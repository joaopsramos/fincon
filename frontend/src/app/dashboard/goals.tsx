import { Goal, updateGoals } from "@/api/goals"
import { useTranslations } from "next-intl"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { PencilSquareIcon } from "@heroicons/react/24/outline"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Label } from "@/components/ui/label"
import { Slider } from "@/components/ui/slider"
import Form from "next/form"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { useToast } from "@/hooks/use-toast"

export default function Goals({ goals }: { goals: Goal[] }) {
  const commonT = useTranslations("Common")
  const t = useTranslations("DashboardPage.goals")
  const queryClient = useQueryClient()
  const [saveDisabled, setSaveDisable] = useState(false)
  const [percentages, setPercentages] = useState<number[]>([])
  const { toast } = useToast()

  const updateMut = useMutation({
    mutationFn: (params: FormData) => updateGoals(goals, params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["summary"] })
      toast({ title: t("goalsUpdatedTitle") })
    },
  })

  useEffect(() => resetPercentages(), [goals])

  useEffect(() => {
    const sum = percentages.reduce((acc, p) => acc + p, 0)

    setSaveDisable(sum !== 100)
  }, [percentages])

  const resetPercentages = () => {
    setPercentages(goals.map((goal) => goal.percentage ?? 0))
  }

  const handleSliderChange = (idx: number, value: number) => {
    const newPercentages = [...percentages]
    newPercentages[idx] = value
    setPercentages(newPercentages)
  }

  return (
    <Dialog>
      <Card className="h-full">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>{t("title")}</CardTitle>

            <DialogTrigger>
              <PencilSquareIcon className="size-5 cursor-pointer" />
            </DialogTrigger>
          </div>
        </CardHeader>

        <CardContent>
          <ul>
            {goals?.map((goal) => (
              <li key={goal.id} className="my-2">
                <div className="flex justify-between">
                  <span>{commonT(goal.name)}</span>
                  <span>{goal.percentage}%</span>
                </div>
                <Separator className="my-1" />
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>

      <DialogContent>
        <DialogHeader>
          <DialogTitle className="dark:text-white">{t("editForm")}</DialogTitle>
        </DialogHeader>

        <DialogDescription>{t("editGoalsDesc")}</DialogDescription>

        <Form action={updateMut.mutate}>
          <ul>
            {goals?.map((goal, idx) => (
              <li key={goal.id} className="my-2 relative dark:text-white">
                <Label htmlFor={"input-" + goal.id}>{goals[idx].name}</Label>
                <Slider
                  id={"input-" + goal.id}
                  name={"percentage-" + goal.id}
                  max={100}
                  min={0}
                  step={5}
                  defaultValue={[goal.percentage]}
                  onValueChange={(val) => handleSliderChange(idx, val[0])}
                  value={[percentages[idx]]}
                  className="mt-2 mb-8"
                />
                <span
                  className="absolute text-sm -bottom-7"
                  style={{
                    left: `${percentages[idx]}%`,
                    transform: "translateX(-50%)",
                  }}>
                  {percentages[idx]}%
                </span>
              </li>
            ))}
          </ul>

          <DialogFooter>
            <Button onClick={() => resetPercentages()} type="button" variant={"secondary"} className="mt-2">
              {t("reset")}
            </Button>

            <DialogClose disabled={saveDisabled} asChild>
              <Button disabled={saveDisabled} type="submit" className="mt-2">
                {t("save")}
              </Button>
            </DialogClose>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
