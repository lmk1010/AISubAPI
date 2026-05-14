package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ssrf"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// UpstreamCostHandler handles upstream cost analysis proxy requests
type UpstreamCostHandler struct {
	client       *http.Client
	localService *service.UpstreamCostService
}

// NewUpstreamCostHandler creates a new upstream cost handler
func NewUpstreamCostHandler(localService *service.UpstreamCostService) *UpstreamCostHandler {
	return &UpstreamCostHandler{
		client:       ssrf.NewSafeHTTPClient(30 * time.Second),
		localService: localService,
	}
}

// --- Request types ---

type upstreamProxyRequest struct {
	ProviderType string `json:"provider_type" binding:"required,oneof=newapi sub2api"` // newapi or sub2api
	BaseURL      string `json:"base_url" binding:"required,url"`
	AccessToken  string `json:"access_token"` // required for newapi
	UserID       int64  `json:"user_id"`      // required for newapi
	Email        string `json:"email"`        // required for sub2api
	Password     string `json:"password"`     // required for sub2api
}

func (r *upstreamProxyRequest) validate() error {
	switch r.ProviderType {
	case "newapi":
		if r.AccessToken == "" {
			return fmt.Errorf("access_token is required for newapi")
		}
		if r.UserID == 0 {
			return fmt.Errorf("user_id is required for newapi")
		}
	case "sub2api":
		if r.Email == "" {
			return fmt.Errorf("email is required for sub2api")
		}
		if r.Password == "" {
			return fmt.Errorf("password is required for sub2api")
		}
	}
	return nil
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

// GetLocalSummary returns local usage-log-based cost aggregation by upstream quota pool.
func (h *UpstreamCostHandler) GetLocalSummary(c *gin.Context) {
	if h.localService == nil {
		response.Error(c, http.StatusInternalServerError, "Upstream cost service is not configured")
		return
	}
	startTime, endTime := parseTimeRange(c)
	summary, err := h.localService.GetLocalSummary(c.Request.Context(), startTime, endTime)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get local upstream cost summary: "+err.Error())
		return
	}
	response.Success(c, summary)
}

// --- Helper: do upstream request ---

