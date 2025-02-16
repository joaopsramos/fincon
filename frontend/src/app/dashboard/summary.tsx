"use client"

import { useTranslations } from "next-intl"
import type { Summary } from "@/api/summary"
import { SummaryGoal } from "@/api/summary"
import { moneyValueToString, sortGoals } from "@/lib/utils"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useMemo } from "react"

type SummaryTotalParams = {
  value: string
  text: string
  valueColor?: string
}

export default function Summary({ summary }: { summary: Summary | undefined }) {
  const t = useTranslations("DashboardPage")

  const sortedGoals = useMemo(() => {
    if (!summary?.goals) return []

    return sortGoals(summary?.goals)
  }, [summary?.goals])

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("summary.title")}</CardTitle>
      </CardHeader>

      <CardContent>
        <Table>
          <TableHeader className="bg-slate-100 dark:bg-slate-900">
            <TableRow className="whitespace-nowrap hover:bg-slate-100 dark:hover:bg-slate-900 dark:border-slate-800">
              <TableHead>{t("summary.budget")}</TableHead>
              <TableHead>{t("summary.spent")}</TableHead>
              <TableHead>{t("summary.mustSpend")}</TableHead>
              <TableHead>{t("summary.used")}</TableHead>
              <TableHead>{t("summary.total")}</TableHead>
            </TableRow>
          </TableHeader >

          <TableBody>
            {sortedGoals.map((goal) => (
              <Row key={goal.name} goal={goal} />
            ))}
          </TableBody>
        </Table >
      </CardContent >

      <CardFooter>
        <div className="flex gap-x-4 gap-y-2 flex-wrap">
          {summary && (
            <>
              <SummaryTotal
                value={moneyValueToString(summary.spent)}
                text={t("summary.totalSpent")}
                valueColor="text-red-500"
              />
              <SummaryTotal
                value={moneyValueToString(summary.must_spend)}
                text={t("summary.mustSpend")}
                valueColor={summary.must_spend > 0 ? "text-green-500" : "text-red-500"}
              />
              <SummaryTotal value={summary.used.toFixed(2).toString() + "%"} text={t("summary.used")} />
            </>
          )}
        </div>
      </CardFooter>
    </Card >
  )
}

function Row({ goal }: { goal: SummaryGoal }) {
  const commonT = useTranslations("Common")

  return (
    <TableRow className="dark:border-slate-800">
      <TableCell>{commonT(goal.name)}</TableCell>
      <TableCell>{moneyValueToString(goal.spent)}</TableCell>
      <TableCell>{moneyValueToString(goal.must_spend)}</TableCell>
      <TableCell>{goal.used.toFixed(2)}%</TableCell>
      <TableCell>{goal.total.toFixed(2)}%</TableCell>
    </TableRow>
  )
}

function SummaryTotal({ value, text, valueColor }: SummaryTotalParams) {
  return (
    <div>
      <p className={`text-lg font-bold -mb-2 ${valueColor}`}>{value}</p>
      <span className="text-xs">{text}</span>
    </div>
  )
}
