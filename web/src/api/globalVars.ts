import { del, get, put } from './client'

export const listGlobalVars = (token: string) => get('/api/global-vars', { token })

export const saveGlobalVar = (
  token: string,
  payload: { key: string; fields: Array<{ name: string; value: string; is_secret: boolean }>; description?: string }
) => put('/api/global-vars', payload, { token })

export const deleteGlobalVar = (token: string, id: string) => del(`/api/global-vars/${id}`, { token })
