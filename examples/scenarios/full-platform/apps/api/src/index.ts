import { serve } from "@hono/node-server";
import { Hono } from "hono";
import { cors } from "hono/cors";
import { serve as serveInngest } from "inngest/hono";
import OpenAI from "openai";
import { pineconeIndex } from "./pinecone.js";
import Anthropic from "@anthropic-ai/sdk";
import Stripe from "stripe";
import { Redis } from "@upstash/redis";
import { Ratelimit } from "@upstash/ratelimit";
import { Receiver } from "@upstash/qstash";
import * as Sentry from "@sentry/node";
import { Logtail } from "@logtail/node";
import Pusher from "pusher";
import { pool, initDb } from "./db.js";
import { requireUserId } from "./auth.js";
import { inngest } from "./inngest/client.js";
import { embedSnippet, weeklyDigest } from "./inngest/functions.js";

if (process.env.SENTRY_DSN) {
  Sentry.init({ dsn: process.env.SENTRY_DSN, tracesSampleRate: 0.1 });
}

const logtail =
  process.env.LOGTAIL_SOURCE_TOKEN &&
  new Logtail(process.env.LOGTAIL_SOURCE_TOKEN, {
    endpoint: process.env.LOGTAIL_ENDPOINT ?? "https://in.logs.betterstack.com",
  });

function logInfo(msg: string, ctx?: Record<string, unknown>) {
  if (logtail) {
    void logtail.info(msg, ctx);
  } else {
    console.log(msg, ctx ?? "");
  }
}

const redis =
  process.env.UPSTASH_REDIS_REST_URL && process.env.UPSTASH_REDIS_REST_TOKEN
    ? new Redis({
        url: process.env.UPSTASH_REDIS_REST_URL,
        token: process.env.UPSTASH_REDIS_REST_TOKEN,
      })
    : null;

const ratelimit = redis
  ? new Ratelimit({
      redis,
      limiter: Ratelimit.slidingWindow(20, "1 m"),
      prefix: "perch-brief",
    })
  : null;

const app = new Hono();

app.use(
  "*",
  cors({
    origin: (origin) =>
      origin &&
      (origin.includes("localhost") ||
        origin.includes("vercel.app") ||
        (process.env.WEB_ORIGIN && origin.startsWith(process.env.WEB_ORIGIN)))
        ? origin
        : "http://localhost:3000",
    allowMethods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
    allowHeaders: ["Content-Type", "Authorization"],
    credentials: true,
  })
);

app.get("/health", (c) => c.json({ ok: true }));

app.use(
  "/api/inngest",
  serveInngest({
    client: inngest,
    functions: [embedSnippet, weeklyDigest],
  })
);

app.post("/webhooks/stripe", async (c) => {
  const stripeKey = process.env.STRIPE_SECRET_KEY;
  const whSecret = process.env.STRIPE_WEBHOOK_SECRET;
  if (!stripeKey || !whSecret) {
    return c.json({ error: "stripe not configured" }, 501);
  }
  const stripe = new Stripe(stripeKey);
  const sig = c.req.header("stripe-signature");
  const body = await c.req.text();
  let event: Stripe.Event;
  try {
    event = stripe.webhooks.constructEvent(body, sig!, whSecret);
  } catch (e) {
    logInfo("stripe webhook verify failed", { error: String(e) });
    return c.json({ error: "invalid signature" }, 400);
  }
  if (event.type === "checkout.session.completed") {
    const session = event.data.object as Stripe.Checkout.Session;
    const clerkId = session.client_reference_id;
    const customerId =
      typeof session.customer === "string" ? session.customer : session.customer?.id;
    if (clerkId && customerId) {
      await pool.query(
        `UPDATE users SET plan = 'pro', stripe_customer_id = $1 WHERE clerk_id = $2`,
        [customerId, clerkId]
      );
    }
  }
  return c.json({ received: true });
});

app.post("/webhooks/qstash", async (c) => {
  const currentKey = process.env.QSTASH_CURRENT_SIGNING_KEY;
  const nextKey = process.env.QSTASH_NEXT_SIGNING_KEY;
  const sig =
    c.req.header("Upstash-Signature") ?? c.req.header("upstash-signature") ?? "";
  if (!currentKey || !sig) {
    return c.json({ error: "qstash not configured" }, 501);
  }
  const receiver = new Receiver({
    currentSigningKey: currentKey,
    nextSigningKey: nextKey ?? currentKey,
  });
  const body = await c.req.text();
  const url = new URL(c.req.url).href;
  try {
    await receiver.verify({ signature: sig, body, url });
  } catch (e) {
    logInfo("qstash verify failed", { error: String(e) });
    return c.json({ error: "invalid signature" }, 401);
  }
  logInfo("qstash webhook ok");
  return c.json({ ok: true });
});

