import Link from "next/link";
import { BriefForm } from "@/components/brief-form";

export default function Home() {
  return (
    <main className="min-h-screen bg-[#f4f6f8]">
      <header className="border-b border-slate-200/80 bg-white">
        <div className="mx-auto flex h-16 max-w-6xl items-center px-5 sm:px-8">
          <Link aria-label="Fugo Creative home" className="text-lg font-bold tracking-tight text-slate-950" href="/">
            fugo<span className="text-indigo-600">.</span>
          </Link>
          <span className="ml-3 border-l border-slate-200 pl-3 text-sm text-slate-500">Creative studio</span>
        </div>
      </header>
      <div className="mx-auto grid max-w-6xl gap-10 px-5 py-10 sm:px-8 sm:py-16 lg:grid-cols-[0.8fr_1.2fr] lg:gap-16">
        <section className="pt-2 lg:sticky lg:top-12 lg:h-fit">
          <p className="text-sm font-semibold uppercase tracking-[0.16em] text-indigo-700">Let’s make it real</p>
          <h1 className="mt-4 max-w-lg text-4xl font-semibold leading-tight tracking-tight text-slate-950 sm:text-5xl">
            Ceritakan ide Anda. Kami bantu wujudkan.
          </h1>
          <p className="mt-5 max-w-md text-base leading-7 text-slate-600">
            Bagikan kebutuhan proyek Anda kepada tim Fugo Creative. Kami akan meninjau brief dan menghubungi Anda untuk langkah selanjutnya.
          </p>
          <div className="mt-8 border-l-2 border-indigo-500 pl-4 text-sm leading-6 text-slate-600">
            Tidak perlu brief sempurna. Beri tahu kami tujuan, konteks, dan target waktu yang sudah Anda ketahui.
          </div>
        </section>
        <section aria-labelledby="brief-form-title">
          <div className="mb-5">
            <h2 className="text-xl font-semibold tracking-tight text-slate-950" id="brief-form-title">Brief proyek</h2>
            <p className="mt-1 text-sm text-slate-600">Isi detail berikut agar kami memahami kebutuhan Anda.</p>
          </div>
          <BriefForm />
        </section>
      </div>
      <footer className="border-t border-slate-200/80 py-6 text-center text-xs text-slate-500">
        © Fugo Creative
      </footer>
    </main>
  );
}
