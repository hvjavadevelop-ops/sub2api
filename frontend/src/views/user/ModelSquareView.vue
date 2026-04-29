<template>
  <AppLayout>
    <div class="w-full max-w-full space-y-4 overflow-x-hidden sm:space-y-5">
      <div class="rounded-2xl border border-emerald-200 bg-gradient-to-r from-emerald-50 via-primary-50 to-white p-4 shadow-sm dark:border-emerald-900/40 dark:from-emerald-950/25 dark:via-primary-950/20 dark:to-dark-800 sm:p-5">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white sm:text-lg">还没有安装 ChatGPT Codex？立即下载安装</h2>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">前往 Codex 官方页面，按系统选择对应安装包。</p>
          </div>
          <a
            href="https://developers.openai.com/codex/app"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex w-fit items-center justify-center rounded-full border border-primary-200 bg-white/80 px-4 py-2 text-sm font-semibold text-primary-600 shadow-sm transition hover:border-primary-300 hover:bg-primary-50 hover:text-primary-700 dark:border-primary-900/50 dark:bg-dark-900/70 dark:text-primary-300 dark:hover:border-primary-700 dark:hover:bg-primary-950/40 dark:hover:text-primary-200"
          >官方页面 →</a>
        </div>
      </div>

      <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(300px,380px)] xl:items-start xl:gap-6">
        <div class="min-w-0 space-y-4">
          <div class="rounded-3xl border border-emerald-200 bg-gradient-to-br from-emerald-50 via-white to-primary-50 p-4 shadow-sm dark:border-emerald-900/40 dark:from-emerald-950/25 dark:via-dark-800 dark:to-primary-950/20 sm:p-5">
            <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="rounded-full bg-emerald-500/10 px-2.5 py-1 text-[11px] font-semibold text-emerald-700 dark:text-emerald-300">推荐</span>
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">只狼云一键配置助手</h2>
                </div>
                <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">复制下面命令运行，按提示选择 Vibe Coding 工具、模型并输入 SK 密钥，即可自动写入对应工具的中转配置。</p>
              </div>
              <div class="flex flex-wrap gap-2 text-xs font-medium sm:justify-end">
                <button
                  type="button"
                  class="rounded-full px-3 py-1.5 transition"
                  :class="installerPlatform === 'unix' ? 'bg-emerald-600 text-white shadow-sm' : 'bg-white/80 text-gray-600 hover:bg-white dark:bg-dark-800 dark:text-gray-300'"
                  @click="installerPlatform = 'unix'"
                >macOS / Linux / WSL2</button>
                <button
                  type="button"
                  class="rounded-full px-3 py-1.5 transition"
                  :class="installerPlatform === 'windows' ? 'bg-emerald-600 text-white shadow-sm' : 'bg-white/80 text-gray-600 hover:bg-white dark:bg-dark-800 dark:text-gray-300'"
                  @click="installerPlatform = 'windows'"
                >Windows PowerShell</button>
              </div>
            </div>
            <div class="mt-4 flex flex-col gap-3 lg:flex-row lg:items-stretch">
              <pre class="min-w-0 flex-1 overflow-x-auto whitespace-pre rounded-2xl bg-gray-950 p-3 text-[12px] leading-6 text-gray-100 sm:p-4 sm:text-sm"><code>{{ installerCommand }}</code></pre>
              <button type="button" class="btn btn-primary w-full justify-center lg:w-auto lg:px-6" @click="copyText(installerCommand)">复制配置命令</button>
            </div>
          </div>
          <div class="rounded-2xl border border-gray-200 bg-white p-3 shadow-sm dark:border-dark-700 dark:bg-dark-800 sm:p-4">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div class="flex min-w-0 flex-1 flex-col gap-2 sm:flex-row sm:items-center">
                <div class="relative min-w-0 flex-1">
                  <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input v-model="searchQuery" class="input pl-10" placeholder="搜索模型" />
                </div>
                <button
                  type="button"
                  class="btn btn-secondary w-full shrink-0 justify-center sm:w-auto"
                  :disabled="loading"
                  @click="loadModels"
                >
                  <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
                  <span>刷新</span>
                </button>
              </div>
              <div class="shrink-0 text-sm text-gray-500 dark:text-gray-400">
                共 {{ filteredModels.length }} / {{ displayModels.length }} 个模型
              </div>
            </div>
          </div>

          <div v-if="loading" class="grid gap-3 sm:grid-cols-2 2xl:grid-cols-3">
            <div v-for="i in 6" :key="i" class="h-32 animate-pulse rounded-2xl bg-gray-100 dark:bg-dark-700"></div>
          </div>

          <div v-else-if="filteredModels.length === 0" class="rounded-2xl border border-dashed border-gray-300 bg-white p-10 text-center dark:border-dark-600 dark:bg-dark-800">
            <p class="text-base font-medium text-gray-700 dark:text-gray-200">暂无匹配模型</p>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">可以换个关键词，或点击刷新恢复展示。</p>
          </div>

          <div v-else class="space-y-5">
            <section
              v-for="category in filteredCategories"
              :key="category.id"
              class="min-w-0 rounded-2xl border border-gray-200 bg-white p-3 shadow-sm dark:border-dark-700 dark:bg-dark-800 sm:p-4"
            >
              <div class="mb-3 flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
                <div>
                  <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ category.label }}</h2>
                  <p v-if="category.description" class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ category.description }}</p>
                </div>
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ category.models.length }} 个模型</span>
              </div>

              <div class="grid min-w-0 gap-3 sm:grid-cols-2 2xl:grid-cols-3">
                <div
                  v-for="model in category.models"
                  :key="`${category.id}:${model.name}`"
                  class="group min-w-0 rounded-2xl border border-gray-200 bg-white p-3 shadow-sm transition hover:border-primary-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-900 dark:hover:border-primary-700 sm:p-4"
                >
                  <div class="flex min-w-0 items-start justify-between gap-2 sm:gap-3">
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-2">
                        <ModelIcon :model="model.name" size="22px" />
                        <h3 class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ model.name }}</h3>
                      </div>
                      <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">{{ model.platformLabel }}</p>
                    </div>
                    <button
                      type="button"
                      class="shrink-0 rounded-lg border border-gray-200 px-2 py-1 text-xs text-gray-600 hover:bg-gray-50 dark:border-dark-600 dark:text-gray-300 dark:hover:bg-dark-700"
                      @click="copyText(model.name)"
                    >
                      复制
                    </button>
                  </div>
                  <div class="mt-3 flex flex-wrap gap-1.5">
                    <span
                      v-for="tag in model.tags"
                      :key="tag"
                      class="rounded-full bg-primary-50 px-2 py-0.5 text-xs text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                    >{{ tag }}</span>
                  </div>
                </div>
              </div>
            </section>
          </div>
        </div>

        <div class="min-w-0 space-y-4">
          <div class="rounded-2xl border border-gray-200 bg-white p-3 text-center shadow-sm dark:border-dark-700 dark:bg-dark-800 sm:p-4">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">联系我！</h2>
            <img :src="qqGroupQrImage" alt="云栖小铺交流群二维码" class="mx-auto mt-3 max-h-[420px] w-full max-w-[240px] rounded-xl object-contain shadow-sm" />
            <p class="mt-3 text-sm font-semibold leading-6 text-emerald-700 dark:text-emerald-300">用心服务好每一位小宝～</p>
            <p class="text-sm font-semibold leading-6 text-gray-700 dark:text-gray-200">欢迎加入交流群！</p>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white p-3 shadow-sm dark:border-dark-700 dark:bg-dark-800 sm:p-4">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">文本模型示例</h2>
            <pre class="mt-3 max-w-full overflow-x-auto whitespace-pre rounded-xl bg-gray-950 p-3 text-[11px] leading-5 text-gray-100 sm:text-xs"><code>{{ chatCurlExample }}</code></pre>
            <button type="button" class="btn btn-secondary mt-3 w-full" @click="copyText(chatCurlExample)">复制文本示例</button>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white p-3 shadow-sm dark:border-dark-700 dark:bg-dark-800 sm:p-4">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">生图模型示例</h2>
            <pre class="mt-3 max-w-full overflow-x-auto whitespace-pre rounded-xl bg-gray-950 p-3 text-[11px] leading-5 text-gray-100 sm:text-xs"><code>{{ imageCurlExample }}</code></pre>
            <button type="button" class="btn btn-secondary mt-3 w-full" @click="copyText(imageCurlExample)">复制生图示例</button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import qqGroupQrImage from '@/assets/images/qrcode-qq-group.jpg'
