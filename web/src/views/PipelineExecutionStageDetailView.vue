<template>
  <PageCard title="流水线详情" :subtitle="`批次 #${displayBatchNumber}`" icon="◆" bg="gradient">
    <template #header>
      <div class="page-header-content">
        <div class="header-left">
          <div class="breadcrumb">
            <span class="breadcrumb-item" @click="goPlanPipelines">流水线</span>
            <span class="breadcrumb-sep">/</span>
            <span class="breadcrumb-item" @click="goPipelineBatches">{{ pipelineName || pipelineId }}</span>
            <span class="breadcrumb-sep">/</span>
            <span class="breadcrumb-current">#{{ displayBatchNumber }} 执行记录</span>
          </div>
          <div class="plan-badge" v-if="planVersion">
            <span class="badge-icon">◈</span>
            关联研发计划：{{ planVersion }}
          </div>
        </div>
      </div>
    </template>

    <div class="detail-grid">
      <a-card :bordered="false" class="detail-card nodes-card">
        <div class="card-header">
          <span class="card-icon">▶</span>
          <h3 class="card-title">BPM 节点</h3>
        </div>
        <div class="nodes-container">
          <div class="nodes-scroll-wrapper">
            <div class="nodes-grid">
              <div
                v-for="(column, colIdx) in stageColumns"
                :key="`col-${colIdx}`"
                class="node-column-wrapper"
              >
                <div class="parallel-block">
                  <div
                    v-for="(node, nodeIdx) in column"
                    :key="`${node.name}-${nodeIdx}`"
                    class="node-card"
                    :class="[`node-status-${node.status}`, { selected: selectedStageName === node.name }]"
                    @click="selectNode(node)"
                    @mouseenter="onNodeHover(node)"
                    @mouseleave="onNodeLeave"
                  >
                    <div class="node-card-header">
                      <span class="node-name">{{ node.name }}</span>
                      <span class="node-status-badge">{{ getStatusLabel(node.status) }}</span>
                    </div>
                    <div class="node-card-body">
                      <div class="node-duration">
                        <ClockCircleOutlined />
                        {{ getNodeDuration(node.name) }}
                      </div>
                    </div>

                    <div v-if="hoveredNode === node.name || selectedStageName === node.name" class="node-preview-popup">
                      <div class="preview-title">{{ node.name }}</div>
                      <div v-for="(item, pidx) in (hoveredNode === node.name ? hoveredNodePreview : selectedNodePreview)" :key="pidx" class="preview-line">
                        {{ item }}
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="colIdx < stageColumns.length - 1" class="node-connector"></div>
              </div>
            </div>
          </div>
        </div>
      </a-card>

      <a-card :bordered="false" class="detail-card meta-card">
        <div class="card-header">
          <span class="card-icon">◈</span>
          <h3 class="card-title">节点运行详情</h3>
        </div>
        <div class="meta-grid">
          <div class="meta-row">
            <span class="meta-label">操作人</span>
            <span class="meta-value">{{ operatorText }}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">开始时间</span>
            <span class="meta-value">{{ startTime }}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">运行时间</span>
            <span class="meta-value">{{ durationText }}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">源码地址</span>
            <span class="meta-value break">{{ repoUrl || '-' }}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">提交记录ID</span>
            <span class="meta-value">
              <a-tooltip v-if="commitId" :title="commitId">
                <span class="commit-short">{{ commitId.slice(0, 9) }}</span>
              </a-tooltip>
              <span v-else>-</span>
              <a class="commit-link" @click="openCommitModal">提交记录</a>
            </span>
          </div>
        </div>
      </a-card>
    </div>

    <a-card :bordered="false" class="detail-card logs-card">
      <div class="card-header">
        <span class="card-icon">▶</span>
        <h3 class="card-title">
          实时日志
          <span class="card-subtitle">· {{ selectedStageName || stageName || stageKey }}</span>
        </h3>
        <div class="log-toolbar">
          <a-button size="small" class="toolbar-btn" @click="loadAll">
            <ReloadOutlined />
            刷新
          </a-button>
          <a-button size="small" class="toolbar-btn" @click="scrollToBottom">
            <ArrowDownOutlined />
            下拉滚动
          </a-button>
          <a-button size="small" class="toolbar-btn" @click="toggleFullscreen">
            <FullscreenOutlined />
            {{ isFullscreen ? '退出全屏' : '全屏' }}
          </a-button>
          <a-button size="small" class="toolbar-btn" @click="downloadLogs">
            <DownloadOutlined />
            下载
          </a-button>
        </div>
      </div>
      <pre ref="logsRef" class="logs-view">{{ logsText }}</pre>
    </a-card>

    <a-modal v-model:open="commitModalOpen" title="提交记录" width="72vw" :footer="null">
      <a-table
        :data-source="commitRows"
        :columns="commitColumns"
        :loading="commitLoading"
        row-key="code_version"
        :pagination="{ pageSize: 8, showSizeChanger: false }"
        class="commit-table"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'code_version'">
            <a-tooltip :title="record.code_version">
              <span class="commit-id-cell">{{ String(record.code_version || '').slice(0, 16) }}...</span>
            </a-tooltip>
          </template>
        </template>
      </a-table>
    </a-modal>
  </PageCard>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowDownOutlined,
  ClockCircleOutlined,
  CodeOutlined,
  DownloadOutlined,
  FullscreenOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import { getBatchStatus, listExecutionCommits, listExecutionLogs } from '../api/executions'
