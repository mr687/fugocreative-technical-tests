# PRD — Fugo Creative Brief Intake

## 1. Ringkasan

MVP internal untuk menerima brief calon klien, menyimpan data, menghasilkan ringkasan/kategori/prioritas dengan AI, dan membantu sales mengelola tindak lanjut melalui dashboard terlindungi.

**Sasaran:** reviewer dapat mengirim brief publik, melihatnya tersimpan beserta hasil AI, masuk dashboard admin, memfilter dan membuka detail, mencoba ulang pemrosesan AI, serta memperbarui pipeline.

**Stack yang ditetapkan:** monorepo dengan `./backend` (Go 1.27, PostgreSQL 18) dan `./frontend` (Next.js, Bun preferred, shadcn/ui). PostgreSQL 18 uuidv7 digunakan untuk ID bila versi/runtime driver mendukungnya dengan baik.

**Prioritas:** seluruh fitur wajib berfungsi end-to-end sebelum penambahan opsional. Estimasi test 4–6 jam; hindari abstraksi dan fitur di luar lingkup.

## 2. Masalah dan nilai

Brief masuk melalui email/WhatsApp tersebar dan harus dibaca satu per satu. Sales memerlukan satu daftar terpusat dengan ringkasan dan prioritas awal agar dapat menentukan tindak lanjut lebih cepat.

## 3. Tujuan dan bukan tujuan

### Tujuan
- Form publik dengan validasi browser dan validasi server.
- Penyimpanan brief tetap berhasil meski AI gagal.
- Hasil AI terstruktur, tervalidasi, tersimpan, dan dapat dicoba ulang.
- Dashboard admin satu akun, terlindungi autentikasi.
- Daftar, filter, detail, dan perubahan status pipeline.
- Deployment publik dan README yang memungkinkan reviewer mencoba aplikasi.

### Bukan tujuan MVP
- Integrasi inbox email/WhatsApp, multi-user/role, queue eksternal, notifikasi, analytics, ekspor, draft email AI.
- CRM penuh, pengeditan/penghapusan brief, atau workflow pipeline kustom.
- Jaminan AI selalu benar; hasil AI adalah rekomendasi untuk ditinjau manusia.

## 4. Pengguna dan akses

| Pengguna | Akses | Kebutuhan |
|---|---|---|
| Calon klien | Form publik | Mengirim brief dan mendapat konfirmasi |
| Admin/sales | Login dashboard | Menilai lead dan mengelola pipeline |

Satu akun admin dikonfigurasi melalui environment variables (username dan password atau hash password). Backend mengembalikan JWT setelah login; token tidak memiliki refresh token. Jangan menyimpan kredensial plaintext di repository atau mengirimkannya ke frontend bundle.

## 5. Alur utama

1. Klien mengisi dan mengirim form.
2. Backend memvalidasi input, lalu dalam satu transaksi menyimpan brief berstatus AI `pending` dan job AI berstatus `queued` di PostgreSQL.
3. Setelah transaksi commit, backend merespons bahwa brief diterima. Worker goroutine mengambil job dari PostgreSQL dan memproses AI di background.
4. Bila sukses, worker menyimpan hasil terstruktur dan status `succeeded`; bila gagal/timeout/JSON tidak valid, brief tetap tersimpan dengan status `failed`.
5. Klien menerima konfirmasi bahwa brief diterima; jangan tampilkan error provider atau data internal.
6. Admin login, melihat daftar terbaru, memfilter kategori/prioritas, dan membuka detail.
7. Admin dapat mengubah pipeline dan meminta pemrosesan ulang AI pada brief.

**Trade-off MVP:** worker goroutine sederhana menghindari infrastruktur queue eksternal. PostgreSQL menjadi sumber job yang durable; pada startup worker mengambil kembali job `queued` dan job `processing` yang ditinggalkan proses sebelumnya. Worker harus membatasi concurrency dan memakai timeout.

## 6. Kebutuhan fungsional

### 6.1 Form publik
Field wajib:
- Nama kontak (teks).
- Email (format email).
- Nama perusahaan (teks).
- Jenis kebutuhan (teks bebas).
- Budget (salah satu rentang tetap yang ditampilkan secara konsisten; rentang final dicantumkan pada UI dan README).
- Target timeline (tanggal kalender).

