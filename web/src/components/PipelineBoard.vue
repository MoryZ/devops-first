<template>
  <div class="pipeline-board">
    <div class="board-toolbar">
      <div class="toolbar-title">流水线</div>
      <div class="toolbar-actions">
        <a-button @click="loadPipelines"><ReloadOutlined /> 刷新</a-button>
        <a-button type="primary" @click="showCreateModal"><PlusOutlined /> 新建流水线</a-button>
      </div>
    </div>

    <div class="pipeline-list">
      <a-empty v-if="pipelines.length === 0" description="暂无流水线" />

      <article v-for="pipeline in pipelines" :key="pipeline.id" class="pipeline-card">
        <div class="pipeline-head">
          <div class="pipeline-name">
            <a-tag :color="appTypeColor[pipeline.app_type]">{{ pipeline.app_type }}</a-tag>
            {{ pipeline.name }}
          </div>
          <div class="pipeline-desc">
            {{ pipeline.description }}
            <a-tag v-if="pipeline.release_unit_name" color="geekblue">发布单元: {{ pipeline.release_unit_name }}</a-tag>
          </div>
          <div class="pipeline-head-actions">
            <a-button type="link" size="small" @click="openPipelineConfig(pipeline)">配置</a-button>
            <a-button type="link" size="small" @click="openBPMDesigner(pipeline)">流程编排</a-button>
            <a-button type="link" size="small" @click="openExecutionHistory(pipeline)">执行历史</a-button>
          </div>
        </div>
      </article>
    </div>

    <a-modal
      v-model:open="createModalOpen"
      title="新建流水线"
      :confirm-loading="creating"
      @ok="handleCreatePipeline"
    >
      <a-form layout="vertical">
        <a-form-item label="流水线名称">
          <a-input v-model:value="newPipeline.name" placeholder="输入流水线名称" />
        </a-form-item>
        <a-form-item label="应用类型">
          <a-select v-model:value="newPipeline.app_type" placeholder="选择应用类型">
            <a-select-option value="java">Java</a-select-option>
            <a-select-option value="node">Node</a-select-option>
            <a-select-option value="sql">SQL</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="发布单元（可选）">
          <a-select
            v-model:value="newPipeline.release_unit_id"
            placeholder="选择发布单元"
            allow-clear
            @change="handleCreateUnitChange"
          >
            <a-select-option v-for="unit in releaseUnits" :key="unit.id" :value="unit.id">
              {{ unit.displayOrder }} - {{ unit.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea v-model:value="newPipeline.description" placeholder="流水线描述（可选）" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import { useWorkspaceStore } from '../stores/workspace'

const props = defineProps({
  token: String,
})

const emit = defineEmits(['edit-pipeline'])
const workspace = useWorkspaceStore()
const router = useRouter()

const pipelines = ref([])
const pipelineConfigMap = ref({})
const releaseUnits = ref([])
const createModalOpen = ref(false)
const creating = ref(false)
const newPipeline = ref({
  name: '',
  app_type: 'java',
  release_unit_id: '',
  description: '',
})

const appTypeColor = {
  java: 'blue',
  node: 'green',
  sql: 'orange',
}

const mapPipeline = (item) => ({
  id: item.ID || item.id,
  system_id: item.SystemID || item.system_id,
  plan_id: item.PlanID || item.plan_id,
  name: item.Name || item.name,
  app_type: item.AppType || item.app_type,
  description: item.Description || item.description,
  created_at: item.CreatedAt || item.created_at,
  updated_at: item.UpdatedAt || item.updated_at,
})

const normalizePipelineConfig = (item) => ({
  pipeline_id: item.pipeline_id,
  name: item.name,
  release_unit_id: item.release_unit_id,
  repository_type: item.repository_type || 'git',
  auto_merge: item.auto_merge ?? true,
  auto_tag: item.auto_tag ?? true,
  display_order: Number(item.display_order) || 0,
  repo_url: item.repo_url,
  branch: item.branch,
  git_username: item.git_username,
  git_token: item.git_token,
  project_path: item.project_path,
  build_type: item.build_type,
  maven_command: item.maven_command,
  gradle_command: item.gradle_command,
  npm_command: item.npm_command,
  deploy_type: item.deploy_type,
  docker_image: item.docker_image,
  docker_container: item.docker_container,
  docker_run_args: item.docker_run_args,
  main_stages: item.main_stages || [],
  env_stages: item.env_stages || [],
})

const loadReleaseUnits = () => {
  if (!workspace.selectedSystemId) {
    releaseUnits.value = []
    return
  }
  try {
    const raw = localStorage.getItem(`releaseUnits:${workspace.selectedSystemId}`)
    const parsed = raw ? JSON.parse(raw) : []
    releaseUnits.value = Array.isArray(parsed)
      ? parsed
          .map((item, index) => ({
            ...item,
            repositoryType: item.repositoryType || 'git',
            autoMerge: item.autoMerge ?? true,
            autoTag: item.autoTag ?? true,
            displayOrder: Number(item.displayOrder) || index + 1,
          }))
          .sort((a, b) => Number(a.displayOrder) - Number(b.displayOrder))
      : []
  } catch {
    releaseUnits.value = []
  }
}

const mergePipelinesWithConfigs = () => {
  pipelines.value = pipelines.value.map((p) => {
    const cfg = pipelineConfigMap.value[p.id]
    if (!cfg) return p
    const unit = releaseUnits.value.find((u) => u.id === cfg.release_unit_id)
    return {
      ...p,
      release_unit_name: unit?.name,
      config: {
        releaseUnitId: cfg.release_unit_id,
        repositoryType: cfg.repository_type,
        autoMerge: cfg.auto_merge,
        autoTag: cfg.auto_tag,
        displayOrder: cfg.display_order,
        repoUrl: cfg.repo_url,
        branch: cfg.branch,
        gitUsername: cfg.git_username,
        gitToken: cfg.git_token,
        projectPath: cfg.project_path,
        buildType: cfg.build_type,
        mavenCommand: cfg.maven_command,
        gradleCommand: cfg.gradle_command,
        npmCommand: cfg.npm_command,
        deployType: cfg.deploy_type,
        dockerImage: cfg.docker_image,
        dockerContainer: cfg.docker_container,
        dockerRunArgs: cfg.docker_run_args,
      },
      mainStages: cfg.main_stages,
      envStages: cfg.env_stages,
    }
  })
}

const loadPipelineConfigs = async () => {
  try {
    const res = await fetch('/api/pipelines', {
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (!res.ok) return
    const data = await res.json()
    pipelineConfigMap.value = (data.items || []).reduce((acc, item) => {
      const normalized = normalizePipelineConfig(item)
      acc[normalized.pipeline_id] = normalized
      return acc
    }, {})
  } catch (err) {
    console.warn('Failed to load pipeline configs:', err)
  }
}

watch(
  () => [workspace.selectedPlanId, workspace.selectedSystemId],
  async ([planId, systemId]) => {
    loadReleaseUnits()
    await loadPipelineConfigs()
    if (planId) {
      await loadPipelinesForPlan(planId)
    } else if (systemId) {
      await loadPipelinesForSystem(systemId)
    } else {
      pipelines.value = []
    }
  }
)

onMounted(async () => {
  loadReleaseUnits()
  await loadPipelineConfigs()
  if (workspace.selectedPlanId) {
    await loadPipelinesForPlan(workspace.selectedPlanId)
  } else if (workspace.selectedSystemId) {
    await loadPipelinesForSystem(workspace.selectedSystemId)
  }
})

const loadPipelinesForPlan = async (planId) => {
  try {
    const res = await fetch(`/api/plans/${planId}/pipelines`, {
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (res.ok) {
      const data = await res.json()
      pipelines.value = (data.items || []).map(mapPipeline)
      mergePipelinesWithConfigs()
    }
  } catch (err) {
    console.warn('Failed to load pipelines for plan:', err)
  }
}

const loadPipelinesForSystem = async (systemId) => {
  try {
    const res = await fetch(`/api/systems/${systemId}/pipelines`, {
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (res.ok) {
      const data = await res.json()
      pipelines.value = (data.items || []).map(mapPipeline)
      mergePipelinesWithConfigs()
    }
  } catch (err) {
    console.warn('Failed to load pipelines for system:', err)
  }
}

const loadPipelines = async () => {
  if (workspace.selectedPlanId) {
    await loadPipelinesForPlan(workspace.selectedPlanId)
  } else if (workspace.selectedSystemId) {
    await loadPipelinesForSystem(workspace.selectedSystemId)
  }
}

const showCreateModal = () => {
  if (!workspace.selectedSystemId) {
    message.warning('请先选择系统')
    return
  }
  createModalOpen.value = true
  newPipeline.value = {
    name: '',
    app_type: 'java',
    release_unit_id: '',
    description: '',
  }
}

const handleCreateUnitChange = (unitId) => {
  if (!unitId) return
  const unit = releaseUnits.value.find((u) => u.id === unitId)
  if (!unit) return
  if (!newPipeline.value.description?.trim()) {
    newPipeline.value.description = `${unit.name} 发布流水线`
  }
}

const handleCreatePipeline = async () => {
  if (!newPipeline.value.name.trim()) {
    message.warning('请输入流水线名称')
    return
  }
  if (!workspace.selectedSystemId) {
    message.warning('请先选择系统')
    return
  }

  creating.value = true
  try {
    const res = await fetch(`/api/systems/${workspace.selectedSystemId}/pipelines`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${props.token}`,
      },
      body: JSON.stringify({
        name: newPipeline.value.name,
        app_type: newPipeline.value.app_type,
        description: newPipeline.value.description,
        plan_id: workspace.selectedPlanId || '',
      }),
    })

    if (!res.ok) {
      const err = await res.json()
      message.error(err.error || '创建流水线失败')
      return
    }

    const created = await res.json()
    if (newPipeline.value.release_unit_id) {
      const unit = releaseUnits.value.find((u) => u.id === newPipeline.value.release_unit_id)
      if (unit) {
        await fetch('/api/pipelines/config', {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${props.token}`,
          },
          body: JSON.stringify({
            pipeline_id: created.ID || created.id,
            name: created.Name || created.name,
            release_unit_id: unit.id,
            repository_type: unit.repositoryType || 'git',
            auto_merge: unit.autoMerge ?? true,
            auto_tag: unit.autoTag ?? true,
            display_order: Number(unit.displayOrder) || 0,
            repo_url: unit.repoUrl || '',
            branch: unit.branch || 'main',
            project_path: '',
            maven_command: 'mvn clean package -DskipTests',
            docker_image: '',
            docker_container: '',
            docker_run_args: '',
            main_stages: [],
            env_stages: [],
          }),
        })
      }
    }

    message.success('流水线创建成功')
    createModalOpen.value = false
    await loadPipelineConfigs()
    await loadPipelines()
  } catch (err) {
    message.error('创建流水线失败: ' + err.message)
  } finally {
    creating.value = false
  }
}

const openPipelineConfig = (pipeline) => {
  emit('edit-pipeline', pipeline)
}

const openExecutionHistory = (pipeline) => {
  router.push({
    path: `/pipelines/${pipeline.id}/executions`,
    query: { name: pipeline.name || '' },
  })
}

const openBPMDesigner = (pipeline) => {
  router.push({
    path: `/pipelines/${pipeline.id}/bpm`,
    query: { name: pipeline.name || '' },
  })
}
</script>

<style scoped>
.pipeline-board {
  flex: 1;
  padding: 18px;
  overflow-y: auto;
}

.board-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.toolbar-title {
  font-size: 22px;
  font-weight: 700;
  color: #1c2d42;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.pipeline-list {
  display: grid;
  gap: 14px;
}

.pipeline-card {
  background: #fff;
  border: 1px solid #dfe7f4;
  border-radius: 12px;
  padding: 14px;
}

.pipeline-head {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
}

.pipeline-name {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #2a5db8;
  font-weight: 700;
}

.pipeline-desc {
  color: #6d7b95;
  font-size: 13px;
}

.pipeline-head-actions {
  display: flex;
  gap: 4px;
}

.stage-row {
  display: flex;
  gap: 6px;
  margin-bottom: 8px;
  overflow-x: auto;
}

.stage-col {
  min-width: 165px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.stage-card {
  width: 100%;
  border: 1px solid #e3e7f1;
  border-radius: 8px;
  background: #fafcff;
  padding: 8px;
}

.stage-top {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #2e3f57;
  font-size: 13px;
}

.stage-duration {
  margin-left: auto;
  color: #7d8ea8;
}

.stage-dot {
  width: 8px;
  height: 8px;
  border-radius: 4px;
}

.dot-success {
  background: #2fb65c;
}

.dot-failed {
  background: #e24b4b;
}

.dot-pending {
  background: #b8c1d3;
}

.stage-progress {
  margin-top: 8px;
}

.progress-line {
  display: block;
  width: 100%;
  height: 4px;
  border-radius: 2px;
}

.line-success {
  background: #2fb65c;
}

.line-failed {
  background: #e24b4b;
}

.line-pending {
  background: #cfd5df;
}

.stage-link {
  width: 14px;
  height: 1px;
  background: #9cafcd;
}
</style>
