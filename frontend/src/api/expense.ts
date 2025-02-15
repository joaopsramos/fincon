import api from "@/api"
import dayjs from "dayjs"

export type Expense = {
  id: number,
  name: string,
  value: number,
  date: Date,
  goal_id: number,

}

export type CreateExpenseParams = {
  name: string,
  value: number,
  date: Date,
  goal_id: number,
}

type EditParams = {
  name: string,
  value: number,
}

export async function getExpenses({ queryKey }: { queryKey: [string, Date, number] }) {
  const [_, date, goalID] = queryKey
  const resp = await api.get(`/goals/${goalID}/expenses?year=${date.getFullYear()}&month=${date.getMonth() + 1}`)

  return resp.data as Expense[]
}

export async function createExpense(params: CreateExpenseParams) {
  await api.post("/expenses", { ...params, date: dayjs(params.date).format("YYYY-MM-DD") })
}

export async function editExpense({ name, value }: EditParams, expenseId: number) {
  await api.patch(`/expenses/${expenseId}`, {
    name: name,
    value: value,
  })
}

export async function deleteExpense(expenseId: number) {
  await api.delete(`/expenses/${expenseId}`)
}

export async function findMatchingNames(query: string) {
  const resp = await api.get(`/expenses/matching-names?query=${query}`)
  return resp.data as string[]
}
