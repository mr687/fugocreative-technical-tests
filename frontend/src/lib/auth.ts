import type { NextAuthOptions } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

const backendApiURL = process.env.BACKEND_API_URL ?? "http://localhost:8080/api";

export const authOptions: NextAuthOptions = {
  secret: process.env.NEXTAUTH_SECRET,
  session: { strategy: "jwt", maxAge: 60 * 60 },
  pages: { signIn: "/login" },
  providers: [
    CredentialsProvider({
      name: "Admin",
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        if (!credentials?.username || !credentials.password) return null;
        try {
          const response = await fetch(`${backendApiURL.replace(/\/$/, "")}/auth/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: credentials.username, password: credentials.password }),
            cache: "no-store",
          });
          if (!response.ok) return null;
          const payload = (await response.json()) as { access_token?: string };
          if (!payload.access_token) return null;
          return {
            id: credentials.username,
            name: credentials.username,
            backendAccessToken: payload.access_token,
          };
        } catch {
          return null;
        }
      },
    }),
  ],
  callbacks: {
    async jwt({ token, user }) {
      if (user && "backendAccessToken" in user) {
        token.backendAccessToken = user.backendAccessToken;
      }
      return token;
    },
    async session({ session, token }) {
      if (session.user) session.user.name = token.name ?? "admin";
      return session;
    },
  },
};
