<template>
  <div class="bpm-page">
    <div class="bpm-topbar">
      <a-button @click="goBack">返回</a-button>
      <div class="title">阶段编排 · {{ pipelineName || pipelineId }}</div>
      <div class="topbar-actions">
        <a-radio-group v-model:value="pipelineTriggerMode" size="small">
          <a-radio-button value="auto">自动触发</a-radio-button>
          <a-radio-button value="manual">手动触发</a-radio-button>
        </a-radio-group>
        <a-button @click="openTemplateModal">添加新阶段</a-button>
        <a-button @click="loadDefinition">刷新</a-button>
        <a-button type="primary" @click="saveDefinition">保存流程</a-button>
      </div>
    </div>

    <section class="stage-wrap">
      <div class="stage-head">
        <div>拖拽可调整顺序，也可直接改序号</div>
        <a-button @click="normalizeOrder">重排序号</a-button>
      </div>

      <a-empty v-if="stages.length === 0" description="暂无阶段，点击 添加新阶段" />

      <div v-else class="stage-list">
        <div
          v-for="(stage, idx) in stages"
          :key="stage.id"
          class="stage-item"
          :class="{ selected: selectedStageId === stage.id }"
          draggable="true"
          @dragstart="onDragStart(idx)"
          @dragover.prevent
          @drop="onDrop(idx)"
          @click="selectStage(stage.id)"
        >
          <div class="drag-handle">≡</div>
          <div class="stage-order">
            <a-input-number
              v-model:value="stage.order"
              :min="1"
              :controls="false"
              @change="applyManualOrder(stage.id)"
            />
          </div>
          <div class="stage-main">
            <div class="name">{{ stage.name }}</div>
            <div class="sub">{{ stage.templateCategory || '自定义阶段' }} / {{ stage.templateSubCategory || '通用' }}</div>
          </div>
          <div class="stage-mode">
            <a-tag :color="stage.runMode === 'parallel' ? 'orange' : 'blue'">
              {{ stage.runMode === 'parallel' ? '并行' : '串行' }}
            </a-tag>
            <a-tag v-if="stage.runMode === 'parallel' && stage.parallelGroup" color="purple">
              并行组 {{ stage.parallelGroup }}
            </a-tag>
            <a-tag :color="stage.triggerMode === 'manual' ? 'red' : 'green'">
              {{ stage.triggerMode === 'manual' ? '手动触发' : '自动触发' }}
            </a-tag>
          </div>
          <div class="stage-actions">
            <a-button type="link" @click.stop="selectStage(stage.id)">配置</a-button>
            <a-button type="link" danger @click.stop="removeStage(stage.id)">删除</a-button>
          </div>
        </div>
      </div>
    </section>

    <a-drawer v-model:open="stageDrawerOpen" width="520" title="阶段配置" placement="right" destroy-on-close>
      <div v-if="selectedStage" class="drawer-sections">
        <div class="sec-title">任务信息</div>
        <a-form layout="vertical">
          <a-form-item label="任务名称">
            <a-input v-model:value="selectedStage.name" />
          </a-form-item>
          <a-form-item label="任务动作">
            <a-select v-model:value="selectedStage.action">
              <a-select-option value="source">source</a-select-option>
              <a-select-option value="build">build</a-select-option>
              <a-select-option value="deploy">deploy</a-select-option>
              <a-select-option value="custom">custom</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="任务类型">
            <a-input v-model:value="selectedStage.taskType" placeholder="如 Java代码扫描" />
          </a-form-item>
          <a-row :gutter="8">
            <a-col :span="12">
              <a-form-item label="执行模式">
                <a-select v-model:value="selectedStage.runMode">
                  <a-select-option value="serial">串行</a-select-option>
                  <a-select-option value="parallel">并行</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="触发模式">
                <a-select v-model:value="selectedStage.triggerMode">
                  <a-select-option value="auto">自动触发</a-select-option>
                  <a-select-option value="manual">手动触发</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item v-if="selectedStage.runMode === 'parallel'" label="并行组">
            <a-select
              v-model:value="selectedStage.parallelGroup"
              mode="tags"
              :max-tag-count="1"
              :token-separators="[',', ' ']"
              placeholder="选择或输入并行组，如 A / B / C"
              :options="parallelGroupOptions"
            />
          </a-form-item>
        </a-form>

        <a-divider />

        <template v-if="selectedStage.action === 'source'">
          <div class="sec-title">源码参数</div>
          <a-form layout="vertical">
            <a-form-item label="仓库地址">
              <a-input v-model:value="selectedStage.presetFields.repoUrl" placeholder="https://github.com/org/repo.git" />
            </a-form-item>
            <a-form-item label="分支">
              <a-input v-model:value="selectedStage.presetFields.branch" placeholder="main" />
            </a-form-item>
            <a-form-item label="认证方式">
              <a-radio-group v-model:value="selectedStage.presetFields.authType" button-style="solid">
                <a-radio-button value="none">无需认证</a-radio-button>
                <a-radio-button value="token">令牌认证</a-radio-button>
              </a-radio-group>
            </a-form-item>
            <a-form-item v-if="selectedStage.presetFields.authType === 'token'" label="凭据命名空间">
              <a-select
                v-model:value="selectedStage.presetFields.gitCredentialKey"
                show-search
                placeholder="选择如 github"
                :options="globalVariableNamespaceOptions"
                :filter-option="filterVariableOption"
                @change="onCredentialNamespaceChange(selectedStage)"
              />
            </a-form-item>
            <a-form-item v-if="selectedStage.presetFields.authType === 'token'" label="用户名字段">
              <a-select
                v-model:value="selectedStage.presetFields.gitUsernameField"
                show-search
                placeholder="选择如 username"
                :options="fieldOptionsForStage(selectedStage)"
                :filter-option="filterVariableOption"
              />
            </a-form-item>
            <a-form-item v-if="selectedStage.presetFields.authType === 'token'" label="令牌字段">
              <a-select
                v-model:value="selectedStage.presetFields.gitTokenField"
                show-search
                placeholder="选择如 token"
                :options="fieldOptionsForStage(selectedStage)"
                :filter-option="filterVariableOption"
              />
            </a-form-item>
          </a-form>
        </template>

        <template v-else>
          <div class="sec-title">构建环境</div>
          <a-form layout="vertical">
            <a-form-item label="构建集群">
              <a-input v-model:value="selectedStage.presetFields.buildCluster" placeholder="如 云效默认构建集群" />
            </a-form-item>
            <a-form-item label="构建节点">
              <a-input v-model:value="selectedStage.presetFields.buildNode" placeholder="如 Linux/amd64" />
            </a-form-item>
            <a-form-item label="构建环境">
              <a-select v-model:value="selectedStage.presetFields.buildEnvironment">
                <a-select-option value="default">默认环境</a-select-option>
                <a-select-option value="container">指定容器环境</a-select-option>
              </a-select>
            </a-form-item>
            <a-form-item label="容器镜像地址">
              <a-input v-model:value="selectedStage.presetFields.containerImage" />
            </a-form-item>
          </a-form>

          <a-divider />

          <div class="sec-head">
            <div class="sec-title">任务步骤</div>
            <a-button type="link" @click="addStep">+ 添加步骤</a-button>
          </div>
          <div v-for="(step, i) in selectedStage.steps" :key="`${selectedStage.id}-step-${i}`" class="step-card">
            <a-input v-model:value="step.name" placeholder="步骤名称" style="margin-bottom: 8px" />
            <a-textarea v-model:value="step.command" :rows="3" placeholder="执行命令" />
            <div class="step-actions">
              <a-button type="link" danger @click="removeStep(i)">删除步骤</a-button>
            </div>
          </div>

          <a-divider />

          <div class="sec-title">高级设置</div>
          <a-form layout="vertical">
            <a-row :gutter="8">
              <a-col :span="12">
                <a-form-item label="构建环境规格">
                  <a-input v-model:value="selectedStage.advancedSettings.buildSpec" placeholder="DEFAULT" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="超时时间(分钟)">
                  <a-input-number v-model:value="selectedStage.advancedSettings.timeoutMinutes" :min="1" style="width: 100%" />
                </a-form-item>
              </a-col>
            </a-row>
            <a-space direction="vertical" style="width: 100%">
              <a-switch v-model:checked="selectedStage.advancedSettings.debugMode" checked-children="Debug开" un-checked-children="Debug关" />
              <a-switch v-model:checked="selectedStage.advancedSettings.deployTask" checked-children="部署任务" un-checked-children="普通任务" />
              <a-switch v-model:checked="selectedStage.advancedSettings.dockerDaemon" checked-children="Docker Daemon开" un-checked-children="Docker Daemon关" />
            </a-space>
          </a-form>

          <a-divider />

          <div class="sec-head">
            <div class="sec-title">任务插件</div>
            <a-button type="link" @click="addPlugin">+ 添加插件</a-button>
          </div>
          <div v-for="(plugin, i) in selectedStage.plugins" :key="`${selectedStage.id}-plugin-${i}`" class="plugin-card">
            <a-input v-model:value="plugin.plugin_name" placeholder="插件名称" style="margin-bottom: 8px" />
            <a-textarea v-model:value="plugin.configText" :rows="2" placeholder='插件配置(JSON，可选) 如 {"severity":"high"}' />
            <div class="step-actions">
              <a-button type="link" danger @click="removePlugin(i)">删除插件</a-button>
            </div>
          </div>

          <a-divider />

          <div class="sec-title">任务输出</div>
          <a-switch v-model:checked="selectedStage.presetFields.taskOutputArtifact" checked-children="输出制品" un-checked-children="不输出" />
        </template>
      </div>
    </a-drawer>

    <a-modal v-model:open="templateModalOpen" title="选择阶段模板" width="980px" :footer="null" destroy-on-close>
      <div class="template-modal">
        <div class="template-cats">
          <div
            v-for="cat in categoryList"
            :key="cat"
            class="cat-item"
            :class="{ active: cat === activeCategory }"
            @click="activeCategory = cat"
          >
            {{ cat }}
          </div>
        </div>
        <div class="template-list">
          <a-empty v-if="filteredTemplates.length === 0" description="该分类暂无模板" />
          <div v-else class="tpl-grid">
            <article v-for="tpl in filteredTemplates" :key="tpl.id" class="tpl-card" @click="createStageFromTemplate(tpl)">
              <div class="tpl-name">{{ tpl.name }}</div>
              <div class="tpl-tags">
                <a-tag color="blue">{{ tpl.category }}</a-tag>
                <a-tag color="geekblue">{{ tpl.sub_category || '通用' }}</a-tag>
              </div>
              <div class="tpl-desc">{{ tpl.description || '无描述' }}</div>
            </article>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

