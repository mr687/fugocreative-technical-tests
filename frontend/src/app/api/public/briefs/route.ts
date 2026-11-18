import { NextResponse } from "next/server";

const backendApiURL = process.env.BACKEND_API_URL ?? "http://localhost:8080/api";

export async function POST(request: Request) {
  return serverApiProxy(request);
}

async function serverApiProxy(request: Request) {
  let body: string;
  try {
    body = await request.text();
  } catch {
    return NextResponse.json(
      { error: { code: "INVALID_REQUEST", message: "Periksa kembali input." } },
      { status: 400 },
    );
  }

  try {
    const response = await fetch(`${backendApiURL.replace(/\/$/, "")}/public/briefs`, {
      method: "POST",
      headers: { "Content-Type": request.headers.get("content-type") ?? "application/json" },
      body,
      cache: "no-store",
    });

    return new NextResponse(await response.text(), {
      status: response.status,
      headers: { "Content-Type": response.headers.get("content-type") ?? "application/json" },
    });
  } catch {
    return NextResponse.json(
      { error: { code: "UPSTREAM_UNAVAILABLE", message: "Layanan sedang tidak tersedia. Coba lagi sebentar." } },
      { status: 502 },
    );
  }
}