import { getPipelineConfig } from '../api/pipelines'
import PageCard from '../components/PageCard.vue'

const route = useRoute()
const router = useRouter()

const token = computed(() => localStorage.getItem('token') || '')
const pipelineId = computed(() => String(route.params.id || ''))
const batchId = computed(() => String(route.params.batchId || ''))
const stageKey = computed(() => String(route.params.stageKey || ''))
const stageName = computed(() => String(route.query.stage_name || ''))
const pipelineName = computed(() => String(route.query.name || ''))
const operator = computed(() => String(route.query.operator || ''))
const planVersion = computed(() => String(route.query.plan_version || ''))
const batchNumber = computed(() => String(route.query.batch_number || ''))

const startTime = ref('-')
const endTime = ref('-')
const durationText = ref('-')
const logsText = ref('')
const stageNodes = ref([])
const hoveredNode = ref(null)
const selectedStageName = ref(stageName.value || '')
const selectedStageKey = ref(stageKey.value || 'task')
const selectedNodePreview = ref([])
const hoveredNodePreview = ref([])
const logsRef = ref(null)
const isFullscreen = ref(false)
const stageColumns = ref([])
const displayBatchNumber = ref(batchNumber.value || '-')
const commitId = ref('')
const repoUrl = ref('')
const triggeredBy = ref('manual')

const commitModalOpen = ref(false)
const commitRows = ref([])
const commitLoading = ref(false)
let timer = null

const commitColumns = [
  { title: '代码版本号', dataIndex: 'code_version', key: 'code_version', width: 200 },
  { title: '内容', dataIndex: 'content', key: 'content' },
  { title: '提交人', dataIndex: 'author', key: 'author', width: 180 },
  { title: '提交时间', dataIndex: 'committed_at', key: 'committed_at', width: 200 },
]

const operatorText = computed(() => operator.value || '-')

const formatDateTime = (value) => {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '-'
  return d.toLocaleString()
}

const toDuration = (ms) => {
  const v = Number(ms)
  if (!v || v <= 0) return '-'
  const s = Math.floor(v / 1000)
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  const r = s % 60
  return `${m}m ${r}s`
}

const statusClass = (status) => {
  if (status === 'success' || status === 'done') return 'success'
  if (status === 'failed' || status === 'error') return 'failed'
  if (status === 'running') return 'running'
  if (status === 'pending') return 'pending'
  return 'idle'
}

