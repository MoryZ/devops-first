<template>
  <div class="dashboard-page">
    <Header
      :token="token"
      :userInitial="userInitial"
      @logout="$emit('logout')"
    />

    <div class="board-main">
      <aside class="sidebar">
        <button class="side-item" :class="{ active: activePage === 'workspace' }" @click="setActivePage('workspace')">
          <DeploymentUnitOutlined />
        </button>
        <button class="side-item" :class="{ active: activePage === 'release-units' }" @click="setActivePage('release-units')">
          <ApartmentOutlined />
        </button>
        <button class="side-item" :class="{ active: activePage === 'global-vars' }" @click="setActivePage('global-vars')">
          <AppstoreOutlined />
        </button>
        <button class="side-item"><BuildOutlined /></button>
        <button class="side-item"><SettingOutlined /></button>

      </aside>

      <div class="board-content">
        <div class="context-strip">
          <span v-for="(item, idx) in contextBreadcrumbs" :key="idx" class="context-item" :class="{ active: idx === contextBreadcrumbs.length - 1 }">
            {{ item }}
            <span v-if="idx < contextBreadcrumbs.length - 1" class="context-sep"> / </span>
          </span>
        </div>

        <div v-if="activePage === 'workspace'" class="two-panel-layout">
          <div class="panel-column plan-column">
            <PlanList :token="token" />
          </div>

          <div class="panel-column pipeline-column">
            <PipelineBoard :token="token" @edit-pipeline="openPipelineConfig" />
          </div>
        </div>

        <GlobalVariablesView v-else-if="activePage === 'global-vars'" />

        <ReleaseUnitPage v-else />
      </div>
    </div>

    <!-- Pipeline Config Drawer -->
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
  SettingOutlined,
} from '@ant-design/icons-vue'
import Header from './Header.vue'
import PlanList from './PlanList.vue'
import PipelineBoard from './PipelineBoard.vue'
import PipelineConfigDrawer from './PipelineConfigDrawer.vue'
import ReleaseUnitPage from './ReleaseUnitPage.vue'
import GlobalVariablesView from '../views/GlobalVariablesView.vue'
import { useWorkspaceStore } from '../stores/workspace'

const props = defineProps({
  token: String,
  currentUser: Object,
})

const emit = defineEmits(['logout'])

const configDrawerOpen = ref(false)
const editingPipeline = ref(null)
const workspace = useWorkspaceStore()
const route = useRoute()
const router = useRouter()

const activePage = computed(() => {
  if (route.name === 'global-vars') return 'global-vars'
  if (route.name === 'release-units') return 'release-units'
  return 'workspace'
})

const contextBreadcrumbs = computed(() => {
  switch (activePage.value) {
    case 'global-vars':
      return ['我的工作台', '全局变量']
    case 'release-units':
      return ['我的工作台', '发布单元']
    default:
      return ['我的工作台', '业务需求管理']
  }
})

const setActivePage = (page) => {
  if (page === 'global-vars') {
    router.push('/global-vars')
  } else if (page === 'release-units') {
    router.push('/release-units')
  } else {
    router.push('/workspace')
  }
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
  grid-template-columns: 62px 1fr;
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
  width: 40px;
  height: 40px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #b6c5e5;
  font-size: 18px;
  cursor: pointer;
}

.side-item.active,
.side-item:hover {
  background: #19325c;
  color: #f4f8ff;
}

.board-content {
  padding: 18px;
  overflow: auto;
}

.context-strip {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  padding: 10px 12px;
  border: 1px solid #dde4f1;
  border-radius: 10px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.context-item {
  color: #5f6f88;
  font-size: 13px;
}

.context-item.active {
  color: #1f56ba;
  font-weight: 600;
}

.context-sep {
  color: #9faecc;
}

.two-panel-layout {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 14px;
  height: 100%;
}

.panel-column {
  min-height: 0;
}

.plan-column {
  overflow: hidden;
}

.pipeline-column {
  display: flex;
  flex-direction: column;
}

@media (max-width: 1400px) {
  .two-panel-layout {
    grid-template-columns: 1fr;
  }

  .board-main {
    grid-template-columns: 1fr;
  }

  .sidebar {
    flex-direction: row;
    justify-content: center;
    padding: 8px;
  }
}
</style>
