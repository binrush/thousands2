<script setup>
import { ref, provide, watch, computed, markRaw } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { useAuth } from './auth'
import { getImageUrl } from './utils/images'

// Import layouts
import DefaultLayout from './layouts/DefaultLayout.vue'
import MapLayout from './layouts/MapLayout.vue'

const currentUser = ref(null)
const { authState } = useAuth()
const route = useRoute()

// Compute which layout to use based on the current route's meta.layout or component's layout option
const currentLayout = computed(() => {
  if (route.meta.layout === 'MapLayout') {
    return MapLayout
  }
  return DefaultLayout
})

// Watch for auth state changes to update currentUser
watch(() => authState.user, (newUser) => {
  currentUser.value = newUser
}, { immediate: true })

provide('currentUser', currentUser)
</script>

<template>
  <component :is="currentLayout">
    <RouterView />
  </component>
</template>