const nodeNameToStageKey = (name) => {
  const n = String(name || '').toLowerCase()
  if (n.includes('触发源') || n.includes('检出') || n.includes('获取代码') || n.includes('source') || n.includes('checkout')) return 'source'
  if (n.includes('编译') || n.includes('build') || n.includes('maven') || n.includes('gradle')) return 'build'
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

const buildStageColumns = (mainStages, envStages, stagesMap) => {
  const allStages = [...(Array.isArray(mainStages) ? mainStages : []), ...(Array.isArray(envStages) ? envStages : [])]
  const columns = []

  allStages.forEach((stage, index) => {
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

const loadBatchAndNodes = async () => {
  if (!token.value || !batchId.value) return
  try {
    const [batch, cfg] = await Promise.all([
      getBatchStatus(token.value, batchId.value),
      getPipelineConfig(token.value, pipelineId.value),
    ])

    startTime.value = formatDateTime(batch.started_at || batch.StartedAt || batch.created_at || batch.CreatedAt)
    endTime.value = formatDateTime(batch.completed_at || batch.CompletedAt)
    durationText.value = toDuration(batch.total_duration || batch.TotalDuration)
    displayBatchNumber.value = String(batch.batch_number || batch.BatchNumber || displayBatchNumber.value || '-')
    commitId.value = String(batch.commit_id || batch.CommitID || '')
    triggeredBy.value = String(batch.triggered_by || batch.TriggeredBy || 'manual')
    repoUrl.value = String(cfg?.repo_url || cfg?.repoUrl || '')

    const stagesMap = normalizeLatestStagesMap(JSON.parse(batch.stages_status_json || batch.StagesStatusJSON || '{}'))
    stageColumns.value = buildStageColumns(cfg?.main_stages || [], cfg?.env_stages || [], stagesMap)
    stageNodes.value = stageColumns.value.flat()

    if (!selectedStageName.value && stageNodes.value.length > 0) {
      selectedStageName.value = stageNodes.value[0].name
      selectedStageKey.value = nodeNameToStageKey(stageNodes.value[0].name)
    }
  } catch {
    // Keep page alive on transient errors.
  }
}

const loadLogs = async () => {
  if (!token.value || !batchId.value) return
  try {
    const data = await listExecutionLogs(token.value, batchId.value, 3000)
    const all = data.logs || []
    const key = selectedStageKey.value || stageKey.value
    const selected = all.filter((l) => String(l.stage || '').toLowerCase() === key)

    selectedNodePreview.value = selected
      .map((i) => String(i.log_line || ''))
      .filter((x) => x && !x.includes('=== Stage:'))
      .slice(0, 5)

    const lines = selected.map((item) => `${formatDateTime(item.created_at)} [${item.log_level}] ${item.log_line}`)
    logsText.value = lines.length ? lines.join('\n') : '暂无该节点日志'
    await nextTick()
    scrollToBottom()
  } catch {
    logsText.value = '日志加载失败'
  }
}

const loadAll = async () => {
  await loadBatchAndNodes()
  await loadLogs()
}

const selectNode = (node) => {
  selectedStageName.value = node.name
  selectedStageKey.value = nodeNameToStageKey(node.name)
  loadLogs()
}

const getNodePreview = async (nodeName) => {
  if (!token.value || !batchId.value) return []
  try {
    const stageKey = nodeNameToStageKey(nodeName)
    const data = await listExecutionLogs(token.value, batchId.value, 3000)
    const all = data.logs || []
    const selected = all.filter((l) => String(l.stage || '').toLowerCase() === stageKey)
    return selected
      .map((i) => String(i.log_line || ''))
      .filter((x) => x && !x.includes('=== Stage:'))
      .slice(0, 5)
  } catch {
    return []
  }
}

const onNodeHover = async (node) => {
  hoveredNode.value = node.name
  hoveredNodePreview.value = await getNodePreview(node.name)
}

const onNodeLeave = () => {
  hoveredNode.value = null
  hoveredNodePreview.value = []
}

const getStatusLabel = (status) => {
  const labels = {
    'success': '成功',
    'done': '成功',
    'failed': '失败',
    'error': '失败',
    'running': '运行中',
    'pending': '待执行',
    'idle': '未执行',
  }
  return labels[status] || '未执行'
}

const getNodeDuration = (nodeName) => {
  const durations = {
    '触发源': '2s',
    '代码检出': '3s',
    '编译': '2s',
    '单元测试': '0s',
    '打包': '0s',
    '部署开发': '0s',
    '部署测试': '0s',
  }
  return durations[nodeName] || '-'
}

const scrollToBottom = () => {
  const el = logsRef.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

const toggleFullscreen = async () => {
  const el = logsRef.value
  if (!el) return
  if (!document.fullscreenElement) {
    await el.requestFullscreen()
    isFullscreen.value = true
  } else {
    await document.exitFullscreen()
    isFullscreen.value = false
  }
}

const downloadLogs = () => {
  const blob = new Blob([logsText.value || ''], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${pipelineName.value || pipelineId.value}_${selectedStageName.value || selectedStageKey.value}_logs.txt`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

const openCommitModal = async () => {
  commitModalOpen.value = true
  commitLoading.value = true
  try {
    const data = await listExecutionCommits(token.value, batchId.value, 20)
    commitRows.value = data.items || []
  } catch {
    commitRows.value = []
  } finally {
    commitLoading.value = false
  }
}

const goPlanPipelines = () => {
  router.push('/workspace')
}

const goPipelineBatches = () => {
  router.push({
    path: `/pipelines/${pipelineId.value}/executions`,
    query: {
      name: pipelineName.value || '',
      plan_version: planVersion.value || '',
    },
  })
}

onMounted(async () => {
  if (stageName.value) selectedStageName.value = stageName.value
  if (stageKey.value) selectedStageKey.value = stageKey.value
  await loadAll()
  timer = setInterval(loadAll, 2000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  isFullscreen.value = false
})
</script>

<style scoped>
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.breadcrumb-item {
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all var(--transition-fast);
}

.breadcrumb-item:hover {
  color: var(--accent-primary);
}

.breadcrumb-sep {
  color: var(--text-muted);
  font-size: 12px;
}

.breadcrumb-current {
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 600;
}

.plan-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(0, 212, 255, 0.3);
  background: rgba(0, 212, 255, 0.08);
  color: var(--accent-primary);
  font-size: 13px;
  font-weight: 500;
  width: fit-content;
}

.badge-icon {
  font-size: 12px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1.8fr 0.8fr;
  gap: 20px;
  margin-bottom: 20px;
}

.detail-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);
}

.detail-card:hover {
  border-color: var(--border-color-accent);
  box-shadow: var(--shadow-lg), var(--shadow-glow);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--border-color);
}

.card-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border-radius: 7px;
  font-size: 11px;
  color: white;
  flex-shrink: 0;
}

.card-title {
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: 0.02em;
}

.card-subtitle {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-tertiary);
  margin-left: 4px;
}

.nodes-container {
  width: 100%;
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
  background: var(--border-color-light);
  border-radius: 3px;
}

.nodes-scroll-wrapper::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

.nodes-grid {
  display: flex;
  gap: 0;
  padding: 4px 0;
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
  min-width: 150px;
  position: relative;
  flex-shrink: 0;
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

.node-card.selected {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.2), var(--shadow-lg);
  background: rgba(0, 212, 255, 0.08);
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
  text-transform: uppercase;
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

.node-preview-popup {
  position: absolute;
  left: 50%;
  bottom: 100%;
  transform: translateX(-50%);
  margin-bottom: 8px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-color-accent);
  border-radius: var(--radius-md);
  padding: 10px 12px;
  min-width: 160px;
  max-width: 260px;
  box-shadow: var(--shadow-lg);
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.preview-title {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 13px;
  margin-bottom: 2px;
}

.preview-line {
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.6;
  word-break: break-all;
}

.meta-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}

.meta-row {
  display: grid;
  grid-template-columns: 100px 1fr;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px dashed var(--border-color);
}

.meta-row:last-child {
  border-bottom: none;
}

.meta-label {
  color: var(--text-tertiary);
  font-size: 13px;
  font-weight: 500;
}

.meta-value {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 13px;
}

.break {
  word-break: break-all;
}

.commit-short {
  font-family: var(--font-mono);
  margin-right: 8px;
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  letter-spacing: 0.02em;
}

.commit-link {
  color: var(--accent-primary);
  cursor: pointer;
  margin-left: 8px;
  font-size: 13px;
  transition: all var(--transition-fast);
}

.commit-link:hover {
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.4);
}

.log-toolbar {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.toolbar-btn {
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color-light);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  height: 30px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-family: var(--font-display);
  transition: all var(--transition-fast);
}

.toolbar-btn:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
  background: var(--bg-elevated);
}

.logs-card {
  margin-top: 0;
}

.logs-view {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 58vh;
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

.commit-table {
  border-radius: var(--radius-md);
  overflow: hidden;
}

.commit-id-cell {
  font-family: var(--font-mono);
  font-size: 12px;
}

@keyframes pulse-border {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(0, 212, 255, 0.3);
  }
  50% {
    box-shadow: 0 0 0 4px rgba(0, 212, 255, 0.1);
  }
}

@media (max-width: 1180px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }

  .meta-grid {
    grid-template-columns: 1fr;
  }

  .meta-row {
    grid-template-columns: 1fr;
  }
}
</style>
