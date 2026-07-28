# Thiết kế cơ sở dữ liệu WealthOS

PostgreSQL là nguồn dữ liệu chuẩn. Khóa chính dùng UUID; bảng nghiệp vụ có `user_id`, `created_at`, `updated_at`. Không xóa cứng sổ cái, loan payment hay valuation đã dùng trong báo cáo.

## Bảng cốt lõi

| Nhóm | Bảng | Trường quan trọng |
|---|---|---|
| Tenant | `users`, `user_members` | owner, role |
| Portfolio | `portfolios`, `portfolio_memberships` | base_currency, ownership_share |
| Cash | `accounts`, `transactions`, `transfers` | type, currency, amount, occurred_at, status |
| Loan | `loans`, `loan_schedules`, `loan_accruals`, `loan_payments` | direction, principal, annual_rate, day_count_basis, due_at |
| Property | `properties`, `property_valuations` | acquired_value, valuation_amount, effective_at, source |
| Other assets | `assets`, `asset_valuations` | asset_type, value, effective_at, source |
| Planning | `recurring_rules`, `budgets`, `budget_allocations`, `forecast_scenarios`, `forecast_assumptions` | period, rrule, assumption type |
| Bank feed | `bank_connections`, `bank_feed_events`, `bank_feed_transactions`, `bank_automation_rules`, `bank_reconciliations`, `payment_requests` | provider, consent, external ID, confidence/evidence, auto-post policy, sync/review state, as-of, source |
| Governance | `exchange_rates`, `audit_logs`, `attachments` | effective_at, actor, before/after JSON |

## Bất biến và ràng buộc

- `loans.direction IN ('receivable', 'payable')`; `principal_initial > 0`; `annual_rate >= 0`.
- `loan_payments.principal_amount`, `interest_amount`, `fee_amount`, `waived_amount >= 0`; tổng phần tiền phải bằng transaction liên kết.
- `loan_accruals` unique theo `(loan_id, accrual_date, accrual_method)` để job chạy lại không ghi lãi trùng.
- `property_valuations` và `asset_valuations` append-only; cấm `effective_at` ở tương lai nếu không phải scenario.
- `transactions.amount > 0`; category bắt buộc với income/expense và để trống khi transfer; `valuation_adjustment` không được tạo cash movement.
- Một transfer phải ghép đúng hai transaction cùng currency/amount trong cùng DB transaction.
- `bank_feed_events` unique theo `(provider, external_event_id)` và `bank_feed_transactions` unique theo `(connection_id, provider_transaction_id)`; retry/replay chỉ được tạo một candidate.
- Bank-feed transaction là immutable import; `approve` hoặc policy auto-post mới tạo `transactions`, `transfers` hoặc `loan_payments` qua service ledger và idempotency key liên kết event nguồn.
- `bank_feed_transactions` lưu `classification_type`, `category_id`, `classification_confidence`, `classification_evidence`, `posting_state` và `posted_transaction_id`; auto-post cũng phải liên kết duy nhất đến một ledger record.
- `bank_automation_rules` thuộc user/account, có priority, điều kiện minh bạch, action/type/category và trạng thái enabled; rule không được tự gán loan payment hoặc transfer nếu thiếu reference/match chắc chắn.
- Token/secret của provider được mã hóa; `bank_connections` lưu consent scope, `revoked_at` và không lưu thông tin đăng nhập ngân hàng.
- `payment_requests.payment_code` unique; webhook match chỉ tạo candidate, không tự ghi principal/interest.

## Chỉ mục và snapshot

- `(user_id, occurred_at DESC)` và `(account_id, occurred_at DESC)` cho transactions.
- `(connection_id, occurred_at DESC)` cho bank feed; unique index provider event, index `(user_id, sync_state)` cho connection và `(account_id, as_of_at DESC)` cho reconciliation.
- `(portfolio_id, status)`, `(loan_id, due_at)`, `(loan_id, accrual_date DESC)`, `(property_id, effective_at DESC)`.
- Tạo `portfolio_snapshots` theo ngày cho dashboard nhanh, gồm net worth và attribution. Snapshot là cache có version nguồn; có thể rebuild từ ledger, valuation và tỷ giá.

## Đa tiền tệ

Giá trị gốc luôn lưu kèm currency. `exchange_rates` có cặp tiền, ngày hiệu lực, nguồn và độ chính xác; snapshot lưu cả số nguyên tệ và số đã quy đổi để truy vết lại kết quả lịch sử.
