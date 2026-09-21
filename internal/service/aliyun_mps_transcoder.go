package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/mts"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/logger"
)

// CloudTranscodeRequest is the provider-neutral input used by TaskManager.
type CloudTranscodeRequest struct {
	TaskID      string
	SourceURL   string
	Resolutions []string
}

// CloudTranscodeJob is aligned with CloudTranscodeRequest.Resolutions.
// Status is one of processing, success, or failed.
type CloudTranscodeJob struct {
	ID         string
	Resolution string
	Status     string
	ObjectKey  string
	Error      string
}

// CloudTranscodeStatus is the normalized result returned by a provider query.
type CloudTranscodeStatus struct {
	ID        string
	Status    string
	ObjectKey string
	Error     string
}

// CloudTranscoder abstracts a remote video transcoding provider.
type CloudTranscoder interface {
	Submit(req CloudTranscodeRequest) ([]CloudTranscodeJob, error)
	Query(jobIDs []string) (map[string]CloudTranscodeStatus, error)
}

// AliyunMPSClient submits OSS-to-OSS jobs to Alibaba Cloud MPS. The API server
// only calls MPS and polls status; media bytes never pass through this process.
type AliyunMPSClient struct {
	client      *mts.Client
	bucket      string
	region      string // MPS region ID, e.g. cn-hangzhou
	ossLocation string // OSS location, e.g. oss-cn-hangzhou
	pipelineID  string
	storageRoot string
	templates   map[string]string
}

// NewAliyunMPSClient creates an MPS client when all mandatory settings exist.
func NewAliyunMPSClient(ossCfg *config.OSSConfig, transcodingCfg *config.TranscodingConfig) (*AliyunMPSClient, error) {
	if ossCfg == nil || transcodingCfg == nil {
		return nil, fmt.Errorf("OSS and transcoding configuration are required")
	}
	if ossCfg.AccessKeyID == "" || ossCfg.AccessKeySecret == "" {
		return nil, fmt.Errorf("Alibaba Cloud access key is not configured")
	}
	bucket := ossCfg.Bucket
	if bucket == "" {
		bucket = ossCfg.BucketName
	}
	if bucket == "" {
		return nil, fmt.Errorf("OSS bucket is not configured")
	}
	ossLocation := ossCfg.Region
	region := transcodingCfg.MPSRegion
	if region == "" {
		region = ossLocation
	}
	region = normalizeMPSRegion(region)
	if region == "" {
		return nil, fmt.Errorf("MPS region is not configured")
	}
	ossLocation = normalizeOSSLocation(ossLocation, region)
	if transcodingCfg.MPSPipelineID == "" {
		return nil, fmt.Errorf("MPS pipeline ID is not configured")
	}

	client, err := mts.NewClientWithAccessKey(region, ossCfg.AccessKeyID, ossCfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create MPS client: %w", err)
	}

	return &AliyunMPSClient{
		client:      client,
		bucket:      bucket,
		region:      region,
		ossLocation: ossLocation,
		pipelineID:  transcodingCfg.MPSPipelineID,
		storageRoot: strings.Trim(transcodingCfg.StorageRoot, "/"),
		templates: map[string]string{
			"original": transcodingCfg.MPSTemplateOriginal,
			"1080p":    transcodingCfg.MPSTemplate1080p,
			"720p":     transcodingCfg.MPSTemplate720p,
			"480p":     transcodingCfg.MPSTemplate480p,
		},
	}, nil
}

