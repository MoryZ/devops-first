<template>
  <div class="pipeline-board">
    <div class="board-toolbar">
      <div class="toolbar-title-wrap">
        <h3 class="toolbar-title">
          <span class="title-icon">▶</span>
          流水线
        </h3>
        <div class="title-underline"></div>
      </div>

      <div class="toolbar-actions">
        <div class="search-shell" :class="{ active: searchFocus || keyword }" @click="focusSearch">
          <SearchOutlined class="search-icon" />
          <input
            ref="searchInputRef"
            v-model="keyword"
            class="search-input"
            placeholder="搜索流水线名..."
            @focus="searchFocus = true"
            @blur="handleSearchBlur"
          />
        </div>

        <a-select
          v-model:value="selectedPlanIdProxy"
          class="plan-select"
          placeholder="选择迭代版本"
          allow-clear
        >
          <a-select-option v-for="plan in plans" :key="plan.id" :value="plan.id">
            {{ plan.version }}
          </a-select-option>
        </a-select>

        <a-dropdown>
          <a-button class="icon-btn">
            <SortAscendingOutlined />
          </a-button>
          <template #overlay>
            <a-menu @click="handleSortChange">
              <a-menu-item key="exec_desc">按执行时间先后排序</a-menu-item>
              <a-menu-item key="name_asc">按流水线名称排序(A-Z)</a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
        <a-button class="icon-btn" @click="filterOpen = true">
          <FilterOutlined />
        </a-button>
        <a-button class="secondary-btn" @click="goAllPipelines">
          <UnorderedListOutlined />
          查看全部
        </a-button>
        <a-button type="primary" class="primary-btn" @click="showCreateModal">
          <PlusOutlined />
          新建
        </a-button>
      </div>
    </div>

    <div class="version-line" v-if="selectedPlanVersion !== '全部版本'">
      <span class="version-label">当前版本</span>
      <span class="version-value">{{ selectedPlanVersion }}</span>
    </div>

    <div class="pipeline-list">
      <a-empty v-if="displayPipelines.length === 0" description="暂无流水线，请先创建或绑定发布单元" />

      <article v-for="(pipeline, index) in displayPipelines" :key="pipeline.id" class="pipeline-card" :style="{ animationDelay: `${index * 0.05}s` }">
        <div class="pipeline-one-line">
          <span class="line-name" @click="openExecutionHistory(pipeline)">{{ getPipelineDisplayName(pipeline) }}</span>
          <span class="line-tags" v-if="(pipeline.tags || []).length > 0">
            <a-tag v-for="tag in pipeline.tags || []" :key="`${pipeline.id}-${tag}`" color="blue" class="custom-tag">{{ tag }}</a-tag>
          </span>
          <template v-if="pipeline.last_execution_at">
            <span class="line-item"><span class="line-label">执行人员</span><span class="line-value">{{ pipeline.last_execution_user || '-' }}</span></span>
            <span class="line-item"><span class="line-label">执行时间</span><span class="line-value">{{ formatDateTime(pipeline.last_execution_at) }}</span></span>
            <span class="line-item"><span class="line-label">分支</span><span class="line-value">{{ pipeline.config?.branch || '-' }}</span></span>
            <span class="line-item"><span class="line-label">最新提交ID</span>
              <a-tooltip v-if="pipeline.last_commit_id" :title="pipeline.last_commit_id">
                <span class="line-value commit-id">{{ pipeline.last_commit_id.slice(0, 9) }}</span>
              </a-tooltip>
              <span v-else class="line-value commit-id">-</span>
            </span>
          </template>

          <span class="line-actions">
            <a-button size="small" class="action-btn" @click="handleRerunPipeline(pipeline)">{{ pipeline.last_execution_at ? '重新执行' : '执行' }}</a-button>
            <a-button size="small" type="primary" ghost class="action-btn-primary" @click="openPipelineConfig(pipeline)">配置</a-button>
          </span>
        </div>

        <div class="stage-lane-wrap">
          <div class="stage-row">
            <template v-for="(stage, idx) in getMainStages(pipeline)" :key="`${pipeline.id}-main-${idx}`">
              <div class="stage-col">
                <div class="stage-card clickable" @click="openStageDetail(pipeline, stage)">
                  <div class="stage-top">
                    <span class="stage-dot" :class="statusDotClass(stage.status)"></span>
                    <a-tooltip :title="getStageTooltipText(stage, pipeline)">
                      <span class="stage-name">{{ stage.name }}</span>
                    </a-tooltip>
                    <span class="stage-duration">{{ stage.duration || '-' }}</span>
                  </div>
                  <div class="stage-progress">
                    <span class="progress-line" :class="statusLineClass(stage.status)"></span>
                  </div>
                </div>
              </div>
              <span v-if="idx < getMainStages(pipeline).length - 1" class="stage-link"></span>
            </template>
            <template v-if="getEnvStages(pipeline).length > 0">
              <div class="parallel-group-wrap">
                <span class="stage-link parallel-entry-link"></span>
                <div class="env-parallel-block">
                  <div v-for="(stage, idx) in getEnvStages(pipeline)" :key="`${pipeline.id}-env-${idx}`" class="stage-col">
                    <div class="stage-card env-card clickable" @click="openStageDetail(pipeline, stage)">
                      <div class="stage-top">
                        <span class="stage-dot" :class="statusDotClass(stage.status)"></span>
                        <a-tooltip :title="getStageTooltipText(stage, pipeline)">
                          <span class="stage-name">{{ stage.name }}</span>
                        </a-tooltip>
                        <span class="stage-duration">{{ stage.duration || '-' }}</span>
                      </div>
                      <div class="stage-progress">
                        <span class="progress-line" :class="statusLineClass(stage.status)"></span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </template>
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

    <a-modal v-model:open="filterOpen" title="标签筛选" @ok="applyTagFilter">
      <a-form layout="vertical">
        <a-form-item label="按标签过滤流水线">
          <a-select
            v-model:value="pendingFilterTags"
            mode="multiple"
            allow-clear
            placeholder="选择标签"
          >
            <a-select-option v-for="tag in allTags" :key="tag" :value="tag">{{ tag }}</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="addUnitOpen" title="添加发布单元" @ok="handleAddUnitToPlan">
      <a-form layout="vertical">
        <a-form-item label="发布单元">
          <a-select v-model:value="pendingUnitId" placeholder="选择当前子系统发布单元">
            <a-select-option v-for="unit in releaseUnits" :key="unit.id" :value="unit.id">
              {{ unit.displayOrder }} - {{ unit.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="流水线名称">
          <a-input v-model:value="pendingUnitPipelineName" placeholder="默认使用发布单元名称" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="historyOpen" title="部署历史" width="72vw" :footer="null">
      <a-table
        :data-source="historyRows"
        :columns="historyColumns"
        row-key="id"
        :pagination="{ pageSize: 6, showSizeChanger: false }"
        :loading="historyLoading"
      />
    </a-modal>
  </div>
</template>

<script setup>
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, SearchOutlined, SortAscendingOutlined, UnorderedListOutlined, FilterOutlined } from '@ant-design/icons-vue'
import { useWorkspaceStore } from '../stores/workspace'
import { listPlansBySystem, listPipelinesByPlan } from '../api/plans'
import { createSystemPipeline, listSystemPipelines } from '../api/systems'
import { listPipelineConfigs, listPipelineExecutions, upsertPipelineConfig } from '../api/pipelines'
import { startPipelineExecution, getBatchStatus } from '../api/executions'

const props = defineProps({
  token: String,
})

const emit = defineEmits(['edit-pipeline'])
const workspace = useWorkspaceStore()
const router = useRouter()

const pipelines = ref([])
const plans = ref([])
const keyword = ref('')
const searchFocus = ref(false)
const searchInputRef = ref(null)
const sortBy = ref('exec_desc')
const filterOpen = ref(false)
const activeFilterTags = ref([])
const pendingFilterTags = ref([])
const selectedPipelineIds = ref([])
const addUnitOpen = ref(false)
const pendingUnitId = ref('')
const pendingUnitPipelineName = ref('')
const historyOpen = ref(false)
const historyRows = ref([])
const historyLoading = ref(false)
const pipelineConfigMap = ref({})
const releaseUnits = ref([])
const runningBatchPollers = ref({})
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

const selectedPlanIdProxy = computed({
  get: () => workspace.selectedPlanId,
  set: (val) => {
    selectedPipelineIds.value = []
    workspace.selectPlan(val || '')
  },
})

const historyColumns = [
  { title: '用户', dataIndex: 'user', key: 'user', width: 120 },
  { title: '部署环境名', dataIndex: 'environment', key: 'environment', width: 180 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 120 },
  { title: '时间', dataIndex: 'time', key: 'time', width: 220 },
]

const filteredPipelines = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return pipelines.value
  return pipelines.value.filter((p) => `${p.name || ''} ${p.description || ''}`.toLowerCase().includes(kw))
})

