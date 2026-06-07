<template>
  <header class="board-header">
    <div class="header-glow"></div>
    <div class="board-header-left">
      <div class="workspace-label">
        <span class="label-dot"></span>
        我的工作台
      </div>
      <SystemSelector :token="token" compact />
    </div>
    <div class="board-brand">
      <span class="brand-icon">◆</span>
      <span class="brand-text">WIZARD</span>
    </div>
    <div class="board-header-right">
      <a-button type="primary" class="run-btn" @click="$emit('run')">
        <PlayCircleOutlined />
        <span class="btn-text">快速执行</span>
      </a-button>
      <div class="avatar-pill">
        <span class="avatar-glow"></span>
        {{ userInitial }}
      </div>
      <a-button type="text" class="logout-btn" @click="$emit('logout')">
        退出
      </a-button>
    </div>
  </header>
</template>

<script setup>
import { PlayCircleOutlined } from '@ant-design/icons-vue'
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

defineEmits(['logout', 'run'])
</script>

<style scoped>
.board-header {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 64px;
  padding: 0 24px;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(14px) saturate(140%);
  border-bottom: 1px solid #e4e8f1;
  animation: slideDown 0.4s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.header-glow {
  position: absolute;
  top: -50%;
  left: 50%;
  transform: translateX(-50%);
  width: 600px;
  height: 100px;
  background: radial-gradient(ellipse at center, rgba(0, 212, 255, 0.16) 0%, transparent 70%);
  pointer-events: none;
}

.board-header-left,
.board-header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.workspace-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #4b5a72;
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  letter-spacing: 0.01em;
}

.label-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
  animation: pulse 2s ease-in-out infinite;
}

.board-brand {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 10px;
  color: #0f1b2d;
  font-family: var(--font-display);
}

.brand-icon {
  font-size: 20px;
  color: var(--accent-primary);
  filter: drop-shadow(0 0 8px rgba(0, 212, 255, 0.45));
  animation: glow 3s ease-in-out infinite;
}

.brand-text {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.15em;
  background: linear-gradient(135deg, #0f1b2d 0%, var(--accent-primary) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.run-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 20px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--accent-primary) 0%, var(--accent-info) 100%);
  border: none;
  font-weight: 600;
  font-family: var(--font-display);
  letter-spacing: 0.02em;
  transition: all var(--transition-base);
  box-shadow: 0 4px 15px rgba(0, 212, 255, 0.25);
  color: #fff;
}

.run-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 212, 255, 0.35);
}

.btn-text {
  font-size: 14px;
}

.avatar-pill {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #f59e0b 0%, #ef4444 100%);
  color: #fff;
  display: grid;
  place-items: center;
  font-weight: 700;
  font-size: 14px;
  font-family: var(--font-mono);
  cursor: pointer;
  transition: all var(--transition-base);
  border: 2px solid rgba(255, 255, 255, 0.2);
}

.avatar-pill:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 15px rgba(245, 158, 11, 0.35);
}

.avatar-glow {
  position: absolute;
  inset: -2px;
  border-radius: 50%;
  background: linear-gradient(135deg, #f59e0b, #ef4444);
  opacity: 0;
  z-index: -1;
  filter: blur(8px);
  transition: opacity var(--transition-base);
}

.avatar-pill:hover .avatar-glow {
  opacity: 0.6;
}

.logout-btn {
  color: #4b5a72;
  font-size: 14px;
  transition: all var(--transition-fast);
}

.logout-btn:hover {
  color: var(--accent-danger);
}

@media (max-width: 1100px) {
  .board-header {
    height: auto;
    min-height: 64px;
    padding: 12px 16px;
    gap: 12px;
    flex-wrap: wrap;
  }

  .board-brand {
    position: static;
    transform: none;
    order: -1;
    width: 100%;
    justify-content: center;
    margin-bottom: 8px;
  }

  .board-header-left {
    min-width: 0;
    flex: 1;
  }

  .board-header-right {
    flex-shrink: 0;
  }
}
</style>
