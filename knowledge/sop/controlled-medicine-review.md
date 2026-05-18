---
title: SOP Review Transaksi Obat Keras dan Pola Pembelian Mencurigakan
category: controlled-review
doc_type: sop
version: 1.0
owner: pharmacy-risk
---

# Tujuan
Memberikan panduan bagi staff, supervisor, dan tim audit untuk menandai transaksi yang berisiko tinggi, khususnya pada obat keras atau pola pembelian yang tidak wajar.

# Ruang Lingkup
SOP ini berlaku untuk item yang memerlukan perhatian lebih tinggi, termasuk item internal `OBATKERAS-X`, transaksi berulang, dan kombinasi item yang berpotensi berisiko.

# Prinsip Umum
Tidak semua transaksi harus ditolak, tetapi transaksi dengan red flag harus masuk jalur review sebelum dianggap aman untuk diproses penuh.

# Red Flags Utama
Transaksi harus masuk review bila ditemukan salah satu kondisi berikut:
- kuantitas pembelian jauh di atas pola penggunaan normal
- pembelian berulang dalam interval waktu singkat oleh identitas yang sama
- pelanggan menolak memberikan informasi dasar yang dibutuhkan untuk verifikasi
- item obat keras dibeli tanpa dokumen pendukung yang sesuai
- kombinasi item dalam satu transaksi terlihat tidak konsisten dengan indikasi penggunaan umum

# Tindakan Staff
Saat menemukan red flag, staff wajib:
- tidak melanjutkan transaksi secara otomatis
- menandai transaksi untuk review supervisor
- memastikan data item, jumlah, dan jejak request tercatat dengan benar
- tidak memberikan opini klinis di luar kewenangannya

# Tindakan Supervisor
Supervisor melakukan review dengan mempertimbangkan:
- histori transaksi bila tersedia
- kewajaran jumlah
- konsistensi item yang dibeli
- kebutuhan eskalasi ke apoteker atau review manual tambahan

# Hubungan dengan Sistem Audit
Dalam sistem FinPharm-AI, transaksi yang tidak aman untuk diproses otomatis dapat:
- masuk status pending review
- masuk status flagged
- atau ditolak bila ada masalah operasional atau verifikasi gagal

# Dokumentasi
Setiap transaksi yang masuk jalur review harus memiliki:
- alasan review yang jelas
- jejak request atau trace id
- item dan kuantitas lengkap
- hasil keputusan akhir supervisor atau sistem audit