import { getPublicModelCatalog, DEFAULT_MODEL_CATALOG, type ModelCatalogCategory } from '@/api/modelCatalog'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

interface ModelCard {
  name: string
  platformLabel: string
  tags: string[]
}

interface DisplayCategory {
  id: string
  label: string
  description?: string
  models: ModelCard[]
}

const appStore = useAppStore()
const loading = ref(false)
const searchQuery = ref('')
const categories = ref<ModelCatalogCategory[]>(DEFAULT_MODEL_CATALOG.categories)

type InstallerPlatform = 'unix' | 'windows'
const installerPlatform = ref<InstallerPlatform>('unix')
const unixInstallerCommand = 'curl -fsSL https://sekirocloud.site:8443/sekiro-install.sh | bash'
const windowsInstallerCommand = 'iwr -useb https://sekirocloud.site:8443/sekiro-install.ps1 | iex'
const installerCommand = computed(() => installerPlatform.value === 'windows' ? windowsInstallerCommand : unixInstallerCommand)

const chatCurlExample = `curl https://sekirocloud.site:8443/v1/chat/completions \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-5.5","messages":[{"role":"user","content":"你好"}]}'`

const imageCurlExample = `curl https://sekirocloud.site:8443/v1/images/generations \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-image-2",
    "prompt": "二次元少女，星空背景，唯美氛围感，高清8K",
    "n": 1,
    "size": "1024x1024",
    "response_format": "b64_json"
  }'`

