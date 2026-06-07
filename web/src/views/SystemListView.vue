<template>
  <div class="system-list-view">
    <!-- Header -->
    <div class="view-header">
      <div class="header-left">
        <div class="breadcrumb">
          <span class="breadcrumb-item" @click="$router.push('/workspace')">工作台</span>
          <span class="breadcrumb-sep">/</span>
          <span class="breadcrumb-current">系统管理</span>
        </div>
      </div>
      <a-button type="primary" class="new-btn" @click="$router.push('/systems/new')">
        <PlusOutlined />
        新建系统
      </a-button>
    </div>

    <!-- Toolbar: search + filters -->
    <div class="view-toolbar">
      <div class="search-wrap">
        <a-input
          v-model:value="keyword"
          placeholder="搜索系统名称或描述..."
          class="search-input"
          allow-clear
          @press-enter="handleSearch"
          @change="debouncedSearch"
        >
          <template #prefix>
            <SearchOutlined class="search-icon" />
          </template>
        </a-input>
      </div>
      <div class="toolbar-meta">
        <span class="total-count">共 <strong>{{ total }}</strong> 个系统</span>
      </div>
    </div>

    <!-- Table -->
    <div class="view-table-wrap">
      <a-table
        :columns="columns"
        :data-source="systems"
        :loading="loading"
        :pagination="false"
        row-key="id"
        :scroll="{ x: 800 }"
        class="system-table"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <div class="name-cell">
              <span class="system-name">{{ record.name }}</span>
            </div>
          </template>

          <template v-else-if="column.key === 'description'">
            <span class="desc-text">{{ record.description || '—' }}</span>
          </template>

          <template v-else-if="column.key === 'status'">
            <a-tag :color="statusColor[record.status]">{{ statusLabel[record.status] }}</a-tag>
          </template>

          <template v-else-if="column.key === 'created_at'">
            <span class="time-text">{{ formatTime(record.createdAt) }}</span>
          </template>

          <template v-else-if="column.key === 'actions'">
            <div class="action-group">
              <a-button
                type="text"
                size="small"
                class="edit-btn"
                :disabled="editingId === record.id"
                @click="startEdit(record)"
              >
                <EditOutlined />
                编辑
              </a-button>
              <a-popconfirm
                title="确定删除该系统？此操作不可恢复。"
                ok-text="删除"
                cancel-text="取消"
                placement="topRight"
                @confirm="handleDelete(record)"
              >
                <a-button type="text" size="small" danger class="delete-btn">
                  <DeleteOutlined />
                  删除
                </a-button>
              </a-popconfirm>
            </div>
          </template>
        </template>

        <template #emptyText>
          <div class="empty-state">
            <div class="empty-icon">◈</div>
            <div class="empty-title">{{ keyword ? '未找到匹配的系统' : '暂无系统' }}</div>
            <div class="empty-desc">
              {{ keyword ? '尝试更换关键词后重试' : '点击右上角「新建系统」创建一个吧' }}
            </div>
          </div>
        </template>
      </a-table>

      <!-- Inline Edit Row -->
      <div v-if="editingId" class="edit-form-wrap">
        <div class="edit-form-card">
          <div class="edit-form-header">
            <span class="edit-icon">✎</span>
            <span>编辑系统：{{ editingOrigin?.name }}</span>
          </div>
          <div class="edit-form-body">
            <div class="edit-form-row">
              <label class="edit-label">系统名称</label>
              <a-input v-model:value="editForm.name" class="edit-input" placeholder="系统名称" />
            </div>
            <div class="edit-form-row">
              <label class="edit-label">系统描述</label>
              <a-input v-model:value="editForm.description" class="edit-input" placeholder="系统描述（可选）" />
            </div>
            <div class="edit-form-row">
              <label class="edit-label">状态</label>
              <a-select v-model:value="editForm.status" class="edit-select">
                <a-select-option value="active">活跃</a-select-option>
                <a-select-option value="planning">规划中</a-select-option>
                <a-select-option value="archived">已归档</a-select-option>
              </a-select>
            </div>
          </div>
          <div class="edit-form-actions">
            <a-button @click="cancelEdit">取消</a-button>
            <a-button type="primary" :loading="saving" @click="handleSave">保存</a-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div class="view-pagination">
      <a-pagination
        v-model:current="page"
        v-model:pageSize="pageSize"
        :total="total"
        :show-size-changer="true"
        :show-quick-jumper="true"
        :page-size-options="['10', '20', '50', '100']"
        show-size-changer
        @change="handlePageChange"
        @showSizeChange="handleSizeChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons-vue'
