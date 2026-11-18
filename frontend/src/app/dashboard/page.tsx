import { Suspense } from "react";
import { getServerSession } from "next-auth";
import { redirect } from "next/navigation";
import { Dashboard } from "@/components/dashboard";
import { authOptions } from "@/lib/auth";

export default function DashboardPage() {
  return <Suspense fallback={<main className="p-8 text-sm text-muted-foreground">Memeriksa sesi...</main>}><ProtectedDashboard /></Suspense>;
}

async function ProtectedDashboard() {
  const session = await getServerSession(authOptions);
  if (!session) redirect("/login?callbackUrl=%2Fdashboard");
  return <Dashboard username={session.user?.name ?? "admin"} />;
}
