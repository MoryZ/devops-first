<template>
  <PageCard title="执行历史" subtitle="查看流水线部署记录和节点状态" icon="⏱">
    <template #actions>
      <a-range-picker
        v-model:value="dateRange"
        format="YYYY-MM-DD"
        :placeholder="['开始日期', '结束日期']"
        size="small"
        class="filter-picker"
        @change="applyFilters"
      />
      <a-select
        v-model:value="selectedOperator"
        placeholder="全部操作人"
        size="small"
        class="filter-select"
        allow-clear
        @change="applyFilters"
      >
        <a-select-option value="">全部</a-select-option>
        <a-select-option v-for="name in operatorOptions" :key="name" :value="name">{{ name }}</a-select-option>
      </a-select>
      <a-button size="small" class="refresh-btn" @click="loadHistory">
        <RedoOutlined />
        刷新
      </a-button>
    </template>

    <a-spin :spinning="loading" class="page-spin">
      <div v-if="rows.length" class="history-list">
        <div v-for="record in rows" :key="record.id" class="history-item">
          <div class="history-item-head">
            <div class="batch-main">
              <div class="batch-info-row">
                <div class="info-item batch-no-item">
                  <span class="batch-no-link" @click="openPipelineDetail(record)">
                    <span class="batch-hash">#</span>{{ record.batch_number || '-' }}
                  </span>
                </div>
                <div class="info-item">
                  <UserOutlined class="info-icon" />
                  <span class="info-value">{{ getOperator(record) }}</span>
                </div>
                <div class="info-item">
                  <ClockCircleOutlined class="info-icon" />
                  <span class="info-value">{{ formatDateTime(record.started_at || record.created_at) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-value">{{ pipelineBranch || '-' }}</span>
                </div>
                <div class="info-item">
                  <CodeOutlined class="info-icon" />
                  <span class="info-value commit-hash">{{ shortCommit(record.commit_id) }}</span>
                </div>
              </div>
            </div>
            <div class="batch-actions">
              <a-button type="text" size="small" class="action-link" @click="viewLogs(record.id)">
                <EyeOutlined /> 查看日志
              </a-button>
              <a-button
                v-if="record.status === 'running' || record.status === 'pending'"
                type="text"
                size="small"
                class="action-link danger"
                @click="cancelBatch(record.id)"
              >
                <StopOutlined /> 中断
              </a-button>
              <a-button
                v-if="['failed', 'cancelled', 'success'].includes(record.status)"
                type="text"
                size="small"
                class="action-link primary"
                @click="openRerunModal(record.id)"
              >
                <ReloadOutlined /> 重跑节点
              </a-button>
            </div>
          </div>

          <div class="nodes-scroll-wrapper">
            <div class="nodes-grid">
              <div
                v-for="(column, colIdx) in getBatchStageColumns(record)"
                :key="`${record.id}-col-${colIdx}`"
                class="node-column-wrapper"
              >
                <div class="parallel-block">
                  <div
                    v-for="(node, nodeIdx) in column"
                    :key="`${record.id}-${colIdx}-${node.name}-${nodeIdx}`"
                    class="node-card"
                    :class="[`node-status-${node.status}`]"
                    @click="openStageDetail(record, node)"
                  >
                    <div class="node-card-header">
                      <span class="node-name">{{ node.name }}</span>
                      <span class="node-status-badge">{{ getStatusLabel(node.status) }}</span>
                    </div>
                    <div class="node-card-body">
                      <div class="node-duration">
                        <ClockCircleOutlined />
                        {{ formatDuration(record.total_duration, node.status) }}
                      </div>
                      <div class="node-hint">点击查看详情</div>
                    </div>
                  </div>
                </div>
                <div v-if="colIdx < getBatchStageColumns(record).length - 1" class="node-connector"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <a-empty v-else description="暂无执行记录" class="page-empty" />
    </a-spin>

    <a-modal v-model:open="logsOpen" title="执行日志" width="72vw" :footer="null">
      <pre class="logs-view">{{ logsText }}</pre>
    </a-modal>

    <a-modal v-model:open="rerunOpen" title="从节点重跑" @ok="confirmRerunNode">
      <a-form layout="vertical">
        <a-form-item label="节点ID">
          <a-input v-model:value="rerunNodeId" placeholder="例如: task_1741234567" />
        </a-form-item>
      </a-form>
    </a-modal>
  </PageCard>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  EyeOutlined,
  StopOutlined,
  ReloadOutlined,
  ClockCircleOutlined,
  UserOutlined,
  CodeOutlined,
  RedoOutlined,
} from '@ant-design/icons-vue'
import { getPipelineConfig, listPipelineExecutions } from '../api/pipelines'
import { cancelExecutionBatch, listExecutionLogs, rerunExecutionNode } from '../api/executions'
import PageCard from './PageCard.vue'

const props = defineProps({
  token: {
    type: String,
    default: '',
  },
  pipelineId: {
    type: String,
    default: '',
  },
})

const route = useRoute()
const router = useRouter()

const rows = ref([])
const rawRows = ref([])
const loading = ref(false)
const logsOpen = ref(false)
const logsText = ref('')
const rerunOpen = ref(false)
const rerunBatchId = ref('')
const rerunNodeId = ref('')
const mainStages = ref([])
const envStages = ref([])
const pipelineBranch = ref('')
const dateRange = ref([])
const selectedOperator = ref('')

const operatorOptions = computed(() => {
  const names = rawRows.value
    .map((record) => getOperator(record))
    .filter((name) => name && name !== '-')
  return Array.from(new Set(names))
})

const formatDateTime = (value) => {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '-'
  return d.toLocaleString()
}

const formatDuration = (ms, status) => {
  const v = Number(ms)
  if (!v || v <= 0) return status === 'idle' ? '未执行' : '-'
  const s = Math.floor(v / 1000)
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  const r = s % 60
  return `${m}m ${r}s`
}

const shortCommit = (commitId) => {
  const value = String(commitId || '').trim()
  return value ? value.slice(0, 9) : '-'
}

const getOperator = (record) => {
  return String(record?.operator || record?.user_name || route.query.operator || record?.triggered_by || '-').trim() || '-'
}

const statusClass = (status) => {
  if (status === 'success' || status === 'done') return 'success'
  if (status === 'failed' || status === 'error' || status === 'cancelled') return 'failed'
  if (status === 'running') return 'running'
  if (status === 'pending') return 'pending'
  return 'idle'
}

const getStatusLabel = (status) => {
  const labels = {
    success: '成功',
    done: '成功',
    failed: '失败',
    error: '失败',
    running: '运行中',
    pending: '待执行',
    cancelled: '已中断',
    idle: '未执行',
  }
  return labels[status] || '未执行'
}

const nodeNameToStageKey = (name) => {
  const n = String(name || '').toLowerCase()
  if (n.includes('触发源') || n.includes('检出') || n.includes('获取代码') || n.includes('source') || n.includes('checkout')) return 'source'
  if (n.includes('编译') || n.includes('build') || n.includes('maven') || n.includes('gradle')) return 'build'
  if (n.includes('测试') || n.includes('test')) return 'task'
  if (n.includes('部署') || n.includes('灰度') || n.includes('生产') || n.includes('deploy')) return 'deploy'
  return 'task'
}

const normalizeLatestStagesMap = (raw) => {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return {}
  const normalized = { ...raw }
  if (normalized['触发源'] === undefined) {
    if (normalized['代码检出'] !== undefined) normalized['触发源'] = normalized['代码检出']
    if (normalized.source !== undefined) normalized['触发源'] = normalized.source
  }
  if (normalized['编译'] === undefined) {
    if (normalized['编译构建'] !== undefined) normalized['编译'] = normalized['编译构建']
    if (normalized.build !== undefined) normalized['编译'] = normalized.build
  }
  return normalized
}

const buildStageColumns = (stagesMap) => {
  const sourceStages = Array.isArray(mainStages.value) ? mainStages.value : []
  const environmentStages = Array.isArray(envStages.value) ? envStages.value : []
  const stages = [...sourceStages, ...environmentStages]
  const columns = []

  stages.forEach((stage, index) => {
    const name = stage?.name || `阶段 ${index + 1}`
    if (!name) return
    const node = {
      name,
      status: statusClass(stagesMap[name] || stage?.status || 'idle'),
    }
    const runMode = String(stage?.run_mode || stage?.runMode || '').toLowerCase()
    const parallelGroup = String(stage?.parallel_group || stage?.parallelGroup || '').trim()
    const isParallel = runMode === 'parallel' || Boolean(parallelGroup)

    if (!isParallel) {
      columns.push({ key: `serial-${index}`, nodes: [node] })
      return
    }

    const groupKey = parallelGroup || `parallel-${index}`
    const lastColumn = columns[columns.length - 1]
    if (lastColumn && lastColumn.key === groupKey) {
      lastColumn.nodes.push(node)
    } else {
      columns.push({ key: groupKey, nodes: [node] })
    }
  })

  return columns
    .map((column) => column.nodes)
    .filter((column) => Array.isArray(column) && column.length > 0)
}

const getBatchStageColumns = (record) => {
  let parsed = {}
  try {
    parsed = JSON.parse(record.stages_status_json || '{}')
  } catch {
    parsed = {}
  }
  const stagesMap = normalizeLatestStagesMap(parsed)
  return buildStageColumns(stagesMap)
}

const applyFilters = () => {
  let allRows = Array.isArray(rawRows.value) ? [...rawRows.value] : []
  if (dateRange.value && dateRange.value.length === 2) {
    const [start, end] = dateRange.value
    const startDate = start?.startOf ? start.startOf('day').toDate() : new Date(start)
    const endDate = end?.endOf ? end.endOf('day').toDate() : new Date(end)
    allRows = allRows.filter((record) => {
      const recordDate = new Date(record.started_at || record.created_at)
      return recordDate >= startDate && recordDate <= endDate
    })
  }
  if (selectedOperator.value) {
    allRows = allRows.filter((record) => getOperator(record) === selectedOperator.value)
  }
  rows.value = allRows
}

const loadHistory = async () => {
  if (!props.pipelineId || !props.token) {
    rows.value = []
    return
  }

  loading.value = true
  try {
    const [history, pipelineConfig] = await Promise.all([
      listPipelineExecutions(props.token, props.pipelineId, 50),
      getPipelineConfig(props.token, props.pipelineId),
    ])
    rawRows.value = history.batches || []
    mainStages.value = pipelineConfig?.main_stages || []
    envStages.value = pipelineConfig?.env_stages || []
    pipelineBranch.value = pipelineConfig?.branch || ''
    applyFilters()
  } catch {
    message.error('加载执行历史失败')
    rows.value = []
    rawRows.value = []
  } finally {
    loading.value = false
  }
}

const viewLogs = async (batchId) => {
  try {
    const data = await listExecutionLogs(props.token, batchId, 500)
    const lines = (data.logs || []).map((item) => `[${item.log_level}] ${item.log_line}`)
    logsText.value = lines.length ? lines.join('\n') : '暂无日志'
    logsOpen.value = true
  } catch {
    message.error('获取日志失败')
  }
}

const cancelBatch = async (batchId) => {
  try {
    await cancelExecutionBatch(props.token, batchId)
    message.success('已发送中断请求')
    await loadHistory()
  } catch {
    message.error('中断失败')
  }
}

const openRerunModal = (batchId) => {
  rerunBatchId.value = batchId
  rerunNodeId.value = ''
  rerunOpen.value = true
}

const confirmRerunNode = async () => {
  if (!rerunNodeId.value.trim()) {
    message.warning('请输入节点ID')
    return
  }
  try {
    await rerunExecutionNode(props.token, rerunBatchId.value, rerunNodeId.value.trim())
    message.success('节点重跑已提交')
    rerunOpen.value = false
    await loadHistory()
  } catch {
    message.error('重跑失败')
  }
}

const openStageDetail = (record, node) => {
  router.push({
    path: `/pipelines/${props.pipelineId}/executions/${record.id}/stage/${encodeURIComponent(nodeNameToStageKey(node.name))}`,
    query: {
      name: String(route.query.name || record.pipeline_name || ''),
      plan_version: String(route.query.plan_version || ''),
      stage_name: node.name,
      operator: record.triggered_by || '',
      batch_number: String(record.batch_number || ''),
    },
  })
}

const openPipelineDetail = () => {
  router.push({
    path: `/pipelines/${props.pipelineId}/bpm`,
    query: {
      name: String(route.query.name || ''),
      plan_version: String(route.query.plan_version || ''),
    },
  })
}

watch(
  () => [props.pipelineId, props.token],
  () => {
    loadHistory()
  },
  { immediate: true }
)
</script>

<style scoped>
.page-spin {
  display: block;
}

.filter-picker {
  width: 220px;
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  border-color: var(--border-color-light);
}

.filter-select {
  width: 160px;
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  border-color: var(--border-color-light);
}

.refresh-btn {
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color-light);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  height: 30px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.refresh-btn:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
  background: var(--bg-elevated);
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.history-item {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  padding: 18px;
  transition: all var(--transition-base);
  animation: slideUp 0.4s ease-out backwards;
  position: relative;
  overflow: hidden;
}

.history-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--accent-primary), transparent);
  opacity: 0;
  transition: opacity var(--transition-base);
}

