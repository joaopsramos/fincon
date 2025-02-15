import { SummaryGoal } from "@/api/summary"
import { moneyValueToString } from "@/lib/utils"
import { ColumnDef } from "@tanstack/react-table"

export const summaryColumns: ColumnDef<SummaryGoal>[] = [
  { accessorKey: "name", header: "Budget", cell: ({ row }) => <div className="">{row.getValue("name")}</div> },
  { accessorKey: "spent", header: "Spent", cell: ({ row }) => moneyValueToString(row.getValue("spent")) },
  { accessorKey: "must_spend", header: "Must spend", cell: ({ row }) => moneyValueToString(row.getValue("spent")) },
  {
    accessorKey: "used",
    header: "Used",
    cell: ({ row }) => {
      const used = parseFloat(row.getValue("used"))
      return used.toFixed(2).toString() + "%"
    },
  },
  {
    accessorKey: "total",
    header: "Total",
    cell: ({ row }) => {
      const used = parseFloat(row.getValue("total"))
      return used.toFixed(2).toString() + "%"
    },
  },
]