const allTags = computed(() => {
  const bucket = new Set()
  pipelines.value.forEach((pipeline) => {
    ;(pipeline.tags || []).forEach((tag) => bucket.add(tag))
  })
  return Array.from(bucket).sort((a, b) => a.localeCompare(b))
})

const displayPipelines = computed(() => {
  let list = [...filteredPipelines.value]
  if (activeFilterTags.value.length > 0) {
    list = list.filter((pipeline) => {
      const tagSet = new Set(pipeline.tags || [])
      return activeFilterTags.value.every((tag) => tagSet.has(tag))
    })
  }

  if (sortBy.value === 'name_asc') {
    list.sort((a, b) => String(a.name || '').localeCompare(String(b.name || '')))
  } else {
    list.sort((a, b) => Number(b.last_execution_ts || 0) - Number(a.last_execution_ts || 0))
  }
  return list
})

const selectedPlanVersion = computed(() => {
  if (!workspace.selectedPlanId) return '全部版本'
  const current = plans.value.find((plan) => String(plan.id) === String(workspace.selectedPlanId))
  return current?.version || '未匹配版本'
})

const defaultMainStages = [
  { name: '触发源', status: 'pending', duration: '' },
  { name: '编译', status: 'pending', duration: '' },
  { name: '代码扫描', status: 'pending', duration: '' },
  { name: '质量门禁', status: 'pending', duration: '' },
]

