package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// UpstreamCostHandler handles upstream cost analysis proxy requests
type UpstreamCostHandler struct {
	client *http.Client
}

// NewUpstreamCostHandler creates a new upstream cost handler
func NewUpstreamCostHandler() *UpstreamCostHandler {
	return &UpstreamCostHandler{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// --- Request types ---

type upstreamProxyRequest struct {
	ProviderType string `json:"provider_type" binding:"required,oneof=newapi sub2api"` // newapi or sub2api
	BaseURL      string `json:"base_url" binding:"required,url"`
	AccessToken  string `json:"access_token" binding:"required"`
	UserID       int64  `json:"user_id" binding:"required"`
}

type upstreamLogsRequest struct {
	upstreamProxyRequest
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Model    string `json:"model"`
	TokenID  int64  `json:"token_id"`
}

// --- New-API response types ---

type newAPIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type newAPIUserInfo struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
	Role         int    `json:"role"`
	Status       int    `json:"status"`
	Quota        int64  `json:"quota"`
	UsedQuota    int64  `json:"used_quota"`
	RequestCount int64  `json:"request_count"`
	Group        string `json:"group"`
}

type newAPILogItem struct {
	ID               int64  `json:"id"`
	UserID           int64  `json:"user_id"`
	CreatedAt        int64  `json:"created_at"`
	Type             int    `json:"type"`
	Content          string `json:"content"`
	Username         string `json:"username"`
	TokenName        string `json:"token_name"`
	ModelName        string `json:"model_name"`
	Quota            int64  `json:"quota"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	UseTime          int    `json:"use_time"`
	IsStream         bool   `json:"is_stream"`
	Channel          int64  `json:"channel"`
	ChannelName      string `json:"channel_name"`
	TokenID          int64  `json:"token_id"`
	Group            string `json:"group"`
	Other            string `json:"other"`
	RequestID        string `json:"request_id"`
}

type newAPILogListData struct {
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Total    int64           `json:"total"`
	Items    []newAPILogItem `json:"items"`
}

type newAPIStatData struct {
	Quota int64 `json:"quota"`
	RPM   int64 `json:"rpm"`
	TPM   int64 `json:"tpm"`
}

// --- Unified response types (returned to our frontend) ---

type upstreamUserInfoResponse struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
	Quota        int64  `json:"quota"`
	UsedQuota    int64  `json:"used_quota"`
	RequestCount int64  `json:"request_count"`
	Group        string `json:"group"`
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
}

type upstreamLogItem struct {
	ID               int64   `json:"id"`
	CreatedAt        int64   `json:"created_at"`
	ModelName        string  `json:"model_name"`
	Quota            int64   `json:"quota"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	UseTime          int     `json:"use_time"`
	IsStream         bool    `json:"is_stream"`
	TokenName        string  `json:"token_name"`
	Channel          int64   `json:"channel"`
	Group            string  `json:"group"`
	QuotaUSD         float64 `json:"quota_usd"` // quota / 500000 for New-API default
}

type upstreamLogsResponse struct {
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
	Items    []upstreamLogItem `json:"items"`
}

type upstreamStatResponse struct {
	Quota    int64   `json:"quota"`
	QuotaUSD float64 `json:"quota_usd"`
	RPM      int64   `json:"rpm"`
	TPM      int64   `json:"tpm"`
}

// --- Helper: do upstream request ---

func (h *UpstreamCostHandler) doUpstreamGet(baseURL, path, accessToken string, userID int64) ([]byte, int, error) {
	url := strings.TrimSuffix(baseURL, "/") + path
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("New-Api-User", fmt.Sprintf("%d", userID))

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB limit
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return body, resp.StatusCode, nil
}

// GetUserInfo fetches user info from upstream
func (h *UpstreamCostHandler) GetUserInfo(c *gin.Context) {
	var req upstreamProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	body, statusCode, err := h.doUpstreamGet(req.BaseURL, "/api/user/self/", req.AccessToken, req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to fetch upstream user info: "+err.Error())
		return
	}
	if statusCode != http.StatusOK {
		var errResp newAPIResponse
		_ = json.Unmarshal(body, &errResp)
		response.Error(c, http.StatusBadGateway, fmt.Sprintf("Upstream error (%d): %s", statusCode, errResp.Message))
		return
	}

	var apiResp newAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse upstream response")
		return
	}

	var user newAPIUserInfo
	if err := json.Unmarshal(apiResp.Data, &user); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse user data")
		return
	}

	response.Success(c, upstreamUserInfoResponse{
		ID:           user.ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Email:        user.Email,
		Quota:        user.Quota,
		UsedQuota:    user.UsedQuota,
		RequestCount: user.RequestCount,
		Group:        user.Group,
		ProviderType: req.ProviderType,
		BaseURL:      req.BaseURL,
	})
}

