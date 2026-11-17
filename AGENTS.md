# AGENTS.md

## Tech stack

- Backend: Go (Golang) menggunakan _framework_ `chi/v5`, _database driver_ `pgx/v5`, dan _migrasi_ `goose`.
- Frontend: Next.js (App router), NextAuth, TypeScript, Bun, shadcn/ui.
- Database: PostgreSQL 18; UUIDv7, JSONB.
- Monorepo: `./backend` dan `./frontend`.
- Ikuti `PRD.md` dan `SPEC.md` untuk kebutuhan dan kontrak.

## Struktur backend

Pertahankan struktur sederhana dan berorientasi fitur:

```text
backend/
├── cmd/server/main.go       # Titik masuk utama
├── internal/
│   ├── config/              # Konfigurasi ENV
│   ├── database/            # Koneksi DB & Goose migration runner
│   ├── handler/             # HTTP Handlers (chi) & validasi request
│   ├── middleware/          # JWT Auth & CORS
│   ├── model/               # Struct data & JSON Schema struct LLM
│   ├── repository/          # Layer SQL query PostgreSQL
│   └── service/             # Business logic & integrasi API LLM
├── migrations/              # File migrasi SQL (.sql)
└── go.mod
```

Hindari package dan abstraksi yang belum dibutuhkan. Pisahkan handler, akses database, dan domain bila membantu keterbacaan.

## Command guidelines

Jalankan command dari folder aplikasi terkait. Gunakan `bun` untuk frontend dan `go` untuk backend.

- Backend dependencies: `cd backend && go mod tidy`
- Backend tests: `cd backend && go test ./...`
- Backend format: `cd backend && gofmt -w .`
- Backend server: `cd backend && go run ./cmd/server`
- Database migration: gunakan migration tool yang benar-benar dipilih/terpasang; dokumentasikan perintah up/down di README. Jangan mengasumsikan tool sebelum dipilih.
- Frontend install: `cd frontend && bun install`
- Frontend dev: `cd frontend && bun run dev`
- Frontend lint/typecheck/build: gunakan script yang tersedia di `frontend/package.json`.
- Jangan menjalankan migrasi destruktif atau menghapus data tanpa persetujuan.
- Jangan commit atau push kecuali diminta.

## Commit guidelines

- Gunakan Conventional Commits: `type(scope): description` (contoh: `feat(api): add brief submission`).
- Tipe yang diizinkan: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.
- Subject imperative, lowercase, tanpa titik di akhir, maksimal 72 karakter.
- Tambahkan body bila perlu untuk menjelaskan alasan; jangan commit kecuali diminta.

## Standar penulisan kode

- Utamakan perubahan minimal, sederhana, dan mudah dibaca; hindari abstraksi spekulatif.
- Validasi semua input di server; gunakan query SQL berparameter.
- Ikuti format idiomatik Go (`gofmt`) dan TypeScript/Next.js yang sudah ada.
- Propagasikan error; jangan abaikan error atau mengirim detail internal ke client.
- Jangan pernah commit secret. Ambil konfigurasi dari environment.
- Ubah kontrak API hanya jika PRD/SPEC dan seluruh pemanggil ikut diperbarui.
- Tambahkan atau perbarui test untuk perubahan perilaku; jalankan pemeriksaan paling relevan.
