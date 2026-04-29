<template>
  <AppLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">模型配置</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          按厂商统一维护对用户展示的模型，不需要在每个账号里重复解析。
        </p>
      </div>

    <div class="rounded-xl border border-blue-100 bg-blue-50 p-4 text-sm text-blue-800 dark:border-blue-900/60 dark:bg-blue-950/30 dark:text-blue-200">
      <div class="font-medium">用法说明</div>
      <div class="mt-1 leading-6">
        这里配置的是用户侧“模型广场”的展示目录。账号编辑里的模型白名单/映射仍然是账号级调度/限制能力，不作为对外模型总目录。
      </div>
    </div>

    <div v-if="loading" class="rounded-xl border border-gray-200 bg-white p-6 text-center text-gray-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400">
      加载中...
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="(category, index) in draft.categories"
        :key="index"
        class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-800"
      >
        <div class="grid gap-3 md:grid-cols-2">
          <label class="block">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">厂商分类名称</span>
            <input
              v-model="category.label"
              type="text"
              class="input mt-1"
              placeholder="例如：OpenAI、Claude、Gemini"
            >
          </label>
          <label class="block">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">分类标识</span>
            <input
              v-model="category.id"
              type="text"
              class="input mt-1"
              placeholder="openai"
            >
          </label>
        </div>

        <label class="mt-3 block">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">分类说明</span>
          <input
            v-model="category.description"
            type="text"
            class="input mt-1"
            placeholder="例如：此分类下对用户展示的模型。"
          >
        </label>

        <div class="mt-3">
          <div class="mb-1 flex items-center justify-between gap-2">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">模型列表</span>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ category.models.length }} 个模型</span>
          </div>
          <textarea
            :value="modelInputs[index] || ''"
            rows="7"
            class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 font-mono text-sm text-gray-900 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20 dark:border-gray-600 dark:bg-gray-900 dark:text-gray-100"
            placeholder="一行一个模型名，也可以粘贴 /v1/models JSON 或逗号分隔文本"
            @input="setModelsText(index, ($event.target as HTMLTextAreaElement).value)"
          />
        </div>

        <div class="mt-3 flex flex-wrap gap-2">
          <button type="button" class="btn btn-danger" @click="removeCategory(index)">
            删除分类
          </button>
        </div>
      </div>

      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <button type="button" class="btn btn-secondary border-dashed" @click="addCategory">
          + 新增模型配置
        </button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="saving"
          @click="saveCatalog"
        >
          {{ saving ? '保存中...' : '保存模型配置' }}
        </button>
      </div>
    </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { getAdminModelCatalog, updateAdminModelCatalog, type ModelCatalogConfig } from '@/api/modelCatalog'
import { parseBulkModelInput } from '@/composables/useModelWhitelist'
import { useAppStore } from '@/stores/app'

const loading = ref(true)
const saving = ref(false)
const appStore = useAppStore()
const draft = reactive<ModelCatalogConfig>({ categories: [] })
const modelInputs = ref<string[]>([])

function replaceDraft(config: ModelCatalogConfig) {
  draft.categories.splice(0, draft.categories.length, ...config.categories.map(category => ({
    id: category.id,
    label: category.label,
    description: category.description || '',
    models: [...category.models],
  })))
  modelInputs.value = draft.categories.map(category => category.models.join('\n'))
}

function setModelsText(index: number, raw: string) {
  modelInputs.value[index] = raw
}

function parseModels(index: number) {
  const parsed = parseBulkModelInput(modelInputs.value[index] || '')
  draft.categories[index].models = parsed
  modelInputs.value[index] = parsed.join('\n')
}

function addCategory() {
  draft.categories.push({
    id: `provider-${draft.categories.length + 1}`,
    label: '新厂商分类',
    description: '',
    models: [],
  })
  modelInputs.value.push('')
}

function removeCategory(index: number) {
  draft.categories.splice(index, 1)
  modelInputs.value.splice(index, 1)
  if (draft.categories.length === 0) addCategory()
}

async function loadCatalog() {
  loading.value = true
  try {
    replaceDraft(await getAdminModelCatalog())
  } finally {
    loading.value = false
  }
}

async function saveCatalog() {
  saving.value = true
  try {
    draft.categories.forEach((_, index) => parseModels(index))
    const saved = await updateAdminModelCatalog(draft)
    replaceDraft(saved)
    appStore.showSuccess('模型配置已保存，用户侧模型广场会按这里展示')
  } catch (error) {
    appStore.showError('保存模型配置失败')
  } finally {
    saving.value = false
  }
}

onMounted(loadCatalog)
</script>
