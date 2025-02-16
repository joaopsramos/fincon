"use client"

import { useQuery } from "@tanstack/react-query"
import Goals from "./goals"
import Summary from "./summary"
import { getGoals } from "@/api/goals"
import Expense from "./expense"
import { useNow } from "next-intl"
import { Suspense, useEffect, useMemo, useState } from "react"
import { sortGoals } from "@/lib/utils"
import Header from "./header"
import Menu from "@/components/menu"
import { useSearchParams } from "next/navigation"
import { getSummary } from "@/api/summary"
import { LoaderCircle } from "lucide-react"

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
  const [isLoading, setIsLoading] = useState(true)

  const queryYear = searchParams.get("year") || new Date().getFullYear()
  const queryMonth = searchParams.get("month") || new Date().getMonth() + 1

  const date = useNow()
  date.setFullYear(Number(queryYear))
  date.setMonth(Number(queryMonth) - 1)

  const { data: goals, isLoading: isLoadingGoals } = useQuery({
    queryKey: ["goals"],
    queryFn: getGoals,
    refetchOnWindowFocus: false,
  })

  const { data: summary, isLoading: isLoadingSummary } = useQuery({
    queryKey: ["summary", date],
    queryFn: getSummary,
    refetchOnWindowFocus: false,
  })

  const sortedGoals = useMemo(() => {
    if (!goals) return []

    return sortGoals(goals)
  }, [goals])

  useEffect(() => {
    setIsLoading(isLoadingGoals || isLoadingSummary)
  }, [isLoadingGoals, isLoadingSummary])

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-dvh">
        <LoaderCircle className="animate-spin size-10" />
      </div>
    )
  }

  return (
    <>
      <div className="mb-2">
        <Header date={date} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-2">
        <div className="lg:col-span-2">
          <Summary summary={summary} />
        </div>
        <div>
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
