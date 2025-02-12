"use client"

import { useTranslations } from "next-intl";
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import Form from "next/form"
import { useMutation } from "@tanstack/react-query"
import { login } from "@/api/user"
import { useRouter } from "next/navigation"
import { useToast } from "@/hooks/use-toast"
import { LoaderCircle } from "lucide-react"
import Link from "next/link";

export default function Login() {
  const t = useTranslations("LoginPage");
  const commonT = useTranslations("Common");
  const router = useRouter()
  const { toast } = useToast()
  const [isNavigating, setIsNavigating] = useState(false)

  const loginMut = useMutation({
    mutationFn: (formData: FormData) => login(formData),
    onSuccess: () => {
      setIsNavigating(true)
      router.replace("/dashboard")
    },
    onError: (e: Error) => {
      toast({ title: "Error", description: e.message, variant: "destructive" })
    }
  })

  const isLoading = loginMut.isPending || isNavigating;

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl font-bold text-center">{t("formTitle")}</CardTitle>
        </CardHeader>

        <CardContent className="pb-2">
          <Form action={loginMut.mutate} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium">
                Email
              </label>
              <Input id="email" name="email" type="email" required />
            </div>
            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium">
                {commonT("passwordInput")}
              </label>
              <Input id="password" name="password" type="password" required />
            </div>
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? (
                <LoaderCircle className="animate-spin size-6" />
              ) : commonT("continueButton")}
            </Button>
          </Form>
        </CardContent>

        <CardFooter>
          <Button variant="outline" className="w-full" asChild>
            <Link href="/signup">
              {t("needAnAccount")}
            </Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}

