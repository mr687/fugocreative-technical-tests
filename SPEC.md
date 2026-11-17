# Technical Specifications

## 1. Database Schema

```sql
CREATE TABLE briefs (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    contact_name text NOT NULL CHECK (length(trim(contact_name)) > 0),
    email text NOT NULL,
    company_name text NOT NULL CHECK (length(trim(company_name)) > 0),
    need_description text NOT NULL CHECK (length(trim(need_description)) > 0),
    budget_range text NOT NULL,
    target_date date NOT NULL CHECK (target_date >= CURRENT_DATE),
    ai_status text NOT NULL DEFAULT 'pending'
        CHECK (ai_status IN ('pending', 'succeeded', 'failed')),
    ai_result jsonb,
    ai_error text,
    pipeline_status text NOT NULL DEFAULT 'New'
        CHECK (pipeline_status IN ('New', 'Contacted', 'Proposal', 'Won', 'Lost')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((ai_status = 'succeeded') = (ai_result IS NOT NULL))
);
CREATE INDEX briefs_created_at_idx ON briefs (created_at DESC);

CREATE TABLE ai_jobs (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    brief_id uuid NOT NULL REFERENCES briefs(id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'queued'
        CHECK (status IN ('queued', 'processing', 'succeeded', 'failed')),
    attempt integer NOT NULL DEFAULT 1 CHECK (attempt > 0),
    error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    finished_at timestamptz
);
CREATE INDEX ai_jobs_queued_idx ON ai_jobs (created_at) WHERE status = 'queued';
CREATE UNIQUE INDEX ai_jobs_one_active_per_brief_idx ON ai_jobs (brief_id)
    WHERE status IN ('queued', 'processing');
```

Pembuatan brief dan job pertama wajib satu transaksi. Worker claim job secara atomik dengan row locking (`FOR UPDATE SKIP LOCKED`). UUIDv7 memerlukan PostgreSQL 18.

## 2. Integrasi LLM

- Provider, model, API key, dan timeout dikonfigurasi melalui environment; key hanya di backend.
- Worker goroutine memanggil provider dengan timeout. Brief diperlakukan sebagai data tak tepercaya, bukan instruksi.
- Minta output JSON sesuai schema di bawah. Parse dan validasi seluruh required fields, tipe, enum, dan larang properti tambahan sebelum menyimpan `ai_result`.
- Sukses: `briefs.ai_status=succeeded`, `ai_result` terisi; job `succeeded`.
- Gagal, timeout, atau output invalid: brief/job `failed`, simpan pesan error internal yang aman; brief tetap tersedia.
- Retry admin membuat job baru hanya bila tidak ada job aktif. Job terminal dipertahankan.