// Submit creates one MPS job per requested output in a single SubmitJobs call.
// "original" is represented by the already uploaded OSS object and therefore
// does not incur a second cloud processing job.
func (c *AliyunMPSClient) Submit(req CloudTranscodeRequest) ([]CloudTranscodeJob, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("MPS client is not initialized")
	}
	sourceKey := objectKeyFromURL(req.SourceURL)
	if sourceKey == "" {
		return nil, fmt.Errorf("cannot derive OSS object key from source URL")
	}
	storageDir := c.storageRoot
	if storageDir == "" {
		storageDir = "videos"
	}

	jobs := make([]CloudTranscodeJob, len(req.Resolutions))
	outputs := make([]map[string]string, 0, len(req.Resolutions))
	outputIndexes := make([]int, 0, len(req.Resolutions))
	for index, resolution := range req.Resolutions {
		jobs[index] = CloudTranscodeJob{Resolution: resolution, Status: "failed"}
		if resolution == "original" {
			jobs[index].Status = "success"
			jobs[index].ObjectKey = sourceKey
			continue
		}

		templateID := c.templates[resolution]
		if templateID == "" {
			jobs[index].Error = fmt.Sprintf("MPS template ID is not configured for %s", resolution)
			continue
		}
		objectKey := fmt.Sprintf("%s/trans/%s/%s.mp4", storageDir, req.TaskID, resolution)
		userData, _ := json.Marshal(map[string]string{
			"task_id":    req.TaskID,
			"resolution": resolution,
		})
		outputs = append(outputs, map[string]string{
			"OutputObject": url.QueryEscape(objectKey),
			"TemplateId":   templateID,
			"UserData":     string(userData),
		})
		outputIndexes = append(outputIndexes, index)
		jobs[index].ObjectKey = objectKey
		jobs[index].Status = "processing"
	}

	if len(outputs) == 0 {
		return jobs, nil
	}

	inputJSON, err := json.Marshal(map[string]string{
		"Bucket":   c.bucket,
		"Location": c.ossLocation,
		"Object":   url.QueryEscape(sourceKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode MPS input: %w", err)
	}
	outputsJSON, err := json.Marshal(outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to encode MPS outputs: %w", err)
	}

	request := mts.CreateSubmitJobsRequest()
	request.Input = string(inputJSON)
	request.Outputs = string(outputsJSON)
	request.OutputBucket = c.bucket
	request.OutputLocation = c.ossLocation
	request.PipelineId = c.pipelineID
	response, err := c.client.SubmitJobs(request)
	if err != nil {
		return nil, fmt.Errorf("MPS SubmitJobs failed: %w", err)
	}

	for resultIndex, result := range response.JobResultList.JobResult {
		if resultIndex >= len(outputIndexes) {
			break
		}
		jobIndex := outputIndexes[resultIndex]
		if !result.Success || result.Job.JobId == "" {
			jobs[jobIndex].Status = "failed"
			jobs[jobIndex].Error = firstNonEmpty(result.Message, result.Code, "MPS rejected the output job")
			continue
		}
		jobs[jobIndex].ID = result.Job.JobId
		jobs[jobIndex].Status = "processing"
	}
	for index := range jobs {
		if jobs[index].Status == "processing" && jobs[index].ID == "" {
			jobs[index].Status = "failed"
			jobs[index].Error = "MPS did not return a job ID"
		}
	}

	logger.Log.Infof("Submitted %d MPS output jobs for task %s", len(outputs), req.TaskID)
	return jobs, nil
}

// Query fetches status for a batch of MPS jobs.
func (c *AliyunMPSClient) Query(jobIDs []string) (map[string]CloudTranscodeStatus, error) {
	statuses := make(map[string]CloudTranscodeStatus, len(jobIDs))
	validJobIDs := make([]string, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		if strings.TrimSpace(jobID) != "" {
			validJobIDs = append(validJobIDs, jobID)
		}
	}
	if len(validJobIDs) == 0 {
		return statuses, nil
	}
	request := mts.CreateQueryJobListRequest()
	request.JobIds = strings.Join(validJobIDs, ",")
	response, err := c.client.QueryJobList(request)
	if err != nil {
		return nil, fmt.Errorf("MPS QueryJobList failed: %w", err)
	}
	for _, job := range response.JobList.Job {
		status := normalizeMPSStatus(job.State)
		objectKey, _ := url.QueryUnescape(job.Output.OutputFile.Object)
		if objectKey == "" {
			objectKey, _ = url.QueryUnescape(job.Output.ExtendData)
		}
		statuses[job.JobId] = CloudTranscodeStatus{
			ID:        job.JobId,
			Status:    status,
			ObjectKey: objectKey,
			Error:     firstNonEmpty(job.Message, job.Code),
		}
	}
	return statuses, nil
}

func normalizeMPSStatus(state string) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "success", "succeeded", "finished", "complete", "completed", "transcodesuccess", "transcode_success":
		return "success"
	case "fail", "failed", "error", "cancelled", "canceled", "transcodefail", "transcodefailed", "transcode_fail", "transcode_failed":
		return "failed"
	default:
		return "processing"
	}
}