Aturan:
- Field wajib tidak boleh kosong setelah trim; email valid; tanggal tidak boleh sebelum tanggal hari ini.
- Validasi di frontend untuk UX dan di backend sebagai sumber kebenaran.
- Saat submit, tampilkan keadaan loading, cegah submit ganda, tampilkan error yang dapat ditindaklanjuti, dan tampilkan pesan sukses setelah brief tersimpan.
- Kegagalan AI tidak boleh membuat submit tampak gagal jika brief telah tersimpan.

### 6.2 AI processing
Setiap brief baru diproses oleh provider/model yang ditetapkan via konfigurasi environment. Prompt meminta JSON saja dengan schema:

```json
{
  "summary": "ringkasan 1–2 kalimat",
  "service_category": "Branding | Website | Mobile App | AI Automation | Social Media | Lainnya",
  "priority": "High | Medium | Low",
  "priority_reason": "alasan singkat",
  "missing_info": ["informasi yang masih kurang dari brief"]
}
```

Kriteria:
- Backend memvalidasi struktur, enum, dan tipe field sebelum menyimpan hasil sebagai JSONB.
- Jangan mempercayai JSON/model output sebagai instruksi sistem; brief adalah data tak tepercaya.
- Worker goroutine mengambil job `queued` dari PostgreSQL secara aman agar beberapa worker tidak memproses job sama; gunakan transaksi/locking dan klaim job atomik.
- Sukses/gagal memperbarui record `ai_jobs` dan status/hasil AI pada `briefs` secara konsisten. Simpan error internal yang aman tanpa API key/secret.
- Retry hanya melalui aksi admin yang membuat job baru atau mengantrekan ulang secara atomik; cegah dua job aktif untuk brief yang sama.
- Tampilkan status pemrosesan dari job/brief (`queued`/`processing`/`succeeded`/`failed`); bila gagal, tampilkan aksi retry.
- README mencantumkan provider/model aktual, env, dan bila mock dipakai menyebutkan keterbatasan secara jujur.

### 6.3 Dashboard admin
- Semua route halaman dan API internal wajib menolak akses tanpa JWT valid.
- Login admin menggunakan kredensial static dari environment variables; backend mengembalikan access token JWT berumur terbatas, tanpa refresh token. Logout dilakukan di sisi klien dengan menghapus token.
- Daftar menampilkan perusahaan, kategori (belum diproses: nilai fallback seperti “Belum dikategorikan”), prioritas, status AI, tanggal masuk, dan status pipeline bila ruang memungkinkan.
- Urutan terbaru dahulu.
- Filter kategori dan prioritas; filter dapat dikombinasikan. URL/query state boleh digunakan bila implementasinya sederhana.
- Detail menampilkan seluruh data form, hasil AI, status pemrosesan dan pipeline, serta retry AI.
- Pipeline: `New → Contacted → Proposal → Won` atau `Lost`. Validasi enum di backend; MVP dapat mengizinkan transisi langsung antar-status agar sales tidak terhambat.
- Status sukses/gagal terlihat jelas; aksi mutasi memberi feedback.

### 6.4 Deployment
- Frontend dan backend dapat diakses melalui URL publik; PostgreSQL terkelola atau host deployment.
- Konfigurasi CORS dibatasi ke origin frontend. HTTPS di production.
- Secrets hanya lewat environment/secret manager; tidak masuk Git, log, atau client bundle.
- README memuat langkah lokal, env, migrasi, arsitektur, skema, keputusan/trade-off, provider/model, dan AI usage log.

## 7. Model data minimum

### `briefs`
- `id`: UUIDv7 primary key.
- `contact_name`, `email`, `company_name`, `need_description`: teks wajib.
- `budget_range`: teks/enum rentang yang didukung.
- `target_date`: `DATE`, tidak kurang dari tanggal hari ini saat insert.
- `ai_status`: enum/constraint `pending | succeeded | failed`.
- `ai_result`: JSONB nullable; hanya diisi setelah validasi.
- `ai_error`: teks nullable, pesan aman/non-rahasia.
- `pipeline_status`: enum/constraint `New | Contacted | Proposal | Won | Lost`, default `New`.
- `created_at`, `updated_at`: timestamp dengan zona waktu.

Index minimum: `created_at DESC`; index pada kategori/prioritas hasil AI hanya bila model/query memerlukannya. Filter JSONB dapat dilakukan langsung untuk volume MVP.

