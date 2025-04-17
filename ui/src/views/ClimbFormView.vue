<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isSubmitting = ref(false)

const climbData = ref({
  date: new Date().toISOString().split('T')[0],
  description: '',
  photos: []
})

function handleFileUpload(event) {
  climbData.value.photos = Array.from(event.target.files)
}

async function submitClimb() {
  if (isSubmitting.value) return
  
  isSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('date', climbData.value.date)
    formData.append('description', climbData.value.description)
    climbData.value.photos.forEach(photo => {
      formData.append('photos', photo)
    })

    const response = await fetch(`/api/summits/${route.params.ridge_id}/${route.params.summit_id}/climbs`, {
      method: 'POST',
      body: formData
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
</script>

<template>
  <div class="climb-form-view">
    <div class="max-w-screen-md mx-auto">
      <h1 class="text-2xl font-bold mb-6">Регистрация восхождения</h1>
      
      <form @submit.prevent="submitClimb" class="space-y-4">
        <div class="form-group">
          <label for="date" class="block text-sm font-medium text-gray-700">Дата восхождения</label>
          <input
            id="date"
            type="date"
            v-model="climbData.date"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            required
          >
        </div>

        <div class="form-group">
          <label for="description" class="block text-sm font-medium text-gray-700">Описание</label>
          <textarea
            id="description"
            v-model="climbData.description"
            rows="4"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            placeholder="Расскажите о вашем восхождении..."
          ></textarea>
        </div>

        <div class="form-group">
          <label for="photos" class="block text-sm font-medium text-gray-700">Фотографии</label>
          <input
            id="photos"
            type="file"
            multiple
            accept="image/*"
            @change="handleFileUpload"
            class="mt-1 block w-full text-sm text-gray-500
              file:mr-4 file:py-2 file:px-4
              file:rounded-md file:border-0
              file:text-sm file:font-semibold
              file:bg-blue-50 file:text-blue-700
              hover:file:bg-blue-100"
          >
        </div>

        <div class="flex justify-end">
          <button
            type="submit"
            class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? 'Сохранение...' : 'Сохранить' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.climb-form-view {
  padding: 20px;
}
</style>
