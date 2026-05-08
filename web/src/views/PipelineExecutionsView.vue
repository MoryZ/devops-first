<template>
  <div class="executions-view">
    <a-card :bordered="false" class="head-card">
      <div class="line-one">
        <a class="link" @click="goBack">流水线</a>
        <span class="sep">/</span>
        <span>{{ pipelineName || pipelineId }}</span>
        <a-button type="text" size="small" class="settings-btn" @click="openSettings">⚙️ 设置</a-button>
      </div>
      <div class="line-two">
        <span>关联研发计划: {{ planVersion || '-' }}</span>
        <a-button type="text" size="small" class="copy-btn" @click="copyPlanVersion">复制</a-button>
      </div>
    </a-card>

    <ExecutionHistoryPanel :token="token" :pipelineId="pipelineId" />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { getPipelineConfig } from '../api/pipelines'
import { listPlansBySystem } from '../api/plans'
import { useWorkspaceStore } from '../stores/workspace'
import ExecutionHistoryPanel from '../components/ExecutionHistoryPanel.vue'

const route = useRoute()
const router = useRouter()
const workspace = useWorkspaceStore()

const token = computed(() => localStorage.getItem('token') || '')
const pipelineId = computed(() => String(route.params.id || ''))
const pipelineName = computed(() => String(route.query.name || ''))
const planVersion = ref(String(route.query.plan_version || ''))

const resolvePlanFromReleaseUnit = async () => {
  if (!token.value || !pipelineId.value) return ''
  try {
    const cfg = await getPipelineConfig(token.value, pipelineId.value)
    const releaseUnitId = String(cfg?.release_unit_id || cfg?.releaseUnitId || '').trim()
    if (!releaseUnitId) return ''

    const keys = Object.keys(localStorage).filter((key) => key.startsWith('releaseUnits:'))
    for (const key of keys) {
      const raw = localStorage.getItem(key)
      if (!raw) continue
      try {
        const units = JSON.parse(raw)
        if (!Array.isArray(units)) continue
        const matched = units.find((unit) => String(unit.id) === releaseUnitId)
        if (matched?.name) return String(matched.name)
      } catch {
        // ignore invalid cached release unit payload
      }
    }
  } catch {
    // ignore fallback resolve failures
  }
  return ''
}

const resolvePlanVersion = async () => {
  if (planVersion.value) return
  if (!token.value) {
    planVersion.value = '-'
    return
  }

  try {
    if (workspace.selectedSystemId && workspace.selectedPlanId) {
      const plansData = await listPlansBySystem(token.value, workspace.selectedSystemId)
      const plan = (plansData.items || []).find((item) => String(item.ID || item.id) === String(workspace.selectedPlanId))
      const v = String(plan?.Version || plan?.version || '').trim()
      if (v) {
        planVersion.value = v
        return
      }
    }
  } catch {
    // continue to other fallbacks
  }

  const releaseUnitName = await resolvePlanFromReleaseUnit()
  planVersion.value = releaseUnitName || '-'
}

const goBack = () => {
  router.push('/workspace')
}

const openSettings = () => {
  message.info('流水线设置功能敬请期待')
}

const copyPlanVersion = () => {
  if (!planVersion.value || planVersion.value === '-') {
    message.warning('暂无计划版本')
    return
  }
  navigator.clipboard.writeText(planVersion.value).then(() => {
    message.success('已复制到剪贴板')
  }).catch(() => {
    message.error('复制失败')
  })
}

onMounted(() => {
  resolvePlanVersion()
})
</script>

<style scoped>
.executions-view {
  min-height: 100vh;
  background: #f3f5f9;
  padding: 18px;
}

.head-card {
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
  cursor: pointer;
  padding: 4px 0;
}

.link:hover {
  text-decoration: underline;
}

.sep {
  color: #8ca0bf;
}

.settings-btn {
  margin-left: 12px;
  font-size: 14px;
  color: #2d67d8;
}

.settings-btn:hover {
  opacity: 0.8;
}

.line-two {
  margin-top: 8px;
  color: #536b8f;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.copy-btn {
  font-size: 12px;
  color: #2d67d8;
  padding: 0 4px;
}

.copy-btn:hover {
  opacity: 0.8;
}
</style>
