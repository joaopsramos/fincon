import type React from "react"
import { forwardRef, useEffect, useImperativeHandle, useState } from "react"
import { useTranslations } from "next-intl"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { PlusIcon } from "@heroicons/react/24/solid"
import ExpenseForm, { ExpenseFormProps } from "./expense_form"

export type UpsertExpenseDialogProps = {
  onDialogClose?: () => void
}

export type UpsertExpenseDialogRef = {
  openDialog: () => void
}

const UpsertExpenseDialog = forwardRef<UpsertExpenseDialogRef, UpsertExpenseDialogProps & ExpenseFormProps>(
  (props, ref) => {
    const { expense, onDialogClose } = props
    const commonT = useTranslations("Common")
    const t = useTranslations("DashboardPage.expenses")
    const [open, setOpen] = useState(false)

    useEffect(() => {
      if (!open) {
        onDialogClose?.()
      }
    }, [open, onDialogClose])

    useImperativeHandle(ref, () => ({
      openDialog() {
        setOpen(true)
      },
    }))

    return (
      <>
        <div className="flex justify-center">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  size="icon"
                  className="bg-slate-900 dark:bg-white rounded-full hover:bg-slate-800 dark:hover:bg-slate-200 [&_svg]:size-6"
                  onClick={() => setOpen(true)}>
                  <PlusIcon className="text-white dark:text-slate-900" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{expense ? t("editTooltip") : t("addExpense")}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        <Dialog open={open} onOpenChange={setOpen}>
          <DialogContent className="top-60">
            <DialogHeader>
              <DialogTitle>{expense ? t("editExpense") : t("addExpense")}</DialogTitle>
            </DialogHeader>

            {/* used to not focus first form input */}
            <input className="fixed left-0 top-0 h-0 w-0" type="checkbox" autoFocus={true} />

            <ExpenseForm {...props} />

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" className="mt-2 sm:mt-0" onClick={() => setOpen(false)}>
                {commonT("cancel")}
              </Button>
              <Button type="submit" form="upsert-form">
                {expense ? commonT("update") : commonT("add")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </>
    )
  }
)

UpsertExpenseDialog.displayName = "UpsertExpenseDialog"

export default UpsertExpenseDialog
