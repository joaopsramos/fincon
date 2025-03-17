import type React from "react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import type { CreateExpenseParams } from "@/api/expense"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useMutation, useQuery } from "@tanstack/react-query"
import { useTranslations } from "next-intl"
import dayjs from "dayjs"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { CalendarIcon } from "lucide-react"
import { Label } from "@/components/ui/label"
import { Goal } from "@/api/goals"
import { createExpense, findMatchingNames } from "@/api/expense"
import { useToast } from "@/hooks/use-toast"
import { PlusIcon } from "@heroicons/react/24/solid"

const createExpenseSchema = z.object({
  goal_id: z.number(),
  name: z.string().min(2, "stringMin"),
  value: z.number().gte(1, "numberGte"),
  date: z.date(),
  installments: z.number().int().gte(1, "numberGte").optional(),
})

type CreateExpenseSchema = z.infer<typeof createExpenseSchema>

export default function CreateExpense({
  selectedGoal: goal,
  goals,
  invalidateQueries,
}: {
  selectedGoal: Goal
  goals: Goal[]
  invalidateQueries: () => Promise<void>
}) {
  const commonT = useTranslations("Common")
  const errorsT = useTranslations("Errors")
  const t = useTranslations("DashboardPage.expenses")
  const { toast } = useToast()
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [nameFocused, setNameFocused] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<CreateExpenseSchema>({
    resolver: zodResolver(createExpenseSchema),
    defaultValues: {
      goal_id: goal.id,
      date: new Date(),
      installments: 1,
    },
  })

  const name = watch("name", "")
  const selectedDate = watch("date")

  const createExpenseMut = useMutation({
    mutationFn: async (data: CreateExpenseParams) => await createExpense(data),
    onSuccess: () => {
      reset()
      setIsDialogOpen(false)
      invalidateQueries()
      setValue("date", selectedDate)
      setValue("value", NaN)
    },
    onError: () =>
      toast({
        title: "Error",
        description: "Something went wrong, please try again.",
        variant: "destructive",
      }),
  })

  const handleCreateExpense = (data: CreateExpenseSchema) => {
    createExpenseMut.mutate(data)
  }

  const { data: matchingNames = [] } = useQuery({
    queryKey: ["matchingNames", name],
    queryFn: () => findMatchingNames(name),
    enabled: name?.length >= 2,
    placeholderData: [],
    refetchOnWindowFocus: false,
  })

  return (
    <>
      <div className="flex justify-center">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                size="icon"
                className="bg-slate-900 dark:bg-white rounded-full hover:bg-slate-800 dark:hover:bg-slate-200 [&_svg]:size-6"
                onClick={() => setIsDialogOpen(true)}>
                <PlusIcon className="text-white dark:text-slate-900" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{t("addTooltip")}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("addTooltip")}</DialogTitle>
          </DialogHeader>

          <form onSubmit={handleSubmit(handleCreateExpense)}>
            <div className="grid gap-4 py-2">
              <div className="grid gap-2">
                <Label htmlFor="name">{t("nameInput")}</Label>
                <div className="relative">
                  <Input
                    id="name"
                    className={`w-full ${errors.name ? "border-red-500" : ""}`}
                    {...register("name")}
                    autoComplete="off"
                    onFocus={() => setNameFocused(true)}
                    onBlur={() => setNameFocused(false)}
                    onKeyDown={(e) => {
                      if (e.key === "Tab") {
                        setNameFocused(false)
                      }
                    }}
                  />

                  {errors.name && (
                    <span className="text-red-500 text-xs">{errorsT(errors.name.message, { min: 2 })}</span>
                  )}

                  {matchingNames.length > 0 && nameFocused && (
                    <div className="absolute top-10 z-10 bg-white dark:bg-slate-900 rounded-lg mt-1 w-full max-h-40 overflow-y-auto scroll border border-slate-300 dark:border-slate-700 shadow">
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
                    {...register("value", { setValueAs: (val) => (val === "NaN" ? undefined : Number(val)) })}
                  />
                </div>

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

                {errors.value && (
                  <span className="text-red-500 text-xs">{errorsT(errors.value.message, { gte: 1 })}</span>
                )}

                {errors.installments && (
                  <span className="col-start-2 text-red-500 text-xs">
                    {errorsT(errors.installments.message, { gte: 1 })}
                  </span>
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
                        {selectedDate ? dayjs(selectedDate).utc().format("DD/MM/YY") : <span>{t("selectDate")}</span>}
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
                      {goals.map((g) => (
                        <SelectItem key={g.id} value={g.id.toString()}>
                          {commonT(g.name)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" className="mt-2 sm:mt-0" onClick={() => setIsDialogOpen(false)}>
                {commonT("cancel")}
              </Button>
              <Button type="submit">{commonT("add")}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}
