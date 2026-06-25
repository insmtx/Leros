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

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
)

const defaultHTTPTimeout = 30 * time.Second

type ServerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewServerClient(serverAddr string) *ServerClient {
	return &ServerClient{
		baseURL:    fmt.Sprintf("http://%s", serverAddr),
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
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

	if target != nil {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("decode response: %w", err)
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

func (c *ServerClient) PresignArtifactUpload(ctx context.Context, req *contract.PresignArtifactUploadRequest) (*contract.PresignArtifactUploadResponse, error) {
	var resp contract.PresignArtifactUploadResponse
	if err := c.doPost(ctx, "/v1/internal/artifacts/presign-upload", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ServerClient) DownloadSkillCache(ctx context.Context, skillID, source, version string) ([]byte, error) {
	baseURL := c.baseURL + "/v1/skill-marketplace/skills/" + url.PathEscape(skillID) + "/download"
	reqURL := fmt.Sprintf("%s?source=%s&version=%s", baseURL, url.QueryEscape(source), url.QueryEscape(version))
	return c.doGetRaw(ctx, reqURL)
}
