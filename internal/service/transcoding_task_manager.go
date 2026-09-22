package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/logger"
	"go-api-starter/internal/platform/oss"
)

// TaskManager manages transcoding tasks
type TaskManager struct {
	db              *sqlx.DB
	cloudTranscoder CloudTranscoder
	providerError   error
}

// NewTaskManager creates a new TaskManager
func NewTaskManager(db *sqlx.DB) *TaskManager {
	return &TaskManager{
		db: db,
	}
}

// SetCloudTranscoder switches newly created tasks to a cloud transcoding
// provider. Existing tasks continue to be tracked by their persisted provider.
func (tm *TaskManager) SetCloudTranscoder(transcoder CloudTranscoder) {
	tm.cloudTranscoder = transcoder
	tm.providerError = nil
}

// SetProviderError makes MPS configuration failures visible to upload callers.
func (tm *TaskManager) SetProviderError(err error) {
	tm.providerError = err
}

// CloudTranscoder returns the configured MPS client for the background poller.
func (tm *TaskManager) CloudTranscoder() CloudTranscoder {
	if tm == nil {
		return nil
	}
	return tm.cloudTranscoder
}

// CreateTask generates a task record and submits the requested outputs to MPS.
func (tm *TaskManager) CreateTask(req model.CreateTaskRequest) (string, error) {
	if tm.providerError != nil {
		return "", tm.providerError
	}
	// Generate unique task_id
	taskID := uuid.New().String()
	if len(req.Resolutions) == 0 {
		req.Resolutions = []string{"original", "1080p", "720p", "480p"}
	}

	if tm.cloudTranscoder == nil {
		return "", fmt.Errorf("Alibaba Cloud MPS is not configured")
	}

	// Create task record
	task := &model.TranscodingTask{
		TaskID:      taskID,
		FileID:      req.FileID,
		SourceURL:   req.SourceURL,
		WebhookURL:  req.WebhookURL,
		Provider:    "aliyun_mps",
		Status:      "processing",
		Resolutions: model.JSONStringList(req.Resolutions),
		Results:     model.JSONResults{},
	}

	// Persist to database
	now := time.Now()
	task.CreatedAt, task.UpdatedAt = now, now
	res, err := tm.db.Exec(`INSERT INTO transcoding_tasks(task_id,file_id,source_url,webhook_url,provider,status,resolutions,external_job_ids,results,error_msg,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, task.TaskID, task.FileID, task.SourceURL, task.WebhookURL, task.Provider, task.Status, task.Resolutions, task.ExternalJobIDs, task.Results, task.ErrorMsg, task.CreatedAt, task.UpdatedAt)
	if err == nil {
		if id, idErr := res.LastInsertId(); idErr == nil {
			task.ID = uint(id)
		}
	}
	if err != nil {
		logger.Log.Errorf("Failed to create task record: %v", err)
		return "", fmt.Errorf("failed to create task record: %w", err)
	}

	logger.Log.Infof("Created task record: %s", taskID)

	// Submit to MPS. The server only creates the task and tracks
	// its external job IDs; it never downloads or transcodes the video.
	jobs, err := tm.cloudTranscoder.Submit(CloudTranscodeRequest{
		TaskID:      taskID,
		SourceURL:   req.SourceURL,
		Resolutions: req.Resolutions,
	})
	if err != nil {
		_ = tm.UpdateTaskStatus(taskID, "failed", nil, err.Error())
		return "", fmt.Errorf("failed to submit cloud transcoding task: %w", err)
	}

	externalJobIDs := make(model.JSONStringList, len(req.Resolutions))
	results := make(model.JSONResults, 0, len(req.Resolutions))
	allTerminal := true
	for index, resolution := range req.Resolutions {
		job := CloudTranscodeJob{Resolution: resolution, Status: "failed", Error: "cloud provider did not return a job"}
		if index < len(jobs) {
			job = jobs[index]
		}
		externalJobIDs[index] = job.ID
		result := model.TranscodingResult{
			Resolution: resolution,
			Status:     job.Status,
			URL:        job.ObjectKey,
			Error:      job.Error,
		}
		if result.Status == "processing" {
			allTerminal = false
		}
		results = append(results, result)
	}

	if _, err := tm.db.Exec(`UPDATE transcoding_tasks SET external_job_ids=?,results=?,updated_at=? WHERE task_id=?`, externalJobIDs, results, time.Now(), taskID); err != nil {
		_ = tm.UpdateTaskStatus(taskID, "failed", nil, err.Error())
		return "", fmt.Errorf("failed to save cloud transcoding jobs: %w", err)
	}
	if allTerminal {
		status := cloudFinalStatus(results)
		if err := tm.UpdateTaskStatus(taskID, status, results, ""); err != nil {
			return "", err
		}
	}
	logger.Log.Infof("Submitted MPS transcoding task: %s", taskID)
	return taskID, nil
}

// ListPendingCloudTasks returns cloud tasks that still need polling.
func (tm *TaskManager) ListPendingCloudTasks() ([]model.TranscodingTask, error) {
	var tasks []model.TranscodingTask
	if err := tm.db.Select(&tasks, `SELECT id,task_id,file_id,source_url,webhook_url,provider,status,resolutions,external_job_ids,results,error_msg,created_at,updated_at,completed_at,deleted_at FROM transcoding_tasks WHERE provider=? AND status=?`, "aliyun_mps", "processing"); err != nil {
		return nil, fmt.Errorf("failed to query pending cloud tasks: %w", err)
	}
	return tasks, nil
}

func cloudFinalStatus(results []model.TranscodingResult) string {
	successCount := 0
	failedCount := 0
	for _, result := range results {
		if result.Status == "success" {
			successCount++
		} else {
			failedCount++
		}
	}
	if failedCount == 0 {
		return "success"
	}
	if successCount == 0 {
		return "failed"
	}
	return "partial_success"
}

// GetTaskStatus retrieves task details from database
func (tm *TaskManager) GetTaskStatus(taskID string) (*model.TranscodingTask, error) {
	var task model.TranscodingTask
	if err := tm.db.Get(&task, `SELECT id,task_id,file_id,source_url,webhook_url,provider,status,resolutions,external_job_ids,results,error_msg,created_at,updated_at,completed_at,deleted_at FROM transcoding_tasks WHERE task_id=? LIMIT 1`, taskID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to query task: %w", err)
	}
	return &task, nil
}

// UpdateTaskStatus updates task status and results
func (tm *TaskManager) UpdateTaskStatus(taskID string, status string, results []model.TranscodingResult, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if results != nil {
		updates["results"] = model.JSONResults(results)
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	// Set completed_at timestamp for terminal states
	if status == "success" || status == "partial_success" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	// Update task record
	resultJSON, _ := json.Marshal(results)
	if _, err := tm.db.Exec(`UPDATE transcoding_tasks SET status=?,results=?,error_msg=?,completed_at=?,updated_at=? WHERE task_id=?`, status, resultJSON, errorMsg, updates["completed_at"], updates["updated_at"], taskID); err != nil {
		logger.Log.Errorf("Failed to update task status: %v", err)
		return fmt.Errorf("failed to update task status: %w", err)
	}

	logger.Log.Infof("Updated task %s status to %s", taskID, status)

	// 如果转码成功,保存转码结果到 video_variants 表
	if (status == "success" || status == "partial_success") && results != nil {
		// 获取任务信息以获取 FileID
		var task model.TranscodingTask
		if err := tm.db.Get(&task, `SELECT id,task_id,file_id,source_url,webhook_url,provider,status,resolutions,external_job_ids,results,error_msg,created_at,updated_at,completed_at,deleted_at FROM transcoding_tasks WHERE task_id=? LIMIT 1`, taskID); err != nil {
			logger.Log.Errorf("Failed to get task for saving variants: %v", err)
			return nil // 不影响主流程
		}

		if task.FileID != nil {
			var file model.File
			if err := tm.db.Get(&file, `SELECT id,transcoding_task_id FROM files WHERE id=?`, *task.FileID); err != nil {
				// The file may have been deleted while a remote job was running.
				logger.Log.Infof("Skipping variants for task %s because its file no longer exists", taskID)
				cleanupTranscodingKeys(results)
				return nil
			}
			if file.TranscodingTaskID == nil || *file.TranscodingTaskID != taskID {
				// A later reprocess owns the file now. Do not let an old MPS
				// callback resurrect stale variants.
				logger.Log.Infof("Skipping stale variants for task %s", taskID)
				cleanupTranscodingKeys(results)
				return nil
			}

			// 保存成功的转码结果到 video_variants
			for _, result := range results {
				if result.Status == "success" && result.URL != "" {
					// result.URL 现在存的是 OSS key（相对路径）
					variant := model.VideoVariant{
						FileID:  *task.FileID,
						Quality: result.Resolution,
						Key:     result.URL,
						Format:  "mp4",
					}
					if result.Size > 0 {
						size := uint(result.Size)
						variant.Size = &size
					}

					// 检查是否已存在
					var existing model.VideoVariant
					err := tm.db.Get(&existing, `SELECT id,file_id,quality,`+"`key`"+`,format,size,width,height,bitrate,fps,duration,created_at FROM video_variants WHERE file_id=? AND quality=? LIMIT 1`, *task.FileID, result.Resolution)
					if err == nil {
						// 已存在,更新
						if existing.Key != "" && existing.Key != variant.Key {
							deleteOSSObject(existing.Key)
						}
						_, _ = tm.db.Exec(`UPDATE video_variants SET `+"`key`"+`=?,format=?,size=? WHERE id=?`, variant.Key, variant.Format, variant.Size, existing.ID)
					} else {
						// 不存在,创建
						if _, err := tm.db.Exec(`INSERT INTO video_variants(file_id,quality,`+"`key`"+`,format,size,created_at) VALUES(?,?,?,?,?,?)`, variant.FileID, variant.Quality, variant.Key, variant.Format, variant.Size, time.Now()); err != nil {
							logger.Log.Errorf("Failed to save video variant: %v", err)
						} else {
							logger.Log.Infof("Saved video variant: %s for file %d", result.Resolution, *task.FileID)
						}
					}
				}
			}
		}
	}

	return nil
}

func deleteOSSObject(key string) {
	if key == "" {
		return
	}
	bucket := oss.GetBucket()
	if bucket == nil {
		return
	}
	if err := bucket.DeleteObject(key); err != nil {
		logger.Log.Warnf("failed to delete transcoding object %s: %v", key, err)
	}
}

func cleanupTranscodingKeys(results []model.TranscodingResult) {
	for _, result := range results {
		if result.Status == "success" {
			deleteOSSObject(result.URL)
		}
	}
}
