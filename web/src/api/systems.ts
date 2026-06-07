import { get, post, put, del } from './client'

export const listSystems = (token: string, params?: { page?: number; page_size?: number; keyword?: string }) => {
  const query = new URLSearchParams()
  if (params?.page) query.set('page', String(params.page))
  if (params?.page_size) query.set('page_size', String(params.page_size))
  if (params?.keyword) query.set('keyword', params.keyword)
  const qs = query.toString()
  return get(`/api/systems${qs ? `?${qs}` : ''}`, { token })
}

export const createSystem = (token: string, payload: { name: string; description?: string }) =>
  post('/api/systems', payload, { token })

export const getSystem = (token: string, systemId: string) =>
  get(`/api/systems/${systemId}`, { token })

export const updateSystem = (token: string, systemId: string, payload: { name?: string; description?: string; status?: string }) =>
  put(`/api/systems/${systemId}`, payload, { token })

export const deleteSystem = (token: string, systemId: string) =>
  del(`/api/systems/${systemId}`, { token })

export const listSystemPlans = (token: string, systemId: string) =>
  get(`/api/systems/${systemId}/plans`, { token })

export const listSystemPipelines = (token: string, systemId: string) =>
  get(`/api/systems/${systemId}/pipelines`, { token })

export const createSystemPipeline = (
  token: string,
  systemId: string,
  payload: Record<string, any>
) => post(`/api/systems/${systemId}/pipelines`, payload, { token })
