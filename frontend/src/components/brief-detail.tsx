"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { signOut } from "next-auth/react";
import { useRouter } from "next/navigation";
import { ArrowLeft, LogOut, RefreshCw, RotateCcw, WandSparkles } from "lucide-react";
import { Dialog } from "@base-ui/react/dialog";
import { toast } from "sonner";
import { Button, buttonVariants } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

type AIResult = {
  summary: string;
  service_category: string;
  priority: string;
  priority_reason: string;
  missing_info: string[];
};

type EmailDraft = { subject: string; body: string };

type BriefDetailData = {
  id: string;
  contact_name: string;
  email: string;
  company_name: string;
  need_description: string;
  budget_range: string;
  target_date: string;
  ai_status: string;
  ai_result: AIResult | null;
  ai_email_draft: EmailDraft | null;
  pipeline_status: string;
  latest_job: { id: string; status: string; attempt: number; error: string | null } | null;
  created_at: string;
  updated_at: string;
};

type APIError = { error?: { message?: string } };

export function BriefDetail({ id, username }: { id: string; username: string }) {
  const router = useRouter();
  const [brief, setBrief] = useState<BriefDetailData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRetrying, setIsRetrying] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const [isDraftDialogOpen, setIsDraftDialogOpen] = useState(false);
  const [isGeneratingDraft, setIsGeneratingDraft] = useState(false);
  const [emailSubject, setEmailSubject] = useState("");
  const [emailBody, setEmailBody] = useState("");
  const isProcessing = brief?.latest_job?.status === "processing";

  const handleUnauthorized = useCallback(async () => {
    toast.error("Sesi berakhir. Silakan login kembali.");
    await signOut({ redirect: false });
    router.replace("/login");
  }, [router]);

  useEffect(() => {
    const controller = new AbortController();
    fetch(`/api/admin/briefs/${encodeURIComponent(id)}`, { cache: "no-store", signal: controller.signal })
      .then(async (response) => {
        if (response.status === 401) {
          await handleUnauthorized();
          return null;
        }
        const payload = (await response.json().catch(() => null)) as BriefDetailData & APIError | null;
        if (!response.ok) throw new Error(payload?.error?.message ?? "Detail brief gagal dimuat.");
        return payload as BriefDetailData;
      })
      .then((payload) => {
        if (payload) {
          setBrief(payload);
          setError(null);
        }
      })
      .catch((cause: unknown) => {
        if (cause instanceof DOMException && cause.name === "AbortError") return;
        const message = cause instanceof Error ? cause.message : "Detail brief gagal dimuat.";
        setError(message);
        toast.error(message);
      })
      .finally(() => {
        if (!controller.signal.aborted) setIsLoading(false);
      });

    return () => controller.abort();
  }, [handleUnauthorized, id, refreshKey]);

  useEffect(() => {
    if (!isProcessing) return;
    const deadline = Date.now() + 60_000;
    const controller = new AbortController();
    let stopped = false;
    let pollTimer: ReturnType<typeof setTimeout>;

    const schedulePoll = () => {
      if (!stopped && Date.now() < deadline) {
        pollTimer = setTimeout(poll, Math.min(5_000, deadline - Date.now()));
      }
    };

    const poll = async () => {
      if (stopped || controller.signal.aborted) return;
      try {
        const response = await fetch(`/api/admin/briefs/${encodeURIComponent(id)}`, {
          cache: "no-store",
          signal: controller.signal,
        });
        if (response.status === 401) {
          stopped = true;
          await handleUnauthorized();
          return;
        }
        if (!response.ok) throw new Error("Status brief gagal diperbarui.");
        const payload = (await response.json()) as BriefDetailData;
        setBrief(payload);
        if (payload.latest_job?.status !== "processing") stopped = true;
        else schedulePoll();
      } catch {
        if (!controller.signal.aborted && Date.now() < deadline) schedulePoll();
      }
    };

    schedulePoll();
    const deadlineTimer = setTimeout(() => {
      stopped = true;
      controller.abort();
      toast.info("Pembaruan otomatis berhenti setelah 1 menit. Muat ulang untuk cek lagi.");
    }, 60_000);

    return () => {
      stopped = true;
      clearTimeout(pollTimer);
      clearTimeout(deadlineTimer);
      controller.abort();
    };
  }, [handleUnauthorized, id, isProcessing, refreshKey]);

  async function retryAI() {
    setIsRetrying(true);
    try {
      const response = await fetch(`/api/admin/briefs/${encodeURIComponent(id)}/reprocess`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: "{}",
      });
      if (response.status === 401) {
        await handleUnauthorized();
        return;
      }
      const payload = (await response.json().catch(() => null)) as
        | { job_id: string; ai_status: string; job_status: string }
        | APIError
        | null;
      if (!response.ok || !payload || !("job_id" in payload)) {
        throw new Error((payload as APIError | null)?.error?.message ?? "Retry pemrosesan AI gagal.");
      }
      setBrief((current) => current ? {
        ...current,
        ai_status: payload.ai_status,
        ai_result: null,
        latest_job: { id: payload.job_id, status: payload.job_status, attempt: 1, error: null },
      } : current);
      toast.success("Brief masuk antrean pemrosesan AI.");
    } catch (cause) {
      toast.error(cause instanceof Error ? cause.message : "Retry pemrosesan AI gagal.");
    } finally {
      setIsRetrying(false);
    }
  }
  const jobActive = brief?.latest_job?.status === "queued" || isProcessing;

  function openEmailDraft() {
    setEmailSubject(brief?.ai_email_draft?.subject ?? "");
    setEmailBody(brief?.ai_email_draft?.body ?? "");
    setIsDraftDialogOpen(true);
  }

  async function generateEmailDraft() {
    setIsGeneratingDraft(true);
    try {
      const response = await fetch(`/api/admin/briefs/${encodeURIComponent(id)}/email-draft`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: "{}",
      });
      if (response.status === 401) {
        await handleUnauthorized();
        return;
      }
      const payload = (await response.json().catch(() => null)) as ({ ai_email_draft?: EmailDraft } & APIError) | null;
      const draft = payload?.ai_email_draft;
      if (!response.ok || !draft || typeof draft.subject !== "string" || typeof draft.body !== "string") {
        throw new Error(payload?.error?.message ?? "Draft email gagal dibuat.");
      }
      setEmailSubject(draft.subject);
      setEmailBody(draft.body);
      setBrief((current) => current ? { ...current, ai_email_draft: draft } : current);
      toast.success("Draft email berhasil dibuat.");
    } catch (cause) {
      toast.error(cause instanceof Error ? cause.message : "Draft email gagal dibuat.");
    } finally {
      setIsGeneratingDraft(false);
    }
  }

  const canSendEmail = emailSubject.trim() !== "" && emailBody.trim() !== "";
  const canGenerateEmailDraft = Boolean(
    brief?.ai_result &&
    brief.ai_result.summary.trim() &&
    Array.isArray(brief.ai_result.missing_info),
  );
  const mailtoHref = brief
    ? `mailto:${brief.email}?subject=${encodeURIComponent(emailSubject)}&body=${encodeURIComponent(emailBody)}`
    : undefined;


  return (
    <main className="min-h-screen bg-muted/35">
      <header className="border-b bg-background">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-8">
          <Link aria-label="Fugo Creative" className="text-xl font-bold tracking-tight" href="/dashboard">fugo<span className="text-indigo-600">.</span></Link>
          <div className="flex items-center gap-3 text-sm">
            <span className="hidden text-muted-foreground sm:inline">{username}</span>
            <Button variant="outline" size="sm" onClick={() => signOut({ callbackUrl: "/login" })}>
              <LogOut className="mr-2 size-4" aria-hidden="true" /> Keluar
            </Button>
          </div>
        </div>
      </header>

      <div className="mx-auto max-w-5xl space-y-6 px-4 py-8 sm:px-8 sm:py-12">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Link className="-ml-3 inline-flex h-8 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium transition-colors hover:bg-muted" href="/dashboard">
            <ArrowLeft className="size-4" aria-hidden="true" />Kembali ke daftar
          </Link>
          <Button variant="outline" disabled={isLoading} onClick={() => { setIsLoading(true); setRefreshKey((value) => value + 1); }}>
            <RefreshCw className={`mr-2 size-4 ${isLoading ? "animate-spin" : ""}`} aria-hidden="true" />Muat ulang
          </Button>
        </div>

        {isLoading && !brief ? (
          <Card><CardContent className="py-12 text-center text-muted-foreground">Memuat detail brief...</CardContent></Card>
        ) : error && !brief ? (
          <Card><CardContent className="py-12 text-center"><p className="text-destructive">{error}</p></CardContent></Card>
        ) : brief ? (
          <>
            <div className="flex flex-wrap items-end justify-between gap-4">
              <div>
                <p className="text-sm font-medium text-indigo-700">Detail brief</p>
                <h1 className="mt-2 text-3xl font-semibold tracking-tight">{brief.company_name}</h1>
                <p className="mt-2 text-muted-foreground">Dikirim {formatDateTime(brief.created_at)}</p>
              </div>
              <div className="flex items-center gap-2">
                <StatusPill value={brief.ai_status} />
                <Dialog.Root
                  open={isDraftDialogOpen}
                  onOpenChange={setIsDraftDialogOpen}
                  triggerId="email-draft-trigger"
                >
                  <Button id="email-draft-trigger" variant="outline" onClick={openEmailDraft}>
                    <WandSparkles className="mr-2 size-4" aria-hidden="true" />
                    Draft Email
                  </Button>
                  <Dialog.Portal>
                    <Dialog.Backdrop className="fixed inset-0 z-50 bg-black/50" />
                    <Dialog.Popup className="fixed left-1/2 top-1/2 z-50 grid max-h-[calc(100vh-2rem)] w-[calc(100%-2rem)] max-w-xl -translate-x-1/2 -translate-y-1/2 gap-5 overflow-y-auto rounded-xl border bg-background p-6 shadow-lg">
                      <header className="space-y-1">
                        <Dialog.Title className="text-lg font-semibold">Draft Email</Dialog.Title>
                        <Dialog.Description className="text-sm text-muted-foreground">
                          Tinjau dan edit draft sebelum membuka aplikasi email.
                        </Dialog.Description>
                      </header>
                      <div className="grid gap-4">
                        <div className="grid gap-2">
                          <Label htmlFor="draft-email-to">Email tujuan</Label>
                          <Input id="draft-email-to" type="email" value={brief.email} disabled />
                        </div>
                        <div className="grid gap-2">
                          <Label htmlFor="draft-email-subject">Subject</Label>
                          <Input
                            id="draft-email-subject"
                            value={emailSubject}
                            onChange={(event) => setEmailSubject(event.target.value)}
                          />
                        </div>
                        <div className="grid gap-2">
                          <Label htmlFor="draft-email-body">Body</Label>
                          <Textarea
                            id="draft-email-body"
                            className="min-h-56 resize-y"
                            value={emailBody}
                            onChange={(event) => setEmailBody(event.target.value)}
                          />
                        </div>
                        {!brief.ai_email_draft && !brief.ai_result ? (
                          <p className="text-sm text-muted-foreground">Draft dapat dibuat setelah ringkasan AI dan informasi yang dibutuhkan tersedia.</p>
                        ) : null}
                      </div>
                      <footer className="flex flex-wrap justify-end gap-2">
                        <Button type="button" variant="outline" onClick={() => setIsDraftDialogOpen(false)}>
                          Cancel
                        </Button>
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={generateEmailDraft}
                          disabled={isGeneratingDraft || !canGenerateEmailDraft}
                        >
                          {isGeneratingDraft ? "Membuat draft..." : brief.ai_email_draft ? "Generate ulang" : "Generate"}
                        </Button>
                        <a
                          className={buttonVariants({ className: canSendEmail ? "" : "pointer-events-none opacity-50" })}
                          href={canSendEmail ? mailtoHref : undefined}
                          aria-disabled={!canSendEmail}
                        >
                          Kirim
                        </a>
                      </footer>
                    </Dialog.Popup>
                  </Dialog.Portal>
                </Dialog.Root>
                <Button onClick={retryAI} disabled={isRetrying || jobActive}>
                  <RotateCcw className={`mr-2 size-4 ${isRetrying ? "animate-spin" : ""}`} aria-hidden="true" />
                  {isRetrying ? "Mengantrekan..." : "Retry LLM summary"}
                </Button>
              </div>
            </div>

            <Card>
              <CardHeader>
                <CardTitle>Informasi brief</CardTitle>
                <CardDescription>Informasi kontak, kebutuhan, dan status pipeline.</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-5 sm:grid-cols-2">
                <DetailField label="Nama kontak" value={brief.contact_name} />
                <DetailField label="Email" value={brief.email} />
                <DetailField label="Perusahaan" value={brief.company_name} />
                <DetailField label="Budget" value={brief.budget_range} />
                <DetailField label="Target tanggal" value={formatDate(brief.target_date)} />
                <DetailField label="Pipeline" value={brief.pipeline_status} />
                <div className="sm:col-span-2"><DetailField label="Kebutuhan" value={brief.need_description} /></div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Ringkasan AI</CardTitle>
                <CardDescription>Status pemrosesan: {brief.latest_job?.status ?? brief.ai_status}{brief.latest_job ? ` · percobaan ${brief.latest_job.attempt}` : ""}</CardDescription>
              </CardHeader>
              <CardContent>
                {brief.ai_result ? (
                  <div className="space-y-5">
                    <DetailField label="Ringkasan" value={brief.ai_result.summary} />
                    <div className="grid gap-5 sm:grid-cols-2">
                      <DetailField label="Kategori layanan" value={brief.ai_result.service_category} />
                      <DetailField label="Prioritas" value={brief.ai_result.priority} />
                    </div>
                    <DetailField label="Alasan prioritas" value={brief.ai_result.priority_reason} />
                    <div>
                      <p className="text-sm font-medium">Informasi yang masih dibutuhkan</p>
                      {brief.ai_result.missing_info.length ? (
                        <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-muted-foreground">
                          {brief.ai_result.missing_info.map((item, index) => <li key={`${index}-${item}`}>{item}</li>)}
                        </ul>
                      ) : <p className="mt-2 text-sm text-muted-foreground">Tidak ada.</p>}
                    </div>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">
                    {brief.ai_status === "failed" ? "Pemrosesan gagal. Anda dapat meminta retry." : "Ringkasan belum tersedia."}
                  </p>
                )}
                {brief.latest_job?.error ? <p className="mt-4 text-sm text-destructive">{brief.latest_job.error}</p> : null}
              </CardContent>
            </Card>
          </>
        ) : null}
      </div>
    </main>
  );
}

function DetailField({ label, value }: { label: string; value: string }) {
  return (
    <div className="space-y-1">
      <p className="text-sm font-medium">{label}</p>
      <p className="whitespace-pre-wrap text-sm text-muted-foreground">{value}</p>
    </div>
  );
}

function StatusPill({ value }: { value: string }) {
  const tone = value === "failed"
    ? "bg-destructive/10 text-destructive"
    : value === "succeeded"
      ? "bg-emerald-100 text-emerald-800"
      : "bg-muted text-muted-foreground";
  return <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${tone}`}>{value}</span>;
}

function formatDate(value: string) {
  const [year, month, day] = value.split("-").map(Number);
  return new Intl.DateTimeFormat("id-ID", { dateStyle: "long" }).format(new Date(year, month - 1, day));
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
