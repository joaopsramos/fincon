import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import Cookies from "js-cookie"
import { AppRouterInstance } from "next/dist/shared/lib/app-router-context.shared-runtime"

const GOALS_ORDER = ["Fixed costs", "Comfort", "Pleasures", "Knowledge", "Financial investments", "Goals"]
const goalsOrderMap = new Map(GOALS_ORDER.map((name, index) => [name, index]))

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export const TOKEN_KEY = "fincon-token"

export function getAuthCookie() {
  return Cookies.get(TOKEN_KEY)
}

export function setAuthCookie(token: string) {
  return Cookies.set(TOKEN_KEY, token, { expires: 7, sameSite: "strict", secure: process.env.NODE_ENV == "production" })
}

export function deleteAuthCookie() {
  return Cookies.remove(TOKEN_KEY)
}

export function sortGoals<T extends { name: string }>(goals: T[]) {
  return [...goals].sort((a, b) => (goalsOrderMap.get(a.name) ?? 1) - (goalsOrderMap.get(b.name) ?? 1))
}

export type Money = {
  amount: number
  currency: string
}

export function moneyToString(money: Money) {
  return money.amount.toLocaleString("pt-BR", { style: "currency", currency: money.currency })
}

export function handleLogout(router: AppRouterInstance) {
  deleteAuthCookie()
  router.replace("/")
}
