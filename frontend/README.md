# Fugo Creative Brief Intake — Frontend

Next.js app for submitting public project briefs.

## Local development

```sh
bun install
cp .env.example .env.local
bun run dev
```

Set `BACKEND_API_URL` to backend API base URL (default `http://localhost:8080/api`). It is read only by the Next.js server; browser requests go through `/api/public/briefs`.

The public form accepts budgets: Di bawah Rp10 juta, Rp10–25 juta, Rp25–50 juta, and Di atas Rp50 juta.
