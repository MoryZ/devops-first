import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useWorkspaceStore = defineStore('workspace', () => {
  const systems = ref([])
  const selectedSystemId = ref('')
  const selectedPlanId = ref('')

  const selectedSystem = computed(() => {
    return systems.value.find((s) => s.id === selectedSystemId.value) || null
  })

  const setSystems = (items) => {
    // Transform API response from PascalCase to camelCase
    systems.value = (items || []).map((item) => ({
      id: item.ID || item.id,
      name: item.Name || item.name,
      description: item.Description || item.description,
      status: item.Status || item.status,
      userId: item.UserID || item.userId,
      createdAt: item.CreatedAt || item.createdAt,
      updatedAt: item.UpdatedAt || item.updatedAt,
    }))
    if (!systems.value.length) {
      selectedSystemId.value = ''
      selectedPlanId.value = ''
      return
    }
    if (!selectedSystemId.value || !systems.value.some((s) => s.id === selectedSystemId.value)) {
      selectedSystemId.value = systems.value[0].id
      selectedPlanId.value = ''
    }
  }

  const selectSystem = (systemId) => {
    if (selectedSystemId.value !== systemId) {
      selectedSystemId.value = systemId
      selectedPlanId.value = ''
    }
  }

  const selectPlan = (planId) => {
    selectedPlanId.value = planId || ''
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
})
