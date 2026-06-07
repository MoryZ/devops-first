<template>
  <div class="system-selector-wrap" :class="{ compact }">
    <div class="system-selector">
      <span v-if="!compact" class="selector-label">当前系统</span>
      <a-select
        :value="currentSystemId"
        class="system-select"
        placeholder="请选择系统"
        :loading="loadingSystems"
        show-search
        option-label-prop="label"
        option-filter-prop="label"
        :filter-option="filterSystemOption"
        @change="handleSystemChange"
      >
        <a-select-option
          v-for="system in sortedSystems"
          :key="system.id"
          :value="system.id"
          :label="getSystemPrimary(system)"
          :title="`${getSystemPrimary(system)} ${getSystemSecondary(system)}`"
        >
          <div class="system-option">
            <div class="option-text">
              <span class="option-primary" :title="getSystemPrimary(system)">{{ getSystemPrimary(system) }}</span>
              <span class="option-secondary" :title="getSystemSecondary(system)">
                {{ getSystemSecondary(system) }}
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
        <a-select-option
          value="__new_system__"
          label="＋ 新建系统"
          class="new-system-option"
        >
          <div class="new-system-option-content">
            <PlusOutlined />
            <span>新建系统</span>
          </div>
        </a-select-option>
      </a-select>
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
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { StarFilled, StarOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { useWorkspaceStore } from '../stores/workspace'
import { listSystems } from '../api/systems'

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
const router = useRouter()

const loadingSystems = ref(false)
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

const getSystemPrimary = (system) => {
  return system?.description?.trim() || system?.name || `ID: ${String(system?.id || '').slice(0, 8)}`
}

const getSystemSecondary = (system) => {
  const name = system?.name || `ID: ${String(system?.id || '').slice(0, 8)}`
  const description = system?.description?.trim() || name
  return `${name}_${description}`
}

const toggleFavorite = (id) => {
  if (favoriteSystemIds.value.has(id)) {
    favoriteSystemIds.value.delete(id)
  } else {
    favoriteSystemIds.value.add(id)
  }
  saveFavorites()
}

const filterSystemOption = (input, option) => {
  const text = `${option?.label || ''} ${option?.title || ''} ${option?.value || ''}`.toLowerCase()
  return text.includes(input.toLowerCase())
}

const loadSystems = async () => {
  loadingSystems.value = true
  try {
    const data = await listSystems(props.token)
    workspace.setSystems(data.items || [])
  } catch (err) {
    console.warn('Failed to load systems:', err)
  } finally {
    loadingSystems.value = false
  }
}

const handleSystemChange = (value) => {
  if (value === '__new_system__') {
    router.push('/systems/new')
    currentSystemId.value = workspace.selectedSystemId
    return
  }
  workspace.selectSystem(value)
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
  gap: 12px;
  padding: 10px 14px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-tertiary);
  transition: all var(--transition-base);
}

.system-selector:hover {
  border-color: var(--border-color-accent);
}

.system-selector-wrap.compact .system-selector {
  padding: 8px;
  border: none;
  background: transparent;
  box-shadow: none;
  gap: 10px;
}

.system-selector-wrap.compact .selector-label {
  color: var(--text-secondary);
  font-size: 14px;
  white-space: nowrap;
  font-weight: 500;
}

.system-selector-wrap.compact :deep(.ant-select) {
  min-width: 240px;
}

.system-selector-wrap.compact :deep(.ant-select-selector) {
  background: rgba(30, 41, 59, 0.8) !important;
  border-color: var(--border-color-light) !important;
  border-radius: var(--radius-md) !important;
  color: var(--text-primary) !important;
}

.system-selector-wrap.compact :deep(.ant-select-arrow) {
  color: var(--text-tertiary) !important;
}

.system-selector-wrap.compact :deep(.ant-select-selection-item) {
  color: var(--text-primary) !important;
}

.system-selector-wrap.compact :deep(.ant-select-selection-placeholder) {
  color: var(--text-muted) !important;
}

.system-selector-wrap.compact :deep(.ant-select-selection-item .option-secondary),
.system-selector-wrap.compact :deep(.ant-select-selection-item .option-star-btn) {
  display: none;
}

.system-selector-wrap.compact :deep(.ant-select-selection-item .system-option) {
  justify-content: flex-start;
}

.system-selector-wrap.compact :deep(.ant-select-selection-item .option-primary) {
  max-width: 140px;
  font-size: 14px;
  color: var(--text-primary) !important;
}

.system-selector-wrap.compact :deep(.ant-select) {
  font-size: 14px;
}

.system-selector-wrap.compact :deep(.ant-select-selection-item) {
  font-size: 14px;
}

.new-system-option {
  color: var(--accent-primary);
}

.new-system-option-content {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.selected-system-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-radius: var(--radius-md);
  padding: 12px 14px;
  border: 1px solid rgba(0, 212, 255, 0.3);
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
  backdrop-filter: blur(8px);
  animation: slideUp 0.3s ease-out;
}

.system-card-main {
  min-width: 0;
}

.system-name-line {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.system-name {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 14px;
  max-width: 320px;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.system-subline {
  color: var(--text-tertiary);
  font-size: 12px;
  min-width: 0;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.star-btn {
  color: var(--text-secondary);
  transition: color var(--transition-fast);
}

.star-btn:hover {
  color: var(--accent-warning);
}

.star-on {
  color: var(--accent-warning);
}

.star-off {
  color: var(--text-tertiary);
}

.selector-label {
  color: var(--text-secondary);
  font-size: 13px;
  min-width: 56px;
  font-weight: 500;
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
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.option-primary {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  max-width: 100%;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.option-secondary {
  font-size: 12px;
  color: var(--text-tertiary);
  min-width: 0;
  max-width: 100%;
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
  transition: transform var(--transition-fast);
}

.option-star-btn:hover {
  transform: scale(1.2);
}

.option-star {
  font-size: 14px;
}

.option-star-on {
  color: var(--accent-warning);
}

.option-star-off {
  color: var(--text-muted);
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
