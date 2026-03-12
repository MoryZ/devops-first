<template>
  <a-card class="exec-panel" :bordered="false" title="执行控制台">
    <div class="selected-line">
      <a-tag color="processing">当前流水线：{{ pipelineName || '未选择' }}</a-tag>
      <a-tag :color="statusColor[status]">{{ statusLabel[status] }}</a-tag>
      <a-tag v-if="releaseUnitName" color="geekblue">发布单元：{{ releaseUnitName }}</a-tag>
      <a-tag v-if="repositoryType" color="purple">触发源：{{ repositoryType }}</a-tag>
      <a-tag :color="autoMerge ? 'blue' : 'default'">自动归并：{{ autoMerge ? 'ON' : 'OFF' }}</a-tag>
      <a-tag :color="autoTag ? 'cyan' : 'default'">自动Tag：{{ autoTag ? 'ON' : 'OFF' }}</a-tag>
    </div>
    <div class="control-row">
      <a-input
        v-model:value="projectPath"
        size="large"
        allow-clear
        placeholder="Input project path, e.g. /Users/name/projects/order-service"
      />
      <a-button type="primary" size="large" :disabled="isDeploying" @click="handleStart">
        Start Deploy
      </a-button>
      <a-button danger size="large" :disabled="!isDeploying" @click="handleStop">
        Stop
      </a-button>
    </div>
    <div ref="terminalEl" class="terminal-box"></div>
  </a-card>
</template>

<script setup>
import { ref, onBeforeUnmount, watch } from 'vue'
import { message } from 'ant-design-vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

const props = defineProps({
  token: String,
  pipelineName: String,
  pipelineId: String,
  systemId: String,
  initialProjectPath: {
    type: String,
    default: '',
  },
  releaseUnitName: {
    type: String,
    default: '',
  },
  repositoryType: {
    type: String,
    default: '',
  },
  autoMerge: {
    type: Boolean,
    default: true,
  },
  autoTag: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['deploy-start', 'deploy-stop'])

const projectPath = ref('')
const isDeploying = ref(false)
const status = ref('idle')
const terminalEl = ref(null)

let ws = null
let terminal = null
let fitAddon = null
let currentBatchId = ''

const statusLabel = {
  idle: 'Idle',
  connecting: 'Connecting',
  running: 'Running',
  done: 'Done',
  error: 'Error',
}

const statusColor = {
  idle: 'default',
  connecting: 'processing',
  running: 'blue',
  done: 'success',
  error: 'error',
}

watch(
  () => props.initialProjectPath,
  (val) => {
    if (typeof val === 'string' && val.trim()) {
      projectPath.value = val.trim()
    }
  },
  { immediate: true }
)

const getWsUrl = (batchId) => {
  const path = `/ws/execute/${encodeURIComponent(batchId)}?token=${encodeURIComponent(props.token)}`
  return path
}

const writeLine = (line) => {
  if (!terminal) return
  terminal.write(`${line}\r\n`)
}

const clearTerminal = () => {
  if (!terminal) return
  terminal.clear()
  terminal.reset()
}

const ensureTerminalMounted = () => {
  if (!terminal) {
    fitAddon = new FitAddon()
    terminal = new Terminal({
      cursorBlink: true,
      convertEol: true,
      fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
      fontSize: 13,
      theme: {
        background: '#0f1722',
        foreground: '#d6e2ff',
        cursor: '#7be0ad',
      },
    })
    terminal.loadAddon(fitAddon)
  }

  if (terminalEl.value && !terminal.element) {
    terminal.open(terminalEl.value)
    fitAddon.fit()
    writeLine('[client] terminal ready')
    return
  }

  if (terminalEl.value) {
    fitAddon.fit()
  }
}

const handleStart = async () => {
  if (isDeploying.value) return
  if (!props.pipelineId || !props.systemId) {
    message.warning('缺少流水线或系统信息')
    return
  }

  ensureTerminalMounted()
  clearTerminal()
  status.value = 'connecting'

  let batchId = ''
  try {
    const submitRes = await fetch(`/api/pipelines/${encodeURIComponent(props.pipelineId)}/execute?system_id=${encodeURIComponent(props.systemId)}&triggered_by=manual`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${props.token}`,
      },
    })
    if (!submitRes.ok) {
      const errPayload = await submitRes.json()
      throw new Error(errPayload.error || '提交执行失败')
    }
    const submitPayload = await submitRes.json()
    batchId = submitPayload.batch_id
    currentBatchId = batchId
  } catch (err) {
    status.value = 'error'
    message.error(err.message || '提交执行失败')
    return
  }

  ws = new WebSocket(getWsUrl(batchId))
  ws.onopen = () => {
    isDeploying.value = true
    status.value = 'running'
    writeLine(`\u001b[36m[client]\u001b[0m connected to execution stream, batch=${batchId}`)
    emit('deploy-start')
  }

  ws.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data)
      if (payload.type === 'log') {
        writeLine(payload.line ?? '')
        return
      }
      writeLine(payload.message ?? JSON.stringify(payload))
    } catch {
      writeLine(String(event.data))
    }
  }

  ws.onerror = () => {
    status.value = 'error'
    message.error('WebSocket error, check backend service status.')
  }

  ws.onclose = () => {
    isDeploying.value = false
    if (status.value !== 'error') {
      status.value = 'done'
    }
    writeLine('\u001b[33m[client]\u001b[0m deployment stream closed')
    ws = null
  }
}

const handleStop = () => {
  if (currentBatchId && props.token) {
    fetch(`/api/executions/${encodeURIComponent(currentBatchId)}/cancel`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${props.token}`,
      },
    }).catch(() => {})
  }
  if (ws) {
    ws.close(1000, 'user requested stop')
  }
  isDeploying.value = false
  status.value = 'idle'
  currentBatchId = ''
  emit('deploy-stop')
}

onBeforeUnmount(() => {
  if (ws) {
    ws.close(1000, 'component unmount')
  }
  if (terminal) {
    terminal.dispose()
  }
})
</script>

<style scoped>
.exec-panel {
  border-radius: 12px;
  box-shadow: 0 10px 24px rgba(17, 36, 64, 0.08);
}

.selected-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.control-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 10px;
  margin-bottom: 10px;
}

.terminal-box {
  width: 100%;
  height: 36vh;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #1d2f45;
}

@media (max-width: 900px) {
  .control-row {
    grid-template-columns: 1fr;
  }

  .terminal-box {
    height: 48vh;
  }
}
</style>
