# MoneyPlan — API Development Guide

## 1. Project Overview

REST API backend untuk **MoneyPlan** (repo/domain: FlowFund), aplikasi zero-based budgeting pribadi: pengguna memasukkan pemasukan bulanan, membaginya ke 3 pilar (Tabungan, Kebutuhan, Keinginan), memecahnya ke sub-alokasi, lalu mencatat transaksi sampai sisa Rp 0.

Frontend ada di repo terpisah `budgeting-app` (Nuxt). Kebutuhan produk ada di `budgeting-app/PRD_Financial_Budgeting_Monitoring.md`.

---

## 2. Tech Stack

| Layer | Package / Layanan |
|---|---|
| Bahasa | Go 1.25 |
| HTTP | `gin-gonic/gin` + `gin-contrib/cors` (`cors.Default()`, semua origin) |
| ORM | `gorm.io/gorm` + `gorm.io/driver/postgres` |
| Password | `golang.org/x/crypto/bcrypt` |
| Database | PostgreSQL di **Neon** (region Singapore, connection string pooler) |
| Hosting | **Vercel** (Go zero-config backend) — production: `https://budgeting-api-henna.vercel.app` |

---

## 3. Project Structure

```
main.go              # ConnectDB → DROP INDEX lama → AutoMigrate → routes.Setup() → Run(:PORT)
config/db.go         # ConnectDB: DATABASE_URL, fallback ke Postgres lokal
routes/routes.go     # Semua route terdaftar di Setup()
handlers/            # user.go, transaction.go, budget.go — validasi + query + response
models/              # User, Transaction, BudgetPlan, BudgetItem, BudgetCategory
```

Belum ada layer service/repository: handler memanggil `config.DB` langsung. Untuk fitur kecil ikuti pola itu; pisahkan ke layer baru hanya lewat plan yang disetujui.

---

## 4. Menjalankan & Deploy

```bash
DATABASE_URL='postgresql://...neon.tech/neondb?sslmode=require' go run .   # pakai DB Neon
go run .                                                                     # tanpa DATABASE_URL → Postgres lokal
go vet ./... && go build ./...                                               # cek sebelum push
```

| Env | Keterangan |
|---|---|
| `DATABASE_URL` | Connection string Neon. Jangan pernah di-commit atau ditempel di chat. |
| `LOCAL_DATABASE_URL` / `DATABASE_URL_LOCAL` | Override DSN lokal (default `host=localhost user=postgres dbname=budgeting_api`). |
| `PORT` | Diisi platform; default `8080`. |
| `GIN_MODE` | `release` di production. |

- **Deploy = push ke `main`.** Vercel mendeteksi `go.mod` + `main.go` dan menjalankannya sebagai server biasa.
- **Jangan tambahkan `vercel.json` rewrite atau folder `api/`.** Rewrite mengubah path request jadi `/api/index` dan semua route jadi 404.
- `AutoMigrate` jalan setiap cold start. Perubahan skema yang bukan sekadar menambah kolom (rename, ubah tipe, hapus) perlu langkah migrasi eksplisit, bukan mengandalkan AutoMigrate.

---

## 5. Coding Conventions

### General

- Jalankan `gofmt` dan `go vet ./...` sebelum commit.
- Route baru didaftarkan di `routes/routes.go`, handler di `handlers/<domain>.go`.
- Input request pakai struct `XxxInput` dengan tag `binding` dan `c.ShouldBindJSON`.
- Model hanya berisi field, tag, dan relasi — tanpa logika bisnis.
- Operasi tulis multi-langkah (hapus lalu buat ulang item, dsb.) dibungkus `config.DB.Transaction(...)`.

### Penamaan field JSON

Saat ini tidak konsisten dan frontend sudah bergantung pada bentuk yang ada:

