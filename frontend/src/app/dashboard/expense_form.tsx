import type React from "react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import type { CreateExpenseParams, Expense, UpdateExpenseParams } from "@/api/expense"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslations } from "next-intl"
import dayjs from "dayjs"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { CalendarIcon } from "lucide-react"
import { Label } from "@/components/ui/label"
import { Goal } from "@/api/goals"
import { createExpense, updateExpense, findMatchingNames } from "@/api/expense"
import { useToast } from "@/hooks/use-toast"

const formSchema = z.object({
  name: z.string().min(2, "stringMin"),
  value: z.number().gte(1, "numberGte"),
  date: z.date(),
  goal_id: z.number(),
  installments: z.number().int().gte(1, "numberGte").optional(),
})

type FormSchema = z.infer<typeof formSchema>

export type ExpenseFormProps = {
  date: Date
  goal: Goal
  allGoals: Goal[]
  expense?: Expense
  invalidateQueries: () => Promise<void>
  onSuccess?: () => void
}

export default function ExpenseForm({ date, goal, allGoals, expense, invalidateQueries, onSuccess }: ExpenseFormProps) {
  const commonT = useTranslations("Common")
  const errorsT = useTranslations("Errors")
  const t = useTranslations("DashboardPage.expenses")
  const { toast } = useToast()
  const [nameFocused, setNameFocused] = useState(false)
  const queryClient = useQueryClient()

  const expenseDate = expense && new Date(expense.date)

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<FormSchema>({
    resolver: zodResolver(formSchema),
    defaultValues: expense
      ? {
          goal_id: expense.goal_id,
          date:
            expenseDate && new Date(expenseDate.getUTCFullYear(), expenseDate.getUTCMonth(), expenseDate.getUTCDate()),
          name: expense.name,
          value: expense.value,
          installments: 1,
        }
      : {
          goal_id: goal.id,
          date: new Date(),
          installments: 1,
        },
  })

  const name = watch("name", "")
  const selectedDate = watch("date")

  const upsertExpenseMut = useMutation({
    mutationFn: async (data: CreateExpenseParams | UpdateExpenseParams) => {
      if (expense) {
        return await updateExpense(data, expense.id)
      }
      return await createExpense(data as CreateExpenseParams)
    },
    onSuccess: (_, params) => {
      invalidateQueries()
      if (params.goal_id != goal.id) {
        queryClient.invalidateQueries({ queryKey: ["expense", date, params.goal_id] })
      }

      reset()
      onSuccess?.()
      setValue("date", selectedDate)
    },
    onError: () =>
      toast({
        title: "Error",
        description: "Something went wrong, please try again.",
        variant: "destructive",
      }),
  })

  const handleCreateExpense = (data: FormSchema) => {
    upsertExpenseMut.mutate(data)
  }

  const { data: matchingNames = [] } = useQuery({
    queryKey: ["matchingNames", name],
    queryFn: () => findMatchingNames(name),
    enabled: name?.length >= 2,
    placeholderData: [],
    refetchOnWindowFocus: false,
  })

  return (
    <form id="upsert-form" onSubmit={handleSubmit(handleCreateExpense)}>
      <div className="grid gap-4 py-2">
        <div className="grid gap-2">
          <Label htmlFor="name">{t("nameInput")}</Label>
          <div className="relative">
            <Input
              {...register("name")}
              id="name"
              className={`w-full ${errors.name ? "border-red-500" : ""}`}
              autoComplete="off"
              onFocus={() => setNameFocused(true)}
              onBlur={() => setNameFocused(false)}
              onKeyDown={(e) => {
                if (e.key === "Tab") {
                  setNameFocused(false)
                }
              }}
            />

            {errors.name && <span className="text-red-500 text-xs">{errorsT(errors.name.message, { min: 2 })}</span>}

            {
              // matchingNames.length > 0 && nameFocused && (
              // <div className="absolute top-10 z-10 bg-white dark:bg-slate-900 rounded-lg mt-1 w-full max-h-40 overflow-y-auto scroll border border-slate-300 dark:border-slate-700 shadow">
              //   <ul>
              //     {matchingNames.map((name) => (
              //       <li
              //         key={name}
              //         className="px-2 py-1 border-b dark:border-slate-700 cursor-pointer hover:bg-gray-100 dark:hover:bg-slate-700 transition-colors"
              //         onMouseDown={(e) => e.preventDefault()}
              //         onClick={() => {
              //           setValue("name", name)
              //           setNameFocused(false)
              //         }}>
              //         {name}
              //       </li>
              //     ))}
              //   </ul>
              // </div>
              //)
            }
          </div>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <div className="grid gap-2">
            <Label htmlFor="value">{t("valueInput")}</Label>
            <Input
              id="value"
              type="number"
              step="0.01"
              className={`w-full ${errors.value ? "border-red-500" : ""}`}
              {...register("value", { setValueAs: (val) => Number(val) })}
            />
          </div>

          {!expense && (
            <div className="grid gap-2">
              <Label htmlFor="installments">{t("installmentsInput")}</Label>
              <Input
                id="installments"
                type="number"
                className={`w-full ${errors.installments ? "border-red-500" : ""}`}
                {...register("installments", {
                  setValueAs: (val) => (val ? Number(val) : undefined),
                })}
              />
            </div>
          )}

          {errors.value && <span className="text-red-500 text-xs">{errorsT(errors.value.message, { gte: 1 })}</span>}

          {errors.installments && (
            <span className="col-start-2 text-red-500 text-xs">{errorsT(errors.installments.message, { gte: 1 })}</span>
          )}
        </div>

        <div className="grid grid-cols-2 gap-2">
          <div className="grid gap-2">
            <Label htmlFor="date">{t("date")}</Label>
            <Popover modal>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className={`w-full justify-start text-left font-normal ${errors.date ? "border-red-500" : ""}`}>
                  <CalendarIcon className="h-4 w-4" />
                  {selectedDate ? dayjs(selectedDate).format("DD/MM/YY") : <span>{t("selectDate")}</span>}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0">
                <Calendar
                  mode="single"
                  selected={selectedDate}
                  onSelect={(date) => date && setValue("date", date as Date)}
                  disabled={(date) => date > new Date() || date < new Date("1900-01-01")}
                />
              </PopoverContent>
            </Popover>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="goal">{t("goalInput")}</Label>
            <Select
              defaultValue={goal.id.toString()}
              onValueChange={(value) => setValue("goal_id", Number.parseInt(value))}>
              <SelectTrigger>
                <SelectValue placeholder={t("goalInputPlaceholder")} />
              </SelectTrigger>
              <SelectContent>
                {allGoals.map((g) => (
                  <SelectItem key={g.id} value={g.id.toString()}>
                    {commonT(g.name)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>
    </form>
  )
}
