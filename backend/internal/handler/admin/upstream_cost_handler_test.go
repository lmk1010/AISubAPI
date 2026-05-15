package admin

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestUpstreamRemoteAccountSnapshotEffectiveUsedQuotaUsesLargestRemoteValue(t *testing.T) {
	tests := []struct {
		name     string
		snapshot upstreamRemoteAccountSnapshot
		want     int64
	}{
		{
			name: "stat greater than user used",
			snapshot: upstreamRemoteAccountSnapshot{
				UsedQuota: 100,
				StatQuota: 120,
			},
			want: 120,
		},
		{
			name: "user used greater than stat",
			snapshot: upstreamRemoteAccountSnapshot{
				UsedQuota: 120,
				StatQuota: 100,
			},
			want: 120,
		},
		{
			name: "stat missing",
			snapshot: upstreamRemoteAccountSnapshot{
				UsedQuota: 90,
			},
			want: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.snapshot.effectiveUsedQuota(); got != tt.want {
				t.Fatalf("effectiveUsedQuota() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseNewAPIDataReturnsErrorOnBusinessFailure(t *testing.T) {
	var out newAPITokenUsageData
	err := parseNewAPIData([]byte(`{"success":false,"message":"bad token","data":null}`), &out)
	if err == nil {
		t.Fatal("expected business failure to return an error")
	}
	if err.Error() != "bad token" {
		t.Fatalf("error = %q, want %q", err.Error(), "bad token")
	}
}

func TestParseNewAPIDataUnwrapsEnvelope(t *testing.T) {
	var out newAPITokenUsageData
	err := parseNewAPIData([]byte(`{"success":true,"data":{"name":"cc","total_used":500000,"available":1000000}}`), &out)
	if err != nil {
		t.Fatalf("parseNewAPIData returned error: %v", err)
	}
	if out.Name != "cc" || out.TotalUsed != 500000 || out.Available != 1000000 {
		t.Fatalf("unexpected parsed token usage: %+v", out)
	}
}

func TestNewAPITokenUsageDataNormalizeSupportsNewAPIAliases(t *testing.T) {
	var out newAPITokenUsageData
	err := parseNewAPIData([]byte(`{"code":true,"message":"ok","data":{"name":"cc1","total_used":8549138,"total_available":-8548725,"unlimited_quota":true,"expires_at":123}}`), &out)
	if err != nil {
		t.Fatalf("parseNewAPIData returned error: %v", err)
	}
	out.normalize()

	if out.Name != "cc1" || out.TotalUsed != 8549138 {
		t.Fatalf("unexpected parsed token usage: %+v", out)
	}
	if out.Available != -8548725 {
		t.Fatalf("available = %d, want %d", out.Available, int64(-8548725))
	}
	if !out.Unlimited {
		t.Fatal("expected unlimited_quota alias to set Unlimited")
	}
	if out.ExpiredTime != 123 {
		t.Fatalf("expired time = %d, want 123", out.ExpiredTime)
	}
}

func TestBuildRealGroupCostSummariesAllocatesByStandardCost(t *testing.T) {
	groups := []service.UpstreamCostGroupBreakdown{
		{GroupID: 1, GroupName: "default", CurrentGroupRate: 1, Requests: 2, TotalTokens: 200, StandardCost: 80, UserCost: 80, DownstreamRevenueRMB: 80},
		{GroupID: 6, GroupName: "codex-vip-1", CurrentGroupRate: 0.08, Requests: 1, TotalTokens: 100, StandardCost: 20, UserCost: 1.6, DownstreamRevenueRMB: 1.6},
	}

	out := buildRealGroupCostSummaries(groups, 100, 50)

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}
	if out[0].AllocatedUpstreamUsedRMB != 40 {
		t.Fatalf("default allocated = %v, want 40", out[0].AllocatedUpstreamUsedRMB)
	}
	if out[1].AllocatedUpstreamUsedRMB != 10 {
		t.Fatalf("vip allocated = %v, want 10", out[1].AllocatedUpstreamUsedRMB)
	}
	if out[1].DownstreamEffectiveRate != 0.08 {
		t.Fatalf("vip effective rate = %v, want 0.08", out[1].DownstreamEffectiveRate)
	}
}
