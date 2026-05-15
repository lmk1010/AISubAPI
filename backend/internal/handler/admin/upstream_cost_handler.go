package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ssrf"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const newAPIQuotaUnitsPerRMB = 500000.0

// UpstreamCostHandler handles upstream cost analysis proxy requests
type UpstreamCostHandler struct {
	client             *http.Client
	localService       *service.UpstreamCostService
	realSummaryCacheMu sync.RWMutex
	realSummaryCache   *upstreamRealSummaryResponse
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

type upstreamRealSummaryResponse struct {
	StartDate   string                    `json:"start_date"`
	EndDate     string                    `json:"end_date"`
	GeneratedAt time.Time                 `json:"generated_at"`
	Scope       string                    `json:"scope"`
	Totals      upstreamRealSummaryTotals `json:"totals"`
	Pools       []upstreamRealPoolSummary `json:"pools"`
}

type upstreamRealSummaryTotals struct {
	AccountCount                int     `json:"account_count"`
	ErrorCount                  int     `json:"error_count"`
	Requests                    int64   `json:"requests"`
	TotalTokens                 int64   `json:"total_tokens"`
	StandardCost                float64 `json:"standard_cost"`
	DownstreamRevenueRMB        float64 `json:"downstream_revenue_rmb"`
	DownstreamUsageQuota        float64 `json:"downstream_usage_quota"`
	BalanceRevenueRMB           float64 `json:"balance_revenue_rmb"`
	SubscriptionQuotaCost       float64 `json:"subscription_quota_cost"`
	SubscriptionRevenueRMB      float64 `json:"subscription_revenue_rmb"`
	LocalAccountCostRMB         float64 `json:"local_account_cost_rmb"`
	UpstreamUsedRMB             float64 `json:"upstream_used_rmb"`
	AllocatedUpstreamUsedRMB    float64 `json:"allocated_upstream_used_rmb"`
	UnallocatedUpstreamUsedRMB  float64 `json:"unallocated_upstream_used_rmb"`
	UpstreamRemainingRMB        float64 `json:"upstream_remaining_rmb"`
	UpstreamUsedUSD             float64 `json:"upstream_used_usd"`
	SuccessfulRechargeRMB       float64 `json:"successful_recharge_rmb"`
	ProfitRMB                   float64 `json:"profit_rmb"`
	UpstreamEffectiveRate       float64 `json:"upstream_effective_rate"`
	DownstreamEffectiveRate     float64 `json:"downstream_effective_rate"`
	WeightedUpstreamAccountRate float64 `json:"weighted_upstream_account_rate"`
}

type upstreamRealPoolSummary struct {
	PoolKey                     string                           `json:"pool_key"`
	PoolName                    string                           `json:"pool_name"`
	ProviderType                string                           `json:"provider_type"`
	BaseURL                     string                           `json:"base_url"`
	AccountCount                int                              `json:"account_count"`
	AccountIDs                  []int64                          `json:"account_ids"`
	Requests                    int64                            `json:"requests"`
	TotalTokens                 int64                            `json:"total_tokens"`
	StandardCost                float64                          `json:"standard_cost"`
	DownstreamRevenueRMB        float64                          `json:"downstream_revenue_rmb"`
	DownstreamUsageQuota        float64                          `json:"downstream_usage_quota"`
	BalanceRevenueRMB           float64                          `json:"balance_revenue_rmb"`
	SubscriptionQuotaCost       float64                          `json:"subscription_quota_cost"`
	SubscriptionRevenueRMB      float64                          `json:"subscription_revenue_rmb"`
	LocalAccountCostRMB         float64                          `json:"local_account_cost_rmb"`
	UpstreamUsedRMB             float64                          `json:"upstream_used_rmb"`
	AllocatedUpstreamUsedRMB    float64                          `json:"allocated_upstream_used_rmb"`
	UnallocatedUpstreamUsedRMB  float64                          `json:"unallocated_upstream_used_rmb"`
	UpstreamRemainingRMB        float64                          `json:"upstream_remaining_rmb"`
	UpstreamUsedUSD             float64                          `json:"upstream_used_usd"`
	SuccessfulRechargeRMB       float64                          `json:"successful_recharge_rmb"`
	ProfitRMB                   float64                          `json:"profit_rmb"`
	UpstreamEffectiveRate       float64                          `json:"upstream_effective_rate"`
	DownstreamEffectiveRate     float64                          `json:"downstream_effective_rate"`
	WeightedUpstreamAccountRate float64                          `json:"weighted_upstream_account_rate"`
	Status                      string                           `json:"status"`
	Errors                      []string                         `json:"errors"`
	Accounts                    []upstreamRealAccountCostSummary `json:"accounts"`
}

type upstreamRealAccountCostSummary struct {
	AccountID               int64                                `json:"account_id"`
	AccountName             string                               `json:"account_name"`
	Platform                string                               `json:"platform"`
	Status                  string                               `json:"status"`
	GroupIDs                []int64                              `json:"group_ids"`
	ProviderType            string                               `json:"provider_type"`
	BaseURL                 string                               `json:"base_url"`
	Requests                int64                                `json:"requests"`
	TotalTokens             int64                                `json:"total_tokens"`
	StandardCost            float64                              `json:"standard_cost"`
	DownstreamRevenueRMB    float64                              `json:"downstream_revenue_rmb"`
	DownstreamUsageQuota    float64                              `json:"downstream_usage_quota"`
	BalanceRevenueRMB       float64                              `json:"balance_revenue_rmb"`
	SubscriptionQuotaCost   float64                              `json:"subscription_quota_cost"`
	SubscriptionRevenueRMB  float64                              `json:"subscription_revenue_rmb"`
	LocalAccountCostRMB     float64                              `json:"local_account_cost_rmb"`
	UpstreamUsedRMB         float64                              `json:"upstream_used_rmb"`
	UpstreamRemainingRMB    float64                              `json:"upstream_remaining_rmb"`
	UpstreamUsedUSD         float64                              `json:"upstream_used_usd"`
	ProfitRMB               float64                              `json:"profit_rmb"`
	UpstreamConfiguredRate  float64                              `json:"upstream_configured_rate"`
	UpstreamEffectiveRate   float64                              `json:"upstream_effective_rate"`
	DownstreamEffectiveRate float64                              `json:"downstream_effective_rate"`
	Source                  string                               `json:"source"`
	TokenName               string                               `json:"token_name"`
	TokenHash               string                               `json:"token_hash"`
	RemoteStatus            string                               `json:"remote_status"`
	Error                   string                               `json:"error"`
	Groups                  []upstreamRealGroupCostSummary       `json:"groups"`
	Trend                   []service.UpstreamCostTrendPoint     `json:"trend"`
	Models                  []service.UpstreamCostModelBreakdown `json:"models"`
}

type upstreamRealGroupCostSummary struct {
	GroupID                  int64   `json:"group_id"`
	GroupName                string  `json:"group_name"`
	CurrentGroupRate         float64 `json:"current_group_rate"`
	Requests                 int64   `json:"requests"`
	TotalTokens              int64   `json:"total_tokens"`
	StandardCost             float64 `json:"standard_cost"`
	DownstreamRevenueRMB     float64 `json:"downstream_revenue_rmb"`
	DownstreamUsageQuota     float64 `json:"downstream_usage_quota"`
	BalanceRevenueRMB        float64 `json:"balance_revenue_rmb"`
	SubscriptionQuotaCost    float64 `json:"subscription_quota_cost"`
	SubscriptionRevenueRMB   float64 `json:"subscription_revenue_rmb"`
	LocalAccountCostRMB      float64 `json:"local_account_cost_rmb"`
	AllocatedUpstreamUsedRMB float64 `json:"allocated_upstream_used_rmb"`
	ProfitRMB                float64 `json:"profit_rmb"`
	DownstreamEffectiveRate  float64 `json:"downstream_effective_rate"`
}

type upstreamRealPoolAccumulator struct {
	upstreamRealPoolSummary
	accountUsedSeen map[string]bool
	balanceSeen     map[string]bool
	rechargeSeen    map[string]bool
	rateWeightSum   float64
}

type upstreamRemoteAccountSnapshot struct {
	IdentityKey           string
	BalanceQuota          int64
	UsedQuota             int64
	StatQuota             int64
	RequestCount          int64
	SuccessfulRechargeRMB float64
	RechargeKnown         bool
	Error                 string
}

func (s upstreamRemoteAccountSnapshot) effectiveUsedQuota() int64 {
	if s.StatQuota > s.UsedQuota {
		return s.StatQuota
	}
	return s.UsedQuota
}

type newAPITokenUsageData struct {
	Object         string `json:"object"`
	Name           string `json:"name"`
	TotalGranted   int64  `json:"total_granted"`
	TotalUsed      int64  `json:"total_used"`
	Available      int64  `json:"available"`
	TotalAvailable int64  `json:"total_available"`
	Unlimited      bool   `json:"unlimited"`
	UnlimitedQuota bool   `json:"unlimited_quota"`
	ExpiredTime    int64  `json:"expired_time"`
	ExpiresAt      int64  `json:"expires_at"`
}

type newAPITopupListData struct {
	Items []newAPITopupItem `json:"items"`
}

type newAPITopupItem struct {
	Amount float64 `json:"amount"`
	Money  float64 `json:"money"`
	Status any     `json:"status"`
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

// GetRealSummary returns a remote-first upstream cost summary.
//
// Remote upstream usage is intentionally not derived from local usage logs. Local logs are
// only used for our downstream revenue and request/token attribution by account.
func (h *UpstreamCostHandler) GetRealSummary(c *gin.Context) {
	if h.localService == nil {
		response.Error(c, http.StatusInternalServerError, "Upstream cost service is not configured")
		return
	}

	refresh := queryBool(c.Query("refresh"))
	if !refresh {
		if cached := h.cachedRealSummary(); cached != nil {
			response.Success(c, cached)
			return
		}
		response.Success(c, emptyRealSummaryResponse(time.Now()))
		return
	}

	now := time.Now()
	startTime := time.Unix(0, 0).UTC()
	endTime := now.Add(24 * time.Hour)
	configuredAccounts, err := h.localService.GetConfiguredAccountSummaries(c.Request.Context(), startTime, endTime)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get local account summaries: "+err.Error())
		return
	}

	identityCounts := make(map[string]int, len(configuredAccounts))
	for _, configured := range configuredAccounts {
		identityCounts[upstreamRemoteIdentityKey(configured)]++
	}

	snapshotCache := make(map[string]upstreamRemoteAccountSnapshot)
	pools := make(map[string]*upstreamRealPoolAccumulator)
	for _, configured := range configuredAccounts {
		cfg := configured.Config
		poolKey := upstreamRealPoolKey(cfg)
		pool := pools[poolKey]
		if pool == nil {
			pool = &upstreamRealPoolAccumulator{
				upstreamRealPoolSummary: upstreamRealPoolSummary{
					PoolKey:      poolKey,
					PoolName:     upstreamRealPoolName(cfg),
					ProviderType: cfg.ProviderType,
					BaseURL:      firstNonEmptyString(cfg.NormalizedBaseURL, cfg.BaseURL),
					AccountIDs:   []int64{},
					Errors:       []string{},
					Accounts:     []upstreamRealAccountCostSummary{},
					Status:       "ok",
				},
				accountUsedSeen: map[string]bool{},
				balanceSeen:     map[string]bool{},
				rechargeSeen:    map[string]bool{},
			}
			pools[poolKey] = pool
		}

		child := h.buildRealAccountCostSummary(c.Request.Context(), configured, identityCounts, snapshotCache)
		identityKey := upstreamRemoteIdentityKey(configured)
		if child.Error != "" {
			pool.Errors = append(pool.Errors, fmt.Sprintf("%s: %s", child.AccountName, child.Error))
		}
		if child.RemoteStatus != "ok" {
			pool.Status = "partial"
		}

		pool.AccountIDs = append(pool.AccountIDs, child.AccountID)
		pool.Accounts = append(pool.Accounts, child)
		pool.Requests += child.Requests
		pool.TotalTokens += child.TotalTokens
		pool.StandardCost += child.StandardCost
		pool.DownstreamRevenueRMB += child.DownstreamRevenueRMB
		pool.DownstreamUsageQuota += child.DownstreamUsageQuota
		pool.BalanceRevenueRMB += child.BalanceRevenueRMB
		pool.SubscriptionQuotaCost += child.SubscriptionQuotaCost
		pool.SubscriptionRevenueRMB += child.SubscriptionRevenueRMB
		pool.LocalAccountCostRMB += child.LocalAccountCostRMB
		pool.AllocatedUpstreamUsedRMB += child.UpstreamUsedRMB
		pool.rateWeightSum += child.UpstreamConfiguredRate * child.StandardCost

		if snapshot, ok := snapshotCache[identityKey]; ok && snapshot.Error == "" {
			if !pool.accountUsedSeen[identityKey] {
				pool.UpstreamUsedRMB += quotaToMoney(snapshot.effectiveUsedQuota())
				pool.accountUsedSeen[identityKey] = true
			}
			if !pool.balanceSeen[identityKey] {
				pool.UpstreamRemainingRMB += quotaToMoney(snapshot.BalanceQuota)
				pool.balanceSeen[identityKey] = true
			}
			if snapshot.RechargeKnown && !pool.rechargeSeen[identityKey] {
				pool.SuccessfulRechargeRMB += snapshot.SuccessfulRechargeRMB
				pool.rechargeSeen[identityKey] = true
			}
		}
	}

	out := upstreamRealSummaryResponse{
		StartDate:   startTime.Format("2006-01-02"),
		EndDate:     now.Format("2006-01-02"),
		GeneratedAt: now,
		Scope:       "remote_upstream_lifetime_and_local_lifetime",
		Pools:       make([]upstreamRealPoolSummary, 0, len(pools)),
	}

	for _, pool := range pools {
		pool.AccountCount = len(pool.Accounts)
		sort.Slice(pool.AccountIDs, func(i, j int) bool { return pool.AccountIDs[i] < pool.AccountIDs[j] })
		sort.Slice(pool.Accounts, func(i, j int) bool {
			if pool.Accounts[i].UpstreamUsedRMB == pool.Accounts[j].UpstreamUsedRMB {
				return pool.Accounts[i].AccountName < pool.Accounts[j].AccountName
			}
			return pool.Accounts[i].UpstreamUsedRMB > pool.Accounts[j].UpstreamUsedRMB
		})
		if pool.UpstreamUsedRMB <= 0 && pool.AllocatedUpstreamUsedRMB > 0 {
			pool.UpstreamUsedRMB = pool.AllocatedUpstreamUsedRMB
			if pool.Status == "ok" {
				pool.Status = "partial"
			}
		}
		if pool.UpstreamUsedRMB > pool.AllocatedUpstreamUsedRMB {
			pool.UnallocatedUpstreamUsedRMB = pool.UpstreamUsedRMB - pool.AllocatedUpstreamUsedRMB
		}
		if pool.Status == "ok" && pool.UnallocatedUpstreamUsedRMB > 0.000001 {
			pool.Status = "partial"
		}
		if pool.Status == "" {
			pool.Status = "ok"
		}
		pool.UpstreamUsedUSD = pool.UpstreamUsedRMB
		pool.ProfitRMB = pool.DownstreamRevenueRMB - pool.UpstreamUsedRMB
		pool.WeightedUpstreamAccountRate = safeRatio(pool.rateWeightSum, pool.StandardCost)
		pool.UpstreamEffectiveRate = pool.WeightedUpstreamAccountRate
		pool.DownstreamEffectiveRate = safeRatio(pool.DownstreamUsageQuota, pool.StandardCost)

		out.Totals.AccountCount += pool.AccountCount
		out.Totals.ErrorCount += len(pool.Errors)
		out.Totals.Requests += pool.Requests
		out.Totals.TotalTokens += pool.TotalTokens
		out.Totals.StandardCost += pool.StandardCost
		out.Totals.DownstreamRevenueRMB += pool.DownstreamRevenueRMB
		out.Totals.DownstreamUsageQuota += pool.DownstreamUsageQuota
		out.Totals.BalanceRevenueRMB += pool.BalanceRevenueRMB
		out.Totals.SubscriptionQuotaCost += pool.SubscriptionQuotaCost
		out.Totals.SubscriptionRevenueRMB += pool.SubscriptionRevenueRMB
		out.Totals.LocalAccountCostRMB += pool.LocalAccountCostRMB
		out.Totals.UpstreamUsedRMB += pool.UpstreamUsedRMB
		out.Totals.AllocatedUpstreamUsedRMB += pool.AllocatedUpstreamUsedRMB
		out.Totals.UnallocatedUpstreamUsedRMB += pool.UnallocatedUpstreamUsedRMB
		out.Totals.UpstreamRemainingRMB += pool.UpstreamRemainingRMB
		out.Totals.UpstreamUsedUSD += pool.UpstreamUsedUSD
		out.Totals.SuccessfulRechargeRMB += pool.SuccessfulRechargeRMB
		out.Totals.ProfitRMB += pool.ProfitRMB

		out.Pools = append(out.Pools, pool.upstreamRealPoolSummary)
	}
	var totalRateWeight float64
	for _, pool := range pools {
		totalRateWeight += pool.rateWeightSum
	}
	out.Totals.WeightedUpstreamAccountRate = safeRatio(totalRateWeight, out.Totals.StandardCost)
	out.Totals.UpstreamEffectiveRate = out.Totals.WeightedUpstreamAccountRate
	out.Totals.DownstreamEffectiveRate = safeRatio(out.Totals.DownstreamUsageQuota, out.Totals.StandardCost)

	sort.Slice(out.Pools, func(i, j int) bool {
		if out.Pools[i].UpstreamUsedRMB == out.Pools[j].UpstreamUsedRMB {
			return out.Pools[i].PoolName < out.Pools[j].PoolName
		}
		return out.Pools[i].UpstreamUsedRMB > out.Pools[j].UpstreamUsedRMB
	})

	h.setCachedRealSummary(&out)
	response.Success(c, out)
}

func (h *UpstreamCostHandler) cachedRealSummary() *upstreamRealSummaryResponse {
	h.realSummaryCacheMu.RLock()
	defer h.realSummaryCacheMu.RUnlock()
	return h.realSummaryCache
}

func (h *UpstreamCostHandler) setCachedRealSummary(summary *upstreamRealSummaryResponse) {
	if summary == nil {
		return
	}
	h.realSummaryCacheMu.Lock()
	defer h.realSummaryCacheMu.Unlock()
	h.realSummaryCache = summary
}

func emptyRealSummaryResponse(now time.Time) upstreamRealSummaryResponse {
	return upstreamRealSummaryResponse{
		StartDate:   time.Unix(0, 0).UTC().Format("2006-01-02"),
		EndDate:     now.Format("2006-01-02"),
		GeneratedAt: now,
		Scope:       "remote_upstream_cache_empty",
		Pools:       []upstreamRealPoolSummary{},
	}
}

func (h *UpstreamCostHandler) buildRealAccountCostSummary(
	ctx context.Context,
	configured service.UpstreamCostConfiguredAccount,
	identityCounts map[string]int,
	snapshotCache map[string]upstreamRemoteAccountSnapshot,
) upstreamRealAccountCostSummary {
	account := configured.Account
	cfg := configured.Config
	local := configured.Summary
	child := upstreamRealAccountCostSummary{
		AccountID:               local.AccountID,
		AccountName:             local.AccountName,
		Platform:                local.Platform,
		Status:                  local.Status,
		GroupIDs:                append([]int64(nil), local.GroupIDs...),
		ProviderType:            cfg.ProviderType,
		BaseURL:                 firstNonEmptyString(cfg.NormalizedBaseURL, cfg.BaseURL),
		Requests:                local.Requests,
		TotalTokens:             local.TotalTokens,
		StandardCost:            local.StandardCost,
		DownstreamRevenueRMB:    local.DownstreamRevenueRMB,
		DownstreamUsageQuota:    local.UserCost,
		BalanceRevenueRMB:       local.BalanceRevenueRMB,
		SubscriptionQuotaCost:   local.SubscriptionQuotaCost,
		SubscriptionRevenueRMB:  local.SubscriptionRevenueRMB,
		LocalAccountCostRMB:     local.UpstreamCost,
		UpstreamConfiguredRate:  local.RateMultiplier,
		DownstreamEffectiveRate: safeRatio(local.UserCost, local.StandardCost),
		Source:                  "unknown",
		RemoteStatus:            "ok",
		Trend:                   local.Trend,
		Models:                  local.Models,
	}

	identityKey := upstreamRemoteIdentityKey(configured)
	snapshot, ok := snapshotCache[identityKey]
	if !ok {
		switch cfg.ProviderType {
		case "newapi":
			snapshot = h.fetchNewAPIAccountSnapshot(ctx, cfg)
		case "sub2api":
			snapshot = h.fetchSub2APIAccountSnapshot(ctx, cfg)
		default:
			snapshot = upstreamRemoteAccountSnapshot{IdentityKey: identityKey, Error: "unsupported upstream provider: " + cfg.ProviderType}
		}
		if snapshot.IdentityKey == "" {
			snapshot.IdentityKey = identityKey
		}
		snapshotCache[identityKey] = snapshot
	}

	if snapshot.Error != "" {
		child.RemoteStatus = "error"
		child.Error = snapshot.Error
	}
	child.UpstreamRemainingRMB = quotaToMoney(snapshot.BalanceQuota)

	apiKey := strings.TrimSpace(stringFromHandlerAny(account.Credentials["api_key"]))
	if cfg.ProviderType == "newapi" && apiKey != "" {
		child.TokenHash = shortSecretHash(apiKey)
		tokenUsage, err := h.fetchNewAPITokenUsage(ctx, cfg.BaseURL, apiKey)
		if err != nil {
			child.RemoteStatus = "partial"
			child.Error = appendError(child.Error, "API key usage failed: "+err.Error())
			child.Source = "token_error"
		} else {
			child.TokenName = tokenUsage.Name
			child.Source = "token"
			if snapshot.Error != "" {
				child.RemoteStatus = "partial"
			}
			child.UpstreamUsedRMB = quotaToMoney(tokenUsage.TotalUsed)
			child.UpstreamUsedUSD = child.UpstreamUsedRMB
			if snapshot.Error != "" && !tokenUsage.Unlimited && tokenUsage.Available >= 0 {
				child.UpstreamRemainingRMB = quotaToMoney(tokenUsage.Available)
			}
		}
	} else if snapshot.Error == "" && identityCounts[identityKey] == 1 {
		child.Source = "account_total"
		child.UpstreamUsedRMB = quotaToMoney(snapshot.effectiveUsedQuota())
		child.UpstreamUsedUSD = child.UpstreamUsedRMB
	} else if snapshot.Error == "" {
		child.RemoteStatus = "partial"
		child.Source = "unallocated"
		child.Error = appendError(child.Error, "missing API key usage; this upstream account is shared and cannot be split safely")
	}

	child.ProfitRMB = child.DownstreamRevenueRMB - child.UpstreamUsedRMB
	child.UpstreamEffectiveRate = child.UpstreamConfiguredRate
	child.Groups = buildRealGroupCostSummaries(local.Groups, child.StandardCost, child.UpstreamUsedRMB)
	return child
}

func buildRealGroupCostSummaries(groups []service.UpstreamCostGroupBreakdown, accountStandardCost, accountUpstreamUsedRMB float64) []upstreamRealGroupCostSummary {
	if len(groups) == 0 {
		return []upstreamRealGroupCostSummary{}
	}
	out := make([]upstreamRealGroupCostSummary, 0, len(groups))
	for _, group := range groups {
		allocatedUpstream := 0.0
		if accountStandardCost > 0 && accountUpstreamUsedRMB > 0 {
			allocatedUpstream = accountUpstreamUsedRMB * group.StandardCost / accountStandardCost
		}
		out = append(out, upstreamRealGroupCostSummary{
			GroupID:                  group.GroupID,
			GroupName:                group.GroupName,
			CurrentGroupRate:         group.CurrentGroupRate,
			Requests:                 group.Requests,
			TotalTokens:              group.TotalTokens,
			StandardCost:             group.StandardCost,
			DownstreamRevenueRMB:     group.DownstreamRevenueRMB,
			DownstreamUsageQuota:     group.UserCost,
			BalanceRevenueRMB:        group.BalanceRevenueRMB,
			SubscriptionQuotaCost:    group.SubscriptionQuotaCost,
			SubscriptionRevenueRMB:   group.SubscriptionRevenueRMB,
			LocalAccountCostRMB:      group.UpstreamCost,
			AllocatedUpstreamUsedRMB: allocatedUpstream,
			ProfitRMB:                group.DownstreamRevenueRMB - allocatedUpstream,
			DownstreamEffectiveRate:  safeRatio(group.UserCost, group.StandardCost),
		})
	}
	return out
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

func (h *UpstreamCostHandler) doUpstreamGetWithContext(ctx context.Context, baseURL, path string, headers map[string]string) ([]byte, int, error) {
	fullURL := strings.TrimSuffix(baseURL, "/") + path
	if err := ssrf.ValidateURL(fullURL); err != nil {
		return nil, 0, fmt.Errorf("SSRF blocked: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	for key, value := range headers {
		if strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return body, resp.StatusCode, nil
}

func (h *UpstreamCostHandler) fetchNewAPIAccountSnapshot(ctx context.Context, cfg service.UpstreamCostProviderConfig) upstreamRemoteAccountSnapshot {
	snapshot := upstreamRemoteAccountSnapshot{IdentityKey: upstreamConfigIdentityKey(cfg)}
	if strings.TrimSpace(cfg.AccessToken) == "" || cfg.UserID == 0 {
		snapshot.Error = "missing New-API access token or user id"
		return snapshot
	}

	headers := map[string]string{
		"Authorization": "Bearer " + cfg.AccessToken,
		"New-Api-User":  fmt.Sprintf("%d", cfg.UserID),
	}
	body, statusCode, err := h.doUpstreamGetWithContext(ctx, cfg.BaseURL, "/api/user/self/", headers)
	if err != nil {
		snapshot.Error = err.Error()
		return snapshot
	}
	if statusCode != http.StatusOK {
		snapshot.Error = fmt.Sprintf("user info failed (%d): %s", statusCode, newAPIErrorMessage(body))
		return snapshot
	}
	var user newAPIUserInfo
	if err := parseNewAPIData(body, &user); err != nil {
		snapshot.Error = "parse user info: " + err.Error()
		return snapshot
	}
	snapshot.BalanceQuota = user.Quota
	snapshot.UsedQuota = user.UsedQuota
	snapshot.RequestCount = user.RequestCount

	if body, statusCode, err := h.doUpstreamGetWithContext(ctx, cfg.BaseURL, "/api/log/self/stat", headers); err == nil && statusCode == http.StatusOK {
		var stat newAPIStatData
		if err := parseNewAPIData(body, &stat); err == nil {
			snapshot.StatQuota = stat.Quota
		}
	}
	if rechargeRMB, ok := h.fetchNewAPITopupRMB(ctx, cfg, headers); ok {
		snapshot.SuccessfulRechargeRMB = rechargeRMB
		snapshot.RechargeKnown = true
	}
	return snapshot
}

func (h *UpstreamCostHandler) fetchNewAPITokenUsage(ctx context.Context, baseURL, apiKey string) (newAPITokenUsageData, error) {
	if strings.TrimSpace(apiKey) == "" {
		return newAPITokenUsageData{}, fmt.Errorf("missing API key")
	}
	body, statusCode, err := h.doUpstreamGetWithContext(ctx, baseURL, "/api/usage/token", map[string]string{
		"Authorization": "Bearer " + apiKey,
	})
	if err != nil {
		return newAPITokenUsageData{}, err
	}
	if statusCode != http.StatusOK {
		return newAPITokenUsageData{}, fmt.Errorf("token usage failed (%d): %s", statusCode, newAPIErrorMessage(body))
	}
	var tokenUsage newAPITokenUsageData
	if err := parseNewAPIData(body, &tokenUsage); err != nil {
		return newAPITokenUsageData{}, fmt.Errorf("parse token usage: %w", err)
	}
	tokenUsage.normalize()
	return tokenUsage, nil
}

func (d *newAPITokenUsageData) normalize() {
	if d.Available == 0 && d.TotalAvailable != 0 {
		d.Available = d.TotalAvailable
	}
	if d.UnlimitedQuota {
		d.Unlimited = true
	}
	if d.ExpiredTime == 0 && d.ExpiresAt != 0 {
		d.ExpiredTime = d.ExpiresAt
	}
}

func (h *UpstreamCostHandler) fetchNewAPITopupRMB(ctx context.Context, cfg service.UpstreamCostProviderConfig, headers map[string]string) (float64, bool) {
	body, statusCode, err := h.doUpstreamGetWithContext(ctx, cfg.BaseURL, "/api/user/topup/self", headers)
	if err != nil || statusCode != http.StatusOK {
		return 0, false
	}
	var topups newAPITopupListData
	if err := parseNewAPIData(body, &topups); err != nil {
		return 0, false
	}
	var total float64
	for _, item := range topups.Items {
		if !newAPITopupSuccess(item.Status) {
			continue
		}
		amount := item.Money
		if amount <= 0 {
			amount = item.Amount
		}
		total += amount
	}
	return total, true
}

func (h *UpstreamCostHandler) fetchSub2APIAccountSnapshot(ctx context.Context, cfg service.UpstreamCostProviderConfig) upstreamRemoteAccountSnapshot {
	_ = ctx
	snapshot := upstreamRemoteAccountSnapshot{IdentityKey: upstreamConfigIdentityKey(cfg)}
	if cfg.Email == "" || cfg.Password == "" {
		snapshot.Error = "missing Sub2API email or password"
		return snapshot
	}

	jwt, err := h.sub2apiLogin(cfg.BaseURL, cfg.Email, cfg.Password)
	if err != nil {
		snapshot.Error = "login failed: " + err.Error()
		return snapshot
	}

	body, statusCode, err := h.doSub2APIGet(cfg.BaseURL, "/api/v1/user/profile", jwt)
	if err != nil {
		snapshot.Error = "profile failed: " + err.Error()
		return snapshot
	}
	if statusCode != http.StatusOK {
		snapshot.Error = fmt.Sprintf("profile failed (%d)", statusCode)
		return snapshot
	}
	apiResp, err := parseSub2APIResponse(body)
	if err != nil {
		snapshot.Error = "parse profile: " + err.Error()
		return snapshot
	}
	var profile struct {
		Balance float64 `json:"balance"`
	}
	if err := json.Unmarshal(apiResp.Data, &profile); err != nil {
		snapshot.Error = "parse profile data: " + err.Error()
		return snapshot
	}
	snapshot.BalanceQuota = int64(profile.Balance * newAPIQuotaUnitsPerRMB)

	body, statusCode, err = h.doSub2APIGet(cfg.BaseURL, "/api/v1/usage/dashboard/stats", jwt)
	if err == nil && statusCode == http.StatusOK {
		if apiResp, err := parseSub2APIResponse(body); err == nil {
			var stats struct {
				TotalActualCost float64 `json:"total_actual_cost"`
			}
			if err := json.Unmarshal(apiResp.Data, &stats); err == nil {
				snapshot.StatQuota = int64(stats.TotalActualCost * newAPIQuotaUnitsPerRMB)
			}
		}
	}
	return snapshot
}

func parseNewAPIData(body []byte, out any) error {
	var apiResp struct {
		Success *bool           `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err == nil {
		if apiResp.Success != nil && !*apiResp.Success {
			message := strings.TrimSpace(apiResp.Message)
			if message == "" {
				message = "upstream returned success=false"
			}
			return fmt.Errorf("%s", message)
		}
		if apiResp.Data != nil {
			if len(apiResp.Data) == 0 || string(apiResp.Data) == "null" {
				return nil
			}
			return json.Unmarshal(apiResp.Data, out)
		}
		if apiResp.Success != nil {
			return nil
		}
	}
	return json.Unmarshal(body, out)
}

func newAPIErrorMessage(body []byte) string {
	var apiResp newAPIResponse
	if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Message != "" {
		return apiResp.Message
	}
	var generic struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &generic); err == nil {
		if generic.Error.Message != "" {
			return generic.Error.Message
		}
		if generic.Message != "" {
			return generic.Message
		}
	}
	return "upstream returned non-OK response"
}

func newAPITopupSuccess(raw any) bool {
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "success", "succeeded", "paid", "completed", "complete", "done", "1":
			return true
		default:
			return false
		}
	case float64:
		return int(value) == 1
	case int:
		return value == 1
	case int64:
		return value == 1
	case json.Number:
		n, _ := value.Int64()
		return n == 1
	default:
		return false
	}
}

func upstreamRealPoolKey(cfg service.UpstreamCostProviderConfig) string {
	baseURL := firstNonEmptyString(cfg.NormalizedBaseURL, normalizeHandlerURL(cfg.BaseURL))
	return cfg.ProviderType + "|" + baseURL
}

func upstreamRealPoolName(cfg service.UpstreamCostProviderConfig) string {
	baseURL := firstNonEmptyString(cfg.NormalizedBaseURL, normalizeHandlerURL(cfg.BaseURL))
	if parsed, err := url.Parse(baseURL); err == nil && parsed.Host != "" {
		if parsed.Path != "" && parsed.Path != "/" {
			return parsed.Host + strings.TrimRight(parsed.Path, "/")
		}
		return parsed.Host
	}
	return firstNonEmptyString(cfg.PoolName, baseURL)
}

func upstreamRemoteIdentityKey(configured service.UpstreamCostConfiguredAccount) string {
	return upstreamConfigIdentityKey(configured.Config)
}

func upstreamConfigIdentityKey(cfg service.UpstreamCostProviderConfig) string {
	baseURL := firstNonEmptyString(cfg.NormalizedBaseURL, normalizeHandlerURL(cfg.BaseURL))
	switch cfg.ProviderType {
	case "newapi":
		return fmt.Sprintf("newapi|%s|uid:%d|token:%s", baseURL, cfg.UserID, shortSecretHash(cfg.AccessToken))
	case "sub2api":
		return fmt.Sprintf("sub2api|%s|email:%s", baseURL, strings.ToLower(strings.TrimSpace(cfg.Email)))
	default:
		return fmt.Sprintf("%s|%s", cfg.ProviderType, baseURL)
	}
}

func normalizeHandlerURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return strings.TrimRight(strings.ToLower(raw), "/")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return strings.TrimRight(parsed.String(), "/")
}

func quotaToMoney(quota int64) float64 {
	return float64(quota) / newAPIQuotaUnitsPerRMB
}

func safeRatio(numerator, denominator float64) float64 {
	if denominator <= 0 {
		return 0
	}
	return numerator / denominator
}

func queryBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func appendError(existing, next string) string {
	if strings.TrimSpace(next) == "" {
		return existing
	}
	if strings.TrimSpace(existing) == "" {
		return next
	}
	return existing + "; " + next
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func shortSecretHash(secret string) string {
	if strings.TrimSpace(secret) == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])[:12]
}

func stringFromHandlerAny(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case fmt.Stringer:
		return value.String()
	case int:
		return strconv.Itoa(value)
	case int64:
		return strconv.FormatInt(value, 10)
	case float64:
		if value == float64(int64(value)) {
			return strconv.FormatInt(int64(value), 10)
		}
		return strconv.FormatFloat(value, 'f', -1, 64)
	case json.Number:
		return value.String()
	default:
		return ""
	}
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
