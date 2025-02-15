import api from "@/api"
import dayjs from "dayjs"

export type SummaryGoal = {
  name: string,
  spent: number,
  must_spend: number,
  used: number,
  total: number
}

export type Summary = {
  goals: SummaryGoal[],
  spent: number,
  must_spend: number,
  used: number
}

export async function getSummary({ queryKey }: { queryKey: [string, Date] }) {
  const [_, date] = queryKey
  const resp = await api.get(`/expenses/summary`, { params: { date: dayjs(date).format("YYYY-MM-DD") } })

  return resp.data as Summary
}
