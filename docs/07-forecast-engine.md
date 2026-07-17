# Engine dự báo WealthOS

## Mục tiêu

Engine trả lời ở các mốc 7/30/90 ngày, cuối năm và mốc người dùng chọn:

- Tiền mặt khả dụng có xuống dưới ngưỡng an toàn không?
- Gốc/lãi dự kiến thu hay trả là bao nhiêu?
- Tài sản ròng dự báo thay đổi ra sao và do giả định nào?
- Một sự kiện như giải ngân khoản vay, bán đất, nộp vốn hoặc chi lớn sẽ tác động thế nào?

Kết quả là mô phỏng có điều kiện, **không phải** tư vấn đầu tư, định giá thị trường hay cam kết lợi nhuận.

## Dữ liệu đầu vào

- Portfolio snapshot đã đối soát tại ngày bắt đầu.
- Loan schedule, accrual policy, lãi/gốc dự kiến thu–trả, maturity và trạng thái trễ hạn.
- Giao dịch pending, recurring rule, thu nhập/chi phí lịch sử; loại bỏ outlier do người dùng quyết định.
- Valuation hiện tại và giả định tăng/giảm giá cho từng asset class.
- Scenario event: external cash flow, giải ngân, tất toán, thay đổi lãi suất, định giá lại, chi phí lớn.

## Mô hình phiên bản đầu

1. Chốt baseline `net_worth`, `cash` và cơ cấu tài sản tại `as_of_at`.
2. Sinh các sự kiện xác định từ lịch: thu lãi/gốc, kỳ trả nợ, recurring cash flow.
3. Tính lãi tương lai theo điều khoản loan; khoản `overdue` dùng giả định thu hồi do scenario chỉ định, không mặc định đủ 100%.
4. Với dòng tiền biến đổi, dùng trung vị của các kỳ đầy đủ gần nhất; không có dữ liệu thì để trống thay vì bịa con số.
5. Áp các assumption người dùng có thể sửa: tỷ lệ định giá, ngày thu, chi phí, tỷ giá, lãi suất.
6. Xuất ba kịch bản `thận trọng`, `cơ sở`, `lạc quan`, mỗi kịch bản có ID, timestamp và danh sách assumption.

## Công thức báo cáo forecast

`NW_t = NW_0 + external_cash_flow_t + realized_income_t - expense_t + valuation_change_t + FX_change_t - liability_change_t`.

Chuyển gốc giữa cash và receivable không làm thay đổi `NW_t`; chỉ tác động thanh khoản. Lãi dự báo chỉ tăng net worth nếu scenario chọn accrual accounting và khoản phải thu chưa bị áp haircut rủi ro.

## Điều kiện cảnh báo

- Cash dự báo thấp hơn ngưỡng thanh khoản do người dùng đặt.
- Loan đến hạn/trễ hạn; tỷ trọng một đối tác hoặc một asset class vượt giới hạn người dùng đặt.
- Valuation đã cũ hoặc forecast phụ thuộc assumption chưa được xác nhận.
- Kế hoạch chi/giải ngân làm giảm cash nhưng chưa chỉ rõ nguồn bù đắp.

Mỗi cảnh báo phải mở được danh sách sự kiện và assumption gây ra nó, có nút chỉnh sửa. Ví dụ “31/12/2026: 1,68 tỷ VND” chỉ được hiển thị cùng scenario và các dòng tiền tạo nên giá trị đó.
