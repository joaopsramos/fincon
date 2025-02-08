"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import Form from "next/form"
import { useMutation } from "@tanstack/react-query"
import { login, signUp } from "@/api/session"
import { useRouter } from "next/navigation"
import { useToast } from "@/hooks/use-toast"
import { LoaderCircle } from "lucide-react"

export default function Login() {
  const router = useRouter()
  const [isSignUp, setIsSignUp] = useState(false)
  const { toast } = useToast()

  const loginMut = useMutation({
    mutationFn: (formData: FormData) => login(formData),
    onSuccess: () => {
      router.replace("/dashboard")
    },
    onError: (e: Error) => {
      toast({ title: "Error", description: e.message, variant: "destructive" })
    }
  })

  const signUpMut = useMutation({
    mutationFn: (formData: FormData) => signUp(formData),
    onSuccess: () => {
      router.replace("/dashboard")
    },
    onError: (e: Error) => {
      toast({ title: "Error", description: e.message, variant: "destructive" })
    }
  })

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl font-bold text-center">{isSignUp ? "Sign Up" : "Sign In"}</CardTitle>
        </CardHeader>

        <CardContent className="pb-2">
          <Form action={isSignUp ? signUpMut.mutate : loginMut.mutate} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium">
                Email
              </label>
              <Input id="email" name="email" type="email" required />
            </div>
            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium">
                Password
              </label>
              <Input id="password" name="password" type="password" required />
            </div>
            <Button type="submit" className="w-full" disabled={loginMut.isPending || signUpMut.isPending}>
              {loginMut.isPending || signUpMut.isPending ? (
                <LoaderCircle className="animate-spin size-6" />
              ) : isSignUp ? "Sign Up" : "Sign In"}
            </Button>
          </Form>
        </CardContent>

        <CardFooter>
          <Button variant="outline" className="w-full" onClick={() => setIsSignUp(!isSignUp)}>
            {isSignUp ? "Already have an account? Sign In" : "Need an account? Sign Up"}
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}

