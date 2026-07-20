"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { useUser } from "@clerk/nextjs";
import posthog from "posthog-js";
import * as Sentry from "@sentry/browser";

export function AnalyticsProviders({ children }: { children: ReactNode }) {
  const { user, isSignedIn } = useUser();

  useEffect(() => {
    const key = process.env.NEXT_PUBLIC_POSTHOG_KEY;
    const host = process.env.NEXT_PUBLIC_POSTHOG_HOST ?? "https://us.i.posthog.com";
    if (key) {
      posthog.init(key, { api_host: host, person_profiles: "identified_only" });
    }
    const dsn = process.env.NEXT_PUBLIC_SENTRY_DSN;
    if (dsn) {
      Sentry.init({ dsn, tracesSampleRate: 0.1 });
    }
  }, []);

  useEffect(() => {
    if (isSignedIn && user) {
      posthog.identify(user.id, { email: user.primaryEmailAddress?.emailAddress });
    }
  }, [isSignedIn, user]);

  return <>{children}</>;
}
