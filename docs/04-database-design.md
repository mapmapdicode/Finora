# Thiết kế cơ sở dữ liệu WealthOS

PostgreSQL là nguồn dữ liệu chuẩn. Khóa chính dùng UUID; bảng nghiệp vụ có `workspace_id`, `created_at`, `updated_at`. Không xóa cứng sổ cái, loan payment hay valuation đã dùng trong báo cáo.

## Bảng cốt lõi

| Nhóm | Bảng | Trường quan trọng |
|---|---|---|
| Tenant | `workspaces`, `workspace_members` | owner, role |
| Portfolio | `portfolios`, `portfolio_memberships` | base_currency, ownership_share |
| Cash | `accounts`, `transactions`, `transfers` | type, currency, amount, occurred_at, status |
| Loan | `loans`, `loan_schedules`, `loan_accruals`, `loan_payments` | direction, principal, annual_rate, day_count_basis, due_at |
| Property | `properties`, `property_valuations` | acquired_value, valuation_amount, effective_at, source |
| Other assets | `assets`, `asset_valuations` | asset_type, value, effective_at, source |
| Planning | `recurring_rules`, `budgets`, `budget_allocations`, `forecast_scenarios`, `forecast_assumptions` | period, rrule, assumption type |
| Governance | `exchange_rates`, `audit_logs`, `attachments` | effective_at, actor, before/after JSON |

## Bất biến và ràng buộc

- `loans.direction IN ('receivable', 'payable')`; `principal_initial > 0`; `annual_rate >= 0`.
- `loan_payments.principal_amount`, `interest_amount`, `fee_amount`, `waived_amount >= 0`; tổng phần tiền phải bằng transaction liên kết.
- `loan_accruals` unique theo `(loan_id, accrual_date, accrual_method)` để job chạy lại không ghi lãi trùng.
- `property_valuations` và `asset_valuations` append-only; cấm `effective_at` ở tương lai nếu không phải scenario.
- `transactions.amount > 0`; category bắt buộc với income/expense và để trống khi transfer; `valuation_adjustment` không được tạo cash movement.
- Một transfer phải ghép đúng hai transaction cùng currency/amount trong cùng DB transaction.

## Chỉ mục và snapshot

- `(workspace_id, occurred_at DESC)` và `(account_id, occurred_at DESC)` cho transactions.
- `(portfolio_id, status)`, `(loan_id, due_at)`, `(loan_id, accrual_date DESC)`, `(property_id, effective_at DESC)`.
- Tạo `portfolio_snapshots` theo ngày cho dashboard nhanh, gồm net worth và attribution. Snapshot là cache có version nguồn; có thể rebuild từ ledger, valuation và tỷ giá.

## Đa tiền tệ

Giá trị gốc luôn lưu kèm currency. `exchange_rates` có cặp tiền, ngày hiệu lực, nguồn và độ chính xác; snapshot lưu cả số nguyên tệ và số đã quy đổi để truy vết lại kết quả lịch sử.
