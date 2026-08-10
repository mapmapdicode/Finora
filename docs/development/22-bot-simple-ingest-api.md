# Bot API tối giản — ghi Thu/Chi và Thu lãi/Gốc

Tài liệu này dành cho bot cá nhân của Finora. Bot chỉ cần biết `userId` của chủ tài khoản, **không cần đăng nhập JWT** và không cần account key. API phù hợp cho automation đáng tin cậy do chính người dùng vận hành.

> Cảnh báo bảo mật: đây là tuyến ghi dữ liệu công khai theo UUID người dùng. Không dùng cho bot bên thứ ba hoặc môi trường không kiểm soát. Với tích hợp bên thứ ba, dùng API an toàn hơn tại [21-bot-public-api.md](./21-bot-public-api.md), có `X-Finora-Account-Key`.

## URL gốc và quy tắc chung

Production hiện tại:

```text
http://110.172.29.117:2001/public/v1
```

- Mọi lệnh `POST` bắt buộc có header `Idempotency-Key`. Dùng một UUID mới cho mỗi nghiệp vụ; gửi lại cùng key và cùng body sẽ không tạo trùng.
- `userId`, `accountId`, `loanId` là UUID Finora. Lấy chúng từ endpoint **Ngữ cảnh bot** bên dưới.
- Tất cả tiền là chuỗi số VND dương, không dấu chấm, không `VND`: `"125000"`, `"6000000"`.
- `occurredAt` nhận `YYYY-MM-DD` hoặc RFC3339. Nên dùng RFC3339 có múi giờ Việt Nam, ví dụ `2026-08-09T19:30:00+07:00`.
- Với câu như “chi hôm nay”, bot nên **bỏ hẳn** `occurredAt`; Finora sẽ dùng giờ server hiện tại. Chỉ gửi `occurredAt` khi người dùng nêu một ngày cụ thể hoặc bot có ngày chắc chắn.
- API trả JSON lỗi theo dạng `{ "code": "...", "message": "..." }`.

### Lấy `userId` một lần

Sau khi đăng nhập Finora, response của `POST /api/v1/auth/login` có trường `user.id`. Lưu UUID này trong cấu hình bot. Với tài khoản `hoangxuan.ks6@gmail.com` trên production hiện tại, `userId` là:

```text
d2932d32-40c3-4654-bf5d-a2f7c4c6f08f
```

## 1. Ngữ cảnh bot: lấy tài khoản và khoản vay đang hiệu lực

```bash
curl 'http://110.172.29.117:2001/public/v1/users/USER_ID/context'
```

Response trả `accounts` và `openLoans`. Bot cần gọi bước này khi bắt đầu phiên hoặc khi chưa có ID cần thiết:

```json
{
  "userId": "USER_ID",
  "accounts": [
    { "id": "ACCOUNT_ID", "name": "Tiền mặt", "currency": "VND" }
  ],
  "openLoans": [
    {
      "id": "LOAN_ID",
      "counterparty": "Nguyễn Văn A",
      "direction": "receivable",
      "principalBalance": "200000000",
      "dailyRatePerMillion": "3000",
      "status": "active"
    }
  ]
}
```

Chỉ khoản có `status` đang hiệu lực được trả trong `openLoans`; khoản đã tất toán không thể nhận thêm thanh toán.

## 2. Báo cáo lãi cộng dồn hằng ngày (dùng thẳng cho cronjob)

```http
GET /public/v1/users/{userId}/loans/accrual-report
```

```bash
curl -s 'http://110.172.29.117:2001/public/v1/users/USER_ID/loans/accrual-report' \
  | jq -r '.markdown'
```

Response có hai phần:

- `loans` và `totals`: JSON số tiền chính xác để bot tiếp tục xử lý.
- `markdown`: chuỗi hoàn chỉnh để bot gửi nguyên văn vào Telegram/Zalo/Discord; server tính theo lịch nhận lãi/gốc thực tế, bot không cần tự cộng ngày.

Ví dụ `markdown`:

```text
📊 Lãi cộng dồn — 10/08/2026

loan_17_0710 (30M) — 30 ngày — 2,700k
loan_01_0209 (100M) — 29 ngày — 8,700k
...

Tổng gốc: 1,198M
Lãi/ngày: 3,594k
Tổng lãi cộng dồn: 57,963k
```

Mỗi dòng gồm mã khoản vay Markdown nếu có (nếu không dùng tên đối tác), gốc còn lại, số ngày từ lần nhận lãi gần nhất và lãi chưa nhận. Khoản đã tất toán hoặc khoản đi vay không được đưa vào báo cáo phải thu này.

## 3. Ghi thu nhập hoặc chi tiêu

```http
POST /public/v1/users/{userId}/transactions
Content-Type: application/json
Idempotency-Key: <UUID mới cho nghiệp vụ này>
```

Ví dụ ghi chi tiền ăn vào tài khoản tiền mặt:

```bash
curl -X POST 'http://110.172.29.117:2001/public/v1/users/USER_ID/transactions' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: 31904415-dc47-48f7-98ca-b3e4a2e626ab' \
  --data '{
    "accountId": "ACCOUNT_ID",
    "type": "expense",
    "amount": "125000",
    "name": "Ăn trưa",
    "note": "Bot ghi từ tin nhắn",
    "occurredAt": "2026-08-09T12:30:00+07:00"
  }'
```

Ví dụ ghi thu nhập:

