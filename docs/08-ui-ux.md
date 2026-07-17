# UI/UX WealthOS

## Nguyên tắc thông tin

Màn hình đầu tiên trả lời “tôi đang sở hữu bao nhiêu và vì sao thay đổi?”, sau đó mới là “tháng này đã chi bao nhiêu?”. Mọi số tài chính quan trọng có `as of`, đơn vị tiền tệ, trạng thái dữ liệu và đường dẫn drill-down.

## Điều hướng chính

`Tổng quan tài sản` · `Danh mục vốn` · `Khoản vay` · `Tài sản & đất` · `Dòng tiền` · `Ngân sách` · `Dự báo` · `Báo cáo` · `Cài đặt`.

Mobile ưu tiên nút **Thêm giao dịch** và **Cập nhật tài sản**; người dùng phổ thông không bị ép tạo portfolio phức tạp khi chỉ muốn ghi chi tiêu.

## Tổng quan tài sản mặc định

1. Hero: **TÀI SẢN RÒNG** + ngày chốt + tỷ lệ thay đổi theo kỳ; chạm để xem công thức.
2. Cơ cấu: cash, receivables/cho vay, property, other assets, liabilities. Không gộp khoản cho vay vào cash.
3. “Lãi hôm nay”: `accrued`, `đã thu`, `còn phải thu` là ba số liệu riêng.
4. Lịch 30 ngày: khoản thu gốc/lãi, khoản phải trả và cảnh báo cash floor.
5. Attribution: thay đổi do vốn nộp/rút, thu nhập/lãi, chi tiêu, định giá, tỷ giá hay adjustment.
6. Nút “Chạy kịch bản” hiển thị mốc forecast và assumption thay vì một lời tiên đoán.

## Luồng quản lý khoản vay

1. Tạo loan: chiều khoản vay, đối tác, gốc, lãi suất, cơ sở tính ngày, ngày đến hạn và lịch thu.
2. Xem portfolio row: dư gốc, lãi/ngày, lãi chưa thu, kỳ gần nhất/kỳ trễ và mức tập trung vốn.
3. Ghi payment: giao diện bắt buộc xác nhận chia gốc/lãi/phí; preview tác động tới cash, receivable và net worth.
4. Khi tái cơ cấu/xóa nợ, dùng flow riêng có lý do, quyền owner và audit trail.

## Luồng định giá bất động sản/tài sản

Hiển thị giá mua, định giá gần nhất, ngày hiệu lực, nguồn và mức biến động. “Cập nhật giá” tạo một bản ghi mới, không sửa số cũ; có cảnh báo nếu quá lâu chưa định giá.

## Cash flow và budget

Thu–chi và budget vẫn là module nhanh, với nút thêm giao dịch, category, recurring rule và cảnh báo vượt mức. Chúng được thiết kế như lớp quản lý thanh khoản, không thay thế dashboard WealthOS.

## Khả năng tiếp cận

- Không dùng màu xanh/đỏ là tín hiệu duy nhất; trạng thái phải có nhãn và biểu tượng.
- Biểu đồ có bảng dữ liệu thay thế, tooltip bàn phím và định dạng tiền theo locale.
- Mục tiêu thao tác tối thiểu 44×44 px; lỗi biểu mẫu nằm cạnh trường cùng cách sửa.
- Forecast/valuation phải được gắn nhãn `ước tính` hoặc `dữ liệu cũ`, không trình bày như số đã đối soát.
