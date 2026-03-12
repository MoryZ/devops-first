<template>
  <a-layout class="page-layout">
    <DashboardPage :token="token" :currentUser="currentUser" @logout="handleLogout" />
  </a-layout>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import DashboardPage from '../components/DashboardPage.vue'

const router = useRouter()

const token = computed(() => localStorage.getItem('token') || '')
const currentUser = computed(() => {
  const raw = localStorage.getItem('user')
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
})

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  router.replace('/login')
}
</script>

<style scoped>
.page-layout {
  min-height: 100vh;
}
</style>
