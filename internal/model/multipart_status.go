package model

// MultipartUploadStatus represents the status of a multipart upload
type MultipartUploadStatus string

const (
	// MultipartStatusInitiated indicates the upload has been initiated but no parts uploaded yet
	MultipartStatusInitiated MultipartUploadStatus = "initiated"
	// MultipartStatusInProgress indicates parts are being uploaded
	MultipartStatusInProgress MultipartUploadStatus = "in_progress"
	// MultipartStatusCompleted indicates the upload has been completed successfully
	MultipartStatusCompleted MultipartUploadStatus = "completed"
	// MultipartStatusAborted indicates the upload has been aborted
	MultipartStatusAborted MultipartUploadStatus = "aborted"
)

// IsValid checks if the status is a valid MultipartUploadStatus
func (s MultipartUploadStatus) IsValid() bool {
	switch s {
	case MultipartStatusInitiated, MultipartStatusInProgress, MultipartStatusCompleted, MultipartStatusAborted:
		return true
	}
	return false
}

// CanTransitionTo checks if the current status can transition to the target status
// Valid transitions:
//   - Initiated -> InProgress, Aborted
//   - InProgress -> Completed, Aborted
//   - Completed -> (none)
//   - Aborted -> (none)
func (s MultipartUploadStatus) CanTransitionTo(target MultipartUploadStatus) bool {
	transitions := map[MultipartUploadStatus][]MultipartUploadStatus{
		MultipartStatusInitiated:  {MultipartStatusInProgress, MultipartStatusAborted},
		MultipartStatusInProgress: {MultipartStatusCompleted, MultipartStatusAborted},
		MultipartStatusCompleted:  {},
		MultipartStatusAborted:    {},
	}

	allowed, ok := transitions[s]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}
