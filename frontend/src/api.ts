import axios from "axios"
import { deleteAuthCookie, getAuthCookie } from "./lib/utils"

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
})

api.interceptors.request.use(
  async (config) => {
    const token = getAuthCookie()

    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error) => Promise.reject(error)
)

api.interceptors.response.use(
  (resp) => resp,
  (error) => {
    if (error.response?.status === 401) {
      deleteAuthCookie()
      window.location.href = "/login"
    }

    return Promise.reject(error)
  }
)

export default api
