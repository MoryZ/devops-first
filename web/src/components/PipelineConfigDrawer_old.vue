<template>
  <a-drawer
    :open="open"
    width="85vw"
    :closable="false"
    :destroyOnClose="false"
    @update:open="updateOpen"
  >
    <template #title>
      <div class="drawer-title-row">
        <span class="drawer-title">{{ pipelineName || 'pipeline' }}</span>
        <div class="drawer-actions">
          <a-button @click="$emit('close')">取消</a-button>
          <a-button type="primary" @click="handleSave">仅保存</a-button>
        </div>
      </div>
    </template>

    <div class="drawer-tabs">
      <button 
        class="drawer-tab" 
        :class="{ active: activeTab === 'basic' }"
        @click="activeTab = 'basic'"
      >基本信息</button>
      <button 
        class="drawer-tab"
        :class="{ active: activeTab === 'flow' }"
        @click="activeTab = 'flow'"
      >流程配置</button>
      <button 
        class="drawer-tab"
        :class="{ active: activeTab === 'trigger' }"
        @click="activeTab = 'trigger'"
      >触发设置</button>
    </div>

    <!-- 基本信息标签页 -->
    <div v-show="activeTab === 'basic'" class="tab-content">
      <a-card :bordered="false" style="margin-bottom: 16px">
        <a-form layout="vertical">
          <a-form-item label="发布单元">
            <a-select
              v-model:value="form.releaseUnitId"
              placeholder="选择发布单元（可选）"
              allow-clear
              @change="handleReleaseUnitChange"
            >
              <a-select-option v-for="unit in releaseUnits" :key="unit.id" :value="unit.id">
                {{ unit.displayOrder }} - {{ unit.name }}
              </a-select-option>
            </a-select>
          </a-form-item>

          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="触发源类型">
                <a-select v-model:value="form.repositoryType" placeholder="选择触发源类型">
                  <a-select-option value="git">git</a-select-option>
                  <a-select-option value="svn">svn</a-select-option>
                  <a-select-option value="artifact">artifact</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="展示顺序">
                <a-input-number v-model:value="form.displayOrder" :min="0" style="width: 100%" />
              </a-form-item>
            </a-col>
          </a-row>

          <a-form-item label="仓库地址">
            <a-input v-model:value="form.repoUrl" placeholder="git@gitlab.example.com:group/project.git" />
          </a-form-item>

          <a-form-item label="分支">
            <a-input v-model:value="form.branch" placeholder="master / feature-xxx" />
          </a-form-item>

          <a-form-item label="本地项目路径">
            <a-input v-model:value="form.projectPath" placeholder="/Users/name/projects/your-service" />
          </a-form-item>

          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="自动代码归并">
                <a-switch v-model:checked="form.autoMerge" checked-children="ON" un-checked-children="OFF" />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="自动打Tag">
                <a-switch v-model:checked="form.autoTag" checked-children="ON" un-checked-children="OFF" />
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </a-card>
    </div>

    <!-- 流程配置标签页 -->
    <div v-show="activeTab === 'flow'" class="tab-content">
      <div class="config-canvas">
        <section class="config-lane">
          <h4>流水线源</h4>
          <a-card class="lane-card" :bordered="false">
            <a-form layout="vertical">
              <a-form-item label="仓库地址">
                <a-input v-model:value="form.repoUrl" placeholder="git@gitlab.example.com:group/project.git" />
              </a-form-item>
              <a-form-item label="分支">
                <a-input v-model:value="form.branch" placeholder="master / feature-xxx" />
              </a-form-item>
            </a-form>
          </a-card>
        </section>

        <section class="config-lane">
          <h4>构建</h4>
          <a-card class="lane-card" :bordered="false">
            <a-form layout="vertical">
              <a-form-item label="构建方式">
                <a-select v-model:value="form.buildType" placeholder="选择构建方式">
                  <a-select-option value="maven">Maven</a-select-option>
                  <a-select-option value="gradle">Gradle</a-select-option>
                  <a-select-option value="npm">NPM</a-select-option>
                  <a-select-option value="none">不构建</a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item v-if="form.buildType === 'maven'" label="Maven 命令">
                <a-input v-model:value="form.mavenCommand" placeholder="mvn clean package -DskipTests" />
              </a-form-item>
              <a-form-item v-else-if="form.buildType === 'npm'" label="NPM 命令">
                <a-input v-model:value="form.npmCommand" placeholder="npm run build" />
              </a-form-item>
            </a-form>
          </a-card>
        </section>

        <section class="config-lane">
          <h4>部署</h4>
          <a-card class="lane-card" :bordered="false">
            <a-form layout="vertical">
              <a-form-item label="部署方式">
                <a-select v-model:value="form.deployType" placeholder="选择部署方式">
                  <a-select-option value="docker">Docker</a-select-option>
                  <a-select-option value="jar">JAR</a-select-option>
                  <a-select-option value="war">WAR</a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item v-if="form.deployType === 'docker'" label="镜像名称">
                <a-input v-model:value="form.dockerImage" placeholder="service-name:latest" />
              </a-form-item>
              <a-form-item v-if="form.deployType === 'docker'" label="容器名称">
                <a-input v-model:value="form.dockerContainer" placeholder="service-name" />
              </a-form-item>
              <a-form-item v-if="form.deployType === 'docker'" label="运行参数">
                <a-input v-model:value="form.dockerRunArgs" placeholder="-d -p 8080:8080" />
              </a-form-item>
            </a-form>
          </a-card>
        </section>
      </div>

      <div class="stage-editor-section">
        <div class="stage-editor-group">
          <div class="stage-editor-label">
            主流程阶段
            <a-button size="small" type="dashed" @click="addMainStage"><PlusOutlined /> 添加</a-button>
          </div>
        <VueDraggable
          v-model="editingMainStages"
          handle=".drag-handle"
          :animation="150"
          class="draggable-stage-list"
        >
          <div v-for="(stage, idx) in editingMainStages" :key="idx" class="draggable-stage-item">
            <span class="drag-handle"><HolderOutlined /></span>
            <a-input v-model:value="stage.name" size="small" class="stage-name-input" />
            <a-select v-model:value="stage.status" size="small" class="stage-status-select">
              <a-select-option value="success">成功</a-select-option>
              <a-select-option value="failed">失败</a-select-option>
              <a-select-option value="pending">待执行</a-select-option>
              <a-select-option value="running">执行中</a-select-option>
            </a-select>
            <a-button type="text" danger size="small" @click="removeMainStage(idx)">
              <MinusCircleOutlined />
            </a-button>
          </div>
        </VueDraggable>
      </div>

      <div class="stage-editor-group">
        <div class="stage-editor-label">
          环境阶段
          <a-button size="small" type="dashed" @click="addEnvStage"><PlusOutlined /> 添加</a-button>
        </div>
        <VueDraggable
          v-model="editingEnvStages"
          handle=".drag-handle"
          :animation="150"
          class="draggable-stage-list"
        >
          <div v-for="(stage, idx) in editingEnvStages" :key="idx" class="draggable-stage-item">
            <span class="drag-handle"><HolderOutlined /></span>
            <a-input v-model:value="stage.name" size="small" class="stage-name-input" />
            <a-select v-model:value="stage.status" size="small" class="stage-status-select">
              <a-select-option value="success">成功</a-select-option>
              <a-select-option value="failed">失败</a-select-option>
              <a-select-option value="pending">待执行</a-select-option>
              <a-select-option value="running">执行中</a-select-option>
            </a-select>
            <a-button type="text" danger size="small" @click="removeEnvStage(idx)">
              <MinusCircleOutlined />
            </a-button>
          </div>
        </VueDraggable>
      </div>
    </div>

    <!-- 触发设置标签页 -->
    <div v-show="activeTab === 'trigger'" class="tab-content">
      <a-card :bordered="false" title="触发方式" style="margin-bottom: 16px">
        <a-form layout="vertical">
          <a-form-item label="手动触发">
            <a-switch 
              v-model:checked="form.triggerManual" 
              checked-children="启用" 
              un-checked-children="禁用" 
            />
          </a-form-item>

          <a-form-item label="Webhook触发">
            <a-switch 
              v-model:checked="form.triggerWebhook" 
              checked-children="启用" 
              un-checked-children="禁用" 
            />
          </a-form-item>

          <a-form-item v-if="form.triggerWebhook" label="Webhook 地址">
            <div style="display: flex; gap: 8px">
              <a-input 
                :value="`${window.location.origin}/webhook/pipeline/${pipelineId}`" 
                disabled 
              />
              <a-button type="primary" size="middle" @click="copyWebhookUrl">复制</a-button>
            </div>
          </a-form-item>

          <a-form-item label="定时触发">
            <a-switch 
              v-model:checked="form.triggerSchedule" 
              checked-children="启用" 
              un-checked-children="禁用" 
            />
          </a-form-item>

          <a-form-item v-if="form.triggerSchedule" label="执行周期 (Cron 表达式)">
            <a-input 
              v-model:value="form.scheduleCron" 
              placeholder="0 0 * * * (每天午夜)" 
            />
          </a-form-item>

          <a-form-item label="并发限制">
            <a-input-number 
              v-model:value="form.concurrencyLimit" 
              :min="1" 
              :max="10"
              style="width: 100%"
            />
          </a-form-item>
        </a-form>
      </a-card>

      <a-card :bordered="false" title="高级设置" style="margin-bottom: 16px">
        <a-form layout="vertical">
          <a-form-item label="失败重试">
            <a-switch 
              v-model:checked="form.enableRetry" 
              checked-children="启用" 
              un-checked-children="禁用" 
            />
          </a-form-item>

          <a-form-item v-if="form.enableRetry" label="重试次数">
            <a-input-number 
              v-model:value="form.retryCount" 
              :min="1" 
              :max="5"
              style="width: 100%"
            />
          </a-form-item>

          <a-form-item label="超时设置 (分钟)">
            <a-input-number 
              v-model:value="form.timeoutMinutes" 
              :min="5" 
              :max="180"
              style="width: 100%"
            />
          </a-form-item>

          <a-form-item label="通知设置">
            <a-checkbox-group v-model:value="form.notifyChannels">
              <a-checkbox value="email">邮件通知</a-checkbox>
              <a-checkbox value="webhook">Webhook通知</a-checkbox>
              <a-checkbox value="dingtalk">钉钉通知</a-checkbox>
            </a-checkbox-group>
          </a-form-item>
        </a-form>
      </a-card>
    </div>

    <!-- Execution Console （保留在底部，所有标签页都可见） -->
    <ExecutionConsole
      v-show="activeTab === 'flow'"
      :token="token"
      :pipelineName="pipelineName"
      :initialProjectPath="form.projectPath"
      :releaseUnitName="selectedReleaseUnit?.name || ''"
      :repositoryType="form.repositoryType"
      :autoMerge="form.autoMerge"
      :autoTag="form.autoTag"
      class="drawer-console"
    />
  </a-drawer>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, HolderOutlined, MinusCircleOutlined } from '@ant-design/icons-vue'
