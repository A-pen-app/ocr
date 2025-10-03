package store

import (
	"context"

	"github.com/A-pen-app/ocr/models"
)

type OCR interface {
	ScanName(ctx context.Context, link string) (string, error)
	ScanRawInfo(ctx context.Context, userID string, link string, professionType models.PlatformType) (*models.OCRRawInfo, error)
}
