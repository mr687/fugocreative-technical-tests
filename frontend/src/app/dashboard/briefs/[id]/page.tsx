import { Suspense } from "react";
import { getServerSession } from "next-auth";
import { redirect } from "next/navigation";
import { BriefDetail } from "@/components/brief-detail";
import { authOptions } from "@/lib/auth";

type PageProps = { params: Promise<{ id: string }> };

export default function BriefDetailPage({ params }: PageProps) {
  return (
    <Suspense fallback={<main className="p-8 text-sm text-muted-foreground">Memuat detail brief...</main>}>
      <ProtectedBriefDetail params={params} />
    </Suspense>
  );
}

async function ProtectedBriefDetail({ params }: PageProps) {
  const session = await getServerSession(authOptions);
  if (!session) redirect("/login?callbackUrl=%2Fdashboard");
  const { id } = await params;
  return <BriefDetail id={id} username={session.user?.name ?? "admin"} />;
}
