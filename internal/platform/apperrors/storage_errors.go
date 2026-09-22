package apperrors

import "net/http"

// Object-storage error codes.
const (
	CodeStorageNotInitialized  = "STORAGE_NOT_INITIALIZED"
	CodeStorageInitError       = "STORAGE_INIT_ERROR"
	CodeStorageUploadError     = "STORAGE_UPLOAD_ERROR"
	CodeMultipartInitError     = "MULTIPART_INIT_ERROR"
	CodeMultipartCompleteError = "MULTIPART_COMPLETE_ERROR"
	CodeMultipartAbortError    = "MULTIPART_ABORT_ERROR"
	CodeStorageDeleteError     = "STORAGE_DELETE_ERROR"
	CodeStorageListError       = "STORAGE_LIST_ERROR"
)

// ErrStorageNotInitialized is returned when the object-storage client is not initialized.
var ErrStorageNotInitialized = &AppError{
	Code:       CodeStorageNotInitialized,
	Message:    "object storage client not initialized",
	HTTPStatus: http.StatusServiceUnavailable,
}

// StorageInitError creates an error for object-storage initialization failures.
func StorageInitError(err error) *AppError {
	return &AppError{
		Code:       CodeStorageInitError,
		Message:    "failed to initialize object storage",
		HTTPStatus: http.StatusServiceUnavailable,
		Err:        err,
	}
}

// StorageUploadError creates an error for object-storage upload failures.
func StorageUploadError(err error, details string) *AppError {
	return &AppError{
		Code:       CodeStorageUploadError,
		Message:    "failed to upload file",
		HTTPStatus: http.StatusBadGateway,
		Err:        err,
		Details:    details,
	}
}

// MultipartInitError creates an error for multipart upload initialization failures
func MultipartInitError(err error) *AppError {
	return &AppError{
		Code:       CodeMultipartInitError,
		Message:    "failed to initialize multipart upload",
		HTTPStatus: http.StatusBadGateway,
		Err:        err,
	}
}

// MultipartCompleteError creates an error for multipart upload completion failures
func MultipartCompleteError(err error) *AppError {
	return &AppError{
		Code:       CodeMultipartCompleteError,
		Message:    "failed to complete multipart upload",
		HTTPStatus: http.StatusBadGateway,
		Err:        err,
	}
}

// MultipartAbortError creates an error for multipart upload abort failures
func MultipartAbortError(err error) *AppError {
	return &AppError{
		Code:       CodeMultipartAbortError,
		Message:    "failed to abort multipart upload",
		HTTPStatus: http.StatusBadGateway,
		Err:        err,
	}
}

// StorageDeleteError creates an error for object deletion failures.
func StorageDeleteError(err error, key string) *AppError {
	return &AppError{
		Code:       CodeStorageDeleteError,
		Message:    "failed to delete object",
		HTTPStatus: http.StatusBadGateway,
		Err:        err,
		Details:    key,
	}
}

// StorageListError creates an error for object listing failures.
func StorageListError(err error) *AppError {
	return &AppError{
		Code:       CodeStorageListError,
		Message:    "failed to list objects",
		HTTPStatus: http.StatusBadGateway,
		Err:        err,
	}
}

// FileExtensionNotAllowed creates an error for disallowed file extensions
func FileExtensionNotAllowed(ext string) *AppError {
	return &AppError{
		Code:       "FILE_EXTENSION_NOT_ALLOWED",
		Message:    "file extension is not allowed",
		HTTPStatus: http.StatusBadRequest,
		Details:    ext,
	}
}

// FileNotFound creates an error for file not found
func FileNotFound(id interface{}) *AppError {
	return &AppError{
		Code:       "FILE_NOT_FOUND",
		Message:    "file not found",
		HTTPStatus: http.StatusNotFound,
		Details:    id,
	}
}
