#!/usr/bin/env node
import { copyFileSync, existsSync } from "fs";
import { dirname, join } from "path";
import { fileURLToPath } from "url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const env = join(root, ".env");
const webEnv = join(root, "apps/web/.env.local");

if (!existsSync(env)) {
  copyFileSync(join(root, ".env.example"), env);
  console.log("Created .env from .env.example");
} else {
  console.log(".env already exists");
}

if (!existsSync(webEnv)) {
  copyFileSync(env, webEnv);
  console.log("Created apps/web/.env.local from .env");
} else {
  console.log("apps/web/.env.local already exists");
}

await import("./check-env.mjs");