import { listSystems, updateSystem, deleteSystem } from '../api/systems'
import { useWorkspaceStore } from '../stores/workspace'

const workspace = useWorkspaceStore()

const keyword = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const saving = ref(false)
const systems = ref([])
const editingId = ref('')
const editingOrigin = ref(null)
const editForm = reactive({ name: '', description: '', status: 'active' })

let searchTimer = null

const columns = [
  { title: '系统名称', key: 'name', width: 200, fixed: 'left' },
  { title: '描述', key: 'description', minWidth: 200 },
  { title: '状态', key: 'status', width: 110 },
  { title: '创建时间', key: 'created_at', width: 160 },
  { title: '操作', key: 'actions', width: 160, fixed: 'right' },
]

const statusColor = { active: 'green', planning: 'orange', archived: 'default' }
const statusLabel = { active: '活跃', planning: '规划中', archived: '已归档' }

const formatTime = (ts) => {
  if (!ts) return '—'
  const d = new Date(ts)
  if (isNaN(d)) return '—'
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const loadSystems = async () => {
  loading.value = true
  try {
    const data = await listSystems(workspace.token || localStorage.getItem('token') || '', {
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
    })
    systems.value = (data.items || []).map((s) => ({
      ...s,
      id: s.ID || s.id,
      name: s.Name || s.name,
      description: s.Description || s.description,
      status: s.Status || s.status,
      createdAt: s.CreatedAt || s.createdAt,
      updatedAt: s.UpdatedAt || s.updatedAt,
    }))
    total.value = data.total || 0
    workspace.setSystems(data.items || [])
  } catch (err) {
    message.error('加载系统列表失败: ' + (err?.message || err))
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadSystems()
}

const debouncedSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    handleSearch()
  }, 400)
}

const handlePageChange = (p) => {
  page.value = p
  loadSystems()
}

const handleSizeChange = (current, size) => {
  pageSize.value = size
  page.value = 1
  loadSystems()
}

const startEdit = (record) => {
  editingId.value = record.id
  editingOrigin.value = record
  editForm.name = record.name
  editForm.description = record.description || ''
  editForm.status = record.status || 'active'
}

const cancelEdit = () => {
  editingId.value = ''
  editingOrigin.value = null
}

const handleSave = async () => {
  if (!editForm.name?.trim()) {
    message.warning('系统名称不能为空')
    return
  }
  saving.value = true
  try {
    await updateSystem(workspace.token || localStorage.getItem('token') || '', editingId.value, {
      name: editForm.name.trim(),
      description: editForm.description.trim(),
      status: editForm.status,
    })
    message.success('系统信息已更新')
    cancelEdit()
    await loadSystems()
  } catch (err) {
    message.error('保存失败: ' + (err?.message || err))
  } finally {
    saving.value = false
  }
}

const handleDelete = async (record) => {
  try {
    await deleteSystem(workspace.token || localStorage.getItem('token') || '', record.id)
    message.success('系统已删除')
    if (workspace.selectedSystemId === record.id) {
      workspace.selectSystem('')
    }
    await loadSystems()
  } catch (err) {
    message.error('删除失败: ' + (err?.message || err))
  }
}

onMounted(loadSystems)
</script>

<style scoped>
.system-list-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  background: var(--core-bg);
  border-radius: 12px;
  overflow: hidden;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px 16px;
  background: var(--core-surface);
  border-bottom: 1px solid var(--core-border);
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.breadcrumb-item {
  color: var(--core-text-secondary);
  cursor: pointer;
  transition: color 150ms;
}

.breadcrumb-item:hover {
  color: var(--accent-primary);
}

.breadcrumb-sep {
  color: var(--core-text-muted);
}

.breadcrumb-current {
  color: var(--core-text);
  font-weight: 600;
}

.new-btn {
  height: 36px;
  padding: 0 18px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-info));
  border: none;
  color: #fff;
  font-weight: 600;
  font-family: var(--font-display);
  box-shadow: 0 6px 18px rgba(0, 212, 255, 0.28);
  transition: all var(--transition-base);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.new-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 24px rgba(0, 212, 255, 0.38);
}

.view-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 24px;
  background: var(--core-surface);
  border-bottom: 1px solid var(--core-border);
}