const route = useRoute()
const globalVariables = ref([])
const router = useRouter()

const pipelineId = computed(() => String(route.params.id || ''))
const pipelineName = computed(() => String(route.query.name || ''))
const token = computed(() => localStorage.getItem('token') || '')

const stages = ref([])
const selectedStageId = ref('')
const stageDrawerOpen = ref(false)
const dragIndex = ref(-1)

const pipelineTriggerMode = ref('auto')

const templateModalOpen = ref(false)
const templates = ref([])
const activeCategory = ref('')
const parallelGroupOptions = ref([
  { label: 'A组', value: 'A' },
  { label: 'B组', value: 'B' },
  { label: 'C组', value: 'C' },
])

const defaultTemplates = () => [
  {
    id: 'builtin-git-source',
    name: 'Git 拉取',
    category: '源码',
    sub_category: 'Git拉取',
    description: '从 Git 仓库拉取代码，默认分支为 main，可按阶段调整',
    preset_fields: {
      repoUrl: '',
      branch: 'main',
      authType: 'none',
      gitCredentialKey: 'github',
      gitUsernameField: 'username',
      gitTokenField: 'token',
    },
    advanced_settings: {},
    steps: [],
    plugins: [],
  },
  {
    id: 'builtin-java-security-scan',
    name: 'Java 安全扫描',
    category: '代码扫描',
    sub_category: 'Java代码扫描',
    description: 'Java 构建 + SpotBugs 安全扫描',
    preset_fields: {
      buildCluster: '云效默认构建集群',
      buildNode: 'Linux/amd64',
      buildEnvironment: 'container',
      containerImage: 'build-steps/alinux3',
      taskOutputArtifact: true,
    },
    advanced_settings: {
      buildSpec: 'DEFAULT',
      timeoutMinutes: 240,
      debugMode: false,
      deployTask: false,
      dockerDaemon: false,
    },
    steps: [
      { name: '配置 MavenSettings 文件', command: 'echo use maven settings' },
      { name: '安装 Java', command: 'echo install java' },
      { name: '执行 Java 构建命令', command: 'mvn -B clean package -Dmaven.test.skip=true' },
      { name: 'Java 安全扫描 Spotbugs', command: 'mvn com.github.spotbugs:spotbugs-maven-plugin:check' },
    ],
    plugins: [{ plugin_name: 'SpotBugs', plugin_config: { severity: 'medium' } }],
  },
  {
    id: 'builtin-java-build',
    name: 'Java 构建',
    category: '构建',
    sub_category: 'Java构建',
    description: 'Maven 构建 Java 项目',
    preset_fields: {
      buildEnvironment: 'container',
      containerImage: 'maven:3.9-eclipse-temurin-17',
    },
    advanced_settings: {
      buildSpec: 'DEFAULT',
      timeoutMinutes: 240,
      debugMode: false,
      deployTask: false,
      dockerDaemon: false,
    },
    steps: [{ name: '执行构建', command: 'mvn -B clean package -Dmaven.test.skip=true' }],
    plugins: [],
  },
  {
    id: 'builtin-test-build',
    name: '测试并构建',
    category: '测试构建',
    sub_category: 'Java测试构建',
    description: '先测试后构建',
    preset_fields: { buildEnvironment: 'container', containerImage: 'maven:3.9-eclipse-temurin-17' },
    advanced_settings: {
      buildSpec: 'DEFAULT',
      timeoutMinutes: 300,
      debugMode: false,
      deployTask: false,
      dockerDaemon: false,
    },
    steps: [
      { name: '执行单元测试', command: 'mvn test' },
      { name: '执行构建', command: 'mvn -B clean package -Dmaven.test.skip=true' },
    ],
    plugins: [],
  },
  {
    id: 'builtin-empty',
    name: '空模板',
    category: '空模板',
    sub_category: '空模板',
    description: '无预设步骤，完全自定义',
    preset_fields: { buildEnvironment: 'default', taskOutputArtifact: false },
    advanced_settings: {
      buildSpec: 'DEFAULT',
      timeoutMinutes: 120,
      debugMode: false,
      deployTask: false,
      dockerDaemon: false,
    },
    steps: [],
    plugins: [],
  },
]
const globalVariableNamespaceOptions = computed(() => globalVariables.value.map((item) => ({
  label: `${item.key}${item.description ? ` | ${item.description}` : ''}${item.is_secret ? ' | 密文' : ''}`,
  value: item.key,
})))

