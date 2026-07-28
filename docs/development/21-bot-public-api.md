# Bot Public API

API này dành cho bot/automation ghi nhận giao dịch vào **một** account Finora. Không dùng JWT; mỗi request bắt buộc có secret riêng theo account trong header `X-Finora-Account-Key`.

## 1. Tạo hoặc xoay secret

Người dùng đăng nhập Finora với vai trò owner/editor của user gọi:

```http
POST /api/v1/accounts/{accountId}/bot-api-key
Authorization: Bearer <user-jwt>
X-User-ID: <userId>
```

Response trả `secret` đúng một lần. Lưu nó trong secret manager của bot. Gọi lại endpoint sẽ **xoay** secret cũ; bot đang dùng secret cũ sẽ bị từ chối. Finora chỉ lưu SHA-256 hash, không log hoặc trả lại secret.

## 2. Ghi một giao dịch

```http
POST /public/v1/accounts/{accountId}/transactions
Content-Type: application/json
X-Finora-Account-Key: finora_bot_...

{
  "type": "expense",
  "amount": "125000",
  "name": "Ăn trưa",
  "categoryId": "<optional-category-id>",
  "note": "Bot ghi nhận",
  "occurredAt": "2026-07-27T12:30:00+07:00"
}
```

- `type`: chỉ `income` hoặc `expense`.
- `amount`: số dương, dùng chuỗi để tránh sai số JSON.
- `occurredAt`: tùy chọn; RFC3339 hoặc `YYYY-MM-DD`; thiếu thì dùng thời điểm server nhận request.
- Currency lấy từ account; bot không thể đổi user, account hay currency.

Response `201` chứa `transaction`. Source luôn là `bot_public_api`.

## 3. Lấy lịch sử theo khoảng ngày

```http
GET /public/v1/accounts/{accountId}/transactions/history?from=2026-07-01&to=2026-07-31
X-Finora-Account-Key: finora_bot_...
```

`from` và `to` là bắt buộc, theo `YYYY-MM-DD` hoặc RFC3339. Với `to` dạng ngày, API bao gồm hết ngày đó. Response `200` trả `accountId`, khoảng thời gian đã chuẩn hóa UTC và `items` theo thứ tự mới nhất trước.

## Bảo mật và vận hành

- Không đưa secret vào prompt, log, Git hoặc mobile app.
- Chỉ gọi HTTPS; secret là credential cấp account, không thay cho JWT người dùng.
- Khi nghi lộ key, gọi lại endpoint tạo key để xoay ngay.
- API không tự tạo transfer, loan hoặc thay đổi transaction đã có.
