<template>
  <div class="auth-container">
    <div class="auth-bg">
      <div class="grid-pattern"></div>
      <div class="glow-orb orb-1"></div>
      <div class="glow-orb orb-2"></div>
    </div>

    <div class="auth-card">
      <div class="card-glow"></div>
      <div class="auth-header">
        <div class="auth-logo">
          <span class="logo-icon">◆</span>
        </div>
        <h1 class="auth-title">WIZARD</h1>
        <p class="auth-subtitle">DevOps 流水线管理平台</p>
      </div>

      <template v-if="!isRegMode">
        <div class="auth-form">
          <div class="input-group">
            <label class="input-label">用户名</label>
            <a-input
              v-model:value="form.username"
              placeholder="请输入用户名"
              size="large"
              class="auth-input"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <UserOutlined />
              </template>
            </a-input>
          </div>

          <div class="input-group">
            <label class="input-label">密码</label>
            <a-input-password
              v-model:value="form.password"
              placeholder="请输入密码"
              size="large"
              class="auth-input"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <LockOutlined />
              </template>
            </a-input-password>
          </div>

          <a-button
            type="primary"
            size="large"
            block
            :loading="loading"
            @click="handleLogin"
            class="auth-button"
          >
            <span v-if="!loading">登 录</span>
            <span v-else>登录中...</span>
          </a-button>
        </div>

        <div class="auth-switch">
          还没有账号？
          <a @click="isRegMode = true">立即注册</a>
        </div>
      </template>

      <template v-else>
        <div class="auth-form">
          <div class="input-group">
            <label class="input-label">用户名</label>
            <a-input
              v-model:value="form.username"
              placeholder="请输入用户名"
              size="large"
              class="auth-input"
            >
              <template #prefix>
                <UserOutlined />
              </template>
            </a-input>
          </div>

          <div class="input-group">
            <label class="input-label">密码</label>
            <a-input-password
              v-model:value="form.password"
              placeholder="密码（至少6位）"
              size="large"
              class="auth-input"
            >
              <template #prefix>
                <LockOutlined />
              </template>
            </a-input-password>
          </div>

          <div class="input-group">
            <label class="input-label">邮箱（可选）</label>
            <a-input
              v-model:value="form.email"
              placeholder="your@email.com"
              size="large"
              class="auth-input"
            >
              <template #prefix>
                <MailOutlined />
              </template>
            </a-input>
          </div>

          <div class="input-group">
            <label class="input-label">备注（可选）</label>
            <a-input
              v-model:value="form.remark"
              placeholder="添加备注信息"
              size="large"
              class="auth-input"
            />
          </div>

          <a-button
            type="primary"
            size="large"
            block
            :loading="loading"
            @click="handleRegister"
            class="auth-button"
          >
            <span v-if="!loading">注 册</span>
            <span v-else>注册中...</span>
          </a-button>
        </div>

        <div class="auth-switch">
          已有账号？
          <a @click="isRegMode = false">立即登录</a>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons-vue'
import { login, register } from '../api/auth'

const emit = defineEmits(['login-success'])

const isRegMode = ref(false)
const loading = ref(false)
const form = ref({
  username: '',
  password: '',
  email: '',
  remark: '',
})

const handleLogin = async () => {
  if (!form.value.username || !form.value.password) {
    message.warning('请输入用户名和密码')
    return
  }

  loading.value = true
  try {
    const data = await login({
      username: form.value.username,
      password: form.value.password,
    })
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
    message.success('登录成功')
    emit('login-success', { token: data.token, user: data.user })
  } catch (err) {
    message.error('登录失败: ' + err.message)
  } finally {
    loading.value = false
  }
}

const handleRegister = async () => {
  if (!form.value.username || !form.value.password) {
    message.warning('用户名和密码不能为空')
    return
  }
  if (form.value.password.length < 6) {
    message.warning('密码至少需要6个字符')
    return
  }

  loading.value = true
  try {
    await register({
      username: form.value.username,
      password: form.value.password,
      email: form.value.email || undefined,
      remark: form.value.remark || undefined,
    })
    isRegMode.value = false
    form.value = {
      username: form.value.username,
      password: '',
      email: '',
      remark: '',
    }
    message.success('注册成功，请登录')
  } catch (err) {
    message.error('注册失败: ' + err.message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 20px;
  position: relative;
  overflow: hidden;
  background: var(--bg-primary);
}

.auth-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  z-index: 0;
}

.grid-pattern {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 212, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.03) 1px, transparent 1px);
  background-size: 50px 50px;
  mask-image: radial-gradient(ellipse at center, black 20%, transparent 70%);
}

