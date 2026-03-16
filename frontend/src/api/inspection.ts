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
  created_at: string
}

export const inspectionApi = {
  list: () => api.get<any, { data: InspectionReport[] }>('/inspections'),

  get: (id: number) => api.get<any, { data: { report: InspectionReport; items: InspectionItem[] } }>(`/inspections/${id}`),

  start: (projectId: number, inspector: string) => 
    api.post<any, { data: InspectionReport }>('/inspections/start', { project_id: projectId, inspector }),

  updateSummary: (id: number, summary: string, remark: string) =>
    api.put(`/inspections/${id}/summary`, { summary, remark })
}
