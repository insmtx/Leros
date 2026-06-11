package filestore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ygpkg/storage-go"
	_ "github.com/ygpkg/storage-go/driver/local"
	_ "github.com/ygpkg/storage-go/driver/minio"

	"github.com/insmtx/Leros/backend/config"
)

var (
	defaultStorage storage.Storage
	defaultBucket  string
)

func Init(cfg *config.StorageConfig) error {
	if cfg == nil {
		var root string
		if exe, err := os.Executable(); err == nil {
			root = filepath.Dir(exe)
		} else {
			root, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
		}
		cfg = &config.StorageConfig{
			Driver:   "local",
			LocalDir: filepath.Join(root, "leros-bucket"),
			Bucket:   "bucket",
		}
	}
	driver := storage.DriverType(cfg.Driver)
	sCfg := storage.Config{
		Endpoint:  cfg.Endpoint,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		Bucket:    cfg.Bucket,
		UseSSL:    cfg.UseSSL,
		BaseDir:   cfg.LocalDir,
	}
	var pb storage.PathBuilder
	if driver == storage.DriverLocal {
		pb = &storage.LocalPathBuilder{
			AbsDir:  cfg.LocalDir,
			BaseURL: cfg.BaseURL,
		}
	} else {
		urlStyle := storage.URLStylePath
		if cfg.URLStyle == "virtual-hosted" {
			urlStyle = storage.URLStyleVirtualHosted
		}
		pb = &storage.S3PathBuilder{
			BaseURL:  cfg.BaseURL,
			Endpoint: cfg.Endpoint,
			URLStyle: urlStyle,
		}
	}
	s, err := storage.New(driver, sCfg, pb)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}
	defaultStorage = s
	defaultBucket = cfg.Bucket
	return nil
}

func GetStorage() storage.Storage {
	return defaultStorage
}

func DefaultBucket() string {
	return defaultBucket
}