// GetStats fetches usage statistics from upstream
func (h *UpstreamCostHandler) GetStats(c *gin.Context) {
	var req upstreamProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	body, statusCode, err := h.doUpstreamGet(req.BaseURL, "/api/log/self/stat", req.AccessToken, req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to fetch upstream stats: "+err.Error())
		return
	}
	if statusCode != http.StatusOK {
		var errResp newAPIResponse
		_ = json.Unmarshal(body, &errResp)
		response.Error(c, http.StatusBadGateway, fmt.Sprintf("Upstream error (%d): %s", statusCode, errResp.Message))
		return
	}

	var apiResp newAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse upstream response")
		return
	}

	var stat newAPIStatData
	if err := json.Unmarshal(apiResp.Data, &stat); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse stat data")
		return
	}

	response.Success(c, upstreamStatResponse{
		Quota:    stat.Quota,
		QuotaUSD: float64(stat.Quota) / 500000.0,
		RPM:      stat.RPM,
		TPM:      stat.TPM,
	})
}

// GetLogs fetches detailed usage logs from upstream
func (h *UpstreamCostHandler) GetLogs(c *gin.Context) {
	var req upstreamLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	path := fmt.Sprintf("/api/log/self/?p=%d&page_size=%d", req.Page, req.PageSize)
	if req.Model != "" {
		path += "&model_name=" + req.Model
	}
	if req.TokenID > 0 {
		path += fmt.Sprintf("&token_id=%d", req.TokenID)
	}

	body, statusCode, err := h.doUpstreamGet(req.BaseURL, path, req.AccessToken, req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to fetch upstream logs: "+err.Error())
		return
	}
	if statusCode != http.StatusOK {
		var errResp newAPIResponse
		_ = json.Unmarshal(body, &errResp)
		response.Error(c, http.StatusBadGateway, fmt.Sprintf("Upstream error (%d): %s", statusCode, errResp.Message))
		return
	}

	var apiResp newAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse upstream response")
		return
	}

	var logData newAPILogListData
	if err := json.Unmarshal(apiResp.Data, &logData); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse log data")
		return
	}

	items := make([]upstreamLogItem, 0, len(logData.Items))
	for _, item := range logData.Items {
		items = append(items, upstreamLogItem{
			ID:               item.ID,
			CreatedAt:        item.CreatedAt,
			ModelName:        item.ModelName,
			Quota:            item.Quota,
			PromptTokens:     item.PromptTokens,
			CompletionTokens: item.CompletionTokens,
			UseTime:          item.UseTime,
			IsStream:         item.IsStream,
			TokenName:        item.TokenName,
			Channel:          item.Channel,
			Group:            item.Group,
			QuotaUSD:         float64(item.Quota) / 500000.0,
		})
	}

	response.Success(c, upstreamLogsResponse{
		Page:     logData.Page,
		PageSize: logData.PageSize,
		Total:    logData.Total,
		Items:    items,
	})
}

// TestConnection tests if the upstream connection is valid
func (h *UpstreamCostHandler) TestConnection(c *gin.Context) {
	var req upstreamProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	body, statusCode, err := h.doUpstreamGet(req.BaseURL, "/api/user/self/", req.AccessToken, req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Connection failed: "+err.Error())
		return
	}
	if statusCode != http.StatusOK {
		var errResp newAPIResponse
		_ = json.Unmarshal(body, &errResp)
		response.Error(c, http.StatusBadGateway, fmt.Sprintf("Auth failed (%d): %s", statusCode, errResp.Message))
		return
	}

	var apiResp newAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse response")
		return
	}

	var user newAPIUserInfo
	_ = json.Unmarshal(apiResp.Data, &user)

	response.Success(c, gin.H{
		"connected":    true,
		"display_name": user.DisplayName,
		"username":     user.Username,
		"user_id":      user.ID,
	})
}
