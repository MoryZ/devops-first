<template>
  <a-card class="dag-card" :bordered="false" title="执行 DAG">
    <div class="dag-lane">
      <div class="lane-title">主流程</div>
      <div class="node-row">
        <template v-for="(stage, idx) in safeMainStages" :key="`main-${idx}`">
          <div class="dag-node" :class="`node-${stage.status || 'pending'}`">
            <div class="node-name">{{ stage.name || `阶段 ${idx + 1}` }}</div>
          </div>
          <span v-if="idx < safeMainStages.length - 1" class="node-link"></span>
        </template>
      </div>
    </div>

    <div class="dag-lane">
      <div class="lane-title">环境阶段</div>
      <div class="node-row">
        <template v-for="(stage, idx) in safeEnvStages" :key="`env-${idx}`">
          <div class="dag-node" :class="`node-${stage.status || 'pending'}`">
            <div class="node-name">{{ stage.name || `环境 ${idx + 1}` }}</div>
          </div>
          <span v-if="idx < safeEnvStages.length - 1" class="node-link"></span>
        </template>
      </div>
    </div>
  </a-card>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  mainStages: {
    type: Array,
    default: () => [],
  },
  envStages: {
    type: Array,
    default: () => [],
  },
})

const safeMainStages = computed(() => (props.mainStages?.length ? props.mainStages : [{ name: '触发源', status: 'pending' }, { name: '构建', status: 'pending' }, { name: '部署', status: 'pending' }]))
const safeEnvStages = computed(() => (props.envStages?.length ? props.envStages : [{ name: '测试环境', status: 'pending' }, { name: '生产环境', status: 'pending' }]))
</script>

<style scoped>
.dag-card {
  border-radius: 12px;
  box-shadow: 0 8px 22px rgba(17, 36, 64, 0.08);
  margin-bottom: 14px;
}

.dag-lane {
  margin-bottom: 12px;
}

.lane-title {
  font-size: 13px;
  color: #627691;
  margin-bottom: 8px;
}

.node-row {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 6px;
}

.dag-node {
  min-width: 120px;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid #dce4f2;
  background: #f7faff;
}

.node-name {
  font-size: 13px;
  color: #24384f;
  white-space: nowrap;
}

.node-link {
  width: 18px;
  height: 2px;
  background: #a3b2cb;
  flex-shrink: 0;
}

.node-success {
  border-color: #8de0a8;
  background: #e9fbf0;
}

.node-running {
  border-color: #8ec5ff;
  background: #ebf5ff;
}

.node-failed {
  border-color: #ffb0b0;
  background: #fff0f0;
}

.node-pending {
  border-color: #dce4f2;
  background: #f7faff;
}
</style>
