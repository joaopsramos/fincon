import dayjs from "dayjs"
import type { CreateExpenseParams, Expense } from "@/api/expense"
import utc from "dayjs/plugin/utc"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { CheckIcon, PencilIcon, PlusIcon, TrashIcon } from "@heroicons/react/24/solid"
import { Goal } from "@/api/goals"
import { Input } from "@/components/ui/input"
import { KeyboardEvent, useState } from "react"
import { QueryClient, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { createExpense, deleteExpense, editExpense, findMatchingNames, getExpenses } from "@/api/expense"
import { moneyValueToString } from "@/lib/utils"
import { useForm } from "react-hook-form"
import { useToast } from "@/hooks/use-toast"
import { useTranslations } from "next-intl"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle } from "lucide-react"

export default function Expense({ goal, date }: { goal: Goal; date: Date }) {
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
                    <TableCell colSpan={3} className="pt-6 text-center text-gray-500 dark:text-gray-400">
                      {t("noExpenses")}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}

        <div className="mt-4">
          <CreateExpense goal={goal} invalidateQueries={invalidateQueries} />
        </div>
      </CardContent>
    </Card >
  )
}

function Row({ expense, invalidateQueries }: { expense: Expense; invalidateQueries: () => Promise<void> }) {
  const t = useTranslations("DashboardPage.expenses")
  const [isEditing, setIsEditing] = useState(false)
  const [name, setName] = useState(expense.name)
  const [value, setValue] = useState(expense.value.toString())

  const editExpenseMut = useMutation({
    mutationFn: () => editExpense({ name, value: parseFloat(value) }, expense.id),
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

  const editExpenseOnEnter = (e: KeyboardEvent<HTMLInputElement>) => {
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
          ) : (
            <div
              className="cursor-pointer bg-green-500 rounded-full p-1 w-min hover:bg-green-600 transition-colors"
              onClick={() => editExpenseMut.mutate()}>
              <CheckIcon className="size-4 text-white" />
            </div>
          )}

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
        </div>
      </TableCell>
    </TableRow>
  )
}

const createExpenseSchema = z.object({
  name: z.string().min(2, "stringMin"),
  value: z.number().gte(0.01, "numberGte"),
})

type CreateExpenseSchema = z.infer<typeof createExpenseSchema>

function CreateExpense({ goal, invalidateQueries }: { goal: Goal; invalidateQueries: () => Promise<void> }) {
  const errorsT = useTranslations("Errors")
  const t = useTranslations("DashboardPage.expenses")
  const [nameFocused, setNameFocused] = useState(false)
  const { toast } = useToast()
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<CreateExpenseSchema>({ resolver: zodResolver(createExpenseSchema) })
  const name = watch("name", "")

  const createExpenseMut = useMutation({
    mutationFn: async (data: CreateExpenseParams) => await createExpense(data),
    onSuccess: () => {
      reset()
      invalidateQueries()
    },
    onError: () =>
      toast({
        title: "Error",
        description: "Something went wrong, please try again.",
        variant: "destructive",
      }),
  })

  const handleCreateExpense = (data: CreateExpenseSchema) => {
    const params: CreateExpenseParams = {
      ...data,
      date: new Date(),
      goal_id: goal.id,
    }

    createExpenseMut.mutate(params)
  }

  const { data: matchingNames = [] } = useQuery({
    queryKey: ["matchingNames", name],
    queryFn: () => findMatchingNames(name),
    enabled: name.length >= 2,
    placeholderData: [],
    refetchOnWindowFocus: false,
  })

  return (
    <form onSubmit={handleSubmit(handleCreateExpense)} className="relative">
      <div className="grid grid-cols-11 gap-1">
        <div className="col-span-5">
          <Input
            className={`rounded-md p-1 w-full dark:bg-slate-900 ${errors.name ? "border-red-500" : ""}`}
            {...register("name")}
            type="text"
            placeholder={t("nameInput")}
            autoComplete="off"
            onFocus={() => setNameFocused(true)}
            onBlur={() => setNameFocused(false)}
            value={name}
          />
        </div>
        <div className="col-span-5">
          <Input
            className={`rounded-md p-1 w-full dark:bg-slate-900 ${errors.value ? "border-red-500" : ""}`}
            {...register("value", { setValueAs: (val) => Number(val) })}
            type="number"
            placeholder={t("valueInput")}
            step="0.01"
          />
        </div>
        <input hidden name="goal_id" value={goal.id} readOnly />

        <Tooltip>
          <TooltipTrigger>
            <span className="block self-center w-min h-min ml-1 -mr-1 sm:ml-4 sm:mr-0 bg-slate-900 dark:bg-white rounded-full">
              <PlusIcon className="size-6 text-white dark:text-slate-900" />
            </span>
          </TooltipTrigger>
          <TooltipContent>
            <p>{t("addTooltip")}</p>
          </TooltipContent>
        </Tooltip>

        {errors.name && (
          <span className="row-start-2 col-span-5 text-red-500 text-xs">
            {errorsT(errors.name.message, { min: 2 })}
          </span>
        )}

        {errors.value && (
          <span className="row-start-2 col-start-6 col-span-5 text-red-500 text-xs">
            {errorsT(errors.value.message, { gte: 0.01 })}
          </span>
        )}
      </div>

      {matchingNames.length > 0 &&
        nameFocused && (
          <div className="absolute top-10 z-10 bg-white dark:bg-slate-900 rounded-lg mt-1 w-min text-nowrap max-h-40 overflow-y-auto scroll border border-slate-300 dark:border-slate-700 shadow">
            <ul>
              {matchingNames.map((name) => (
                <li
                  key={name}
                  className="px-2 py-1 border-b dark:border-slate-700 cursor-pointer hover:bg-gray-100 dark:hover:bg-slate-700 transition-colors"
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => {
                    setValue("name", name)
                    setNameFocused(false)
                  }}>
                  {name}
                </li>
              ))}
            </ul>
          </div>
        )}
    </form>
  )
}

function buildInvalidateQueriesFn(queryClient: QueryClient, date: Date, goalId: number) {
  return async () => {
    await queryClient.invalidateQueries({ queryKey: ["expense", date, goalId] })
    await queryClient.invalidateQueries({ queryKey: ["summary"] })
  }
}
