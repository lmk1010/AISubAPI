import { apiClient } from './client'

export interface UpstreamProvider {
  provider_type: 'newapi' | 'sub2api'
  base_url: string
  access_token: string
  user_id: number
  email?: string
  password?: string
}

export interface UpstreamLocalSummaryParams {
  start_date?: string
  end_date?: string
  timezone?: string
  refresh?: boolean
}

export interface UpstreamUserInfo {
  id: number
  username: string
  display_name: string
  email: string
  quota: number
  used_quota: number
  request_count: number
  group: string
  provider_type: string
  base_url: string
}

export interface UpstreamStat {
  quota: number
  quota_usd: number
  rpm: number
  tpm: number
}

export interface UpstreamLogItem {
  id: number
  created_at: number
  model_name: string
  quota: number
  prompt_tokens: number
  completion_tokens: number
  use_time: number
  is_stream: boolean
  token_name: string
  channel: number
  group: string
  quota_usd: number
}

export interface UpstreamLogsResponse {
  page: number
  page_size: number
  total: number
  items: UpstreamLogItem[]
}

export interface UpstreamCostTotals {
  requests: number
  total_tokens: number
  standard_cost: number
  upstream_cost: number
  user_cost: number
  downstream_revenue_rmb: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
  profit: number
}

export interface UpstreamCostTrendPoint {
  date: string
  requests: number
  input_tokens: number
  output_tokens: number
  cache_tokens: number
  total_tokens: number
  standard_cost: number
  upstream_cost: number
  user_cost: number
  downstream_revenue_rmb: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
}

export interface UpstreamCostModelBreakdown {
  model: string
  requests: number
  total_tokens: number
  standard_cost: number
  upstream_cost: number
  user_cost: number
  downstream_revenue_rmb: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
}

export interface UpstreamCostAccountSummary {
  account_id: number
  account_name: string
  platform: string
  status: string
  group_ids: number[]
  rate_multiplier: number
  provider_type: string
  base_url: string
  pool_key: string
  pool_name: string
  requests: number
  total_tokens: number
  standard_cost: number
  upstream_cost: number
  user_cost: number
  downstream_revenue_rmb: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
  profit: number
  groups: UpstreamCostGroupBreakdown[]
  trend: UpstreamCostTrendPoint[]
  models: UpstreamCostModelBreakdown[]
}

export interface UpstreamCostGroupBreakdown {
  group_id: number
  group_name: string
  current_group_rate: number
  requests: number
  total_tokens: number
  standard_cost: number
  upstream_cost: number
  user_cost: number
  downstream_revenue_rmb: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
}

export interface UpstreamCostPoolSummary {
  pool_key: string
  pool_name: string
  provider_type: string
  base_url: string
  account_count: number
  account_ids: number[]
  requests: number
  total_tokens: number
  standard_cost: number
  upstream_cost: number
  user_cost: number
  downstream_revenue_rmb: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
  profit: number
  trend: UpstreamCostTrendPoint[]
  models: UpstreamCostModelBreakdown[]
  accounts: UpstreamCostAccountSummary[]
}

export interface UpstreamCostLocalSummary {
  start_date: string
  end_date: string
  totals: UpstreamCostTotals
  pools: UpstreamCostPoolSummary[]
}

export interface UpstreamRealSummaryTotals {
  account_count: number
  error_count: number
  requests: number
  total_tokens: number
  standard_cost: number
  downstream_revenue_rmb: number
  downstream_usage_quota: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
  local_account_cost_rmb: number
  upstream_used_rmb: number
  allocated_upstream_used_rmb: number
  unallocated_upstream_used_rmb: number
  upstream_remaining_rmb: number
  upstream_used_usd: number
  successful_recharge_rmb: number
  profit_rmb: number
  upstream_effective_rate: number
  downstream_effective_rate: number
  weighted_upstream_account_rate: number
}

