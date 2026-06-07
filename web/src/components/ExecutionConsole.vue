<template>
  <div class="exec-panel">
    <div class="panel-header">
      <div class="header-left">
        <span class="panel-icon">▶</span>
        <h3 class="panel-title">执行控制台</h3>
      </div>
      <div class="status-indicators">
        <span class="status-dot" :class="statusClass"></span>
        <span class="status-text">{{ statusLabel[status] }}</span>
      </div>
    </div>

    <div class="selected-line">
      <a-tag color="processing" class="custom-tag">
        <span class="tag-prefix">◆</span>
        {{ pipelineName || '未选择' }}
      </a-tag>
      <a-tag v-if="releaseUnitName" color="geekblue" class="custom-tag">
        <span class="tag-prefix">◈</span>
        发布单元：{{ releaseUnitName }}
      </a-tag>
      <a-tag v-if="repositoryType" color="purple" class="custom-tag">
        <span class="tag-prefix">⬡</span>
        触发源：{{ repositoryType }}
      </a-tag>
      <a-tag :color="autoMerge ? 'blue' : 'default'" class="custom-tag">
        <span class="tag-prefix">◆</span>
        自动归并：{{ autoMerge ? 'ON' : 'OFF' }}
      </a-tag>
      <a-tag :color="autoTag ? 'cyan' : 'default'" class="custom-tag">
        <span class="tag-prefix">◆</span>
        自动Tag：{{ autoTag ? 'ON' : 'OFF' }}
      </a-tag>
    </div>

    <div class="control-row">
      <a-input
        v-model:value="projectPath"
        size="large"
        allow-clear
        placeholder="输入项目路径，例如 /Users/name/projects/order-service"
        class="path-input"
      >
        <template #prefix>
          <FolderOpenOutlined />
        </template>
      </a-input>
      <a-button type="primary" size="large" :disabled="isDeploying" @click="handleStart" class="start-btn">
        <PlayCircleOutlined />
        启动部署
      </a-button>
      <a-button danger size="large" :disabled="!isDeploying" @click="handleStop" class="stop-btn">
        <StopOutlined />
        停止
      </a-button>
    </div>
    <div ref="terminalEl" class="terminal-box"></div>
  </div>
</template>

<script setup>
import { ref, computed, onBeforeUnmount, watch } from 'vue'
import { message } from 'ant-design-vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { cancelExecutionBatch, startPipelineExecution } from '../api/executions'
import { PlayCircleOutlined, StopOutlined, FolderOpenOutlined } from '@ant-design/icons-vue'
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
  idle: '空闲',
  connecting: '连接中',
  running: '运行中',
  done: '完成',
  error: '错误',
}

const statusClass = computed(() => {
  const map = {
    idle: 'status-idle',
    connecting: 'status-connecting',
    running: 'status-running',
    done: 'status-done',
    error: 'status-error',
  }
  return map[status.value] || 'status-idle'
})

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
      fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
      fontSize: 13,
      theme: {
        background: '#0f1722',
        foreground: '#d6e2ff',
        cursor: '#00d4ff',
        selection: 'rgba(0, 212, 255, 0.3)',
      },
    })
    terminal.loadAddon(fitAddon)
  }

  if (terminalEl.value && !terminal.element) {
    terminal.open(terminalEl.value)
    fitAddon.fit()
    writeLine('\x1b[36m[client]\x1b[0m terminal ready')
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
    const submitPayload = await startPipelineExecution(props.token, props.pipelineId, props.systemId)
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
    writeLine(`\x1b[36m[client]\x1b[0m connected to execution stream, batch=${batchId}`)
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
    writeLine('\x1b[33m[client]\x1b[0m deployment stream closed')
    ws = null
  }
}

const handleStop = () => {
  if (currentBatchId && props.token) {
    cancelExecutionBatch(props.token, currentBatchId).catch(() => {})
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
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  animation: slideUp 0.4s ease-out;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-tertiary);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.panel-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border-radius: 6px;
  font-size: 10px;
  color: white;
}

.panel-title {
  margin: 0;
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 700;
  font-family: var(--font-display);
  letter-spacing: 0.02em;
}

.status-indicators {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-muted);
}

.status-dot.status-idle {
  background: var(--text-muted);
}

.status-dot.status-connecting {
  background: var(--accent-warning);
  animation: pulse 1s ease-in-out infinite;
}

.status-dot.status-running {
  background: var(--accent-primary);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.6);
  animation: pulse 1s ease-in-out infinite;
}

.status-dot.status-done {
  background: var(--accent-success);
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.status-dot.status-error {
  background: var(--accent-danger);
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
}

.status-text {
  font-size: 13px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
  font-weight: 500;
}

.selected-line {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color);
  flex-wrap: wrap;
}

.custom-tag {
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.02em;
  padding: 2px 10px;
  border: 1px solid;
  background: transparent;
  font-family: var(--font-display);
}

.custom-tag :deep(.ant-tag-text) {
  display: flex;
  align-items: center;
  gap: 4px;
}

.tag-prefix {
  font-size: 10px;
  opacity: 0.7;
}

.control-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-tertiary);
}

.path-input {
  border-radius: var(--radius-md) !important;
  background: var(--bg-secondary) !important;
  border: 1px solid var(--border-color-light) !important;
}

.path-input :deep(.ant-input) {
  background: transparent !important;
  border: none !important;
  color: var(--text-primary) !important;
  font-family: var(--font-mono) !important;
  font-size: 13px !important;
}

.path-input :deep(.ant-input::placeholder) {
  color: var(--text-muted) !important;
}

.path-input :deep(.ant-input-prefix) {
  color: var(--text-tertiary) !important;
  margin-right: 8px !important;
}

.path-input:hover,
.path-input:focus-within {
  border-color: var(--accent-primary) !important;
}

.start-btn {
  height: 42px !important;
  padding: 0 24px !important;
  border-radius: var(--radius-md) !important;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-info)) !important;
  border: none !important;
  font-weight: 600 !important;
  font-family: var(--font-display) !important;
  box-shadow: 0 4px 15px rgba(0, 212, 255, 0.25) !important;
  transition: all var(--transition-base) !important;
}

.start-btn:hover:not(:disabled) {
  transform: translateY(-2px) !important;
  box-shadow: 0 6px 20px rgba(0, 212, 255, 0.4) !important;
}

.stop-btn {
  height: 42px !important;
  padding: 0 24px !important;
  border-radius: var(--radius-md) !important;
  border: 1px solid var(--accent-danger) !important;
  background: transparent !important;
  color: var(--accent-danger) !important;
  font-weight: 600 !important;
  font-family: var(--font-display) !important;
  transition: all var(--transition-base) !important;
}

.stop-btn:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.1) !important;
  box-shadow: 0 0 15px rgba(239, 68, 68, 0.2) !important;
}

.terminal-box {
  width: 100%;
  height: 36vh;
  overflow: hidden;
  background: #0a0f1a;
  position: relative;
}

.terminal-box::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--accent-primary), transparent);
  opacity: 0.3;
  z-index: 1;
}

.terminal-box :deep(.xterm) {
  padding: 12px;
  height: 100%;
}

.terminal-box :deep(.xterm-viewport) {
  background: transparent !important;
}

.terminal-box :deep(.xterm-rows) {
  font-family: 'JetBrains Mono', 'Fira Code', monospace !important;
}

@media (max-width: 900px) {
  .control-row {
    grid-template-columns: 1fr;
  }

  .terminal-box {
    height: 48vh;
  }

  .start-btn,
  .stop-btn {
    width: 100%;
  }
}
</style>
