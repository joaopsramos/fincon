import { redirect, RedirectType } from "next/navigation"

export default function Index() {
  redirect("/login", RedirectType.replace)
}