| Model | Bentuk JSON | Contoh |
|---|---|---|
| `Transaction` | **Tanpa tag** → PascalCase | `ID`, `Type`, `Amount`, `Note`, `Date` |
| `User` | campuran | `ID`, `name`, `email` (password `json:"-"`) |
| `Budget*` | snake_case | `id`, `plan_id`, `income`, `nominal` |

Model baru pakai snake_case. Jangan ubah tag model lama tanpa mengubah frontend di PR yang sama.

### Komentar

**Default: jangan tulis komentar.** Kode harus menjelaskan dirinya sendiri lewat nama variabel, nama fungsi, dan struktur yang jelas — bukan lewat paragraf komentar di atasnya. Komentar yang hanya mengulang apa yang sudah terbaca dari kode membuat file penuh dan cepat basi saat kode berubah.

| Jangan | Sebabnya |
|---|---|
| Komentar yang menarasikan apa yang dilakukan baris di bawahnya (`// ambil budget` di atas `config.DB.Where(...).First(&plan)`) | Mengulang kode |
| Komentar pembuka di tiap fungsi yang hanya menyebut ulang namanya (`// GetBudget gets budget`) | Tidak menambah informasi |
| Komentar penanda seksi (`// ── Helpers ──`) | Kalau file butuh penanda seksi, file-nya yang terlalu besar |

**Satu-satunya pengecualian**: alasan (*why*) yang tidak bisa disimpulkan dari kode dan akan membuat orang berikutnya salah mengubahnya — keputusan yang disengaja, workaround, atau aturan bisnis yang tampak janggal. Tulis sebagai satu kalimat pendek, jelaskan alasannya, bukan mekanismenya.

```go
// BENAR — alasan yang tak terbaca dari kode
// Drop the old single-column unique index so GORM can recreate it as a composite (user_id, month) unique index
config.DB.Exec("DROP INDEX IF EXISTS idx_user_month")

// SALAH — menarasikan ulang kode
// Simpan plan ke database
config.DB.Save(&plan)
```

Aturan ini berlaku untuk kode baru maupun saat memodifikasi kode existing. Komentar existing yang sudah ada **jangan dihapus massal** tanpa diminta — hapus hanya yang berada di blok yang memang sedang diubah.

---

## 6. API Response Format

```go
c.JSON(http.StatusOK, gin.H{"message": "Item added successfully", "data": item})   // sukses
c.JSON(http.StatusOK, gin.H{"data": plan})                                          // sukses, baca
c.JSON(http.StatusBadRequest, gin.H{"error": "..."})                                // error
```

- Data selalu di key `data`, error di key `error`.
- `GET /budget` tanpa data mengembalikan **200** dengan `"data": null` (bukan 404) — frontend mengandalkan ini.
- Pesan error yang ditampilkan ke pengguna sebaiknya bahasa Indonesia. Jangan kirim `err.Error()` dari validator mentah (`Key: 'RegisterInput.Password' Error:...`) — ubah jadi kalimat yang bisa dibaca pengguna.

---

## 7. Auth & Data Ownership — Kondisi Saat Ini

> [!WARNING]
> **API belum punya autentikasi.** `POST /login` hanya mengembalikan data user tanpa token. Semua endpoint lain menerima `user_id` dari query/body dan mempercayainya begitu saja, jadi siapa pun bisa membaca, mengubah, atau menghapus data user lain. `GET /transactions` tanpa `user_id` mengembalikan transaksi **semua** user, dan `DELETE`/`PUT` by id tidak mengecek pemilik.

Saat menyentuh endpoint mana pun:

- Jangan menambah endpoint baru yang mengambil identitas user dari query/body.
- Perbaikan yang direncanakan: JWT saat login → middleware yang menaruh user id di context → handler membaca `c.GetUint("user_id")`, dan setiap query `WHERE id = ? AND user_id = ?`.

---

## 8. Domain & Known Issues