.history-item:hover {
  border-color: var(--border-color-accent);
  box-shadow: var(--shadow-lg), var(--shadow-glow);
}

.history-item:hover::before {
  opacity: 1;
}

.history-item-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 14px;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--border-color);
}

.batch-main {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.batch-info-row {
  display: flex;
  align-items: center;
  gap: 18px;
  flex-wrap: wrap;
  color: var(--text-secondary);
  font-size: 13px;
}

.info-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.info-icon {
  font-size: 12px;
  color: var(--text-muted);
}

.batch-no-item {
  gap: 0;
}

.batch-no-link {
  color: var(--text-primary);
  font-weight: 700;
  font-size: 15px;
  text-decoration: none;
  cursor: pointer;
  transition: all var(--transition-fast);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.batch-hash {
  color: var(--accent-primary);
}

.batch-no-link:hover {
  color: var(--accent-primary);
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

.info-value {
  color: var(--text-primary);
  font-weight: 600;
  white-space: nowrap;
  font-family: var(--font-mono);
  font-size: 12px;
}

.commit-hash {
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: 4px;
  letter-spacing: 0.02em;
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.action-link {
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 500;
  transition: all var(--transition-fast);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.action-link:hover {
  color: var(--accent-primary);
}

.action-link.danger {
  color: var(--accent-danger);
}

.action-link.danger:hover {
  color: #f87171;
}

.action-link.primary {
  color: var(--accent-info);
}

.action-link.primary:hover {
  color: #60a5fa;
}

.nodes-scroll-wrapper {
  display: flex;
  align-items: center;
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 6px;
  -webkit-overflow-scrolling: touch;
}

.nodes-scroll-wrapper::-webkit-scrollbar {
  height: 6px;
}

.nodes-scroll-wrapper::-webkit-scrollbar-track {
  background: transparent;
}

.nodes-scroll-wrapper::-webkit-scrollbar-thumb {
  background: var(--border-color-light);
  border-radius: 3px;
}

.nodes-scroll-wrapper::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

.nodes-grid {
  display: flex;
  gap: 0;
  min-width: min-content;
  align-items: stretch;
}

.node-column-wrapper {
  display: flex;
  align-items: center;
}

.parallel-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.node-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 12px;
  background: var(--bg-tertiary);
  cursor: pointer;
  transition: all var(--transition-base);
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 170px;
  flex-shrink: 0;
  position: relative;
  overflow: hidden;
}

.node-card::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  opacity: 0;
  transition: opacity var(--transition-base);
  pointer-events: none;
}

.node-card:hover {
  border-color: var(--accent-primary);
  background: var(--bg-elevated);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.3);
}

.node-card:hover::after {
  opacity: 1;
  box-shadow: inset 0 0 20px rgba(0, 212, 255, 0.05);
}

.node-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}

.node-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
  word-break: break-word;
  line-height: 1.4;
}

