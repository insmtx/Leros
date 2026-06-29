package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/insmtx/Leros/backend/internal/api/dto"
)

const defaultHTTPTimeout = 30 * time.Second

type ServerClient struct {
	baseURL    string
	httpClient *http.Client
	appKey     string
}

func NewServerClient(serverAddr, appKey string) *ServerClient {
	return &ServerClient{
		baseURL:    fmt.Sprintf("http://%s", serverAddr),
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		appKey:     appKey,
	}
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *ServerClient) doPost(ctx context.Context, endpoint string, reqBody interface{}, target interface{}) error {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	reqURL := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAppKey(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if apiResp.Code != dto.CodeSuccess {
		return fmt.Errorf("api error [%d]: %s", apiResp.Code, apiResp.Message)
	}

	if target != nil {
		if err := json.Unmarshal(apiResp.Data, target); err != nil {
			return fmt.Errorf("decode data: %w", err)
		}
	}

	return nil
}

func (c *ServerClient) doGet(ctx context.Context, path string, target interface{}) error {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.setAppKey(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found (404)")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if apiResp.Code != dto.CodeSuccess {
		return fmt.Errorf("api error [%d]: %s", apiResp.Code, apiResp.Message)
	}

	if target != nil {
		if err := json.Unmarshal(apiResp.Data, target); err != nil {
			return fmt.Errorf("decode data: %w", err)
		}
	}

	return nil
}

func (c *ServerClient) doGetRaw(ctx context.Context, path string) ([]byte, error) {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAppKey(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found (404)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(io.LimitReader(resp.Body, 100_000_000))
}

func (c *ServerClient) DownloadSkillCache(ctx context.Context, skillID, source, version string) ([]byte, error) {
	baseURL := c.baseURL + "/v1/skill-marketplace/skills/" + url.PathEscape(skillID) + "/download"
	reqURL := fmt.Sprintf("%s?source=%s&version=%s", baseURL, url.QueryEscape(source), url.QueryEscape(version))
	return c.doGetRaw(ctx, reqURL)
}

// StorageConfig represents storage configuration returned by the server.
type StorageConfig struct {
	Scheme string `json:"scheme"`
	Bucket string `json:"bucket"`
}

// GetStorageConfig fetches the storage configuration from the server.
func (c *ServerClient) GetStorageConfig(ctx context.Context) (*StorageConfig, error) {
	var resp StorageConfig
	if err := c.doGet(ctx, "/v1/static/storage-config", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPresignUploadURL fetches a presigned upload URL for the given bucket and key.
func (c *ServerClient) GetPresignUploadURL(ctx context.Context, bucket, key string) (string, error) {
	path := fmt.Sprintf("/v1/static/%s/%s?operation=upload", url.PathEscape(bucket), key)
	data, err := c.doGetRaw(ctx, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *ServerClient) setAppKey(req *http.Request) {
	if c.appKey != "" {
		req.Header.Set("X-App-Key", c.appKey)
	}
}
