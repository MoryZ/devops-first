import { get, post } from './client'

export const listPlansBySystem = (token: string, systemId: string) =>
  get(`/api/systems/${systemId}/plans`, { token })

export const createPlan = (token: string, systemId: string, payload: Record<string, any>) =>
  post(`/api/systems/${systemId}/plans`, payload, { token })

export const listPipelinesByPlan = (token: string, planId: string) =>
  get(`/api/plans/${planId}/pipelines`, { token })