export interface UpstreamRealAccountSummary {
  account_id: number
  account_name: string
  platform: string
  status: string
  group_ids: number[]
  provider_type: string
  base_url: string
  requests: number
  total_tokens: number
  standard_cost: number
  downstream_revenue_rmb: number
  downstream_usage_quota: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
  local_account_cost_rmb: number
  upstream_used_rmb: number
  upstream_remaining_rmb: number
  upstream_used_usd: number
  profit_rmb: number
  upstream_configured_rate: number
  upstream_effective_rate: number
  downstream_effective_rate: number
  source: 'token' | 'account_total' | 'unallocated' | 'token_error' | 'unknown' | string
  token_name: string
  token_hash: string
  remote_status: 'ok' | 'partial' | 'error' | string
  error: string
  groups: UpstreamRealGroupSummary[]
  trend: UpstreamCostTrendPoint[]
  models: UpstreamCostModelBreakdown[]
}

export interface UpstreamRealGroupSummary {
  group_id: number
  group_name: string
  current_group_rate: number
  requests: number
  total_tokens: number
  standard_cost: number
  downstream_revenue_rmb: number
  downstream_usage_quota: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
  local_account_cost_rmb: number
  allocated_upstream_used_rmb: number
  profit_rmb: number
  downstream_effective_rate: number
}

export interface UpstreamRealPoolSummary {
  pool_key: string
  pool_name: string
  provider_type: string
  base_url: string
  account_count: number
  account_ids: number[]
  requests: number
  total_tokens: number
  standard_cost: number
  downstream_revenue_rmb: number
  downstream_usage_quota: number
  balance_revenue_rmb: number
  subscription_quota_cost: number
  subscription_revenue_rmb: number
  local_account_cost_rmb: number
  upstream_used_rmb: number
  allocated_upstream_used_rmb: number
  unallocated_upstream_used_rmb: number
  upstream_remaining_rmb: number
  upstream_used_usd: number
  successful_recharge_rmb: number
  profit_rmb: number
  upstream_effective_rate: number
  downstream_effective_rate: number
  weighted_upstream_account_rate: number
  status: 'ok' | 'partial' | 'error' | string
  errors: string[]
  accounts: UpstreamRealAccountSummary[]
}

export interface UpstreamRealSummary {
  start_date: string
  end_date: string
  generated_at: string
  scope: string
  totals: UpstreamRealSummaryTotals
  pools: UpstreamRealPoolSummary[]
}

export async function getLocalSummary(params?: UpstreamLocalSummaryParams): Promise<UpstreamCostLocalSummary> {
  const { data } = await apiClient.get<UpstreamCostLocalSummary>('/admin/upstream-cost/local-summary', { params })
  return data
}

export async function getRealSummary(params?: UpstreamLocalSummaryParams): Promise<UpstreamRealSummary> {
  const { data } = await apiClient.get<UpstreamRealSummary>('/admin/upstream-cost/real-summary', { params })
  return data
}

export async function testConnection(provider: UpstreamProvider) {
  const { data } = await apiClient.post('/admin/upstream-cost/test-connection', provider)
  return data
}

export async function getUserInfo(provider: UpstreamProvider): Promise<UpstreamUserInfo> {
  const { data } = await apiClient.post<UpstreamUserInfo>('/admin/upstream-cost/user-info', provider)
  return data
}

export async function getStats(provider: UpstreamProvider): Promise<UpstreamStat> {
  const { data } = await apiClient.post<UpstreamStat>('/admin/upstream-cost/stats', provider)
  return data
}

export async function getLogs(
  provider: UpstreamProvider,
  options?: { page?: number; page_size?: number; model?: string; token_id?: number }
): Promise<UpstreamLogsResponse> {
  const { data } = await apiClient.post<UpstreamLogsResponse>('/admin/upstream-cost/logs', {
    ...provider,
    ...options
  })
  return data
}

export const upstreamCostAPI = { getLocalSummary, getRealSummary, testConnection, getUserInfo, getStats, getLogs }
export default upstreamCostAPI