const defaultEnvStages = [
  { name: '灰度环境', status: 'pending', duration: '' },
  { name: '生产环境', status: 'pending', duration: '' },
]

const normalizeStage = (stage, index, fallbackName) => ({
  name: stage?.name || fallbackName || `阶段 ${index + 1}`,
  status: stage?.status || 'pending',
  duration: stage?.duration || '',
})

const resolveStageStatus = (baseStatus, liveStages, stageName) => {
  if (liveStages && liveStages[stageName] !== undefined) return liveStages[stageName]
  if (baseStatus === 'pending') return 'idle'
  return baseStatus || 'idle'
}

const getMainStages = (pipeline) => {
  const liveStages = pipeline.live_stages
  if (Array.isArray(pipeline.mainStages) && pipeline.mainStages.length > 0) {
    return pipeline.mainStages.map((stage, index) => {
      const base = normalizeStage(stage, index, defaultMainStages[index]?.name)
      base.status = resolveStageStatus(base.status, liveStages, base.name)
      return base
    })
  }
  return defaultMainStages.map((s) => ({ ...s, status: resolveStageStatus(s.status, liveStages, s.name) }))
}

const getEnvStages = (pipeline) => {
  const liveStages = pipeline.live_stages
  if (Array.isArray(pipeline.envStages) && pipeline.envStages.length > 0) {
    return pipeline.envStages.map((stage, index) => {
      const base = normalizeStage(stage, index, defaultEnvStages[index]?.name)
      base.status = resolveStageStatus(base.status, liveStages, base.name)
      return base
    })
  }
  return defaultEnvStages.map((s) => ({ ...s, status: resolveStageStatus(s.status, liveStages, s.name) }))
}

const statusDotClass = (status) => {
  if (status === 'success' || status === 'done') return 'dot-success'
  if (status === 'failed' || status === 'error') return 'dot-failed'
  if (status === 'running') return 'dot-running'
  if (status === 'idle') return 'dot-idle'
  return 'dot-pending'
}

const statusLineClass = (status) => {
  if (status === 'success' || status === 'done') return 'line-success'
  if (status === 'failed' || status === 'error') return 'line-failed'
  if (status === 'running') return 'line-running'
  if (status === 'idle') return 'line-idle'
  return 'line-pending'
}

const stageDisplayResult = (status) => {
  if (status === 'success' || status === 'done') return '运行成功'
  if (status === 'failed' || status === 'error') return '运行失败'
  if (status === 'running') return '运行中'
  if (status === 'pending') return '等待执行'
  return '未执行'
}

const stageActionText = (stage, pipeline) => {
  const name = String(stage?.name || '')
  if (name.includes('触发源') || name.includes('检出') || name.includes('获取代码')) return '获取代码'
  if (name.includes('编译')) return '普通编译'
  return name || String(pipeline?.name || '节点任务')
}

const getStageTooltipText = (stage, pipeline) => {
  return `${stageActionText(stage, pipeline)}\n${stageDisplayResult(stage?.status)}`
}

const stageNameToKey = (name) => {
  const n = String(name || '').toLowerCase()
  if (n.includes('触发源') || n.includes('检出') || n.includes('获取代码') || n.includes('source') || n.includes('checkout')) return 'source'
  if (n.includes('编译') || n.includes('build') || n.includes('maven') || n.includes('gradle')) return 'build'
  if (n.includes('部署') || n.includes('灰度') || n.includes('生产') || n.includes('deploy')) return 'deploy'
  if (n.includes('扫描') || n.includes('门禁') || n.includes('test') || n.includes('task')) return 'task'
  return 'task'
}

const openStageDetail = (pipeline, stage) => {
  const batchId = pipeline?.last_batch_id
  if (!batchId) {
    message.warning('暂无可查看的执行批次')
    return
  }
  router.push({
    path: `/pipelines/${pipeline.id}/executions/${batchId}/stage/${encodeURIComponent(stageNameToKey(stage?.name))}`,
    query: {
      name: pipeline?.name || '',
      stage_name: stage?.name || '',
      operator: pipeline?.last_execution_user || '',
      plan_version: selectedPlanVersion.value || '',
      batch_number: pipeline?.last_batch_number || '',
    },
  })
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
  selectedPipelineIds.value = selectedPipelineIds.value.filter((id) => pipelines.value.some((p) => p.id === id))
}

const getTagStorageKey = () => `pipelineTags:${workspace.selectedSystemId || 'default'}`

