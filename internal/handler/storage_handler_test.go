package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"go-api-starter/internal/middleware"
	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/internal/transport"
)

type storageServiceStub struct {
	service.StorageServiceInterface
	file        *model.File
	deletedFile string
}

func (s *storageServiceStub) GetFileByUID(string) (*model.File, error) { return s.file, nil }

func (s *storageServiceStub) DeleteFile(uid string) error {
	s.deletedFile = uid
	return nil
}

func TestPrivateFileRequiresOwner(t *testing.T) {
	for _, tc := range []struct {
		name       string
		userID     uint
		isPrivate  bool
		wantStatus int
	}{
		{name: "public file is visible anonymously", wantStatus: http.StatusOK},
		{name: "private file is hidden anonymously", isPrivate: true, wantStatus: http.StatusForbidden},
		{name: "private file is visible to owner", userID: 7, isPrivate: true, wantStatus: http.StatusOK},
		{name: "private file is hidden from another user", userID: 8, isPrivate: true, wantStatus: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &storageServiceStub{file: &model.File{UID: "file-uid", UserID: 7, IsPrivate: tc.isPrivate}}
			handler := NewStorageHandler(svc, nil)
			r := transport.New()
			r.Use(middleware.ErrorHandler())
			r.GET("/files/:uid", func(c *transport.Context) {
				if tc.userID > 0 {
					c.Set("userID", tc.userID)
				}
				c.Next()
			}, handler.GetFile)

			req := httptest.NewRequest(http.MethodGet, "/files/file-uid", nil)
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			require.Equal(t, tc.wantStatus, res.Code)
		})
	}
}

func TestDeleteFileRequiresOwner(t *testing.T) {
	for _, tc := range []struct {
		name       string
		userID     uint
		wantStatus int
		wantDelete bool
	}{
		{name: "anonymous denied", wantStatus: http.StatusUnauthorized},
		{name: "different user denied", userID: 8, wantStatus: http.StatusForbidden},
		{name: "owner can delete", userID: 7, wantStatus: http.StatusOK, wantDelete: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &storageServiceStub{file: &model.File{UID: "file-uid", UserID: 7}}
			handler := NewStorageHandler(svc, nil)
			r := transport.New()
			r.Use(middleware.ErrorHandler())
			r.DELETE("/files/:uid", func(c *transport.Context) {
				if tc.userID > 0 {
					c.Set("userID", tc.userID)
				}
				c.Next()
			}, handler.DeleteFile)

			req := httptest.NewRequest(http.MethodDelete, "/files/file-uid", nil)
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			require.Equal(t, tc.wantStatus, res.Code)
			if tc.wantDelete {
				require.Equal(t, "file-uid", svc.deletedFile)
			} else {
				require.Empty(t, svc.deletedFile)
			}
		})
	}
}
