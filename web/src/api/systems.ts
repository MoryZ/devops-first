import { get, post } from './client'

export const listSystems = (token: string) => get('/api/systems', { token })

export const createSystem = (token: string, payload: { name: string; description?: string }) =>
  post('/api/systems', payload, { token })

export const listSystemPlans = (token: string, systemId: string) =>
  get(`/api/systems/${systemId}/plans`, { token })

export const listSystemPipelines = (token: string, systemId: string) =>
  get(`/api/systems/${systemId}/pipelines`, { token })

export const createSystemPipeline = (
  token: string,
  systemId: string,
  payload: Record<string, any>
) => post(`/api/systems/${systemId}/pipelines`, payload, { token })
