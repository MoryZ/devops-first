<template>
  <div class="release-units-panel">
    <div class="panel-header">
      <div class="header-main">
        <div class="header-text">
          <span class="panel-title">发布单元配置</span>
          <span class="panel-subtitle">{{ workspace.selectedPlanId ? '当前迭代关联的发布单元' : '请先选择迭代计划' }}</span>
        </div>
        <div class="header-actions">
          <a-button type="link" size="small" :disabled="!workspace.selectedPlanId" @click="openBindModal">
            <PlusOutlined />
            添加发布单元
          </a-button>
          <span class="action-divider">|</span>
          <a-popconfirm
            title="确认批量移除当前迭代下所有发布单元关联吗？"
            ok-text="确认"
            cancel-text="取消"
            @confirm="handleRemoveAllBindings"
          >
            <a-button type="link" danger size="small" :disabled="!workspace.selectedPlanId || releaseUnits.length === 0" :loading="removing">
              <DeleteOutlined />
              批量移除
            </a-button>
          </a-popconfirm>
        </div>
      </div>
    </div>

    <div v-if="!workspace.selectedPlanId" class="empty-state">
      <a-empty description="请先选择迭代计划" />
    </div>

    <div v-else class="units-list">
      <a-spin :spinning="loading">
        <div v-if="releaseUnits.length === 0" class="empty-hint">
          当前迭代暂无发布单元，点击右上角“添加发布单元”进行关联
        </div>

        <div v-for="unit in releaseUnits" :key="unit.id" class="unit-card">
          <div class="unit-header">
            <span class="unit-name">{{ unit.name }}</span>
            <a-tag :color="unit.enabled ? 'green' : 'default'">{{ unit.enabled ? '启用' : '停用' }}</a-tag>
          </div>

          <div class="unit-info">
            <div class="info-row">
              <span class="info-label">模块:</span>
              <span class="info-value">{{ unit.module || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">代码库:</span>
              <span class="info-value" :title="unit.repoUrl">{{ unit.repoUrl || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">分支:</span>
              <span class="info-value">{{ unit.branch || 'main' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">源类型:</span>
              <span class="info-value">{{ unit.repositoryType || 'git' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">自动归并:</span>
              <a-tag :color="unit.autoMerge ? 'blue' : 'default'">{{ unit.autoMerge ? 'ON' : 'OFF' }}</a-tag>
            </div>
            <div class="info-row">
              <span class="info-label">自动Tag:</span>
              <a-tag :color="unit.autoTag ? 'cyan' : 'default'">{{ unit.autoTag ? 'ON' : 'OFF' }}</a-tag>
            </div>
          </div>
        </div>
      </a-spin>
    </div>

    <a-modal
      v-model:open="bindModalOpen"
      title="添加发布单元"
      ok-text="确认关联"
      cancel-text="取消"
      :okButtonProps="{ disabled: selectedUnitIdsToBind.length === 0 }"
      :confirm-loading="binding"
      @ok="handleBindReleaseUnit"
    >
      <a-form layout="vertical">
        <a-form-item label="选择一个空间设置中已配置的发布单元">
          <a-select
            v-model:value="selectedUnitIdsToBind"
            mode="multiple"
            placeholder="请选择空间设置中已配置的发布单元"
            show-search
            option-filter-prop="label"
            :filter-option="filterUnitOption"
          >
            <a-select-option
              v-for="unit in allSystemUnits"
              :key="unit.id"
              :value="unit.id"
              :label="unit.name || '未命名发布单元'"
            >
              {{ unit.name || '未命名发布单元' }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
      <div class="modal-hint">可多选。系统会将所选发布单元依次关联到当前迭代下的流水线。</div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { useWorkspaceStore } from '../stores/workspace'
import { listPipelinesByPlan } from '../api/plans'
import { listPipelineConfigs, upsertPipelineConfig, upsertPipelineReleaseUnitBinding } from '../api/pipelines'
import { createSystemPipeline } from '../api/systems'

const props = defineProps({
  token: String,
})

const workspace = useWorkspaceStore()
const releaseUnits = ref([])
const loading = ref(false)
const bindModalOpen = ref(false)
const binding = ref(false)
const removing = ref(false)
const selectedUnitIdsToBind = ref([])
const allSystemUnits = ref([])

const readField = (obj, snakeKey, camelKey, fallback) => {
  if (obj?.[snakeKey] !== undefined && obj?.[snakeKey] !== null) return obj[snakeKey]
  if (camelKey && obj?.[camelKey] !== undefined && obj?.[camelKey] !== null) return obj[camelKey]
  return fallback
}

const buildConfigPayload = ({ pipelineId, pipelineName, baseConfig = {}, unit, releaseUnitId }) => {
  const resolvedReleaseUnitId =
    releaseUnitId !== undefined
      ? releaseUnitId
      : unit?.id || readField(baseConfig, 'release_unit_id', 'releaseUnitId', '')

  return {
    ...baseConfig,
    pipeline_id: pipelineId,
    name: pipelineName || readField(baseConfig, 'name', 'name', ''),
    release_unit_id: resolvedReleaseUnitId,
    repository_type: unit?.repositoryType || readField(baseConfig, 'repository_type', 'repositoryType', 'git'),
    auto_merge: unit?.autoMerge ?? readField(baseConfig, 'auto_merge', 'autoMerge', true),
    auto_tag: unit?.autoTag ?? readField(baseConfig, 'auto_tag', 'autoTag', true),
    display_order: Number(unit?.displayOrder) || Number(readField(baseConfig, 'display_order', 'displayOrder', 0)) || 0,
    repo_url: unit?.repoUrl || readField(baseConfig, 'repo_url', 'repoUrl', ''),
    branch: unit?.branch || readField(baseConfig, 'branch', 'branch', 'main'),
    project_path: readField(baseConfig, 'project_path', 'projectPath', ''),
    build_type: readField(baseConfig, 'build_type', 'buildType', 'maven'),
    maven_command: readField(baseConfig, 'maven_command', 'mavenCommand', 'mvn clean package -DskipTests'),
    deploy_type: readField(baseConfig, 'deploy_type', 'deployType', 'docker'),
    docker_image: readField(baseConfig, 'docker_image', 'dockerImage', ''),
    docker_container: readField(baseConfig, 'docker_container', 'dockerContainer', ''),
    docker_run_args: readField(baseConfig, 'docker_run_args', 'dockerRunArgs', ''),
    main_stages: readField(baseConfig, 'main_stages', 'mainStages', []),
    env_stages: readField(baseConfig, 'env_stages', 'envStages', []),
  }
}

const getPlanPipelineConfigMap = async (planPipelines) => {
  const ids = new Set(planPipelines.map((p) => String(p.ID || p.id)))
  const allConfigsData = await listPipelineConfigs(props.token)
  const allConfigs = allConfigsData?.items || allConfigsData?.data || []
  const map = new Map()
  for (const cfg of allConfigs) {
    const pid = String(readField(cfg, 'pipeline_id', 'pipelineId', ''))
    if (!pid || !ids.has(pid)) continue
    map.set(pid, cfg)
  }
  return map
}

const loadAllSystemUnits = () => {
  const storageKey = `releaseUnits:${workspace.selectedSystemId}`
  try {
    const raw = localStorage.getItem(storageKey)
    const items = raw ? JSON.parse(raw) : []
    allSystemUnits.value = Array.isArray(items)
      ? items.map((item, idx) => ({
          ...item,
          id: String(item.id || ''),
          displayOrder: Number(item.displayOrder) || idx + 1,
        }))
      : []
  } catch {
    allSystemUnits.value = []
  }
}

const filterUnitOption = (input, option) => {
  const text = String(option?.label || '').toLowerCase()
  return text.includes(String(input || '').toLowerCase())
}

const openBindModal = () => {
  if (!workspace.selectedPlanId) {
    message.warning('请先选择迭代计划')
    return
  }
  loadAllSystemUnits()
  if (!allSystemUnits.value.length) {
    message.warning('当前系统暂无发布单元，请先在“发布单元”页面创建')
    return
  }
  selectedUnitIdsToBind.value = []
  bindModalOpen.value = true
}

const handleBindReleaseUnit = async () => {
  if (!selectedUnitIdsToBind.value.length) {
    message.warning('请至少选择一个发布单元')
    return
  }
  if (!workspace.selectedPlanId) {
    message.warning('请先选择迭代计划')
    return
  }

  binding.value = true
  try {
    const selectedIds = Array.from(new Set(selectedUnitIdsToBind.value.map((id) => String(id))))
    const unitById = new Map(allSystemUnits.value.map((u) => [String(u.id), u]))

    const pipelinesData = await listPipelinesByPlan(props.token, workspace.selectedPlanId)
    let pipelines = pipelinesData.items || []

    // No pipeline in current plan: create base pipelines first, then bind release units.
    if (!pipelines.length) {
      const createdPipelines = []

      for (const unitId of selectedIds) {
        const unit = unitById.get(unitId)
        if (!unit) continue

        const created = await createSystemPipeline(props.token, workspace.selectedSystemId, {
          plan_id: workspace.selectedPlanId,
          name: `${unit.name || '发布单元'}-pipeline`,
          app_type: 'java',
          description: `${unit.name || '发布单元'} 自动生成流水线`,
        })

        await upsertPipelineConfig(props.token, {
          ...buildConfigPayload({
            pipelineId: String(created.ID || created.id),
            pipelineName: created.Name || created.name,
            unit,
            releaseUnitId: unit.id,
          }),
        })

        createdPipelines.push(created)
      }

      pipelines = createdPipelines
    } else {
      const pipelineStates = await Promise.all(
        pipelines.map(async (pipeline) => {
          const pipelineId = String(pipeline.ID || pipeline.id)
          return {
            pipeline,
            pipelineId,
            currentUnitId: String(
              readField(
                pipeline,
                'release_unit_id',
                'releaseUnitId',
                readField(pipeline, 'ReleaseUnitID', 'ReleaseUnitId', '')
              )
            ),
          }
        })
      )

      const usedSelectedIds = new Set()

      // First pass: update pipelines already bound to selected release units.
      for (const state of pipelineStates) {
        if (!state.currentUnitId || !selectedIds.includes(state.currentUnitId)) continue
        const unit = unitById.get(state.currentUnitId)
        if (!unit) continue
        usedSelectedIds.add(state.currentUnitId)
        await upsertPipelineReleaseUnitBinding(props.token, state.pipelineId, state.currentUnitId)
      }

      // Second pass: bind remaining selected release units to unbound/unselected pipelines.
      const remainingSelectedIds = selectedIds.filter((id) => !usedSelectedIds.has(id))
      const candidatePipelines = pipelineStates.filter((state) => !selectedIds.includes(state.currentUnitId))

      for (let i = 0; i < remainingSelectedIds.length && i < candidatePipelines.length; i += 1) {
        const unitId = remainingSelectedIds[i]
        const unit = unitById.get(unitId)
        const state = candidatePipelines[i]
        if (!unit || !state) continue
        usedSelectedIds.add(unitId)
        await upsertPipelineReleaseUnitBinding(props.token, state.pipelineId, unitId)
      }

      // Third pass: create pipelines for selected release units still not bound.
      const needCreateIds = selectedIds.filter((id) => !usedSelectedIds.has(id))
      for (const unitId of needCreateIds) {
        const unit = unitById.get(unitId)
        if (!unit) continue
        const created = await createSystemPipeline(props.token, workspace.selectedSystemId, {
          plan_id: workspace.selectedPlanId,
          name: `${unit.name || '发布单元'}-pipeline`,
          app_type: 'java',
          description: `${unit.name || '发布单元'} 自动生成流水线`,
        })
        await upsertPipelineConfig(props.token, {
          ...buildConfigPayload({
            pipelineId: String(created.ID || created.id),
            pipelineName: created.Name || created.name,
            unit,
            releaseUnitId: unit.id,
          }),
        })
      }
    }

    bindModalOpen.value = false
    message.success('发布单元关联成功，流水线基础信息已就绪')
    await loadReleaseUnits()
  } catch (err) {
    message.error('关联发布单元失败: ' + (err?.message || '未知错误'))
  } finally {
    binding.value = false
  }
}

const handleRemoveAllBindings = async () => {
  if (!workspace.selectedPlanId) {
    message.warning('请先选择迭代计划')
    return
  }

  removing.value = true
  try {
    const pipelinesData = await listPipelinesByPlan(props.token, workspace.selectedPlanId)
    const pipelines = pipelinesData.items || []

    if (!pipelines.length) {
      message.info('当前迭代下暂无流水线')
      return
    }

    await Promise.all(
      pipelines.map(async (pipeline) => {
        const pipelineId = String(pipeline.ID || pipeline.id)
        await upsertPipelineReleaseUnitBinding(props.token, pipelineId, '')
      })
    )

    message.success('已批量移除当前迭代下所有发布单元关联')
    await loadReleaseUnits()
  } catch (err) {
    message.error('批量移除失败: ' + (err?.message || '未知错误'))
  } finally {
    removing.value = false
  }
}

const loadReleaseUnits = async () => {
  if (!workspace.selectedPlanId || !workspace.selectedSystemId) {
    releaseUnits.value = []
    return
  }

  loading.value = true
  try {
    // First, get all pipelines in the plan
    const pipelinesData = await listPipelinesByPlan(props.token, workspace.selectedPlanId)
    const pipelines = pipelinesData.items || []

    const configMap = await getPlanPipelineConfigMap(pipelines)

    // Collect release units from all pipelines
    const unitMap = new Map()

    for (const pipeline of pipelines) {
      const pipelineId = pipeline.ID || pipeline.id

      const config = configMap.get(String(pipelineId)) || {}
      const releaseUnitId = readField(config, 'release_unit_id', 'releaseUnitId', '')

      if (releaseUnitId && !unitMap.has(releaseUnitId)) {
        unitMap.set(releaseUnitId, {
          id: releaseUnitId,
          pipelineId: pipelineId,
        })
      }
    }

    // Fetch release unit details from localStorage (since they're stored per system)
    loadAllSystemUnits()
    const allUnits = allSystemUnits.value

    releaseUnits.value = allUnits.filter((unit) => unitMap.has(String(unit.id)))
  } catch (err) {
    message.error('加载发布单元配置失败: ' + err.message)
    releaseUnits.value = []
  } finally {
    loading.value = false
  }
}

watch(
  () => workspace.selectedPlanId,
  () => {
    loadReleaseUnits()
  }
)

watch(
  () => workspace.selectedSystemId,
  () => {
    loadReleaseUnits()
  }
)

watch(
  () => workspace.selectedPlanId,
  () => {
    loadAllSystemUnits()
  },
  { immediate: true }
)
</script>

<style scoped>
.release-units-panel {
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
  flex-direction: column;
  gap: 8px;
  padding: 14px;
  border-bottom: 1px solid #e8e8e8;
}

.header-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.action-divider {
  color: #d0d7e6;
  font-size: 12px;
  line-height: 1;
}

.header-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #0a1630;
}

.panel-subtitle {
  font-size: 12px;
  color: #8c8c8c;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.units-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.empty-hint {
  padding: 40px 20px;
  text-align: center;
  color: #8c8c8c;
  font-size: 12px;
}

.modal-hint {
  color: #8c8c8c;
  font-size: 12px;
}

.unit-card {
  margin-bottom: 12px;
  padding: 12px;
  border: 1px solid #dce5f2;
  border-radius: 8px;
  background: #fafbfe;
}

.unit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.unit-name {
  font-weight: 600;
  color: #0a1630;
  font-size: 14px;
}

.unit-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.info-label {
  color: #8c8c8c;
  min-width: 60px;
  flex-shrink: 0;
}

.info-value {
  color: #262626;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
