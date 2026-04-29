import { apiClient } from './client'

export interface ModelCatalogCategory {
  id: string
  label: string
  description?: string
  models: string[]
}

export interface ModelCatalogConfig {
  categories: ModelCatalogCategory[]
}

export const DEFAULT_MODEL_CATALOG: ModelCatalogConfig = {
  categories: [
    {
      id: 'openai',
      label: 'OpenAI',
      description: '可用模型列表。',
      models: [
        'gpt-5.2',
        'gpt-5.3-codex',
        'gpt-5.3-codex-spark',
        'gpt-5.4',
        'gpt-5.4-mini',
        'gpt-5.5',
        'gpt-image-2',
      ],
    },
  ],
}

function normalizeModelName(model: unknown): string | null {
  if (typeof model !== 'string') return null
  const trimmed = model.trim()
  return trimmed || null
}

export function normalizeModelCatalogConfig(input: unknown): ModelCatalogConfig {
  if (!input || typeof input !== 'object') return DEFAULT_MODEL_CATALOG
  const rawCategories = (input as { categories?: unknown }).categories
  if (!Array.isArray(rawCategories)) return DEFAULT_MODEL_CATALOG

  const categories = rawCategories
    .map((category, index): ModelCatalogCategory | null => {
      if (!category || typeof category !== 'object') return null
      const raw = category as Record<string, unknown>
      const models = Array.isArray(raw.models)
        ? Array.from(new Set(raw.models.map(normalizeModelName).filter(Boolean) as string[]))
        : []
      if (models.length === 0) return null
      const id = typeof raw.id === 'string' && raw.id.trim() ? raw.id.trim() : `category-${index + 1}`
      const label = typeof raw.label === 'string' && raw.label.trim() ? raw.label.trim() : id
      const description = typeof raw.description === 'string' ? raw.description.trim() : ''
      return { id, label, description, models }
    })
    .filter(Boolean) as ModelCatalogCategory[]

  return categories.length > 0 ? { categories } : DEFAULT_MODEL_CATALOG
}

export async function getPublicModelCatalog(): Promise<ModelCatalogConfig> {
  const { data } = await apiClient.get<ModelCatalogConfig>('/model-catalog')
  return normalizeModelCatalogConfig(data)
}

export async function getAdminModelCatalog(): Promise<ModelCatalogConfig> {
  const { data } = await apiClient.get<ModelCatalogConfig>('/admin/model-catalog')
  return normalizeModelCatalogConfig(data)
}

export async function updateAdminModelCatalog(config: ModelCatalogConfig): Promise<ModelCatalogConfig> {
  const { data } = await apiClient.put<ModelCatalogConfig>('/admin/model-catalog', config)
  return data
}
