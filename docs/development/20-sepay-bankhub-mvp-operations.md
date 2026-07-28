# SePay Bank Hub MVP — vận hành và rollout

## Cấu hình

Thiết lập các biến môi trường **chỉ tại backend**:

```dotenv
SEPAY_BANKHUB_CLIENT_ID=
SEPAY_BANKHUB_CLIENT_SECRET=
SEPAY_BANKHUB_COMPANY_XID=
SEPAY_BANKHUB_IPN_API_KEY=
SEPAY_BANKHUB_BASE_URL=https://bankhub-api.sepay.vn
SEPAY_BANKHUB_PILOT_BANK_CODES=MBB,TPB,VCB
```

`SEPAY_BANKHUB_BASE_URL` có thể thay bằng sandbox theo tài liệu Bank Hub. Không đưa bất kỳ biến nào vào Flutter, frontend build-time config, log hoặc audit payload. Trong production, dùng secret manager và xoay IPN API key theo quy trình của SePay.

## Thiết lập provider

SePay có hai luồng webhook khác nhau; không trỏ chúng vào cùng một URL:

| Sản phẩm SePay | URL Finora | Payload / xác thực |
| --- | --- | --- |
| SePay Webhooks trên dashboard | `https://<api-host>/hooks/sepay-webhook` | `id`, `transferType`, `transferAmount`; HMAC `SEPAY_WEBHOOK_SECRET` |
| Bank Hub cho Hosted Link của MVP | `https://<api-host>/hooks/sepay-bankhub-ipn` | `bank_account_xid`, `transaction_id`, `credit`/`debit`; API key `SEPAY_BANKHUB_IPN_API_KEY` |

1. Đăng ký company sandbox trước, sau đó production; lưu `company_xid` tương ứng môi trường.
2. Đăng ký IPN `POST https://<api-host>/hooks/sepay-bankhub-ipn`. Ví dụ production: `https://api.finora.vn/hooks/sepay-bankhub-ipn`. URL này không chứa version API để giữ ổn định khi deploy. Endpoint cũ `/api/v1/webhooks/sepay/bankhub/ipn` vẫn được giữ tương thích; event server-side (nếu SePay cấp) dùng `/api/v1/webhooks/sepay/bankhub/events`.
3. Đặt header `Authorization: Apikey <SEPAY_BANKHUB_IPN_API_KEY>` tại provider.
4. Bật duy nhất 2–3 bank code trong `SEPAY_BANKHUB_PILOT_BANK_CODES`. Account chỉ map được khi Bank Hub báo kết nối/hoạt động và bank code nằm trong allowlist; trong MVP, allowlist là capability registry cho cả tiền vào lẫn tiền ra.

## Luồng và invariant

- Mobile chỉ nhận `hosted_link_url`, không nhận OAuth/client token hoặc API key.
- IPN xác thực API key, insert event durable theo provider-account + provider transaction ID, trả `{"success":true}`, và worker xử lý sau đó. Event hợp lệ đến trước mapping được lưu `sepay_unmapped_events` với trạng thái `quarantined` nhưng vẫn ACK; không được bỏ qua hoặc retry vô hạn.
- Worker claim event theo compare-and-set. Reconciliation dùng overlap 5 phút, provider ID và `company_xid` trong profile của chính user để re-import an toàn.
- Raw provider data là read-only; feed mới luôn `needs_review`. Rule, history và semantic matching chỉ tạo suggestion. Chỉ `confirm` hoặc `correct` của owner tạo transaction Finora.
- Giao dịch transfer, loan và refund không được auto-resolve trong MVP.

## Kiểm tra sandbox trước pilot

| Case | Kỳ vọng |
|---|---|
| Inbound / outbound | Một feed `needs_review`; không có ledger transaction tự tạo. |
| IPN sai key | `401`, không có event. |
| Same transaction 100 lần | Một source event và một feed source duy nhất. |
| Replay sau reconciliation | Không tạo feed mới. |
| Cancel / expired hosted link | Không tạo bank account hoặc mapping. |
| Account one-way / ngoài allowlist | Hiện `unsupported`, không map được. |
| Confirm/correct/ignore | Có audit + feedback; immutable provider fields không đổi. |

## Theo dõi và rollback

Scrape `GET /metrics` (network-restricted trong production) cho `sepay_webhook_failures_total`, `sepay_unknown_bank_account_total`, `sepay_queue_lag_seconds_total`, `sepay_reconciliation_failures_total`, `sepay_ai_suggestions_corrected_total` và `sepay_user_confirmed_rule_total`. Kết hợp log theo request ID cho queue retry, confirm/correct/ignore, và suggestion source/confidence. Không log raw account number, OTP, token, hoặc nội dung giao dịch đầy đủ.

Nếu pilot có lỗi: xóa bank code khỏi allowlist, ngừng tạo link session mới, giữ worker ingest để tránh mất event, và để người dùng xử lý các feed hiện có qua inbox. Không xóa source event hoặc transaction lịch sử.

## Assumption cần xác minh với SePay

- `bank_account_xid` được coi là opaque string, không được giả định UUID.
- Mobile Hosted Link nhận sự kiện `FINISHED_BANK_ACCOUNT_LINK` qua `window.postMessage`, sau đó gọi endpoint user-owned `POST /me/sepay/bank-accounts/sync` với account number từ sự kiện. Backend dùng bearer token của mình gọi Bank Hub `GET /v1/bank-account` và chỉ lưu kết quả có số tài khoản khớp chính xác. Không tin `bank_account_xid`, capability hay bank code do mobile gửi lên.
- Endpoint webhook `BANK_ACCOUNT_LINKED` vẫn được giữ tương thích với các triển khai gửi event server-side, nhưng không là luồng hoàn tất chính của Hosted Link.
- `transaction_id` là định danh ổn định trong phạm vi provider account. Nếu provider bảo đảm global unique thì key scoped hiện tại vẫn an toàn.
