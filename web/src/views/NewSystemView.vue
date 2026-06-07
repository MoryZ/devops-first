<template>
  <PageCard title="新建系统" subtitle="创建一个新的 DevOps 子系统" icon="+" bg="gradient">
    <template #header>
      <div class="page-header-content">
        <div class="header-left">
          <div class="breadcrumb">
            <span class="breadcrumb-item" @click="goBack">工作台</span>
            <span class="breadcrumb-sep">/</span>
            <span class="breadcrumb-current">新建系统</span>
          </div>
        </div>
      </div>
    </template>

    <a-card :bordered="false" class="form-card">
      <a-form
        layout="vertical"
        :model="form"
        :rules="rules"
        ref="formRef"
        class="system-form"
        @finish="handleSubmit"
      >
        <div class="form-section">
          <div class="section-title">
            <span class="section-icon">◈</span>
            基础信息
          </div>

          <a-form-item label="系统名称" name="name">
            <a-input
              v-model:value="form.name"
              placeholder="输入系统名称，例如：order-service"
              size="large"
              class="form-input"
            >
              <template #prefix>
                <FolderOutlined />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item label="系统描述">
            <a-textarea
              v-model:value="form.description"
              placeholder="描述该系统的职责、团队或业务域（可选）"
              :rows="4"
              class="form-textarea"
            />
          </a-form-item>

          <a-form-item label="状态">
            <a-radio-group v-model:value="form.status" button-style="solid" class="status-radio-group">
              <a-radio-button value="active">活跃</a-radio-button>
              <a-radio-button value="planning">规划中</a-radio-button>
              <a-radio-button value="archived">已归档</a-radio-button>
            </a-radio-group>
          </a-form-item>
        </div>

        <div class="form-actions">
          <a-button class="back-btn" @click="goBack">
            <ArrowLeftOutlined />
            返回
          </a-button>
          <a-button type="primary" html-type="submit" :loading="submitting" class="submit-btn">
            <PlusOutlined />
            创建系统
          </a-button>
        </div>
      </a-form>
    </a-card>

    <a-card v-if="recentSystems.length" :bordered="false" class="recent-card">
      <div class="recent-header">
        <span class="recent-icon">◈</span>
        <h3 class="recent-title">最近创建</h3>
      </div>
      <div class="recent-list">
        <a-card
          v-for="system in recentSystems"
          :key="system.id"
          :bordered="false"
          class="recent-item"
          @click="selectSystem(system)"
        >
          <div class="recent-item-name">{{ system.name }}</div>
          <div class="recent-item-desc">{{ system.description || '暂无描述' }}</div>
          <a-tag :color="statusColor[system.status]" class="recent-tag">{{ system.status }}</a-tag>
        </a-card>
      </div>
    </a-card>
  </PageCard>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { ArrowLeftOutlined, FolderOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { createSystem, listSystems } from '../api/systems'
import { useWorkspaceStore } from '../stores/workspace'
import PageCard from '../components/PageCard.vue'

const route = useRoute()
const router = useRouter()
const workspace = useWorkspaceStore()

const formRef = ref(null)
const submitting = ref(false)
const systems = ref([])

const form = reactive({
  name: '',
  description: '',
  status: 'active',
})

const rules = {
  name: [
    { required: true, message: '请输入系统名称', trigger: 'blur' },
    { min: 2, max: 64, message: '系统名称长度应在 2-64 个字符之间', trigger: 'blur' },
  ],
}

const statusColor = {
  active: 'green',
  planning: 'orange',
  archived: 'gray',
}

const recentSystems = ref([])

onMounted(async () => {
  await loadRecentSystems()
})

const loadRecentSystems = async () => {
  try {
    const data = await listSystems(workspace.token || localStorage.getItem('token') || '')
    systems.value = data.items || []
    recentSystems.value = systems.value.slice(0, 6)
  } catch (err) {
    console.warn('Failed to load systems:', err)
  }
}

const handleSubmit = async () => {
  try {
    const valid = await formRef.value?.validate?.()
    if (!valid) return
  } catch {
    return
  }

  submitting.value = true
  try {
    const created = await createSystem(workspace.token || localStorage.getItem('token') || '', {
      name: form.name.trim(),
      description: form.description.trim(),
      status: form.status,
    })

    message.success('系统创建成功')
    await loadRecentSystems()

    form.name = ''
    form.description = ''
    form.status = 'active'

    workspace.selectSystem(created.ID || created.id)
    router.replace('/workspace')
  } catch (err) {
    message.error('创建失败: ' + (err?.message || '未知错误'))
  } finally {
    submitting.value = false
  }
}

const selectSystem = (system) => {
  workspace.selectSystem(system.id)
  router.replace('/workspace')
}

const goBack = () => {
  router.replace('/workspace')
}
</script>

<style scoped>
.form-card {
  border-radius: 16px;
  border: 1px solid var(--core-border);
  background: var(--core-surface);
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.08);
  margin-bottom: 20px;
  animation: slideUp 0.4s ease-out;
}

