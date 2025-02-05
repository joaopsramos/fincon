import { Money } from "@/lib/utils"
import api from "@/api"
import dayjs from "dayjs"

export type Expense = {
  id: number,
  name: string,
  value: Money,
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

export async function createExpense(formData: FormData, goalId: number) {
  await api.post("/expenses", {
    name: formData.get("name"),
    value: parseFloat(formData.get("value") as string),
    date: dayjs().format("YYYY-MM-DD"),
    goal_id: goalId,
  })
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
