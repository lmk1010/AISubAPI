package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/lib/pq"
)

const upstreamCostPageSize = 1000

type upstreamCostAccountLister interface {
	List(ctx context.Context, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error)
}

// UpstreamCostService builds local, log-based cost summaries for upstream quota pools.
type UpstreamCostService struct {
	db          *sql.DB
	accountRepo upstreamCostAccountLister
}

func NewUpstreamCostService(db *sql.DB, accountRepo AccountRepository) *UpstreamCostService {
	return &UpstreamCostService{db: db, accountRepo: accountRepo}
}

type UpstreamCostLocalSummary struct {
	StartDate string                    `json:"start_date"`
	EndDate   string                    `json:"end_date"`
	Totals    UpstreamCostTotals        `json:"totals"`
	Pools     []UpstreamCostPoolSummary `json:"pools"`
}

type UpstreamCostTotals struct {
	Requests     int64   `json:"requests"`
	TotalTokens  int64   `json:"total_tokens"`
	StandardCost float64 `json:"standard_cost"`
	UpstreamCost float64 `json:"upstream_cost"`
	UserCost     float64 `json:"user_cost"`
	Profit       float64 `json:"profit"`
}

type UpstreamCostPoolSummary struct {
	PoolKey      string                       `json:"pool_key"`
	PoolName     string                       `json:"pool_name"`
	ProviderType string                       `json:"provider_type"`
	BaseURL      string                       `json:"base_url"`
	AccountCount int                          `json:"account_count"`
	AccountIDs   []int64                      `json:"account_ids"`
	Requests     int64                        `json:"requests"`
	TotalTokens  int64                        `json:"total_tokens"`
	StandardCost float64                      `json:"standard_cost"`
	UpstreamCost float64                      `json:"upstream_cost"`
	UserCost     float64                      `json:"user_cost"`
	Profit       float64                      `json:"profit"`
	Trend        []UpstreamCostTrendPoint     `json:"trend"`
	Models       []UpstreamCostModelBreakdown `json:"models"`
	Accounts     []UpstreamCostAccountSummary `json:"accounts"`
}

type UpstreamCostAccountSummary struct {
	AccountID      int64                        `json:"account_id"`
	AccountName    string                       `json:"account_name"`
	Platform       string                       `json:"platform"`
	Status         string                       `json:"status"`
	GroupIDs       []int64                      `json:"group_ids"`
	RateMultiplier float64                      `json:"rate_multiplier"`
	ProviderType   string                       `json:"provider_type"`
	BaseURL        string                       `json:"base_url"`
	PoolKey        string                       `json:"pool_key"`
	PoolName       string                       `json:"pool_name"`
	Requests       int64                        `json:"requests"`
	TotalTokens    int64                        `json:"total_tokens"`
	StandardCost   float64                      `json:"standard_cost"`
	UpstreamCost   float64                      `json:"upstream_cost"`
	UserCost       float64                      `json:"user_cost"`
	Profit         float64                      `json:"profit"`
	Trend          []UpstreamCostTrendPoint     `json:"trend"`
	Models         []UpstreamCostModelBreakdown `json:"models"`
}

