# Dữ liệu mẫu và dữ liệu người dùng thật

## Fixture WealthOS

Không dùng dữ liệu cá nhân thật ở local, test, staging, screenshot hoặc tài liệu. Fixture dưới đây là giả định để kiểm thử dashboard, không phải danh mục của một người thật.

| Nhóm | Giá trị mẫu | Mục đích |
|---|---:|---|
| Ngân hàng | 160.000.000 VND | Cash/bank balance |
| Dư gốc phải thu đang hoạt động | 1.010.000.000 VND | Loan portfolio và lịch thu gốc/lãi |
| Tài sản ròng tối thiểu từ hai dòng trên | 1.170.000.000 VND (1,17 tỷ VND) | Snapshot/breakdown net worth |
| Chi ăn uống | 65.000 VND | Cash flow/budget |

Không suy diễn lãi/ngày từ fixture nếu thiếu lãi suất, cơ sở tính ngày, ngày bắt đầu và quy tắc compounding của từng khoản vay. Forecast mẫu cũng phải chứa scenario input, không chỉ có số đầu ra.

## Nguyên tắc dữ liệu thật

1. Thu thập tối thiểu; chỉ lấy trường cần cho tính năng mà người dùng bật.
2. Có consent rõ ràng trước khi import, kết nối ngân hàng hoặc chia sẻ workspace; cho phép revoke và nêu tác động của revoke.
3. Mã hóa khi truyền/lưu; nguyên tắc least privilege cho người dùng lẫn nhân sự vận hành.
4. Export có xác thực lại, URL hết hạn và audit log; việc xóa tuân theo chính sách được công bố nhưng không được làm sai audit/đối soát bắt buộc.
5. Dùng bản sao ẩn danh/tổng hợp cho phân tích; không đưa số tài khoản, ghi chú, đối tác khoản vay hoặc token vào log.

## Chất lượng và provenance

- Lưu `source`: `manual`, `csv_import`, `bank_sync`, `valuation_import`; lưu actor và thời điểm cho loan/valuation.
- Preview/mapping trước import; phát hiện trùng, xác nhận trước khi hạch toán.
- Định giá phải có ngày hiệu lực và nguồn; dashboard cảnh báo khi stale.
- Khi nguồn đồng bộ sửa/xóa giao dịch, giữ change record và cho người dùng xem tác động đến cash, receivable và net worth.
