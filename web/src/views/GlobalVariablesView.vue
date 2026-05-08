<template>
  <div class="global-vars-page">
    <div class="vars-header">
      <h2>全局变量管理</h2>
      <a-button type="primary" @click="openAddVariableModal">+ 新建变量</a-button>
    </div>

    <div class="vars-container">
      <a-empty v-if="globalVariables.length === 0" description="暂无全局变量" style="margin-top: 40px" />

      <div v-else class="var-list">
        <div v-for="item in globalVariables" :key="item.id" class="var-card">
          <div class="var-header">
            <div class="var-key-badge">{{ item.key }}</div>
            <div class="var-secret-badge" v-if="item.is_secret">密文</div>
          </div>
          <div class="var-body">
            <div class="var-desc" v-if="item.description">{{ item.description }}</div>
            <div v-if="(item.fields || []).length > 0" class="field-list">
              <div v-for="field in item.fields" :key="`${item.id}-${field.name}`" class="field-item">
                <span class="field-name">{{ field.name }}</span>
                <span class="field-value">{{ field.value_preview || '-' }}</span>
                <span class="field-badge" :class="{ secret: field.is_secret }">{{ field.is_secret ? '密文' : '明文' }}</span>
              </div>
            </div>
            <div v-else class="var-preview">字段: -</div>
          </div>
          <div class="var-footer">
            <a-button type="link" size="small" @click="editVariable(item)">编辑</a-button>
            <a-divider type="vertical" style="margin: 0 4px" />
            <a-popconfirm
              title="删除变量"
              description="确认删除此全局变量吗？"
              ok-text="确定"
              cancel-text="取消"
              @confirm="deleteVariable(item.id)"
            >
              <a-button type="link" danger size="small">删除</a-button>
            </a-popconfirm>
          </div>
        </div>
      </div>
    </div>

    <a-modal v-model:open="varModalOpen" :title="editingVar.id ? '编辑变量' : '新建变量'" @ok="saveVariable" destroy-on-close width="720px">
      <a-form :model="editingVar" autocomplete="off" layout="vertical">
        <a-form-item label="命名空间键" required>
          <a-input v-model:value="editingVar.key" placeholder="如 github" :disabled="!!editingVar.id" />
        </a-form-item>
        <a-form-item label="字段列表" required>
          <div class="edit-field-list">
            <div v-for="(field, idx) in editingVar.fields" :key="`edit-field-${idx}`" class="edit-field-row">
              <a-input v-model:value="field.name" placeholder="字段名，如 username" />
              <a-input v-model:value="field.value" placeholder="字段值" />
              <a-checkbox v-model:checked="field.is_secret">密文</a-checkbox>
              <a-button danger type="text" @click="removeEditField(idx)">删除</a-button>
            </div>
            <a-button type="dashed" block @click="addEditField">+ 添加字段</a-button>
          </div>
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model:value="editingVar.description" placeholder="变量说明" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'
import { deleteGlobalVar, listGlobalVars, saveGlobalVar } from '../api/globalVars'

const globalVariables = ref([])
const varModalOpen = ref(false)
const editingVar = ref({
  id: '',
  key: '',
  description: '',
  fields: [],
})

const token = computed(() => localStorage.getItem('token') || '')

const loadGlobalVariables = async () => {
  if (!token.value) {
    globalVariables.value = []
    return
  }
  try {
    const data = await listGlobalVars(token.value)
    globalVariables.value = Array.isArray(data.items) ? data.items : []
  } catch (err) {
    message.error(`加载变量失败: ${err.message}`)
  }
}

const openAddVariableModal = () => {
  editingVar.value = {
    id: '',
    key: '',
    description: '',
    fields: [
      { name: 'username', value: '', is_secret: false },
      { name: 'token', value: '', is_secret: true },
    ],
  }
  varModalOpen.value = true
}

const editVariable = (item) => {
  editingVar.value = {
    id: item.id,
    key: item.key,
    description: item.description || '',
    fields: Array.isArray(item.fields)
      ? item.fields.map((f) => ({
        name: f.name || '',
        value: '',
        is_secret: !!f.is_secret,
      }))
      : [],
  }
  if (editingVar.value.fields.length === 0) {
    editingVar.value.fields = [{ name: '', value: '', is_secret: false }]
  }
  varModalOpen.value = true
}

const addEditField = () => {
  editingVar.value.fields.push({ name: '', value: '', is_secret: false })
}

const removeEditField = (idx) => {
  editingVar.value.fields.splice(idx, 1)
}