| Hal | Detail |
|---|---|
| Bulan | String `YYYY-MM`. Satu `BudgetPlan` per (`user_id`, `month`), unique index `idx_user_month`. |
| Uang | `int` dalam rupiah penuh, tanpa desimal. |
| `Transaction.Note` | Satu string berformat `"<sub-alokasi> - <deskripsi> - <bank> - <catatan>"`. Frontend mem-parse dengan `split(' - ')` dan menghitung realisasi sub-alokasi lewat pencocokan nama di `Note`. Tidak ada foreign key ke `BudgetItem`. |
| `BudgetCategory` | **Tidak punya kolom nominal & bank.** Frontend mengirim `allocatedAmount` dan `bankName`, tapi tidak tersimpan — nominal 3 pilar hilang setelah reload. |
| `SaveBudget` | Menghapus semua item & kategori lalu membuat ulang dengan ID baru (tanpa transaksi DB). |
| Hapus user | Belum ada endpoint. Akun tes `qa.flowfund.567106@example.com` (user id 2) masih ada di DB. |

---

## 9. Planning Mode Guidelines

When asked to create an implementation plan (Planning Mode), you MUST strictly adhere to the following rule:
**DO NOT write any code, execute modifying commands, or alter system state until the user has explicitly approved the implementation plan or given explicit instructions to start coding.**

Your sole responsibility during this phase is to research (read files, grep, etc.) and write the plan artifact.

### Klarifikasi Sebelum Menulis Plan

Sebelum menulis dokumen plan, jika ada hal yang belum jelas dan **jawabannya mengubah isi plan** — scope, pendekatan teknis, aturan bisnis, penamaan, atau trade-off — tanyakan lebih dulu:

- **Maksimal 4 pertanyaan**, diajukan sekaligus dalam satu kali tanya, bukan bertahap satu per satu.
- Hanya untuk hal yang benar-benar tidak bisa disimpulkan dari kode, konvensi repo, atau permintaan user. Jangan tanyakan hal yang sudah punya default jelas — ambil default-nya dan sebutkan di plan.
- Setiap pertanyaan sertakan opsi konkret beserta konsekuensinya, dan tandai mana yang direkomendasikan.
- Jika semuanya sudah jelas, langsung tulis plan tanpa bertanya.
- Pertanyaan yang muncul **setelah** plan ditulis dan tidak memblokir penulisan masuk ke komponen **Open Questions** (§10 komponen 13), bukan ditanyakan di muka.

## 10. Implementation Plan Structure

Every implementation plan MUST include the following components **in this order**. Skip a component only if it is genuinely not applicable to the task.

### Header Dokumen

Setiap dokumen plan dibuka dengan judul `# <Nama Fitur>` lalu **tabel metadata** — sebelum kotak "Ringkasan Singkat" dan sebelum komponen 1. Tabel ini yang menjawab: dokumen ini versi berapa, sudah disetujui atau belum, kapan terakhir disentuh, dan siapa yang menulisnya.

```markdown
# Nama Fitur

| | |
|---|---|
| **Versi** | 1.0 |
| **Status** | Draft |
| **Tanggal Dibuat** | 2026-10-03 |
| **Terakhir Diperbarui** | 2026-10-03 |
| **Author** | fadilnuris |
| **Reviewer** | — |
```

**Aturan pengisian:**

| Field | Aturan |
|---|---|
| **Versi** | `MAJOR.MINOR`. Mulai dari `1.0`. Naikkan MINOR untuk revisi isi (klarifikasi, tambah detail, perbaikan). Naikkan MAJOR bila pendekatan/scope berubah sehingga plan lama tidak lagi valid. |
| **Status** | Salah satu dari: `Draft` → `In Review` → `Approved` → `Implemented` → `Superseded`. Plan baru selalu `Draft`. |
| **Tanggal Dibuat** | Tanggal dokumen pertama kali ditulis, format `YYYY-MM-DD`. Tidak pernah berubah. |
| **Terakhir Diperbarui** | Tanggal revisi terakhir, format `YYYY-MM-DD`. Wajib diperbarui setiap kali isi dokumen diubah. |
| **Author** | Nama penulis plan (default: git user pada repo). |
| **Reviewer** | Nama yang me-review/approve. Isi `—` bila belum ada. |

