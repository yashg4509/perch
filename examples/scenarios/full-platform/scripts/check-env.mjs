#!/usr/bin/env node
import { readFileSync, existsSync } from "fs";
import { dirname, join } from "path";
import { fileURLToPath } from "url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const envPath = join(root, ".env");

const required = [
  ["NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY", "Clerk → API keys"],
  ["CLERK_SECRET_KEY", "Clerk → API keys"],
  ["DATABASE_URL", "Neon → connection string (must include neon.tech or start Docker Postgres)"],
  ["OPENAI_API_KEY", "platform.openai.com → API keys"],
  ["PINECONE_API_KEY", "app.pinecone.io"],
  ["PINECONE_INDEX", "Pinecone index name"],
  ["ANTHROPIC_API_KEY", "console.anthropic.com (or set ANSWER_MODEL=openai in .env)"],
];

function parseEnv(raw) {
  const m = {};
  for (const line of raw.split("\n")) {
    const t = line.trim();
    if (!t || t.startsWith("#")) continue;
    const i = t.indexOf("=");
    if (i === -1) continue;
    m[t.slice(0, i).trim()] = t.slice(i + 1).trim();
  }
  return m;
}

if (!existsSync(envPath)) {
  console.error("Missing .env — run: npm run setup");
  process.exit(1);
}

const env = parseEnv(readFileSync(envPath, "utf8"));
const missing = [];

for (const [key, hint] of required) {
  if (key === "ANTHROPIC_API_KEY" && env.ANSWER_MODEL === "openai") continue;
  const v = env[key];
  if (!v || v === '""' || v === "''") missing.push({ key, hint });
}

if (missing.length) {
  console.error("\nFill these in .env before npm run dev:\n");
  for (const { key, hint } of missing) {
    console.error(`  ${key}  (${hint})`);
  }
  console.error("\nAccount links: README.md#accounts\n");
  process.exit(1);
}

console.log("Required .env keys look set.");
