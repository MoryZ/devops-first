<template>
  <div class="stage-detail-view">
    <a-card :bordered="false" class="head-card">
      <div class="line-one">
        <a class="link" @click="goPlanPipelines">流水线</a>
        <span class="sep">/</span>
        <a class="link" @click="goPipelineBatches">{{ pipelineName || pipelineId }}</a>
        <span class="sep">/</span>
        <span>#{{ displayBatchNumber }}(流水线批次ID) 执行记录</span>
      </div>
      <div class="line-two">关联研发计划: {{ planVersion || '-' }}</div>
    </a-card>

    <a-card :bordered="false" class="middle-card">
      <div class="middle-grid">
        <div class="left-pane">
          <div class="pane-title">BPM 节点</div>
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
                        <div class="node-duration">{{ getNodeDuration(node.name) }}</div>
                      </div>

                      <!-- 悬停预览 -->
                      <div v-if="hoveredNode === node.name || selectedStageName === node.name" class="node-preview-popup">
                        <div class="preview-title">{{ node.name }}</div>
                        <div v-for="(item, pidx) in (hoveredNode === node.name ? hoveredNodePreview : selectedNodePreview)" :key="`${pidx}-${item}`" class="preview-line">
                          {{ item }}
                        </div>
                      </div>
                    </div>
                  </div>
                  <!-- 连接线 -->
                  <div v-if="colIdx < stageColumns.length - 1" class="node-connector"></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="right-pane">
          <div class="pane-title">节点运行详情</div>
          <div class="meta-row"><span class="k">操作人</span><span class="v">{{ operatorText }}</span></div>
          <div class="meta-row"><span class="k">开始时间</span><span class="v">{{ startTime }}</span></div>
          <div class="meta-row"><span class="k">运行时间</span><span class="v">{{ durationText }}</span></div>
          <div class="meta-row"><span class="k">源码地址</span><span class="v break">{{ repoUrl || '-' }}</span></div>
          <div class="meta-row">
            <span class="k">提交记录ID</span>
            <span class="v">
              <a-tooltip v-if="commitId" :title="commitId">
                <span class="commit-short">{{ commitId.slice(0, 9) }}</span>
              </a-tooltip>
              <span v-else>-</span>
              <a class="commit-link" @click="openCommitModal">提交记录</a>
            </span>
          </div>
        </div>
      </div>
    </a-card>

    <a-card :bordered="false" class="logs-card" :title="`实时日志 · ${selectedStageName || stageName || stageKey}`">
      <div class="log-toolbar">
        <a-button size="small" @click="loadAll">刷新</a-button>
        <a-button size="small" @click="scrollToBottom">下拉滚动</a-button>
        <a-button size="small" @click="toggleFullscreen">{{ isFullscreen ? '退出全屏' : '全屏' }}</a-button>
        <a-button size="small" @click="downloadLogs">下载</a-button>
      </div>
      <pre ref="logsRef" class="logs-view">{{ logsText }}</pre>
    </a-card>

    <a-modal v-model:open="commitModalOpen" title="提交记录" width="70vw" :footer="null">
      <a-table
        :data-source="commitRows"
        :columns="commitColumns"
        :loading="commitLoading"
        row-key="code_version"
        :pagination="{ pageSize: 8, showSizeChanger: false }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'code_version'">
            <a-tooltip :title="record.code_version">
              <span>{{ String(record.code_version || '').slice(0, 16) }}...</span>
            </a-tooltip>
          </template>
        </template>
      </a-table>
    </a-modal>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getBatchStatus, listExecutionCommits, listExecutionLogs } from '../api/executions'
import { getPipelineConfig } from '../api/pipelines'

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
const pipelineTypeName = computed(() => `${triggeredBy.value || 'manual'}-${pipelineName.value || pipelineId.value}`)

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
    'idle': '未执行'
  }
  return labels[status] || '未执行'
}

