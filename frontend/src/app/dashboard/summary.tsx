import { getSummary, SummaryGoal } from "@/api/summary"
import { moneyToString } from "@/lib/utils"
import { useQuery } from "@tanstack/react-query"

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

  const thClass = "text-slate-700 pb-2"

  return (
    <div className="bg-slate-200 rounded-md p-4 pb-1 w-full">
      <h1 className="text-xl font-bold">Summary</h1>

      <div className="mt-4 overflow-auto pb-1">
        <table className="w-full text-left table-auto">
          <thead>
            <tr>
              <th className={`min-w-28 ${thClass}`}>Budget</th>
              <th className={`min-w-28 ${thClass}`}>Spent</th>
              <th className={`min-w-28 ${thClass}`}>Must spend</th>
              <th className={`min-w-20 ${thClass}`}>Used</th>
              <th className={thClass}>Total</th>
            </tr>
          </thead>

          <tbody>
            {summary?.goals.map(entry => (
              <Row entry={entry} key={entry.name} />
            ))}
          </tbody>
        </table>

        <div className="mt-6 flex gap-8">
          {summary && (
            <>
              <SummaryTotal value={moneyToString(summary.spent)} text="Total spent" valueColor="text-red-500" />
              <SummaryTotal value={moneyToString(summary.must_spend)} text="Must spent" valueColor="text-green-600" />
              <SummaryTotal value={summary.used.toFixed(2).toString() + "%"} text="Used" />
            </>
          )}
        </div>
      </div>
    </div>
  )
}

function Row({ entry }: { entry: SummaryGoal }) {
  const tdClass = "py-1 border-b border-slate-300"

  return (
    <tr>
      <td className={tdClass}>{entry.name}</td>
      <td className={tdClass}>{moneyToString(entry.spent)}</td>
      <td className={tdClass}>{moneyToString(entry.must_spend)}</td>
      <td className={tdClass}>{entry.used.toFixed(2)}%</td>
      <td className={tdClass}>{entry.total.toFixed(2)}%</td>
    </tr>
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