const globalVariableFieldMap = computed(() => {
  const out = {}
  globalVariables.value.forEach((item) => {
    out[item.key] = Array.isArray(item.fields)
      ? item.fields.map((f) => (typeof f === 'string' ? f : f?.name)).filter(Boolean)
      : []
  })
  return out
})

const selectedStage = computed(() => stages.value.find((s) => s.id === selectedStageId.value) || null)

const categoryList = computed(() => Array.from(new Set(templates.value.map((t) => t.category).filter(Boolean))))
const filteredTemplates = computed(() => {
  if (!activeCategory.value) return templates.value
  return templates.value.filter((t) => t.category === activeCategory.value)
})

const goBack = () => router.push('/workspace')
const genId = (prefix) => `${prefix}_${Date.now()}_${Math.random().toString(16).slice(2, 7)}`

const ensureSourcePresetFields = (stage) => {
  if (!stage) return
  stage.presetFields = stage.presetFields || {}
  if ((stage.presetFields.branch || '').trim() === '') {
    stage.presetFields.branch = 'main'
  }
  if (!stage.presetFields.authType) {
    stage.presetFields.authType = 'none'
  }
  if (!Object.prototype.hasOwnProperty.call(stage.presetFields, 'gitCredentialKey')) {
    stage.presetFields.gitCredentialKey = 'github'
  }
  if (!Object.prototype.hasOwnProperty.call(stage.presetFields, 'gitUsernameField')) {
    stage.presetFields.gitUsernameField = 'username'
  }
  if (!Object.prototype.hasOwnProperty.call(stage.presetFields, 'gitTokenField')) {
    stage.presetFields.gitTokenField = 'token'
  }
}

