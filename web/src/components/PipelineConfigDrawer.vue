<template>
  <a-drawer
    :open="open"
    width="76vw"
    :closable="false"
    :destroyOnClose="false"
    :bodyStyle="{ padding: '0', display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }"
    @update:open="updateOpen"
  >
    <template #title>
      <div class="drawer-header">
        <span class="drawer-title">{{ pipelineName || 'pipeline' }}</span>
        <div class="templates-row">
          <span class="templates-label">快速模板</span>
          <button class="tpl-btn" @click="applyTemplate('java')">Java</button>
          <button class="tpl-btn" @click="applyTemplate('node')">Node.js</button>
          <button class="tpl-btn" @click="applyTemplate('static')">Static</button>
        </div>
        <div class="header-actions">
          <a-button @click="$emit('close')">取消</a-button>
          <a-button type="primary" :disabled="missingFields.length > 0" @click="handleSave">保存配置</a-button>
        </div>
      </div>
    </template>

    <div class="health-bar-wrap">
      <div class="health-bar-track">
        <div class="health-bar-fill" :class="healthClass" :style="{ width: completenessPercent + '%' }"></div>
      </div>
      <div class="health-bar-text" :class="healthClass">
        <template v-if="missingFields.length === 0">
          <CheckCircleOutlined style="margin-right: 5px" />配置完整
        </template>
        <template v-else>
          <ExclamationCircleOutlined style="margin-right: 5px" />还需填写：{{ missingFields.join('、') }}
        </template>
      </div>
    </div>

    <div class="drawer-body">
      <div class="config-form">
        <a-collapse v-model:activeKey="openSections" :bordered="false" expand-icon-position="end" class="config-collapse">
          <a-collapse-panel key="source" class="config-panel">
            <template #header>
              <div class="panel-header">
                <span class="panel-num">1</span>
                <span class="panel-title">源码</span>
                <span v-if="form.repoUrl" class="panel-hint">{{ form.repoUrl }}</span>
                <span v-else-if="form.projectPath" class="panel-hint">使用已有目录: {{ form.projectPath }}</span>
                <span v-else class="panel-hint panel-hint-missing">未配置源码来源</span>
              </div>
            </template>

            <a-form layout="vertical" class="panel-form">
              <a-form-item label="发布单元">
                <a-select v-model:value="form.releaseUnitId" placeholder="选择发布单元（可选）" allow-clear @change="handleReleaseUnitChange">
                  <a-select-option v-for="unit in releaseUnits" :key="unit.id" :value="unit.id">
                    {{ unit.displayOrder }} - {{ unit.name }}
                  </a-select-option>
                </a-select>
              </a-form-item>

              <a-row :gutter="12">
                <a-col :span="16">
                  <a-form-item label="仓库地址（HTTPS）">
                    <a-input v-model:value="form.repoUrl" placeholder="https://github.com/org/repo.git" />
                  </a-form-item>
                </a-col>
                <a-col :span="8">
                  <a-form-item label="分支">
                    <a-input v-model:value="form.branch" placeholder="main" />
                  </a-form-item>
                </a-col>
              </a-row>

              <a-row :gutter="12">
                <a-col :span="24">
                  <a-form-item label="凭证命名空间">
                    <a-select
                      v-model:value="form.gitCredentialKey"
                      placeholder="选择全局变量命名空间（如 github）"
                      :loading="loadingGlobalVariables"
                    >
                      <a-select-option v-for="gv in globalVariables" :key="gv.key" :value="gv.key">
                        {{ gv.key }}
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                </a-col>
              </a-row>

              <a-row :gutter="12" v-if="form.gitCredentialKey">
                <a-col :span="12">
                  <a-form-item label="Git 用户名字段">
                    <a-select
                      v-model:value="form.gitUsernameField"
                      placeholder="选择用户名字段"
                      show-search
                      allow-clear
                    >
                      <a-select-option v-for="field in getGlobalVariableOptions(form.gitCredentialKey)" :key="field.value" :value="field.value">
                        {{ field.label }}
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item label="Git 令牌字段">
                    <a-select
                      v-model:value="form.gitTokenField"
                      placeholder="选择令牌字段"
                      show-search
                      allow-clear
                    >
                      <a-select-option v-for="field in getGlobalVariableOptions(form.gitCredentialKey)" :key="field.value" :value="field.value">
                        {{ field.label }}
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                </a-col>
              </a-row>

              <a-form-item label="项目路径（可选）">
                <a-input v-model:value="form.projectPath" placeholder="/data/apps/order-service" />
              </a-form-item>
            </a-form>
          </a-collapse-panel>

          <a-collapse-panel key="build" class="config-panel">
            <template #header>
              <div class="panel-header">
                <span class="panel-num">2</span>
                <span class="panel-title">构建</span>
                <span class="panel-hint">{{ buildSummary }}</span>
              </div>
            </template>
            <a-form layout="vertical" class="panel-form">
              <a-form-item label="构建方式">
                <a-radio-group v-model:value="form.buildType" button-style="solid">
                  <a-radio-button value="maven">Maven</a-radio-button>
                  <a-radio-button value="gradle">Gradle</a-radio-button>
                  <a-radio-button value="npm">NPM</a-radio-button>
                  <a-radio-button value="none">不构建</a-radio-button>
                </a-radio-group>
              </a-form-item>
              <a-form-item v-if="form.buildType === 'maven'" label="Maven 命令">
                <a-input v-model:value="form.mavenCommand" placeholder="mvn clean package -DskipTests" />
              </a-form-item>
              <a-form-item v-if="form.buildType === 'gradle'" label="Gradle 命令">
                <a-input v-model:value="form.gradleCommand" placeholder="./gradlew clean bootJar" />
              </a-form-item>
              <a-form-item v-if="form.buildType === 'npm'" label="NPM 命令">
                <a-input v-model:value="form.npmCommand" placeholder="npm run build" />
              </a-form-item>
            </a-form>
          </a-collapse-panel>

          <a-collapse-panel key="deploy" class="config-panel">
            <template #header>
              <div class="panel-header">
                <span class="panel-num">3</span>
                <span class="panel-title">部署</span>
                <span class="panel-hint">{{ form.deployType }}</span>
              </div>
            </template>
            <a-form layout="vertical" class="panel-form">
              <a-form-item label="部署方式">
                <a-radio-group v-model:value="form.deployType" button-style="solid">
                  <a-radio-button value="docker">Docker</a-radio-button>
                  <a-radio-button value="jar">JAR</a-radio-button>
                  <a-radio-button value="war">WAR</a-radio-button>
                </a-radio-group>
              </a-form-item>
              <template v-if="form.deployType === 'docker'">
                <a-row :gutter="12">
                  <a-col :span="12">
                    <a-form-item label="镜像名称">
                      <a-input v-model:value="form.dockerImage" placeholder="order-service:latest" />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="容器名称">
                      <a-input v-model:value="form.dockerContainer" placeholder="order-service-prod" />
                    </a-form-item>
                  </a-col>
                </a-row>
                <a-form-item label="运行参数">
                  <a-input v-model:value="form.dockerRunArgs" placeholder="-d --restart unless-stopped -p 8080:8080" />
                </a-form-item>
              </template>
            </a-form>
          </a-collapse-panel>
        </a-collapse>
      </div>

      <div class="bpm-placeholder">
        <div class="bpm-title">BPM 流程编排（下一阶段）</div>
        <div class="bpm-desc">
          已移除假节点预览和假阶段编辑。下一步替换为真实 BPM 编排器（开始、任务、审批、智能体、网关、结束节点），支持拖拽连线、属性编辑、版本化。
        </div>
      </div>
    </div>
  </a-drawer>
