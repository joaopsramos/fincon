'use client'
import { useQuery } from "@tanstack/react-query";
import Goals from "./goals";
import Summary from "./summary";
import { getGoals } from "@/api/goals";
import Expense from "./expense";

export default function Index() {
  const { data: goals } = useQuery({
    queryKey: ["goals"],
    queryFn: getGoals
  })

  return (
    <div className="m-4">
      <div className="flex">
        <div className="grow">
          <Summary />
        </div>
        <div className="grow">
          <Goals goals={goals || []} />
        </div>

      </div>

      <div className="mt-4 grid grid-cols-3 gap-8">
        {goals?.map(goal => (
          <Expense key={goal.id} goal={goal} />
        ))}
      </div>
    </div>
  );
}
