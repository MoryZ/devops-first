<template>
  <aside class="version-sidebar">
    <div class="sidebar-head">
      <span class="title">版本计划</span>
      <div class="head-actions">
        <a-button size="small" @click="reloadPlans">排序</a-button>
        <a-button size="small" @click="reloadPlans">筛选</a-button>
      </div>
    </div>

    <div class="plan-list">
      <button
        v-for="plan in plans"
        :key="plan.id"
        class="plan-item"
        :class="{ active: selectedPlanId === plan.id }"
        @click="selectPlan(plan.id)"
      >
        <div class="plan-version">{{ plan.version }}</div>
        <div class="plan-meta">{{ plan.status || 'planning' }} · {{ plan.planned_date || '-' }}</div>
      </button>
      <div v-if="plans.length === 0" class="empty">暂无版本计划</div>
    </div>

    <a-button block type="default" class="view-all" @click="goAllPipelines">查看全部</a-button>
  </aside>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useWorkspaceStore } from '../stores/workspace'
import { listPlansBySystem } from '../api/plans'

const props = defineProps({
  token: {
    type: String,
    default: '',
  },
})

const router = useRouter()
const workspace = useWorkspaceStore()
const plans = ref([])

const selectedPlanId = computed(() => workspace.selectedPlanId)

const loadPlans = async () => {
  if (!workspace.selectedSystemId) {
    plans.value = []
    workspace.selectPlan('')
    return
  }

  try {
    const data = await listPlansBySystem(props.token, workspace.selectedSystemId)
    plans.value = (data.items || []).map((item) => ({
      id: item.ID || item.id,
      version: item.Version || item.version,
      status: item.Status || item.status,
      planned_date: item.PlannedDate || item.planned_date,
    }))
    if (!plans.value.some((p) => p.id === workspace.selectedPlanId)) {
      workspace.selectPlan(plans.value[0]?.id || '')
    }
  } catch (err) {
    message.error('加载版本计划失败: ' + err.message)
  }
}

watch(
  () => workspace.selectedSystemId,
  async () => {
    await loadPlans()
  },
  { immediate: true }
)

const selectPlan = (planId) => {
  workspace.selectPlan(planId)
}

const reloadPlans = async () => {
  await loadPlans()
}

const goAllPipelines = () => {
  router.push('/pipelines/all')
}
</script>

<style scoped>
.version-sidebar {
  border: 1px solid #dbe3ef;
  border-radius: 12px;
  background: #fff;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 400px;
}

.sidebar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title {
  font-size: 14px;
  font-weight: 700;
  color: #1e2d43;
}

.head-actions {
  display: flex;
  gap: 6px;
}

.plan-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: auto;
  min-height: 0;
}

.plan-item {
  border: 1px solid #e2e8f3;
  border-radius: 8px;
  padding: 8px;
  text-align: left;
  background: #f8fbff;
  cursor: pointer;
}

.plan-item.active {
  border-color: #2d67d8;
  background: #edf4ff;
}

.plan-version {
  font-size: 13px;
  font-weight: 600;
  color: #234269;
}

.plan-meta {
  font-size: 12px;
  color: #7284a0;
  margin-top: 3px;
}

.empty {
  color: #8b9bb5;
  font-size: 12px;
  padding: 12px 4px;
}

.view-all {
  margin-top: auto;
}
</style>
