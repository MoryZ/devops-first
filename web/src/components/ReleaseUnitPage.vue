<template>
  <div class="release-unit-page">
    <div class="unit-toolbar">
      <div>
        <div class="unit-title">发布单元管理</div>
        <div class="unit-subtitle">按系统维护发布单元，供流水线发布阶段复用</div>
      </div>
      <a-button type="primary" @click="showCreateModal"><PlusOutlined /> 添加发布单元</a-button>
    </div>

    <a-alert
      v-if="!workspace.selectedSystemId"
      type="info"
      show-icon
      message="请先在上方选择一个系统，再创建发布单元"
      class="unit-alert"
    />

    <div class="unit-list-wrap">
      <a-empty v-if="units.length === 0" description="当前系统暂无发布单元" />

      <a-table
        v-else
        :data-source="sortedUnits"
        :pagination="false"
        :row-key="(row) => row.id"
        size="small"
      >
        <a-table-column title="顺序" data-index="displayOrder" key="displayOrder" width="70" />
        <a-table-column title="发布单元" data-index="name" key="name" />
        <a-table-column title="应用模块" data-index="module" key="module" />
        <a-table-column title="触发源类型" data-index="repositoryType" key="repositoryType" width="110" />
        <a-table-column title="代码库地址" data-index="repoUrl" key="repoUrl" />
        <a-table-column title="默认分支" data-index="branch" key="branch" />
        <a-table-column title="自动归并" key="autoMerge" width="95">
          <template #default="{ record }">
            <a-tag :color="record.autoMerge ? 'blue' : 'default'">{{ record.autoMerge ? 'ON' : 'OFF' }}</a-tag>
          </template>
        </a-table-column>
        <a-table-column title="自动Tag" key="autoTag" width="95">
          <template #default="{ record }">
            <a-tag :color="record.autoTag ? 'geekblue' : 'default'">{{ record.autoTag ? 'ON' : 'OFF' }}</a-tag>
          </template>
        </a-table-column>
        <a-table-column title="状态" key="status">
          <template #default="{ record }">
            <a-tag :color="record.enabled ? 'green' : 'default'">{{ record.enabled ? '启用' : '停用' }}</a-tag>
          </template>
        </a-table-column>
        <a-table-column title="操作" key="actions" align="right" width="230">
          <template #default="{ record }">
            <a-space>
              <a-button type="link" size="small" @click="showEditModal(record)">编辑</a-button>
              <a-button type="link" size="small" @click="toggleEnabled(record)">
                {{ record.enabled ? '停用' : '启用' }}
              </a-button>
              <a-button type="link" danger size="small" @click="removeUnit(record.id)">删除</a-button>
            </a-space>
          </template>
        </a-table-column>
      </a-table>
    </div>

    <a-modal
      v-model:open="createModalOpen"
      :title="isEditing ? '编辑发布单元' : '新增发布单元'"
      cancel-text="取消"
      ok-text="确认"
      :confirm-loading="creating"
      @ok="handleSaveUnit"
    >
      <a-form layout="vertical">
        <a-form-item label="发布单元名称">
          <a-input v-model:value="newUnit.name" placeholder="例如: marketing-domain" />
        </a-form-item>
        <a-form-item label="应用模块">
          <a-input v-model:value="newUnit.module" placeholder="例如: marketing-domain" />
        </a-form-item>
        <a-form-item label="代码库地址">
          <a-input v-model:value="newUnit.repoUrl" placeholder="https://git.example.com/group/repo.git" />
        </a-form-item>
        <a-form-item label="默认分支">
          <a-input v-model:value="newUnit.branch" placeholder="main / release-xxx" />
        </a-form-item>
        <a-form-item label="触发源类型">
          <a-select v-model:value="newUnit.repositoryType" placeholder="选择触发源类型">
            <a-select-option value="git">git</a-select-option>
            <a-select-option value="svn">svn</a-select-option>
            <a-select-option value="artifact">artifact</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="展示顺序">
          <a-input-number v-model:value="newUnit.displayOrder" :min="1" style="width: 100%" />
        </a-form-item>
        <a-form-item label="自动代码归并">
          <a-switch v-model:checked="newUnit.autoMerge" checked-children="ON" un-checked-children="OFF" />
        </a-form-item>
        <a-form-item label="自动打Tag">
          <a-switch v-model:checked="newUnit.autoTag" checked-children="ON" un-checked-children="OFF" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { useWorkspaceStore } from '../stores/workspace'

const workspace = useWorkspaceStore()

const units = ref([])
const createModalOpen = ref(false)
const creating = ref(false)
const editingUnitId = ref('')
const newUnit = ref({
  name: '',
  module: '',
  repoUrl: '',
  branch: 'main',
  repositoryType: 'git',
  autoMerge: true,
  autoTag: true,
  displayOrder: 1,
})

const isEditing = computed(() => Boolean(editingUnitId.value))