</template>

<script setup>
import { computed, ref, watch, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { CheckCircleOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  open: Boolean,
  token: String,
  pipelineId: String,
  pipelineName: String,
  pipeline: Object,
  systemId: String,
})

const emit = defineEmits(['close', 'save'])
const openSections = ref(['source', 'build', 'deploy'])

const form = ref({
  releaseUnitId: '',
  repositoryType: 'git',
  autoMerge: true,
  autoTag: true,
  displayOrder: 0,
  repoUrl: '',
  branch: '',
  gitUsername: '',
  gitToken: '',
  projectPath: '',
  buildType: 'maven',
  mavenCommand: 'mvn clean package -DskipTests',
  gradleCommand: './gradlew clean bootJar',
  npmCommand: 'npm run build',
  deployType: 'docker',
  dockerImage: '',
  dockerContainer: '',
  dockerRunArgs: '',
  gitCredentialKey: 'github',
  gitUsernameField: 'username',
  gitTokenField: 'token',
})

const releaseUnits = ref([])
const globalVariables = ref([])
const globalVariableFieldMap = ref({})
const loadingGlobalVariables = ref(false)

const missingFields = computed(() => {
  const missing = []
  if (!form.value.repoUrl && !form.value.projectPath) missing.push('仓库地址或项目路径')
  if (form.value.repoUrl && !form.value.branch) missing.push('分支')
  if (form.value.repoUrl && !form.value.gitCredentialKey) missing.push('凭证命名空间')
  if (form.value.repoUrl && !form.value.gitUsernameField) missing.push('用户名字段')
  if (form.value.repoUrl && !form.value.gitTokenField) missing.push('令牌字段')
  if (form.value.deployType === 'docker') {
    if (!form.value.dockerImage) missing.push('镜像名称')
    if (!form.value.dockerContainer) missing.push('容器名称')
  }
  return missing
})

