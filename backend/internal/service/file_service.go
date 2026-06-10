package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/ygpkg/storage-go"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	infradb "github.com/insmtx/Leros/backend/internal/infra/db"
	appstorage "github.com/insmtx/Leros/backend/internal/infra/storage"
	"github.com/insmtx/Leros/backend/types"
	"github.com/ygpkg/yg-go/encryptor/snowflake"
)

type fileService struct {
	db *gorm.DB
}

var _ contract.FileService = (*fileService)(nil)

func NewFileService(db *gorm.DB) contract.FileService {
	return &fileService{db: db}
}

func (s *fileService) UploadFile(ctx context.Context, req *contract.UploadFileRequest) (*contract.UploadFileResult, error) {
	caller, err := requireCallerOrg(ctx)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(req.File)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	hash := sha256.Sum256(data)
	sha256Hex := hex.EncodeToString(hash[:])

	ext := ""
	if idx := strings.LastIndex(req.Filename, "."); idx >= 0 {
		ext = req.Filename[idx:]
	}
	storeFilename := fmt.Sprintf("%s%s", snowflake.GenerateIDBase58(), ext)
	key := fmt.Sprintf("%s/%d/%s", req.Purpose, caller.OrgID, storeFilename)

	st := appstorage.Get()
	bucket := appstorage.DefaultBucket()

	result, err := st.PutObject(ctx, bucket, key, bytes.NewReader(data),
		storage.WithContentType(req.MimeType),
	)
	if err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}

	publicID := fmt.Sprintf("file_%s", snowflake.GenerateIDBase58())
	file := &types.FileUpload{
		PublicID:     publicID,
		OrgID:        caller.OrgID,
		OwnerID:      caller.Uin,
		Filename:     storeFilename,
		OriginalName: req.Filename,
		MimeType:     req.MimeType,
		FileSize:     int64(len(data)),
		StoragePath:  result.Path.Path(),
		Sha256:       sha256Hex,
		Purpose:      req.Purpose,
		Status:       "active",
	}

	if err := infradb.CreateFileUpload(ctx, s.db, file); err != nil {
		return nil, fmt.Errorf("create file upload record: %w", err)
	}

	return &contract.UploadFileResult{
		FileUploadID: publicID,
		Filename:     storeFilename,
		OriginalName: req.Filename,
		MimeType:     req.MimeType,
		FileSize:     file.FileSize,
		Sha256:       sha256Hex,
		URL:          result.Path.PublicURL(),
	}, nil
}

func (s *fileService) GetFileDownloadURL(ctx context.Context, orgID uint, fileID string) (*contract.FileDownloadURL, error) {
	file, err := infradb.GetFileUploadByPublicID(ctx, s.db, orgID, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

	st := appstorage.Get()
	bucket := appstorage.DefaultBucket()

	ttl := 30 * time.Minute
	url, err := st.PresignGetObject(ctx, bucket, file.StoragePath, ttl)
	if err != nil {
		return nil, fmt.Errorf("presign url: %w", err)
	}

	return &contract.FileDownloadURL{
		URL:       url,
		Filename:  file.OriginalName,
		MimeType:  file.MimeType,
		FileSize:  file.FileSize,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}, nil
}