.search-wrap {
  flex: 1;
  max-width: 400px;
}

.search-input {
  border-radius: var(--radius-md) !important;
  border-color: var(--core-border-accent) !important;
  background: var(--core-bg) !important;
  font-family: var(--font-display);
}

.search-input :deep(.ant-input) {
  background: transparent !important;
  color: var(--core-text) !important;
}

.search-input :deep(.ant-input::placeholder) {
  color: var(--core-text-muted) !important;
}

.search-input :deep(.ant-input-prefix) {
  color: var(--core-text-muted) !important;
  margin-right: 8px;
}

.search-icon {
  font-size: 15px;
}

.toolbar-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.total-count {
  font-size: 13px;
  color: var(--core-text-secondary);
  white-space: nowrap;
}

.total-count strong {
  color: var(--core-text);
  font-weight: 600;
}

.view-table-wrap {
  flex: 1;
  overflow: auto;
  padding: 16px 24px 0;
}

.system-table {
  background: var(--core-surface);
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--core-border);
}

.system-table :deep(.ant-table) {
  background: transparent;
  color: var(--core-text);
}

.system-table :deep(.ant-table-thead > tr > th) {
  background: #f0f3f8 !important;
  color: var(--core-text) !important;
  font-weight: 600;
  font-size: 13px;
  border-bottom: 1px solid var(--core-border) !important;
  font-family: var(--font-display);
}

.system-table :deep(.ant-table-tbody > tr > td) {
  border-bottom: 1px solid var(--core-border) !important;
  color: var(--core-text);
  font-size: 14px;
}

.system-table :deep(.ant-table-tbody > tr:hover > td) {
  background: #f8faff !important;
}

.system-table :deep(.ant-table-wrapper .ant-table-pagination) {
  margin: 16px 0 0;
}

.name-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.system-name {
  font-weight: 600;
  color: var(--core-text);
  font-size: 14px;
}

.desc-text {
  font-size: 13px;
  color: var(--core-text-secondary);
}

.time-text {
  font-size: 13px;
  color: var(--core-text-secondary);
  font-family: var(--font-mono);
}

.action-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.edit-btn {
  color: var(--accent-primary) !important;
  font-size: 13px;
  font-weight: 500;
}

.edit-btn:hover {
  background: rgba(0, 212, 255, 0.08) !important;
}

.delete-btn {
  color: var(--accent-danger) !important;
  font-size: 13px;
  font-weight: 500;
}

.delete-btn:hover {
  background: rgba(239, 68, 68, 0.08) !important;
}

.empty-state {
  padding: 48px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.empty-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.12), rgba(124, 58, 237, 0.12));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--accent-primary);
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--core-text);
}

.empty-desc {
  font-size: 13px;
  color: var(--core-text-secondary);
}

.edit-form-wrap {
  margin-top: 16px;
}

.edit-form-card {
  background: var(--core-surface);
  border: 1px solid var(--core-border-accent);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.1);
  animation: slideUp 0.25s ease-out;
}

.edit-form-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: rgba(0, 212, 255, 0.06);
  border-bottom: 1px solid var(--core-border);
  font-size: 14px;
  font-weight: 600;
  color: var(--core-text);
}

.edit-icon {
  color: var(--accent-primary);
  font-size: 16px;
}

.edit-form-body {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.edit-form-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.edit-label {
  width: 80px;
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 500;
  color: var(--core-text-secondary);
  text-align: right;
}

.edit-input,
.edit-select {
  flex: 1;
}

.edit-input {
  border-radius: var(--radius-md) !important;
  border-color: var(--core-border-accent) !important;
  background: var(--core-bg) !important;
}

.edit-input :deep(.ant-input) {
  background: transparent !important;
  color: var(--core-text) !important;
}

.edit-form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 18px;
  border-top: 1px solid var(--core-border);
}

.view-pagination {
  padding: 16px 24px 20px;
  background: var(--core-surface);
  border-top: 1px solid var(--core-border);
  display: flex;
  justify-content: flex-end;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 768px) {
  .view-header,
  .view-toolbar,
  .view-table-wrap,
  .view-pagination {
    padding-left: 16px;
    padding-right: 16px;
  }

  .view-toolbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .search-wrap {
    max-width: 100%;
    width: 100%;
  }

  .edit-form-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .edit-label {
    width: auto;
    text-align: left;
  }
}
</style>
