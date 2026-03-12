import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'
import LoginView from '../views/LoginView.vue'
import PipelineExecutionsView from '../views/PipelineExecutionsView.vue'
import PipelineBPMDesignerView from '../views/PipelineBPMDesignerView.vue'

const routes = [
  {
    path: '/',
    redirect: '/workspace',
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: { guestOnly: true },
  },
  {
    path: '/workspace',
    name: 'workspace',
    component: DashboardView,
    meta: { requiresAuth: true },
  },
  {
    path: '/release-units',
    name: 'release-units',
    component: DashboardView,
    meta: { requiresAuth: true },
  },
  {
    path: '/global-vars',
    name: 'global-vars',
    component: DashboardView,
    meta: { requiresAuth: true },
  },
  {
    path: '/pipelines/:id/executions',
    name: 'pipeline-executions',
    component: PipelineExecutionsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/pipelines/:id/bpm',
    name: 'pipeline-bpm',
    component: PipelineBPMDesignerView,
    meta: { requiresAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/workspace',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  const isAuthenticated = Boolean(token)

  if (to.meta.requiresAuth && !isAuthenticated) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  if (to.meta.guestOnly && isAuthenticated) {
    const redirect = typeof to.query.redirect === 'string' ? to.query.redirect : '/workspace'
    return redirect
  }

  return true
})

export default router
