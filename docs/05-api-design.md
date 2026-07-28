# Thiết kế API WealthOS

Base URL: `/api/v1`. API dùng JSON, UTC ISO-8601, decimal dưới dạng chuỗi và access token suy ra user. Tất cả response tài sản ròng/forecast phải có `asOfAt`, `baseCurrency`, `dataQuality` và các nguồn cấu thành.

## Quy ước

- Danh sách phân trang cursor: `?limit=50&cursor=...`; mặc định ngày hiệu lực giảm dần.
- Header `Idempotency-Key` bắt buộc cho create transaction, payment, valuation và job-triggering request.
- API không trả một `netWorth` rời rạc: phải kèm `assets`, `liabilities`, `attribution` hoặc link drill-down.
- Lỗi theo mẫu `{ "code": "LOAN_NOT_ACTIVE", "message": "...", "traceId": "..." }`.

## Endpoint cốt lõi

| Method | Đường dẫn | Mục đích |
|---|---|---|
| `GET, POST` | `/portfolios` | Danh mục WealthOS |
| `GET` | `/portfolios/{id}/net-worth` | Net worth as-of, cơ cấu và attribution |
| `GET` | `/portfolios/{id}/snapshots` | Chuỗi thời gian net worth/growth |
| `GET, POST` | `/accounts` | Tiền mặt, ngân hàng, thẻ |
| `GET, POST` | `/transactions` | Thu/chi/dòng tiền, có phân loại nghiệp vụ |
| `POST` | `/transfers` | Chuyển tiền nguyên tử |
| `GET, POST` | `/loans` | Khoản phải thu/phải trả và điều khoản |
| `GET` | `/loans/{id}/accruals` | Lãi phát sinh, đã thu và chưa thu |
| `POST` | `/loans/{id}/payments` | Thu/trả gốc, lãi, phí |
| `GET, POST` | `/properties` | Bất động sản |
| `POST` | `/properties/{id}/valuations` | Thêm định giá append-only |
| `GET, POST` | `/assets` | Tài sản thủ công và định giá |
| `GET, PUT` | `/budgets/{period}` | Budget cho expense |
| `GET, POST` | `/forecast-scenarios` | Lưu kịch bản dự báo |
| `POST` | `/forecast-scenarios/{id}/run` | Chạy mô phỏng, trả nguồn giả định |
| `POST` | `/integrations/sepay/connect` | Bắt đầu Hosted Link/OAuth SePay, chỉ yêu cầu scope đọc cần thiết |
| `GET` | `/integrations/sepay/callback` | Hoàn tất liên kết server-side sau khi xác thực state/PKCE |
| `GET` | `/bank-connections` | Xem consent, trạng thái sync, data quality và connection health |
| `POST` | `/bank-connections/{id}/sync` | Yêu cầu backfill/đối soát, có rate limit |
| `POST` | `/bank-connections/{id}/revoke` | Thu hồi consent và vô hiệu connection/token có audit |
| `POST` | `/webhooks/sepay` | Provider-only: verify signature, dedupe, enqueue rồi trả thành công |
| `GET` | `/bank-feed-transactions` | Liệt kê giao dịch imported/candidate theo account và trạng thái |
| `POST` | `/bank-feed-transactions/{id}/approve` | Hạch toán candidate qua ledger service, idempotent |
| `POST` | `/bank-feed-transactions/{id}/reclassify` | Sửa type/category của import hoặc transaction tự ghi, tạo adjustment/audit |
| `POST` | `/bank-feed-transactions/{id}/ignore` | Bỏ qua candidate có lý do, giữ event nguồn/audit |
| `GET, POST` | `/bank-automation-rules` | Xem/tạo rule nhận diện tiền vào và phân loại chi tiêu tự động |
| `POST` | `/bank-automation-rules/preview` | Chạy rule chưa lưu trên import gần đây, chỉ trả affected count/sample; không mutation |
| `PATCH, DELETE` | `/bank-automation-rules/{id}` | Bật/tắt/sửa rule hoặc ngừng áp dụng rule mới; không sửa giao dịch lịch sử |
| `POST` | `/loans/{id}/payment-requests` | Tạo VietQR payment request cho một kỳ thu loan |
| `POST` | `/assistant/telegram/webhook` | Nhận Telegram update đã xác thực webhook secret |
| `POST` | `/assistant/commands` | Tạo command từ web/mobile hoặc adapter Telegram |
| `GET` | `/assistant/commands` | Danh sách command của user hiện tại |
| `GET` | `/assistant/commands/{id}` | Xem trạng thái, plan, approval và kết quả |
| `POST` | `/assistant/commands/{id}/approve` | Phê duyệt command cần confirmation một lần |
| `POST` | `/assistant/commands/{id}/cancel` | Hủy command chưa hoàn tất |
| `POST` | `/assistant/executors/{id}/events` | Hermes executor gửi execution event có chữ ký |

## Ví dụ trả về net worth

```json
{
  "asOfAt": "2026-07-17T23:59:59+07:00",
  "baseCurrency": "VND",
  "netWorth": "1170000000.00",
  "assets": { "cash": "160000000.00", "receivables": "1010000000.00" },
  "liabilities": "0.00",
  "dataQuality": { "reconciledAccounts": 2, "staleValuations": 0 },
  "attribution": { "externalCashFlow": "0.00", "accruedInterest": "0.00", "valuationChange": "0.00" }
}
```

Server, không phải client, chịu trách nhiệm phân quyền, đối soát loại giao dịch và công thức portfolio.

Chi tiết contract Telegram–Gateway–Hermes: [15-hermes-telegram-assistant.md](15-hermes-telegram-assistant.md).

Chi tiết consent, webhook, đối soát và VietQR với SePay: [17-sepay-bank-integration.md](17-sepay-bank-integration.md).
