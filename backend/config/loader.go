package config

import (
	"fmt"

	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

// Load 从指定路径或默认位置加载配置
func Load(configPath string) (*Config, error) {
	var cfg Config

	if configPath != "" {
		err := ygconfig.LoadYamlLocalFile(configPath, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
		logs.Infof("Loaded config from: %s", configPath)
		return &cfg, nil
	}

	pathsToTry := []string{"./config.yaml", "/app/config.yaml"}
	var lastErr error

	for _, path := range pathsToTry {
		lastErr = ygconfig.LoadYamlLocalFile(path, &cfg)
		if lastErr == nil {
			logs.Infof("Loaded config from: %s", path)
			return &cfg, nil
		}
	}

	logs.Warnf("Could not load config from any path, will proceed without config")
	return &Config{}, nil
}