const loadTagMap = () => {
  try {
    const raw = localStorage.getItem(getTagStorageKey())
    const parsed = raw ? JSON.parse(raw) : {}
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}

const saveTagMap = (tagMap) => {
  localStorage.setItem(getTagStorageKey(), JSON.stringify(tagMap))
}

const bindTagsToPipelines = () => {
  const tagMap = loadTagMap()
  pipelines.value = pipelines.value.map((pipeline) => ({
    ...pipeline,
    tags: Array.isArray(tagMap[pipeline.id]) ? tagMap[pipeline.id] : [],
  }))
}

const updatePipelineTags = (pipelineId, values) => {
  const tags = Array.from(new Set((values || []).map((v) => String(v).trim()).filter(Boolean)))
  const tagMap = loadTagMap()
  tagMap[pipelineId] = tags
  saveTagMap(tagMap)
  pipelines.value = pipelines.value.map((p) => (p.id === pipelineId ? { ...p, tags } : p))
}

const applyTagFilter = () => {
  activeFilterTags.value = [...pendingFilterTags.value]
  filterOpen.value = false
}

const togglePipelineSelection = (pipelineId, checked) => {
  if (checked) {
    if (!selectedPipelineIds.value.includes(pipelineId)) {
      selectedPipelineIds.value.push(pipelineId)
    }
  } else {
    selectedPipelineIds.value = selectedPipelineIds.value.filter((id) => id !== pipelineId)
  }
}

const buildConfigPayload = (pipeline, overrides = {}) => {
  const cfg = pipeline.config || {}
  return {
    pipeline_id: pipeline.id,
    name: pipeline.name,
    release_unit_id: overrides.release_unit_id ?? cfg.releaseUnitId ?? '',
    repository_type: cfg.repositoryType || 'git',
    auto_merge: cfg.autoMerge ?? true,
    auto_tag: cfg.autoTag ?? true,
    display_order: Number(cfg.displayOrder) || 0,
    repo_url: cfg.repoUrl || '',
    branch: cfg.branch || 'main',
    git_username: cfg.gitUsername || '',
    git_token: cfg.gitToken || '',
    project_path: cfg.projectPath || '',
    build_type: cfg.buildType || pipeline.app_type || 'maven',
    maven_command: cfg.mavenCommand || 'mvn clean package -DskipTests',
    gradle_command: cfg.gradleCommand || '',
    npm_command: cfg.npmCommand || '',
    deploy_type: cfg.deployType || 'docker',
    docker_image: cfg.dockerImage || '',
    docker_container: cfg.dockerContainer || '',
    docker_run_args: cfg.dockerRunArgs || '',
    main_stages: pipeline.mainStages || [],
    env_stages: pipeline.envStages || [],
  }
}

const updatePipelineReleaseUnit = async (pipeline, releaseUnitId) => {
  try {
    const payload = buildConfigPayload(pipeline, { release_unit_id: releaseUnitId || '' })
    await upsertPipelineConfig(props.token, payload)
    await loadPipelineConfigs()
    await loadPipelines()
    message.success('发布单元已更新')
  } catch (err) {
    message.error('发布单元更新失败: ' + err.message)
  }
}

const handleBatchRemove = async () => {
  if (selectedPipelineIds.value.length === 0) {
    message.warning('请先选择要移除的流水线')
    return
  }
  try {
    for (const pipelineId of selectedPipelineIds.value) {
      const pipeline = pipelines.value.find((item) => item.id === pipelineId)
      if (!pipeline) continue
      const payload = buildConfigPayload(pipeline, { release_unit_id: '' })
      await upsertPipelineConfig(props.token, payload)
    }
    selectedPipelineIds.value = []
    await loadPipelineConfigs()
    await loadPipelines()
    message.success('批量移除完成')
  } catch (err) {
    message.error('批量移除失败: ' + err.message)
  }
}

const showAddUnitModal = () => {
  if (!workspace.selectedPlanId) {
    message.warning('请先选择迭代计划')
    return
  }
  pendingUnitId.value = ''
  pendingUnitPipelineName.value = ''
  addUnitOpen.value = true
}

const handleAddUnitToPlan = async () => {
  if (!workspace.selectedSystemId || !workspace.selectedPlanId) {
    message.warning('缺少当前子系统或迭代计划')
    return
  }
  if (!pendingUnitId.value) {
    message.warning('请选择发布单元')
    return
  }

  const unit = releaseUnits.value.find((item) => item.id === pendingUnitId.value)
  if (!unit) {
    message.warning('发布单元不存在')
    return
  }

  const pipelineName = pendingUnitPipelineName.value.trim() || `${unit.name}-pipeline`

  try {
    const created = await createSystemPipeline(props.token, workspace.selectedSystemId, {
      plan_id: workspace.selectedPlanId,
      name: pipelineName,
      app_type: 'java',
      description: `${unit.name} 发布流水线`,
    })
    await upsertPipelineConfig(props.token, {
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
      build_type: 'maven',
      maven_command: 'mvn clean package -DskipTests',
      deploy_type: 'docker',
      docker_image: '',
      docker_container: '',
      docker_run_args: '',
      main_stages: [],
      env_stages: [],
    })

    addUnitOpen.value = false
    await loadPipelineConfigs()
    await loadPipelines()
    message.success('发布单元已添加到当前迭代')
  } catch (err) {
    message.error('添加发布单元失败: ' + err.message)
  }
}

const openHistoryModal = async () => {
  if (!workspace.selectedPlanId) {
    message.warning('请先选择迭代计划')
    return
  }

  historyOpen.value = true
  historyLoading.value = true
  try {
    const data = await listPipelinesByPlan(props.token, workspace.selectedPlanId)
    const items = data.items || []

    const history = []
    for (const pipeline of items) {
      const pid = pipeline.ID || pipeline.id
      const pname = pipeline.Name || pipeline.name
      const execData = await listPipelineExecutions(props.token, pid, 20)
      const rows = execData.batches || execData.items || []
      rows.forEach((row) => {
        history.push({
          id: row.ID || row.id || `${pid}-${row.BatchNumber || row.batch_number || Math.random()}`,
          user: row.UserID || row.user_id || (JSON.parse(localStorage.getItem('user') || '{}').username || '当前用户'),
          environment: pname,
          status: row.Status || row.status || '-',
          time: formatDateTime(row.CreatedAt || row.created_at || row.StartedAt || row.started_at),
          _ts: new Date(row.CreatedAt || row.created_at || row.StartedAt || row.started_at || 0).getTime(),
        })
      })
    }

    historyRows.value = history.sort((a, b) => b._ts - a._ts)
  } catch (err) {
    message.error('查询部署历史失败: ' + err.message)
  } finally {
    historyLoading.value = false
  }
}

const loadLastExecutionTimes = async () => {
  if (!props.token || pipelines.value.length === 0) return
  const currentUser = JSON.parse(localStorage.getItem('user') || '{}')?.username || ''

  const normalizeLatestStagesMap = (raw) => {
    if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return {}
    const normalized = { ...raw }

    if (normalized['触发源'] === undefined) {
      if (normalized['代码检出'] !== undefined) normalized['触发源'] = normalized['代码检出']
      if (normalized['source'] !== undefined) normalized['触发源'] = normalized['source']
      if (normalized['checkout'] !== undefined) normalized['触发源'] = normalized['checkout']
    }
    if (normalized['编译'] === undefined) {
      if (normalized['编译构建'] !== undefined) normalized['编译'] = normalized['编译构建']
      if (normalized['build'] !== undefined) normalized['编译'] = normalized['build']
    }

    return normalized
  }

  const results = await Promise.all(
    pipelines.value.map(async (pipeline) => {
      try {
        const data = await listPipelineExecutions(props.token, pipeline.id)
        const latest = (data.items || data.batches || [])[0]
        const at = latest?.CreatedAt || latest?.created_at || latest?.StartedAt || latest?.started_at || ''
        const user =
          latest?.UserName ||
          latest?.user_name ||
          latest?.Operator ||
          latest?.operator ||
          latest?.TriggeredBy ||
          latest?.triggered_by ||
          currentUser ||
          '-'
        const commitId =
          latest?.CommitID ||
          latest?.commit_id ||
          latest?.GitCommitID ||
          latest?.git_commit_id ||
          latest?.Revision ||
          latest?.revision ||
          ''
        const batchId = latest?.ID || latest?.id || ''
        const batchNumber = latest?.BatchNumber || latest?.batch_number || ''

        let liveStages = {}
        try {
          const rawStages = latest?.StagesStatusJSON || latest?.stages_status_json || '{}'
          const parsed = JSON.parse(rawStages)
          liveStages = normalizeLatestStagesMap(parsed)
        } catch {
          liveStages = {}
        }

        return { id: pipeline.id, ts: at ? new Date(at).getTime() : 0, at, user, commitId, batchId, batchNumber, liveStages }
      } catch {
        return { id: pipeline.id, ts: 0, at: '', user: currentUser || '-', commitId: '', batchId: '', batchNumber: '', liveStages: {} }
      }
    })
  )

  const map = results.reduce((acc, item) => {
    acc[item.id] = item
    return acc
  }, {})

  pipelines.value = pipelines.value.map((pipeline) => {
    const info = map[pipeline.id] || { ts: 0, at: '', user: '-', commitId: '', batchId: '', batchNumber: '', liveStages: {} }
    return {
      ...pipeline,
      last_execution_ts: info.ts,
      last_execution_at: info.at,
      last_execution_user: info.user,
      last_commit_id: info.commitId,
      last_batch_id: info.batchId,
      last_batch_number: info.batchNumber,
      live_stages: info.liveStages,
    }
  })
}

const finalizePipelineList = async () => {
  mergePipelinesWithConfigs()
  bindTagsToPipelines()
  await loadLastExecutionTimes()
}

const loadPipelineConfigs = async () => {
  try {
    const data = await listPipelineConfigs(props.token)
    pipelineConfigMap.value = (data.items || []).reduce((acc, item) => {
      const normalized = normalizePipelineConfig(item)
      acc[normalized.pipeline_id] = normalized
      return acc
    }, {})
  } catch (err) {
    console.warn('Failed to load pipeline configs:', err)
  }
}

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
    }))
    if (!plans.value.some((p) => p.id === workspace.selectedPlanId)) {
      workspace.selectPlan(plans.value[0]?.id || '')
    }
  } catch (err) {
    console.warn('Failed to load plans:', err)
  }
}

