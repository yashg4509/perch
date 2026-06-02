import OpenAI from "openai";
import Pusher from "pusher";
import { inngest } from "./client.js";
import { pool } from "../db.js";
import { pineconeIndex } from "../pinecone.js";

function openai(): OpenAI {
  return new OpenAI({ apiKey: process.env.OPENAI_API_KEY });
}

function pusher(): Pusher {
  return new Pusher({
    appId: process.env.PUSHER_APP_ID!,
    key: process.env.PUSHER_KEY!,
    secret: process.env.PUSHER_SECRET!,
    cluster: process.env.PUSHER_CLUSTER!,
    useTLS: true,
  });
}

export const embedSnippet = inngest.createFunction(
  {
    id: "embed-snippet",
    name: "Embed snippet into Pinecone",
    onFailure: async ({ event }) => {
      const { snippetId } = event.data.event.data as { snippetId: string };
      await pool.query(
        `UPDATE snippets SET embedding_status = 'failed' WHERE id = $1`,
        [snippetId]
      );
    },
  },
  { event: "snippet/created" },
  async ({ event, step }) => {
    const { snippetId, userId } = event.data as {
      snippetId: string;
      userId: string;
    };

    const row = await step.run("load-snippet", async () => {
      const r = await pool.query<{ id: string; title: string; body: string }>(
        `SELECT s.id, s.title, s.body FROM snippets s JOIN users u ON s.user_id = u.id
         WHERE s.id = $1 AND u.clerk_id = $2`,
        [snippetId, userId]
      );
      return r.rows[0];
    });
    if (!row) {
      return { skipped: true };
    }

    const embedding = await step.run("embed", async () => {
      const o = openai();
      const res = await o.embeddings.create({
        model: "text-embedding-3-small",
        input: `${row.title}\n\n${row.body}`,
      });
      return res.data[0].embedding;
    });

    await step.run("pinecone-upsert", async () => {
      const idx = pineconeIndex();
      // Pinecone JS v6: upsert takes an array of records, not { records: [...] }.
      await idx.upsert([
        {
          id: row.id,
          values: embedding,
          metadata: {
            userId,
            snippetId: row.id,
            title: row.title,
            body: truncateMeta(row.body, 4000),
          },
        },
      ]);
    });

    await step.run("db-update", async () => {
      await pool.query(
        `UPDATE snippets SET embedding_status = 'ready', pinecone_id = $1 WHERE id = $2`,
        [row.id, row.id]
      );
    });

    await step.run("pusher", async () => {
      try {
        const push = pusher();
        await push.trigger(`private-user-${userId}`, "snippet-ready", {
          snippetId: row.id,
        });
      } catch {
        // Pusher optional in dev without keys
      }
    });

    return { ok: true, snippetId: row.id };
  }
);

export const weeklyDigest = inngest.createFunction(
  { id: "weekly-digest", name: "Weekly digest email" },
  { cron: "TZ=America/New_York 0 9 * * 1" },
  async ({ step }) => {
    await step.run("send-digests", async () => {
      const { Resend } = await import("resend");
      const apiKey = process.env.RESEND_API_KEY;
      if (!apiKey) {
        return { skipped: true };
      }
      const resend = new Resend(apiKey);
      const users = await pool.query<{ email: string | null; clerk_id: string }>(
        `SELECT email, clerk_id FROM users WHERE email IS NOT NULL LIMIT 50`
      );
      const from = process.env.RESEND_FROM_EMAIL ?? "onboarding@resend.dev";
      for (const u of users.rows) {
        if (!u.email) {
          continue;
        }
        await resend.emails.send({
          from,
          to: u.email,
          subject: "Your Perch Brief weekly digest",
          html: `<p>Hi — here is your weekly nudge to add a few more snippets in <strong>Perch Brief</strong>.</p>`,
        });
      }
      return { sent: users.rows.length };
    });
  }
);

function truncateMeta(s: string, max: number): string {
  if (s.length <= max) return s;
  return s.slice(0, max);
}
