import { Money } from "@/util/money"

export type Expense = {
  id: number,
  name: string,
  value: Money,
  date: Date,
  goal_id: number,

}

export async function getExpenses({ queryKey }: { queryKey: [string, number] }) {
  const [_, goalID] = queryKey
  const resp = await fetch(`http://127.0.0.1:4000/api/goals/${goalID}/expenses`)
  const data = await resp.json()

  return data as Expense[]
}
