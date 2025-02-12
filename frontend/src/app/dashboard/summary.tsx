import { useTranslations } from "next-intl"
import { getSummary, SummaryGoal } from "@/api/summary"
import { moneyToString, sortGoals } from "@/lib/utils"
import { useQuery } from "@tanstack/react-query"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useMemo } from "react"

type SummaryTotalParams = {
  value: string,
  text: string,
  valueColor?: string
}

export default function Summary({ date }: { date: Date }) {
  const t = useTranslations("DashboardPage")
  const { data: summary } = useQuery({
    queryKey: ["summary", date],
    queryFn: getSummary,
    refetchOnWindowFocus: false
  })

  const sortedGoals = useMemo(() => {
    if (!summary?.goals) return [];

    return sortGoals(summary?.goals)
  }, [summary?.goals]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("summary.title")}</CardTitle>
      </CardHeader>

      <CardContent>
        <Table>
          <TableHeader className="bg-slate-100">
            <TableRow className="whitespace-nowrap">
              <TableHead>{t("summary.budget")}</TableHead>
              <TableHead>{t("summary.spent")}</TableHead>
              <TableHead>{t("summary.mustSpend")}</TableHead>
              <TableHead>{t("summary.used")}</TableHead>
              <TableHead>{t("summary.total")}</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {sortedGoals.map(goal => <Row key={goal.name} goal={goal} />)}
          </TableBody>
        </Table>
      </CardContent>

      <CardFooter>
        <div className="flex gap-x-4 gap-y-2 flex-wrap">
          {summary && (
            <>
              <SummaryTotal value={moneyToString(summary.spent)} text={t("summary.totalSpent")} valueColor="text-red-500" />
              <SummaryTotal value={moneyToString(summary.must_spend)} text={t("summary.mustSpend")} valueColor="text-green-600" />
              <SummaryTotal value={summary.used.toFixed(2).toString() + "%"} text={t("summary.used")} />
            </>
          )}
        </div>
      </CardFooter>
    </Card >
  )
}

function Row({ goal }: { goal: SummaryGoal }) {
  return (
    <TableRow>
      <TableCell>{goal.name}</TableCell>
      <TableCell>{moneyToString(goal.spent)}</TableCell>
      <TableCell>{moneyToString(goal.must_spend)}</TableCell>
      <TableCell>{goal.used.toFixed(2)}%</TableCell>
      <TableCell>{goal.total.toFixed(2)}%</TableCell>
    </TableRow>
  )
}

function SummaryTotal({ value, text, valueColor }: SummaryTotalParams) {
  return (
    <div>
      <p className={`text-lg font-bold -mb-2 ${valueColor}`}>{value}</p >
      <span className="text-xs">{text}</span>
    </div>
  )
}
