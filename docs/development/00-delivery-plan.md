# 00 — Kế hoạch giao hàng theo milestone

## Mục tiêu delivery

Xây dựng một ứng dụng Angular + Go API mà người dùng có thể bắt đầu bằng cash/thu–chi, sau đó quản lý tài sản, loan, forecast và bank feed SePay mà không cần thay hệ thống. Mỗi milestone cho ra một lát dọc hoàn chỉnh: UI, API, DB migration, audit, test và dashboard vận hành.

## Phạm vi release đầu

| Có trong release | Chưa có |
|---|---|
| Workspace cá nhân, account, transaction, net-worth snapshot | Household sharing UI, bank sync diện rộng |
| Loan receivable/payable, accrual, payment split | Tư vấn đầu tư/credit score |
| Asset/property valuation append-only | Giá thị trường tự động không có nguồn xác thực |
| SePay connect, webhook, review/auto-post có guardrail | Lệnh chuyển tiền ngân hàng |
| Telegram read-only sau cùng | Hermes external action trong lần phát hành đầu |

## Backlog theo lát dọc

### M0 — Khung kỹ thuật và bảo mật

**Frontend**

- Angular workspace, strict TypeScript, ESLint/Prettier, route shell, design tokens, i18n `vi-VN` và currency formatter.
- Auth guard, HTTP interceptor gắn correlation ID, error boundary/toast, loading/skeleton và empty state chuẩn.
- App layout desktop/mobile, navigation skeleton và trang 403/404/maintenance.

**Backend**

- Go API/worker, PostgreSQL migration, config/secrets, health/readiness, OpenAPI generation.
- User/workspace/member/RBAC, JWT/session, audit middleware, request ID và rate limit login.
- `Idempotency-Key` store cho mọi POST mutation; standard error envelope; seed fixtures giả.

**Done khi**

- User đăng nhập, vào đúng workspace, viewer không thể gọi mutation qua API dù tự sửa UI.
- Migration chạy từ DB trống; smoke test `/healthz`, `/readyz`, auth và audit qua CI.
- UI đạt keyboard navigation cơ bản và có layout mobile 360 px.

### M1 — Cash flow và net-worth snapshot

**Frontend**: account list, transaction list/filter, form thêm thu/chi/transfer, dashboard net worth và drill-down.

**Backend**: account/transaction/transfer ledger, category, portfolio snapshot rebuild job và budget read model tối thiểu.

**Done khi**: transaction posted hiện trên cash flow; transfer không đổi tổng cash; dashboard trả `asOfAt`, attribution và data-quality; UI không hiển thị snapshot cũ như realtime.

### M2 — Asset, property và đối soát

**Frontend**: asset/property list, detail timeline, valuation form/preview, reconciliation inbox và trạng thái stale.

**Backend**: append-only valuation, attachment metadata, portfolio snapshot theo ngày, review/lock snapshot, data-quality projection.

**Done khi**: valuation mới không sửa lịch sử; user thấy nguồn và ngày hiệu lực; adjustment có audit/truy vết.

### M3 — Loan portfolio

**Frontend**: wizard tạo loan, loan detail, schedule, payment split preview, overdue inbox.

**Backend**: loan terms/schedule/accrual worker/payment service, counterparty tối thiểu, payment/loan invariant.

**Done khi**: thu gốc không đổi net worth, thu lãi tăng cash/income đúng; accrual job chạy lại không nhân đôi.

### M4 — Forecast, goals và alerts

**Frontend**: scenario editor, assumption drawer, forecast chart/table, goal progress, alert explanation.

**Backend**: forecast engine async, versioned assumptions, cache result, concentration/liquidity calculation.

**Done khi**: mỗi output drill-down được nguồn event/assumption; không có UI gọi kết quả là “đảm bảo”.

### M5 — SePay bank feed

**Frontend**: connect/revoke bank, sync health, imported transaction inbox, rule builder, review/reclassify, VietQR request.

**Backend**: adapter, webhook ingress, raw event/idempotency queue, matcher/classifier, auto-post policy và reconciliation.

**Done khi**: retry webhook không double-post; tiền ra auto expense trừ ngoại lệ; tiền vào chỉ auto income khi evidence đủ; có canary/kill switch.

### M6 — Hardening và private beta

- Load/failure/restore drill; accessibility review; privacy review; monitoring SLO; data export; support runbook.
- Mời 5–10 workspace test, sau đó 25 và 100 users theo cohort; không mở bank auto-post đại trà trước khi reclassify rate chấp nhận được.

## Dependency và ownership

| Hạng mục | Bắt đầu sau | FE đầu mối | BE đầu mối |
|---|---|---|---|
| Dashboard | Auth + account read model | Dashboard feature | Portfolio/snapshot |
| Transaction form | Category/account API | Cash-flow feature | Ledger |
| Loan payment | Loan read model + transaction service | Loan feature | Loan + ledger |
| Valuation | Asset/property detail | Asset feature | Portfolio |
| Forecast | Snapshot/loan schedule ổn định | Forecast feature | Planning engine |
| SePay | Account/ledger/idempotency/audit | Integration feature | Bank-feed adapter |

Frontend có thể dùng MSW/mock contract trước khi endpoint hoàn chỉnh. Không mock công thức ledger ở client: mọi preview có nhãn “ước tính”; server trả kết quả chính thức sau submit.

## Nhịp triển khai cho mỗi user story

1. Chốt acceptance criteria, permission, error states và analytics event.
2. Viết API/schema migration + unit/integration test trước mutation.
3. Cập nhật OpenAPI; FE tạo typed client và mock success/error/empty/loading.
4. Xây UI desktop/mobile/accessibility cùng với backend contract.
5. E2E happy path + boundary case; feature flag staging → canary → production.
6. Review metric trong 48 giờ; rollback flag hoặc migration-compatible fix khi cần.

## Definition of Done dùng chung

- [ ] UX states: loading, empty, error, forbidden, offline/retry và success feedback.
- [ ] Mobile + desktop + keyboard + screen-reader label cho action chính.
- [ ] Backend auth/RBAC, validation, idempotency, audit, trace ID và rate limit phù hợp.
- [ ] Migration forward-compatible, index cần thiết, rollback/compensation được mô tả.
- [ ] Unit + integration + E2E cập nhật; fixture không chứa dữ liệu thật.
- [ ] Metric/log/alert và runbook cho asynchronous job/provider integration.
- [ ] Product/engineering review xác nhận không vi phạm `02-business-rules.md`.
