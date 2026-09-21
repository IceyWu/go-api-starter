package service

import (
	"encoding/json"
	"strings"
	"time"

	"go-api-starter/internal/model"
)

type ClientMediaMetadata struct {
	SchemaVersion int                  `json:"schema_version"`
	Basic         ClientBasicMetadata  `json:"basic"`
	Image         *ClientImageMetadata `json:"image,omitempty"`
	Video         *ClientVideoMetadata `json:"video,omitempty"`
}
type ClientBasicMetadata struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Extension string `json:"extension"`
	Size      int64  `json:"size"`
	MD5       string `json:"md5"`
	Width     *uint  `json:"width,omitempty"`
	Height    *uint  `json:"height,omitempty"`
}
type ClientImageMetadata struct {
	Blurhash     *string            `json:"blurhash,omitempty"`
	Arthash      *string            `json:"arthash,omitempty"`
	ArthashCodec *string            `json:"arthash_codec,omitempty"`
	Colors       []ClientMediaColor `json:"colors,omitempty"`
	Exif         ClientImageExif    `json:"exif"`
}
type ClientMediaColor struct {
	Hex        string  `json:"hex"`
	R          uint8   `json:"r"`
	G          uint8   `json:"g"`
	B          uint8   `json:"b"`
	Percentage float64 `json:"percentage"`
	IsPrimary  bool    `json:"is_primary"`
	Rank       uint8   `json:"rank"`
}
type ClientImageExif struct {
	Lat          *float64               `json:"lat,omitempty"`
	Lng          *float64               `json:"lng,omitempty"`
	Altitude     *float64               `json:"altitude,omitempty"`
	TakenAt      string                 `json:"taken_at,omitempty"`
	DeviceMake   string                 `json:"device_make,omitempty"`
	DeviceModel  string                 `json:"device_model,omitempty"`
	LensModel    string                 `json:"lens_model,omitempty"`
	FNumber      string                 `json:"f_number,omitempty"`
	ExposureTime string                 `json:"exposure_time,omitempty"`
	ISO          *int                   `json:"iso,omitempty"`
	FocalLength  string                 `json:"focal_length,omitempty"`
	Raw          map[string]interface{} `json:"raw,omitempty"`
}
type ClientVideoMetadata struct {
	Width        *uint                  `json:"width,omitempty"`
	Height       *uint                  `json:"height,omitempty"`
	Duration     *float64               `json:"duration,omitempty"`
	CreationTime string                 `json:"creation_time,omitempty"`
	Codec        string                 `json:"codec,omitempty"`
	Bitrate      *uint                  `json:"bitrate,omitempty"`
	FrameRate    *float64               `json:"frame_rate,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Exif         *ClientImageExif       `json:"exif,omitempty"`
}

func applyClientMetadata(file *model.File, md *ClientMediaMetadata) {
	file.Width, file.Height = md.Basic.Width, md.Basic.Height
	if md.Image != nil {
		file.Blurhash, file.Arthash, file.ArthashCodec = md.Image.Blurhash, md.Image.Arthash, md.Image.ArthashCodec
		applyExif(file, md.Image.Exif)
		return
	}
	if md.Video != nil {
		file.Width, file.Height, file.Duration, file.FrameRate = md.Video.Width, md.Video.Height, md.Video.Duration, md.Video.FrameRate
		file.Codec = optionalString(md.Video.Codec)
		file.Bitrate = md.Video.Bitrate
		if len(md.Video.Metadata) > 0 {
			file.VideoMetadata, _ = json.Marshal(md.Video.Metadata)
		}
		if md.Video.CreationTime != "" {
			if t, err := time.Parse(time.RFC3339, md.Video.CreationTime); err == nil {
				file.TakenAt = &t
			}
		}
		if md.Video.Exif != nil {
			applyExif(file, *md.Video.Exif)
		}
	}
}
func applyExif(file *model.File, exif ClientImageExif) {
	file.Lat, file.Lng, file.Altitude, file.ISO = exif.Lat, exif.Lng, exif.Altitude, exif.ISO
	file.DeviceMake, file.DeviceModel, file.LensModel = optionalString(exif.DeviceMake), optionalString(exif.DeviceModel), optionalString(exif.LensModel)
	file.FNumber, file.ExposureTime, file.FocalLength = optionalString(exif.FNumber), optionalString(exif.ExposureTime), optionalString(exif.FocalLength)
	if exif.TakenAt != "" {
		if t, err := time.Parse(time.RFC3339, exif.TakenAt); err == nil {
			file.TakenAt = &t
		}
	}
	if len(exif.Raw) > 0 {
		file.ExifRaw, _ = json.Marshal(exif.Raw)
	}
}
func optionalString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