const completenessTotal = computed(() => (form.value.deployType === 'docker' ? 5 : 3))
const completenessPercent = computed(() => {
  const filled = completenessTotal.value - missingFields.value.length
  return Math.max(0, Math.round((filled / completenessTotal.value) * 100))
})

const healthClass = computed(() => {
  const pct = completenessPercent.value
  if (pct === 100) return 'health-good'
  if (pct >= 40) return 'health-warn'
  return 'health-bad'
})

const buildSummary = computed(() => {
  const t = form.value.buildType
  if (t === 'maven') return `Maven · ${form.value.mavenCommand || '未填写命令'}`
  if (t === 'gradle') return `Gradle · ${form.value.gradleCommand || '未填写命令'}`
  if (t === 'npm') return `NPM · ${form.value.npmCommand || '未填写命令'}`
  return '不构建'
})

const applyTemplate = (type) => {
  if (type === 'java') {
    form.value.buildType = 'maven'
    form.value.mavenCommand = 'mvn clean package -DskipTests'
    form.value.deployType = 'docker'
    if (!form.value.dockerRunArgs) form.value.dockerRunArgs = '-d --restart unless-stopped -p 8080:8080'
  } else if (type === 'node') {
    form.value.buildType = 'npm'
    form.value.npmCommand = 'npm run build'
    form.value.deployType = 'docker'
    if (!form.value.dockerRunArgs) form.value.dockerRunArgs = '-d --restart unless-stopped -p 3000:3000 -e NODE_ENV=production'
  } else if (type === 'static') {
    form.value.buildType = 'npm'
    form.value.npmCommand = 'npm run build'
    form.value.deployType = 'docker'
    if (!form.value.dockerRunArgs) form.value.dockerRunArgs = '-d --restart unless-stopped -p 80:80'
  }
  message.success('模板已应用')
}

