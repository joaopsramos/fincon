import api from "@/api"

export type Goal = {
  id: number,
  name: string,
  percentage: number,
}

export async function getGoals() {
  const resp = await api.get("/goals")
  return resp.data as Goal[]
}