async function ensureUser(clerkId: string, email?: string | null, avatarUrl?: string | null) {
  const existing = await pool.query(`SELECT id FROM users WHERE clerk_id = $1`, [clerkId]);
  if (existing.rows.length) {
    if (email || avatarUrl) {
      await pool.query(
        `UPDATE users SET email = COALESCE($2, email), avatar_url = COALESCE($3, avatar_url) WHERE clerk_id = $1`,
        [clerkId, email, avatarUrl]
      );
    }
    return existing.rows[0].id as string;
  }
  const ins = await pool.query(
    `INSERT INTO users (clerk_id, email, avatar_url) VALUES ($1, $2, $3) RETURNING id`,
    [clerkId, email ?? null, avatarUrl ?? null]
  );
  return ins.rows[0].id as string;
}

app.get("/v1/me", async (c) => {
  try {
    const clerkId = await requireUserId(c.req.header("Authorization"));
    await ensureUser(clerkId);
    const r = await pool.query(
      `SELECT email, avatar_url, plan FROM users WHERE clerk_id = $1`,
      [clerkId]
    );
    return c.json({ user: r.rows[0] ?? null });
  } catch {
    return c.json({ error: "unauthorized" }, 401);
  }
});

app.patch("/v1/me", async (c) => {
  try {
    const clerkId = await requireUserId(c.req.header("Authorization"));
    const body = await c.req.json<{ avatarUrl?: string }>();
    await pool.query(`UPDATE users SET avatar_url = $2 WHERE clerk_id = $1`, [
      clerkId,
      body.avatarUrl ?? null,
    ]);
    return c.json({ ok: true });
  } catch {
    return c.json({ error: "unauthorized" }, 401);
  }
});

app.get("/v1/snippets", async (c) => {
  try {
    const clerkId = await requireUserId(c.req.header("Authorization"));
    await ensureUser(clerkId);
    const r = await pool.query(
      `SELECT s.id, s.title, s.body, s.embedding_status, s.created_at
       FROM snippets s JOIN users u ON s.user_id = u.id
       WHERE u.clerk_id = $1 ORDER BY s.created_at DESC`,
      [clerkId]
    );
    return c.json({ snippets: r.rows });
  } catch {
    return c.json({ error: "unauthorized" }, 401);
  }
});

app.post("/v1/snippets", async (c) => {
  try {
    const clerkId = await requireUserId(c.req.header("Authorization"));
    const uid = await ensureUser(clerkId);
    const body = await c.req.json<{ title: string; body: string }>();
    if (!body.title?.trim() || !body.body?.trim()) {
      return c.json({ error: "title and body required" }, 400);
    }
    const planRow = await pool.query<{ plan: string }>(
      `SELECT plan FROM users WHERE clerk_id = $1`,
      [clerkId]
    );
    const plan = planRow.rows[0]?.plan ?? "free";
    const countRow = await pool.query<{ n: string }>(
      `SELECT count(*)::text as n FROM snippets s JOIN users u ON s.user_id = u.id WHERE u.clerk_id = $1`,
      [clerkId]
    );
    const n = Number(countRow.rows[0]?.n ?? 0);
    if (plan === "free" && n >= 10) {
      return c.json({ error: "snippet limit reached; upgrade to Pro" }, 402);
    }
    const ins = await pool.query<{ id: string }>(
      `INSERT INTO snippets (user_id, title, body) VALUES ($1, $2, $3) RETURNING id`,
      [uid, body.title.trim(), body.body.trim()]
    );
    const snippetId = ins.rows[0].id;
    await inngest.send({
      name: "snippet/created",
      data: { snippetId, userId: clerkId },
    });
    logInfo("snippet created", { snippetId });
    return c.json({ snippet: { id: snippetId } });
  } catch (e) {
    if (String(e).includes("unauthorized")) {
      return c.json({ error: "unauthorized" }, 401);
    }
    throw e;
  }
});

