import api from './index'

export interface InspectionReport {
  id: number
  project_id: number
  project_name: string
  inspector: string
  start_time: string
  end_time: string
  status: string
  total_items: number
  warning_count: number
  critical_count: number
  summary: string
  remark: string
  created_at: string
}

export interface InspectionItem {
  id: number
  report_id: number
  rule_id: number
  group_id: number
  group_name: string
  rule_name: string
  instance: string
  value: number
  status: string
  show_in_table: boolean
  trend_data: string
  labels: string
  unit: string
  // 表格列配置
  table_column_order: number
  table_column_width: number
  table_column_type: string
  table_column_label: string
  table_column_merge: boolean
  created_at: string
}

export interface ListParams {
  keyword?: string
  page?: number
  page_size?: number
}

export interface ListResult {
  total: number
  page: number
  page_size: number
  list: InspectionReport[]
}

export const inspectionApi = {
  list: (params?: ListParams) => {
    const queryParams = new URLSearchParams()
    if (params?.keyword) queryParams.append('keyword', params.keyword)
    if (params?.page) queryParams.append('page', params.page.toString())
    if (params?.page_size) queryParams.append('page_size', params.page_size.toString())
    const query = queryParams.toString()
    return api.get<any, { data: ListResult }>(`/inspections${query ? '?' + query : ''}`)
  },

  get: (id: number) => api.get<any, { data: { report: InspectionReport; items: InspectionItem[] } }>(`/inspections/${id}`),

  start: (projectId: number, inspector: string) => 
    api.post<any, { data: InspectionReport }>('/inspections/start', { project_id: projectId, inspector }),

  updateSummary: (id: number, summary: string, remark: string) =>
    api.put(`/inspections/${id}/summary`, { summary, remark })
}

// 系统配置API
export interface SystemConfig {
  report_retention_days: string
}

export const systemApi = {
  getConfigs: () => api.get<any, { data: SystemConfig }>('/system/configs'),
  
  updateConfig: (key: string, value: string) => 
    api.put('/system/configs', { key, value }),
  
  cleanupReports: (days: number) => 
    api.post<any, { message: string; count: number }>(`/system/cleanup?days=${days}`)
}
