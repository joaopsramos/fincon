export type Goal = {
  id: number,
  name: string,
  percentage: number,
}

export async function getGoals() {
  const resp = await fetch("http://127.0.0.1:4000/api/goals")
  const data = await resp.json()

  return data as Goal[]
}