const displayCategories = computed<DisplayCategory[]>(() =>
  categories.value.map(category => ({
    id: category.id,
    label: category.label,
    description: category.description,
    models: category.models.map(name => ({
      name,
      platformLabel: category.label,
      tags: buildTags(name),
    })),
  })),
)

const displayModels = computed(() => displayCategories.value.flatMap(category => category.models))

const filteredCategories = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return displayCategories.value
  return displayCategories.value
    .map(category => ({
      ...category,
      models: category.models.filter(model =>
        model.name.toLowerCase().includes(q) ||
        model.platformLabel.toLowerCase().includes(q) ||
        model.tags.some(tag => tag.toLowerCase().includes(q)),
      ),
    }))
    .filter(category => category.models.length > 0)
})

const filteredModels = computed(() => filteredCategories.value.flatMap(category => category.models))

function buildTags(name: string): string[] {
  const lower = name.toLowerCase()
  const tags: string[] = []
  if (lower.includes('image')) tags.push('图片')
  else tags.push('文本')
  if (lower.includes('codex')) tags.push('编程')
  if (lower.includes('mini')) tags.push('轻量')
  return tags
}

async function loadModels() {
  loading.value = true
  try {
    const catalog = await getPublicModelCatalog()
    categories.value = catalog.categories
  } catch (err: unknown) {
    categories.value = DEFAULT_MODEL_CATALOG.categories
    appStore.showError(extractApiErrorMessage(err, '已显示默认模型列表，模型配置加载失败'))
  } finally {
    loading.value = false
  }
}

async function copyText(text: string) {
  await navigator.clipboard.writeText(text)
  appStore.showSuccess('已复制')
}

onMounted(loadModels)
</script>
