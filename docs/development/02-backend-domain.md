# 02 — Kế hoạch Backend, domain và API

## Cấu trúc Go

```text
cmd/
  api/                 main HTTP server
  worker/              main job runner/scheduler
internal/
  platform/http/       router, middleware, JSON/problem response
  platform/postgres/   pool, tx helper, migration, repository primitives
  platform/jobs/       durable queue, lease, retry, outbox
  platform/security/   auth, RBAC, encryption, audit, secret interface
  identity/            user/member/session use cases
  ledger/              account, category, transaction, transfer
  portfolio/           asset/property/valuation/snapshot
  loans/               loan, schedule, accrual, payment split
  planning/            recurring/budget/forecast/goal
  bankfeed/            provider adapter, ingestion, matching, classifier
  assistant/           command gateway/executor adapter
  contracts/           request/response DTO and generated OpenAPI glue
```

Handler chỉ parse/auth/validate sơ bộ → gọi application use case → map response. Use case mở DB transaction, kiểm permission và gọi repository/domain policy. Repository không chứa business decision; job không gọi HTTP handler. Chia theo module nhưng build/deploy một binary API và một binary worker.

## Chuẩn HTTP/API

- Base `/api/v1`, JSON UTF-8, UTC ISO-8601; input decimal là string, VND không giả định float.
- Mọi response có `traceId`; list cursor pagination gồm `items`, `nextCursor`.
- `POST` mutation yêu cầu `Idempotency-Key` UUID do client tạo, trừ callback/webhook provider.
- Error: `{ "code": "…", "message": "…", "fields": { "amount": "…" }, "traceId": "…" }`.
- JWT xác định actor; user chọn từ path/header chỉ sau membership check. Không tin `user_id` từ body.
- `GET` không mutation; `PATCH` chỉ cho update metadata cho phép; financial correction dùng endpoint adjustment/void chuyên biệt.

## Middleware theo thứ tự

1. Request ID / trusted proxy / HTTPS redirect.
2. Body-size limit, content type, JSON decoder strict.
3. Rate limit theo IP + actor/route.
4. Authentication; `RequireUserRole` theo endpoint.
5. Idempotency middleware cho mutation: tìm existing result hoặc reserve key trong DB transaction.
6. Handler/use case; audit interceptor lấy before/after redacted.
7. Structured access log, metric/status/latency, panic recovery.

Webhook SePay dùng router riêng: raw body trước JSON parsing, HMAC verification, provider-specific rate limit, không dùng end-user JWT/idempotency header.

## Database transaction patterns

### Mutation chuẩn

```text
BEGIN
  assert actor membership + role
  lock/validate referenced aggregate
  enforce idempotency unique (user, actor, key)
  append domain record(s)
  append audit log
  append outbox/job if projection/notification required
  persist idempotency response
COMMIT
```

Không gửi email, gọi SePay, render export hay rebuild snapshot bên trong transaction. Nếu commit thành công nhưng client timeout, cùng idempotency key trả response gốc. Nếu fail, rollback toàn bộ including audit/outbox.

### Điều chỉnh và void

- `posted`/`reconciled` financial record không UPDATE amount/type trực tiếp.
- Reclassify tạo adjustment/compensating transaction liên kết `original_transaction_id`, audit nêu reason/actor/source.
- Void chỉ bởi role phù hợp; UI/API trả trạng thái lịch sử thay vì hard delete.
- Snapshot invalidation tạo job versioned; read model có thể hơi trễ nhưng không được viết ngược vào ledger.

## Domain use case theo module

### Identity/user

| Use case | Validate | Side effects |
|---|---|---|
| Create user | base currency, owner | portfolio mặc định, audit |
| Invite/change role | owner, role policy | membership/audit/notification job |
| Export | re-auth/permission/range | export job + signed URL + audit |
| Delete/revoke | retention/legal/audit policy | async workflow, không hard-delete ledger |

### Ledger

| Use case | Input chính | Invariant/ghi dữ liệu |
|---|---|---|
| Create account | type/currency/opening balance | account + opening transaction/source rõ |
| Post income/expense | account/amount/category/occurredAt | amount > 0; category bắt buộc; append transaction |
| Create transfer | source/destination/amount | lock hai account; tạo hai vế nguyên tử; no category |
| Reclassify | original + new type/category + reason | compensation/adjustment, audit, snapshot job |
| Reconcile account | as-of/provider balance | không sửa ledger; tạo reconciliation result/exception |

