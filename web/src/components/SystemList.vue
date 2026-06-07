<template>
  <div class="system-list-panel">
    <div class="panel-header">
      <span class="panel-title">系统列表</span>
      <a-button type="primary" @click="goNewSystem"><PlusOutlined /> 新建系统</a-button>
    </div>
    <div class="system-list">
      <a-card
        v-for="system in systems"
        :key="system.id"
        class="system-card"
        :bordered="false"
        :class="{ active: selectedSystemId === system.id }"
        @click="selectSystem(system)"
      >
        <div class="system-name">{{ system.name }}</div>
        <div class="system-desc">{{ system.description }}</div>
        <div class="system-status">
          <a-tag :color="statusColor[system.status]">{{ system.status }}</a-tag>
        </div>
      </a-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { PlusOutlined } from '@ant-design/icons-vue'
import { listSystems } from '../api/systems'
import { useWorkspaceStore } from '../stores/workspace'

const props = defineProps({
  token: String,
})

const emit = defineEmits(['select-system'])

const router = useRouter()
const workspace = useWorkspaceStore()

const systems = ref([])
const selectedSystemId = ref('')

const statusColor = {
  active: 'green',
  planning: 'orange',
  archived: 'gray',
}

onMounted(async () => {
  await loadSystems()
})

const loadSystems = async () => {
  try {
    const data = await listSystems(props.token)
    systems.value = data.items || []
    if (systems.value.length > 0 && !selectedSystemId.value) {
      selectSystem(systems.value[0])
    }
  } catch (err) {
    console.warn('Failed to load systems:', err)
  }
}

const selectSystem = (system) => {
  selectedSystemId.value = system.id
  emit('select-system', system)
}

const goNewSystem = () => {
  router.push('/systems/new')
}
</script>

<style scoped>
.system-list-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 14px rgba(18, 35, 58, 0.07);
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px;
  border-bottom: 1px solid #e8e8e8;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2d3d;
}

.system-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.system-card {
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid #e8e8e8 !important;
  background: #fafbff;
}

.system-card:hover {
  background: #f0f5ff;
  border-color: #d0d7e8 !important;
}

.system-card.active {
  background: #e6f4ff;
  border: 2px solid #1890ff !important;
}

.system-name {
  font-weight: 600;
  color: #0c47a1;
  margin-bottom: 4px;
}

.system-desc {
  font-size: 12px;
  color: #8a92a8;
  margin-bottom: 8px;
  line-height: 1.4;
}

.system-status {
  display: flex;
  gap: 4px;
}
</style>