### JSON Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "required": ["summary", "service_category", "priority", "priority_reason", "missing_info"],
  "properties": {
    "summary": { "type": "string", "minLength": 1 },
    "service_category": {
      "type": "string",
      "enum": ["Branding", "Website", "Mobile App", "AI Automation", "Social Media", "Lainnya"]
    },
    "priority": { "type": "string", "enum": ["High", "Medium", "Low"] },
    "priority_reason": { "type": "string", "minLength": 1 },
    "missing_info": { "type": "array", "items": { "type": "string" } }
  }
}
```

## 3. API Contracts

Semua response JSON. Error umum: `{"error":{"code":"...","message":"..."}}`. Admin endpoint memakai `Authorization: Bearer <JWT>`. Field tanggal format `YYYY-MM-DD`.

### `POST /api/public/briefs` — publik

Request:
```json
{
  "contact_name": "Ayu Putri",
  "email": "ayu@example.com",
  "company_name": "Contoh Studio",
  "need_description": "Membutuhkan website company profile",
  "budget_range": "Rp10–25 juta",
  "target_date": "2026-12-20"
}
```
Success `202`:
```json
{"id":"019...","ai_status":"pending","job_status":"queued","message":"Brief berhasil dikirim"}
```
Error: `400 VALIDATION_ERROR`, `500 INTERNAL_ERROR`.

### `POST /api/auth/login` — publik

Request:
```json
{"username":"admin","password":"..."}
```
Success `200`:
```json
{"access_token":"<JWT>","token_type":"Bearer","expires_in":3600}
```
Error: `400 VALIDATION_ERROR`, `401 INVALID_CREDENTIALS`, `500 INTERNAL_ERROR`.

### `GET /api/auth/me` — JWT

Request body: tidak ada.

Success `200`:
```json
{"authenticated":true,"username":"admin"}
```
Error: `401 UNAUTHORIZED`.

### `GET /api/admin/briefs?priority=&category=` — JWT

Request body: tidak ada. Query filter opsional; kategori/priority harus enum valid.

Success `200`:
```json
{"items":[{"id":"019...","company_name":"Contoh Studio","service_category":"Website","priority":"Medium","ai_status":"succeeded","pipeline_status":"New","created_at":"2026-11-17T08:00:00Z"}]}
```
Error: `400 INVALID_FILTER`, `401 UNAUTHORIZED`, `500 INTERNAL_ERROR`.

### `GET /api/admin/briefs/{id}` — JWT

Request body: tidak ada.

Success `200`:
```json
{
  "id":"019...",
  "contact_name":"Ayu Putri",
  "email":"ayu@example.com",
  "company_name":"Contoh Studio",
  "need_description":"Membutuhkan website company profile",
  "budget_range":"Rp10–25 juta",
  "target_date":"2026-12-20",
  "ai_status":"succeeded",
  "ai_result":{"summary":"Brief meminta website company profile.","service_category":"Website","priority":"Medium","priority_reason":"...","missing_info":[]},
  "pipeline_status":"New",
  "latest_job":{"id":"019...","status":"succeeded","attempt":1,"error":null},
  "created_at":"2026-11-17T08:00:00Z",
  "updated_at":"2026-11-17T08:00:00Z"
}
```
Error: `401 UNAUTHORIZED`, `404 NOT_FOUND`, `500 INTERNAL_ERROR`.

### `POST /api/admin/briefs/{id}/reprocess` — JWT

Request body: `{}`.

Success `202`:
```json
{"job_id":"019...","ai_status":"pending","job_status":"queued"}
```
Error: `401 UNAUTHORIZED`, `404 NOT_FOUND`, `409 JOB_ALREADY_ACTIVE`, `500 INTERNAL_ERROR`.

### `PATCH /api/admin/briefs/{id}/pipeline` — JWT

Request:
```json
{"pipeline_status":"Contacted"}
```
Success `200`:
```json
{"id":"019...","pipeline_status":"Contacted","updated_at":"2026-11-17T08:10:00Z"}
```
Error: `400 VALIDATION_ERROR`, `401 UNAUTHORIZED`, `404 NOT_FOUND`, `500 INTERNAL_ERROR`.

## 4. Error Handling

| Kondisi | HTTP | Perilaku |
|---|---:|---|
| JSON rusak, field wajib/enum/tanggal invalid | 400 | `VALIDATION_ERROR`; jangan menyimpan data |
| Kredensial login salah | 401 | `INVALID_CREDENTIALS`; pesan generik |
| JWT hilang, invalid, atau expired | 401 | `UNAUTHORIZED`; jangan ungkap detail verifikasi |
| Brief/job tidak ditemukan | 404 | `NOT_FOUND` |
| Job aktif sudah ada saat retry | 409 | `JOB_ALREADY_ACTIVE` |
| Database gagal sebelum commit submit | 500 | Jangan mengklaim brief tersimpan |
| Provider gagal, timeout, atau output invalid | Tidak memengaruhi response submit | Tandai job dan brief `failed`; simpan error aman; brief tetap tersimpan |
| Worker crash/restart | — | Job `queued` tetap durable; pulihkan `processing` stale agar bisa di-claim ulang |
| Error tak terduga | 500 | `INTERNAL_ERROR`; detail hanya di log server tanpa secret/credential |

Format error konsisten: `{"error":{"code":"VALIDATION_ERROR","message":"Periksa kembali input."}}`. Jangan kirim stack trace, prompt internal, token, API key, atau error mentah provider ke client.
