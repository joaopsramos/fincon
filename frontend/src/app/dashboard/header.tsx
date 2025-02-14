import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { ArrowRightStartOnRectangleIcon } from "@heroicons/react/24/outline"
import { CheckIcon, PencilIcon } from "@heroicons/react/24/solid"
import { useTranslations } from "next-intl"
import { useCallback, useState } from "react"
import { handleLogout } from "@/lib/utils"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { getSalary, updateSalary } from "@/api/salary"
import { z } from "zod"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useToast } from "@/hooks/use-toast"
import { LoaderCircle } from "lucide-react"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

const salarySchema = z.object({
  amount: z.number().min(1, "valuteTooLow"),
})

type SalarySchema = z.infer<typeof salarySchema>

const INITIAL_YEAR = 2025

export default function Header({ date }: { date: Date }) {
  const router = useRouter()
  const menuT = useTranslations("Menu")

  return (
    <div className="grid grid-flow-col grid-rows-2 gap-2 grid-cols-2 md:grid-cols-3 md:grid-rows-1">
      <div className="md:row-span-2">
        <Salary />
      </div>

      <div className="justify-self-start md:row-span-2 md:justify-self-center">
        <DateSelector date={date} />
      </div>

      <div className="hidden min-[375px]:block justify-self-end col-span-2">
        <Button size={"sm"} onClick={() => handleLogout(router)}>
          <ArrowRightStartOnRectangleIcon className="size-5 text-white" />
          {menuT("logout")}
        </Button>
      </div>
    </div >
  )
}

function DateSelector({ date }: { date: Date }) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const pathname = usePathname()
  const queryYear = date.getFullYear()
  const queryMonth = date.getMonth() + 1

  const currentYear = new Date().getFullYear()

  const years = Array.from({ length: currentYear - INITIAL_YEAR + 1 }, (_, i) => INITIAL_YEAR + i)
  const months = Array.from({ length: 12 }, (_, i) => ({
    value: String(i + 1),
    label: new Date(2025, i).toLocaleString("pt-BR", { month: "long" }),
  }))

  const createQueryString = useCallback(
    (name: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString())
      params.set(name, value)
      return params.toString()
    },
    [searchParams],
  )

  const capitalize = (str: string) => str.charAt(0).toUpperCase() + str.slice(1)

  return (
    <div className="flex items-center gap-2 w-min">
      <Select
        defaultValue={String(queryMonth)}
        onValueChange={(value) => {
          router.push(pathname + "?" + createQueryString("month", value))
        }}
      >
        <SelectTrigger className="bg-white">
          <SelectValue placeholder="Mês" />
          <span className="pl-2"></span>
        </SelectTrigger>
        <SelectContent>
          {months.map((month) => (
            <SelectItem key={month.value} value={month.value}>
              {capitalize(month.label)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {years.length > 0 && (
        <Select
          defaultValue={String(queryYear)}
          onValueChange={(value) => {
            router.push(pathname + "?" + createQueryString("year", value))
          }}
        >
          <SelectTrigger className="bg-white">
            <SelectValue placeholder="Ano" />
            <span className="pl-2"></span>
          </SelectTrigger>
          <SelectContent>
            {years.map((year) => (
              <SelectItem key={year} value={String(year)}>
                {year}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}
    </div>
  )
}

function Salary() {
  const { toast } = useToast()
  const t = useTranslations("DashboardPage")
  const queryClient = useQueryClient()
  const [isEditing, setIsEditing] = useState(false)

  const { data: salary, isPending: isPendingSalary } = useQuery({
    queryKey: ["salary"],
    queryFn: getSalary,
    refetchOnWindowFocus: false,
  })

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SalarySchema>({ resolver: zodResolver(salarySchema), defaultValues: { amount: salary?.amount } })

  const updateSalaryMut = useMutation({
    mutationFn: (data: SalarySchema) => updateSalary(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["summary"] })
      queryClient.invalidateQueries({ queryKey: ["salary"] })
    },
    onError: (e: Error) => {
      toast({ title: "Error", description: e.message, variant: "destructive" })
    },
  })

  const handleUpdateSalary = (data: SalarySchema) => {
    setIsEditing(false)
    updateSalaryMut.mutate(data)
  }

  return (
    <form onSubmit={handleSubmit(handleUpdateSalary)}>
      <div className="flex items-center min-w-56">
        <Label className="pr-2">{t("header.salary")}:</Label>
        <div className="bg-white">
          <Input
            {...register("amount", { setValueAs: (val) => Number(val) })}
            type="number"
            step="0.01"
            defaultValue={salary?.amount}
            className={`max-w-32 ${errors.amount && "border-red-500"} ${!isEditing && "hidden"}`}
          />

          {isPendingSalary ? (
            <LoaderCircle className="animate-spin size-6" />
          ) : (
            <Input
              defaultValue={salary?.amount.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })}
              disabled
              className={`max-w-32 disabled:opacity-100 ${isEditing && "hidden"}`}
            />
          )}
        </div>

        {!isPendingSalary && <SalaryEditIcons isEditing={isEditing} setIsEditing={setIsEditing} />}
      </div>
    </form>
  )
}

function SalaryEditIcons({ isEditing, setIsEditing }: { isEditing: boolean; setIsEditing: (value: boolean) => void }) {
  return (
    <>
      {isEditing ? (
        <div className="ml-2 cursor-pointer bg-green-500 rounded-full p-1 w-min hover:bg-green-600 transition-colors">
          <button type="submit" className="block">
            <CheckIcon className="size-4 text-white" />
          </button>
        </div>
      ) : (
        <div
          className="ml-2 cursor-pointer bg-yellow-400 rounded-full p-1 w-min hover:bg-yellow-500 transition-colors"
          onClick={() => setIsEditing(true)}>
          <PencilIcon className="size-4 text-white" />
        </div>
      )}
    </>
  )
}
