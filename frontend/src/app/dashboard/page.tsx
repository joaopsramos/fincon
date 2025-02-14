"use client"

import { useQuery } from "@tanstack/react-query"
import Goals from "./goals"
import Summary from "./summary"
import { getGoals } from "@/api/goals"
import Expense from "./expense"
import { useNow } from "next-intl"
import { Suspense, useMemo } from "react"
import { sortGoals } from "@/lib/utils"
import Header from "./header"
import Menu from "@/components/menu"
import { useSearchParams } from "next/navigation"

export default function Dashboard() {
  return (
    <>
      <div className="m-4">
        <Suspense>
          <Content />
        </Suspense>
      </div>

      <Menu />
    </>
  )
}

function Content() {
  const searchParams = useSearchParams()

  const queryYear = searchParams.get("year") || new Date().getFullYear()
  const queryMonth = searchParams.get("month") || new Date().getMonth() + 1

  const date = useNow()
  date.setFullYear(Number(queryYear))
  date.setMonth(Number(queryMonth) - 1)

  const { data: goals } = useQuery({
    queryKey: ["goals"],
    queryFn: getGoals,
    refetchOnWindowFocus: false,
  })

  const sortedGoals = useMemo(() => {
    if (!goals) return []

    return sortGoals(goals)
  }, [goals])


  return (
    <>
      <div className="mb-2">
        <Header date={date} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-2">
        <div className="lg:col-span-2">
          <Summary date={date} />
        </div>
        <div className="">
          <Goals goals={sortedGoals || []} />
        </div>
      </div>

      <div className="mt-2 grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-2">
        {sortedGoals?.map((goal) => (
          <Expense key={goal.id} goal={goal} date={date} />
        ))}
      </div>
    </>
  )
}
