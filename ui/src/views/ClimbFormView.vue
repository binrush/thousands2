<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { RouterLink } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isSubmitting = ref(false)
const isDeleting = ref(false)
const summit = ref(null)
const isLoading = ref(true)

const formData = ref({
  date: '',
  comment: ''
})

// Function to format date object to string
function formatDateToString(date) {
  if (!date) return ''
  if (!date.Year) return ''
  
  const parts = []
  if (date.Day) parts.push(date.Day.toString().padStart(2, '0'))
  if (date.Month) parts.push(date.Month.toString().padStart(2, '0'))
  parts.push(date.Year.toString())
  
  return parts.join('.')
}

async function fetchSummit() {
  try {
    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}`)
    if (!response.ok) throw new Error('Failed to fetch summit')
    summit.value = await response.json()
    
    // Initialize form data
    if (summit.value.climb_data) {
      formData.value.date = formatDateToString(summit.value.climb_data.date)
      formData.value.comment = summit.value.climb_data.comment || ''
    }
  } catch (error) {
    console.error('Error fetching summit:', error)
  } finally {
    isLoading.value = false
  }
}

async function submitClimb() {
  if (isSubmitting.value) return
  
  isSubmitting.value = true
  try {
    const formDataToSubmit = new FormData()
    formDataToSubmit.append('date', formData.value.date)
    formDataToSubmit.append('comment', formData.value.comment)

    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}`, {
      method: 'PUT',
      body: formDataToSubmit
    })

    if (response.ok) {
      router.push({ name: 'summit', params: route.params })
    } else {
      throw new Error('Failed to submit climb')
    }
  } catch (error) {
    console.error('Error submitting climb:', error)
    alert('Произошла ошибка при сохранении восхождения')
  } finally {
    isSubmitting.value = false
  }
}

async function deleteClimb() {
  if (!confirm('Вы уверены, что хотите удалить запись о восхождении?')) {
    return
  }

  if (isDeleting.value) return
  
  isDeleting.value = true
  try {
    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}`, {
      method: 'DELETE'
    })

    if (response.ok) {
      router.push({ name: 'summit', params: route.params })
    } else {
      throw new Error('Failed to delete climb')
    }
  } catch (error) {
    console.error('Error deleting climb:', error)
    alert('Произошла ошибка при удалении восхождения')
  } finally {
    isDeleting.value = false
  }
}

onMounted(() => {
  fetchSummit()
})
</script>

<template>
  <div class="climb-form-view">
    <div class="max-w-screen-md mx-auto">
      <h1 class="text-2xl font-bold mb-8">Регистрация восхождения</h1>
      
      <!-- Summit Information -->
      <div v-if="isLoading" class="flex justify-center py-4">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
      
      <div v-else-if="summit" class="mb-8 bg-white">
        <div class="flex items-center space-x-2">
          <RouterLink 
            :to="{ name: 'summit', params: { ridge_id: route.params.ridge_id, summit_id: route.params.summit_id }}"
            class="text-lg font-medium text-blue-600 hover:text-blue-800"
          >
            {{ summit.name ? `${summit.name}, ` : '' }}{{ summit.height }}м
          </RouterLink>
          <span class="text-gray-500">|</span>
          <span class="text-gray-600">хребет {{ summit.ridge.name }}</span>
        </div>
      
        <form @submit.prevent="submitClimb" class="space-y-6 mt-6">
          <fieldset :disabled="isLoading" class="space-y-6">
            <div class="bg-white">
              <div class="space-y-6">
                <div class="form-group">
                  <label for="date" class="block text-sm text-gray-500 dark:text-gray-300">Дата восхождения</label>
                  <input
                    id="date"
                    type="text"
                    v-model="formData.date"
                    class="block mt-2 w-full placeholder-gray-400/70 dark:placeholder-gray-500 rounded-lg border border-gray-200 bg-white px-5 py-2.5 text-gray-700 focus:border-blue-400 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-40 dark:border-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:focus:border-blue-300 disabled:bg-gray-100 disabled:text-gray-500"
                    placeholder="дд.мм.гггг"
                  >
                  <p class="mt-2 text-xs text-gray-500">
                    Необязательное поле. Если точная дата неизвестна, можно ввести только месяц (например 2.2012) или только год (например 2012)
                  </p>
                </div>

                <div class="form-group">
                  <label for="comment" class="block text-sm text-gray-500 dark:text-gray-300">Описание</label>
                  <textarea
                    id="comment"
                    v-model="formData.comment"
                    rows="6"
                    maxlength="1000"
                    class="block mt-2 w-full placeholder-gray-400/70 dark:placeholder-gray-500 rounded-lg border border-gray-200 bg-white px-4 h-32 py-2.5 text-gray-700 focus:border-blue-400 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-40 dark:border-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:focus:border-blue-300 disabled:bg-gray-100 disabled:text-gray-500"
                    placeholder="Расскажите о вашем восхождении..."
                  ></textarea>
                  <div class="flex justify-between mt-2">
                    <p class="text-xs text-gray-500">
                      Необязательное поле.
                    </p>
                    <p class="text-sm text-gray-500">
                      {{ formData.comment.length }}/1000
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div class="flex justify-end space-x-3">
              <button
                v-if="summit.climb_data"
                type="button"
                @click="deleteClimb"
                class="px-6 py-2 font-medium tracking-wide text-white capitalize transition-colors duration-300 transform bg-red-600 rounded-lg hover:bg-red-500 focus:outline-none focus:ring focus:ring-red-300 focus:ring-opacity-80 disabled:bg-red-400"
                :disabled="isDeleting || isLoading"
              >
                {{ isDeleting ? 'Удаление...' : 'Удалить' }}
              </button>
              <button
                type="submit"
                class="px-6 py-2 font-medium tracking-wide text-white capitalize transition-colors duration-300 transform bg-blue-600 rounded-lg hover:bg-blue-500 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-80 disabled:bg-blue-400"
                :disabled="isSubmitting || isLoading"
              >
                {{ isSubmitting ? 'Сохранение...' : 'Сохранить' }}
              </button>
            </div>
          </fieldset>
        </form>
      </div>
      
      <div v-else class="text-red-600 text-center py-4">
        Не удалось загрузить информацию о вершине
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>