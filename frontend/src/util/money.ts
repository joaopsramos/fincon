export type Money = {
  amount: number,
  currency: string
}

export function moneyToString(money: Money) {
  return money.amount.toLocaleString("pt-BR", { style: "currency", currency: money.currency })
}