func (h *UpstreamCostHandler) doUpstreamGet(baseURL, path, accessToken string, userID int64) ([]byte, int, error) {
	fullURL := strings.TrimSuffix(baseURL, "/") + path

	// SSRF protection: validate that the target URL does not resolve to internal networks
	if err := ssrf.ValidateURL(fullURL); err != nil {
		return nil, 0, fmt.Errorf("SSRF blocked: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
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

// --- Sub2API helpers ---

// sub2apiLogin logs into a remote Sub2API instance and returns the JWT access token.
func (h *UpstreamCostHandler) sub2apiLogin(baseURL, email, password string) (string, error) {
	loginURL := strings.TrimSuffix(baseURL, "/") + "/api/v1/auth/login"
	if err := ssrf.ValidateURL(loginURL); err != nil {
		return "", fmt.Errorf("SSRF blocked: %w", err)
	}

	payload, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req, err := http.NewRequest(http.MethodPost, loginURL, strings.NewReader(string(payload)))
	if err != nil {
		return "", fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read login response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed (%d): %s", resp.StatusCode, string(body))
	}

	var loginResp struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
		Data struct {
			AccessToken string `json:"access_token"`
			Requires2FA bool   `json:"requires_2fa"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("parse login response: %w", err)
	}
	if loginResp.Data.Requires2FA {
		return "", fmt.Errorf("upstream Sub2API account has 2FA enabled, not supported")
	}
	if loginResp.Data.AccessToken == "" {
		return "", fmt.Errorf("login returned empty token: %s", loginResp.Msg)
	}
	return loginResp.Data.AccessToken, nil
}

// doSub2APIGet performs authenticated GET to a remote Sub2API instance.
func (h *UpstreamCostHandler) doSub2APIGet(baseURL, path, jwt string) ([]byte, int, error) {
	fullURL := strings.TrimSuffix(baseURL, "/") + path
	if err := ssrf.ValidateURL(fullURL); err != nil {
		return nil, 0, fmt.Errorf("SSRF blocked: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return body, resp.StatusCode, nil
}

// sub2apiResponse is the standard wrapper for Sub2API JSON responses: { code, message, data }.
type sub2apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func parseSub2APIResponse(body []byte) (*sub2apiResponse, error) {
	var r sub2apiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetUserInfo fetches user info from upstream
func (h *UpstreamCostHandler) GetUserInfo(c *gin.Context) {
	var req upstreamProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := req.validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.ProviderType == "sub2api" {
		h.getSub2APIUserInfo(c, &req)
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

func (h *UpstreamCostHandler) getSub2APIUserInfo(c *gin.Context, req *upstreamProxyRequest) {
	jwt, err := h.sub2apiLogin(req.BaseURL, req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Sub2API login failed: "+err.Error())
		return
	}

	body, statusCode, err := h.doSub2APIGet(req.BaseURL, "/api/v1/user/profile", jwt)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to fetch Sub2API profile: "+err.Error())
		return
	}
	if statusCode != http.StatusOK {
		response.Error(c, http.StatusBadGateway, fmt.Sprintf("Sub2API profile error (%d)", statusCode))
		return
	}

	apiResp, err := parseSub2APIResponse(body)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse Sub2API response")
		return
	}

	var profile struct {
		ID       int64   `json:"id"`
		Email    string  `json:"email"`
		Username string  `json:"username"`
		Role     string  `json:"role"`
		Balance  float64 `json:"balance"`
		Status   string  `json:"status"`
	}
	if err := json.Unmarshal(apiResp.Data, &profile); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse Sub2API profile data")
		return
	}

	// Convert balance (USD float) to quota units (balance * 500000) for unified display
	quota := int64(profile.Balance * 500000)
	response.Success(c, upstreamUserInfoResponse{
		ID:           profile.ID,
		Username:     profile.Username,
		DisplayName:  profile.Username,
		Email:        profile.Email,
		Quota:        quota,
		UsedQuota:    0,
		RequestCount: 0,
		Group:        "",
		ProviderType: "sub2api",
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
	if err := req.validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.ProviderType == "sub2api" {
		h.getSub2APIStats(c, &req)
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

func (h *UpstreamCostHandler) getSub2APIStats(c *gin.Context, req *upstreamProxyRequest) {
	jwt, err := h.sub2apiLogin(req.BaseURL, req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Sub2API login failed: "+err.Error())
		return
	}

	body, statusCode, err := h.doSub2APIGet(req.BaseURL, "/api/v1/usage/dashboard/stats", jwt)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to fetch Sub2API stats: "+err.Error())
		return
	}
	if statusCode != http.StatusOK {
		response.Error(c, http.StatusBadGateway, fmt.Sprintf("Sub2API stats error (%d)", statusCode))
		return
	}

	apiResp, err := parseSub2APIResponse(body)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse Sub2API response")
		return
	}

	var stats struct {
		TotalActualCost float64 `json:"total_actual_cost"`
		RPM             int64   `json:"rpm"`
		TPM             int64   `json:"tpm"`
	}
	if err := json.Unmarshal(apiResp.Data, &stats); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse Sub2API stats data")
		return
	}

	// Sub2API actual_cost is already in USD; convert to quota units for consistency
	quota := int64(stats.TotalActualCost * 500000)
	response.Success(c, upstreamStatResponse{
		Quota:    quota,
		QuotaUSD: stats.TotalActualCost,
		RPM:      stats.RPM,
		TPM:      stats.TPM,
	})
}

// GetLogs fetches detailed usage logs from upstream
func (h *UpstreamCostHandler) GetLogs(c *gin.Context) {
	var req upstreamLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := req.validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	if req.ProviderType == "sub2api" {
		h.getSub2APILogs(c, &req)
		return
	}

	path := fmt.Sprintf("/api/log/self/?p=%d&page_size=%d", req.Page, req.PageSize)
	if req.Model != "" {
		path += "&model_name=" + url.QueryEscape(req.Model)
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

func (h *UpstreamCostHandler) getSub2APILogs(c *gin.Context, req *upstreamLogsRequest) {
	jwt, err := h.sub2apiLogin(req.BaseURL, req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Sub2API login failed: "+err.Error())
		return
	}

	path := fmt.Sprintf("/api/v1/usage?page=%d&page_size=%d", req.Page+1, req.PageSize) // Sub2API is 1-indexed
	if req.Model != "" {
		path += "&model=" + url.QueryEscape(req.Model)
	}

	body, statusCode, err := h.doSub2APIGet(req.BaseURL, path, jwt)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to fetch Sub2API logs: "+err.Error())
		return
	}
	if statusCode != http.StatusOK {
		response.Error(c, http.StatusBadGateway, fmt.Sprintf("Sub2API logs error (%d)", statusCode))
		return
	}

	apiResp, err := parseSub2APIResponse(body)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse Sub2API response")
		return
	}

	// Sub2API usage list response matches response.Paginated format:
	// { items: [...], total, page, page_size, pages }
	var logData struct {
		Items []struct {
			ID           int64   `json:"id"`
			CreatedAt    string  `json:"created_at"`
			Model        string  `json:"model"`
			InputTokens  int64   `json:"input_tokens"`
			OutputTokens int64   `json:"output_tokens"`
			TotalCost    float64 `json:"total_cost"`
			ActualCost   float64 `json:"actual_cost"`
			DurationMs   *int    `json:"duration_ms"`
			Stream       bool    `json:"stream"`
			APIKey       *struct {
				Name string `json:"name"`
			} `json:"api_key"`
			Group *struct {
				Name string `json:"name"`
			} `json:"group"`
		} `json:"items"`
		Total    int64 `json:"total"`
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
	}
	if err := json.Unmarshal(apiResp.Data, &logData); err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to parse Sub2API log data")
		return
	}

	items := make([]upstreamLogItem, 0, len(logData.Items))
	for _, item := range logData.Items {
		var ts int64
		if t, err := time.Parse(time.RFC3339, item.CreatedAt); err == nil {
			ts = t.Unix()
		}
		var durationSec int
		if item.DurationMs != nil {
			durationSec = *item.DurationMs / 1000
		}
		var keyName, groupName string
		if item.APIKey != nil {
			keyName = item.APIKey.Name
		}
		if item.Group != nil {
			groupName = item.Group.Name
		}
		items = append(items, upstreamLogItem{
			ID:               item.ID,
			CreatedAt:        ts,
			ModelName:        item.Model,
			Quota:            int64(item.ActualCost * 500000),
			PromptTokens:     item.InputTokens,
			CompletionTokens: item.OutputTokens,
			UseTime:          durationSec,
			IsStream:         item.Stream,
			TokenName:        keyName,
			Channel:          0,
			Group:            groupName,
			QuotaUSD:         item.ActualCost,
		})
	}

	response.Success(c, upstreamLogsResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
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
	if err := req.validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.ProviderType == "sub2api" {
		jwt, err := h.sub2apiLogin(req.BaseURL, req.Email, req.Password)
		if err != nil {
			response.Error(c, http.StatusBadGateway, "Sub2API login failed: "+err.Error())
			return
		}
		body, statusCode, err := h.doSub2APIGet(req.BaseURL, "/api/v1/user/profile", jwt)
		if err != nil {
			response.Error(c, http.StatusBadGateway, "Connection failed: "+err.Error())
			return
		}
		if statusCode != http.StatusOK {
			response.Error(c, http.StatusBadGateway, fmt.Sprintf("Profile fetch failed (%d)", statusCode))
			return
		}
		apiResp, _ := parseSub2APIResponse(body)
		var profile struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}
		if apiResp != nil {
			_ = json.Unmarshal(apiResp.Data, &profile)
		}
		response.Success(c, gin.H{
			"connected":    true,
			"display_name": profile.Username,
			"username":     profile.Username,
			"user_id":      profile.ID,
		})
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
