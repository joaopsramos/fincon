"use client"

import { TooltipProvider } from "@/components/ui/tooltip"
import { isServer, QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { AbstractIntlMessages, NextIntlClientProvider } from "next-intl"

let browserQueryClient: QueryClient | undefined = undefined

type ProvidersProps = {
  children: React.ReactNode
  messages: AbstractIntlMessages
  locale: string
}

export default function Providers({ children, messages, locale }: ProvidersProps) {
  // NOTE: Avoid useState when initializing the query client if you don't
  //       have a suspense boundary between this and the code that may
  //       suspend because React will throw away the client on the initial
  //       render if it suspends and there is no boundary
  const queryClient = getQueryClient()

  return (
    <QueryClientProvider client={queryClient}>
      <NextIntlClientProvider messages={messages} locale={locale} timeZone="America/Sao_Paulo">
        <TooltipProvider>{children}</TooltipProvider>
      </NextIntlClientProvider>
    </QueryClientProvider>
  )
}

function getQueryClient() {
  if (isServer) {
    return makeQueryClient()
  } else {
    // Browser: make a new query client if we don't already have one
    // This is very important, so we don't re-make a new client if React
    // suspends during the initial render. This may not be needed if we
    // have a suspense boundary BELOW the creation of the query client
    if (!browserQueryClient) browserQueryClient = makeQueryClient()
    return browserQueryClient
  }
}

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // With SSR, we usually want to set some default staleTime
        // above 0 to avoid refetching immediately on the client
        staleTime: 60 * 1000,
      },
    },
  })
}
