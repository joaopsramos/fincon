import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { ArrowRightStartOnRectangleIcon } from "@heroicons/react/24/outline"
import { CheckIcon, PencilIcon } from "@heroicons/react/24/solid"
import { useTranslations } from "next-intl"
import { useState } from "react"
import { handleLogout } from "@/lib/utils"
import { useRouter } from "next/navigation"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { getSalary, updateSalary } from "@/api/salary"
import { z } from "zod"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useToast } from "@/hooks/use-toast"
import { LoaderCircle } from "lucide-react"

const salarySchema = z.object({
  amount: z.number().min(1, "valuteTooLow"),
})

type SalarySchema = z.infer<typeof salarySchema>

export default function Header({ date: _ }: { date: Date }) {
  const router = useRouter()
  const t = useTranslations("DashboardPage")
  const { toast } = useToast()
  const queryClient = useQueryClient()
  const menuT = useTranslations("Menu")
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
    <div className="flex items-center justify-between">
      <form onSubmit={handleSubmit(handleUpdateSalary)}>
        <div className="flex items-center">
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

      <div className="hidden min-[425px]:block">
        <Button size={"sm"} onClick={() => handleLogout(router)}>
          <ArrowRightStartOnRectangleIcon className="size-5 text-white" />
          {menuT("logout")}
        </Button>
      </div>
    </div>
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