.glow-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
}

.orb-1 {
  top: -20%;
  right: -10%;
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(0, 212, 255, 0.3) 0%, transparent 70%);
  animation: float 8s ease-in-out infinite;
}

.orb-2 {
  bottom: -20%;
  left: -10%;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(124, 58, 237, 0.25) 0%, transparent 70%);
  animation: float 10s ease-in-out infinite reverse;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(30px, -30px) scale(1.1); }
}

.auth-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  padding: 40px 36px;
  background: rgba(21, 29, 46, 0.85);
  backdrop-filter: blur(20px);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg), 0 0 60px rgba(0, 212, 255, 0.1);
  animation: slideUp 0.6s ease-out;
}

.card-glow {
  position: absolute;
  top: -1px;
  left: 20%;
  right: 20%;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--accent-primary), transparent);
  border-radius: 1px;
}

.auth-header {
  text-align: center;
  margin-bottom: 36px;
}

.auth-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border-radius: 16px;
  margin-bottom: 16px;
  box-shadow: 0 8px 24px rgba(0, 212, 255, 0.3);
  animation: pulse 3s ease-in-out infinite;
}

.logo-icon {
  font-size: 24px;
  color: white;
}

.auth-title {
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 700;
  letter-spacing: 0.2em;
  margin: 0 0 8px;
  background: linear-gradient(135deg, var(--text-primary) 0%, var(--accent-primary) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.auth-subtitle {
  color: var(--text-tertiary);
  font-size: 14px;
  margin: 0;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.input-label {
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.01em;
}

.auth-input :deep(.ant-input) {
  background: var(--bg-tertiary) !important;
  border: 1px solid var(--border-color-light) !important;
  color: var(--text-primary) !important;
  border-radius: var(--radius-md) !important;
  height: 44px !important;
  font-family: var(--font-display) !important;
  transition: all var(--transition-fast) !important;
}

.auth-input :deep(.ant-input):hover {
  border-color: var(--accent-primary) !important;
}

.auth-input :deep(.ant-input:focus),
.auth-input :deep(.ant-input-focused) {
  border-color: var(--accent-primary) !important;
  box-shadow: 0 0 0 3px rgba(0, 212, 255, 0.15) !important;
}

.auth-input :deep(.ant-input::placeholder) {
  color: var(--text-muted) !important;
}

.auth-input :deep(.ant-input-prefix) {
  color: var(--text-muted) !important;
  margin-right: 10px !important;
}

.auth-input :deep(.ant-input-password) {
  background: var(--bg-tertiary) !important;
  border: 1px solid var(--border-color-light) !important;
  border-radius: var(--radius-md) !important;
}

.auth-input :deep(.ant-input-password:hover) {
  border-color: var(--accent-primary) !important;
}

.auth-input :deep(.ant-input-password .ant-input) {
  background: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

.auth-button {
  height: 48px !important;
  border-radius: var(--radius-md) !important;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-info)) !important;
  border: none !important;
  font-weight: 600 !important;
  font-size: 15px !important;
  letter-spacing: 0.05em !important;
  margin-top: 8px !important;
  box-shadow: 0 4px 15px rgba(0, 212, 255, 0.3) !important;
  transition: all var(--transition-base) !important;
}

.auth-button:hover {
  transform: translateY(-2px) !important;
  box-shadow: 0 6px 20px rgba(0, 212, 255, 0.4) !important;
}

.auth-switch {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
  color: var(--text-tertiary);
}

.auth-switch a {
  color: var(--accent-primary);
  cursor: pointer;
  margin-left: 4px;
  font-weight: 500;
  transition: all var(--transition-fast);
}

.auth-switch a:hover {
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.5);
}

@media (max-width: 480px) {
  .auth-card {
    padding: 28px 20px;
  }

  .auth-title {
    font-size: 22px;
  }
}
</style>