### Portfolio and valuation

- `CreateAsset/CreateProperty`: ownership, currency, acquisition metadata; không trộn valuation vào transaction cash.
- `AddValuation`: amount > 0, effective date hợp lệ, source required; append-only; invalidate snapshots từ effective date.
- `GetNetWorth`: đọc snapshot mới nhất phù hợp as-of; nếu missing/stale, enqueue rebuild và trả data quality rõ.
- `ReviewSnapshot/LockSnapshot`: owner/editor theo policy; adjustment sau lock tạo record mới, không mutate snapshot cũ.

### Loan

- `CreateLoan`: validate direction, principal, rate/day-count, start/due/schedule; sinh schedule transactionally hoặc job rõ version.
- `AccrueInterest`: worker lấy loan active, unique `(loan_id, date, method)`; decimal library, no float; lock/check run date.
- `RecordPayment`: link bank/ledger transaction, validate split total, principal ≤ outstanding (trừ approved adjustment); cập nhật derived status và snapshot job.
- `Restructure/WriteOff`: owner, reason/attachment, explicit effect policy, audit trước/sau.

### Planning

- Recurring rule chỉ tạo candidate/job idempotent theo `(rule_id, occurrence_at)`.
- Forecast run lưu input snapshot version, assumptions, generated events, engine version và output; caller nhận `jobId`.
- Budget query dựa posted expense trong kỳ, không tính transfer/valuation/loan principal.

## Job queue trong PostgreSQL

`jobs(id, type, payload_json, run_at, status, attempts, max_attempts, locked_at, locked_by, last_error, correlation_id, created_at)`.

Worker claim batch:

```sql
SELECT id FROM jobs
WHERE status = 'queued' AND run_at <= now()
ORDER BY run_at, id
FOR UPDATE SKIP LOCKED
LIMIT $1;
```

Sau claim, set lease. Success mark `completed`; lỗi retryable exponential backoff + jitter; lỗi non-retryable `failed`; quá attempts `dead_letter`. Scheduler requeue lease hết hạn. Mỗi handler idempotent và đo duration/success/failure/lag.

| Job | Trigger | Idempotency | Retry |
|---|---|---|---|
| `rebuild_snapshot` | transaction/valuation/loan change | portfolio + as-of/source version | Có |
| `loan_accrual` | daily scheduler | loan/date/method unique | Có |
| `process_bank_event` | webhook/backfill | provider transaction unique | Có |
| `reconcile_connection` | schedule/manual | connection/as-of window | Có/rate-limited |
| `run_forecast` | user action | scenario/input version | Có |
| `generate_export` | user action | export request ID | Có |
| `send_notification` | inbox/event | notification dedupe key | Có |

## Authorization matrix

| Action | Owner | Editor | Viewer | Service/provider |
|---|---|---|---|---|
| Read permitted portfolio | ✓ | ✓ | ✓ | scope-bound |
| Post transaction/valuation | ✓ | ✓ | — | bank-feed policy only |
| Loan restructure/write-off | ✓ | policy optional | — | — |
| Connect/revoke bank | ✓ | — | — | callback only |
| Export full user | ✓ | — | — | — |
| Manage members/policy | ✓ | — | — | — |
| Receive provider webhook | — | — | — | signed provider endpoint |

Authorization is server-side in every use case, not only router/UI. Attachment access runs the same scope check before issuing a signed URL.

## API contract implementation order

1. Auth/user/categories/accounts.
2. Transactions/transfers/net-worth read/snapshots.
3. Assets/property/valuations/attachments.
4. Loans/accrual/payment/agenda.
5. Forecast/budget/inbox.
6. Bank connections/feed/rules/payment requests.
7. Assistant read API, then command workflow.

For every endpoint: OpenAPI request/response/error examples, permission test, validation test, idempotency test if mutate and FE contract mock must be added in the same pull request.

## Observability fields

All logs/metrics/traces include `trace_id`, route/job type, user hash (not raw ID where avoidable), actor type and provider correlation ID. Financial amount/account note/full account number/token must not be in log. Audit is not a debug log and is stored separately with redaction policy.
