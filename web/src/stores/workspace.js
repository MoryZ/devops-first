import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useWorkspaceStore = defineStore(
  'workspace',
  () => {
    const normalizeId = (id) => {
      if (id === null || id === undefined) return ''
      return String(id)
    }

    const systems = ref([])
    const selectedSystemId = ref('')
    const selectedPlanId = ref('')

    const selectedSystem = computed(() => {
      return systems.value.find((s) => s.id === selectedSystemId.value) || null
    })

    const setSystems = (items) => {
      // Transform API response from PascalCase to camelCase
      systems.value = (items || []).map((item) => ({
        id: normalizeId(item.ID || item.id),
        name: item.Name || item.name,
        description: item.Description || item.description,
        status: item.Status || item.status,
        userId: normalizeId(item.UserID || item.userId),
        createdAt: item.CreatedAt || item.createdAt,
        updatedAt: item.UpdatedAt || item.updatedAt,
      }))
      if (!systems.value.length) {
        selectedSystemId.value = ''
        selectedPlanId.value = ''
        return
      }
      const currentSystemId = normalizeId(selectedSystemId.value)
      if (!currentSystemId || !systems.value.some((s) => s.id === currentSystemId)) {
        selectedSystemId.value = systems.value[0].id
        selectedPlanId.value = ''
      }
    }

    const selectSystem = (systemId) => {
      const normalizedSystemId = normalizeId(systemId)
      if (selectedSystemId.value !== normalizedSystemId) {
        selectedSystemId.value = normalizedSystemId
        selectedPlanId.value = ''
      }
    }

    const selectPlan = (planId) => {
      selectedPlanId.value = normalizeId(planId)
    }

    return {
      systems,
      selectedSystemId,
      selectedPlanId,
      selectedSystem,
      setSystems,
      selectSystem,
      selectPlan,
    }
  },
  {
    persist: {
      key: 'workspace',
      storage: localStorage,
      pick: ['selectedSystemId', 'selectedPlanId'],
    },
  }
)