> [!IMPORTANT]
> **Header wajib diperbarui setiap kali plan direvisi, bukan hanya saat dibuat.** Revisi yang tidak menaikkan versi dan tidak mengubah tanggal membuat pembaca tidak bisa membedakan plan yang sudah dikoreksi dari plan yang basi.

**Riwayat Revisi** *(opsional, mulai dipakai saat versi ≥ 1.1)* — tabel di bawah header:

```markdown
| Versi | Tanggal | Perubahan |
|---|---|---|
| 1.1 | 2026-10-04 | Tambah kolom bank per pilar |
| 1.0 | 2026-10-03 | Versi awal |
```

---

### Required Components

1. **Problem Statement** *(bahasa non-teknis — lihat catatan di bawah)* — Apa masalahnya, kenapa perlu diubah, dan dampaknya jika tidak diubah.
2. **Business Rules** *(bahasa non-teknis)* — Aturan bisnis yang harus dipenuhi oleh solusi — syarat & batasan dari sisi produk.
3. **Approach / Solution Overview** *(bahasa non-teknis)* — Pendekatan solusi yang dipilih, beserta perbandingan dengan alternatif lain (pros/cons table jika ada lebih dari 1 opsi).
4. **UI/UX Design** *(jika ada perubahan UI)* — Wajib menyertakan wireframe LoFi (low-fidelity, boleh berupa sketsa ASCII/box-layout sederhana, tidak perlu detail visual) untuk tiap state layar baru, plus mockup bila relevan, user flow, state & feedback (loading, success, error, empty state), dan responsive behavior.
5. **Perubahan UI Existing** *(jika ada perubahan UI)* — Wajib ada untuk setiap perubahan UI, berpasangan dengan komponen 4: before/after halaman atau komponen yang berubah, elemen baru yang ditambahkan beserta posisi dan interaksinya.
6. **Database / Data Design** — Struktur data baru dan modifikasi data existing — tabel, kolom, tipe data, relasi, index, constraint.
7. **Flow Diagram** *(bahasa non-teknis)* — Visualisasi alur proses menggunakan mermaid diagram:
   - **Write path** — kapan dan bagaimana data ditulis/diubah.
   - **Read path** — bagaimana data dibaca dan urutan prioritasnya.
8. **Event Summary Table** — Tabel ringkasan: per-event sistem, field apa yang berubah dan apa yang tidak.
9. **Scenario Walkthrough** — Contoh skenario konkret step-by-step — happy path, edge case, dan error case.
10. **Impacted Files / Components** — Daftar file atau komponen yang perlu dibuat (`[NEW]`) atau dimodifikasi (`[MODIFY]`) atau dihapus (`[DELETE]`), dikelompokkan per layer (database, model, handler, route, UI, dll).
11. **Data Migration / Backfill Strategy** — Bagaimana menangani data existing yang sudah ada sebelum fitur ini diimplementasi.
12. **Decisions Requiring Review** — Keputusan desain yang memerlukan persetujuan — naming, scope, UX, trade-off. Gunakan alert `[!IMPORTANT]` atau `[!WARNING]`.
13. **Open Questions** — Pertanyaan yang belum terjawab dan bisa mempengaruhi implementasi.
14. **Verification Plan** — Rencana pengujian:
    - **Automated tests** — daftar test case spesifik.
    - **Manual verification** — langkah validasi manual.

### Riwayat Prompt

Paling bawah dokumen, setelah komponen 14. Satu entri per prompt pemicu plan/revisi, verbatim (jangan diringkas/parafrase), urut terbaru→terlama:

```markdown
---

## Riwayat Prompt

### v1.1 — 2026-10-04
> Bank per pilar juga perlu disimpan ya

### v1.0 — 2026-10-03
> Buatkan plan supaya nominal 3 pilar tersimpan ke database
```

