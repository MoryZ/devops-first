<template>
  <div class="all-pipelines-panel">
    <div class="panel-head">
      <div class="title">全部流水线</div>
      <a-button @click="goBack">返回流水线看板</a-button>
    </div>

    <a-table
      :data-source="rows"
      :columns="columns"
      row-key="id"
      :pagination="pagination"
      :loading="loading"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'status'">
          <a-tag :color="record.status === 'active' ? 'green' : 'default'">{{ record.status || 'active' }}</a-tag>
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-button type="link" size="small" @click="openExecutions(record)">执行</a-button>
          <a-button type="link" size="small" @click="openBPM(record)">编排</a-button>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useWorkspaceStore } from '../stores/workspace'
import { listSystemPipelines } from '../api/systems'
import { listPlansBySystem } from '../api/plans'

const props = defineProps({
  token: {
    type: String,
    default: '',
  },
})

const router = useRouter()
const workspace = useWorkspaceStore()
const rows = ref([])
const loading = ref(false)

const columns = [
  { title: '流水线ID', dataIndex: 'id', key: 'id', width: 220 },
  { title: '流水线名称', dataIndex: 'name', key: 'name' },
  { title: '关联版本计划', dataIndex: 'plan_name', key: 'plan_name', width: 180 },
  { title: '运行次数', dataIndex: 'run_count', key: 'run_count', width: 100 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '最后活动时间', dataIndex: 'updated_at', key: 'updated_at', width: 180 },
  { title: '操作', key: 'actions', width: 140 },
]

const pagination = computed(() => ({
  pageSize: 10,
  showSizeChanger: false,
  showTotal: (total) => `共 ${total} 条`,
}))

const loadPipelines = async () => {
  if (!workspace.selectedSystemId) {
    rows.value = []
    return
  }

  loading.value = true
  try {
    const [pipelinesData, plansData] = await Promise.all([
      listSystemPipelines(props.token, workspace.selectedSystemId),
      listPlansBySystem(props.token, workspace.selectedSystemId).catch(() => ({ items: [] })),
    ])
    const planMap = (plansData.items || []).reduce((acc, p) => {
      acc[p.ID || p.id] = p.Version || p.version
      return acc
    }, {})

    rows.value = (pipelinesData.items || []).map((item) => ({
      id: item.ID || item.id,
      name: item.Name || item.name,
      plan_id: item.PlanID || item.plan_id,
      plan_name: planMap[item.PlanID || item.plan_id] || '-',
      run_count: 1,
      status: 'active',
      updated_at: new Date((item.UpdatedAt || item.updated_at || Date.now())).toLocaleString(),
    }))
  } catch (err) {
    message.error('加载流水线失败: ' + err.message)
  } finally {
    loading.value = false
  }
}

watch(
  () => workspace.selectedSystemId,
  async () => {
    await loadPipelines()
  },
  { immediate: true }
)

const openExecutions = (record) => {
  router.push({
    path: `/pipelines/${record.id}/executions`,
    query: {
      name: record.name,
      plan_version: record.plan_name || '',
    },
  })
}

const openBPM = (record) => {
  router.push({
    path: `/pipelines/${record.id}/bpm`,
    query: {
      name: record.name,
      plan_version: record.plan_name || '',
    },
  })
}

const goBack = () => {
  router.push('/workspace')
}
</script>

<style scoped>
.all-pipelines-panel {
  border: 1px solid #dde6f2;
  border-radius: 12px;
  background: #fff;
  padding: 14px;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.title {
  font-size: 20px;
  font-weight: 700;
  color: #1d2e44;
}
</style>