type UpstreamCostTrendPoint struct {
	Date         string  `json:"date"`
	Requests     int64   `json:"requests"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	CacheTokens  int64   `json:"cache_tokens"`
	TotalTokens  int64   `json:"total_tokens"`
	StandardCost float64 `json:"standard_cost"`
	UpstreamCost float64 `json:"upstream_cost"`
	UserCost     float64 `json:"user_cost"`
}

type UpstreamCostModelBreakdown struct {
	Model        string  `json:"model"`
	Requests     int64   `json:"requests"`
	TotalTokens  int64   `json:"total_tokens"`
	StandardCost float64 `json:"standard_cost"`
	UpstreamCost float64 `json:"upstream_cost"`
	UserCost     float64 `json:"user_cost"`
}

type upstreamProviderCostConfig struct {
	ProviderType string
	BaseURL      string
	PoolKey      string
	PoolName     string
}

type upstreamCostStats struct {
	Requests     int64
	InputTokens  int64
	OutputTokens int64
	CacheTokens  int64
	TotalTokens  int64
	StandardCost float64
	UpstreamCost float64
	UserCost     float64
}

// GetLocalSummary aggregates local usage logs by upstream quota pool.
//
// The upstream cost is account-cost perspective:
//
//	SUM(COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1))
//
// UserCost is downstream billing perspective:
//
//	SUM(actual_cost)
func (s *UpstreamCostService) GetLocalSummary(ctx context.Context, startTime, endTime time.Time) (*UpstreamCostLocalSummary, error) {
	if s == nil || s.db == nil || s.accountRepo == nil {
		return nil, errors.New("upstream cost service is not configured")
	}
	if !endTime.After(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	accounts, err := s.listConfiguredAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	if len(accounts) == 0 {
		return &UpstreamCostLocalSummary{
			StartDate: startTime.Format("2006-01-02"),
			EndDate:   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
			Pools:     []UpstreamCostPoolSummary{},
		}, nil
	}

	accountIDs := make([]int64, 0, len(accounts))
	accountByID := make(map[int64]Account, len(accounts))
	cfgByID := make(map[int64]upstreamProviderCostConfig, len(accounts))
	for _, account := range accounts {
		cfg, ok := upstreamCostProviderConfig(account)
		if !ok {
			continue
		}
		accountIDs = append(accountIDs, account.ID)
		accountByID[account.ID] = account
		cfgByID[account.ID] = cfg
	}
	if len(accountIDs) == 0 {
		return &UpstreamCostLocalSummary{
			StartDate: startTime.Format("2006-01-02"),
			EndDate:   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
			Pools:     []UpstreamCostPoolSummary{},
		}, nil
	}

	statsByAccount, err := s.queryAccountStats(ctx, startTime, endTime, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("query account stats: %w", err)
	}
	trendByAccount, err := s.queryAccountTrend(ctx, startTime, endTime, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("query account trend: %w", err)
	}
	modelsByAccount, err := s.queryAccountModels(ctx, startTime, endTime, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("query account models: %w", err)
	}

	pools := make(map[string]*UpstreamCostPoolSummary)
	for _, accountID := range accountIDs {
		account := accountByID[accountID]
		cfg := cfgByID[accountID]
		stats := statsByAccount[accountID]
		accountSummary := UpstreamCostAccountSummary{
			AccountID:      account.ID,
			AccountName:    account.Name,
			Platform:       account.Platform,
			Status:         account.Status,
			GroupIDs:       append([]int64(nil), account.GroupIDs...),
			RateMultiplier: account.BillingRateMultiplier(),
			ProviderType:   cfg.ProviderType,
			BaseURL:        cfg.BaseURL,
			PoolKey:        cfg.PoolKey,
			PoolName:       cfg.PoolName,
			Requests:       stats.Requests,
			TotalTokens:    stats.TotalTokens,
			StandardCost:   stats.StandardCost,
			UpstreamCost:   stats.UpstreamCost,
			UserCost:       stats.UserCost,
			Profit:         stats.UserCost - stats.UpstreamCost,
			Trend:          trendByAccount[accountID],
			Models:         modelsByAccount[accountID],
		}

		pool := pools[cfg.PoolKey]
		if pool == nil {
			pool = &UpstreamCostPoolSummary{
				PoolKey:      cfg.PoolKey,
				PoolName:     cfg.PoolName,
				ProviderType: cfg.ProviderType,
				BaseURL:      cfg.BaseURL,
				AccountIDs:   []int64{},
				Trend:        []UpstreamCostTrendPoint{},
				Models:       []UpstreamCostModelBreakdown{},
				Accounts:     []UpstreamCostAccountSummary{},
			}
			pools[cfg.PoolKey] = pool
		}
		pool.AccountIDs = append(pool.AccountIDs, account.ID)
		pool.Accounts = append(pool.Accounts, accountSummary)
		pool.Requests += accountSummary.Requests
		pool.TotalTokens += accountSummary.TotalTokens
		pool.StandardCost += accountSummary.StandardCost
		pool.UpstreamCost += accountSummary.UpstreamCost
		pool.UserCost += accountSummary.UserCost
		pool.Profit = pool.UserCost - pool.UpstreamCost
	}

	out := &UpstreamCostLocalSummary{
		StartDate: startTime.Format("2006-01-02"),
		EndDate:   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		Pools:     make([]UpstreamCostPoolSummary, 0, len(pools)),
	}
	for _, pool := range pools {
		pool.AccountCount = len(pool.Accounts)
		sort.Slice(pool.Accounts, func(i, j int) bool {
			if pool.Accounts[i].UpstreamCost == pool.Accounts[j].UpstreamCost {
				return pool.Accounts[i].AccountName < pool.Accounts[j].AccountName
			}
			return pool.Accounts[i].UpstreamCost > pool.Accounts[j].UpstreamCost
		})
		sort.Slice(pool.AccountIDs, func(i, j int) bool { return pool.AccountIDs[i] < pool.AccountIDs[j] })
		pool.Trend = mergeUpstreamCostTrend(pool.Accounts)
		pool.Models = mergeUpstreamCostModels(pool.Accounts)
		out.Pools = append(out.Pools, *pool)
		out.Totals.Requests += pool.Requests
		out.Totals.TotalTokens += pool.TotalTokens
		out.Totals.StandardCost += pool.StandardCost
		out.Totals.UpstreamCost += pool.UpstreamCost
		out.Totals.UserCost += pool.UserCost
	}
	out.Totals.Profit = out.Totals.UserCost - out.Totals.UpstreamCost
	sort.Slice(out.Pools, func(i, j int) bool {
		if out.Pools[i].UpstreamCost == out.Pools[j].UpstreamCost {
			return out.Pools[i].PoolName < out.Pools[j].PoolName
		}
		return out.Pools[i].UpstreamCost > out.Pools[j].UpstreamCost
	})
	return out, nil
}

func (s *UpstreamCostService) listConfiguredAccounts(ctx context.Context) ([]Account, error) {
	var out []Account
	for page := 1; ; page++ {
		items, result, err := s.accountRepo.List(ctx, pagination.PaginationParams{
			Page:      page,
			PageSize:  upstreamCostPageSize,
			SortBy:    "id",
			SortOrder: pagination.SortOrderAsc,
		})
		if err != nil {
			return nil, err
		}
		for _, account := range items {
			if _, ok := upstreamCostProviderConfig(account); ok {
				out = append(out, account)
			}
		}
		if result == nil || int64(page*upstreamCostPageSize) >= result.Total || len(items) < upstreamCostPageSize {
			break
		}
	}
	return out, nil
}

func upstreamCostProviderConfig(account Account) (upstreamProviderCostConfig, bool) {
	raw, ok := account.Extra["upstream_provider"]
	if !ok || raw == nil {
		return upstreamProviderCostConfig{}, false
	}
	cfg, ok := raw.(map[string]any)
	if !ok {
		return upstreamProviderCostConfig{}, false
	}
	baseURL := strings.TrimSpace(stringFromAny(cfg["base_url"]))
	if baseURL == "" {
		return upstreamProviderCostConfig{}, false
	}
	providerType := strings.ToLower(strings.TrimSpace(stringFromAny(cfg["type"])))
	if providerType == "" {
		providerType = strings.ToLower(strings.TrimSpace(stringFromAny(cfg["provider_type"])))
	}
	if providerType == "" {
		providerType = "newapi"
	}
	normalizedURL := normalizeUpstreamPoolURL(baseURL)
	poolKey := strings.TrimSpace(firstNonEmptyUpstreamCostString(
		stringFromAny(cfg["pool_key"]),
		stringFromAny(cfg["upstream_pool_key"]),
		stringFromAny(cfg["pool_id"]),
		stringFromAny(account.Extra["upstream_pool_key"]),
	))
	if poolKey == "" {
		poolKey = providerType + "|" + normalizedURL
	}
	poolName := strings.TrimSpace(firstNonEmptyUpstreamCostString(
		stringFromAny(cfg["pool_name"]),
		stringFromAny(cfg["upstream_pool_name"]),
		stringFromAny(account.Extra["upstream_pool_name"]),
	))
	if poolName == "" {
		poolName = upstreamPoolNameFromURL(normalizedURL)
	}
	return upstreamProviderCostConfig{
		ProviderType: providerType,
		BaseURL:      baseURL,
		PoolKey:      poolKey,
		PoolName:     poolName,
	}, true
}

func (s *UpstreamCostService) queryAccountStats(ctx context.Context, startTime, endTime time.Time, accountIDs []int64) (map[int64]upstreamCostStats, error) {
	query := `
		SELECT
			account_id,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COALESCE(SUM(cache_creation_tokens + cache_read_tokens), 0) AS cache_tokens,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) AS total_tokens,
			COALESCE(SUM(total_cost), 0) AS standard_cost,
			COALESCE(SUM(COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1)), 0) AS upstream_cost,
			COALESCE(SUM(actual_cost), 0) AS user_cost
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2 AND account_id = ANY($3)
		GROUP BY account_id
	`
	rows, err := s.db.QueryContext(ctx, query, startTime, endTime, pq.Array(accountIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]upstreamCostStats, len(accountIDs))
	for rows.Next() {
		var accountID int64
		var stats upstreamCostStats
		if err := rows.Scan(
			&accountID,
			&stats.Requests,
			&stats.InputTokens,
			&stats.OutputTokens,
			&stats.CacheTokens,
			&stats.TotalTokens,
			&stats.StandardCost,
			&stats.UpstreamCost,
			&stats.UserCost,
		); err != nil {
			return nil, err
		}
		out[accountID] = stats
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *UpstreamCostService) queryAccountTrend(ctx context.Context, startTime, endTime time.Time, accountIDs []int64) (map[int64][]UpstreamCostTrendPoint, error) {
	query := `
		SELECT
			account_id,
			TO_CHAR(created_at, 'YYYY-MM-DD') AS date,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COALESCE(SUM(cache_creation_tokens + cache_read_tokens), 0) AS cache_tokens,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) AS total_tokens,
			COALESCE(SUM(total_cost), 0) AS standard_cost,
			COALESCE(SUM(COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1)), 0) AS upstream_cost,
			COALESCE(SUM(actual_cost), 0) AS user_cost
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2 AND account_id = ANY($3)
		GROUP BY account_id, date
		ORDER BY date ASC
	`
	rows, err := s.db.QueryContext(ctx, query, startTime, endTime, pq.Array(accountIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64][]UpstreamCostTrendPoint, len(accountIDs))
	for rows.Next() {
		var accountID int64
		var point UpstreamCostTrendPoint
		if err := rows.Scan(
			&accountID,
			&point.Date,
			&point.Requests,
			&point.InputTokens,
			&point.OutputTokens,
			&point.CacheTokens,
			&point.TotalTokens,
			&point.StandardCost,
			&point.UpstreamCost,
			&point.UserCost,
		); err != nil {
			return nil, err
		}
		out[accountID] = append(out[accountID], point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *UpstreamCostService) queryAccountModels(ctx context.Context, startTime, endTime time.Time, accountIDs []int64) (map[int64][]UpstreamCostModelBreakdown, error) {
	query := `
		SELECT
			account_id,
			COALESCE(NULLIF(requested_model, ''), NULLIF(model, ''), 'unknown') AS model_name,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) AS total_tokens,
			COALESCE(SUM(total_cost), 0) AS standard_cost,
			COALESCE(SUM(COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1)), 0) AS upstream_cost,
			COALESCE(SUM(actual_cost), 0) AS user_cost
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2 AND account_id = ANY($3)
		GROUP BY account_id, COALESCE(NULLIF(requested_model, ''), NULLIF(model, ''), 'unknown')
		ORDER BY upstream_cost DESC
	`
	rows, err := s.db.QueryContext(ctx, query, startTime, endTime, pq.Array(accountIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64][]UpstreamCostModelBreakdown, len(accountIDs))
	for rows.Next() {
		var accountID int64
		var item UpstreamCostModelBreakdown
		if err := rows.Scan(
			&accountID,
			&item.Model,
			&item.Requests,
			&item.TotalTokens,
			&item.StandardCost,
			&item.UpstreamCost,
			&item.UserCost,
		); err != nil {
			return nil, err
		}
		out[accountID] = append(out[accountID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func mergeUpstreamCostTrend(accounts []UpstreamCostAccountSummary) []UpstreamCostTrendPoint {
	byDate := make(map[string]*UpstreamCostTrendPoint)
	for _, account := range accounts {
		for _, point := range account.Trend {
			merged := byDate[point.Date]
			if merged == nil {
				copyPoint := UpstreamCostTrendPoint{Date: point.Date}
				merged = &copyPoint
				byDate[point.Date] = merged
			}
			merged.Requests += point.Requests
			merged.InputTokens += point.InputTokens
			merged.OutputTokens += point.OutputTokens
			merged.CacheTokens += point.CacheTokens
			merged.TotalTokens += point.TotalTokens
			merged.StandardCost += point.StandardCost
			merged.UpstreamCost += point.UpstreamCost
			merged.UserCost += point.UserCost
		}
	}
	out := make([]UpstreamCostTrendPoint, 0, len(byDate))
	for _, point := range byDate {
		out = append(out, *point)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out
}

func mergeUpstreamCostModels(accounts []UpstreamCostAccountSummary) []UpstreamCostModelBreakdown {
	byModel := make(map[string]*UpstreamCostModelBreakdown)
	for _, account := range accounts {
		for _, item := range account.Models {
			merged := byModel[item.Model]
			if merged == nil {
				copyItem := UpstreamCostModelBreakdown{Model: item.Model}
				merged = &copyItem
				byModel[item.Model] = merged
			}
			merged.Requests += item.Requests
			merged.TotalTokens += item.TotalTokens
			merged.StandardCost += item.StandardCost
			merged.UpstreamCost += item.UpstreamCost
			merged.UserCost += item.UserCost
		}
	}
	out := make([]UpstreamCostModelBreakdown, 0, len(byModel))
	for _, item := range byModel {
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpstreamCost == out[j].UpstreamCost {
			return out[i].Model < out[j].Model
		}
		return out[i].UpstreamCost > out[j].UpstreamCost
	})
	return out
}

func normalizeUpstreamPoolURL(raw string) string {
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
	if parsed.Path == "" {
		parsed.Path = ""
	}
	return strings.TrimRight(parsed.String(), "/")
}

func upstreamPoolNameFromURL(normalizedURL string) string {
	parsed, err := url.Parse(normalizedURL)
	if err == nil && parsed.Host != "" {
		if parsed.Path != "" && parsed.Path != "/" {
			return parsed.Host + strings.TrimRight(parsed.Path, "/")
		}
		return parsed.Host
	}
	return normalizedURL
}

func stringFromAny(v any) string {
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
	default:
		return ""
	}
}

func firstNonEmptyUpstreamCostString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
