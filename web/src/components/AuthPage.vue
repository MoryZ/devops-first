<template>
  <div class="auth-container">
    <a-card class="auth-card" title="DevOps Deploy Console" :bordered="false">
      <template v-if="!isRegMode">
        <div class="auth-subtitle">Login</div>
        <a-input
          v-model:value="form.username"
          placeholder="Username"
          size="large"
          class="auth-input"
          @keyup.enter="handleLogin"
        />
        <a-input-password
          v-model:value="form.password"
          placeholder="Password"
          size="large"
          class="auth-input"
          @keyup.enter="handleLogin"
        />
        <a-button
          type="primary"
          size="large"
          block
          :loading="loading"
          @click="handleLogin"
          class="auth-button"
        >
          Login
        </a-button>
        <div class="auth-switch">
          Don't have an account?
          <a @click="isRegMode = true">Register now</a>
        </div>
      </template>

      <template v-else>
        <div class="auth-subtitle">Register</div>
        <a-input
          v-model:value="form.username"
          placeholder="Username"
          size="large"
          class="auth-input"
        />
        <a-input-password
          v-model:value="form.password"
          placeholder="Password (min 6 chars)"
          size="large"
          class="auth-input"
        />
        <a-input
          v-model:value="form.email"
          type="email"
          placeholder="Email (optional)"
          size="large"
          class="auth-input"
        />
        <a-input
          v-model:value="form.remark"
          placeholder="Remark (optional)"
          size="large"
          class="auth-input"
        />
        <a-button
          type="primary"
          size="large"
          block
          :loading="loading"
          @click="handleRegister"
          class="auth-button"
        >
          Register
        </a-button>
        <div class="auth-switch">
          Already have an account?
          <a @click="isRegMode = false">Login here</a>
        </div>
      </template>
    </a-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { message } from 'ant-design-vue'

const emit = defineEmits(['login-success'])

const isRegMode = ref(false)
const loading = ref(false)
const form = ref({
  username: '',
  password: '',
  email: '',
  remark: '',
})

const readErrorMessage = async (response, fallback) => {
  try {
    const payload = await response.json()
    return payload?.error || fallback
  } catch {
    const text = await response.text()
    return text || fallback
  }
}

const handleLogin = async () => {
  if (!form.value.username || !form.value.password) {
    message.warning('Please enter username and password')
    return
  }

  loading.value = true
  try {
    const response = await fetch('/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: form.value.username,
        password: form.value.password,
      }),
    })

    if (!response.ok) {
      const errMsg = await readErrorMessage(response, 'Login failed')
      message.error(errMsg)
      return
    }

    const data = await response.json()
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
    message.success('Login successful')
    emit('login-success', { token: data.token, user: data.user })
  } catch (err) {
    message.error('Network error: ' + err.message)
  } finally {
    loading.value = false
  }
}

const handleRegister = async () => {
  if (!form.value.username || !form.value.password) {
    message.warning('Username and password are required')
    return
  }
  if (form.value.password.length < 6) {
    message.warning('Password must be at least 6 characters')
    return
  }

  loading.value = true
  try {
    const response = await fetch('/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: form.value.username,
        password: form.value.password,
        email: form.value.email || undefined,
        remark: form.value.remark || undefined,
      }),
    })

    if (!response.ok) {
      const errMsg = await readErrorMessage(response, 'Registration failed')
      message.error(errMsg)
      return
    }

    await response.json()
    isRegMode.value = false
    form.value = {
      username: form.value.username,
      password: '',
      email: '',
      remark: '',
    }
    message.success('Registration successful, please login')
  } catch (err) {
    message.error('Network error: ' + err.message)
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
  background: radial-gradient(circle at 10% 5%, #ffe6cc 0%, #f3f7ff 35%, #eaf5ee 100%);
}

.auth-card {
  width: 100%;
  max-width: 400px;
  border-radius: 16px;
  box-shadow: 0 16px 34px rgba(17, 36, 64, 0.15);
}

.auth-subtitle {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #1f2d3d;
}

.auth-input {
  margin-bottom: 12px;
}

.auth-button {
  margin-top: 8px;
  margin-bottom: 16px;
}

.auth-switch {
  text-align: center;
  font-size: 14px;
  color: #666;
}

.auth-switch a {
  color: #1890ff;
  cursor: pointer;
  margin-left: 4px;
}
</style>
