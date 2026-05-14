package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type upstreamCostAccountListerStub struct {
	accounts []Account
}

func (s upstreamCostAccountListerStub) List(context.Context, pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	return s.accounts, &pagination.PaginationResult{Total: int64(len(s.accounts)), Page: 1, PageSize: upstreamCostPageSize}, nil
}

func TestUpstreamCostService_GetLocalSummary_GroupsSharedPoolAndKeepsAccountMultipliers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rateA := 1.2
	rateB := 1.4
	accounts := []Account{
		{
			ID:             1,
			Name:           "foxnio-kiro-a",
			Platform:       "claude",
			Status:         StatusActive,
			RateMultiplier: &rateA,
			GroupIDs:       []int64{101},
			Extra: map[string]any{
				"upstream_provider": map[string]any{
					"type":     "sub2api",
					"base_url": "https://api.foxnio.com/",
				},
			},
		},
		{
			ID:             2,
			Name:           "foxnio-kiro-b",
			Platform:       "claude",
			Status:         StatusActive,
			RateMultiplier: &rateB,
			GroupIDs:       []int64{102},
			Extra: map[string]any{
				"upstream_provider": map[string]any{
					"type":     "sub2api",
					"base_url": "https://api.foxnio.com",
				},
			},
		},
	}

	statsRows := sqlmock.NewRows([]string{
		"account_id", "requests", "input_tokens", "output_tokens", "cache_tokens", "total_tokens", "standard_cost", "upstream_cost", "user_cost",
	}).
		AddRow(int64(1), int64(2), int64(100), int64(50), int64(20), int64(170), 10.0, 12.0, 18.0).
		AddRow(int64(2), int64(1), int64(80), int64(20), int64(10), int64(110), 5.0, 7.0, 8.0)
	mock.ExpectQuery("SELECT\\s+account_id,\\s+COUNT\\(\\*\\) AS requests").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(statsRows)

	trendRows := sqlmock.NewRows([]string{
		"account_id", "date", "requests", "input_tokens", "output_tokens", "cache_tokens", "total_tokens", "standard_cost", "upstream_cost", "user_cost",
	}).
		AddRow(int64(1), "2026-05-12", int64(2), int64(100), int64(50), int64(20), int64(170), 10.0, 12.0, 18.0).
		AddRow(int64(2), "2026-05-12", int64(1), int64(80), int64(20), int64(10), int64(110), 5.0, 7.0, 8.0)
	mock.ExpectQuery("TO_CHAR\\(created_at, 'YYYY-MM-DD'\\) AS date").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(trendRows)

	modelRows := sqlmock.NewRows([]string{
		"account_id", "model_name", "requests", "total_tokens", "standard_cost", "upstream_cost", "user_cost",
	}).
		AddRow(int64(1), "claude-sonnet-4.5", int64(2), int64(170), 10.0, 12.0, 18.0).
		AddRow(int64(2), "claude-sonnet-4.5", int64(1), int64(110), 5.0, 7.0, 8.0)
	mock.ExpectQuery("COALESCE\\(NULLIF\\(requested_model").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(modelRows)

	svc := &UpstreamCostService{
		db:          db,
		accountRepo: upstreamCostAccountListerStub{accounts: accounts},
	}
	start := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	summary, err := svc.GetLocalSummary(context.Background(), start, end)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Len(t, summary.Pools, 1)

	pool := summary.Pools[0]
	require.Equal(t, "sub2api|https://api.foxnio.com", pool.PoolKey)
	require.Equal(t, "api.foxnio.com", pool.PoolName)
	require.Equal(t, 2, pool.AccountCount)
	require.InDelta(t, 19.0, pool.UpstreamCost, 0.000001)
	require.InDelta(t, 26.0, pool.UserCost, 0.000001)
	require.InDelta(t, 7.0, pool.Profit, 0.000001)
	require.InDelta(t, 19.0, summary.Totals.UpstreamCost, 0.000001)
	require.InDelta(t, 26.0, summary.Totals.UserCost, 0.000001)
	require.Len(t, pool.Trend, 1)
	require.InDelta(t, 19.0, pool.Trend[0].UpstreamCost, 0.000001)
	require.Len(t, pool.Models, 1)
	require.InDelta(t, 19.0, pool.Models[0].UpstreamCost, 0.000001)
}

func TestUpstreamCostProviderConfig_ManualPoolKeySplitsSameURL(t *testing.T) {
	account := Account{
		ID:             1,
		Name:           "manual-pool",
		RateMultiplier: ptrFloat64ForUpstreamCostTest(1.0),
		Extra: map[string]any{
			"upstream_provider": map[string]any{
				"type":      "newapi",
				"base_url":  "https://example.com/",
				"pool_key":  "vendor-a-main",
				"pool_name": "Vendor A Main",
			},
		},
	}

	cfg, ok := upstreamCostProviderConfig(account)
	require.True(t, ok)
	require.Equal(t, "vendor-a-main", cfg.PoolKey)
	require.Equal(t, "Vendor A Main", cfg.PoolName)
}

func ptrFloat64ForUpstreamCostTest(v float64) *float64 {
	return &v
}

var _ upstreamCostAccountLister = upstreamCostAccountListerStub{}
