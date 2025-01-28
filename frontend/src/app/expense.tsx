import type { Expense } from "@/api/expense"
import { getExpenses } from "@/api/expense"
import { Goal } from "@/api/goals"
import { moneyToString } from "@/util/money"
import { useQuery } from "@tanstack/react-query"
import dayjs from "dayjs"

function Row({ expense }: { expense: Expense }) {
  const tdClass = "py-1 border-b border-slate-300"

  return (
    <tr>
      <td className={tdClass}>{expense.name}</td>
      <td className={tdClass}>{moneyToString(expense.value)}</td>
      <td className={tdClass}>{dayjs(expense.date).format("DD/MM/YY")}</td>
    </tr>
  )
}

export default function Expense({ goal }: { goal: Goal }) {
  const { data: expenses } = useQuery({
    queryKey: ["expense", goal.id],
    queryFn: getExpenses
  })

  const thClass = "text-slate-700 pb-2"

  return (
    <div className="bg-slate-200 rounded-md p-4 h-full">
      <h1 className="text-xl font-bold">{goal.name}</h1>

      <div className="mt-4 max-h-72 overflow-y-auto
        [&::-webkit-scrollbar]:w-1
        [&::-webkit-scrollbar-track]:bg-none
        [&::-webkit-scrollbar-thumb]:bg-gray-300
      ">
        <table className="w-full text-left table-auto">
          <thead>
            <tr>
              <th className={thClass}>Custos</th>
              <th className={thClass}>Valor gasto</th>
              <th className={thClass}>Data</th>
            </tr>
          </thead>

          <tbody>
            {expenses?.map(e => (
              <Row expense={e} key={e.id} />
            ))}
          </tbody>
        </table>
      </div>
    </div >
  )
}

