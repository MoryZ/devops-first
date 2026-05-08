<template>
  <div class="dashboard-page">
    <Header
      :token="token"
      :userInitial="userInitial"
      @logout="$emit('logout')"
    />

    <div class="board-main">
      <aside class="sidebar">
        <button class="side-item" :class="{ active: activePage === 'workspace' }" @click="setActivePage('workspace')" title="流水线">
          <DeploymentUnitOutlined />
          <span class="side-label">流水线</span>
        </button>
        <button class="side-item" :class="{ active: activePage === 'iteration-plans' }" @click="setActivePage('iteration-plans')" title="迭代计划">
          <ProfileOutlined />
          <span class="side-label">迭代计划</span>
        </button>
        <button class="side-item" :class="{ active: activePage === 'release-units' }" @click="setActivePage('release-units')" title="发布单元">
          <ApartmentOutlined />
          <span class="side-label">发布单元</span>
        </button>
        <button class="side-item" :class="{ active: activePage === 'global-vars' }" @click="setActivePage('global-vars')" title="全局变量">
          <AppstoreOutlined />
          <span class="side-label">全局变量</span>
        </button>
        <button class="side-item" disabled title="敬请期待">
          <BuildOutlined />
          <span class="side-label">构建</span>
        </button>
        <button class="side-item" disabled title="敬请期待">
          <SettingOutlined />
          <span class="side-label">设置</span>
        </button>
      </aside>

      <div class="board-content">
        <div v-if="activePage === 'iteration-plans'" class="iteration-layout">
          <div class="iteration-left">
            <PlanList :token="token" />
          </div>
          <div class="iteration-right">
            <PlanReleaseUnitsPanel :token="token" />
          </div>
        </div>

        <div v-else-if="activePage === 'workspace'" class="pipeline-layout">
          <PipelineBoard :token="token" @edit-pipeline="openPipelineConfig" />
        </div>

        <PipelineAllListPanel v-else-if="activePage === 'pipeline-all'" :token="token" />

        <GlobalVariablesView v-else-if="activePage === 'global-vars'" />

        <ReleaseUnitPage v-else />
      </div>
    </div>

    <PipelineConfigDrawer
      :open="configDrawerOpen"
      :token="token"
      :pipelineId="editingPipeline?.id"
      :pipelineName="editingPipeline?.name"
      :pipeline="editingPipeline"
      :systemId="editingPipeline?.system_id || workspace.selectedSystemId"
      @close="closePipelineConfig"
      @save="handleConfigSave"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ApartmentOutlined,
  AppstoreOutlined,
  BuildOutlined,
  DeploymentUnitOutlined,
  ProfileOutlined,
  SettingOutlined,
} from '@ant-design/icons-vue'
import Header from './Header.vue'
import PlanList from './PlanList.vue'
import PlanReleaseUnitsPanel from './PlanReleaseUnitsPanel.vue'
import PipelineAllListPanel from './PipelineAllListPanel.vue'
import PipelineBoard from './PipelineBoard.vue'
import PipelineConfigDrawer from './PipelineConfigDrawer.vue'
import ReleaseUnitPage from './ReleaseUnitPage.vue'
import GlobalVariablesView from '../views/GlobalVariablesView.vue'
import { useWorkspaceStore } from '../stores/workspace'

const props = defineProps({
  token: String,
  currentUser: Object,
})

defineEmits(['logout'])

const configDrawerOpen = ref(false)
const editingPipeline = ref(null)
const workspace = useWorkspaceStore()
const route = useRoute()
const router = useRouter()

const activePage = computed(() => {
  if (route.name === 'pipeline-all') return 'pipeline-all'
  if (route.name === 'iteration-plans') return 'iteration-plans'
  if (route.name === 'global-vars') return 'global-vars'
  if (route.name === 'release-units') return 'release-units'
  return 'workspace'
})

const setActivePage = (page) => {
  if (page === 'iteration-plans') {
    router.push('/plans')
    return
  }
  if (page === 'global-vars') {
    router.push('/global-vars')
    return
  }
  if (page === 'release-units') {
    router.push('/release-units')
    return
  }
  router.push('/workspace')
}

const userInitial = computed(() => {
  return props.currentUser?.username?.slice(0, 1)?.toUpperCase() || 'U'
})

const openPipelineConfig = (pipeline) => {
  editingPipeline.value = pipeline
  configDrawerOpen.value = true
}

const closePipelineConfig = () => {
  configDrawerOpen.value = false
  editingPipeline.value = null
}

const handleConfigSave = () => {
  closePipelineConfig()
}
</script>

<style scoped>
.dashboard-page {
  min-height: 100vh;
  background: #f3f5f9;
  display: flex;
  flex-direction: column;
}

.board-main {
  display: grid;
  grid-template-columns: 160px 1fr;
  min-height: calc(100vh - 56px);
  flex: 1;
}

.sidebar {
  background: #0a1630;
  padding: 14px 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.side-item {
  width: 100%;
  height: auto;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #b6c5e5;
  cursor: pointer;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: flex-start;
  gap: 4px;
  padding: 6px 8px;
  font-size: 14px;
  transition: all 0.2s ease;
}

.side-item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.side-item.active,
.side-item:hover:not(:disabled) {
  background: #19325c;
  color: #f4f8ff;
}

.side-item svg {
  font-size: 14px;
  flex-shrink: 0;
  min-width: 14px;
}

.side-label {
  font-size: 14px;
  line-height: 1.2;
  flex: 1;
}

.board-content {
  padding: 18px;
  overflow: auto;
}

.plans-panel {
  max-width: 296px;
}

.iteration-layout {
  display: grid;
  grid-template-columns: 296px 1fr;
  gap: 18px;
  min-width: 0;
  min-height: calc(100vh - 92px);
  align-items: stretch;
}

.iteration-left {
  min-width: 0;
  display: flex;
}

.iteration-right {
  min-width: 0;
}

.pipeline-layout {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

@media (max-width: 1400px) {
  .board-main {
    grid-template-columns: 1fr;
  }

  .sidebar {
    flex-direction: row;
    justify-content: flex-start;
    padding: 8px;
    gap: 8px;
    overflow-x: auto;
    border-bottom: 1px solid #e8e8e8;
  }

  .side-item {
    width: auto;
    min-width: 80px;
    flex-shrink: 0;
  }
}
</style>
