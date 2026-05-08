import { get, put } from './client'

export const listTaskTemplates = (token: string) => get('/api/task-templates', { token })

export const getPipelineBpm = (token: string, pipelineId: string) =>
  get(`/api/pipelines/${pipelineId}/bpm`, { token })

export const savePipelineBpm = (token: string, pipelineId: string, payload: Record<string, any>) =>
  put(`/api/pipelines/${pipelineId}/bpm`, payload, { token })
