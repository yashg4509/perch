# init-signals scenario

**Detection-only:** this folder has `vercel.json` and `package.json` dependencies (`@supabase/supabase-js`, `stripe`, `openai`) so `perch init` can infer providers and edges.

`perch.yaml` is **gitignored** here so the repo stays the source of truth for signals; generate it locally:

```bash
/path/to/perch init --name perch-init-demo
```

Then inspect `perch.yaml` and run `perch graph --json` from this directory.

Full instructions: [../../README.md](../../README.md).
