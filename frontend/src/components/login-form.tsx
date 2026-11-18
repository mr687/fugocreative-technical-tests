"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { signIn } from "next-auth/react";
import { useRouter } from "next/navigation";
import { ArrowRight, LockKeyhole } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { loginSchema, type LoginValues } from "@/lib/login-schema";
export function LoginForm() {
  const router = useRouter();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: "", password: "" },
  });

  async function submit(values: LoginValues) {
    try {
      const result = await signIn("credentials", { ...values, redirect: false });
      if (!result || result.error) {
        toast.error("Username atau password salah.");
        return;
      }
      const callbackUrl = new URLSearchParams(window.location.search).get("callbackUrl");
      const destination = callbackUrl?.startsWith("/") && !callbackUrl.startsWith("//") ? callbackUrl : "/dashboard";
      router.replace(destination);
      router.refresh();
    } catch {
      toast.error("Login gagal. Pastikan layanan tersedia, lalu coba lagi.");
    }
  }

  return (
    <Card className="w-full max-w-md border-border/80 shadow-lg shadow-black/5">
      <CardHeader className="space-y-3">
        <div className="flex size-11 items-center justify-center rounded-xl bg-primary text-primary-foreground">
          <LockKeyhole className="size-5" aria-hidden="true" />
        </div>
        <div>
          <CardTitle className="text-2xl">Masuk ke dashboard</CardTitle>
          <CardDescription className="mt-2">Gunakan akun admin untuk mengelola brief masuk.</CardDescription>
        </div>
      </CardHeader>
      <CardContent>
        <form className="space-y-5" noValidate onSubmit={handleSubmit(submit)}>
          <Field data-invalid={!!errors.username}>
            <FieldLabel htmlFor="username">Username</FieldLabel>
            <Input
              autoComplete="username"
              aria-describedby={errors.username ? "username-error" : undefined}
              aria-invalid={!!errors.username}
              id="username"
              {...register("username")}
            />
            <FieldError id="username-error" errors={[errors.username]} />
          </Field>
          <Field data-invalid={!!errors.password}>
            <FieldLabel htmlFor="password">Password</FieldLabel>
            <Input
              autoComplete="current-password"
              aria-describedby={errors.password ? "password-error" : undefined}
              aria-invalid={!!errors.password}
              id="password"
              type="password"
              {...register("password")}
            />
            <FieldError id="password-error" errors={[errors.password]} />
          </Field>
          <Button className="w-full" disabled={isSubmitting} type="submit">
            {isSubmitting ? "Memeriksa..." : "Masuk"}
            {!isSubmitting && <ArrowRight className="ml-2 size-4" aria-hidden="true" />}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
