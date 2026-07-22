# Quy tắc nghiệp vụ WealthOS

## Tài sản ròng và định giá

1. Tài sản ròng luôn được tính **tại một thời điểm** và theo một tiền tệ báo cáo.
2. Khoản cho vay còn gốc là `receivable` (tài sản), không phải tiền mặt; thu gốc chỉ chuyển giá trị từ receivable sang cash.
3. Lãi phát sinh được tách thành `accrued_interest`; lãi chỉ làm tăng tiền mặt khi có `interest_payment` đã posted. Cài đặt báo cáo xác định lãi phát sinh có được tính vào net worth theo accrual hay không.
4. Bất động sản/tài sản thủ công dùng định giá mới nhất trước thời điểm báo cáo; giá mua, giá trị hiện tại và chi phí liên quan là ba số liệu khác nhau.
5. Định giá quá hạn ngưỡng cấu hình phải có nhãn `stale`; hệ thống không tự coi giá cũ là giá thị trường hiện tại.
6. Vay phải trả được trừ khỏi tài sản ròng; không được bù trừ với khoản cho vay/phải thu chỉ vì cùng một đối tác.

## Khoản vay và lãi

1. Một loan phải có chiều `receivable` hoặc `payable`, số gốc ban đầu, tiền tệ, ngày bắt đầu, cơ sở tính ngày và quy tắc lãi.
2. Thu lãi **không** làm giảm dư gốc. Thu gốc làm giảm dư gốc; tổng hai phần bằng số tiền nhận thực tế.
3. Lãi ngày dùng dư gốc đầu ngày, trừ khi hợp đồng quy định khác; không tính lãi chồng lãi nếu `interest_compounding` không cho phép.
4. Mỗi payment phải tham chiếu loan và phân tách `principal_amount`, `interest_amount`, `fee_amount`, `waived_amount`.
5. Không cho phép thu gốc làm dư gốc âm, trừ giao dịch điều chỉnh/tất toán được owner phê duyệt và có audit log.
6. Khoản trễ lịch không tự động trở thành mất vốn. Chỉ trạng thái `written_off` có lý do và quyền phù hợp mới giảm giá trị phải thu theo chính sách kế toán đã chọn.

## Giao dịch và dòng tiền

1. Giao dịch có loại `income`, `expense`, `transfer`, `investment_funding`, `loan_disbursement`, `loan_payment`, `valuation_adjustment`.
2. Chuyển tiền giữa hai account và thu gốc khoản cho vay không được tính là thu nhập/chi tiêu; phải phân loại để dashboard attribution đúng.
3. Chỉ `posted` đi vào actual; `pending` chỉ phục vụ forecast; `voided` giữ lịch sử và lý do.
4. Không xóa cứng giao dịch tài chính. Mọi chỉnh sửa sau đối soát là adjustment mới liên kết giao dịch gốc.
5. Giao dịch ngân hàng từ provider có `transferType = out` được tự ghi `expense` và `posted` sau khi qua dedupe, trừ khi có bằng chứng nó là transfer nội bộ, giải ngân loan hoặc cấp vốn đầu tư; các ngoại lệ này phải được hạch toán theo loại nghiệp vụ đúng.
6. Giao dịch `in` chỉ tự ghi `income` khi rule/matching có độ tin cậy đủ cao. Ưu tiên nhận diện transfer nội bộ và loan payment trước; tiền vào không đủ bằng chứng phải ở `pending_review`, không mặc định là thu nhập.
7. Transaction tự ghi từ bank feed phải giữ `source = bank_feed`, provider transaction ID, confidence và evidence. Người dùng có thể sửa category/type; hệ thống tạo adjustment/audit, không sửa hay xóa raw import.

## Ngân sách và quyền

1. Budget áp dụng cho expense theo danh mục/kỳ; không áp dụng cho định giá, chuyển tiền hoặc thu gốc.
2. Vượt ngân sách tạo cảnh báo, không chặn giao dịch hợp lệ.
3. Dữ liệu bị cô lập theo workspace; `owner`, `editor`, `viewer` có quyền rõ ràng. Owner chịu trách nhiệm xóa workspace, xuất dữ liệu và phê duyệt điều chỉnh nhạy cảm.
4. Thao tác thay đổi giá trị tài sản, điều khoản loan, định giá hoặc lịch sử hạch toán phải có audit log trước/sau.

## Trợ lý Telegram và Hermes Agent

1. Telegram là kênh nhận lệnh, không phải nguồn xác thực duy nhất. Mỗi `telegram_chat_id` phải được liên kết và xác nhận với một User/Workspace trước khi thực thi.
2. Lệnh từ chat được phân loại `read`, `draft`, `write`, `external_action`. `read` chỉ đọc; `draft` tạo bản nháp; `write` thay đổi dữ liệu WealthOS; `external_action` điều khiển phần mềm trên Mac Mini hoặc dịch vụ bên ngoài.
3. `write` và `external_action` cần confirmation một lần bằng inline button có `approval_id` do server tạo, hết hạn nhanh và chỉ dùng một lần. Không chấp nhận “OK” dạng văn bản như bằng chứng phê duyệt.
4. Lệnh không được chuyển nguyên văn từ Telegram sang Hermes. Gateway chuyển thành action có schema, policy và quyền hạn đã kiểm tra.
5. Hermes không được có quyền ghi trực tiếp PostgreSQL hoặc truy cập secrets của WealthOS. Nó chỉ nhận action đã được phép và trả execution event.
6. Mọi lệnh ghi/điều khiển phải lưu request, actor, mục tiêu, policy decision, approval, executor, kết quả và correlation ID trong audit log.
