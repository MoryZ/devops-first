import { get, put } from './client'

export const listPipelineConfigs = (token: string) => get('/api/pipelines', { token })

export const upsertPipelineConfig = (token: string, payload: Record<string, any>) =>
  put('/api/pipelines/config', payload, { token })

export const listPipelineExecutions = (token: string, pipelineId: string, limit?: number) =>
  get(`/api/pipelines/${pipelineId}/executions${limit ? `?limit=${limit}` : ''}`, { token })

export const getPipelineConfig = (token: string, pipelineId: string) =>
  get(`/api/pipelines/${pipelineId}/config`, { token })

export const upsertPipelineReleaseUnitBinding = (
  token: string,
  pipelineId: string,
  releaseUnitId: string
) => put(`/api/pipelines/${pipelineId}/release-unit`, { release_unit_id: releaseUnitId }, { token })

export const fetchPipelineLatestCommit = (token: string, pipelineId: string) =>
  get(`/api/pipelines/${pipelineId}/latest-commit`, { token })
