package filestore

import (
	"context"
	"time"
)

const defaultPresignTTL = 1 * time.Hour

// PresignUpload 生成预签名上传 URL
func PresignUpload(ctx context.Context, bucket, key string) (string, time.Time, error) {
	st := GetStorage()
	expiresAt := time.Now().Add(defaultPresignTTL)
	url, err := st.PresignPutObject(ctx, bucket, key, defaultPresignTTL)
	if err != nil {
		return "", time.Time{}, err
	}
	return url, expiresAt, nil
}

// PresignDownload 生成预签名下载 URL
func PresignDownload(ctx context.Context, bucket, key string) (string, time.Time, error) {
	st := GetStorage()
	expiresAt := time.Now().Add(defaultPresignTTL)
	url, err := st.PresignGetObject(ctx, bucket, key, defaultPresignTTL)
	if err != nil {
		return "", time.Time{}, err
	}
	return url, expiresAt, nil
}
