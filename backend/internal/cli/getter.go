package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
)

// GetSession 调用服务端 GetSession API 并返回解析后的结果。
func GetSession(ctx context.Context, serverAddr, sessionID string) (*contract.Session, error) {
	var result contract.Session
	if err := doPostRequest(ctx, serverAddr, "GetSession",
		map[string]string{"session_id": sessionID}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTask 调用服务端 GetTask API 并返回解析后的结果。
func GetTask(ctx context.Context, serverAddr, publicID string) (*contract.Task, error) {
	var result contract.Task
	if err := doPostRequest(ctx, serverAddr, "GetTask",
		map[string]string{"public_id": publicID}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetProject 调用服务端 GetProject API 并返回解析后的结果。
func GetProject(ctx context.Context, serverAddr, publicID string) (*contract.Project, error) {
	var result contract.Project
	if err := doPostRequest(ctx, serverAddr, "GetProject",
		map[string]string{"public_id": publicID}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DetailProject 调用服务端 DetailProject API 并返回解析后的结果。
func DetailProject(ctx context.Context, serverAddr, publicID string) (*contract.ProjectDetail, error) {
	var result contract.ProjectDetail
	if err := doPostRequest(ctx, serverAddr, "DetailProject",
		map[string]string{"public_id": publicID}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSessionMessages 调用服务端 GetSessionMessages API 并返回解析后的结果。
func GetSessionMessages(ctx context.Context, serverAddr, sessionID string, page, perPage int) (*contract.MessageList, error) {
	var result contract.MessageList
	if err := doPostRequest(ctx, serverAddr, "GetSessionMessages",
		map[string]interface{}{
			"session_id": sessionID,
			"page":       page,
			"per_page":   perPage,
		}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDigitalAssistantByID 调用服务端 GetDigitalAssistant API 并返回解析后的结果。
func GetDigitalAssistantByID(ctx context.Context, serverAddr string, id uint) (*contract.DigitalAssistantDetail, error) {
	var result contract.DigitalAssistantDetail
	if err := doPostRequest(ctx, serverAddr, "GetDigitalAssistant",
		map[string]interface{}{"id": id}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ResolveUserName 通过 Uin 解析用户名称，优先使用 GetUserOrg，回退到 GetOrgMember。
func ResolveUserName(ctx context.Context, serverAddr string, uin uint) string {
	org, err := getUserOrgByUin(ctx, serverAddr, uin)
	if err == nil {
		if org.UserName != "" {
			return org.UserName
		}
		return org.UserLogin
	}
	member, err := getOrgMemberByUin(ctx, serverAddr, uin)
	if err == nil {
		if member.UserName != "" {
			return member.UserName
		}
		return member.UserLogin
	}
	return ""
}

func getUserOrgByUin(ctx context.Context, serverAddr string, uin uint) (*contract.UserOrg, error) {
	var result contract.UserOrg
	if err := doPostRequest(ctx, serverAddr, "GetUserOrg",
		map[string]interface{}{"uin": uin}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func getOrgMemberByUin(ctx context.Context, serverAddr string, uin uint) (*contract.OrgMember, error) {
	var result contract.OrgMember
	if err := doPostRequest(ctx, serverAddr, "GetOrgMember",
		map[string]interface{}{"uin": uin}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListTaskArtifacts 调用服务端 ListTaskArtifacts API 并返回解析后的结果。
func ListTaskArtifacts(ctx context.Context, serverAddr, taskID string) ([]contract.Artifact, error) {
	var result []contract.Artifact
	if err := doGetRequest(ctx, serverAddr, fmt.Sprintf("tasks/%s/artifacts", taskID), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// doPostRequest 发送 POST JSON API 请求的通用封装。
func doPostRequest(ctx context.Context, serverAddr, endpoint string, reqBody, target interface{}) error {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: defaultHTTPTimeout}
	url := fmt.Sprintf("http://%s/v1/%s", serverAddr, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return doRequest(client, req, target)
}

// doGetRequest 发送 GET API 请求的通用封装（用于 REST 风格端点）。
func doGetRequest(ctx context.Context, serverAddr, path string, target interface{}) error {
	client := &http.Client{Timeout: defaultHTTPTimeout}
	url := fmt.Sprintf("http://%s/v1/%s", serverAddr, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	return doRequest(client, req, target)
}

func doRequest(client *http.Client, req *http.Request, target interface{}) error {
	resp, err := client.Do(req)
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

	if err := json.Unmarshal(apiResp.Data, target); err != nil {
		return fmt.Errorf("decode data: %w", err)
	}

	return nil
}
