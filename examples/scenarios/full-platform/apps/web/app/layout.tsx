import type { Metadata } from "next";
import type { ReactNode } from "react";
import { ClerkProvider } from "@clerk/nextjs";
import { AnalyticsProviders } from "./providers";
import "./globals.css";

export const metadata: Metadata = {
  title: "Perch Brief",
  description: "Team knowledge snippets and AI answers — Perch full-platform example",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-zinc-950 text-zinc-100 antialiased">
        <ClerkProvider>
          <AnalyticsProviders>{children}</AnalyticsProviders>
        </ClerkProvider>
      </body>
    </html>
  );
}
