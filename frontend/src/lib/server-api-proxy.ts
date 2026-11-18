import { getToken } from "next-auth/jwt";
import type { NextRequest } from "next/server";

const backendApiURL = process.env.BACKEND_API_URL ?? "http://localhost:8080/api";

async function getBackendAccessToken(request: Request) {
  const token = await getToken({ req: request as NextRequest, secret: process.env.NEXTAUTH_SECRET });
  return typeof token?.backendAccessToken === "string" ? token.backendAccessToken : null;
}

export async function serverApiProxy(request: Request, path: string, init: RequestInit = {}) {
  const accessToken = await getBackendAccessToken(request);
  if (!accessToken) {
    return Response.json(
      { error: { code: "UNAUTHORIZED", message: "Sesi berakhir. Silakan login kembali." } },
      { status: 401 },
    );
  }

  try {
    return await fetch(`${backendApiURL.replace(/\/$/, "")}${path}`, {
      ...init,
      headers: {
        ...Object.fromEntries(new Headers(init.headers).entries()),
        Authorization: `Bearer ${accessToken}`,
      },
      cache: "no-store",
    });
  } catch {
    return Response.json(
      { error: { code: "UPSTREAM_UNAVAILABLE", message: "Layanan sedang tidak tersedia. Coba lagi sebentar." } },
      { status: 502 },
    );
  }
}
