"use client"

import { useTranslations } from "next-intl";
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { useMutation } from "@tanstack/react-query"
import { signUp } from "@/api/user"
import { useRouter } from "next/navigation"
import { useToast } from "@/hooks/use-toast"
import { LoaderCircle } from "lucide-react"
import Link from "next/link";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

const signUpSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8, "stringMin"),
  salary: z.number().min(1, "numberMin"),
})

type SignUpSchema = z.infer<typeof signUpSchema>

export default function SignUp() {
  const commonT = useTranslations("Common");
  const errorsT = useTranslations("Errors");
  const t = useTranslations("SignUpPage");
  const router = useRouter()
  const { toast } = useToast()
  const [isNavigating, setIsNavigating] = useState(false)
  const { register, handleSubmit, formState: { errors }, } = useForm<SignUpSchema>({ resolver: zodResolver(signUpSchema) })

  const signUpMut = useMutation({
    mutationFn: (data: SignUpSchema) => signUp(data),
    onSuccess: () => {
      setIsNavigating(true)
      router.replace("/dashboard")
    },
    onError: (e: Error) => {
      toast({ title: "Error", description: e.message, variant: "destructive" })
    }
  })

  const isLoading = signUpMut.isPending || isNavigating;

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl font-bold text-center">{t("formTitle")}</CardTitle>
        </CardHeader>

        <CardContent className="pb-2">
          <form onSubmit={handleSubmit((data) => signUpMut.mutate(data))} className="space-y-4">
            <div>
              <label htmlFor="email" className="text-sm font-medium">
                Email
              </label>
              <Input {...register("email")} id="email" name="email" type="email" required className={"${errors.value ? 'border-red-500' : ''}"} />
              {errors.email && (
                <span
                  className="row-start-2 col-span-5 text-red-500 text-xs"
                >
                  {errorsT(errors.email.message)}
                </span>
              )}
            </div>

            <div>
              <label htmlFor="password" className="text-sm font-medium">
                {commonT("passwordInput")}
              </label>
              <Input {...register("password")} id="password" name="password" type="password" required />
              {errors.password && (
                <span
                  className="row-start-2 col-span-5 text-red-500 text-xs"
                >
                  {errorsT(errors.password.message, { min: 8 })}
                </span>
              )}
            </div>

            <div>
              <label htmlFor="salary" className="text-sm font-medium">
                {t("salaryInput")}
              </label>
              <Input {...register("salary", { setValueAs: val => Number(val) })} id="salary" name="salary" type="number" step={1} required className="max-w-32" />
              {errors.salary && (
                <span
                  className="row-start-2 col-span-5 text-red-500 text-xs"
                >
                  {errorsT(errors.salary.message, { min: 1 })}
                </span>
              )}
            </div>

            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? (
                <LoaderCircle className="animate-spin size-6" />
              ) : commonT("continueButton")}
            </Button>
          </form>
        </CardContent>

        <CardFooter>
          <Button variant="outline" className="w-full" asChild>
            <Link href="/login">
              {t("alreadyHaveAnAccount")}
            </Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}
