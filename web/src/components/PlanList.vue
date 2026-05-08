<template>
  <div class="plan-list-panel">
    <div class="panel-header">
      <span class="panel-title">迭代计划</span>
      <a-button type="primary" class="create-btn" :disabled="!workspace.selectedSystemId" @click="showCreateModal">
        新建
      </a-button>
    </div>

    <div class="plan-search-row">
      <a-input-search v-model:value="keyword" placeholder="输入关键词/负责人/版本" allow-clear />
    </div>

    <div class="plan-filter-row">
      <a-tag :color="activeFilter === 'mine' ? 'blue' : 'default'" class="filter-tag" @click="activeFilter = 'mine'">我的计划</a-tag>
    </div>

    <div class="plan-list">
      <div
        v-for="plan in filteredPlans"
        :key="plan.id"
        class="plan-row"
        :class="{ active: selectedPlanId === plan.id }"
        @click="selectPlan(plan)"
      >
        <span class="plan-version">{{ plan.version }}</span>
        <span class="plan-status"><a-tag :color="statusColor[plan.status]">{{ plan.status }}</a-tag></span>
      </div>

      <div v-if="filteredPlans.length === 0" class="empty-hint">当前筛选条件下暂无迭代计划</div>
    </div>

    <div class="plan-footer">
      <a-button block class="view-all-btn" @click="activeFilter = 'all'">查看全部计划</a-button>
    </div>

    <!-- Create Plan Modal -->
    <a-modal v-model:open="createModalOpen" title="新建迭代计划" @ok="handleCreatePlan">
      <a-form layout="vertical">
        <a-form-item label="版本号">
          <a-input v-model:value="newPlan.version" placeholder="e.g., 1.1.0" />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model:value="newPlan.status" placeholder="选择状态">
            <a-select-option value="planning">规划中</a-select-option>
            <a-select-option value="developing">开发中</a-select-option>
            <a-select-option value="released">已发布</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="计划发布日期">
          <a-input v-model:value="newPlan.planned_date" type="date" />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea v-model:value="newPlan.description" placeholder="计划描述（可选）" :rows="2" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useWorkspaceStore } from '../stores/workspace'
import { createPlan, listPipelinesByPlan, listPlansBySystem } from '../api/plans'
import { createSystemPipeline } from '../api/systems'
import { getPipelineConfig, upsertPipelineConfig } from '../api/pipelines'

const props = defineProps({
  token: String,
})

const workspace = useWorkspaceStore()

const plans = ref([])
const selectedPlanId = computed({
  get: () => workspace.selectedPlanId,
  set: (value) => workspace.selectPlan(value),
})
const keyword = ref('')
const activeFilter = ref('all')
const createdPlanIds = ref(new Set())
const createModalOpen = ref(false)
const newPlan = ref({
  version: '',
  status: 'planning',
  planned_date: '',
  description: '',
})

const statusColor = {
  planning: 'orange',
  developing: 'blue',
  released: 'green',
}

const filteredPlans = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  return plans.value.filter((plan) => {
    const hitKeyword = !kw || `${plan.version} ${plan.description || ''}`.toLowerCase().includes(kw)

    if (!hitKeyword) return false
    if (activeFilter.value === 'mine') return createdPlanIds.value.has(plan.id)
    return true
  })
})

const loadCreatedPlanIds = () => {
  try {
    const raw = localStorage.getItem('createdPlanIds')
    const arr = raw ? JSON.parse(raw) : []
    createdPlanIds.value = new Set(Array.isArray(arr) ? arr : [])
  } catch {
    createdPlanIds.value = new Set()
  }
}

const saveCreatedPlanIds = () => {
  localStorage.setItem('createdPlanIds', JSON.stringify(Array.from(createdPlanIds.value)))
}

const normalizePlan = (item) => ({
  id: String(item.ID || item.id),
  system_id: String(item.SystemID || item.system_id),
  version: item.Version || item.version,
  status: item.Status || item.status || 'planning',
  planned_date: item.PlannedDate || item.planned_date,
  description: item.Description || item.description,
  created_at: item.CreatedAt || item.created_at,
  updated_at: item.UpdatedAt || item.updated_at,
})

const planTime = (plan) => {
  const candidates = [plan.planned_date, plan.updated_at, plan.created_at]
  for (const v of candidates) {
    if (!v) continue
    const t = new Date(v).getTime()
    if (!Number.isNaN(t)) return t
  }
  return 0
}

