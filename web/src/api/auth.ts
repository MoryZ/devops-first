import { post } from './client'

export const login = (payload: { username: string; password: string }) => post('/auth/login', payload)

export const register = (payload: {
  username: string
  password: string
  email?: string
  remark?: string
}) => post('/auth/register', payload)
