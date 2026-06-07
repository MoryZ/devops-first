<template>
  <div class="page-wrapper" :class="[`bg-${bg}`, { 'page-animate': animate }]">
    <div class="page-bg" v-if="bg !== 'solid'">
      <div class="bg-gradient"></div>
    </div>
    <div class="page-container">
      <header v-if="$slots.header || title" class="page-header">
        <slot name="header">
          <div class="page-header-content">
            <div class="header-left">
              <span v-if="icon" class="header-icon">{{ icon }}</span>
              <h1 class="page-title">{{ title }}</h1>
              <p v-if="subtitle" class="page-subtitle">{{ subtitle }}</p>
            </div>
            <div v-if="$slots.actions" class="header-actions">
              <slot name="actions" />
            </div>
          </div>
          <div class="header-underline"></div>
        </slot>
      </header>

      <main class="page-content">
        <div class="core-surface">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: String,
  subtitle: String,
  icon: String,
  bg: {
    type: String,
    default: 'gradient',
  },
  animate: {
    type: Boolean,
    default: true,
  },
})
</script>

<style scoped>
.page-wrapper {
  min-height: 100vh;
  padding: 24px;
  position: relative;
}

.page-wrapper.bg-solid {
  background: var(--bg-primary);
}

.page-wrapper.bg-gradient {
  background: var(--bg-primary);
}

.page-bg {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.bg-gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 15% 0%, rgba(0, 212, 255, 0.08) 0%, transparent 50%),
    radial-gradient(ellipse at 85% 100%, rgba(124, 58, 237, 0.06) 0%, transparent 50%);
}

.page-container {
  position: relative;
  z-index: 1;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;
  animation: slideDown 0.5s ease-out;
}

.page-header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  flex-wrap: wrap;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.header-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border-radius: 10px;
  font-size: 16px;
  color: white;
  box-shadow: 0 4px 14px rgba(0, 212, 255, 0.3);
  margin-bottom: 4px;
}

.page-title {
  font-family: var(--font-display);
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-tertiary);
  margin: 0;
  font-weight: 400;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.header-underline {
  height: 3px;
  margin-top: 14px;
  background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary), transparent);
  border-radius: 2px;
  opacity: 0.6;
}

.page-content {
  animation: fadeIn 0.5s ease-out 0.1s backwards;
}

.page-animate .page-header {
  animation: slideDown 0.5s ease-out;
}

.page-animate .page-content {
  animation: slideUp 0.5s ease-out 0.15s backwards;
}

.core-surface {
  background: var(--core-surface);
  border: 1px solid var(--core-border);
  border-radius: 16px;
  padding: 22px;
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.08);
  color: var(--core-text);
}

.core-surface :deep(.ant-form-item-label > label) {
  color: var(--core-text);
  font-weight: 500;
}

.core-surface :deep(.ant-radio-button-wrapper) {
  background: var(--core-surface);
  color: var(--core-text-secondary);
  border-color: var(--core-border-accent);
}

.core-surface :deep(.ant-radio-button-wrapper-checked) {
  background: rgba(0, 212, 255, 0.12);
  border-color: var(--accent-primary);
  color: #084c61;
  box-shadow: none;
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

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .page-wrapper {
    padding: 16px;
  }

  .page-title {
    font-size: 22px;
  }

  .page-header-content {
    flex-direction: column;
  }

  .header-actions {
    width: 100%;
  }

  .core-surface {
    padding: 16px;
  }
}
</style>
