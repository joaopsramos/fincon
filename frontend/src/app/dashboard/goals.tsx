import { Goal } from "@/api/goals"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export default function Goals({ goals }: { goals: Goal[] }) {
  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>Goals</CardTitle>
      </CardHeader>

      <CardContent>
        <ul>
          {goals?.map(goal => (
            <li key={goal.id} className="flex justify-between my-2">
              <span>{goal.name}</span>
              <span>{goal.percentage}%</span>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card >
  )
}