const fieldOptionsForStage = (stage) => {
  const namespaceKey = String(stage?.presetFields?.gitCredentialKey || '').trim()
  const fields = globalVariableFieldMap.value[namespaceKey] || []
  return fields.map((f) => ({ label: f, value: f }))
}

const onCredentialNamespaceChange = (stage) => {
  const options = fieldOptionsForStage(stage)
  const values = options.map((x) => x.value)
  if (!values.includes(stage.presetFields.gitUsernameField)) {
    stage.presetFields.gitUsernameField = values.includes('username') ? 'username' : (values[0] || '')
  }
  if (!values.includes(stage.presetFields.gitTokenField)) {
    stage.presetFields.gitTokenField = values.includes('token') ? 'token' : (values[0] || '')
  }
}

const filterVariableOption = (input, option) => String(option?.value || '').toLowerCase().includes(String(input || '').toLowerCase())

const fetchJson = async (url, options = {}) => {
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token.value}`,
      ...(options.headers || {}),
    },
  })
  const text = await res.text()
  let data = {}
  try {
    data = text ? JSON.parse(text) : {}
  } catch {
    data = {}
  }
  if (!res.ok) {
    throw new Error(data?.error || data?.message || `HTTP ${res.status}`)
  }
  return data
}

const loadGlobalVariables = async () => {
  if (!token.value) {
    globalVariables.value = []
    return
  }
  try {
    const data = await fetchJson('/api/global-vars', { method: 'GET' })
    globalVariables.value = Array.isArray(data.items) ? data.items : []
  } catch {
    globalVariables.value = []
  }
}

const normalizeOrder = () => {
  stages.value.forEach((s, i) => {
    s.order = i + 1
  })
}

const applyManualOrder = (stageId) => {
  const target = stages.value.find((s) => s.id === stageId)
  if (!target) return
  target.order = Number(target.order) || 1
  stages.value.sort((a, b) => a.order - b.order)
  normalizeOrder()
}

const onDragStart = (idx) => {
  dragIndex.value = idx
}

const onDrop = (idx) => {
  if (dragIndex.value < 0 || dragIndex.value === idx) return
  const moved = stages.value.splice(dragIndex.value, 1)[0]
  stages.value.splice(idx, 0, moved)
  dragIndex.value = -1
  normalizeOrder()
}

const selectStage = (id) => {
  selectedStageId.value = id
  ensureSourcePresetFields(selectedStage.value)
  if (selectedStage.value?.action === 'source') {
    onCredentialNamespaceChange(selectedStage.value)
  }
  stageDrawerOpen.value = true
}

const removeStage = (id) => {
  stages.value = stages.value.filter((s) => s.id !== id)
  if (selectedStageId.value === id) {
    selectedStageId.value = ''
    stageDrawerOpen.value = false
  }
  normalizeOrder()
}

const addStep = () => {
  if (!selectedStage.value) return
  selectedStage.value.steps = selectedStage.value.steps || []
  selectedStage.value.steps.push({ name: '新步骤', command: '' })
}

const removeStep = (idx) => {
  if (!selectedStage.value) return
  selectedStage.value.steps.splice(idx, 1)
}

const addPlugin = () => {
  if (!selectedStage.value) return
  selectedStage.value.plugins = selectedStage.value.plugins || []
  selectedStage.value.plugins.push({ plugin_name: '新插件', configText: '' })
}

const removePlugin = (idx) => {
  if (!selectedStage.value) return
  selectedStage.value.plugins.splice(idx, 1)
}

const mapTemplateAction = (tpl) => {
  const text = `${tpl.category || ''} ${tpl.sub_category || ''} ${tpl.name || ''}`.toLowerCase()
  if (text.includes('部署') || text.includes('deploy')) return 'deploy'
  if (text.includes('构建') || text.includes('build')) return 'build'
  if (text.includes('源码') || text.includes('git') || text.includes('source')) return 'source'
  return 'custom'
}

const openTemplateModal = async () => {
  await loadTemplates()
  templateModalOpen.value = true
}

const createStageFromTemplate = (tpl) => {
  const stage = {
    id: genId('task'),
    type: 'task',
    name: tpl.name,
    order: stages.value.length + 1,
    action: mapTemplateAction(tpl),
    taskType: tpl.sub_category || tpl.category || 'custom',
    runMode: 'serial',
    triggerMode: 'auto',
    parallelGroup: '',
    templateId: tpl.id,
    templateCategory: tpl.category || '',
    templateSubCategory: tpl.sub_category || '',
    presetFields: { ...(tpl.preset_fields || {}) },
    advancedSettings: { ...(tpl.advanced_settings || {}) },
    steps: (tpl.steps || []).map((s) => ({ name: s.name || '', command: s.command || '' })),
    plugins: (tpl.plugins || []).map((p) => ({
      plugin_name: p.plugin_name || p.PluginName || '',
      configText: typeof p.plugin_config === 'string' ? p.plugin_config : JSON.stringify(p.plugin_config || {}),
    })),
  }

  if (stage.action === 'source') {
    ensureSourcePresetFields(stage)
  }

  if (stage.steps.length === 0) {
    stage.steps = [{ name: '执行命令', command: '' }]
  }

  if (stage.action === 'source') {
    stage.steps = []
  }

  stages.value.push(stage)
  selectStage(stage.id)
  templateModalOpen.value = false
  message.success(`已添加阶段: ${tpl.name}`)
}

const normalizeTemplate = (item) => ({
  id: item.id || item.ID,
  name: item.name || item.Name,
  category: item.category || item.Category,
  sub_category: item.sub_category || item.SubCategory,
  description: item.description || item.Description,
  preset_fields: item.preset_fields || item.PresetFields || {},
  advanced_settings: item.advanced_settings || item.AdvancedSettings || {},
  steps: item.steps || item.Steps || [],
  plugins: item.plugins || item.Plugins || [],
})

const loadTemplates = async () => {
  templates.value = defaultTemplates()

  try {
    const data = await fetchJson('/api/task-templates', { method: 'GET' })
    const list = Array.isArray(data.data) ? data.data : []
    if (list.length > 0) {
      templates.value = list.map(normalizeTemplate)
    }
  } catch {
    // Keep built-in defaults when backend route is unavailable.
  }

  if (!activeCategory.value && categoryList.value.length > 0) {
    activeCategory.value = categoryList.value[0]
  }
}

const saveDefinition = async () => {
  const payload = {
    pipeline_id: pipelineId.value,
    definition: {
      meta: {
        pipelineTriggerMode: pipelineTriggerMode.value,
      },
      nodes: stages.value.map((s) => {
        const plugins = (s.plugins || []).map((p) => {
          let config = {}
          try {
            config = p.configText ? JSON.parse(p.configText) : {}
          } catch {
            config = { raw: p.configText || '' }
          }
          return {
            plugin_name: p.plugin_name || '',
            plugin_config: config,
          }
        })

        return {
          id: s.id,
          type: 'task',
          name: s.name,
          order: s.order,
          action: s.action,
          taskType: s.taskType,
          runMode: s.runMode,
          triggerMode: s.triggerMode,
          parallelGroup: s.parallelGroup || '',
          presetFields: s.presetFields || {},
          advancedSettings: s.advancedSettings || {},
          steps: s.steps || [],
          plugins,
        }
      }),
      edges: [],
    },
  }

  try {
    await fetchJson(`/api/pipelines/${pipelineId.value}/bpm`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
    message.success('流程已保存')
  } catch (err) {
    message.error(`保存失败: ${err.message}`)
  }
}

const loadDefinition = async () => {
  try {
    const data = await fetchJson(`/api/pipelines/${pipelineId.value}/bpm`, { method: 'GET' })
    const def = data.definition || {}
    pipelineTriggerMode.value = def?.meta?.pipelineTriggerMode || 'auto'

    if (Array.isArray(def.stages)) {
      stages.value = def.stages
    } else {
      const nodes = Array.isArray(def.nodes) ? def.nodes : []
      stages.value = nodes
        .filter((n) => String(n.type || '').toLowerCase() === 'task')
        .map((n, idx) => ({
          id: n.id || genId('task'),
          type: 'task',
          name: n.name || `阶段${idx + 1}`,
          order: Number(n.order) || idx + 1,
          action: n.action || 'custom',
          taskType: n.taskType || '',
          runMode: n.runMode || 'serial',
          triggerMode: n.triggerMode || 'auto',
          parallelGroup: n.parallelGroup || '',
          templateId: n.templateId || '',
          templateCategory: n.templateCategory || '',
          templateSubCategory: n.templateSubCategory || '',
          presetFields: n.presetFields || {},
          advancedSettings: n.advancedSettings || {},
          steps: Array.isArray(n.steps) ? n.steps : [],
          plugins: Array.isArray(n.plugins)
            ? n.plugins.map((p) => ({
              plugin_name: p.plugin_name || p.PluginName || '',
              configText: JSON.stringify(p.plugin_config || p.PluginConfig || {}),
            }))
            : [],
        }))
    }

    stages.value.forEach((stage) => {
      if (stage.action === 'source') {
        ensureSourcePresetFields(stage)
        stage.steps = []
      }
    })

    stages.value.sort((a, b) => (a.order || 0) - (b.order || 0))
    normalizeOrder()
  } catch (err) {
    message.error(`加载流程失败: ${err.message}`)
  }
}

onMounted(async () => {
  await loadGlobalVariables()
  await loadDefinition()
})
</script>

<style scoped>
.bpm-page {
  min-height: 100vh;
  padding: 16px;
  background: linear-gradient(180deg, #f4f7fb 0%, #f9fbff 100%);
}

.bpm-topbar {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.title {
  font-size: 18px;
  font-weight: 700;
  color: #243a57;
  text-align: center;
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}

.stage-wrap {
  border: 1px solid #dce5f4;
  border-radius: 10px;
  background: #fff;
  padding: 12px;
}

.stage-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  color: #4a6285;
}

.stage-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.stage-item {
  display: grid;
  grid-template-columns: 30px 80px 1fr auto auto;
  gap: 10px;
  align-items: center;
  border: 1px solid #d9e4f4;
  border-radius: 10px;
  padding: 10px;
  background: #fbfdff;
  cursor: pointer;
}

.stage-item.selected {
  border-color: #1677ff;
  box-shadow: 0 0 0 2px rgba(22, 119, 255, 0.2);
}

.drag-handle {
  color: #8ba0bf;
  font-size: 16px;
  user-select: none;
}

.stage-main .name {
  font-size: 15px;
  font-weight: 700;
  color: #1f3552;
}

.stage-main .sub {
  font-size: 12px;
  color: #6d84a4;
}

.stage-mode {
  display: flex;
  gap: 6px;
}

.step-card,
.plugin-card {
  border: 1px solid #e0e8f4;
  border-radius: 8px;
  padding: 8px;
  margin-bottom: 8px;
  background: #fafcff;
}

.step-actions {
  margin-top: 6px;
  text-align: right;
}

.drawer-sections {
  padding-bottom: 16px;
}

.sec-title {
  font-size: 14px;
  font-weight: 700;
  color: #29425f;
  margin-bottom: 8px;
}

.sec-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.template-modal {
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 14px;
  min-height: 420px;
}

.template-cats {
  border-right: 1px solid #e3e8f3;
  padding-right: 10px;
}

.cat-item {
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  color: #355170;
}

.cat-item:hover {
  background: #eef4ff;
}

.cat-item.active {
  background: #e6f0ff;
  color: #124188;
  font-weight: 600;
}

.tpl-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 10px;
}

.tpl-card {
  border: 1px solid #dbe5f3;
  border-radius: 10px;
  padding: 10px;
  background: #fff;
  cursor: pointer;
}

.tpl-card:hover {
  border-color: #1677ff;
  box-shadow: 0 8px 16px rgba(22, 119, 255, 0.12);
}

.tpl-name {
  font-size: 15px;
  font-weight: 700;
  color: #203855;
}

.tpl-tags {
  margin-top: 6px;
}

.tpl-desc {
  margin-top: 6px;
  font-size: 12px;
  color: #6f87a8;
}

@media (max-width: 1200px) {
  .bpm-topbar {
    grid-template-columns: 1fr;
  }

  .title {
    text-align: left;
  }

  .topbar-actions {
    flex-wrap: wrap;
    white-space: normal;
  }

  .stage-item {
    grid-template-columns: 24px 72px 1fr;
  }

  .stage-mode,
  .stage-actions {
    grid-column: 1 / span 3;
  }

  .template-modal {
    grid-template-columns: 1fr;
  }

  .template-cats {
    border-right: 0;
    border-bottom: 1px solid #e3e8f3;
    padding-right: 0;
    padding-bottom: 8px;
  }
}
</style>
