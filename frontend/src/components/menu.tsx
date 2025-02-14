import { ArrowRightStartOnRectangleIcon } from "@heroicons/react/24/solid"
import { Button } from "@/components/ui/button"
import { handleLogout } from "@/lib/utils"
import { useRouter } from "next/navigation"

export default function Menu() {
  const router = useRouter()

  return (
    <div className="fixed z-50 bottom-0 bg-white border-t w-full p-2 flex items-center justify-center min-[375px]:hidden">
      <Button size={"sm"} onClick={() => handleLogout(router)} variant="ghost">
        <ArrowRightStartOnRectangleIcon className="size-6 text-slate-800" />
      </Button>
    </div>
  )
}
