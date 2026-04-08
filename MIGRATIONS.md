# Database Migrations

โปรเจ็กต์นี้ใช้ versioned migrations ผ่าน package `migrations/` และตาราง `schema_migrations`

## คำสั่ง

รันเฉพาะ migration แล้วออก:

```bash
./loan-app-linux migrate
```

รันแอปตามปกติ:

```bash
./loan-app-linux
```

ตอนเริ่มแอป ระบบจะตรวจและรัน pending migrations ให้อัตโนมัติ

## ตารางที่ใช้ติดตาม

- `schema_migrations`

## แนวทางเพิ่ม migration ใหม่

1. เพิ่มรายการใหม่ใน `migrations/migrations.go`
2. ใช้ version แบบเรียงลำดับ เช่น `2026040803`
3. ห้ามแก้ migration ที่ deploy ไปแล้ว ให้เพิ่มรายการใหม่แทน