watch(
  () => workspace.selectedSystemId,
  async () => {
    activeFilterTags.value = []
    pendingFilterTags.value = []
    await loadPlans()
  },
  { immediate: true }
)

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
  await loadPlans()
  loadReleaseUnits()
  await loadPipelineConfigs()
  if (workspace.selectedPlanId) {
    await loadPipelinesForPlan(workspace.selectedPlanId)
  } else if (workspace.selectedSystemId) {
    await loadPipelinesForSystem(workspace.selectedSystemId)
  }

  onUnmounted(() => {
    Object.keys(runningBatchPollers.value).forEach((id) => stopBatchPoller(id))
  })
})

const loadPipelinesForPlan = async (planId) => {
  try {
    const data = await listPipelinesByPlan(props.token, planId)
    pipelines.value = (data.items || []).map(mapPipeline)
    await finalizePipelineList()
  } catch (err) {
    console.warn('Failed to load pipelines for plan:', err)
  }
}

const loadPipelinesForSystem = async (systemId) => {
  try {
    const data = await listSystemPipelines(props.token, systemId)
    pipelines.value = (data.items || []).map(mapPipeline)
    await finalizePipelineList()
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

const goAllPipelines = () => {
  router.push('/pipelines/all')
}

const reloadPlans = async () => {
  await loadPlans()
  await loadPipelines()
}

const formatDateTime = (value) => {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '-'
  return d.toLocaleString()
}

const focusSearch = () => {
  searchFocus.value = true
  if (searchInputRef.value) {
    searchInputRef.value.focus()
  }
}

const handleSearchBlur = () => {
  if (!keyword.value.trim()) {
    searchFocus.value = false
  }
}

const handleSortChange = ({ key }) => {
  if (key === 'exec_desc' || key === 'name_asc') {
    sortBy.value = key
  }
}

const getPipelineDisplayName = (pipeline) => {
  const raw = String(pipeline?.name || '').trim()
  if (raw.toLowerCase().endsWith('-pipeline')) {
    return raw.slice(0, -9)
  }
  return raw || '-'
}

const getPipelineDisplayDesc = (pipeline) => {
  const raw = String(pipeline?.description || '').trim()
  if (!raw) return ''
  const cleaned = raw.replace(/\s*自动生成流水线$/, '').trim()
  const displayName = getPipelineDisplayName(pipeline)
  return cleaned && cleaned !== displayName ? cleaned : ''
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
    const created = await createSystemPipeline(props.token, workspace.selectedSystemId, {
      name: newPipeline.value.name,
      app_type: newPipeline.value.app_type,
      description: newPipeline.value.description,
      plan_id: workspace.selectedPlanId || '',
    })
    if (newPipeline.value.release_unit_id) {
      const unit = releaseUnits.value.find((u) => u.id === newPipeline.value.release_unit_id)
      if (unit) {
        await upsertPipelineConfig(props.token, {
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

const stopBatchPoller = (pipelineId) => {
  if (runningBatchPollers.value[pipelineId]) {
    clearInterval(runningBatchPollers.value[pipelineId])
    delete runningBatchPollers.value[pipelineId]
  }
}

const startBatchPoller = (pipelineId, batchId) => {
  stopBatchPoller(pipelineId)
  pipelines.value = pipelines.value.map((p) => {
    if (p.id !== pipelineId) return p

    const mainStages = Array.isArray(p.mainStages) && p.mainStages.length > 0
      ? p.mainStages.map((s, i) => normalizeStage(s, i, defaultMainStages[i]?.name))
      : defaultMainStages
    const envStages = Array.isArray(p.envStages) && p.envStages.length > 0
      ? p.envStages.map((s, i) => normalizeStage(s, i, defaultEnvStages[i]?.name))
      : defaultEnvStages
    const ordered = [...mainStages, ...envStages]

    const pendingMap = {}
    ordered.forEach((s) => {
      if (s?.name) pendingMap[s.name] = 'pending'
    })
    const firstName = ordered[0]?.name
    if (firstName) pendingMap[firstName] = 'running'

    return { ...p, live_stages: pendingMap }
  })

  const timerId = setInterval(async () => {
    try {
      const batch = await getBatchStatus(props.token, batchId)
      let stagesMap = {}
      try {
        stagesMap = JSON.parse(batch.stages_status_json || batch.StagesStatusJSON || '{}')
      } catch {}
      pipelines.value = pipelines.value.map((p) =>
        p.id === pipelineId ? { ...p, live_stages: stagesMap } : p
      )
      const batchStatus = batch.status || batch.Status || ''
      if (['success', 'failed', 'cancelled'].includes(batchStatus)) {
        stopBatchPoller(pipelineId)
        await loadLastExecutionTimes()
      }
    } catch {
      stopBatchPoller(pipelineId)
    }
  }, 2000)
  runningBatchPollers.value[pipelineId] = timerId
}

const handleRerunPipeline = async (pipeline) => {
  if (!workspace.selectedSystemId) {
    message.warning('请先选择系统')
    return
  }
  try {
    const resp = await startPipelineExecution(props.token, pipeline.id, workspace.selectedSystemId)
    const batchId = resp?.batch_id || resp?.BatchID || resp?.id || resp?.ID
    message.success('已触发执行')
    if (batchId) {
      startBatchPoller(pipeline.id, batchId)
    } else {
      await loadPipelines()
    }
  } catch (err) {
    message.error('执行失败: ' + (err?.message || '未知错误'))
  }
}

const openExecutionHistory = (pipeline) => {
  router.push({
    path: `/pipelines/${pipeline.id}/executions`,
    query: {
      name: pipeline.name || '',
      plan_version: selectedPlanVersion.value || '',
    },
  })
}

const openBPMDesigner = (pipeline) => {
  router.push({
    path: `/pipelines/${pipeline.id}/bpm`,
    query: {
      name: pipeline.name || '',
      plan_version: selectedPlanVersion.value || '',
    },
  })
}
</script>

<style scoped>
.pipeline-board {
  flex: 1;
  padding: 0;
  overflow-y: auto;
}

.board-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  padding: 20px 24px;
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  position: relative;
  overflow: hidden;
}

.board-toolbar::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--accent-primary), transparent);
  opacity: 0.5;
}

.toolbar-title-wrap {
  flex-shrink: 0;
  position: relative;
}

.toolbar-title {
  margin: 0;
  color: var(--text-primary);
  font-size: 24px;
  line-height: 1.3;
  font-weight: 700;
  font-family: var(--font-display);
  letter-spacing: -0.01em;
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border-radius: 8px;
  font-size: 12px;
  color: white;
  box-shadow: 0 4px 12px rgba(0, 212, 255, 0.3);
}

.title-underline {
  width: 40px;
  height: 3px;
  background: linear-gradient(90deg, var(--accent-primary), transparent);
  border-radius: 2px;
  margin-top: 6px;
}

.toolbar-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  margin-left: auto;
}

.search-shell {
  width: 40px;
  height: 38px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color-light);
  background: var(--bg-tertiary);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  cursor: pointer;
  transition: all var(--transition-base);
}

.search-shell:hover {
  border-color: var(--accent-primary);
  background: var(--bg-elevated);
}

.search-shell.active {
  width: 260px;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px rgba(0, 212, 255, 0.1);
}

.search-icon {
  color: var(--text-tertiary);
  font-size: 14px;
  flex-shrink: 0;
}

.search-input {
  border: 0;
  outline: 0;
  width: 0;
  opacity: 0;
  font-size: 13px;
  color: var(--text-primary);
  background: transparent;
  transition: all var(--transition-base);
  font-family: var(--font-display);
}

.search-shell.active .search-input {
  width: 100%;
  opacity: 1;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.plan-select {
  width: 180px;
}

.icon-btn {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color-light);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.icon-btn:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
  background: var(--bg-elevated);
  transform: translateY(-1px);
}

.secondary-btn {
  height: 38px;
  padding: 0 16px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color-light);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all var(--transition-fast);
  font-family: var(--font-display);
}

