import { z } from "zod";

export const budgetRanges = [
  "Di bawah Rp10 juta",
  "Rp10–25 juta",
  "Rp25–50 juta",
  "Di atas Rp50 juta",
] as const;

export function getLocalDate() {
  const today = new Date();
  const month = String(today.getMonth() + 1).padStart(2, "0");
  const day = String(today.getDate()).padStart(2, "0");
  return `${today.getFullYear()}-${month}-${day}`;
}

export const briefSchema = z.object({
  contact_name: z.string().trim().min(1, "Nama kontak wajib diisi.").max(200, "Nama maksimal 200 karakter."),
  email: z.string().trim().min(1, "Email wajib diisi.").email("Masukkan alamat email yang valid.").max(320, "Email maksimal 320 karakter."),
  company_name: z.string().trim().min(1, "Nama perusahaan wajib diisi.").max(200, "Nama maksimal 200 karakter."),
  need_description: z.string().trim().min(1, "Ceritakan kebutuhan Anda.").max(5000, "Kebutuhan maksimal 5.000 karakter."),
  budget_range: z.enum(budgetRanges, { error: "Pilih rentang budget." }),
  target_date: z.iso.date({ error: "Pilih tanggal yang valid." }).refine((date) => date >= getLocalDate(), {
    message: "Target tanggal tidak boleh sebelum hari ini.",
  }),
});

export type BriefFormValues = z.infer<typeof briefSchema>;
