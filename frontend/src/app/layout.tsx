import { Geist, Geist_Mono } from "next/font/google"
import "./globals.css"
import { Toaster } from "@/components/ui/toaster"
import { getLocale, getMessages } from "next-intl/server"
import Providers from "./providers"

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
})

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
})

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const locale = await getLocale()
  const messages = await getMessages()

  return (
    <html lang={locale}>
      <head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      </head>

      <body className={`${geistSans.variable} ${geistMono.variable} antialiased bg-slate-100 dark:bg-slate-900`}>
        <Providers messages={messages} locale={locale}>
          {children}
          <Toaster />
        </Providers>
      </body>
    </html>
  )
}
