package apperrors

import "net/http"

// OSS-related error codes
const (
	CodeOSSNotInitialized    = "OSS_NOT_INITIALIZED"
	CodeOSSInitError         = "OSS_INIT_ERROR"
	CodeOSSUploadError       = "OSS_UPLOAD_ERROR"
	CodeMultipartInitError   = "MULTIPART_INIT_ERROR"
	CodeMultipartCompleteError = "MULTIPART_COMPLETE_ERROR"
	CodeMultipartAbortError  = "MULTIPART_ABORT_ERROR"
	CodeOSSDeleteError       = "OSS_DELETE_ERROR"
	CodeOSSListError         = "OSS_LIST_ERROR"
)

// ErrOSSNotInitialized is returned when OSS client is not initialized
var ErrOSSNotInitialized = &AppError{
	Code:       CodeOSSNotInitialized,
	Message:    "OSS client not initialized",
	HTTPStatus: http.StatusServiceUnavailable,
}

// OSSInitError creates an error for OSS initialization failures
func OSSInitError(err error) *AppError {
	return &AppError{
		Code:       CodeOSSInitError,
		Message:    "failed to initialize OSS",
		HTTPStatus: http.StatusServiceUnavailable,
		Err:        err,
	}
}

// OSSUploadError creates an error for OSS upload failures
func OSSUploadError(err error, details string) *AppError {
	return &AppError{
		Code:       CodeOSSUploadError,
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

// OSSDeleteError creates an error for OSS file deletion failures
func OSSDeleteError(err error, key string) *AppError {
	return &AppError{
		Code:       CodeOSSDeleteError,
		Message:    "failed to delete file from OSS",
		HTTPStatus: http.StatusBadGateway,
		Err:        err,
		Details:    key,
	}
}

// OSSListError creates an error for OSS list operations failures
func OSSListError(err error) *AppError {
	return &AppError{
		Code:       CodeOSSListError,
		Message:    "failed to list files from OSS",
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
