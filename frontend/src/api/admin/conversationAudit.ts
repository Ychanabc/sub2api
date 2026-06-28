import { apiClient } from '../client'

export interface ConversationAuditLog {
  id: number
  request_id: string
  user_id: number
  api_key_id: number
  account_id: number
  group_id?: number | null
  model: string
  endpoint: string
  request_type: number
  request_excerpt: string
  response_excerpt: string
  request_hash: string
  response_hash: string
  status_code: number
  duration_ms: number
  created_at: string
}

export interface ConversationAuditParams {
  page?: number
  page_size?: number
  q?: string
  model?: string
  user_id?: number
  api_key_id?: number
  account_id?: number
  start_time?: string
  end_time?: string
}

export interface ConversationAuditResponse {
  items: ConversationAuditLog[]
  pagination: {
    page: number
    page_size: number
    total: number
    pages: number
  }
}

export interface ConversationAuditAccessLog {
  id: number
  user_id: number
  email: string
  ip_address: string
  user_agent: string
  fingerprint: string
  success: boolean
  failure_reason: string
  created_at: string
}

export interface ConversationAuditStatus {
  secondary_password_configured: boolean
  totp_enabled: boolean
  unlock_ttl_seconds: number
  cleanup_enabled: boolean
  retention_days: number
}

export interface ConversationAuditUnlockResponse {
  unlock_token: string
  expires_at: string
  expires_in: number
}

function auditTokenConfig(unlockToken?: string) {
  return unlockToken
    ? { headers: { 'X-Conversation-Audit-Token': unlockToken } }
    : undefined
}

export const conversationAuditAPI = {
  async status(): Promise<ConversationAuditStatus> {
    const { data } = await apiClient.get<ConversationAuditStatus>('/admin/conversation-audits/status')
    return data
  },

  async unlock(payload: { secondary_password: string; totp_code: string; fingerprint: string }): Promise<ConversationAuditUnlockResponse> {
    const { data } = await apiClient.post<ConversationAuditUnlockResponse>('/admin/conversation-audits/unlock', payload)
    return data
  },

  async list(params: ConversationAuditParams, unlockToken?: string): Promise<ConversationAuditResponse> {
    const { data } = await apiClient.get<ConversationAuditResponse>('/admin/conversation-audits', {
      params,
      ...auditTokenConfig(unlockToken)
    })
    return data
  },

  async clearAll(unlockToken: string): Promise<{ deleted: number }> {
    const { data } = await apiClient.delete<{ deleted: number }>('/admin/conversation-audits', auditTokenConfig(unlockToken))
    return data
  },

  async accessLogs(params: { page?: number; page_size?: number }, unlockToken?: string): Promise<{
    items: ConversationAuditAccessLog[]
    pagination: ConversationAuditResponse['pagination']
  }> {
    const { data } = await apiClient.get<{
      items: ConversationAuditAccessLog[]
      pagination: ConversationAuditResponse['pagination']
    }>('/admin/conversation-audits/access-logs', {
      params,
      ...auditTokenConfig(unlockToken)
    })
    return data
  },
}