func objectKeyFromURL(source string) string {
	parsed, err := url.Parse(source)
	if err == nil && parsed.Host != "" {
		key, _ := url.PathUnescape(strings.TrimPrefix(parsed.Path, "/"))
		return key
	}
	return strings.TrimPrefix(source, "/")
}

func normalizeMPSRegion(region string) string {
	return strings.TrimSpace(strings.TrimPrefix(region, "oss-"))
}

func normalizeOSSLocation(location, region string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return "oss-" + region
	}
	if strings.HasPrefix(location, "oss-") {
		return location
	}
	return "oss-" + normalizeMPSRegion(location)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// MPSTaskPoller keeps the existing task status API unchanged while completing
// asynchronous MPS jobs and writing video_variants through TaskManager.
type MPSTaskPoller struct {
	taskManager  *TaskManager
	transcoder   CloudTranscoder
	webhook      *WebhookNotifier
	pollInterval time.Duration
}

func NewMPSTaskPoller(taskManager *TaskManager, transcoder CloudTranscoder, webhook *WebhookNotifier, interval time.Duration) *MPSTaskPoller {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	return &MPSTaskPoller{taskManager: taskManager, transcoder: transcoder, webhook: webhook, pollInterval: interval}
}

func (p *MPSTaskPoller) Start(ctxDone <-chan struct{}) {
	if p == nil || p.taskManager == nil || p.transcoder == nil {
		return
	}
	p.PollOnce()
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.PollOnce()
		case <-ctxDone:
			return
		}
	}
}

func (p *MPSTaskPoller) PollOnce() {
	tasks, err := p.taskManager.ListPendingCloudTasks()
	if err != nil {
		logger.Log.Warnf("failed to list pending MPS tasks: %v", err)
		return
	}
	allJobIDs := make([]string, 0)
	for _, task := range tasks {
		for _, jobID := range task.ExternalJobIDs {
			if strings.TrimSpace(jobID) != "" {
				allJobIDs = append(allJobIDs, jobID)
			}
		}
	}
	statuses, err := p.transcoder.Query(allJobIDs)
	if err != nil {
		logger.Log.Warnf("failed to poll MPS tasks: %v", err)
		return
	}

	for _, task := range tasks {
		results := make([]model.TranscodingResult, 0, len(task.Resolutions))
		allTerminal := true
		for index, resolution := range task.Resolutions {
			result := model.TranscodingResult{Resolution: resolution, Status: "failed", Error: "MPS job was not returned"}
			if index < len(task.Results) {
				result = task.Results[index]
			}
			if index < len(task.ExternalJobIDs) && task.ExternalJobIDs[index] != "" {
				jobID := task.ExternalJobIDs[index]
				if status, ok := statuses[jobID]; ok {
					result.Status = status.Status
					if status.ObjectKey != "" {
						result.URL = status.ObjectKey
					}
					result.Error = status.Error
				} else {
					result.Status = "processing"
				}
			}
			if result.Status == "processing" {
				allTerminal = false
			}
			results = append(results, result)
		}
		if !allTerminal {
			continue
		}
		status := cloudFinalStatus(results)
		if err := p.taskManager.UpdateTaskStatus(task.TaskID, status, results, ""); err != nil {
			logger.Log.Warnf("failed to update MPS task %s: %v", task.TaskID, err)
			continue
		}
		if task.WebhookURL != "" && p.webhook != nil {
			go func(taskID, webhookURL, finalStatus string, finalResults []model.TranscodingResult) {
				if err := p.webhook.NotifyWebhook(webhookURL, WebhookPayload{TaskID: taskID, Status: finalStatus, Results: finalResults}); err != nil {
					logger.Log.Warnf("failed to send MPS webhook for task %s: %v", taskID, err)
				}
			}(task.TaskID, task.WebhookURL, status, results)
		}
	}
}
