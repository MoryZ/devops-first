<template>
  <div class="system-selector-wrap" :class="{ compact }">
    <div class="system-selector">
      <span v-if="!compact" class="selector-label">当前系统</span>
      <a-select
        v-model:value="currentSystemId"
        class="system-select"
        placeholder="请选择系统"
        :loading="loadingSystems"
        show-search
        option-filter-prop="label"
        :filter-option="filterSystemOption"
      >
        <a-select-option
          v-for="system in sortedSystems"
          :key="system.id"
          :value="system.id"
          :label="`${system.name} · ${system.description || ''}`"
        >
          <div class="system-option">
            <div class="option-text">
              <span class="option-main" :title="system.name">{{ system.name }}</span>
              <span class="option-sub" :title="system.description || `ID: ${system.id.slice(0, 8)}`">
                {{ system.description || `ID: ${system.id.slice(0, 8)}` }}
              </span>
            </div>
            <button
              class="option-star-btn"
              type="button"
              :aria-label="isFavorite(system.id) ? '取消收藏' : '收藏系统'"
              @click.stop="toggleFavorite(system.id)"
            >
              <StarFilled v-if="isFavorite(system.id)" class="option-star option-star-on" />
              <StarOutlined v-else class="option-star option-star-off" />
            </button>
          </div>
        </a-select-option>
      </a-select>
      <a-button v-if="!compact" type="primary" class="new-system-btn" :loading="creating" @click="showCreateModal">
        <PlusOutlined />
        新建系统
      </a-button>
    </div>

    <div v-if="selectedSystem && !compact" class="selected-system-card">
      <div class="system-card-main">
        <div class="system-name-line">
          <span class="system-name" :title="selectedSystem.name">{{ selectedSystem.name }}</span>
          <span class="system-subline" :title="selectedSystem.description || '暂无系统描述'">
            {{ selectedSystem.description || '暂无系统描述' }}
          </span>
          <a-button type="text" size="small" class="star-btn" @click="toggleFavorite(selectedSystem.id)">
            <StarFilled v-if="isFavorite(selectedSystem.id)" class="star-on" />
            <StarOutlined v-else class="star-off" />
          </a-button>
        </div>
      </div>
      <a-tag color="blue">{{ selectedSystem.status || 'active' }}</a-tag>
    </div>
    <a-modal
      v-model:open="createModalOpen"
      title="新建系统"
      :confirm-loading="creating"
      @ok="handleCreateSystem"
    >
      <a-form layout="vertical">
        <a-form-item label="系统名称">
          <a-input v-model:value="newSystem.name" placeholder="输入系统名称" />
        </a-form-item>
        <a-form-item label="系统描述">
          <a-textarea v-model:value="newSystem.description" placeholder="系统描述（可选）" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, StarFilled, StarOutlined } from '@ant-design/icons-vue'
import { useWorkspaceStore } from '../stores/workspace'

const props = defineProps({
  token: {
    type: String,
    required: true,
  },
  compact: {
    type: Boolean,
    default: false,
  },
})

const workspace = useWorkspaceStore()

const loadingSystems = ref(false)
const creating = ref(false)
const createModalOpen = ref(false)
const newSystem = ref({ name: '', description: '' })

const currentSystemId = computed({
  get: () => workspace.selectedSystemId,
  set: (value) => workspace.selectSystem(value),
})

const selectedSystem = computed(() => workspace.selectedSystem)
const favoriteSystemIds = ref(new Set())

const sortedSystems = computed(() => {
  return [...workspace.systems].sort((a, b) => {
    const favA = favoriteSystemIds.value.has(a.id) ? 1 : 0
    const favB = favoriteSystemIds.value.has(b.id) ? 1 : 0
    if (favA !== favB) return favB - favA
    return a.name.localeCompare(b.name)
  })
})

const loadFavorites = () => {
  try {
    const raw = localStorage.getItem('favoriteSystemIds')
    const arr = raw ? JSON.parse(raw) : []
    favoriteSystemIds.value = new Set(Array.isArray(arr) ? arr : [])
  } catch {
    favoriteSystemIds.value = new Set()
  }
}

const saveFavorites = () => {
  localStorage.setItem('favoriteSystemIds', JSON.stringify(Array.from(favoriteSystemIds.value)))
}

const isFavorite = (id) => favoriteSystemIds.value.has(id)

const toggleFavorite = (id) => {
  if (favoriteSystemIds.value.has(id)) {
    favoriteSystemIds.value.delete(id)
  } else {
    favoriteSystemIds.value.add(id)
  }
  saveFavorites()
}

const filterSystemOption = (input, option) => {
  const text = `${option?.label || ''} ${option?.value || ''}`.toLowerCase()
  return text.includes(input.toLowerCase())
}

const parseErrorText = async (res, fallback) => {
  const text = await res.text()
  if (!text) return fallback
  try {
    const json = JSON.parse(text)
    return json.error || fallback
  } catch {
    return text
  }
}