### Ordering Logic

**WHY** (1–2) → **WHAT** user lihat (3–5) → **HOW** secara teknis (6–9) → **IMPACT** (10–11) → **UNRESOLVED** (12–13) → **VERIFY** (14).

### Presentation Rules

Daftar komponen di atas menentukan *apa* isinya; aturan berikut menentukan *bagaimana* menyajikannya agar dokumen bisa dipindai cepat.

- **Kotak "Ringkasan Singkat" di paling atas** — tepat di bawah header dokumen dan sebelum komponen 1, berupa blockquote berisi 3–5 poin: apa yang dibuat, pemicu/kondisi utamanya, aksi yang dilakukan, dan yang sengaja tidak dilakukan. Pembaca harus menangkap inti fitur tanpa membaca seluruh dokumen. Tulis dengan bahasa non-teknis (lihat aturan di bawah).
- **Bahasa non-teknis untuk Ringkasan Singkat, Problem Statement, Business Rules, Approach/Solution Overview, dan Flow Diagram** — tulis dari sudut pandang dampak ke pengguna, pakai istilah sehari-hari, bukan istilah kode. Kalau istilah teknis memang perlu disebut supaya bisa dilacak ke implementasi (nama tabel, kolom, handler, endpoint, nama variabel di diagram), taruh dalam kurung setelah penjelasan non-teknisnya, jangan jadi kalimat utama.

  ```markdown
  Salah: `allocated_amount` ditambahkan ke `budget_categories` dan diisi di `SaveBudget`.
  Benar: Nominal tiap pilar sekarang ikut tersimpan, jadi tidak hilang saat halaman dibuka lagi (kolom baru `budget_categories.allocated_amount`, diisi oleh handler `SaveBudget`).
  ```

  Komponen lain (UI/UX Design, Database/Data Design, Event Summary Table, Impacted Files, dll) boleh sepenuhnya teknis karena pembacanya adalah engineer.
- **Tabel lebih baik daripada prosa** — daftar yang punya dimensi berulang (syarat, skenario, file terdampak, test case, temuan) ditulis sebagai tabel, bukan bullet panjang. Bullet hanya untuk daftar pendek satu dimensi.
- **Paragraf pendek** — maksimal 3–4 kalimat. Paragraf padat yang penuh nama kolom/struct dipecah menjadi tabel.
- **Pemisah antar-seksi** — gunakan `---` di antara komponen utama.
- **Alert diawali kalimat tebal** — tiap blok `[!IMPORTANT]` / `[!WARNING]` dibuka satu kalimat tebal yang merangkum isinya, supaya bisa dipindai tanpa membaca seluruh paragraf. Urutkan dari risiko terbesar.
- **Mermaid harus benar-benar render** — pakai label polos: tanpa tanda baca (`:` `=` `+` `.`), tanpa `<br/>`, tanpa `\n`, dan hindari label edge yang diawali `--`. Pecah kalimat panjang menjadi beberapa node, jangan dijejalkan ke satu label.
- **Rujukan antar-seksi wajib valid** — sebelum dokumen dianggap selesai, pastikan setiap rujukan `§n` menunjuk ke seksi yang benar dan tidak ada rujukan ke seksi yang tidak pernah dibuat.
- **Context memuat temuan, bukan hanya masalah** — bila hasil penelusuran kode mengubah bentuk solusi (mis. kolom yang ternyata tidak pernah dipakai kode manapun), sajikan sebagai tabel "Temuan → Implikasi", bukan dikubur di dalam paragraf.

### Lokasi & Penamaan File

Simpan dokumen plan di `docs/plans/<nama-fitur-kebab-case>.md` (contoh: `docs/plans/simpan-nominal-pilar.md`), kecuali user meminta lokasi lain. Plan yang menyentuh API dan frontend sekaligus disimpan di repo yang perubahannya paling besar, dan repo lainnya cukup merujuk ke sana.
