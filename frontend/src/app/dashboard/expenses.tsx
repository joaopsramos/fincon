"use client"

import type React from "react"
import { useRef, useState } from "react"
import type { Expense } from "@/api/expense"
import { useMutation, useQuery, useQueryClient, type QueryClient } from "@tanstack/react-query"
import { useTranslations } from "next-intl"
import dayjs from "dayjs"
import utc from "dayjs/plugin/utc"
import { PencilIcon, TrashIcon } from "@heroicons/react/24/solid"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip"
import { Goal } from "@/api/goals"
import { deleteExpense, getExpenses } from "@/api/expense"
import { moneyValueToString } from "@/lib/utils"
import { LoaderCircle } from "lucide-react"
import UpsertExpenseDialog, { UpsertExpenseDialogRef } from "./upsert_expense_dialog"

type ExpensesProps = {
  allGoals: Goal[]
  goal: Goal
  date: Date
}

export default function Expenses({ goal, allGoals, date }: ExpensesProps) {
  dayjs.extend(utc)

  const commonT = useTranslations("Common")
  const t = useTranslations("DashboardPage.expenses")
  const queryClient = useQueryClient()
  const invalidateQueries = buildInvalidateQueriesFn(queryClient, date, goal.id)
  const [expenseToEdit, setExpenseToEdit] = useState<Expense>()
  const dialogRef = useRef<UpsertExpenseDialogRef>(null)

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
                  <Row
                    key={e.id}
                    expense={e}
                    invalidateQueries={invalidateQueries}
                    setExpenseToEdit={setExpenseToEdit}
                    dialogRef={dialogRef}
                  />
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
          <UpsertExpenseDialog
            ref={dialogRef}
            date={date}
            goal={goal}
            allGoals={allGoals}
            expense={expenseToEdit}
            onDialogClose={() => setExpenseToEdit(undefined)}
            invalidateQueries={invalidateQueries}
          />
        </div>
      </CardContent>
    </Card>
  )
}

type RowProps = {
  expense: Expense
  setExpenseToEdit: React.Dispatch<React.SetStateAction<Expense | undefined>>
  dialogRef: React.RefObject<UpsertExpenseDialogRef | null>
  invalidateQueries: () => Promise<void>
}

function Row({ expense, dialogRef, setExpenseToEdit, invalidateQueries }: RowProps) {
  const t = useTranslations("DashboardPage.expenses")

  const deleteExpenseMut = useMutation({
    mutationFn: () => deleteExpense(expense.id),
    onSuccess: () => {
      invalidateQueries()
    },
  })

  return (
    <TableRow className="group hover:bg-inherit dark:border-slate-800">
      <TableCell>
        <span className="py-1 inline-block">{expense.name}</span>
      </TableCell>
      <TableCell>{moneyValueToString(expense.value)}</TableCell>
      <TableCell>{dayjs(expense.date).utc().format("DD/MM/YY")}</TableCell>
      <TableCell>
        <div className={"flex justify-end gap-1 invisible group-hover:visible"}>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <div
                  className="cursor-pointer bg-yellow-400 rounded-full p-1 w-min hover:bg-yellow-500 transition-colors"
                  onClick={() => {
                    setExpenseToEdit(expense)
                    dialogRef.current?.openDialog()
                  }}>
                  <PencilIcon className="size-4 text-white" />
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <p>{t("editTooltip")}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>

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