app.post("/v1/ask", async (c) => {
  try {
    const clerkId = await requireUserId(c.req.header("Authorization"));
    await ensureUser(clerkId);
    if (ratelimit) {
      const { success } = await ratelimit.limit(clerkId);
      if (!success) {
        return c.json({ error: "rate limited" }, 429);
      }
    }
    const body = await c.req.json<{ question: string }>();
    if (!body.question?.trim()) {
      return c.json({ error: "question required" }, 400);
    }
    const oai = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });
    const qEmb = await oai.embeddings.create({
      model: "text-embedding-3-small",
      input: body.question,
    });
    const vector = qEmb.data[0].embedding;
    let index;
    try {
      index = pineconeIndex();
    } catch {
      return c.json({ error: "Pinecone not configured" }, 501);
    }
    const q = await index.query({
      vector,
      topK: 8,
      includeMetadata: true,
      filter: { userId: { $eq: clerkId } },
    });
    const ctx = (q.matches ?? [])
      .map((m) => {
        const meta = m.metadata as { title?: string; body?: string } | undefined;
        const title = meta?.title ?? "snippet";
        const body = meta?.body?.trim();
        if (body) {
          return `- ${title}:\n${body}`;
        }
        return `- ${title}: ${JSON.stringify(m.metadata)}`;
      })
      .join("\n\n");
    const mode = process.env.ANSWER_MODEL === "openai" ? "openai" : "anthropic";
    let answer: string;
    if (mode === "openai") {
      const chat = await oai.chat.completions.create({
        model: process.env.OPENAI_CHAT_MODEL ?? "gpt-4o-mini",
        messages: [
          {
            role: "system",
            content:
              "You are Perch Brief. Answer using the context snippets when relevant; be concise.",
          },
          {
            role: "user",
            content: `Context:\n${ctx || "(no snippets yet)"}\n\nQuestion: ${body.question}`,
          },
        ],
      });
      answer = chat.choices[0]?.message?.content ?? "";
    } else {
      const anthropic = new Anthropic({ apiKey: process.env.ANTHROPIC_API_KEY });
      const msg = await anthropic.messages.create({
        model: process.env.ANTHROPIC_MODEL ?? "claude-3-5-sonnet-20241022",
        max_tokens: 1024,
        messages: [
          {
            role: "user",
            content: `Context:\n${ctx || "(no snippets yet)"}\n\nQuestion: ${body.question}`,
          },
        ],
      });
      const block = msg.content[0];
      answer = block.type === "text" ? block.text : "";
    }
    logInfo("ask answered", { clerkId });
    return c.json({ answer });
  } catch (e) {
    if (String(e).includes("unauthorized")) {
      return c.json({ error: "unauthorized" }, 401);
    }
    throw e;
  }
});

app.post("/v1/billing/checkout", async (c) => {
  try {
    const clerkId = await requireUserId(c.req.header("Authorization"));
    const stripeKey = process.env.STRIPE_SECRET_KEY;
    const priceId = process.env.STRIPE_PRICE_ID;
    if (!stripeKey || !priceId) {
      return c.json({ error: "stripe not configured" }, 501);
    }
    const stripe = new Stripe(stripeKey);
    const base = process.env.WEB_ORIGIN ?? "http://localhost:3000";
    const session = await stripe.checkout.sessions.create({
      mode: "subscription",
      client_reference_id: clerkId,
      line_items: [{ price: priceId, quantity: 1 }],
      success_url: `${base}/?checkout=success`,
      cancel_url: `${base}/?checkout=cancel`,
    });
    return c.json({ url: session.url });
  } catch (e) {
    if (String(e).includes("unauthorized")) {
      return c.json({ error: "unauthorized" }, 401);
    }
    throw e;
  }
});

app.post("/v1/pusher/auth", async (c) => {
  try {
    const clerkId = await requireUserId(c.req.header("Authorization"));
    const ct = c.req.header("content-type") ?? "";
    let socketId: string;
    let channel: string;
    if (ct.includes("application/x-www-form-urlencoded")) {
      const text = await c.req.text();
      const params = new URLSearchParams(text);
      socketId = params.get("socket_id") ?? "";
      channel = params.get("channel_name") ?? "";
    } else {
      const body = await c.req.parseBody();
      socketId = body.socket_id as string;
      channel = body.channel_name as string;
    }
    if (!channel?.startsWith("private-user-")) {
      return c.json({ error: "invalid channel" }, 400);
    }
    if (channel !== `private-user-${clerkId}`) {
      return c.json({ error: "forbidden" }, 403);
    }
    const push = new Pusher({
      appId: process.env.PUSHER_APP_ID!,
      key: process.env.PUSHER_KEY!,
      secret: process.env.PUSHER_SECRET!,
      cluster: process.env.PUSHER_CLUSTER!,
      useTLS: true,
    });
    const auth = push.authorizeChannel(socketId, channel);
    return c.json(auth);
  } catch {
    return c.json({ error: "unauthorized" }, 401);
  }
});

const port = Number(process.env.PORT ?? process.env.API_PORT ?? 4000);

initDb()
  .then(() => {
    serve({ fetch: app.fetch, port });
    logInfo(`api listening on ${port}`);
  })
  .catch((e) => {
    console.error(e);
    process.exit(1);
  });