watch(
  () => props.pipeline,
  (newPipeline) => {
    if (!newPipeline) return
    form.value = {
      ...form.value,
      ...(newPipeline.config || {}),
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

const loadGlobalVariables = async () => {
  if (!props.token) return
  try {
    loadingGlobalVariables.value = true
    const res = await fetch('/api/global-vars', {
      headers: { Authorization: `Bearer ${props.token}` },
    })
    if (!res.ok) return
    const data = await res.json()
    globalVariables.value = data.data || []
    
    // Build fieldMap: namespace -> [field1, field2, ...]
    const fieldMap = {}
    globalVariables.value.forEach((gv) => {
      try {
        const val = JSON.parse(gv.value || '{}')
        fieldMap[gv.key] = Object.keys(val)
      } catch {
        fieldMap[gv.key] = []
      }
    })
    globalVariableFieldMap.value = fieldMap
  } catch (err) {
    console.error('Failed to load global variables:', err)
  } finally {
    loadingGlobalVariables.value = false
  }
}

const getGlobalVariableOptions = (namespace) => {
  const fields = globalVariableFieldMap.value[namespace] || []
  return fields.map((f) => ({
    label: f,
    value: f,
  }))
}

const updateOpen = (newVal) => {
  if (!newVal) emit('close')
  if (newVal) loadGlobalVariables()
}

const handleSave = async () => {
  if (!form.value.repoUrl && !form.value.projectPath) {
    message.warning('请至少填写仓库地址或项目路径')
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
        git_username: form.value.gitUsername,
        git_token: form.value.gitToken,
        project_path: form.value.projectPath,
        build_type: form.value.buildType,
        maven_command: form.value.mavenCommand,
        gradle_command: form.value.gradleCommand,
        npm_command: form.value.npmCommand,
        deploy_type: form.value.deployType,
        docker_image: form.value.dockerImage,
        docker_container: form.value.dockerContainer,
        docker_run_args: form.value.dockerRunArgs,
        git_credential_key: form.value.gitCredentialKey,
        git_username_field: form.value.gitUsernameField,
        git_token_field: form.value.gitTokenField,
        main_stages: [],
        env_stages: [],
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
.drawer-header {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.drawer-title {
  font-size: 18px;
  font-weight: 700;
  color: #1f2f46;
}

.templates-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}

.templates-label {
  font-size: 12px;
  color: #8c9bb5;
}

.tpl-btn {
  border: 1px solid #d0daea;
  background: #f4f7fd;
  color: #4a5d80;
  border-radius: 6px;
  padding: 2px 10px;
  font-size: 12px;
  cursor: pointer;
}

.header-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.health-bar-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 20px;
  background: #f7f9fc;
  border-bottom: 1px solid #e8edf5;
}

.health-bar-track {
  flex: 0 0 160px;
  height: 6px;
  background: #e0e8f5;
  border-radius: 6px;
  overflow: hidden;
}

.health-bar-fill {
  height: 100%;
}

.health-bar-fill.health-good { background: #52c41a; }
.health-bar-fill.health-warn { background: #faad14; }
.health-bar-fill.health-bad { background: #ff4d4f; }

.health-bar-text { font-size: 13px; }
.health-bar-text.health-good { color: #389e0d; }
.health-bar-text.health-warn { color: #d48806; }
.health-bar-text.health-bad { color: #cf1322; }

.drawer-body {
  display: grid;
  grid-template-columns: 1fr 320px;
  flex: 1;
  min-height: 0;
}

.config-form {
  overflow-y: auto;
  padding: 16px 16px 16px 20px;
  border-right: 1px solid #e8edf5;
}

.config-collapse :deep(.ant-collapse-item) {
  border: 1px solid #e0e8f5;
  border-radius: 12px !important;
  margin-bottom: 10px;
  overflow: hidden;
  background: #fff;
}

.config-collapse :deep(.ant-collapse-header) {
  padding: 12px 16px !important;
  background: #fafbff;
}

.config-collapse :deep(.ant-collapse-content-box) {
  padding: 16px 16px 8px !important;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.panel-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: #1a5fd0;
  color: #fff;
  border-radius: 50%;
  font-size: 12px;
}

.panel-title {
  font-weight: 600;
  font-size: 15px;
  color: #1f2f46;
}

.panel-hint {
  font-size: 12px;
  color: #8c9bb5;
  max-width: 280px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.panel-hint-missing {
  color: #faad14;
}

.bpm-placeholder {
  padding: 16px;
  background: #f8fafc;
}

.bpm-title {
  font-size: 14px;
  font-weight: 700;
  color: #1f2f46;
  margin-bottom: 8px;
}

.bpm-desc {
  font-size: 13px;
  color: #6a7892;
  line-height: 1.6;
}

@media (max-width: 1100px) {
  .drawer-body {
    grid-template-columns: 1fr;
  }

  .config-form {
    border-right: none;
  }
}
</style>