const getNodeDuration = (nodeName) => {
  // 获取该节点的执行时间（秒），可以从日志中解析或从批次信息中获取
  // 暂时返回示例数据
  const durations = {
    '触发源': '2s',
    '代码检出': '3s',
    '编译': '2s',
    '单元测试': '0s',
    '打包': '0s',
    '部署开发': '0s',
    '部署测试': '0s'
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
.stage-detail-view {
  min-height: 100vh;
  background: #f3f5f9;
  padding: 18px;
}

.head-card,
.middle-card,
.logs-card {
  border-radius: 12px;
  box-shadow: 0 8px 22px rgba(17, 36, 64, 0.08);
  margin-bottom: 14px;
}

.line-one {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #1f2f46;
  font-weight: 600;
  flex-wrap: wrap;
}

.link {
  color: #2d67d8;
}

.sep {
  color: #8ca0bf;
}

.line-two {
  margin-top: 8px;
  color: #536b8f;
  font-size: 14px;
}

.middle-grid {
  display: grid;
  grid-template-columns: 1.8fr 0.8fr;
  gap: 16px;
}

.pane-title {
  font-size: 15px;
  font-weight: 700;
  color: #24364f;
  margin-bottom: 10px;
}

.nodes-wrap {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 12px;
  margin-bottom: 12px;
  padding: 8px 0;
}

.node-item {
  border: 2px solid #e2e9f5;
  border-radius: 10px;
  padding: 12px;
  background: #fbfdff;
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.node-item:hover {
  border-color: #b3c7e8;
  box-shadow: 0 4px 12px rgba(45, 103, 216, 0.1);
}

.node-item .dot {
  display: none;
}

.node-item.node-success {
  border-color: #2fb65c;
  background: #f0f9f5;
}

.node-item.node-failed {
  border-color: #e24b4b;
  background: #fef5f5;
}

.node-item.node-running {
  border-color: #2e7cf2;
  background: #eef5ff;
}

.node-item.node-pending {
  border-color: #faad14;
  background: #fffbeb;
}

.node-item.active {
  border-color: #2e7cf2;
  border-width: 3px;
  box-shadow: 0 4px 16px rgba(45, 103, 216, 0.25);
  background: #eef5ff;
}

.node-preview {
  border: 1px solid #e2e9f5;
  border-radius: 8px;
  background: #fbfdff;
  padding: 10px;
}

.preview-title {
  color: #29456f;
  font-weight: 600;
  margin-bottom: 6px;
}

.preview-line {
  color: #586f91;
  font-size: 12px;
  line-height: 1.7;
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
  background: #d0d7e1;
  border-radius: 3px;
}

.nodes-scroll-wrapper::-webkit-scrollbar-thumb:hover {
  background: #b0b8c8;
}

.nodes-grid {
  display: flex;
  gap: 0;
  padding: 8px 0;
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
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 140px;
  position: relative;
  flex-shrink: 0;
}

.node-card:hover {
  border-color: #b3c7e8;
  box-shadow: 0 4px 12px rgba(45, 103, 216, 0.1);
}

.node-card.selected {
  border-color: #2e7cf2;
  border-width: 3px;
  box-shadow: 0 4px 16px rgba(45, 103, 216, 0.25);
  background: #eef5ff;
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
  text-transform: uppercase;
}

.node-status-success .node-status-badge {
  background: #d6f5e6;
  color: #1d6e3f;
}

.node-status-failed .node-status-badge {
  background: #fce2e2;
  color: #7a1f1f;
}

.node-status-running .node-status-badge {
  background: #dce9fc;
  color: #1d3f7a;
}

.node-status-pending .node-status-badge {
  background: #fff9e6;
  color: #7a5a1f;
}

.node-status-idle .node-status-badge {
  background: #f0f2f5;
  color: #595959;
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

.node-preview-popup {
  position: absolute;
  left: 50%;
  bottom: 100%;
  transform: translateX(-50%);
  margin-bottom: 8px;
  background: #ffffff;
  border: 1px solid #e2e9f5;
  border-radius: 8px;
  padding: 10px;
  min-width: 150px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  z-index: 100;
  white-space: normal;
  max-width: 250px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.preview-title {
  color: #29456f;
  font-weight: 600;
  font-size: 13px;
}

.preview-line {
  color: #586f91;
  font-size: 12px;
  line-height: 1.6;
}

.meta-row {
  display: grid;
  grid-template-columns: 140px 1fr;
  gap: 8px;
  margin-bottom: 8px;
}

.k {
  color: #6e7f99;
}

.v {
  color: #24364f;
  font-weight: 600;
}

.break {
  word-break: break-all;
}

.commit-short {
  font-family: Menlo, Monaco, Consolas, 'Courier New', monospace;
  margin-right: 8px;
}

.commit-link {
  margin-left: 8px;
  color: #2d67d8;
}

.log-toolbar {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.logs-view {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 58vh;
  overflow: auto;
  background: #0f1722;
  color: #d6e2ff;
  padding: 12px;
  border-radius: 8px;
}

@media (max-width: 1180px) {
  .middle-grid {
    grid-template-columns: 1fr;
  }
  .meta-row {
    grid-template-columns: 1fr;
  }
}
</style>
