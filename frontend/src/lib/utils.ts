import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import Cookies from "js-cookie"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

const TOKEN_KEY = "@fincon:token"

export function getAuthCookie() {
  return Cookies.get(TOKEN_KEY)
}

export function setAuthCookie(token: string) {
  return Cookies.set(TOKEN_KEY, token)
}

export type Money = {
  amount: number,
  currency: string
}

export function moneyToString(money: Money) {
  return money.amount.toLocaleString("pt-BR", { style: "currency", currency: money.currency })
}

