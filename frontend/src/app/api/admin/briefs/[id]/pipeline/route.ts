import { serverApiProxy } from "@/lib/server-api-proxy";

type RouteContext = { params: Promise<{ id: string }> };

export async function PATCH(request: Request, { params }: RouteContext) {
  const { id } = await params;
  return serverApiProxy(request, `/admin/briefs/${encodeURIComponent(id)}/pipeline`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: await request.text(),
  });
}
