import Form from "next/form"
import dayjs from "dayjs"
import type { Expense } from "@/api/expense"
import utc from "dayjs/plugin/utc"
import { Goal } from "@/api/goals"
import { createExpense, deleteExpense, editExpense, findMatchingNames, getExpenses } from "@/api/expense"
import { moneyToString } from "@/lib/utils"
import { QueryClient, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { CheckIcon, PencilIcon, PlusCircleIcon, PlusIcon, TrashIcon } from "@heroicons/react/20/solid"
import { KeyboardEvent, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Input } from "@/components/ui/input"

export default function Expense({ goal, date }: { goal: Goal, date: Date }) {
  dayjs.extend(utc)

  const queryClient = useQueryClient()
  const invalidateQueries = buildInvalidateQueriesFn(queryClient, date, goal.id)

  const { data: expenses } = useQuery({
    queryKey: ["expense", date, goal.id],
    queryFn: getExpenses,
    refetchOnWindowFocus: false
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>{goal.name}</CardTitle>
      </CardHeader>

      <CardContent>
        <div className="max-h-72 overflow-auto scroll">
          <Table withoutWrapper>
            <TableHeader className="sticky top-0 bg-slate-100">
              <TableRow>
                <TableHead className="min-w-24 lg:w-5/12 xl:w-4/12 2xl:w-5/12">Expense</TableHead>
                <TableHead className="min-w-24 lg:w-3/12 2xl:w-2/12">Amount</TableHead>
                <TableHead className="min-w-20">Date</TableHead>
                <TableHead></TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              {expenses?.map(e => (
                <Row key={e.id} expense={e} invalidateQueries={invalidateQueries} />
              ))}
            </TableBody>
          </Table>
        </div>

        <div className="mt-4">
          <CreateExpense goal={goal} invalidateQueries={invalidateQueries} />
        </div>
      </CardContent>
    </Card>
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
    onSuccess: () => {
      setIsEditing(false)
      invalidateQueries()
    }
  })

  const editExpenseOnEnter = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key == "Enter") {
      editExpenseMut.mutate()
    }
  }

  const inputClass = "w-10/12 rounded-md px-2 text-sm h-auto"

  return (
    <TableRow className="group hover:bg-inherit">
      <TableCell>
        {!isEditing ? <span className="py-1 inline-block">{expense.name}</span> : (
          <Input
            type="text"
            className={inputClass}
            value={name}
            onChange={e => setName(e.target.value)}
            onKeyDown={editExpenseOnEnter}
          />
        )}
      </TableCell>
      <TableCell>
        {!isEditing ? moneyToString(expense.value) : (
          <Input
            type="number"
            step="0.01"
            className={inputClass}
            value={value}
            onChange={e => setValue(e.target.value)}
            onKeyDown={editExpenseOnEnter}
          />
        )}
      </TableCell>
      <TableCell>{dayjs(expense.date).utc().format("DD/MM/YY")}</TableCell>
      <TableCell>
        <div className={`flex justify-end gap-1 ${isEditing ? "" : "invisible group-hover:visible"}`}>
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
      </TableCell>
    </TableRow>
  )
}

function CreateExpense({ goal, invalidateQueries }: { goal: Goal, invalidateQueries: () => Promise<void> }) {
  const [name, setName] = useState("")
  const [nameFocused, setNameFocused] = useState(false)

  const createExpenseMut = useMutation({
    mutationFn: (formData: FormData) => createExpense(formData, goal.id),
    onSuccess: () => {
      setName("")
      invalidateQueries()
    }
  })

  const { data: matchingNames = [] } = useQuery({
    queryKey: ['matchingNames', name],
    queryFn: () => findMatchingNames(name),
    enabled: name.length >= 2,
    placeholderData: [],
    refetchOnWindowFocus: false
  })

  return (
    <Form action={createExpenseMut.mutate}>
      <div className="flex items-center">
        <Input
          className="rounded-md p-1 w-6/12"
          name="name"
          type="text"
          placeholder="Name"
          autoComplete="off"
          onFocus={() => setNameFocused(true)}
          onBlur={() => setNameFocused(false)}
          onChange={e => setName(e.target.value)}
          value={name}
        />
        <Input className="ml-2 rounded-md p-1 w-6/12" name="value" type="number" placeholder="Value" step="0.01" />
        <input hidden name="goal_id" value={goal.id} readOnly />

        <button type="submit" className="ml-1 -mr-1 sm:ml-4 sm:mr-0 bg-slate-900 rounded-full">
          <PlusIcon className="size-6 text-white" />
        </button>
      </div>
      <div className="absolute z-10 bg-white rounded-lg mt-1 w-min text-nowrap max-h-40 overflow-y-auto scroll">
        <ul>
          {matchingNames.length > 0 && nameFocused && matchingNames.map(name => (
            <li
              key={name}
              className="px-2 py-1 border-b cursor-pointer hover:bg-gray-100 transition-colors"
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => {
                setName(name)
                setNameFocused(false)
              }}>{name}</li>
          ))}
        </ul>
      </div>
    </Form>
  )
}

function buildInvalidateQueriesFn(queryClient: QueryClient, date: Date, goalId: number) {
  return async () => {
    await queryClient.invalidateQueries({ queryKey: ["expense", date, goalId] })
    await queryClient.invalidateQueries({ queryKey: ["summary"] })
  }
}
