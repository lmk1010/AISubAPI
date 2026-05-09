import { apiClient } from './client'

export interface UpstreamProvider {
  provider_type: 'newapi' | 'sub2api'
  base_url: string
  access_token: string
  user_id: number
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

export const upstreamCostAPI = { testConnection, getUserInfo, getStats, getLogs }
export default upstreamCostAPI