### `ai_jobs`
- `id`: UUIDv7 primary key.
- `brief_id`: UUID foreign key ke `briefs`.
- `status`: `queued | processing | succeeded | failed`.
- `attempt`: jumlah percobaan.
- `error`: pesan internal aman nullable.
- `created_at`, `started_at`, `finished_at`: timestamp dengan zona waktu.

Tambahkan index untuk mengambil job `queued` secara efisien dan constraint/locking untuk mencegah lebih dari satu job aktif (`queued`/`processing`) per brief. Penyimpanan brief dan job awal harus atomik. Job terminal tetap disimpan untuk state tracking; retry membuat job baru.

Admin tidak disimpan di database; kredensial static berasal dari environment variables. Jangan simpan output mentah provider kecuali dibutuhkan untuk diagnosis dan dipastikan aman.

## 8. API minimum

Prefix `/api` (kontrak JSON konsisten, tanggal ISO-8601/date, error `{ "error": "..." }` tanpa detail sensitif).

| Method & path | Auth | Perilaku |
|---|---|---|
| `POST /api/public/briefs` | Publik | Validasi dan dalam transaksi simpan brief + job `queued`; respons segera mengindikasikan brief diterima |
| `POST /api/auth/login` | Publik | Verifikasi kredensial env dan kembalikan JWT |
| `GET /api/auth/me` | JWT | Periksa klaim/validitas token |
| `GET /api/admin/briefs?priority=&category=` | JWT | Daftar terbaru dan filter |
| `GET /api/admin/briefs/{id}` | JWT | Detail brief dan status job |
| `POST /api/admin/briefs/{id}/reprocess` | JWT | Buat job retry jika tidak ada job aktif |
| `PATCH /api/admin/briefs/{id}/pipeline` | JWT | Perbarui status pipeline |

Kode status harus membedakan input invalid, tidak terautentikasi/token tidak valid, tidak ditemukan, konflik bila job aktif sudah ada, dan kegagalan server. JWT ditransmisikan melalui `Authorization: Bearer`; autentikasi ditegakkan di backend.

## 9. UI dan navigasi

Frontend Next.js dengan shadcn/ui sebagai komponen dasar; desain sederhana, responsif, aksesibel melalui label form, fokus keyboard, dan pesan error yang terhubung ke field.

- `/` atau `/brief`: landing/form publik.
- `/login`: login admin.
- `/dashboard`: tabel/list brief dengan filter.
- `/dashboard/briefs/[id]`: detail, hasil AI, retry, dan pipeline.

Jangan menambahkan dashboard analytics atau visualisasi sebelum alur wajib stabil.

## 10. Keamanan dan keandalan minimum

- Validasi server untuk semua input dan batas ukuran teks yang wajar.
- Query database terparameterisasi; jangan membangun SQL dari input filter mentah.
- Verifikasi JWT, expiry, issuer/audience bila digunakan, dan algoritma allowlist; JWT secret/private key hanya di backend env. JWT disimpan dan dikirim frontend secara aman; jangan menaruhnya pada URL.
- CORS allowlist, HTTPS production, timeout provider dan database.
- Jangan mengembalikan stack trace, prompt rahasia, credential, atau error provider mentah ke klien.
- Batasi satu job aktif per brief; worker melakukan claim atomik dan menangani job tertinggal saat restart.
- Log request/error tanpa secret dan tanpa menyalin seluruh data pribadi brief bila tidak perlu.

## 11. Acceptance criteria MVP

1. Brief valid tersimpan bersama job AI `queued` secara atomik dan form segera menunjukkan konfirmasi tanpa menunggu LLM.
2. Input kosong, email invalid, dan tanggal lampau ditolak oleh backend; frontend memberi feedback.
3. Brief tetap ada bila provider gagal, timeout, atau output AI invalid; job/brief berstatus `failed`.
4. Output valid mematuhi enum/schema, disimpan, dan tampil pada detail; UI mencerminkan status background job.
5. Retry admin membuat job yang dapat diproses worker dan memperbarui status/hasil.
6. Pengunjung tanpa JWT valid tidak dapat membaca/mengubah dashboard atau memanggil endpoint admin.
7. Admin dapat melihat daftar terbaru, memfilter prioritas/kategori, membuka detail, dan mengubah pipeline.
8. Semua status pipeline yang diizinkan tersimpan dan terlihat setelah reload.
9. Aplikasi dapat dijalankan lokal mengikuti README dan tersedia di URL publik.
10. Tidak ada API key/password/JWT secret dalam repository atau frontend bundle.

