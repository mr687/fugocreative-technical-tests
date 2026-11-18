"use client";

import Link from "next/link";

import { useCallback, useEffect, useState } from "react";
import { signOut } from "next-auth/react";
import { useRouter } from "next/navigation";
import { LogOut, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Brief = {
  id: string;
  company_name: string;
  service_category: string;
  priority: string;
  ai_status: string;
  pipeline_status: string;
  created_at: string;
};

type BriefPage = { items: Brief[]; page: number; page_size: number; total: number; total_pages: number };
type Filter = { value: string; label: string };

const categories: Filter[] = [
  { value: "Branding", label: "Branding" },
  { value: "Website", label: "Website" },
  { value: "Mobile App", label: "Mobile App" },
  { value: "AI Automation", label: "AI Automation" },
  { value: "Social Media", label: "Social Media" },
  { value: "Lainnya", label: "Lainnya" },
];
const priorities: Filter[] = [
  { value: "High", label: "Tinggi" },
  { value: "Medium", label: "Sedang" },
  { value: "Low", label: "Rendah" },
];
const pipelineStatuses = ["New", "Contacted", "Proposal", "Won", "Lost"] as const;

export function Dashboard({ username }: { username: string }) {
  const router = useRouter();
  const [category, setCategory] = useState("all");
  const [priority, setPriority] = useState("all");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<BriefPage | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [refreshKey, setRefreshKey] = useState(0);
  const [updatingPipelineId, setUpdatingPipelineId] = useState<string | null>(null);

  const handleUnauthorized = useCallback(async () => {
    toast.error("Sesi berakhir. Silakan login kembali.");
    await signOut({ redirect: false });
    router.replace("/login");
  }, [router]);

  useEffect(() => {
    const controller = new AbortController();
    const params = new URLSearchParams({ page: String(page), page_size: "20" });
    if (category !== "all") params.set("category", category);
    if (priority !== "all") params.set("priority", priority);

    fetch(`/api/admin/briefs?${params}`, { cache: "no-store", signal: controller.signal })
      .then(async (response) => {
        if (response.status === 401) {
          await handleUnauthorized();
          return null;
        }
        const payload = await response.json();
        if (!response.ok) throw new Error(payload?.error?.message ?? "Brief gagal dimuat.");
        return payload as BriefPage;
      })
      .then((payload) => {
        if (payload) setData(payload);
      })
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === "AbortError") return;
        toast.error(error instanceof Error ? error.message : "Brief gagal dimuat.");
      })
      .finally(() => {
        if (!controller.signal.aborted) setIsLoading(false);
      });

    return () => controller.abort();
  }, [category, handleUnauthorized, page, priority, refreshKey]);

  function updateFilter(setFilter: (value: string) => void, value: string | null) {
    setIsLoading(true);
    setPage(1);
    setFilter(value ?? "all");
  }

  async function updatePipelineStatus(id: string, pipelineStatus: string) {
    setUpdatingPipelineId(id);
    try {
      const response = await fetch(`/api/admin/briefs/${encodeURIComponent(id)}/pipeline`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ pipeline_status: pipelineStatus }),
      });
      if (response.status === 401) {
        await handleUnauthorized();
        return;
      }
      const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      if (!response.ok) throw new Error(payload?.error?.message ?? "Status pipeline gagal diperbarui.");
      setData((current) => current ? {
        ...current,
        items: current.items.map((item) => item.id === id ? { ...item, pipeline_status: pipelineStatus } : item),
      } : current);
      toast.success("Status pipeline diperbarui.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Status pipeline gagal diperbarui.");
    } finally {
      setUpdatingPipelineId(null);
    }
  }

  return (
    <main className="min-h-screen bg-muted/35">
      <header className="border-b bg-background">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-8">
          <Link aria-label="Fugo Creative" className="text-xl font-bold tracking-tight" href="/">fugo<span className="text-indigo-600">.</span></Link>
          <div className="flex items-center gap-3 text-sm">
            <span className="hidden text-muted-foreground sm:inline">{username}</span>
            <Button variant="outline" size="sm" onClick={() => signOut({ callbackUrl: "/login" })}>
              <LogOut className="mr-2 size-4" aria-hidden="true" /> Keluar
            </Button>
          </div>
        </div>
      </header>

      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-8 sm:py-12">
        <div className="mb-7 flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-indigo-700">Fugo Creative · Admin</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">Brief masuk</h1>
            <p className="mt-2 text-muted-foreground">Tinjau dan prioritaskan kebutuhan calon klien.</p>
          </div>
          <Button variant="outline" onClick={() => { setIsLoading(true); setRefreshKey((key) => key + 1); }} disabled={isLoading}>
            <RefreshCw className={`mr-2 size-4 ${isLoading ? "animate-spin" : ""}`} aria-hidden="true" />
            Muat ulang
          </Button>
        </div>

        <Card>
          <CardHeader className="gap-4 border-b sm:flex-row sm:items-center sm:justify-between">
            <div>
              <CardTitle>Semua brief</CardTitle>
              <CardDescription className="mt-1">{data ? `${data.total} brief ditemukan` : "Daftar brief terbaru"}</CardDescription>
            </div>
            <div className="flex flex-wrap gap-3">
              <Select value={category} onValueChange={(value) => updateFilter(setCategory, value)}>
                <SelectTrigger aria-label="Filter kategori" className="w-48"><SelectValue>{category === "all" ? "Semua Kategori" : undefined}</SelectValue></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Kategori</SelectItem>
                  {categories.map((item) => <SelectItem key={item.value} value={item.value}>{item.label}</SelectItem>)}
                </SelectContent>
              </Select>
              <Select value={priority} onValueChange={(value) => updateFilter(setPriority, value)}>
                <SelectTrigger aria-label="Filter prioritas" className="w-44"><SelectValue>{priority === "all" ? "Semua prioritas" : undefined}</SelectValue></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua prioritas</SelectItem>
                  {priorities.map((item) => <SelectItem key={item.value} value={item.value}>{item.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <Table className="[&_th]:px-4 [&_th]:py-3 [&_td]:px-4 [&_td:not([colspan])]:py-3">
              <TableHeader>
                <TableRow>
                  <TableHead>No.</TableHead><TableHead>Perusahaan</TableHead><TableHead>Kategori</TableHead><TableHead>Prioritas</TableHead>
                  <TableHead>Status AI</TableHead><TableHead>Pipeline</TableHead><TableHead className="text-right">Tanggal masuk</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading && !data ? <TableRow><TableCell className="py-12 text-center text-muted-foreground" colSpan={7}>Memuat brief...</TableCell></TableRow> : null}
                {!isLoading && data?.items.length === 0 ? <TableRow><TableCell className="py-12 text-center text-muted-foreground" colSpan={7}>Belum ada brief untuk filter ini.</TableCell></TableRow> : null}
                {data?.items.map((brief, index) => (
                  <TableRow key={brief.id}>
                    <TableCell className="tabular-nums text-muted-foreground">{((data?.page ?? page) - 1) * (data?.page_size ?? 20) + index + 1}</TableCell>
                    <TableCell className="font-medium"><Link className="hover:underline" href={`/dashboard/briefs/${brief.id}`}>{brief.company_name}</Link></TableCell>
                    <TableCell>{brief.service_category || "Belum dikategorikan"}</TableCell>
                    <TableCell><StatusPill value={brief.priority || "Belum diproses"} /></TableCell>
                    <TableCell><StatusPill value={brief.ai_status} /></TableCell>
                    <TableCell>
                      <Select value={brief.pipeline_status} onValueChange={(value) => { if (value) void updatePipelineStatus(brief.id, value); }}>
                        <SelectTrigger aria-label={`Status pipeline ${brief.company_name}`} className="w-36" disabled={updatingPipelineId !== null}>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {pipelineStatuses.map((status) => <SelectItem key={status} value={status}>{status}</SelectItem>)}
                        </SelectContent>
                      </Select>
                    </TableCell>
                    <TableCell className="text-right text-muted-foreground">{new Intl.DateTimeFormat("id-ID", { dateStyle: "medium" }).format(new Date(brief.created_at))}</TableCell>
                  </TableRow>
                ))}
                {isLoading && data ? <TableRow><TableCell className="py-3 text-center text-muted-foreground" colSpan={7}>Memperbarui daftar...</TableCell></TableRow> : null}
              </TableBody>
            </Table>
            <div className="flex items-center justify-between border-t px-4 py-3 text-sm text-muted-foreground">
              <span>Halaman {data?.page ?? page} dari {data?.total_pages ?? 0}</span>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" disabled={isLoading || page <= 1} onClick={() => { setIsLoading(true); setPage((current) => current - 1); }}>Sebelumnya</Button>
                <Button size="sm" variant="outline" disabled={isLoading || !data || page >= data.total_pages} onClick={() => { setIsLoading(true); setPage((current) => current + 1); }}>Berikutnya</Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}

function StatusPill({ value }: { value: string }) {
  const tone = value === "High" || value === "failed"
    ? "bg-destructive/10 text-destructive"
    : value === "succeeded" || value === "Won"
      ? "bg-emerald-100 text-emerald-800"
      : "bg-muted text-muted-foreground";
  const label = value === "High" ? "Tinggi" : value === "Medium" ? "Sedang" : value === "Low" ? "Rendah" : value;
  return <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${tone}`}>{label}</span>;
}