const requestWithTimeout = async (url, options = {}, timeoutMs = 10000) => {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs)
  try {
    return await fetch(url, { ...options, signal: controller.signal })
  } finally {
    clearTimeout(timeoutId)
  }
}

const loadSystems = async () => {
  loadingSystems.value = true
  try {
    const res = await requestWithTimeout('/api/systems', {
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (!res.ok) {
      const errMsg = await parseErrorText(res, '系统列表加载失败')
      message.error(errMsg)
      return
    }
    const data = await res.json()
    workspace.setSystems(data.items || [])
  } catch (err) {
    const msg = err?.name === 'AbortError' ? '请求超时，请检查前后端服务状态后重试' : err.message
    message.error('系统列表加载失败: ' + msg)
  } finally {
    loadingSystems.value = false
  }
}

const showCreateModal = () => {
  createModalOpen.value = true
  newSystem.value = { name: '', description: '' }
}

const handleCreateSystem = async () => {
  if (!newSystem.value.name.trim()) {
    message.warning('请输入系统名称')
    return
  }

  creating.value = true
  try {
    const res = await requestWithTimeout('/api/systems', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${props.token}`,
      },
      body: JSON.stringify(newSystem.value),
    })

    if (!res.ok) {
      const errMsg = await parseErrorText(res, '创建系统失败')
      message.error(errMsg)
      return
    }

    const created = await res.json()
    await loadSystems()
    workspace.selectSystem(created.ID || created.id)
    createModalOpen.value = false
    message.success('系统创建成功')
  } catch (err) {
    const msg = err?.name === 'AbortError' ? '请求超时，请检查前后端服务状态后重试' : err.message
    message.error('创建系统失败: ' + msg)
  } finally {
    creating.value = false
  }
}

onMounted(loadSystems)
onMounted(loadFavorites)
</script>

<style scoped>
.system-selector-wrap {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.system-selector-wrap.compact {
  width: auto;
}

.system-selector {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border-radius: 12px;
  border: 1px solid #dce5f2;
  background: #fff;
  box-shadow: 0 2px 10px rgba(13, 32, 58, 0.06);
}

.system-selector-wrap.compact .system-selector {
  padding: 0;
  border: none;
  background: transparent;
  box-shadow: none;
  gap: 0;
}

.system-selector-wrap.compact .selector-label {
  color: #d7e5ff;
  font-size: 12px;
  white-space: nowrap;
}

.system-selector-wrap.compact :deep(.ant-select) {
  width: 190px !important;
}

.system-selector-wrap.compact :deep(.ant-select-selector) {
  background: rgba(255, 255, 255, 0.92) !important;
  border-color: rgba(184, 206, 246, 0.5) !important;
  border-radius: 8px !important;
}

.system-selector-wrap.compact :deep(.ant-select-arrow) {
  color: #4f617c;
}

/* 选中态只显示系统名，隐藏 sub + 星星 */
.system-selector-wrap.compact :deep(.ant-select-selection-item .option-sub),
.system-selector-wrap.compact :deep(.ant-select-selection-item .option-star-btn) {
  display: none;
}

.system-selector-wrap.compact :deep(.ant-select-selection-item .system-option) {
  justify-content: flex-start;
}

.system-selector-wrap.compact :deep(.ant-select-selection-item .option-main) {
  max-width: 140px;
  font-size: 13px;
}

.system-selector-wrap.compact .new-system-btn {
  height: 30px;
  padding: 0 10px;
}

.selected-system-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-radius: 12px;
  padding: 10px 12px;
  border: 1px solid #cfe0ff;
  background: linear-gradient(135deg, #1e67da 0%, #2a7aee 100%);
}

.system-card-main {
  min-width: 0;
}

.system-name-line {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.system-name {
  color: #ffffff;
  font-weight: 600;
  font-size: 14px;
  max-width: 320px;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.system-subline {
  color: #dbe8ff;
  font-size: 12px;
  min-width: 0;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.star-btn {
  color: #fff;
}

.star-on {
  color: #ffd666;
}

.star-off {
  color: #dbe8ff;
}

.selector-label {
  color: #4f617c;
  font-size: 13px;
  min-width: 56px;
}

.system-select {
  width: 520px;
}

:deep(.system-select .ant-select-selection-item) {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.system-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.option-text {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
}

.option-main {
  font-size: 13px;
  font-weight: 600;
  color: #1f2d3d;
  max-width: 220px;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.option-sub {
  font-size: 12px;
  color: #7e8ca6;
  min-width: 0;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.option-star-btn {
  border: 0;
  background: transparent;
  padding: 0;
  line-height: 1;
  cursor: pointer;
}

.option-star {
  font-size: 14px;
}

.option-star-on {
  color: #f8b800;
}

.option-star-off {
  color: #b7c5dd;
}

.new-system-btn {
  border-radius: 8px;
}

@media (max-width: 900px) {
  .system-selector {
    width: 100%;
    flex-wrap: wrap;
  }

  .system-select {
    width: 100%;
  }
}
</style>
