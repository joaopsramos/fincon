import { getSummary, SummaryGoal } from "@/api/summary"
import { moneyToString } from "@/lib/utils"
import { useQuery } from "@tanstack/react-query"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

type SummaryTotalParams = {
  value: string,
  text: string,
  valueColor?: string
}

export default function Summary({ date }: { date: Date }) {
  const { data: summary } = useQuery({
    queryKey: ["summary", date],
    queryFn: getSummary,
    refetchOnWindowFocus: false
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>Summary</CardTitle>
      </CardHeader>

      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Budget</TableHead>
              <TableHead>Spent</TableHead>
              <TableHead>Must spend</TableHead>
              <TableHead>Used</TableHead>
              <TableHead>Total</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {summary?.goals.map(goal => <Row key={goal.name} goal={goal} />)}
          </TableBody>
        </Table>
      </CardContent>

      <CardFooter>
        <div className="flex gap-x-4 gap-y-2 flex-wrap">
          {summary && (
            <>
              <SummaryTotal value={moneyToString(summary.spent)} text="Total spent" valueColor="text-red-500" />
              <SummaryTotal value={moneyToString(summary.must_spend)} text="Must spent" valueColor="text-green-600" />
              <SummaryTotal value={summary.used.toFixed(2).toString() + "%"} text="Used" />
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
