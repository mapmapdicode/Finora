# Roadmap WealthOS

Roadmap là **asset-first**: đưa người dùng tới tài sản ròng đúng và giải thích được, sau đó mới tăng tự động hóa. Thu–chi/budget vẫn có từ đầu để theo dõi thanh khoản, nhưng không định hình thứ tự phát triển.

| Giai đoạn | Phạm vi | Tiêu chí hoàn thành |
|---|---|---|
| 0. Foundation | Auth, workspace, audit, portfolio, account, migration, decimal/multi-currency | Tenant isolation, backup, quyền và ledger tái tạo được |
| 1. Wealth snapshot | Cash/bank, tài sản thủ công, net worth as-of, valuation history | Net worth có breakdown, nguồn dữ liệu, trạng thái stale và drill-down |
| 2. Loan portfolio | Khoản phải thu/phải trả, lịch thu, accrual, payment split, overdue | Dư gốc/lãi đúng, payment không làm sai cash hay net worth |
| 3. Cash flow | Thu/chi/chuyển tiền, recurring, import CSV, lịch nghĩa vụ | Cash forecast ngắn hạn khớp ledger và lịch loan |
| 4. Property & growth | Property, valuation, attribution, growth rate, báo cáo portfolio | Tách external cash flow khỏi hiệu quả danh mục; định giá có lịch sử |
| 5. Scenario engine | Kịch bản 30/90 ngày/cuối năm, cash floor, concentration alert | Mỗi forecast truy vết được assumption và sự kiện tạo nên nó |
| 6. Budget & integrations | Budget, household sharing, bank sync theo thị trường | Consent/revoke, xử lý trùng, không phá vỡ sổ cái |
| 7. Assistant Gateway | Telegram chat, command/audit/approval, Hermes executor private | Chỉ user đã liên kết chat được dùng; write/external action cần approval; executor không public |

## Không làm sớm

- Tư vấn đầu tư cá nhân hóa, chấm điểm tín dụng hoặc hứa hẹn lợi nhuận.
- Đồng bộ ngân hàng “toàn cầu” khi chưa có đối tác, consent và vận hành đối soát.
- Dịch vụ đàm phán/hủy subscription thay người dùng.
- Đưa một giá trị forecast không có kịch bản, ngày chốt và giả định nguồn.
- Cho Telegram hoặc Hermes gọi thẳng database, nhận token rộng quyền hoặc điều khiển phần mềm mà không có policy/approval/audit.

## Phân biệt với thị trường hiện có

Benchmark tại [13-nghien-cuu-thi-truong-quan-ly-thu-nhap-chi-tieu-toan-cau.md](13-nghien-cuu-thi-truong-quan-ly-thu-nhap-chi-tieu-toan-cau.md) cho thấy phần lớn sản phẩm nổi bật ở budgeting, bank sync hoặc subscription. WealthOS dùng các năng lực đó để làm rõ **thanh khoản**, còn khác biệt chính là portfolio loan/property, net-worth attribution và forecast tài sản có thể giải thích.
