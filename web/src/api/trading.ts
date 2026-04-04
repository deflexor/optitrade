import { api } from './client'

export type TradingStatus = {
  account_status: string
  bot_mode: string
  runner_running: boolean
}

export type OpportunityLeg = {
  instrument: string
  bid: number
  ask: number
}

export type OpportunityRow = {
  id: string
  strategy_name: string
  status: string
  legs: OpportunityLeg[]
  greeks_note?: string
  max_profit: string
  max_loss: string
  recommendation: string
  rationale: string
  expected_edge: string
  edge_after_costs: number
}

export type OpportunitiesResponse = {
  paused?: boolean
  disabled?: boolean
  message?: string
  resume_hint?: string
  updated_at_ms: number
  rows: OpportunityRow[]
}

export async function getTradingStatus(): Promise<TradingStatus> {
  const { data } = await api.get<TradingStatus>('/trading/status')
  return data
}

export async function putTradingMode(
  bot_mode: 'manual' | 'auto' | 'paused',
): Promise<{ bot_mode: string }> {
  const { data } = await api.put<{ bot_mode: string }>('/trading/mode', { bot_mode })
  return data
}

export async function getOpportunities(): Promise<OpportunitiesResponse> {
  const { data } = await api.get<OpportunitiesResponse>('/opportunities')
  return data
}

export async function postOpportunityOpen(id: string): Promise<void> {
  await api.post(`/opportunities/${encodeURIComponent(id)}/open`)
}

export async function postOpportunityCancel(id: string): Promise<void> {
  await api.post(`/opportunities/${encodeURIComponent(id)}/cancel`)
}

export async function postOpportunityClose(id: string): Promise<void> {
  await api.post(`/opportunities/${encodeURIComponent(id)}/close`)
}
