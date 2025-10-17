<script setup>
import { inject, ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { formatRussianDate } from '../utils/dates'
import UserAvatar from '../components/UserAvatar.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ErrorMessage from '../components/ErrorMessage.vue'
const props = defineProps({
  user_id: {
    type: String,
    default: null
  }
})

const currentUser = inject('currentUser')
const user = ref(null)
const climbs = ref([])
const isLoading = ref(true)
const error = ref(null)
const route = useRoute()

// Watch for user data changes to update page title
watch(user, (newUser) => {
  if (newUser) {
    document.title = `${newUser.name} | Тысячники Южного Урала`
  }
}, { immediate: true })

async function loadUser() {
  try {
    isLoading.value = true
    error.value = null

    const userId = props.user_id || route.params.user_id
    if (!userId) {
      error.value = 'Неверный URL профиля'
      return
    }

    if (userId === 'me') {
      if (!currentUser.value) {
        error.value = 'Пожалуйста, войдите в систему для просмотра своего профиля'
        return
      }
      user.value = currentUser.value
    } else {
      const response = await fetch(`/api/user/${userId}`)
      if (response.ok) {
        user.value = await response.json()
      } else {
        error.value = 'Не удалось загрузить данные пользователя'
        return
      }
    }

    // Load climbs
    const actualUserId = userId === 'me' ? currentUser.value.id : userId
    const climbsResponse = await fetch(`/api/user/${actualUserId}/climbs`)
    if (climbsResponse.ok) {
      climbs.value = await climbsResponse.json()
    } else {
      error.value = 'Не удалось загрузить список восхождений'
    }
  } catch (error) {
    console.error('Error loading user:', error)
    error.value = 'Не удалось загрузить данные пользователя'
  } finally {
    isLoading.value = false
  }
}

// Watch for route changes to reload user data
watch(() => route.params.user_id, () => {
  loadUser()
})

function getSocialLinkText(src) {
  switch(src) {
    case 1:
      return 'Профиль VK'
    case 2:
      return 'Профиль southural.ru'
    default:
      return 'Социальная сеть'
  }
}

onMounted(() => {
  loadUser()
})
</script>

<template>
  <div class="max-w-screen-md mx-auto overflow-hidden">
    <!-- Loading State -->
    <LoadingSpinner v-if="isLoading" />

    <!-- Error State -->
    <ErrorMessage v-else-if="error" :message="error" />

    <!-- Content -->
    <div v-else-if="user" class="space-y-6">
      <!-- User Info Card -->
      <div class="flex items-center space-x-4">
        <UserAvatar 
          :image-url="user.image_m"
          :alt-text="user.name"
        />
        <div>
          <h1 class="text-2xl font-bold text-gray-900">{{ user.name }}</h1>
          <a v-if="user.social_link" :href="user.social_link" target="_blank" rel="noopener noreferrer"
            class="text-sm text-blue-600 hover:text-blue-800 hover:underline">
            {{ getSocialLinkText(user.src) }}
          </a>
        </div>
      </div>

      <!-- Climbs Section -->
      <h2 class="text-xl font-semibold text-gray-900 mb-4">
        Восхождения <span v-if="climbs.length > 0" class="text-sm font-normal text-gray-500">({{ climbs.length }})</span>
      </h2>
      <div v-if="!climbs.length" class="text-center text-gray-500">
        Пока нет зарегистрированных восхождений
      </div>
      <div v-else>
        <ul>
          <li v-for="climb in climbs" :key="climb.id" class="py-2 border-b last:border-b-0 flex flex-col">
            <div class="flex justify-between items-baseline">
              <div>
                <RouterLink :to="`/${climb.ridge.id}/${climb.id}`"
                  class="font-bold text-lg text-gray-900 hover:text-blue-600">
                  {{ climb.name || climb.height }}
                </RouterLink>
                <span class="text-sm text-gray-600 ml-2">хребет {{ climb.ridge.name }}</span>
              </div>
              <div class="flex items-center gap-2">
                <div class="text-sm text-gray-500 whitespace-nowrap ml-4">
                  {{ formatRussianDate(climb.climb_data?.date) }}
                </div>
                <template v-if="(props.user_id === 'me' || user?.id === currentUser?.id)">
                  <RouterLink :to="`/${climb.ridge.id}/${climb.id}/climb`"
                    class="ml-2 text-blue-500 hover:underline text-xs">
                    Редактировать
                  </RouterLink>
                </template>
              </div>
            </div>
            <div v-if="climb.climb_data?.comment" class="text-sm mt-1">
              {{ climb.climb_data.comment }}
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
