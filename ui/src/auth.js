import { reactive, readonly } from 'vue'

const state = reactive({
  user: null,
  isInitialized: false
})

let authInstance = null

export const useAuth = () => {
  if (!authInstance) {
    const fetchAuthStatus = async () => {
      try {
        const response = await fetch('/api/user/me')
        if (response.ok) {
          const user = await response.json()
          state.user = user
        } else {
          state.user = null
        }
      } catch (error) {
        state.user = null
      } finally {
        state.isInitialized = true
      }
    }
    
    authInstance = {
      fetchAuthStatus,
      authState: readonly(state),
    }
  }
  
  return authInstance
}
