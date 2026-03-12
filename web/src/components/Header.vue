<template>
  <header class="board-header">
    <div class="board-header-left">
      <SystemSelector :token="token" compact />
    </div>
    <div class="board-brand">wizard</div>
    <div class="board-header-right">
      <a-button type="text" class="icon-btn"><SearchOutlined /></a-button>
      <a-button type="text" class="icon-btn"><FilterOutlined /></a-button>
      <a-button type="primary" class="run-btn">
        <RocketOutlined />
        快速执行
      </a-button>
      <div class="avatar-pill">{{ userInitial }}</div>
      <a-button type="link" danger @click="handleLogout">退出</a-button>
    </div>
  </header>
</template>

<script setup>
import { SearchOutlined, FilterOutlined, RocketOutlined } from '@ant-design/icons-vue'
import SystemSelector from './SystemSelector.vue'

defineProps({
  token: {
    type: String,
    default: '',
  },
  userInitial: {
    type: String,
    default: 'U',
  },
})

const emit = defineEmits(['logout'])

const handleLogout = () => {
  emit('logout')
}
</script>

<style scoped>
.board-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 52px;
  padding: 0 18px;
  background: linear-gradient(100deg, #0d1b33 0%, #13254a 72%, #1a3863 100%);
}

.board-header-left,
.board-header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.board-header-left {
  min-width: 0;
}

.board-brand {
  color: #e5eeff;
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 0.5px;
  font-family: 'Avenir Next', 'Trebuchet MS', sans-serif;
}

.icon-btn {
  color: #dce8ff;
}

.run-btn {
  border-radius: 8px;
  height: 34px;
}

.avatar-pill {
  width: 28px;
  height: 28px;
  border-radius: 14px;
  background: linear-gradient(140deg, #d4a55a, #b56822);
  color: #fff;
  display: grid;
  place-items: center;
  font-weight: 700;
  font-size: 12px;
}

@media (max-width: 1100px) {
  .board-header {
    height: auto;
    min-height: 52px;
    padding: 8px 12px;
    gap: 8px;
    flex-wrap: wrap;
  }

  .board-header-left {
    min-width: 0;
    flex: none;
  }
}
</style>