const inheritReleaseUnitsFromLatestReleasedPlan = async (created) => {
  const newPlanId = String(created.ID || created.id || '')
  if (!newPlanId || !workspace.selectedSystemId) return 0

  const planData = await listPlansBySystem(props.token, workspace.selectedSystemId)
  const allPlans = (planData?.items || []).map(normalizePlan)

  const sourcePlan = allPlans
    .filter((p) => p.id !== newPlanId && String(p.status || '').toLowerCase() === 'released')
    .sort((a, b) => planTime(b) - planTime(a))[0]

  if (!sourcePlan) return 0

  const sourcePipelinesData = await listPipelinesByPlan(props.token, sourcePlan.id)
  const sourcePipelines = sourcePipelinesData?.items || []
  if (!sourcePipelines.length) return 0

  let inheritedCount = 0

  for (const src of sourcePipelines) {
    const srcPipelineId = String(src.ID || src.id)
    let srcConfig
    try {
      srcConfig = await getPipelineConfig(props.token, srcPipelineId)
    } catch {
      continue
    }

    const releaseUnitId = String(srcConfig?.release_unit_id || srcConfig?.releaseUnitId || '')
    if (!releaseUnitId) continue

    const srcName = src.Name || src.name || `pipeline-${inheritedCount + 1}`
    const appType = src.AppType || src.app_type || 'java'
    const srcDesc = src.Description || src.description || `${srcName} 自动继承`

    let createdPipeline
    try {
      createdPipeline = await createSystemPipeline(props.token, workspace.selectedSystemId, {
        plan_id: newPlanId,
        name: srcName,
        app_type: appType,
        description: srcDesc,
      })
    } catch {
      createdPipeline = await createSystemPipeline(props.token, workspace.selectedSystemId, {
        plan_id: newPlanId,
        name: `${srcName}-${created.Version || created.version || 'new'}`,
        app_type: appType,
        description: srcDesc,
      })
    }

    await upsertPipelineConfig(props.token, {
      ...srcConfig,
      pipeline_id: createdPipeline.ID || createdPipeline.id,
      name: createdPipeline.Name || createdPipeline.name || srcName,
      release_unit_id: releaseUnitId,
    })

    inheritedCount += 1
  }

  return inheritedCount
}

const loadPlans = async (systemId) => {
  try {
    const data = await listPlansBySystem(props.token, systemId)
    plans.value = (data?.items || []).map(normalizePlan)
    const currentPlanId = String(selectedPlanId.value || '')
    if (plans.value.length > 0 && !plans.value.some((p) => p.id === currentPlanId)) {
      selectedPlanId.value = plans.value[0].id
    }
    activeFilter.value = 'all'
  } catch (err) {
    message.error('加载计划列表失败: ' + (err?.message || '未知错误'))
  }
}

watch(
  () => workspace.selectedSystemId,
  async (newSystemId) => {
    if (newSystemId) {
      await loadPlans(newSystemId)
    } else {
      plans.value = []
      selectedPlanId.value = ''
    }
  },
  { immediate: true }
)

const selectPlan = (plan) => {
  selectedPlanId.value = String(plan.id)
}

const showCreateModal = () => {
  if (!workspace.selectedSystemId) {
    message.warning('请先选择系统')
    return
  }
  createModalOpen.value = true
  newPlan.value = { version: '', status: 'planning', planned_date: '', description: '' }
}

const handleCreatePlan = async () => {
  if (!newPlan.value.version) {
    message.warning('请输入版本号')
    return
  }

  try {
    const created = await createPlan(props.token, workspace.selectedSystemId, newPlan.value)
    let inheritedCount = 0
    try {
      inheritedCount = await inheritReleaseUnitsFromLatestReleasedPlan(created)
    } catch (inheritErr) {
      console.warn('Failed to inherit release units from latest released plan:', inheritErr)
    }

    createdPlanIds.value.add(created.ID || created.id)
    saveCreatedPlanIds()
    if (inheritedCount > 0) {
      message.success(`计划创建成功，已默认关联上个已发布迭代的 ${inheritedCount} 个发布单元`)
    } else {
      message.success('计划创建成功')
    }
    createModalOpen.value = false
    await loadPlans(workspace.selectedSystemId)
  } catch (err) {
    message.error('创建失败: ' + (err?.message || '未知错误'))
  }
}

loadCreatedPlanIds()
</script>

<style scoped>
.plan-list-panel {
  height: 100%;
  min-height: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 14px rgba(18, 35, 58, 0.07);
  overflow: hidden;
}

.plan-search-row {
  padding: 10px 10px 0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 12px 8px;
  border-bottom: 1px solid #e8e8e8;
}

.create-btn {
  min-width: 84px;
  height: 32px;
  border-radius: 8px;
}

.plan-filter-row {
  display: flex;
  gap: 8px;
  padding: 8px 10px 6px;
  background: #fff;
}

.filter-tag {
  cursor: pointer;
  user-select: none;
}

.panel-title {
  font-size: 18px;
  font-weight: 700;
  color: #1f2d3d;
}

.plan-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 8px 8px;
  background: linear-gradient(180deg, #f8faff 0%, #f2f6fc 100%);
}

.plan-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  align-items: center;
  margin-bottom: 6px;
  padding: 9px 10px;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid #dbe4f2 !important;
  background: #ffffff;
  border-radius: 9px;
}

.plan-row:hover {
  background: #f4f8ff;
  border-color: #b4c7ea !important;
}

.plan-row.active {
  background: linear-gradient(135deg, #3d7df2 0%, #2a66da 100%);
  border-color: #2a66da !important;
  box-shadow: 0 10px 20px rgba(42, 102, 218, 0.22);
}

.plan-row.active .plan-version {
  color: #fff;
}

.plan-version {
  font-size: 13px;
  font-weight: 700;
  color: #21406d;
  line-height: 1.3;
}

.plan-status :deep(.ant-tag) {
  margin-inline-end: 0;
  border-radius: 999px;
}

.empty-hint {
  padding: 40px 12px;
  text-align: center;
  color: #8b99b2;
  font-size: 12px;
}

.plan-footer {
  margin-top: auto;
  padding: 10px;
  border-top: 1px solid #e8eef7;
  background: #fff;
}

.view-all-btn {
  height: 36px;
  border-radius: 9px;
  border-color: #d8e2f0;
  color: #385680;
  background: #f7faff;
}

.plan-row.active .plan-version,
.plan-row.active .plan-date {
  color: #eaf2ff;
}
</style>