.secondary-btn:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
  background: var(--bg-elevated);
  transform: translateY(-1px);
}

.primary-btn {
  height: 38px;
  padding: 0 20px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-info));
  border: none;
  color: var(--bg-primary);
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all var(--transition-base);
  font-family: var(--font-display);
  box-shadow: 0 4px 15px rgba(0, 212, 255, 0.25);
}

.primary-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 212, 255, 0.4);
}

.version-line {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 12px 16px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  animation: slideUp 0.3s ease-out;
}

.version-label {
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-family: var(--font-mono);
}

.version-value {
  color: var(--accent-primary);
  font-size: 14px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.pipeline-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pipeline-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 20px;
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);
  animation: slideUp 0.4s ease-out backwards;
  position: relative;
  overflow: hidden;
}

.pipeline-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--accent-primary), transparent);
  opacity: 0;
  transition: opacity var(--transition-base);
}

.pipeline-card:hover {
  border-color: var(--border-color-accent);
  box-shadow: var(--shadow-lg), var(--shadow-glow);
  transform: translateY(-2px);
}

.pipeline-card:hover::before {
  opacity: 1;
}

.pipeline-one-line {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 14px;
}

.line-name {
  color: var(--text-primary);
  font-weight: 700;
  font-size: 15px;
  margin-right: 4px;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-display);
}

