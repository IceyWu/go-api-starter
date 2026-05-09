package utils

import (
	"fmt"
	"image"
	// Register common decoders so image.DecodeConfig recognizes them.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"mime/multipart"
)

// GetImageDimensions returns the width and height of an image file without fully decoding it.
// The file pointer is rewound to the beginning before and after reading.
func GetImageDimensions(file multipart.File) (int, int, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return 0, 0, fmt.Errorf("failed to reset file pointer: %w", err)
	}
	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return 0, 0, fmt.Errorf("failed to reset file pointer: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}
