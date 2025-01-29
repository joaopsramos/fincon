import api from "@/api"
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
  const resp = await api.get(`/expenses/summary`, { params: { date: dayjs(date).format("YYYY-MM-DD") } })

  return resp.data as Summary
}
