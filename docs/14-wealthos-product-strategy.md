# Quyết định định vị — WealthOS asset-first

## Quyết định

Sản phẩm được định vị là **WealthOS — Hệ điều hành tài sản cá nhân**. Các màn hình thu–chi và ngân sách vẫn phục vụ người dùng phổ thông, nhưng là năng lực theo dõi thanh khoản, không phải lời hứa giá trị trung tâm.

## Lý do

1. App budget thuần đã là một thị trường đông với UX và bank sync mạnh; benchmark tại tài liệu 13 mô tả các ví dụ đại diện.
2. Người có danh mục cho vay, bất động sản hoặc tài sản thủ công cần một mô hình khác: gốc, lãi phát sinh, lịch thu, định giá và nguồn gốc biến động net worth.
3. Một dashboard asset-first tạo đường nâng cấp tự nhiên: người dùng chỉ ghi thu–chi vẫn có ích, khi có tài sản họ không phải chuyển sang hệ thống khác.

## Biên sản phẩm

| WealthOS làm | WealthOS không hứa |
|---|---|
| Ghi nhận và giải thích portfolio/net worth theo dữ liệu người dùng | Định giá thị trường chính xác hay tư vấn đầu tư |
| Mô phỏng kịch bản từ assumption minh bạch | Dự báo lợi nhuận bảo đảm |
| Theo dõi khoản cho vay/phải thu, lãi, gốc và lịch thu | Xác minh pháp lý, cưỡng chế thu hồi nợ hoặc quản lý tín dụng thay người dùng |
| Theo dõi property/tài sản qua valuation history | Thay thế thẩm định viên hoặc môi giới |
| Trợ lý chat tạo lệnh có audit và approval | Mở quyền điều khiển máy tính không giới hạn từ Telegram |

## Hệ quả thiết kế

- Trang đầu là net worth as-of có breakdown và attribution.
- Khoản cho vay được model là receivable; principal, interest accrued và interest paid không được gộp.
- Forecast luôn gắn kịch bản, assumption, data quality và mốc thời gian.
- Growth rate phải tách tiền nộp/rút bên ngoài khỏi hiệu quả danh mục.
- Bank sync được coi là tích hợp theo thị trường, không phải điều kiện để bắt đầu dùng sản phẩm.
- Assistant Gateway là lớp policy/audit; Hermes chỉ là executor có quyền tối thiểu, thay được bằng executor khác sau này.

## Câu hỏi cần chốt trước khi code

1. Loan cần hỗ trợ loại lãi nào: simple, compound, dư nợ giảm dần, lãi trả trước hay lãi cuối kỳ?
2. Trong net worth, có tính lãi phát sinh chưa thu theo accrual hay chỉ tính khi thực thu? Có cần cấu hình theo portfolio không?
3. Asset/property được định giá bằng người dùng nhập, giá tham chiếu, hay cả hai? Ngưỡng stale bao nhiêu ngày?
4. Chủ hộ kinh doanh có cần tách hoàn toàn portfolio cá nhân và kinh doanh, hay chỉ tag/nhóm tài khoản?
5. Báo cáo growth rate cần chuẩn kế toán đơn giản hay TWR/MWR nâng cao cho nhiều dòng nộp/rút vốn?
