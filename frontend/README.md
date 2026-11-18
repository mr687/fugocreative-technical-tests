# Fugo Creative Brief Intake — Frontend

Next.js app for public brief submission and protected admin brief review.

## Local development

```sh
bun install
cp .env.example .env.local
bun run dev
```

Set `BACKEND_API_URL` to backend API base URL (default `http://localhost:8080/api`). It is read only by the Next.js server; browser requests go through Next.js route handlers. Set `NEXTAUTH_SECRET` to a unique random secret and `NEXTAUTH_URL` to the frontend origin. The admin login validates credentials against backend `POST /auth/login`; the backend access token stays in the encrypted server-side NextAuth JWT cookie and is never exposed in the browser session.

For local smoke testing, use the backend-provided test account `admin` / `admin123`. Do not use that password in production.

The public form accepts budgets: Di bawah Rp10 juta, Rp10–25 juta, Rp25–50 juta, and Di atas Rp50 juta.

Admin dashboard: `/dashboard`. Its server-rendered route checks the NextAuth session, and `src/proxy.ts` redirects requests without a valid session to `/login`. Admin data requests go through `/api/admin/briefs`; backend `401` responses expire the session in the browser and redirect to login.