.line-name:hover {
  color: var(--accent-primary);
  text-shadow: 0 0 20px rgba(0, 212, 255, 0.5);
}

.line-tags {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.custom-tag {
  border-radius: 6px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.02em;
  padding: 2px 10px;
  border: 1px solid rgba(59, 130, 246, 0.3);
  background: rgba(59, 130, 246, 0.1);
  color: var(--accent-info);
}

.line-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 13px;
  font-family: var(--font-mono);
}

.line-label {
  color: var(--text-muted);
  font-size: 12px;
  font-family: var(--font-display);
  font-weight: 500;
}

.line-value {
  color: var(--text-primary);
  font-weight: 600;
}

.commit-id {
  font-family: var(--font-mono);
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  letter-spacing: 0.02em;
}

.line-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.action-btn {
  height: 32px;
  padding: 0 14px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color-light);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-weight: 500;
  transition: all var(--transition-fast);
  font-family: var(--font-display);
}

.action-btn:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
  background: var(--bg-elevated);
}

.action-btn-primary {
  height: 32px;
  padding: 0 14px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--accent-primary);
  background: transparent;
  color: var(--accent-primary);
  font-weight: 500;
  transition: all var(--transition-fast);
  font-family: var(--font-display);
}

.action-btn-primary:hover {
  background: rgba(0, 212, 255, 0.1);
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
}

