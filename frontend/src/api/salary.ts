import api from "@/api"

export type Salary = {
  amount: number,
}

type UpdateParams = {
  amount: number,
}

export async function getSalary() {
  const resp = await api.get("/salary")
  return resp.data as Salary
}

export async function updateSalary(params: UpdateParams) {
  const resp = await api.patch("/salary", params)
  return resp.data as Salary
}
