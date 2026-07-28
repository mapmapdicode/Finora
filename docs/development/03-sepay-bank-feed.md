# 03 — Build plan SePay bank feed và tự ghi nhận thu–chi

## Mục tiêu kỹ thuật

Nhận event SePay nhanh và chính xác, để tiền ra tự tạo expense, tiền vào chỉ tạo income khi có evidence đủ mạnh. Không để retry/replay tạo bản ghi trùng, không để classifier vượt qua loan/transfer rule, và luôn cho user sửa kết quả.

## Backend sequence

### A. Kết nối và consent

1. `POST /integrations/sepay/connect` tạo connection draft: user/account target, nonce/state, requested read scopes, expiry.
2. Redirect user tới Hosted Link/OAuth; callback kiểm state, PKCE và actor/user server-side.
3. Mã hóa token/refresh token, lưu provider account mapping/capabilities, tạo `bank_connections.active`.
4. Frontend nhận connection status, không nhận token; user có thể revoke → disable jobs/webhook mapping, revoke provider token nếu API hỗ trợ, audit.

### B. Webhook ingress

1. Router đọc raw body giới hạn kích thước, kiểm HMAC timestamp/signature constant-time.
2. Validate schema tối thiểu: event ID, transaction date, account number reference, direction, integer amount.
3. Trong transaction: `INSERT bank_feed_events` unique provider/event ID; nếu mới, enqueue `process_bank_event`; commit.
4. Trả đúng success response trong thời hạn provider. Duplicate đã tồn tại cũng trả success; signature/schema fail trả lỗi, không enqueue.
5. Không chạy AI, match hoặc ledger write trong request webhook.

### C. Normalizer và matcher worker

1. Resolve `bank_connection` bằng provider account mapping; event không map được → quarantined + alert, không đoán user.
2. Normalize date/timezone, direction `in/out`, amount, content/reference masked/full field policy; insert `bank_feed_transactions` unique.
3. Chạy matcher theo thứ tự: reversal/correction → internal transfer → loan/VietQR payment request → explicit business rule → income/expense classification.
4. Lưu matched entity, evidence JSON, classifier/rule version, confidence và decision; tạo ledger record hoặc review item qua transaction riêng/idempotent.

## Auto-post policy cụ thể

### Tiền ra (`out`)

```text
if internal transfer match: post transfer
else if approved loan disbursement rule: post loan_disbursement
else if investment funding rule: post investment_funding
else: post expense(category = highest-priority rule or Uncategorized)
```

Điều kiện transfer match tối thiểu: account khác cùng user, chiều đối lập, same currency/amount và window thời gian cấu hình (ví dụ ±2 ngày); nếu nhiều candidate hoặc confidence thấp thì tạo review, không tự ghép. Auto expense có `source=bank_feed`, `classification=auto`, `posted_transaction_id` và audit.

### Tiền vào (`in`)

```text
if internal transfer match: post/offer transfer
else if payment request or loan reference matches: create payment_candidate
else if income score >= configured threshold: post income
else: pending_review
```

Score chỉ lấy từ evidence có thể giải thích. Default: rule exact = 100; user-confirmed recurring pattern = 70; user keyword = 45; unknown = 0; threshold = 70. Không dùng model/LLM để auto-post income trong MVP. Mọi transaction auto income phải hiện reason/rule/score trong UI và có endpoint reclassify.

## Rule engine v1

`bank_automation_rules` fields: `id`, `user_id`, nullable `account_id`, `priority`, `direction`, `content_pattern`, `reference_pattern`, `min_amount`, `max_amount`, `recurrence_hint`, `action_type`, `category_id`, `enabled`, `version`, `created_by`.

- Sort account rule trước user rule, sau đó priority giảm dần và updated time tăng dần.
- Regex phải bị giới hạn độ dài/thời gian hoặc dùng safe matcher; không execute regex user-controlled không giới hạn.
- Preview chạy read-only trên tối đa 100 imported transaction và trả số affected + sample; enable cần confirmation.
- Rule edit chỉ áp dụng event mới. “Áp dụng lại lịch sử” là job riêng với preview/approval/audit.

## FE task breakdown

| View | Thành phần | API/state |
|---|---|---|
| Connect bank | consent modal, capability list, redirect state | connect/callback/connection query |
| Connection detail | health, last sync, provider capability, revoke | connection/sync/revoke |
| Review inbox | filters, tabs, row/result badges | feed list cursor query |
| Feed drawer | masked source, evidence, ledger link, audit | feed detail/reclassify/ignore |
| Rule builder | condition builder, priority, preview, confirm | CRUD + preview rule |
| VietQR request | amount/expiry/code/QR, payment status | create request/status |

## Error and recovery policy

| Failure | System action | User/operator action |
|---|---|---|
| HMAC invalid | 401/record security metric, no DB event | alert threshold only |
| DB unavailable | no ACK → provider retry | incident; no manual ledger entry |
| Queue stuck | event persisted, alert lag | restart/scale worker, replay safe job |
| Provider API 429 | backoff, retain cursor/window | show sync delayed |
| Unknown connection | quarantine event | investigate mapping; do not assign user |
| Auto-post wrong | reclassify creates adjustment/audit | offer rule for future, measure rate |

## Acceptance tests

- 100 concurrent duplicate webhook deliveries result in one imported event and one ledger record maximum.
- An `out` unmatched event produces exactly one posted expense and snapshot invalidation job.
- An `out` paired transfer does not change total cash or budget expense.
- An `in` with loan payment code creates candidate, never income.
- An unknown `in` remains review; exact salary rule produces income with evidence 100.
- Reclassify an auto expense preserves raw event and produces audit/adjustment.
- Connection revoke stops future sync and does not erase historic ledger/import.
