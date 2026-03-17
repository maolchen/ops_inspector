import api from './index'

export interface RuleGroup {
  id: number
  name: string
  code: string
  description: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface Rule {
  id: number
  group_id: number
  name: string
  type: boolean
  show_in_table: boolean
  description: string
  query: string
  trend_query: string
  threshold: number | null
  unit: string
  labels: string
  threshold_type: string
  project_scope: string
  enabled: boolean
  sort_order: number
  // 表格列配置
  table_column_order: number
  table_column_width: number
  table_column_type: string
  table_column_label: string
  table_column_merge: boolean
  created_at: string
  updated_at: string
  group?: RuleGroup
}

export const ruleGroupApi = {
  list: () => api.get<any, { data: RuleGroup[] }>('/rule-groups'),

  get: (id: number) => api.get<any, { data: RuleGroup }>(`/rule-groups/${id}`),

  create: (data: Partial<RuleGroup>) => api.post<any, { data: RuleGroup }>('/rule-groups', data),

  update: (id: number, data: Partial<RuleGroup>) => api.put<any, { data: RuleGroup }>(`/rule-groups/${id}`, data),

  delete: (id: number) => api.delete(`/rule-groups/${id}`)
}

export const ruleApi = {
  list: (groupId?: number) => {
    const params = groupId ? { group_id: groupId } : {}
    return api.get<any, { data: Rule[] }>('/rules', { params })
  },

  get: (id: number) => api.get<any, { data: Rule }>(`/rules/${id}`),

  create: (data: Partial<Rule>) => api.post<any, { data: Rule }>('/rules', data),

  update: (id: number, data: Partial<Rule>) => api.put<any, { data: Rule }>(`/rules/${id}`, data),

  delete: (id: number) => api.delete(`/rules/${id}`),

  toggle: (id: number) => api.post(`/rules/${id}/toggle`)
}