.system-form {
  padding: 24px;
}

.form-section {
  margin-bottom: 24px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 700;
  color: var(--core-text);
  margin-bottom: 18px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--core-border);
}

.section-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border-radius: 7px;
  font-size: 11px;
  color: white;
  flex-shrink: 0;
}

.form-input,
.form-textarea {
  border-radius: var(--radius-md) !important;
  background: var(--core-surface) !important;
  border: 1px solid var(--core-border-accent) !important;
  transition: all var(--transition-fast) !important;
}

.form-input :deep(.ant-input),
.form-textarea :deep(.ant-input) {
  background: transparent !important;
  border: none !important;
  color: var(--core-text) !important;
  font-family: var(--font-display) !important;
}

.form-input :deep(.ant-input::placeholder),
.form-textarea :deep(.ant-input::placeholder) {
  color: var(--core-text-muted) !important;
}

.form-input :deep(.ant-input-prefix) {
  color: var(--core-text-secondary) !important;
  margin-right: 10px !important;
}

.form-input:hover,
.form-input:focus-within,
.form-textarea:hover,
.form-textarea:focus-within {
  border-color: var(--accent-primary) !important;
  box-shadow: 0 0 0 3px rgba(0, 212, 255, 0.12) !important;
}

.status-radio-group {
  width: 100%;
}

.status-radio-group :deep(.ant-radio-button-wrapper) {
  background: var(--core-surface) !important;
  border-color: var(--core-border-accent) !important;
  color: var(--core-text-secondary) !important;
  font-family: var(--font-display) !important;
  transition: all var(--transition-fast) !important;
}

.status-radio-group :deep(.ant-radio-button-wrapper:hover) {
  border-color: var(--accent-primary) !important;
  color: var(--accent-primary) !important;
}

.status-radio-group :deep(.ant-radio-button-wrapper-checked) {
  background: rgba(0, 212, 255, 0.12) !important;
  border-color: var(--accent-primary) !important;
  color: #084c61 !important;
  box-shadow: none !important;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid var(--core-border);
  margin-top: 8px;
}

.back-btn {
  height: 42px;
  padding: 0 24px;
  border-radius: var(--radius-md);
  border: 1px solid var(--core-border-accent);
  background: var(--core-surface);
  color: var(--core-text-secondary);
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: all var(--transition-fast);
  font-family: var(--font-display);
}

.back-btn:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
  background: var(--core-surface-hover);
}

.submit-btn {
  height: 42px;
  padding: 0 28px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-info));
  border: none;
  color: #ffffff;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 10px 24px rgba(0, 212, 255, 0.28);
  transition: all var(--transition-base);
  font-family: var(--font-display);
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 14px 28px rgba(0, 212, 255, 0.38);
}

.recent-card {
  border-radius: 16px;
  border: 1px solid var(--core-border);
  background: var(--core-surface);
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.08);
  animation: slideUp 0.4s ease-out 0.1s backwards;
}

.recent-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--core-border);
  margin-bottom: 16px;
}

.recent-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border-radius: 7px;
  font-size: 11px;
  color: white;
  flex-shrink: 0;
}

.recent-title {
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 700;
  color: var(--core-text);
  margin: 0;
  letter-spacing: 0.02em;
}

.recent-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 12px;
}

.recent-item {
  border-radius: var(--radius-md);
  border: 1px solid var(--core-border);
  background: var(--core-surface);
  cursor: pointer;
  transition: all var(--transition-base);
  padding: 14px;
}

.recent-item:hover {
  border-color: var(--accent-primary);
  background: var(--core-surface-hover);
  transform: translateY(-2px);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.1);
}

.recent-item-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--core-text);
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.recent-item-desc {
  font-size: 12px;
  color: var(--core-text-secondary);
  margin-bottom: 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.recent-tag {
  border-radius: 6px;
  font-size: 11px;
  font-weight: 500;
}

@media (max-width: 768px) {
  .system-form {
    padding: 16px;
  }

  .form-actions {
    flex-direction: column-reverse;
  }

  .back-btn,
  .submit-btn {
    width: 100%;
    justify-content: center;
  }

  .recent-list {
    grid-template-columns: 1fr;
  }
}
</style>
