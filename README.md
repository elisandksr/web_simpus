# SIMPUS (Sistem Informasi Manajemen Perpustakaan)

Aplikasi ini adalah backend untuk sistem manajemen perpustakaan yang dibangun menggunakan Go dan MySQL.

## Prasyarat

1.  **Go** (Golang) sudah terinstall.
2.  **MySQL** server sudah berjalan.

## Konfigurasi Database

Secara default, aplikasi akan mencoba terhubung ke database dengan konfigurasi berikut:
- **Host**: `localhost`
- **Port**: `3306`
- **User**: `root`
- **Password**: `""` (kosong)
- **Database Name**: `jwt_auth_db`

### Langkah 1: Buat Database
Pastikan database sudah dibuat di MySQL sebelum menjalankan aplikasi. Jalankan query berikut di MySQL Client Anda:

```sql
CREATE DATABASE jwt_auth_db;
```

*(Tabel-tabel akan dibuat otomatis oleh aplikasi saat pertama kali dijalankan)*

## Cara Menjalankan

Buka terminal (Command Prompt atau PowerShell) di folder proyek ini, lalu jalankan perintah:

```powershell
go run main.go
```

Jika berhasil, Anda akan melihat output seperti:
```
✅ Successfully connected to MySQL database
🚀 Server running on http://localhost:8080
```

## Konfigurasi Lanjutan (Opsional)

Jika konfigurasi MySQL Anda berbeda (misalnya ada password), Anda bisa set environment variable sebelum menjalankan aplikasi:

**PowerShell:**
```powershell
$env:DB_PASS="password_anda"; $env:DB_NAME="nama_db_anda"; go run main.go
```

**CMD:**
```cmd
set DB_PASS=password_anda && set DB_NAME=nama_db_anda && go run main.go
```
