<script setup>
import { inject, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const currentUser = inject('currentUser')
const user = ref(null)
const isLoading = ref(true)
const error = ref(null)
const route = useRoute()

async function loadUser() {
  try {
    const userId = route.params.user_id
    if (userId === 'me') {
      if (!currentUser.value) {
        error.value = 'Пожалуйста, войдите в систему для просмотра своего профиля'
        return
      }
      user.value = currentUser.value
    } else {
      const response = await fetch(`/api/users/${userId}`)
      if (response.ok) {
        user.value = await response.json()
      } else {
        error.value = 'Не удалось загрузить данные пользователя'
      }
    }
  } catch (error) {
    console.error('Error loading user:', error)
    error.value = 'Не удалось загрузить данные пользователя'
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadUser()
})
</script>

<template>
  <div class="max-w-screen-md mx-auto px-4 py-8">
    <!-- Loading State -->
    <div v-if="isLoading" class="flex justify-center py-8">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4 text-center text-red-600">
      {{ error }}
    </div>

    <!-- Content -->
    <div v-else-if="user" class="space-y-6">
      <!-- User Info Card -->
      <div class="bg-white rounded-lg shadow-md overflow-hidden">
        <div class="p-6">
          <h1 class="text-2xl font-bold text-gray-900">{{ user.name }}</h1>
          <p v-if="user.email" class="text-gray-600 mt-1">{{ user.email }}</p>
        </div>
      </div>

      <!-- Climbs Section -->
      <div class="bg-white rounded-lg shadow-md overflow-hidden">
        <div class="p-6">
          <h2 class="text-xl font-semibold text-gray-900 mb-4">Восхождения</h2>
          
          <div v-if="!user.climbs?.length" class="text-center text-gray-500 py-4">
            Пока нет зарегистрированных восхождений
          </div>

          <div v-else class="space-y-4">
            <div 
              v-for="climb in user.climbs" 
              :key="climb.id" 
              class="border border-gray-200 rounded-lg p-4 hover:bg-gray-50"
            >
              <div class="flex justify-between items-start">
                <div>
                  <RouterLink 
                    :to="`/${climb.ridge_id}/${climb.summit_id}`"
                    class="text-lg font-medium text-gray-900 hover:text-blue-600"
                  >
                    {{ climb.summit_name }}
                  </RouterLink>
                  <p class="text-sm text-gray-600">Хребет: {{ climb.ridge_name }}</p>
                  <p class="text-sm text-gray-600">Высота: {{ climb.height }}м</p>
                </div>
                <div class="text-sm text-gray-500">
                  {{ new Date(climb.date).toLocaleDateString() }}
                </div>
              </div>
              <p v-if="climb.description" class="mt-2 text-sm text-gray-600">
                {{ climb.description }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