const sortedUnits = computed(() => {
  return [...units.value].sort((a, b) => {
    const orderA = Number.isFinite(Number(a.displayOrder)) ? Number(a.displayOrder) : 9999
    const orderB = Number.isFinite(Number(b.displayOrder)) ? Number(b.displayOrder) : 9999
    if (orderA !== orderB) return orderA - orderB
    return (b.createdAt || 0) - (a.createdAt || 0)
  })
})

const getStorageKey = () => `releaseUnits:${workspace.selectedSystemId || 'none'}`

const loadUnits = () => {
  if (!workspace.selectedSystemId) {
    units.value = []
    return
  }
  try {
    const raw = localStorage.getItem(getStorageKey())
    const parsed = raw ? JSON.parse(raw) : []
    units.value = Array.isArray(parsed)
      ? parsed.map((item, index) => ({
          ...item,
          repositoryType: item.repositoryType || 'git',
          autoMerge: item.autoMerge ?? true,
          autoTag: item.autoTag ?? true,
          displayOrder: Number(item.displayOrder) || index + 1,
        }))
      : []
  } catch {
    units.value = []
  }
}

const saveUnits = () => {
  localStorage.setItem(getStorageKey(), JSON.stringify(units.value))
}

watch(
  () => workspace.selectedSystemId,
  () => {
    loadUnits()
  },
  { immediate: true }
)

const showCreateModal = () => {
  if (!workspace.selectedSystemId) {
    message.warning('请先选择系统')
    return
  }
  editingUnitId.value = ''
  newUnit.value = {
    name: '',
    module: '',
    repoUrl: '',
    branch: 'main',
    repositoryType: 'git',
    autoMerge: true,
    autoTag: true,
    displayOrder: units.value.length + 1,
  }
  createModalOpen.value = true
}

const showEditModal = (record) => {
  editingUnitId.value = record.id
  newUnit.value = {
    name: record.name || '',
    module: record.module || '',
    repoUrl: record.repoUrl || '',
    branch: record.branch || 'main',
    repositoryType: record.repositoryType || 'git',
    autoMerge: record.autoMerge ?? true,
    autoTag: record.autoTag ?? true,
    displayOrder: Number(record.displayOrder) || 1,
  }
  createModalOpen.value = true
}

const handleSaveUnit = () => {
  if (!newUnit.value.name.trim()) {
    message.warning('请输入发布单元名称')
    return
  }
  if (!newUnit.value.module.trim()) {
    message.warning('请输入应用模块')
    return
  }
  if (!newUnit.value.repoUrl.trim()) {
    message.warning('请输入代码库地址')
    return
  }
  if (!Number.isFinite(Number(newUnit.value.displayOrder)) || Number(newUnit.value.displayOrder) < 1) {
    message.warning('展示顺序必须是大于等于 1 的数字')
    return
  }

  creating.value = true
  try {
    const editMode = Boolean(editingUnitId.value)
    if (editingUnitId.value) {
      units.value = units.value.map((item) => {
        if (item.id !== editingUnitId.value) return item
        return {
          ...item,
          name: newUnit.value.name.trim(),
          module: newUnit.value.module.trim(),
          repoUrl: newUnit.value.repoUrl.trim(),
          branch: newUnit.value.branch.trim() || 'main',
          repositoryType: newUnit.value.repositoryType,
          autoMerge: Boolean(newUnit.value.autoMerge),
          autoTag: Boolean(newUnit.value.autoTag),
          displayOrder: Number(newUnit.value.displayOrder),
        }
      })
    } else {
      units.value.unshift({
        id: `${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
        name: newUnit.value.name.trim(),
        module: newUnit.value.module.trim(),
        repoUrl: newUnit.value.repoUrl.trim(),
        branch: newUnit.value.branch.trim() || 'main',
        repositoryType: newUnit.value.repositoryType,
        autoMerge: Boolean(newUnit.value.autoMerge),
        autoTag: Boolean(newUnit.value.autoTag),
        displayOrder: Number(newUnit.value.displayOrder),
        enabled: true,
        createdAt: Date.now(),
      })
    }
    saveUnits()
    createModalOpen.value = false
    editingUnitId.value = ''
    message.success(editMode ? '发布单元已更新' : '发布单元已创建')
  } finally {
    creating.value = false
  }
}

const toggleEnabled = (record) => {
  record.enabled = !record.enabled
  saveUnits()
}

const removeUnit = (id) => {
  units.value = units.value.filter((item) => item.id !== id)
  saveUnits()
  message.success('发布单元已删除')
}
</script>

<style scoped>
.release-unit-page {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #dfe7f4;
  padding: 16px;
}

.unit-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 14px;
}

.unit-title {
  font-size: 20px;
  font-weight: 700;
  color: #1c2d42;
}

.unit-subtitle {
  margin-top: 4px;
  color: #6d7b95;
  font-size: 13px;
}

.unit-alert {
  margin-bottom: 12px;
}

.unit-list-wrap {
  background: #f8fbff;
  border: 1px solid #e6edf8;
  border-radius: 10px;
  padding: 10px;
}
</style>