.node-status-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
  white-space: nowrap;
  letter-spacing: 0.02em;
}

.node-card-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.node-duration {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.node-duration svg {
  font-size: 11px;
  color: var(--text-muted);
}

.node-hint {
  font-size: 11px;
  color: var(--text-muted);
}

.node-connector {
  width: 20px;
  height: 2px;
  background: linear-gradient(90deg, var(--border-color-light), transparent);
  flex-shrink: 0;
  position: relative;
}

.node-connector::before {
  content: '';
  position: absolute;
  right: -3px;
  top: -4px;
  width: 8px;
  height: 8px;
  background: var(--border-color-light);
  border-radius: 50%;
}

.node-status-success {
  border-color: rgba(16, 185, 129, 0.4);
}

.node-status-success .node-status-badge {
  background: rgba(16, 185, 129, 0.15);
  color: var(--accent-success);
}

.node-status-failed {
  border-color: rgba(239, 68, 68, 0.4);
}

.node-status-failed .node-status-badge {
  background: rgba(239, 68, 68, 0.15);
  color: var(--accent-danger);
}

.node-status-running {
  border-color: rgba(0, 212, 255, 0.4);
  animation: pulse-border 2s ease-in-out infinite;
}

.node-status-running .node-status-badge {
  background: rgba(0, 212, 255, 0.15);
  color: var(--accent-primary);
}

.node-status-pending {
  border-color: rgba(245, 158, 11, 0.4);
}

.node-status-pending .node-status-badge {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.node-status-idle {
  border-color: var(--border-color);
}

.node-status-idle .node-status-badge {
  background: var(--bg-elevated);
  color: var(--text-muted);
}

.logs-view {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 60vh;
  overflow: auto;
  background: #0a0f1a;
  color: #d6e2ff;
  padding: 14px;
  border-radius: var(--radius-md);
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.6;
  border: 1px solid var(--border-color);
}

.page-empty {
  padding: 48px 0;
}

.page-empty :deep(.ant-empty-description) {
  color: var(--text-tertiary);
}

@keyframes pulse-border {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(0, 212, 255, 0.3);
  }
  50% {
    box-shadow: 0 0 0 4px rgba(0, 212, 255, 0.1);
  }
}

@media (max-width: 900px) {
  .history-item {
    padding: 14px;
  }

  .node-card {
    min-width: 140px;
  }

  .filter-picker {
    width: 100%;
  }

  .filter-select {
    width: 100%;
  }
}
</style>
