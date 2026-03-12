<template>
  <div class="plan-list-panel">
    <div class="plan-search-row">
      <a-input-search v-model:value="keyword" placeholder="输入关键词/负责人/版本" allow-clear />
    </div>

    <div class="panel-header">
      <span class="panel-title">迭代计划</span>
    </div>
    <div class="plan-filter-row">
      <a-tag class="filter-tag action-tag" :class="{ disabled: !workspace.selectedSystemId }" @click="showCreateModal">
        <PlusOutlined />
        创建计划
      </a-tag>
      <a-tag :color="activeFilter === 'mine' ? 'blue' : 'default'" class="filter-tag" @click="activeFilter = 'mine'">我的计划</a-tag>
      <a-tag :color="activeFilter === 'all' ? 'blue' : 'default'" class="filter-tag" @click="activeFilter = 'all'">全部计划</a-tag>
    </div>
    <div class="plan-list-header-row">
      <span>版本</span>
      <span>状态</span>
      <span>计划日期</span>
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
        <span class="plan-date">{{ plan.planned_date || '-' }}</span>
      </div>

      <div v-if="filteredPlans.length === 0" class="empty-hint">当前筛选条件下暂无迭代计划</div>
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
import { PlusOutlined } from '@ant-design/icons-vue'
import { useWorkspaceStore } from '../stores/workspace'

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

const loadPlans = async (systemId) => {
  try {
    const res = await fetch(`/api/systems/${systemId}/plans`, {
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (res.ok) {
      const data = await res.json()
      plans.value = (data.items || []).map((item) => ({
        id: item.ID || item.id,
        system_id: item.SystemID || item.system_id,
        version: item.Version || item.version,
        status: item.Status || item.status || 'planning',
        planned_date: item.PlannedDate || item.planned_date,
        description: item.Description || item.description,
        created_at: item.CreatedAt || item.created_at,
        updated_at: item.UpdatedAt || item.updated_at,
      }))
      if (!plans.value.some((p) => p.id === selectedPlanId.value)) {
        selectedPlanId.value = ''
      }
      activeFilter.value = 'all'
    } else {
      message.error('加载计划列表失败')
    }
  } catch (err) {
    message.error('加载计划列表失败: ' + err.message)
  }
}

const selectPlan = (plan) => {
  selectedPlanId.value = plan.id
  workspace.selectPlan(plan.id)
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
    const res = await fetch(`/api/systems/${workspace.selectedSystemId}/plans`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${props.token}`,
      },
      body: JSON.stringify(newPlan.value),
    })
    if (res.ok) {
      const created = await res.json()
      createdPlanIds.value.add(created.ID || created.id)
      saveCreatedPlanIds()
      message.success('计划创建成功')
      createModalOpen.value = false
      await loadPlans(workspace.selectedSystemId)
    } else {
      const err = await res.json()
      message.error(err.error || '创建失败')
    }
  } catch (err) {
    message.error('网络错误: ' + err.message)
  }
}

loadCreatedPlanIds()
</script>

<style scoped>
.plan-list-panel {
  height: 100%;
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
  padding: 14px;
  border-bottom: 1px solid #e8e8e8;
}

.plan-filter-row {
  display: flex;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid #e8e8e8;
  background: #fff;
}

.filter-tag {
  cursor: pointer;
  user-select: none;
}

.action-tag {
  color: #1f56ba;
  border-color: #b7cff8;
}

.action-tag.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2d3d;
}

.plan-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
  background: linear-gradient(180deg, #f8faff 0%, #f2f6fc 100%);
}

.plan-list-header-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 10px;
  padding: 8px 12px;
  color: #7d8ea9;
  font-size: 12px;
  border-bottom: 1px solid #e3e9f3;
  background: #fff;
}

.plan-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 10px;
  align-items: center;
  margin-bottom: 8px;
  padding: 10px;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid #dbe4f2 !important;
  background: #ffffff;
  border-radius: 10px;
}

.plan-row:hover {
  background: #f4f8ff;
  border-color: #b4c7ea !important;
}

.plan-row.active {
  background: linear-gradient(135deg, #1d67d9 0%, #2d79ea 100%);
  border: 1px solid #1d67d9 !important;
}

.plan-version {
  font-weight: 600;
  color: #0c47a1;
  font-size: 14px;
  margin-bottom: 4px;
}

.plan-date {
  color: #8a92a8;
  font-size: 12px;
}

.empty-hint {
  text-align: center;
  color: #8fa0bb;
  font-size: 12px;
  margin-top: 20px;
}

.plan-row.active .plan-version,
.plan-row.active .plan-date {
  color: #eaf2ff;
}
</style>
