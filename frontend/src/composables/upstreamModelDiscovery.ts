import axios from 'axios'

export function buildOpenAIModelsEndpoint(baseUrl: string): string {
  const normalized = baseUrl.trim().replace(/\/+$/, '')
  if (/\/v1$/i.test(normalized)) {
    return `${normalized}/models`
  }
  return `${normalized}/v1/models`
}

export function extractOpenAIModelIds(payload: unknown): string[] {
  if (!payload || typeof payload !== 'object') {
    throw new Error('模型列表格式不正确')
  }

  const data = (payload as { data?: unknown }).data
  if (!Array.isArray(data)) {
    throw new Error('模型列表格式不正确')
  }

  const models = data
    .map((item) => {
      if (typeof item === 'string') return item
      if (item && typeof item === 'object') {
        const id = (item as { id?: unknown }).id
        return typeof id === 'string' ? id : ''
      }
      return ''
    })
    .map((model) => model.trim())
    .filter(Boolean)

  const deduped = Array.from(new Set(models))
  if (deduped.length === 0) {
    throw new Error('未从 /v1/models 返回中解析到模型')
  }
  return deduped
}

export async function fetchOpenAICompatibleModels(
  baseUrl: string,
  apiKey: string,
  options?: { signal?: AbortSignal }
): Promise<string[]> {
  const endpoint = buildOpenAIModelsEndpoint(baseUrl)
  const { data } = await axios.get(endpoint, {
    headers: {
      Authorization: `Bearer ${apiKey}`,
      'Content-Type': 'application/json'
    },
    timeout: 30000,
    signal: options?.signal
  })
  return extractOpenAIModelIds(data)
}

export function formatOpenAIModelDiscoveryError(error: unknown): string {
  const maybeResponse = (error as { response?: { status?: number; data?: unknown }; message?: string })?.response
  if (axios.isAxiosError(error) || maybeResponse) {
    const status = maybeResponse?.status
    const data = maybeResponse?.data as any
    const message =
      data?.message ||
      data?.detail ||
      data?.error?.message ||
      data?.error ||
      (error as { message?: string })?.message ||
      '获取模型失败'
    return status ? `HTTP ${status}: ${message}` : String(message)
  }

  if (error instanceof Error) {
    return error.message
  }

  return '获取模型失败'
}
