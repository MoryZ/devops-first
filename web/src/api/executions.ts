import { get, post } from './client'

export const startPipelineExecution = (token: string, pipelineId: string, systemId: string) =>
  post(
    `/api/pipelines/${encodeURIComponent(pipelineId)}/execute?system_id=${encodeURIComponent(systemId)}&triggered_by=manual`,
    undefined,
    { token }
  )

export const getBatchStatus = (token: string, batchId: string) =>
  get(`/api/executions/${encodeURIComponent(batchId)}`, { token })

export const listExecutionLogs = (token: string, batchId: string, limit = 500) =>
  get(`/api/executions/${batchId}/logs?limit=${limit}`, { token })

export const listExecutionCommits = (token: string, batchId: string, limit = 20) =>
  get(`/api/executions/${batchId}/commits?limit=${limit}`, { token })

export const cancelExecutionBatch = (token: string, batchId: string) =>
  post(`/api/executions/${batchId}/cancel`, undefined, { token })

export const rerunExecutionNode = (token: string, batchId: string, nodeId: string) =>
  post(`/api/executions/${batchId}/rerun-node`, { node_id: nodeId }, { token })
