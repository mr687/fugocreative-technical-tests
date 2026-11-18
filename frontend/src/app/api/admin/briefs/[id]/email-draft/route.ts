import { serverApiProxy } from "@/lib/server-api-proxy";

type RouteContext = { params: Promise<{ id: string }> };

export async function POST(request: Request, { params }: RouteContext) {
  const { id } = await params;
  return serverApiProxy(request, `/admin/briefs/${encodeURIComponent(id)}/email-draft`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  });
}
