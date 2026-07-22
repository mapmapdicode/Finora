# Roadmap WealthOS

Roadmap là **asset-first**: đưa người dùng tới tài sản ròng đúng và giải thích được, sau đó mới tăng tự động hóa. Thu–chi/budget vẫn có từ đầu để theo dõi thanh khoản, nhưng không định hình thứ tự phát triển.

| Giai đoạn | Phạm vi | Tiêu chí hoàn thành |
|---|---|---|
| 0. Foundation | Auth, workspace, RBAC nền tảng, audit, portfolio, account, migration, decimal/multi-currency | Tenant isolation, backup, quyền và ledger tái tạo được |
| 1. Wealth snapshot | Cash/bank, tài sản thủ công, net worth as-of, valuation history, đối soát snapshot và data-quality indicators | Net worth có breakdown, nguồn dữ liệu, trạng thái stale/drill-down; snapshot có `draft`/`reviewed`/`locked` và điều chỉnh được audit |
| 2. Loan portfolio | Khoản phải thu/phải trả, lịch thu, accrual, payment split, overdue, counterparty tối thiểu và liên kết chứng từ | Dư gốc/lãi đúng, payment không làm sai cash hay net worth; trạng thái loan không bị nhầm với credit score hay nhận định pháp lý |
| 3. Cash flow | Thu/chi/chuyển tiền, recurring, import CSV theo batch, lịch nghĩa vụ, financial inbox | Cash forecast ngắn hạn khớp ledger và lịch loan; import có preview/validation/audit, inbox truy vết được về dữ liệu nguồn |
| 4. Property & growth | Property, valuation, attribution, growth rate, báo cáo portfolio, mục tiêu tài sản | Tách external cash flow khỏi hiệu quả danh mục; định giá có lịch sử; độ lệch mục tiêu giải thích được bằng dữ liệu/kịch bản |
| 5. Scenario engine | Kịch bản 30/90 ngày/cuối năm, cash floor, concentration/liquidity/currency insight | Mỗi forecast và alert truy vết được assumption, sự kiện, ngưỡng và thời điểm tính; không đưa khuyến nghị đầu tư mệnh lệnh |
| 6. Budget & integrations | Budget, household sharing theo RBAC, bank sync theo thị trường; SePay Bank Hub/OAuth/Webhook là integration Việt Nam đầu tiên; VietQR payment request cho loan; auto-post tiền ra và classifier tiền vào theo rule/confidence | Consent/revoke, xử lý trùng, không phá vỡ sổ cái; tiền ra tự ghi chi trừ ngoại lệ transfer/loan/investment, tiền vào chỉ tự ghi thu khi đủ evidence; thành viên chỉ thấy dữ liệu/attachment được cấp quyền |
| 7. Assistant Gateway | Telegram chat, command/audit/approval, Hermes executor private | Chỉ user đã liên kết chat được dùng; write/external action cần approval; executor không public |

## Không làm sớm

- Tư vấn đầu tư cá nhân hóa, chấm điểm tín dụng hoặc hứa hẹn lợi nhuận.
- Đồng bộ ngân hàng “toàn cầu” khi chưa có đối tác, consent và vận hành đối soát.
- Dịch vụ đàm phán/hủy subscription thay người dùng.
- Đưa một giá trị forecast không có kịch bản, ngày chốt và giả định nguồn.
- Cho Telegram hoặc Hermes gọi thẳng database, nhận token rộng quyền hoặc điều khiển phần mềm mà không có policy/approval/audit.

## Phân biệt với thị trường hiện có

Benchmark tại [13-nghien-cuu-thi-truong-quan-ly-thu-nhap-chi-tieu-toan-cau.md](13-nghien-cuu-thi-truong-quan-ly-thu-nhap-chi-tieu-toan-cau.md) cho thấy phần lớn sản phẩm nổi bật ở budgeting, bank sync hoặc subscription. WealthOS dùng các năng lực đó để làm rõ **thanh khoản**, còn khác biệt chính là portfolio loan/property, net-worth attribution và forecast tài sản có thể giải thích.

Các tính năng mở rộng được đánh giá theo mức đóng góp vào tính đúng, khả năng hành động và chất lượng dữ liệu của net worth tại [16-phan-tich-tinh-nang-mo-rong.md](16-phan-tich-tinh-nang-mo-rong.md).

Thiết kế tích hợp SePay cho bank feed, đối soát và VietQR: [17-sepay-bank-integration.md](17-sepay-bank-integration.md).

Kiến trúc vận hành mục tiêu cho 100 người dùng: [18-architecture-100-users.md](18-architecture-100-users.md).
