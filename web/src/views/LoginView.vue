<template>
  <a-layout class="page-layout">
    <AuthPage @login-success="handleLoginSuccess" />
  </a-layout>
</template>

<script setup>
import { useRoute, useRouter } from 'vue-router'
import AuthPage from '../components/AuthPage.vue'

const route = useRoute()
const router = useRouter()

const handleLoginSuccess = () => {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/workspace'
  router.replace(redirect)
}
</script>

<style scoped>
.page-layout {
  min-height: 100vh;
}
</style>
