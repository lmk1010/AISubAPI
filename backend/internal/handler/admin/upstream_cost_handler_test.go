package admin

import "testing"

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
