# Test cases trọng yếu WealthOS

| ID | Kịch bản | Kết quả mong đợi |
|---|---|---|
| NW-01 | Cash 160.000.000 + receivable 1.010.000.000, không có nợ | Net worth tại ngày chốt là 1.170.000.000 VND; breakdown rõ cash/receivable |
| NW-02 | Thu 100.000.000 gốc khoản vay vào ngân hàng | Cash tăng, receivable giảm cùng số tiền; net worth không đổi |
| NW-03 | Thu 3.000.000 lãi | Cash và income/lãi đã thu tăng; dư gốc không đổi |
| NW-04 | Cập nhật định giá đất tăng 200.000.000 | Net worth tăng qua valuation attribution; không tạo cash income |
| NW-05 | Nộp thêm 500.000.000 vốn ngoài | Net worth tăng; growth performance không coi 500.000.000 là lợi nhuận |
| LN-01 | Loan 1.010.000.000, lãi suất/cơ sở ngày hợp lệ | Accrual ngày đúng decimal và không bị ghi trùng khi job chạy lại |
| LN-02 | Payment chứa gốc/lãi/phí | Tổng phần tiền bằng transaction; mỗi phần ảnh hưởng đúng aggregate |
| LN-03 | Loan quá hạn | Gắn trạng thái/cảnh báo; không tự xóa nợ hay giảm 100% giá trị |
| AS-01 | Thêm valuation mới rồi xem báo cáo lịch sử | Báo cáo ngày cũ dùng valuation có hiệu lực trước ngày đó, không dùng giá mới hồi tố |
| FC-01 | Scenario cuối năm có loan schedule và chi phí định kỳ | Kết quả kèm assumption, source events và cash floor |
| FC-02 | Scenario thu gốc | Chỉ chuyển cơ cấu cash/receivable; net worth không tăng do riêng sự kiện đó |
| TX-01 | Tạo chi tiêu 65.000 VND | Cash giảm, budget tăng chi, net worth giảm 65.000 |
| TX-02 | Tạo transfer giữa hai account | Tạo hai bút toán nguyên tử; tổng cash và net worth không đổi |
| SEC-01 | Viewer sửa loan/valuation | 403, không thay đổi dữ liệu; audit log không bị giả mạo |
| ASST-01 | Telegram update từ chat chưa liên kết | Gateway không gọi Hermes, trả hướng dẫn liên kết an toàn |
| ASST-02 | User gửi “mở Chrome và vào URL” | Gateway tạo `external_action` và plan; Hermes chưa chạy trước approval |
| ASST-03 | User bấm approve inline button hai lần | Chỉ command đầu được dispatch; approval token còn lại bị từ chối |
| ASST-04 | Hermes gửi event với credential sai | Gateway từ chối event, không đổi trạng thái command |
| ASST-05 | Hermes offline hoặc timeout | Command có trạng thái timeout/failed, Telegram nhận kết quả rõ ràng và audit đầy đủ |
| ASST-06 | User hỏi “tài sản ròng hiện tại?” | Assistant chỉ thực hiện read, trả as-of/data quality và không cần approval |
| PERF-01 | Tải chuỗi net-worth snapshot 3 năm | Phân trang/nén theo SLO; không tính lại full ledger ở mỗi request |

## Chiến lược kiểm thử

- Unit test engine với decimal, loan day-count, external cash flow và fixture định giá theo ngày.
- Integration test transaction DB cho payment, transfer, valuation append-only và phân quyền portfolio.
- Property test: thu gốc cùng currency không làm thay đổi net worth; transfer không đổi tổng cash.
- Golden test cho snapshot, attribution và forecast scenario để phát hiện sai số không chủ ý.
- E2E: tạo portfolio → thêm cash/loan/property → xác nhận dashboard → chạy scenario → drill-down số liệu.
