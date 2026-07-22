# WealthOS - Angular 21 + Go Backend Scaffold

Repo này đã được dựng theo tài liệu tính năng để tạo nền tảng:
- Frontend: Angular 21 + Tailwind CSS
- Backend: Go API theo hướng Modular Monolith
- Giao tiếp API theo đường dẫn `/api/v1` với các endpoint theo file [docs/05-api-design.md](./docs/05-api-design.md)

## 1) Chạy backend

```bash
cd backend
go mod download
go run ./cmd/server
```

Mặc định backend chạy tại `http://localhost:8080`.

### Biến môi trường

- `APP_ENV`: môi trường môi trường
- `APP_PORT`: cổng API (mặc định 8080)
- `APP_STATIC_TOKEN`: token demo cho request nhanh
- `DATABASE_URL`: chuỗi kết nối PostgreSQL

## 2) Chạy frontend

```bash
cd frontend
npm install
npm start
```

Frontend mặc định chạy tại `http://localhost:4200` và gọi API `http://localhost:8080`.

## 3) Login nhanh

- Đăng nhập với user mặc định (demo): `demo@wealthos.vn` / `demo-pass`.
- Backend hiện đang hỗ trợ profile in-memory cho giai đoạn khởi tạo.

## 4) Dùng Docker Compose

```bash
docker compose up --build
```

## 5) Ghi chú mở rộng

- Đây là khung dự án để tiếp tục implement đầy đủ nghiệp vụ tài chính theo tài liệu:
  - Tự động nhận diện tiền vào (inbound rule + confidence)
  - Auto-post chi theo quy tắc khi gửi đi
  - SePay OAuth/Webhook + reconciliations
  - Loan accrual jobs, forecast engine, assistant command flow



- `SEPAY_WEBHOOK_SECRET`: chu?i b� m?t SePay webhook (d�ng cho x�c th?c signature webhooks)
