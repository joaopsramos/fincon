import { Goal } from "@/api/goals"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"

export default function Goals({ goals }: { goals: Goal[] }) {
  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>Goals</CardTitle>
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
