<template>
  <div class="executions-view">
    <div class="head-row">
      <a-button @click="goBack">返回流水线</a-button>
      <div class="title">执行历史 · {{ pipelineName || pipelineId }}</div>
    </div>

    <ExecutionHistoryPanel :token="token" :pipelineId="pipelineId" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ExecutionHistoryPanel from '../components/ExecutionHistoryPanel.vue'

const route = useRoute()
const router = useRouter()

const token = computed(() => localStorage.getItem('token') || '')
const pipelineId = computed(() => String(route.params.id || ''))
const pipelineName = computed(() => String(route.query.name || ''))

const goBack = () => {
  router.push('/workspace')
}
</script>

<style scoped>
.executions-view {
  min-height: 100vh;
  background: #f3f5f9;
  padding: 18px;
}

.head-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.title {
  font-size: 18px;
  font-weight: 700;
  color: #1f2f46;
}
</style>
