"use client"

import type React from "react"
import { useState } from "react"
import type { Expense } from "@/api/expense"
import { useMutation, useQuery, useQueryClient, type QueryClient } from "@tanstack/react-query"
import { useTranslations } from "next-intl"
import dayjs from "dayjs"
import utc from "dayjs/plugin/utc"
import { CheckIcon, PencilIcon, TrashIcon } from "@heroicons/react/24/solid"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip"
import { Goal } from "@/api/goals"
import { deleteExpense, editExpense, getExpenses } from "@/api/expense"
import { moneyValueToString } from "@/lib/utils"
import CreateExpense from "./create_expense"
import { LoaderCircle } from "lucide-react"

export default function Expense({
  selectedGoal: goal,
  goals,
  date,
}: {
  goals: Goal[]
  selectedGoal: Goal
  date: Date
}) {
  dayjs.extend(utc)

  const commonT = useTranslations("Common")
  const t = useTranslations("DashboardPage.expenses")
  const queryClient = useQueryClient()
  const invalidateQueries = buildInvalidateQueriesFn(queryClient, date, goal.id)

  const { data: expenses, isLoading } = useQuery({
    queryKey: ["expense", date, goal.id],
    queryFn: getExpenses,
    refetchOnWindowFocus: false,
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>{commonT(goal.name)}</CardTitle>
      </CardHeader>

      <CardContent>
        {isLoading ? (
          <div className="flex justify-center items-center pb-5">
            <LoaderCircle className="animate-spin size-8" />
          </div>
        ) : (
          <div className="max-h-72 overflow-auto scroll">
            <Table withoutWrapper>
              <TableHeader className="sticky top-0 bg-slate-100 dark:bg-slate-900 rounded-full">
                <TableRow className="dark:border-slate-800">
                  <TableHead className="min-w-24 lg:w-5/12 xl:w-4/12 2xl:w-5/12">{t("expense")}</TableHead>
                  <TableHead className="min-w-24 lg:w-3/12 2xl:w-2/12">{t("amount")}</TableHead>
                  <TableHead className="min-w-20">{t("date")}</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>

              <TableBody>
                {expenses?.map((e) => (
                  <Row key={e.id} expense={e} invalidateQueries={invalidateQueries} />
                ))}

                {(!expenses || expenses.length === 0) && (
                  <TableRow>
                    <TableCell colSpan={4} className="pt-6 text-center text-gray-500 dark:text-gray-400">
                      {t("noExpenses")}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}

        <div className="mt-4">
          <CreateExpense selectedGoal={goal} goals={goals} invalidateQueries={invalidateQueries} />
        </div>
      </CardContent>
    </Card>
  )
}

function Row({ expense, invalidateQueries }: { expense: Expense; invalidateQueries: () => Promise<void> }) {
  const t = useTranslations("DashboardPage.expenses")
  const [isEditing, setIsEditing] = useState(false)
  const [name, setName] = useState(expense.name)
  const [value, setValue] = useState(expense.value.toString())

  const editExpenseMut = useMutation({
    mutationFn: () => editExpense({ name, value: Number.parseFloat(value) }, expense.id),
    onSuccess: async () => {
      await invalidateQueries()
      setIsEditing(false)
    },
  })

  const deleteExpenseMut = useMutation({
    mutationFn: () => deleteExpense(expense.id),
    onSuccess: () => {
      setIsEditing(false)
      invalidateQueries()
    },
  })

  const editExpenseOnEnter = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key == "Enter") {
      editExpenseMut.mutate()
    }
  }

  const inputClass = "w-10/12 rounded-md px-2 text-sm h-auto dark:bg-slate-800"

  return (
    <TableRow className="group hover:bg-inherit dark:border-slate-800">
      <TableCell>
        {!isEditing ? (
          <span className="py-1 inline-block">{expense.name}</span>
        ) : (
          <Input
            type="text"
            className={inputClass}
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={editExpenseOnEnter}
          />
        )}
      </TableCell>
      <TableCell>
        {!isEditing ? (
          moneyValueToString(expense.value)
        ) : (
          <Input
            type="number"
            step="0.01"
            className={inputClass}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={editExpenseOnEnter}
          />
        )}
      </TableCell>
      <TableCell>{dayjs(expense.date).utc().format("DD/MM/YY")}</TableCell>
      <TableCell>
        <div className={`flex justify-end gap-1 ${isEditing ? "" : "invisible group-hover:visible"}`}>
          {!isEditing ? (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div
                    className="cursor-pointer bg-yellow-400 rounded-full p-1 w-min hover:bg-yellow-500 transition-colors"
                    onClick={() => setIsEditing(true)}>
                    <PencilIcon className="size-4 text-white" />
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{t("editTooltip")}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          ) : (
            <div
              className="cursor-pointer bg-green-500 rounded-full p-1 w-min hover:bg-green-600 transition-colors"
              onClick={() => editExpenseMut.mutate()}>
              <CheckIcon className="size-4 text-white" />
            </div>
          )}

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <div
                  className="cursor-pointer bg-red-500 rounded-full p-1 w-min hover:bg-red-600 transition-colors"
                  onClick={() => {
                    if (window.confirm(t("deleteMsg", { name: expense.name }))) {
                      deleteExpenseMut.mutate()
                    }
                  }}>
                  <TrashIcon className="size-4 text-white" />
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <p>{t("deleteTooltip")}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </TableCell>
    </TableRow>
  )
}

function buildInvalidateQueriesFn(queryClient: QueryClient, date: Date, goalId: number) {
  return async () => {
    await queryClient.invalidateQueries({ queryKey: ["expense", date, goalId] })
    await queryClient.invalidateQueries({ queryKey: ["summary"] })
  }
}
