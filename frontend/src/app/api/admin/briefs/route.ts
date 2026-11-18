import { serverApiProxy } from "@/lib/server-api-proxy";

export async function GET(request: Request) {
  const query = new URL(request.url).search;
  return serverApiProxy(request, `/admin/briefs${query}`);
}
