# 04 — Chất lượng, bảo mật và kế hoạch phát hành

## Test pyramid

| Tầng | Công cụ/gợi ý | Phạm vi bắt buộc |
|---|---|---|
| Unit | Go test, Angular unit test | Decimal, domain policy, mapper, form validation, UI state |
| Integration | PostgreSQL thật bằng container | Repository, migration, transaction rollback, RBAC, idempotency, jobs |
| Contract | OpenAPI validation + FE typed client | Request/response/error, enum/status, decimal/date format |
| E2E | Playwright | User journeys desktop/mobile, auth, accessibility smoke |
| Load/failure | k6 + controlled provider fixture | 20 req/s + 20 webhook/s, retry/restart/queue lag |
| Security | dependency scan, SAST, secret scan, DAST staging | auth, IDOR, webhook, export/attachment signed URL |

## Test cases theo feature

### Frontend

- Money input không mất chữ số khi paste `1.170.000.000`; DTO luôn là decimal string.
- Form submit double click/network retry giữ một idempotency key; toast không nói “đã lưu” khi request fail.
- Route guard, hidden CTA và server 403 đều được test; không dựa vào UI để bảo mật.
- Dashboard hiển thị `asOf`/data quality; loading không làm số cũ trông như số mới.
- Keyboard test: tab order, focus return modal, submit form, close drawer, chart alternative table.
- Visual tests ở 360/768/1024/1440 px, dark mode nếu được hỗ trợ, long Vietnamese label và error state.

### Ledger/domain

- Transfer tạo hai vế cùng transaction, không đổi total cash/net worth.
- Loan principal payment giảm receivable và tăng cash, không phải income.
- Interest payment và fee split tổng bằng transaction amount; accrual rerun không trùng.
- Valuation append-only; as-of cũ không thấy valuation tương lai.
- Reclassify/void không hard delete source và audit có before/after/reason.
- Viewer/editor/owner matrix được test cho từng write endpoint.

### SePay

- HMAC trên raw body, timestamp expiry, malformed payload, duplicate/replay, concurrent duplicate.
- Out auto expense, matched transfer exception, income score threshold, loan payment candidate, ignored/reclassify path.
- Backfill cursor bao phủ gap nhưng không double import; `accumulated=0` là unavailable.
- Revoke connection chặn job mới nhưng vẫn xem được history theo retention policy.

## CI pipeline

1. Format/lint/typecheck, secret scan và dependency vulnerability scan.
2. Unit tests + coverage gate cho module financial/bank-feed.
3. Build Angular production và Go API/worker immutable artifact; SBOM/image scan.
4. Start PostgreSQL ephemeral, chạy migration, integration tests và OpenAPI compatibility check.
5. Playwright smoke (login, create expense, transfer, dashboard) với fixture giả.
6. Publish artifacts; deploy staging tự động; production cần release approval + migration check.

Pull request thay financial logic phải có domain review. Pull request đổi OpenAPI phải cập nhật generated client/mock và contract test cùng lúc.

## Migration và deployment strategy

- Expand/contract: thêm nullable column/index trước; deploy code đọc cả mới/cũ; backfill async; chỉ sau đó mới NOT NULL/drop cũ.
- Index lớn dùng concurrent strategy phù hợp PostgreSQL; không khóa bảng transaction trong giờ hoạt động.
- Migration one-way có backup/restore plan, owner và thời lượng ước lượng.
- API 2 replica rolling; worker version tương thích job payload cũ. Không deploy migration phá compatibility trước worker/API mới.
- Feature flags: `sepay_connect`, `sepay_webhook_ingest`, `sepay_auto_expense`, `sepay_auto_income`, `telegram_read`. Flag server-side, có audit khi owner bật policy workspace.

## Observability và SLO

| Signal | Dashboard/alert |
|---|---|
| API | request rate, 4xx/5xx, p50/p95/p99, auth fail, DB pool wait |
| Database | CPU, storage, connections, slow query, lock wait, replication/PITR health |
| Worker | queue depth/lag, job success/retry/dead letter, lease expiry, job duration |
| Financial | snapshot rebuild delay, reconcile difference, accrual failure/duplicate prevention |
| SePay | signature failure, duplicate ratio, unknown connection, auto-post/reclassify rate, provider 429 |
| Security | abnormal login/rate-limit, export generated/downloaded, secret rotation failure |

SLO initial: API read/write p95 < 400 ms; webhook durable ACK < 2 s; bank-feed lag < 60 s normal; no unacknowledged dead-letter over 15 minutes; RPO ≤ 15 minutes and restore RTO ≤ 4 hours. Alert có runbook link, severity và owner; không alert trên từng expected retry đơn lẻ.

## Security/privacy checklist

- [ ] TLS everywhere, HSTS, secure cookie/JWT rotation, CSRF strategy nếu cookie session.
- [ ] Password/provider token/account content không log; secret injected at runtime, rotated and revoked.
- [ ] Encryption at rest for provider tokens and attachment; signed URL ngắn hạn + authorization trước issue.
- [ ] Tenant isolation test: đổi ID/path/query không đọc được workspace khác (IDOR suite).
- [ ] HMAC raw body + replay window + dedupe for SePay; no unauthenticated production webhook.
- [ ] Audit immutable/restricted write; export/delete/revoke/restructure have actor/reason.
- [ ] Retention/redaction documented for raw bank event and attachment.

## Rollout plan

### Staging

- Mỗi deploy dùng DB/secrets/provider sandbox tách production.
- Seed fixture chứa cash/loan/property/bank events giả; nightly runs migration + E2E + webhook replay suite.

### Private beta

1. Internal workspace: manual transactions and snapshot only.
2. 5 users: asset/loan flow, support feedback, no auto income.
3. 10 users: SePay ingest/review; auto expense behind per-account flag.
4. 25 users: income rules only exact/user-confirmed; daily reconciliation monitoring.
5. 100 users: cohort rollout after 7 ngày không có severity-1, duplicate ledger = 0, queue/SLO đạt và reclassify rate theo ngưỡng product chấp thuận.

### Rollback

- UI/API bug: rollback artifact; DB schema phải compatible.
- Classifier bug: tắt `sepay_auto_income`/`auto_expense` flag, vẫn ingest event vào review inbox.
- Provider incident: pause sync/backfill, không tự tạo missing transaction bằng tay; resume cursor + dedupe khi recovery.
- Data defect: incident record, identify affected source IDs, tạo compensating adjustments qua reviewed tool; không sửa row ledger trực tiếp.

## Runbook tối thiểu

| Sự cố | 5 phút đầu | Khôi phục |
|---|---|---|
| API 5xx tăng | xem deploy/DB pool/log trace; giữ read nếu an toàn | rollback API hoặc scale; không bypass validation |
| Webhook lag | kiểm DB commit, queue/worker health | scale/restart worker; replay jobs from persisted event |
| Double-post nghi ngờ | tắt auto-post flag, query provider IDs/audit | compensation reviewed + root-cause idempotency |
| DB outage | declare incident, stop mutation if inconsistent | managed failover/PITR runbook, verify ledger/snapshot |
| SePay auth lỗi | disable provider calls, keep events quarantined | rotate secret/reconnect after verification |

## Release checklist

- [ ] Changelog, migration, feature flag default và rollback owner.
- [ ] Product acceptance cho happy/error/empty/accessibility states.
- [ ] Test suite green, load result attached nếu thay throughput/queue.
- [ ] Monitoring dashboard/alert/runbook đã tồn tại trước enable flag.
- [ ] Backup recent + restore drill không quá hạn.
- [ ] Support biết cách giải thích source, data quality, auto-post và reclassify.
