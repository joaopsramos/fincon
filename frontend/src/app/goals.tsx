import { Goal } from "@/api/goals"

export default function Goals({ goals }: { goals: Goal[] }) {
  return (
    <div className="bg-slate-200 rounded-md p-4 max-w-md h-full">
      <h1 className="text-xl font-bold">Metas</h1>

      <div className="mt-4">
        <ul>
          {goals?.map(goal => (
            <li key={goal.id} className="flex justify-between my-2">
              <span>{goal.name}</span>
              <span>{goal.percentage}%</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
