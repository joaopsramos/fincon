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

export async function updateGoals(goals: Goal[], params: FormData) {
  const reqBody = goals.map(goal => {
    const newPercentage = params.get(`percentage-${goal.id}`)?.toString() ?? ""

    return { id: goal.id, percentage: Number.parseInt(newPercentage) }
  })

  const resp = await api.post("/goals", reqBody)
  return resp.data as Goal[]
}
