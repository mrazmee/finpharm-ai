# Day 33 — RabbitMQ via Docker Compose

## Tujuan

- Menyiapkan RabbitMQ sebagai message broker untuk phase event-driven
- Menjalankan RabbitMQ via Docker Compose
- Menyediakan management UI agar mudah dicek saat development

---

# Scope Day 33

Hari ini belum ada perubahan flow bisnis aplikasi.

Fokusnya hanya:
- RabbitMQ hidup via Docker Compose
- port AMQP siap dipakai
- management UI siap dipakai
- script start / stop / logs tersedia
- dokumentasi Day 33 tersedia

## Kenapa Day 33 Hanya Infra?

Karena sebelum `Transaction Service` publish event, kita harus pastikan dulu broker-nya:
- bisa jalan stabil
- bisa diakses lokal
- mudah dicek
- mudah di-reset

Kalau broker belum sehat, maka Day 34 (publish event) akan lebih sulit dibedakan apakah itu bug publisher, bug koneksi, atau broker belum benar-benar hidup.

---

# Konfigurasi RabbitMQ

## Image Yang Dipakai

Kita pakai: `rabbitmq:3.13-management`

**Kenapa `management`?**
Karena untuk belajar dan portfolio, management UI sangat membantu untuk:
- cek broker hidup
- cek queue
- cek exchange
- cek connection
- cek publish/consume nanti

## Port Yang Dipakai

- `5672` → AMQP port untuk aplikasi/service
- `15672` → management UI

## Credential Default Lokal

Untuk development lokal:
- username: `finpharm`
- password: `finpharm`

*Catatan: Ini hanya untuk local development, bukan production.*

---

# File Yang Ditambahkan

```text
[ADD] docker-compose.rabbitmq.yml
[ADD] scripts/rabbitmq-up.ps1
[ADD] scripts/rabbitmq-down.ps1
[ADD] scripts/rabbitmq-logs.ps1
[ADD] docs/day33.md
```

## Kenapa Compose Dibuat File Terpisah?

Supaya aman:
- tidak mengganggu compose file lain yang mungkin sudah ada
- mudah dijalankan khusus untuk RabbitMQ
- gampang dibaca reviewer
- gampang dihapus / di-reset

---

# Cara Menjalankan

## 1. Start RabbitMQ

```powershell
.\scripts\rabbitmq-up.ps1
```

## 2. Lihat Logs

```powershell
.\scripts\rabbitmq-logs.ps1
```

## 3. Stop RabbitMQ

```powershell
.\scripts\rabbitmq-down.ps1
```

---

# Cara Verifikasi

## 1. Cek Container Hidup

```bash
docker compose -f .\docker-compose.rabbitmq.yml ps
```

Expected:
- service rabbitmq status `running` / `healthy`

## 2. Cek Management UI

Buka browser: `http://localhost:15672`

Login dengan:
- username: `finpharm`
- password: `finpharm`

Expected:
- halaman RabbitMQ Management bisa diakses

## 3. Cek Port AMQP

Port yang nanti dipakai service: `localhost:5672`

---

# Apa Yang Belum Dilakukan Hari Ini?

Day 33 belum mencakup:
- publisher di Transaction Service
- consumer / worker service
- exchange / queue declaration di aplikasi
- retry / dead-letter
- event contract

Itu baru masuk hari-hari berikutnya.

---

# Self-Review

- Kenapa message broker disiapkan dulu sebelum publish event?
- Kenapa pakai image `management`, bukan image RabbitMQ biasa?
- Kenapa credential lokal dipisahkan dari concern production?
- Kenapa compose RabbitMQ dibuat terpisah, bukan langsung dicampur ke file compose lain?

---

# Verifikasi Akhir Day 33

Jalankan ini:

```powershell
.\scripts\rabbitmq-up.ps1
```

Lalu cek di browser:
- `http://localhost:15672`
- Login: `finpharm` / `finpharm`

**Expected Hasil Day 33:**
- RabbitMQ container hidup
- UI terbuka
- belum ada perubahan ke logic aplikasi
- siap lanjut Day 34 untuk publish event