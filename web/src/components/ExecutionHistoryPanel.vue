<template>
  <a-card class="history-card" :bordered="false">
    <div class="history-toolbar">
      <div class="history-tabs">
        <span class="tab-item active">执行记录</span>
        <span class="tab-item">修改历史</span>
      </div>
      <div class="filter-group">
        <a-range-picker
          v-model:value="dateRange"
          format="YYYY-MM-DD"
          :placeholder="['开始日期', '结束日期']"
          size="small"
          style="width: 240px"
          @change="applyFilters"
        />
        <a-select
          v-model:value="selectedOperator"
          placeholder="全部"
          size="small"
          style="width: 140px"
          allow-clear
          @change="applyFilters"
        >
          <a-select-option value="">全部</a-select-option>
          <a-select-option v-for="name in operatorOptions" :key="name" :value="name">{{ name }}</a-select-option>
        </a-select>
        <a-button size="small" @click="loadHistory">刷新</a-button>
      </div>
    </div>

    <a-spin :spinning="loading">
      <div v-if="rows.length" class="history-list">
        <div v-for="record in rows" :key="record.id" class="history-item">
          <div class="history-item-head">
            <div class="batch-main">
              <div class="batch-info-row">
                <div class="info-item batch-no-item">
                  <a class="batch-no-link" @click="openPipelineDetail(record)">#{{ record.batch_number || '-' }}</a>
                </div>
                <div class="info-item">
                  <span class="info-value">{{ getOperator(record) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-value">{{ formatDateTime(record.started_at || record.created_at) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-value">{{ pipelineBranch || '-' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-value">{{ shortCommit(record.commit_id) }}</span>
                </div>
              </div>
            </div>
            <div class="batch-actions">
              <a-button type="link" size="small" @click="viewLogs(record.id)">查看日志</a-button>
              <a-button
                v-if="record.status === 'running' || record.status === 'pending'"
                type="link"
                size="small"
                danger
                @click="cancelBatch(record.id)"
              >中断</a-button>
              <a-button
                v-if="record.status === 'failed' || record.status === 'cancelled' || record.status === 'success'"
                type="link"
                size="small"
                @click="openRerunModal(record.id)"
              >重跑节点</a-button>
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
                      <div class="node-duration">{{ formatDuration(record.total_duration, node.status) }}</div>
                      <div class="node-hint">点击查看节点详情</div>
                    </div>
                  </div>
                </div>
                <div v-if="colIdx < getBatchStageColumns(record).length - 1" class="node-connector"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <a-empty v-else description="暂无执行记录" />
    </a-spin>

    <a-modal v-model:open="logsOpen" title="执行日志" width="70vw" :footer="null">
      <pre class="logs-view">{{ logsText }}</pre>
    </a-modal>

    <a-modal v-model:open="rerunOpen" title="从节点重跑" @ok="confirmRerunNode">
      <a-form layout="vertical">
        <a-form-item label="节点ID">
          <a-input v-model:value="rerunNodeId" placeholder="例如: task_1741234567" />
        </a-form-item>
      </a-form>
    </a-modal>
  </a-card>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { getPipelineConfig, listPipelineExecutions } from '../api/pipelines'
import { cancelExecutionBatch, listExecutionLogs, rerunExecutionNode } from '../api/executions'

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

const buildStageNodes = (stagesMap) => {
  return buildStageColumns(stagesMap).flat().map((stage) => {
      return {
        name: stage.name,
        status: stage.status,
      }
    })
}

const getBatchStageNodes = (record) => {
  const stagesMap = normalizeLatestStagesMap(JSON.parse(record.stages_status_json || '{}'))
  return buildStageNodes(stagesMap)
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
    path: `/pipelines/${props.pipelineId}/executions/${record.id}/stage/${nodeNameToStageKey(node.name)}`,
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
.history-card {
  border-radius: 12px;
  margin-top: 14px;
  box-shadow: 0 8px 22px rgba(17, 36, 64, 0.08);
}

.history-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.history-tabs {
  display: inline-flex;
  align-items: center;
  gap: 16px;
}

.tab-item {
  font-size: 16px;
  font-weight: 600;
  color: #6f819d;
}

.tab-item.active {
  color: #1f2f46;
}

.filter-group {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.history-item {
  border: 1px solid #e6ebf5;
  border-radius: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #f9fbff 100%);
  padding: 16px;
}

.history-item-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 12px;
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
  gap: 22px;
  flex-wrap: wrap;
  color: #60738f;
  font-size: 14px;
}

.info-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.batch-no-item {
  gap: 0;
}

.batch-no-link {
  color: #24364f;
  font-weight: 700;
  text-decoration: none;
  cursor: pointer;
  transition: color 0.2s ease;
}

.batch-no-link:hover {
  color: #2d67d8;
}

.info-value {
  color: #24364f;
  font-weight: 600;
  white-space: nowrap;
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.nodes-scroll-wrapper {
  display: flex;
  align-items: center;
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 8px;
  -webkit-overflow-scrolling: touch;
}

.nodes-scroll-wrapper::-webkit-scrollbar {
  height: 6px;
}

.nodes-scroll-wrapper::-webkit-scrollbar-track {
  background: transparent;
}

.nodes-scroll-wrapper::-webkit-scrollbar-thumb {
  background: #d0d7e1;
  border-radius: 3px;
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
  gap: 8px;
}

.node-card {
  border: 2px solid #e2e9f5;
  border-radius: 10px;
  padding: 12px;
  background: #fbfdff;
  cursor: pointer;
  transition: all 0.25s ease;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 165px;
  flex-shrink: 0;
}

.node-card:hover {
  border-color: #2e7cf2;
  box-shadow: 0 4px 12px rgba(45, 103, 216, 0.12);
  transform: translateY(-1px);
}

.node-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 6px;
}

.node-name {
  font-size: 14px;
  font-weight: 600;
  color: #24364f;
  flex: 1;
  word-break: break-word;
}

.node-status-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.node-card-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.node-duration {
  font-size: 13px;
  font-weight: 500;
  color: #536b8f;
}

.node-hint {
  font-size: 12px;
  color: #8a9bb4;
}

.node-connector {
  width: 20px;
  height: 2px;
  background: #d0d7e1;
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
  background: #d0d7e1;
  border-radius: 50%;
}

.node-status-success {
  border-color: #2fb65c;
  background: #f0f9f5;
}

.node-status-success .node-status-badge {
  background: #d6f5e6;
  color: #1d6e3f;
}

.node-status-failed {
  border-color: #e24b4b;
  background: #fef5f5;
}

.node-status-failed .node-status-badge {
  background: #fce2e2;
  color: #7a1f1f;
}

.node-status-running {
  border-color: #2e7cf2;
  background: #eef5ff;
}

.node-status-running .node-status-badge {
  background: #dce9fc;
  color: #1d3f7a;
}

.node-status-pending {
  border-color: #faad14;
  background: #fffbeb;
}

.node-status-pending .node-status-badge {
  background: #fff9e6;
  color: #7a5a1f;
}

.node-status-idle {
  border-color: #dce4f2;
  background: #f7faff;
}

.node-status-idle .node-status-badge {
  background: #f0f2f5;
  color: #595959;
}

.logs-view {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 60vh;
  overflow: auto;
  background: #0f1722;
  color: #d6e2ff;
  padding: 12px;
  border-radius: 8px;
}

@media (max-width: 900px) {
  .history-item {
    padding: 14px;
  }

  .node-card {
    min-width: 132px;
  }
}
</style>