import { VueDraggable } from 'vue-draggable-plus'
import ExecutionConsole from './ExecutionConsole.vue'

const props = defineProps({
  open: Boolean,
  token: String,
  pipelineId: String,
  pipelineName: String,
  pipeline: Object,
  systemId: String,
})

const emit = defineEmits(['close', 'save'])

const activeTab = ref('basic')

const form = ref({
  releaseUnitId: '',
  repositoryType: 'git',
  autoMerge: true,
  autoTag: true,
  displayOrder: 0,
  repoUrl: '',
  branch: '',
  projectPath: '',
  buildType: 'maven',
  mavenCommand: 'mvn clean package -DskipTests',
  npmCommand: 'npm run build',
  deployType: 'docker',
  dockerImage: '',
  dockerContainer: '',
  dockerRunArgs: '',
  // 触发设置
  triggerManual: true,
  triggerWebhook: false,
  triggerSchedule: false,
  scheduleCron: '0 0 * * *',
  concurrencyLimit: 1,
  enableRetry: false,
  retryCount: 2,
  timeoutMinutes: 60,
  notifyChannels: [],
})

const releaseUnits = ref([])

const selectedReleaseUnit = computed(() => {
  if (!form.value.releaseUnitId) return null
  return releaseUnits.value.find((u) => u.id === form.value.releaseUnitId) || null
})

