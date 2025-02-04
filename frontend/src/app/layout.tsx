'use client'
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "@/components/ui/toaster";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const queryClient = new QueryClient()

export default function RootLayout({ children, }: Readonly<{ children: React.ReactNode; }>) {
  return (
    <QueryClientProvider client={queryClient}>
      <html lang="en">
        <head>
          <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        </head>

        <body
          className={`${geistSans.variable} ${geistMono.variable} antialiased`}
        >
          {children}
          <Toaster />
        </body>
      </html>
    </QueryClientProvider>
  );
}
