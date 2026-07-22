# Kế hoạch phát triển WealthOS

Thư mục này chuyển các quyết định sản phẩm hiện có thành kế hoạch để đội frontend và backend có thể xây dựng, kiểm thử và phát hành từng phần mà không làm sai sổ cái.

## Thứ tự đọc

1. [00-delivery-plan.md](00-delivery-plan.md) — phạm vi, milestone, thứ tự dependency và Definition of Done.
2. [01-frontend-ui-ux.md](01-frontend-ui-ux.md) — cấu trúc Angular, layout, màn hình, state và interaction.
3. [02-backend-domain.md](02-backend-domain.md) — module Go, database write path, API và job.
4. [03-sepay-bank-feed.md](03-sepay-bank-feed.md) — kết nối bank feed, auto-post tiền ra và phân tích tiền vào.
5. [04-quality-release.md](04-quality-release.md) — test matrix, CI/CD, observability, security và rollout.

## Nguyên tắc không đổi

- PostgreSQL ledger và audit log là nguồn chân lý; UI, snapshot và search chỉ là projection/cache.
- Mỗi mutation có authorization, idempotency, audit và phản hồi lỗi có thể xử lý.
- Mọi số tài chính phải hiển thị `asOf`, currency, provenance/data quality và drill-down.
- Tính năng tự động từ bank feed được phép nhanh, nhưng không được làm transfer, loan payment hoặc valuation thành income/expense sai.
- Từng milestone phải deploy độc lập, feature-flag được và có rollback an toàn.

Các quyết định product/domain gốc vẫn nằm ở `../01` đến `../18`; tài liệu trong thư mục này không thay thế các invariant đó.
