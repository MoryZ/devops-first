export class ApiError extends Error {
  status: number
  data: unknown

  constructor(message: string, status: number, data: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data
  }
}

type RequestOptions = {
  method?: string
  token?: string
  headers?: Record<string, string>
  data?: unknown
  timeoutMs?: number
}

const parseResponse = async (res: Response) => {
  const contentType = res.headers.get('content-type') || ''
  const isJson = contentType.includes('application/json')

  if (isJson) {
    try {
      return await res.json()
    } catch {
      return null
    }
  }

  const text = await res.text()
  return text || null
}

const withTimeout = async (fetcher: (signal: AbortSignal) => Promise<Response>, timeoutMs = 10000) => {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs)

  try {
    return await fetcher(controller.signal)
  } catch (err: any) {
    if (err?.name === 'AbortError') {
      throw new Error('请求超时，请检查前后端服务状态后重试')
    }
    throw err
  } finally {
    clearTimeout(timeoutId)
  }
}

export const request = async (url: string, options: RequestOptions = {}) => {
  const { method = 'GET', token, headers = {}, data, timeoutMs = 10000 } = options

  const mergedHeaders: Record<string, string> = { ...headers }

  if (token) {
    mergedHeaders.Authorization = `Bearer ${token}`
  }

  let body: BodyInit | undefined
  if (data !== undefined) {
    const contentType = mergedHeaders['Content-Type'] || 'application/json'
    mergedHeaders['Content-Type'] = contentType
    body = contentType === 'application/json' ? JSON.stringify(data) : (data as BodyInit)
  }

  const res = await withTimeout(
    (signal) =>
      fetch(url, {
        method,
        headers: mergedHeaders,
        body,
        signal,
      }),
    timeoutMs
  )

  const payload = await parseResponse(res)

  if (!res.ok) {
    const message =
      (payload && typeof payload === 'object' && ((payload as any).error || (payload as any).message)) ||
      (typeof payload === 'string' && payload) ||
      `请求失败(${res.status})`
    throw new ApiError(message, res.status, payload)
  }

  return payload
}

export const get = (url: string, options: Omit<RequestOptions, 'method' | 'data'> = {}) => request(url, { ...options, method: 'GET' })
export const post = (url: string, data?: unknown, options: Omit<RequestOptions, 'method' | 'data'> = {}) => request(url, { ...options, method: 'POST', data })
export const put = (url: string, data?: unknown, options: Omit<RequestOptions, 'method' | 'data'> = {}) => request(url, { ...options, method: 'PUT', data })
export const del = (url: string, options: Omit<RequestOptions, 'method' | 'data'> = {}) => request(url, { ...options, method: 'DELETE' })
