import { Money } from "@/util/money"
import dayjs from "dayjs"

export type SummaryGoal = {
  name: string,
  spent: Money,
  must_spend: Money,
  used: number,
  total: number
}

export type Summary = {
  goals: SummaryGoal[],
  spent: Money,
  must_spend: Money,
  used: number
}

export async function getSummary({ queryKey }: { queryKey: [string, Date] }) {
  const [_, date] = queryKey
  const params = new URLSearchParams({ date: dayjs(date).format("YYYY-MM-DD") })
  const resp = await fetch(`http://127.0.0.1:4000/api/expenses/summary?${params}`)
  const data = await resp.json()

  return data as Summary
}
