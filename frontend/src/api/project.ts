import api from './index'

export interface Project {
  id: number
  name: string
  description: string
  prometheus_url: string
  token?: string
  created_at: string
  updated_at: string
}

export const projectApi = {
  list: () => api.get<any, { data: Project[] }>('/projects'),

  get: (id: number) => api.get<any, { data: Project }>(`/projects/${id}`),

  create: (data: Partial<Project>) => api.post<any, { data: Project }>('/projects', data),

  update: (id: number, data: Partial<Project>) => api.put<any, { data: Project }>(`/projects/${id}`, data),

  delete: (id: number) => api.delete(`/projects/${id}`)
}
