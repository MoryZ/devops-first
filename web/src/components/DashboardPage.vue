<template>
  <div class="dashboard-page">
    <Header
      :token="token"
      :userInitial="userInitial"
      @logout="$emit('logout')"
    />

    <div class="board-main">
      <aside class="sidebar">
        <div class="sidebar-glow"></div>
        <button
          v-for="item in menuItems"
          :key="item.key"
          class="side-item"
          :class="{ active: activePage === item.key, disabled: item.disabled }"
          :disabled="item.disabled"
          @click="!item.disabled && setActivePage(item.key)"
          :title="item.title"
        >
          <component :is="item.icon" />
          <span class="side-label">{{ item.label }}</span>
          <span v-if="item.badge" class="side-badge">{{ item.badge }}</span>
        </button>
      </aside>

      <main class="board-content">
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

        <SystemListView v-else-if="activePage === 'system-list'" />

        <ReleaseUnitPage v-else />
      </main>
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
  DatabaseOutlined,
} from '@ant-design/icons-vue'
import Header from './Header.vue'
import PlanList from './PlanList.vue'
import PlanReleaseUnitsPanel from './PlanReleaseUnitsPanel.vue'
import PipelineAllListPanel from './PipelineAllListPanel.vue'
import PipelineBoard from './PipelineBoard.vue'
import PipelineConfigDrawer from './PipelineConfigDrawer.vue'
import ReleaseUnitPage from './ReleaseUnitPage.vue'
import SystemListView from '../views/SystemListView.vue'
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

const menuItems = [
  { key: 'workspace', label: '流水线', icon: DeploymentUnitOutlined, title: '流水线' },
  { key: 'system-list', label: '系统管理', icon: DatabaseOutlined, title: '系统管理' },
  { key: 'iteration-plans', label: '迭代计划', icon: ProfileOutlined, title: '迭代计划' },
  { key: 'release-units', label: '发布单元', icon: ApartmentOutlined, title: '发布单元' },
  { key: 'global-vars', label: '全局变量', icon: AppstoreOutlined, title: '全局变量' },
  { key: 'build', label: '构建', icon: BuildOutlined, title: '敬请期待', disabled: true },
  { key: 'settings', label: '设置', icon: SettingOutlined, title: '敬请期待', disabled: true },
]

const activePage = computed(() => {
  if (route.name === 'pipeline-all') return 'pipeline-all'
  if (route.name === 'system-list') return 'system-list'
  if (route.name === 'iteration-plans') return 'iteration-plans'
  if (route.name === 'global-vars') return 'global-vars'
  if (route.name === 'release-units') return 'release-units'
  return 'workspace'
})

const setActivePage = (page) => {
  if (page === 'system-list') {
    router.push('/systems')
    return
  }
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
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
  position: relative;
}

.dashboard-page::before {
  content: '';
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background:
    radial-gradient(ellipse at 20% 0%, rgba(0, 212, 255, 0.08) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 100%, rgba(124, 58, 237, 0.06) 0%, transparent 50%);
  pointer-events: none;
  z-index: 0;
}

.board-main {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 220px 1fr;
  min-height: calc(100vh - 64px);
  flex: 1;
}

.sidebar {
  position: relative;
  background: rgba(17, 24, 39, 0.95);
  backdrop-filter: blur(16px);
  padding: 18px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
  overflow-y: auto;
}

.sidebar-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 120px;
  background: linear-gradient(180deg, rgba(0, 212, 255, 0.12) 0%, transparent 100%);
  pointer-events: none;
}

.side-item {
  position: relative;
  width: 100%;
  height: auto;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
  padding: 12px 14px;
  font-size: 14px;
  font-weight: 500;
  transition: all var(--transition-base);
  font-family: var(--font-display);
  letter-spacing: 0.01em;
}

.side-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 0;
  background: var(--accent-primary);
  border-radius: 0 2px 2px 0;
  transition: height var(--transition-base);
}

.side-item:hover:not(:disabled) {
  background: rgba(0, 212, 255, 0.1);
  color: var(--text-primary);
  border-color: rgba(0, 212, 255, 0.25);
  transform: translateX(4px);
}

.side-item.active {
  background: rgba(0, 212, 255, 0.14);
  color: var(--accent-primary);
  border-color: rgba(0, 212, 255, 0.35);
  box-shadow: inset 0 0 24px rgba(0, 212, 255, 0.12);
}

.side-item.active::before {
  height: 24px;
}

.side-item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.side-item svg {
  font-size: 16px;
  flex-shrink: 0;
  min-width: 16px;
  transition: transform var(--transition-fast);
}

.side-item:hover:not(:disabled) svg {
  transform: scale(1.1);
}

.side-label {
  font-size: 14px;
  line-height: 1.3;
  flex: 1;
  text-align: left;
}

.side-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  background: rgba(0, 212, 255, 0.22);
  color: var(--accent-primary);
  font-family: var(--font-mono);
}

.side-item.active .side-badge {
  background: rgba(0, 212, 255, 0.32);
}

.board-content {
  background: var(--core-bg);
  padding: 24px;
  overflow: auto;
  color: var(--core-text);
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.board-content > * {
  flex: 1;
  min-height: 0;
}

.plans-panel {
  max-width: 320px;
}

.iteration-layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 24px;
  min-width: 0;
  min-height: calc(100vh - 112px);
  align-items: stretch;
  animation: slideUp 0.5s ease-out;
}

.iteration-left {
  min-width: 0;
  display: flex;
}

.iteration-right {
  min-width: 0;
  animation: slideUp 0.5s ease-out 0.1s both;
}

.pipeline-layout {
  display: flex;
  flex-direction: column;
  min-width: 0;
  animation: fadeIn 0.4s ease-out;
}

@media (max-width: 1400px) {
  .board-main {
    grid-template-columns: 1fr;
  }

  .sidebar {
    flex-direction: row;
    justify-content: flex-start;
    padding: 10px;
    gap: 8px;
    overflow-x: auto;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
  }

  .sidebar-glow {
    display: none;
  }

  .side-item {
    width: auto;
    min-width: 90px;
    flex-shrink: 0;
  }

  .side-item::before {
    display: none;
  }

  .side-item.active {
    border-color: rgba(0, 212, 255, 0.45);
  }
}

@media (max-width: 768px) {
  .board-content {
    padding: 16px;
  }

  .iteration-layout {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}
</style>