const editingMainStages = ref([])
const editingEnvStages = ref([])

watch(
  () => props.pipeline,
  (newPipeline) => {
    if (newPipeline) {
      form.value = {
        releaseUnitId: '',
        repositoryType: 'git',
        autoMerge: true,
        autoTag: true,
        displayOrder: 0,
        repoUrl: '',
        branch: '',
        projectPath: '',
        buildType: 'maven',
        mavenCommand: 'mvn clean package -DskipTests',
        npmCommand: 'npm run build',
        deployType: 'docker',
        dockerImage: '',
        dockerContainer: '',
        dockerRunArgs: '',
        triggerManual: true,
        triggerWebhook: false,
        triggerSchedule: false,
        scheduleCron: '0 0 * * *',
        concurrencyLimit: 1,
        enableRetry: false,
        retryCount: 2,
        timeoutMinutes: 60,
        notifyChannels: [],
        ...(newPipeline.config || {}),
      }
      editingMainStages.value = (newPipeline.mainStages || []).map((s) => ({ ...s }))
      editingEnvStages.value = (newPipeline.envStages || []).map((s) => ({ ...s }))
    }
  },
  { deep: true }
)

watch(
  () => props.systemId,
  (systemId) => {
    if (!systemId) {
      releaseUnits.value = []
      return
    }
    try {
      const raw = localStorage.getItem(`releaseUnits:${systemId}`)
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
  },
  { immediate: true }
)

const handleReleaseUnitChange = (unitId) => {
  const unit = releaseUnits.value.find((u) => u.id === unitId)
  if (!unit) return
  form.value.repositoryType = unit.repositoryType || 'git'
  form.value.autoMerge = unit.autoMerge ?? true
  form.value.autoTag = unit.autoTag ?? true
  form.value.displayOrder = Number(unit.displayOrder) || 0
  form.value.repoUrl = unit.repoUrl || form.value.repoUrl
  form.value.branch = unit.branch || form.value.branch
}

const updateOpen = (newVal) => {
  if (!newVal) {
    emit('close')
  }
}

const addMainStage = () => {
  editingMainStages.value.push({ name: '新阶段', duration: '-', status: 'pending' })
}

const removeMainStage = (idx) => {
  editingMainStages.value.splice(idx, 1)
}

const addEnvStage = () => {
  editingEnvStages.value.push({ name: '新环境', duration: '-', status: 'pending' })
}

const removeEnvStage = (idx) => {
  editingEnvStages.value.splice(idx, 1)
}

const copyWebhookUrl = () => {
  const url = `${window.location.origin}/webhook/pipeline/${props.pipelineId}`
  navigator.clipboard.writeText(url).then(() => {
    message.success('Webhook地址已复制')
  }).catch(() => {
    message.error('复制失败，请手动复制')
  })
}

const handleSave = async () => {
  if (!form.value.projectPath) {
    message.warning('请填写项目路径')
    return
  }

  try {
    const res = await fetch('/api/pipelines/config', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${props.token}`,
      },
      body: JSON.stringify({
        pipeline_id: props.pipelineId,
        name: props.pipelineName,
        release_unit_id: form.value.releaseUnitId,
        repository_type: form.value.repositoryType,
        auto_merge: form.value.autoMerge,
        auto_tag: form.value.autoTag,
        display_order: Number(form.value.displayOrder) || 0,
        repo_url: form.value.repoUrl,
        branch: form.value.branch,
        project_path: form.value.projectPath,
        maven_command: form.value.mavenCommand,
        docker_image: form.value.dockerImage,
        docker_container: form.value.dockerContainer,
        docker_run_args: form.value.dockerRunArgs,
        main_stages: editingMainStages.value,
        env_stages: editingEnvStages.value,
        // 触发设置
        trigger_manual: form.value.triggerManual,
        trigger_webhook: form.value.triggerWebhook,
        trigger_schedule: form.value.triggerSchedule,
        schedule_cron: form.value.scheduleCron,
        concurrency_limit: form.value.concurrencyLimit,
        enable_retry: form.value.enableRetry,
        retry_count: form.value.retryCount,
        timeout_minutes: form.value.timeoutMinutes,
        notify_channels: form.value.notifyChannels,
      }),
    })
    if (!res.ok) {
      const err = await res.json()
      message.error(err.error || '保存失败')
      return
    }
    message.success('流程配置已保存')
    emit('save')
  } catch (err) {
    message.error('网络错误: ' + err.message)
  }
}
</script>

<style scoped>
.drawer-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.drawer-title {
  font-size: 20px;
  font-weight: 700;
  color: #1f2f46;
}

.drawer-actions {
  display: flex;
  gap: 8px;
}

.drawer-tabs {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 0 6px;
  border-bottom: 1px solid #dce3ef;
  margin-bottom: 16px;
}

.drawer-tab {
  border: 0;
  background: transparent;
  color: #617089;
  font-size: 16px;
  padding: 8px 12px;
  border-radius: 10px;
  cursor: pointer;
}

.drawer-tab.active {
  color: #1a5fd0;
  background: #dbeeff;
  font-weight: 600;
}

.config-canvas {
  display: grid;
  grid-template-columns: repeat(3, minmax(260px, 1fr));
  gap: 14px;
  margin-bottom: 16px;
}

.config-lane {
  min-height: 300px;
  border-left: 1px solid #d4dde9;
  padding-left: 14px;
}

.config-lane h4 {
  margin: 0 0 10px;
  color: #6c7c94;
  font-size: 32px;
  font-weight: 300;
}

.lane-card {
  border-radius: 14px;
  box-shadow: 0 10px 24px rgba(18, 35, 58, 0.09);
}

.stage-editor-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 14px;
}

.stage-editor-group {
  background: #fff;
  border-radius: 12px;
  padding: 14px;
  box-shadow: 0 4px 14px rgba(18, 35, 58, 0.07);
}

.stage-editor-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  color: #3a4d6a;
  margin-bottom: 10px;
  font-size: 14px;
}

.draggable-stage-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 32px;
}

.draggable-stage-item {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f5f8ff;
  border: 1px solid #dce6f5;
  border-radius: 8px;
  padding: 6px 8px;
}

.drag-handle {
  color: #a0aec0;
  cursor: grab;
  font-size: 14px;
  padding: 0 2px;
}

.drag-handle:active {
  cursor: grabbing;
}

.stage-name-input {
  flex: 1;
}

.stage-status-select {
  width: 96px;
}

.drawer-console {
  margin-top: 14px;
}

.tab-content {
  animation: fadeIn 0.2s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@media (max-width: 900px) {
  .config-canvas {
    grid-template-columns: 1fr;
  }

  .config-lane {
    border-left: 0;
    border-top: 1px solid #d4dde9;
    padding-left: 0;
    padding-top: 10px;
  }

  .stage-editor-section {
    grid-template-columns: 1fr;
  }
}
</style>