.stage-lane-wrap {
  margin-top: 6px;
  padding-top: 14px;
  border-top: 1px dashed var(--border-color);
}

.lane-caption {
  font-size: 12px;
  color: var(--text-muted);
  margin: 8px 0 10px;
  font-weight: 600;
  font-family: var(--font-display);
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.stage-row {
  display: flex;
  gap: 8px;
  align-items: center;
  overflow-x: auto;
  padding-bottom: 4px;
}

.env-parallel-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
  position: relative;
  padding-left: 18px;
}

.env-parallel-block::before {
  content: '';
  position: absolute;
  left: 0;
  top: 20px;
  bottom: 20px;
  width: 2px;
  background: linear-gradient(180deg, var(--border-color-light), var(--accent-primary), var(--border-color-light));
  border-radius: 1px;
}

.parallel-group-wrap {
  display: flex;
  align-items: stretch;
}

.parallel-entry-link {
  align-self: center;
  margin-right: 8px;
}

.env-row {
  margin-bottom: 0;
}

.stage-col {
  min-width: 170px;
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}

.env-parallel-block .stage-col::before {
  content: '';
  position: absolute;
  left: -18px;
  top: 50%;
  width: 18px;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--border-color-light));
}

.stage-card {
  width: 100%;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  padding: 10px 12px;
  transition: all var(--transition-base);
  position: relative;
}

.stage-card::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  opacity: 0;
  transition: opacity var(--transition-base);
  pointer-events: none;
}

.stage-card.clickable {
  cursor: pointer;
}

.stage-card.clickable:hover {
  border-color: var(--accent-primary);
  background: var(--bg-elevated);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.3);
}

.stage-card.clickable:hover::after {
  opacity: 1;
  box-shadow: inset 0 0 20px rgba(0, 212, 255, 0.05);
}

.env-card {
  background: rgba(0, 212, 255, 0.03);
  border-color: rgba(0, 212, 255, 0.15);
}

.stage-top {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 500;
}

.stage-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stage-duration {
  margin-left: auto;
  color: var(--text-tertiary);
  font-size: 12px;
  font-family: var(--font-mono);
}

.stage-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  position: relative;
}

.dot-success {
  background: var(--accent-success);
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.6);
}

.dot-failed {
  background: var(--accent-danger);
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.6);
}

.dot-running {
  background: var(--accent-primary);
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.6);
  animation: pulse 1.5s ease-in-out infinite;
}

.dot-idle {
  background: var(--text-muted);
}

.dot-pending {
  background: var(--accent-warning);
  box-shadow: 0 0 8px rgba(245, 158, 11, 0.6);
}

.stage-progress {
  margin-top: 10px;
}

.progress-line {
  display: block;
  width: 100%;
  max-width: 140px;
  height: 3px;
  border-radius: 2px;
  transition: all var(--transition-base);
}

.line-success {
  background: linear-gradient(90deg, var(--accent-success), rgba(16, 185, 129, 0.3));
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.4);
}

.line-failed {
  background: linear-gradient(90deg, var(--accent-danger), rgba(239, 68, 68, 0.3));
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.4);
}

.line-running {
  background: linear-gradient(90deg, var(--accent-primary), rgba(0, 212, 255, 0.3));
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.4);
  animation: shimmer 2s linear infinite;
  background-size: 200% 100%;
}

.line-idle {
  background: var(--border-color-light);
}

.line-pending {
  background: linear-gradient(90deg, var(--accent-warning), rgba(245, 158, 11, 0.3));
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.4);
}

.stage-link {
  width: 16px;
  height: 2px;
  background: linear-gradient(90deg, var(--border-color-light), transparent);
  flex-shrink: 0;
}

@media (max-width: 1280px) {
  .board-toolbar {
    flex-direction: column;
    align-items: flex-start;
    padding: 16px;
  }

  .toolbar-actions {
    margin-left: 0;
    width: 100%;
  }

  .line-actions {
    margin-left: 0;
  }

  .toolbar-title {
    font-size: 20px;
  }
}
</style>
