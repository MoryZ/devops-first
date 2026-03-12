<template>
  <a-card class="history-card" :bordered="false" title="执行历史">
    <div class="history-toolbar">
      <a-button size="small" @click="loadHistory">刷新</a-button>
    </div>

    <a-table
      :data-source="rows"
      :columns="columns"
      row-key="id"
      :pagination="{ pageSize: 5, size: 'small' }"
      size="small"
      :loading="loading"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'status'">
          <a-tag :color="statusColor[record.status] || 'default'">{{ record.status }}</a-tag>
        </template>
        <template v-else-if="column.key === 'duration'">
          {{ formatDuration(record.total_duration) }}
        </template>
        <template v-else-if="column.key === 'action'">
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
        </template>
      </template>
    </a-table>

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
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'

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

const rows = ref([])
const loading = ref(false)
const logsOpen = ref(false)
const logsText = ref('')
const rerunOpen = ref(false)
const rerunBatchId = ref('')
const rerunNodeId = ref('')

const columns = [
  { title: '批次', dataIndex: 'batch_number', key: 'batch_number', width: 80 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '触发方式', dataIndex: 'triggered_by', key: 'triggered_by', width: 100 },
  { title: '开始时间', dataIndex: 'created_at', key: 'created_at' },
  { title: '耗时', dataIndex: 'total_duration', key: 'duration', width: 100 },
  { title: '操作', key: 'action', width: 220 },
]

const statusColor = {
  pending: 'default',
  running: 'processing',
  success: 'success',
  failed: 'error',
  cancelled: 'warning',
}

const formatDuration = (ms) => {
  if (!ms || Number(ms) <= 0) return '-'
  const sec = Math.floor(Number(ms) / 1000)
  return `${sec}s`
}

const loadHistory = async () => {
  if (!props.pipelineId || !props.token) {
    rows.value = []
    return
  }

  loading.value = true
  try {
    const res = await fetch(`/api/pipelines/${props.pipelineId}/executions?limit=50`, {
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (!res.ok) {
      rows.value = []
      return
    }
    const data = await res.json()
    rows.value = data.batches || []
  } catch (err) {
    message.error('加载执行历史失败')
    rows.value = []
  } finally {
    loading.value = false
  }
}

const viewLogs = async (batchId) => {
  try {
    const res = await fetch(`/api/executions/${batchId}/logs?limit=500`, {
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (!res.ok) {
      message.error('获取日志失败')
      return
    }
    const data = await res.json()
    const lines = (data.logs || []).map((item) => `[${item.log_level}] ${item.log_line}`)
    logsText.value = lines.length ? lines.join('\n') : '暂无日志'
    logsOpen.value = true
  } catch {
    message.error('获取日志失败')
  }
}

const cancelBatch = async (batchId) => {
  try {
    const res = await fetch(`/api/executions/${batchId}/cancel`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (!res.ok) {
      const err = await res.json()
      message.error(err.error || '中断失败')
      return
    }
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
    const res = await fetch(`/api/executions/${rerunBatchId.value}/rerun-node`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${props.token}`,
      },
      body: JSON.stringify({ node_id: rerunNodeId.value.trim() }),
    })
    if (!res.ok) {
      const err = await res.json()
      message.error(err.error || '重跑失败')
      return
    }
    message.success('节点重跑已提交')
    rerunOpen.value = false
    await loadHistory()
  } catch {
    message.error('重跑失败')
  }
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
  justify-content: flex-end;
  margin-bottom: 10px;
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
</style>