## 12. Prioritas implementasi

### P0 — wajib, urutan pengerjaan
1. Bootstrap monorepo, konfigurasi env, koneksi PostgreSQL dan migrasi tabel `briefs` serta `ai_jobs`.
2. Backend: model, validasi input, endpoint submit dan transaksi simpan brief + job.
3. Backend: worker goroutine, klaim job atomik, timeout, pemrosesan AI, validasi schema, status sukses/gagal dan pemulihan job tertinggal.
4. Backend: login kredensial static dari env, penerbitan/verifikasi JWT, middleware perlindungan endpoint.
5. Backend: list/filter/detail, retry job, update pipeline.
6. Frontend: form submit/success/error; konfirmasi tidak menunggu AI.
7. Frontend: login dan pemakaian JWT, dashboard/filter, detail/status job/retry/pipeline.
8. Deploy database, backend, frontend; pastikan worker berjalan pada deployment backend dan smoke test end-to-end.
9. README dan AI usage log faktual.

### P1 — hanya setelah P0
- Satu test kecil untuk validasi schema AI dan satu test endpoint utama.
- Docker Compose untuk menjalankan app/database lokal.

### P2 — tidak dikerjakan kecuali semua di atas selesai
- Draft email AI, queue eksternal, multi-admin, analytics, integrasi kanal.

## 13. Keputusan teknis dan trade-off

- Monorepo dua aplikasi: `./backend`, `./frontend`; kontrak API HTTP/JSON sederhana.
- Go 1.27 sesuai permintaan; verifikasi ketersediaan toolchain/runtime pada waktu implementasi, jangan mengasumsikan versi tersedia. Jika belum tersedia, laporkan dan gunakan versi Go stabil yang tersedia hanya setelah keputusan eksplisit proyek.
- PostgreSQL 18; gunakan native UUIDv7 bila tersedia melalui fungsi database/toolchain dan driver. Jika tidak, gunakan generator UUIDv7 yang kompatibel tanpa mengubah kontrak ID.
- Bun untuk dependency/install frontend bila kompatibel dengan toolchain Next.js dan deployment; tetap gunakan lockfile.
- Worker goroutine memproses job durable di PostgreSQL; tidak ada broker eksternal di MVP.
- Kredensial admin static dari env; JWT access token berumur terbatas tanpa refresh token.
- JSONB menyimpan hasil AI sebagai satu unit terstruktur; validasi aplikasi tetap wajib.

## 14. Risiko dan mitigasi

| Risiko | Mitigasi MVP |
|---|---|
| Provider lambat/gagal | Worker async, timeout; brief/job durable, status failed dan retry |
| Output AI melenceng dari schema | Parsing + validasi tipe/enum; jangan simpan sebagai sukses |
| Worker crash/restart | Claim atomik dan pemulihan job `processing` yang stale saat startup |
| JWT secret/config salah | Validasi env saat startup; token expiry dan verifikasi algoritma |
| Toolchain Go 1.27/Postgres 18 belum tersedia | Verifikasi saat setup; gunakan versi kompatibel dan nyatakan deviasi bila perlu |
| Deploy menghabiskan waktu | Deploy stack minimal sedini mungkin setelah alur vertikal dasar |
| Brief berisi data sensitif | Batasi akses, jangan log payload lengkap, gunakan TLS dan secret env |
| Deploy menghabiskan waktu | Deploy stack minimal sedini mungkin setelah alur vertikal dasar |
| Brief berisi data sensitif | Batasi akses, jangan log payload lengkap, gunakan TLS dan secret env |

## 15. Definition of Done

- Acceptance criteria P0 terpenuhi di deployment publik.
- Migrasi database berjalan dari instruksi README.
- Endpoint admin menolak token tanpa/invalid/expired dan menerima JWT valid.
- Kegagalan AI tidak menghilangkan brief; worker memproses job durable dan retry berfungsi.
- Tidak ada secret ter-commit.
- README menjelaskan arsitektur, skema, env, cara lokal, trade-off, provider/model dan AI usage log; dua contoh koreksi AI ditulis berdasarkan kejadian nyata.
