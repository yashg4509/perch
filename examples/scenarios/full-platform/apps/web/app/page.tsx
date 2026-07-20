"use client";

import { useAuth, UserButton } from "@clerk/nextjs";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import Pusher from "pusher-js";
import { apiFetch } from "@/lib/api";

type Snippet = {
  id: string;
  title: string;
  body: string;
  embedding_status: string;
  created_at: string;
};

export default function Home() {
  const { getToken, userId, isLoaded } = useAuth();
  const [snippets, setSnippets] = useState<Snippet[]>([]);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [question, setQuestion] = useState("");
  const [answer, setAnswer] = useState("");
  const [status, setStatus] = useState<string | null>(null);
  const [embedEvents, setEmbedEvents] = useState<string[]>([]);
  const [uiError, setUiError] = useState<string | null>(null);

  const [loadingSnippets, setLoadingSnippets] = useState(false);
  const [savingSnippet, setSavingSnippet] = useState(false);
  const [asking, setAsking] = useState(false);
  const [checkingOut, setCheckingOut] = useState(false);

  const apiBase = process.env.NEXT_PUBLIC_API_URL ?? "http://127.0.0.1:4000";

  function getApiErrorHint() {
    return `Could not reach the API at ${apiBase}. Make sure the API server is running (Terminal: \`npm run dev -w @perch-brief/api\`) and then reload.`;
  }

  function embeddingBadge(embeddingStatus: string) {
    const normalized = embeddingStatus.trim().toLowerCase();
    if (normalized === "ready") {
      return {
        label: "Ready",
        className: "border-emerald-400/20 bg-emerald-400/10 text-emerald-200",
      };
    }
    if (normalized === "pending") {
      return {
        label: "Pending",
        className: "border-amber-400/20 bg-amber-400/10 text-amber-200",
      };
    }
    if (normalized === "failed") {
      return {
        label: "Failed",
        className: "border-red-400/20 bg-red-400/10 text-red-200",
      };
    }
    return {
      label: embeddingStatus || "Unknown",
      className: "border-zinc-700 bg-zinc-800/50 text-zinc-200",
    };
  }

  const load = useCallback(async () => {
    setUiError(null);
    setStatus(null);
    setLoadingSnippets(true);
    try {
      const token = await getToken();
      if (!token) return;
      const res = await apiFetch("/v1/snippets", token);
      const data = await res.json();
      if (res.ok) setSnippets(data.snippets ?? []);
      else setUiError(data.error ?? "Failed to load snippets.");
    } catch {
      setUiError(getApiErrorHint());
    } finally {
      setLoadingSnippets(false);
    }
  }, [getToken]);

  useEffect(() => {
    if (isLoaded && userId) {
      void load();
    }
  }, [isLoaded, userId, load]);

  useEffect(() => {
    const key = process.env.NEXT_PUBLIC_PUSHER_KEY;
    if (!userId || !key) {
      return;
    }
    const cluster = process.env.NEXT_PUBLIC_PUSHER_CLUSTER ?? "us2";
    const authUrl = `${process.env.NEXT_PUBLIC_API_URL ?? "http://127.0.0.1:4000"}/v1/pusher/auth`;
    const pusher = new Pusher(key, {
      cluster,
      authorizer: (channel) => ({
        authorize: (socketId, callback) => {
          void (async () => {
            try {
              const token = await getToken();
              const body = new URLSearchParams({
                socket_id: socketId,
                channel_name: channel.name,
              });
              const res = await fetch(authUrl, {
                method: "POST",
                headers: {
                  "Content-Type": "application/x-www-form-urlencoded",
                  ...(token ? { Authorization: `Bearer ${token}` } : {}),
                },
                body,
              });
              const data = (await res.json()) as { auth?: string };
              if (!res.ok || !data.auth) {
                callback(new Error("pusher auth failed"), null);
                return;
              }
              callback(null, { auth: data.auth });
            } catch (e) {
              callback(e as Error, null);
            }
          })();
        },
      }),
    });
    const ch = pusher.subscribe(`private-user-${userId}`);
    ch.bind("snippet-ready", (payload: { snippetId: string }) => {
      setEmbedEvents((e) => [`Snippet ${payload.snippetId} embedded`, ...e].slice(0, 5));
      void load();
    });
    return () => {
      ch.unbind_all();
      pusher.disconnect();
    };
  }, [getToken, load, userId]);

  async function addSnippet(e: React.FormEvent) {
    e.preventDefault();
    setUiError(null);
    setStatus(null);
    setSavingSnippet(true);
    try {
      const token = await getToken();
      if (!token) return;
      const res = await apiFetch("/v1/snippets", token, {
        method: "POST",
        body: JSON.stringify({ title, body }),
      });
      const data = await res.json();
      if (!res.ok) {
        setStatus(data.error ?? "Error");
        return;
      }
      setTitle("");
      setBody("");
      await load();
      setStatus("Snippet saved — embedding in background.");
    } catch {
      setUiError(getApiErrorHint());
    } finally {
      setSavingSnippet(false);
    }
  }

  async function ask(e: React.FormEvent) {
    e.preventDefault();
    setUiError(null);
    setStatus(null);
    setAnswer("");
    setAsking(true);
    try {
      const token = await getToken();
      if (!token) return;
      const res = await apiFetch("/v1/ask", token, {
        method: "POST",
        body: JSON.stringify({ question }),
      });
      const data = await res.json();
      if (!res.ok) {
        setStatus(data.error ?? "Error");
        return;
      }
      setAnswer(data.answer ?? "");
    } catch {
      setUiError(getApiErrorHint());
    } finally {
      setAsking(false);
    }
  }

  async function checkout() {
    setUiError(null);
    setCheckingOut(true);
    try {
      const token = await getToken();
      if (!token) return;
      const res = await apiFetch("/v1/billing/checkout", token, { method: "POST" });
      const data = await res.json();
      if (data.url) window.location.href = data.url;
      else setStatus(data.error ?? "Billing unavailable");
    } catch {
      setUiError(getApiErrorHint());
    } finally {
      setCheckingOut(false);
    }
  }

  if (!isLoaded) {
    return (
      <div className="min-h-screen p-10 text-zinc-400">
        Loading…
      </div>
    );
  }

  return (
    <main className="min-h-screen bg-gradient-to-b from-zinc-950 via-zinc-900 to-zinc-950">
      <div className="mx-auto max-w-5xl px-4 py-10">
        <header className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-3xl font-semibold tracking-tight">Perch Brief</h1>
            <p className="mt-1 text-sm text-zinc-400">
              Snippets → embeddings → answers. Full-platform example for{" "}
              <Link
                className="text-sky-400 underline decoration-sky-400/40 underline-offset-2"
                href="https://github.com/yashg4509/perch"
                target="_blank"
                rel="noreferrer"
              >
                perch
              </Link>
              .
            </p>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden rounded-full border border-zinc-800/70 bg-zinc-900/40 px-3 py-1 text-xs text-zinc-200 backdrop-blur sm:block">
              API: {apiBase.replace(/^https?:\/\//, "")}
            </div>
            <UserButton afterSignOutUrl="/" />
          </div>
        </header>

        {uiError && (
          <div className="mb-6 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200">
            <div className="font-medium">Something went wrong</div>
            <div className="mt-1 opacity-90">{uiError}</div>
          </div>
        )}

        {status && (
          <div className="mb-6 rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-200">
            {status}
          </div>
        )}

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <section className="rounded-2xl border border-zinc-800/70 bg-zinc-900/40 p-5 backdrop-blur lg:col-span-1">
            <h2 className="text-lg font-semibold">Add snippet</h2>
            <p className="mt-1 text-sm text-zinc-400">
              Save knowledge for your team. Background job embeds it into Pinecone.
            </p>
            <form onSubmit={addSnippet} className="mt-4 flex flex-col gap-3">
              <label className="text-xs text-zinc-400">Title</label>
              <input
                className="rounded-xl border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 shadow-sm outline-none ring-0 transition focus:border-sky-500/50 focus:ring-2 focus:ring-sky-500/20"
                placeholder="e.g. Release checklist"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                required
                disabled={savingSnippet}
              />
              <label className="text-xs text-zinc-400">Content</label>
              <textarea
                className="min-h-[110px] rounded-xl border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 shadow-sm outline-none ring-0 transition focus:border-sky-500/50 focus:ring-2 focus:ring-sky-500/20"
                placeholder="What should the assistant remember?"
                value={body}
                onChange={(e) => setBody(e.target.value)}
                required
                disabled={savingSnippet}
              />
              <button
                type="submit"
                disabled={savingSnippet}
                className="rounded-xl bg-sky-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-sky-500 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {savingSnippet ? "Saving..." : "Save snippet"}
              </button>
            </form>
            {embedEvents.length > 0 && (
              <div className="mt-4 rounded-xl border border-emerald-400/20 bg-emerald-400/10 px-3 py-2">
                <div className="text-xs font-medium text-emerald-200">Embedding updates</div>
                <ul className="mt-2 space-y-1 text-xs text-emerald-200/90">
                  {embedEvents.map((x, i) => (
                    <li key={i}>{x}</li>
                  ))}
                </ul>
              </div>
            )}
          </section>

          <section className="rounded-2xl border border-zinc-800/70 bg-zinc-900/40 p-5 backdrop-blur lg:col-span-2">
            <h2 className="text-lg font-semibold">Ask your knowledge base</h2>
            <p className="mt-1 text-sm text-zinc-400">
              Retrieval from Pinecone (filtered by user) + LLM answers with retrieved snippets.
            </p>
            <form onSubmit={ask} className="mt-4 flex flex-col gap-3">
              <label className="text-xs text-zinc-400">Question</label>
              <textarea
                className="min-h-[90px] rounded-xl border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 shadow-sm outline-none ring-0 transition focus:border-violet-500/50 focus:ring-2 focus:ring-violet-500/20"
                placeholder="e.g. What’s our release checklist?"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                required
                disabled={asking}
              />
              <button
                type="submit"
                disabled={asking}
                className="rounded-xl bg-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-violet-500 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {asking ? "Asking..." : "Ask"}
              </button>
            </form>
            {answer && (
              <article className="mt-4 rounded-xl border border-zinc-800 bg-zinc-950/40 p-4 text-sm leading-relaxed text-zinc-200">
                <div className="mb-2 text-xs font-medium text-zinc-400">Answer</div>
                <div className="whitespace-pre-wrap">{answer}</div>
              </article>
            )}
          </section>

          <section className="rounded-2xl border border-zinc-800/70 bg-zinc-900/40 p-5 backdrop-blur lg:col-span-1">
            <h2 className="text-lg font-semibold">Snippets</h2>
            <p className="mt-1 text-sm text-zinc-400">
              {loadingSnippets ? "Loading..." : `${snippets.length} saved`}
            </p>
            <div className="mt-4 space-y-3">
              {snippets.length === 0 ? (
                <div className="rounded-xl border border-zinc-800 bg-zinc-950/30 p-4 text-sm text-zinc-400">
                  No snippets yet. Add one in the left panel to begin.
                </div>
              ) : (
                snippets.map((s) => {
                  const badge = embeddingBadge(s.embedding_status);
                  return (
                    <div key={s.id} className="rounded-xl border border-zinc-800 bg-zinc-950/30 p-4">
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <div className="truncate font-semibold text-zinc-100">{s.title}</div>
                          <div className="mt-2 line-clamp-3 whitespace-pre-wrap text-xs text-zinc-300/90">
                            {s.body}
                          </div>
                        </div>
                        <div className={`shrink-0 rounded-full border px-2 py-1 text-[11px] ${badge.className}`}>
                          {badge.label}
                        </div>
                      </div>
                      <div className="mt-3 text-[11px] text-zinc-500">
                        {new Date(s.created_at).toLocaleString()}
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </section>

          <section className="rounded-2xl border border-zinc-800/70 bg-zinc-900/40 p-5 backdrop-blur lg:col-span-2">
            <h2 className="text-lg font-semibold">Billing</h2>
            <p className="mt-1 text-sm text-zinc-400">
              Free tier: up to 10 snippets. Pro unlocks higher limits (Stripe test mode).
            </p>
            <button
              type="button"
              onClick={() => void checkout()}
              disabled={checkingOut}
              className="mt-4 rounded-xl border border-zinc-700 bg-zinc-950/40 px-4 py-2 text-sm font-semibold text-zinc-100 shadow-sm transition hover:bg-zinc-900 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {checkingOut ? "Opening checkout..." : "Upgrade with Stripe"}
            </button>
          </section>
        </div>
      </div>
    </main>
  );
}