const saveVariable = async () => {
  if (!editingVar.value.key) {
    message.error('命名空间键不能为空')
    return
  }

  const fields = (editingVar.value.fields || [])
    .map((f) => ({
      name: String(f.name || '').trim(),
      value: String(f.value || ''),
      is_secret: !!f.is_secret,
    }))
    .filter((f) => f.name !== '')

  if (fields.length === 0) {
    message.error('至少需要一个字段')
    return
  }

  const nameSet = new Set()
  for (const f of fields) {
    if (nameSet.has(f.name)) {
      message.error(`字段名重复: ${f.name}`)
      return
    }
    nameSet.add(f.name)
  }

  if (editingVar.value.id) {
    for (const f of fields) {
      if (!f.value.trim()) {
        message.error(`编辑时字段值不能为空: ${f.name}`)
        return
      }
    }
  }

  if (!editingVar.value.id) {
    for (const f of fields) {
      if (!f.value.trim()) {
        message.error(`字段值不能为空: ${f.name}`)
        return
      }
    }
  }

  const apiFields = fields.map((f) => ({
    name: f.name,
    value: f.value,
    is_secret: f.is_secret,
  }))

  if (apiFields.length === 0) {
    message.error('字段列表不能为空')
    return
  }

  try {
    await saveGlobalVar(token.value, {
      key: editingVar.value.key,
      fields: apiFields,
      description: editingVar.value.description || '',
    })
    message.success(editingVar.value.id ? '变量已保存' : '变量已创建')
    varModalOpen.value = false
    await loadGlobalVariables()
  } catch (err) {
    message.error(`保存失败: ${err.message}`)
  }
}

const deleteVariable = async (id) => {
  try {
    await deleteGlobalVar(token.value, id)
    message.success('变量已删除')
    await loadGlobalVariables()
  } catch (err) {
    message.error(`删除失败: ${err.message}`)
  }
}

onMounted(async () => {
  await loadGlobalVariables()
})
</script>

<style scoped>
.global-vars-page {
  min-height: 100vh;
  padding: 20px;
  background: linear-gradient(180deg, #f4f7fb 0%, #f9fbff 100%);
}

.vars-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.vars-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #1f3552;
}

.vars-container {
  max-width: 1000px;
}

.var-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.var-card {
  border: 1px solid #dce5f4;
  border-radius: 10px;
  padding: 16px;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  transition: all 0.3s ease;
}

.var-card:hover {
  border-color: #1677ff;
  box-shadow: 0 4px 12px rgba(22, 119, 255, 0.15);
}

.var-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.var-key-badge {
  font-size: 14px;
  font-weight: 600;
  color: #1f3552;
  background: #f0f5ff;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: monospace;
}

.var-secret-badge {
  font-size: 11px;
  color: #d46a6a;
  background: #fef2f0;
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 600;
}

.var-body {
  margin-bottom: 12px;
}

.var-desc {
  font-size: 12px;
  color: #6f87a8;
  margin-bottom: 6px;
  line-height: 1.4;
}

.var-preview {
  font-size: 12px;
  color: #8ba0bf;
  font-family: monospace;
  background: #f0f5ff;
  padding: 6px 8px;
  border-radius: 4px;
  word-break: break-all;
}

.field-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-item {
  display: grid;
  grid-template-columns: 120px 1fr auto;
  gap: 8px;
  align-items: center;
  background: #f0f5ff;
  border-radius: 4px;
  padding: 6px 8px;
}

.field-name {
  font-family: monospace;
  color: #1f3552;
  font-size: 12px;
}

.field-value {
  font-family: monospace;
  color: #6f87a8;
  font-size: 12px;
  word-break: break-all;
}

.field-badge {
  font-size: 11px;
  color: #3e6ea8;
  background: #e8f3ff;
  border-radius: 3px;
  padding: 2px 6px;
}

.field-badge.secret {
  color: #d46a6a;
  background: #fef2f0;
}

.edit-field-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.edit-field-row {
  display: grid;
  grid-template-columns: 1fr 1fr auto auto;
  gap: 8px;
  align-items: center;
}

.var-footer {
  display: flex;
  align-items: center;
  gap: 0;
  padding-top: 8px;
  border-top: 1px solid #f0f5ff;
}

@media (max-width: 768px) {
  .var-list {
    grid-template-columns: 1fr;
  }

  .vars-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .edit-field-row {
    grid-template-columns: 1fr;
  }

  .field-item {
    grid-template-columns: 1fr;
  }
}
</style>
