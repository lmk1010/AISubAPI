/**
 * Admin Channels API endpoints
 * Handles channel management for administrators
 */

import { apiClient } from '../client'
import type { BillingMode, ChannelStatus, BillingModelSource } from '@/constants/channel'

export type { BillingMode } from '@/constants/channel'

export interface PricingInterval {
  id?: number
  min_tokens: number
  max_tokens: number | null
  tier_label: string
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_write_5m_price: number | null
  cache_write_1h_price: number | null
  cache_read_price: number | null
  per_request_price: number | null
  sort_order: number
}

export interface ChannelModelPricing {
  id?: number
  platform: string
  models: string[]
  billing_mode: BillingMode
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_write_5m_price: number | null
  cache_write_1h_price: number | null
  cache_read_price: number | null
  image_output_price: number | null
  per_request_price: number | null
  intervals: PricingInterval[]
}

export interface AccountStatsPricingRule {
  id?: number
  name: string
  group_ids: number[]
  account_ids: number[]
  pricing: ChannelModelPricing[]
}

export interface Channel {
  id: number
  name: string
  description: string
  status: ChannelStatus
  billing_model_source: BillingModelSource
  restrict_models: boolean
  features_config?: Record<string, unknown>
  group_ids: number[]
  model_pricing: ChannelModelPricing[]
  model_mapping: Record<string, Record<string, string>> // platform → {src→dst}
  apply_pricing_to_account_stats: boolean
  account_stats_pricing_rules: AccountStatsPricingRule[]
  created_at: string
  updated_at: string
}

export interface CreateChannelRequest {
  name: string
  description?: string
  group_ids?: number[]
  model_pricing?: ChannelModelPricing[]
  model_mapping?: Record<string, Record<string, string>>
  billing_model_source?: string
  restrict_models?: boolean
  features_config?: Record<string, unknown>
  apply_pricing_to_account_stats?: boolean
  account_stats_pricing_rules?: AccountStatsPricingRule[]
}

export interface UpdateChannelRequest {
  name?: string
  description?: string
  status?: string
  group_ids?: number[]
  model_pricing?: ChannelModelPricing[]
  model_mapping?: Record<string, Record<string, string>>
  billing_model_source?: string
  restrict_models?: boolean
  features_config?: Record<string, unknown>
  apply_pricing_to_account_stats?: boolean
  account_stats_pricing_rules?: AccountStatsPricingRule[]
}

interface PaginatedResponse<T> {
  items: T[]
  total: number
}

/**
 * List channels with pagination
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: string
    search?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<Channel>> {
  const { data } = await apiClient.get<PaginatedResponse<Channel>>('/admin/channels', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    },
    signal: options?.signal
  })
  return data
}

/**
 * Get channel by ID
 */
export async function getById(id: number): Promise<Channel> {
  const { data } = await apiClient.get<Channel>(`/admin/channels/${id}`)
  return data
}

/**
 * Create a new channel
 */
export async function create(req: CreateChannelRequest): Promise<Channel> {
  const { data } = await apiClient.post<Channel>('/admin/channels', req)
  return data
}

/**
 * Update a channel
 */
export async function update(id: number, req: UpdateChannelRequest): Promise<Channel> {
  const { data } = await apiClient.put<Channel>(`/admin/channels/${id}`, req)
  return data
}

/**
 * Delete a channel
 */
export async function remove(id: number): Promise<void> {
  await apiClient.delete(`/admin/channels/${id}`)
}

export interface ModelDefaultPricing {
  found: boolean
  input_price?: number    // per-token price
  output_price?: number
  cache_write_price?: number
  cache_write_5m_price?: number
  cache_write_1h_price?: number
  cache_read_price?: number
  image_output_price?: number
}

export async function getModelDefaultPricing(model: string): Promise<ModelDefaultPricing> {
  const { data } = await apiClient.get<ModelDefaultPricing>('/admin/channels/model-pricing', {
    params: { model }
  })
  return data
}

// ===== LiteLLM sync =====

export interface LiteLLMSuggestionItem {
  model: string
  input_price: number
  output_price: number
  cache_read_price: number
  cache_write_price: number
  cache_write_5m_price: number
  cache_write_1h_price: number
  image_output_price: number
  mode: string
  supports_cache: boolean
  already_configured: boolean
}

export interface LiteLLMSuggestionsResponse {
  platform: string
  items: LiteLLMSuggestionItem[]
}

export async function getLiteLLMSuggestions(
  channelId: number,
  platform: string
): Promise<LiteLLMSuggestionsResponse> {
  const { data } = await apiClient.get<LiteLLMSuggestionsResponse>(
    `/admin/channels/${channelId}/litellm-suggestions`,
    { params: { platform } }
  )
  return data
}

export interface ImportLiteLLMResponse {
  imported: string[]
  skipped: string[]
  missing: string[]
}

export async function importLiteLLMModels(
  channelId: number,
  platform: string,
  models: string[]
): Promise<ImportLiteLLMResponse> {
  const { data } = await apiClient.post<ImportLiteLLMResponse>(
    `/admin/channels/${channelId}/import-litellm-models`,
    { platform, models }
  )
  return data
}

const channelsAPI = {
  list,
  getById,
  create,
  update,
  remove,
  getModelDefaultPricing,
  getLiteLLMSuggestions,
  importLiteLLMModels
}
export default channelsAPI
