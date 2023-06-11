import { reactive, readonly } from 'vue'

const state = reactive({
  user: null,
})

export const useAuth = () => {
    const fetchAuthStatus = async () => {
        try {
          const response = await fetch('/api/user/me')
          const user = await response.json()
          if (response.ok) {
            state.user= user
          } else {
            state.user = null
          }
        } catch (error) {
          state.user = null
        }
      }
      
      return {
        fetchAuthStatus,
        authState: readonly(state),
      }    
 }
