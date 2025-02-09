import { Goal } from "@/api/goals"
import { useTranslations } from "next-intl"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"

export default function Goals({ goals }: { goals: Goal[] }) {
  const t = useTranslations("DashboardPage")

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>{t("goals.title")}</CardTitle>
      </CardHeader>

      <CardContent>
        <ul>
          {goals?.map(goal => (
            <li key={goal.id} className="my-2">
              <div className="flex justify-between">
                <span>{goal.name}</span>
                <span>{goal.percentage}%</span>
              </div>
              <Separator className="my-2" />
            </li>
          ))}
        </ul>
      </CardContent>
    </Card >
  )
}
