import api from "@/api";
import { setAuthCookie } from "@/lib/utils";

export async function signUp(formData: FormData) {
  try {
    const resp = await api.post("/users", {
      email: formData.get("email"),
      password: formData.get("password"),
    })

    setAuthCookie(resp.data.token)

    return
  } catch (e: any) {
    if (e.response?.data?.error) {
      throw new Error(e.response.data.error)
    }

    throw new Error("Failed to sign up. Please try again.")
  }
}

export async function login(formData: FormData) {
  try {
    const resp = await api.post("/sessions", {
      email: formData.get("email"),
      password: formData.get("password"),
    })

    setAuthCookie(resp.data.token)

    return
  } catch (e: any) {
    if (e.response?.data?.error) {
      throw new Error(e.response.data.error)
    }

    console.log(e, "oi")

    throw new Error("Failed to login. Please try again.")
  }
}