```json
{
  "accountId": "ACCOUNT_ID",
  "type": "income",
  "amount": "15000000",
  "name": "Lương tháng 8",
  "note": "Chuyển khoản",
  "occurredAt": "2026-08-05"
}
```

| Trường | Bắt buộc | Quy tắc |
|---|---:|---|
| `type` | Có | Chỉ `income` (thu) hoặc `expense` (chi). |
| `amount` | Có | Chuỗi số dương VND. |
| `name` | Nên có | Tên ngắn để nhìn trong sổ giao dịch. |
| `accountId` | Không | Không gửi thì Finora dùng tài khoản đầu tiên của user. Bot nên gửi rõ để tránh ghi nhầm khi có nhiều tài khoản. |
| `categoryId`, `note`, `occurredAt` | Không | Có thể bỏ trống. |

Response `201` chứa `{ "transaction": {...}, "accountId": "..." }`.

## 4. Sửa giao dịch bot đã ghi nhầm

Bot có thể sửa giao dịch nó đã tạo bằng `transaction.id` trả về khi tạo. Chỉ giao dịch có `source: "bot_user_id_api"` được sửa bằng API này; Finora chặn sửa bút toán khoản vay, chuyển tiền và ngân hàng để không làm sai số dư liên quan.

```http
PATCH /public/v1/users/{userId}/transactions/{transactionId}
Content-Type: application/json
Idempotency-Key: <UUID mới cho lần sửa này>
```

Ví dụ sửa giao dịch “Ăn cả ngày” bị bot ghi sai ngày 02/08 thành tối 09/08:

```bash
curl -X PATCH 'http://110.172.29.117:2001/public/v1/users/USER_ID/transactions/TRANSACTION_ID' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: 6ab1f8c2-3960-4450-a8b5-551037c0dd2d' \
  --data '{
    "occurredAt": "2026-08-09T23:47:00+07:00",
    "amount": "150000",
    "name": "Ăn cả ngày",
    "note": "Bot sửa lại ngày theo yêu cầu người dùng"
  }'
```

Tất cả trường trong body là tùy chọn; chỉ trường xuất hiện mới bị thay đổi: `accountId`, `type` (`income`/`expense`), `amount`, `name`, `categoryId`, `note`, `occurredAt`. Response `200` trả `{ "transaction": {...} }` đã cập nhật.

## 5. Ghi nhận thu lãi, thu gốc hoặc cả hai

```http
POST /public/v1/users/{userId}/loan-payments
Content-Type: application/json
Idempotency-Key: <UUID mới cho nghiệp vụ này>
```

### Chỉ thu lãi

```bash
curl -X POST 'http://110.172.29.117:2001/public/v1/users/USER_ID/loan-payments' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: 6ef17697-ece1-4875-817f-8a4cb33df1e0' \
  --data '{
    "loanId": "LOAN_ID",
    "interestAmount": "6000000",
    "occurredAt": "2026-08-20T09:00:00+07:00"
  }'
```

### Chỉ thu gốc

```json
{
  "loanId": "LOAN_ID",
  "principalAmount": "50000000",
  "occurredAt": "2026-08-25"
}
```

### Thu cả lãi và gốc

```json
{
  "loanId": "LOAN_ID",
  "accountId": "ACCOUNT_ID",
  "interestAmount": "6000000",
  "principalAmount": "50000000",
  "occurredAt": "2026-08-25T10:00:00+07:00"
}
```

| Trường | Bắt buộc | Quy tắc |
|---|---:|---|
| `loanId` | Có | Phải là ID trong `openLoans` của chính `userId`. |
| `interestAmount` | Không | Lãi nhận được; bỏ trống tương đương `0`. |
| `principalAmount` | Không | Gốc nhận được; bỏ trống tương đương `0`. Không được lớn hơn dư nợ hiện tại. |
| `accountId` | Không | Tài khoản nhận/trả tiền. Bỏ trống sẽ dùng tài khoản giải ngân của khoản vay, rồi tài khoản đầu tiên nếu khoản vay chưa có tài khoản đó. |
| `interestDays`, `feeAmount`, `waivedAmount`, `occurredAt` | Không | Dùng khi bot có thông tin chi tiết. Mặc định là `0`/thời điểm gọi API. |

Ít nhất một trong `interestAmount`, `principalAmount`, `feeAmount` phải lớn hơn 0. Khi toàn bộ dư nợ gốc được thu, hợp đồng chuyển sang `closed` (đã tất toán). Response `201` chứa `{ "payment": {...}, "loanId": "..." }` và đồng thời tạo dòng tiền `loan_payment` cho tài khoản nhận tiền.

## Mẫu xử lý cho bot

1. Gọi `/users/{userId}/context` để tìm account/loan phù hợp.
2. Hiểu ý định người dùng: `thu`, `chi`, `thu lãi`, `thu gốc`, hoặc `thu cả hai`.
3. Chuẩn hóa tiền thành chuỗi chữ số VND; không tự thêm dấu âm.
4. Tạo `Idempotency-Key` mới, gọi đúng endpoint một lần.
5. Chỉ báo thành công sau khi nhận HTTP `201`; nếu timeout, gửi lại **cùng** key và body để an toàn.

Các lỗi thường gặp: `USER_NOT_FOUND`, `ACCOUNT_NOT_FOUND`, `LOAN_NOT_FOUND`, `LOAN_CLOSED`, `MISSING_IDEMPOTENCY_KEY`, `DUPLICATE_IDEMPOTENCY_KEY` và `BAD_REQUEST`.
