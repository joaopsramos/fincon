'use client'
import { useQuery } from "@tanstack/react-query";
import Goals from "./goals";
import Summary from "./summary";
import { getGoals } from "@/api/goals";
import Expense from "./expense";

const date = new Date()

export default function Index() {
  const { data: goals } = useQuery({
    queryKey: ["goals"],
    queryFn: getGoals,
    refetchOnWindowFocus: false
  })

  return (
    <div className="m-4">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-2">
        <div className="lg:col-span-2">
          <Summary date={date} />
        </div>
        <div className="">
          <Goals goals={goals || []} />
        </div>

      </div>

      <div className="mt-2 grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-2">
        {goals?.map(goal => (
          <Expense key={goal.id} goal={goal} />
        ))}
      </div>
    </div>
  );
}
