import Form from "next/form"
import dayjs from "dayjs"
import type { Expense } from "@/api/expense"
import utc from "dayjs/plugin/utc"
import { Goal } from "@/api/goals"
import { createExpense, deleteExpense, editExpense, getExpenses } from "@/api/expense"
import { moneyToString } from "@/util/money"
import { QueryClient, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { CheckIcon, PencilIcon, PlusCircleIcon, TrashIcon } from "@heroicons/react/20/solid"
import { KeyboardEvent, useState } from "react"

export default function Expense({ goal }: { goal: Goal }) {
  dayjs.extend(utc)

  const queryClient = useQueryClient()
  const invalidateQueries = buildInvalidateQueriesFn(queryClient, goal.id)

  const { data: expenses } = useQuery({
    queryKey: ["expense", goal.id],
    queryFn: getExpenses,
    refetchOnWindowFocus: false
  })

  const createExpenseMut = useMutation({
    mutationFn: (formData: FormData) => createExpense(formData, goal.id),
    onSuccess: invalidateQueries
  })

  const thClass = "text-slate-700 pb-2"

  return (
    <div className="bg-slate-200 rounded-md p-4 h-full">
      <h1 className="text-xl font-bold">{goal.name}</h1>

      <div>
        <div className="mt-4 max-h-72 overflow-y-auto
        [&::-webkit-scrollbar]:w-1
        [&::-webkit-scrollbar-track]:bg-none
        [&::-webkit-scrollbar-thumb]:bg-gray-300
      ">
          <table className="w-full text-left table-fixed">
            <thead className="sticky top-0 bg-slate-200">
              <tr>
                <th className={`w-5/12 ${thClass}`}>Expense</th>
                <th className={`w-2/12 ${thClass}`}>Amount</th>
                <th className={`w-2/12 ${thClass}`}>Date</th>
                <th className="w-1/12"></th>
              </tr>
            </thead>

            <tbody>
              {expenses?.map(e => (
                <Row key={e.id} expense={e} invalidateQueries={invalidateQueries} />
              ))}
            </tbody>
          </table>
        </div>

        <div className="mt-4">
          <Form action={createExpenseMut.mutate}>
            <div className="flex items-center">
              <input className="rounded-md p-1 w-6/12" name="name" type="text" placeholder="Name" />
              <input className="ml-12 rounded-md p-1 w-6/12" name="value" type="number" placeholder="Value" step="0.01" />
              <input hidden name="goal_id" value={goal.id} readOnly />

              <button type="submit" className="ml-4">
                <PlusCircleIcon className="size-9 text-sky-500" />
              </button>
            </div>
          </Form>
        </div>

      </div>
    </div >
  )
}

function Row({ expense, invalidateQueries }: { expense: Expense, invalidateQueries: () => Promise<void> }) {
  const [isEditing, setIsEditing] = useState(false);
  const [name, setName] = useState(expense.name);
  const [value, setValue] = useState(expense.value.amount.toString());


  const editExpenseMut = useMutation({
    mutationFn: () => editExpense({ name, value: parseFloat(value) }, expense.id),
    onSuccess: async () => {
      await invalidateQueries()
      setIsEditing(false)
    }
  })

  const deleteExpenseMut = useMutation({
    mutationFn: () => deleteExpense(expense.id),
    onSuccess: invalidateQueries
  })

  const editExpenseOnEnter = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key == "Enter") {
      editExpenseMut.mutate()
    }
  }

  const tdClass = "py-1 border-b border-slate-300"
  const inputClass = "w-10/12 rounded-md p-1 text-sm"

  return (
    <tr className="group">
      <td className={tdClass}>
        {!isEditing ? expense.name : (
          <input
            type="text"
            className={inputClass}
            value={name}
            onChange={e => setName(e.target.value)}
            onKeyDown={editExpenseOnEnter}
          />
        )}
      </td>
      <td className={tdClass}>
        {!isEditing ? moneyToString(expense.value) : (
          <input
            type="number"
            step="0.01"
            className={inputClass}
            value={value}
            onChange={e => setValue(e.target.value)}
            onKeyDown={editExpenseOnEnter}
          />
        )}
      </td>
      <td className={tdClass}>{dayjs(expense.date).utc().format("DD/MM/YY")}</td>
      <td className={tdClass}>
        <div className={`flex gap-1 ${isEditing ? "" : "invisible group-hover:visible"}`}>
          {!isEditing ? (
            <div
              className="cursor-pointer bg-yellow-400 rounded-full p-1 w-min hover:bg-yellow-500 transition-colors"
              onClick={() => setIsEditing(true)}
            >
              <PencilIcon className="size-4 text-white" />
            </div>
          ) : (
            <div
              className="cursor-pointer bg-green-500 rounded-full p-1 w-min hover:bg-green-600 transition-colors"
              onClick={() => editExpenseMut.mutate()}
            >
              <CheckIcon className="size-4 text-white" />
            </div>
          )}
          <div
            className="cursor-pointer bg-red-500 rounded-full p-1 w-min hover:bg-red-600 transition-colors"
            onClick={() => {
              if (window.confirm(`Do you want to delete the expense "${expense.name}"?`)) {
                deleteExpenseMut.mutate();
              }
            }}
          >
            <TrashIcon className="size-4 text-white" />
          </div>
        </div>
      </td>
    </tr>
  )
}

function buildInvalidateQueriesFn(queryClient: QueryClient, goalId: number) {
  return async () => {
    await queryClient.invalidateQueries({ queryKey: ["expense", goalId] })
    await queryClient.invalidateQueries({ queryKey: ["summary"] })
  }
}
