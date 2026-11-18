"use client";

import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowRight, CheckCircle2 } from "lucide-react";
import { toast } from "sonner";
import { briefSchema, budgetRanges, type BriefFormValues } from "@/lib/brief-schema";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";

type APIError = { error?: { message?: string } };

export function BriefForm() {
  const [submitted, setSubmitted] = useState(false);
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<BriefFormValues>({
    resolver: zodResolver(briefSchema),
    defaultValues: {
      contact_name: "",
      email: "",
      company_name: "",
      need_description: "",
      budget_range: undefined,
      target_date: "",
    },
  });

  async function submitBrief(values: BriefFormValues) {
    try {
      const response = await fetch("/api/public/briefs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(values),
      });
      const result = (await response.json().catch(() => null)) as APIError | null;
      if (!response.ok) {
        throw new Error(result?.error?.message ?? "Brief belum terkirim. Periksa koneksi lalu coba lagi.");
      }
      setSubmitted(true);
      toast.success("Brief berhasil dikirim", {
        description: "Tim kami akan meninjau kebutuhan Anda.",
      });
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Brief belum terkirim. Coba lagi sebentar.");
    }
  }

  if (submitted) {
    return (
      <Card className="border-emerald-200 bg-white shadow-sm">
        <CardContent className="flex flex-col items-center px-6 py-12 text-center sm:px-10">
          <CheckCircle2 aria-hidden="true" className="mb-4 size-12 text-emerald-600" />
          <h2 className="text-2xl font-semibold tracking-tight">Brief Anda sudah kami terima</h2>
          <p className="mt-2 max-w-md text-sm leading-6 text-muted-foreground">
            Data tersimpan. Tim kami akan meninjau kebutuhan Anda dan menghubungi melalui email.
          </p>
          <Button
            className="mt-7"
            onClick={() => {
              reset();
              setSubmitted(false);
            }}
            type="button"
            variant="outline"
          >
            Kirim brief lainnya
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-border/80 bg-white shadow-sm">
      <CardContent className="p-6 sm:p-8">
        <form noValidate onSubmit={handleSubmit(submitBrief)}>
          <FieldGroup>
            <div className="grid gap-5 sm:grid-cols-2">
              <Field data-invalid={!!errors.contact_name}>
                <FieldLabel htmlFor="contact_name">Nama kontak</FieldLabel>
                <Input
                  autoComplete="name"
                  id="contact_name"
                  placeholder="Contoh: Ayu Putri"
                  aria-invalid={!!errors.contact_name}
                  aria-describedby={errors.contact_name ? "contact_name-error" : undefined}
                  {...register("contact_name")}
                />
                {errors.contact_name && <FieldError id="contact_name-error" errors={[errors.contact_name]} />}
              </Field>
              <Field data-invalid={!!errors.email}>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input
                  autoComplete="email"
                  id="email"
                  placeholder="nama@perusahaan.com"
                  type="email"
                  aria-invalid={!!errors.email}
                  aria-describedby={errors.email ? "email-error" : undefined}
                  {...register("email")}
                />
                {errors.email && <FieldError id="email-error" errors={[errors.email]} />}
              </Field>
              <Field data-invalid={!!errors.company_name}>
                <FieldLabel htmlFor="company_name">Nama perusahaan</FieldLabel>
                <Input
                  autoComplete="organization"
                  id="company_name"
                  placeholder="Nama perusahaan Anda"
                  aria-invalid={!!errors.company_name}
                  aria-describedby={errors.company_name ? "company_name-error" : undefined}
                  {...register("company_name")}
                />
                {errors.company_name && <FieldError id="company_name-error" errors={[errors.company_name]} />}
              </Field>
              <Field data-invalid={!!errors.budget_range}>
                <FieldLabel htmlFor="budget_range">Rentang budget</FieldLabel>
                <Controller
                  control={control}
                  name="budget_range"
                  render={({ field }) => (
                    <Select value={field.value || null} onValueChange={(value) => field.onChange(value ?? "")}>
                      <SelectTrigger
                        id="budget_range"
                        aria-invalid={!!errors.budget_range}
                        aria-describedby={errors.budget_range ? "budget_range-error" : undefined}
                        onBlur={field.onBlur}
                      >
                        <SelectValue placeholder="Pilih rentang budget" />
                      </SelectTrigger>
                      <SelectContent>
                        {budgetRanges.map((range) => (
                          <SelectItem key={range} value={range}>{range}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
                {errors.budget_range && <FieldError id="budget_range-error" errors={[errors.budget_range]} />}
              </Field>
            </div>
            <Field data-invalid={!!errors.need_description}>
              <FieldLabel htmlFor="need_description">Ceritakan kebutuhan Anda</FieldLabel>
              <Textarea
                id="need_description"
                placeholder="Apa yang ingin Anda bangun atau capai? Sertakan detail yang membantu kami memahami kebutuhan Anda."
                rows={5}
                aria-invalid={!!errors.need_description}
                aria-describedby={errors.need_description ? "need_description-error" : undefined}
                {...register("need_description")}
              />
              {errors.need_description && <FieldError id="need_description-error" errors={[errors.need_description]} />}
            </Field>
            <Field data-invalid={!!errors.target_date}>
              <FieldLabel htmlFor="target_date">Target tanggal mulai</FieldLabel>
              <Input
                id="target_date"
                type="date"
                aria-invalid={!!errors.target_date}
                aria-describedby={errors.target_date ? "target_date-error" : undefined}
                {...register("target_date")}
              />
              {errors.target_date && <FieldError id="target_date-error" errors={[errors.target_date]} />}
            </Field>
          </FieldGroup>
          <p className="mt-6 text-xs leading-5 text-muted-foreground">
            Brief Anda disimpan terlebih dahulu. Analisis kebutuhan diproses setelah pengiriman.
          </p>
          <Button className="mt-5 w-full" disabled={isSubmitting} size="lg" type="submit">
            {isSubmitting ? "Mengirim brief…" : "Kirim brief"}
            {!isSubmitting && <ArrowRight aria-hidden="true" className="ml-2 size-4" />}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
