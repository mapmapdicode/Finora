# Phân tích tính năng mở rộng — WealthOS

## Mục đích

Sau các năng lực lõi về tài sản ròng, khoản phải thu, dòng tiền và kịch bản, các tính năng mới chỉ nên được ưu tiên khi giúp người dùng trả lời nhanh hơn một trong ba câu hỏi:

1. Tài sản ròng thay đổi vì đâu?
2. Khoản tiền nào cần xử lý tiếp theo?
3. Dữ liệu nào đang thiếu hoặc không còn đáng tin?

Tính năng không cải thiện một trong ba câu hỏi trên không nên vào roadmap gần, dù có vẻ hấp dẫn về mặt giao diện hoặc AI.

## Ma trận ưu tiên

| Tính năng | Giá trị cho định vị asset-first | Độ phức tạp/rủi ro | Phụ thuộc | Khuyến nghị |
|---|---|---|---|---|
| Inbox việc cần làm tài chính | Cao | Thấp | Loan, valuation, cash flow | Làm sớm |
| Đối soát snapshot tài sản ròng | Cao | Trung bình | Ledger, valuation history | Làm sớm |
| Hồ sơ đối tác khoản vay | Cao | Trung bình | Loan portfolio, privacy | Làm sau loan lõi |
| Tài liệu/chứng từ cho tài sản | Trung bình–cao | Trung bình | Asset, storage, access control | Làm sau MVP |
| Cảnh báo concentration và liquidity | Cao | Trung bình | Asset classification, forecast | Làm sau data-quality layer |
| Mục tiêu tài sản và tiến độ | Trung bình | Thấp | Net worth, scenario | Làm sau snapshot |
| Chia sẻ gia đình có quyền riêng tư | Cao | Cao | User/RBAC/audit | Thiết kế sớm, triển khai muộn |
| Import/migration từ Excel | Cao | Trung bình | Mapping, dedupe, audit | Ưu tiên cao khi onboarding |
| Kết nối ngân hàng/broker | Trung bình | Cao | Consent, đối soát, đối tác theo thị trường | Theo từng thị trường |
| Gợi ý AI/tư vấn đầu tư | Thấp hoặc rủi ro cao | Cao | Dữ liệu chất lượng, compliance | Chưa làm |

## 1. Inbox tài chính: biến dữ liệu thành việc cần làm

Đây là tính năng nên bổ sung đầu tiên sau loan portfolio và cash flow. Nó không thay thế dashboard; nó chọn ra các sự kiện cần hành động từ dashboard.

### Ví dụ item

| Loại | Quy tắc tạo | Hành động của người dùng |
|---|---|---|
| Khoản thu sắp đến hạn | Kỳ thu gốc/lãi còn 7 ngày | Xem chi tiết, đánh dấu đã liên hệ hoặc ghi nhận thanh toán |
| Quá hạn | Chưa đủ payment sau ngày đến hạn | Ghi nhận hẹn mới, thêm ghi chú, đánh dấu tranh chấp |
| Định giá stale | Quá ngưỡng stale của loại tài sản | Cập nhật valuation hoặc xác nhận giữ nguyên |
| Cash floor risk | Forecast tiền mặt xuống dưới ngưỡng người dùng đặt | Mở kịch bản và xem các khoản tạo ra rủi ro |
| Cần đối soát | Account/asset chưa được xác nhận trong chu kỳ | Nhập số dư hoặc xác nhận số đang hiển thị |

### Nguyên tắc thiết kế

- Inbox chỉ tạo **triage**, không tự đổi trạng thái tài chính.
- Mỗi item phải có lý do, dữ liệu nguồn, thời gian tính và deep link tới bản ghi gốc.
- Dùng severity (`info`, `attention`, `urgent`) thay vì màu cảnh báo tràn lan.
- Không coi quá hạn là vỡ nợ; trạng thái rủi ro là do người dùng hoặc rule minh bạch xác định.

### Dữ liệu tối thiểu

`action_item(id, user_id, type, subject_type, subject_id, severity, due_at, state, rule_version, created_at, resolved_at)`; nội dung hiển thị được tái tạo từ dữ liệu nguồn để không thành một bản sao sự thật thứ hai.

## 2. Đối soát tài sản ròng và data quality

Net worth chỉ hữu ích nếu người dùng biết con số nào đã được kiểm chứng. Bổ sung một luồng đối soát định kỳ sẽ khác biệt hơn việc chỉ thêm nhiều biểu đồ.

### Luồng đề xuất

1. Hệ thống tạo snapshot theo ngày chốt, kèm từng thành phần và nguồn dữ liệu.
2. Người dùng xác nhận, chỉnh số dư hoặc giải thích chênh lệch.
3. Điều chỉnh phải tạo transaction/valuation mới hoặc adjustment có audit; không sửa im lặng số lịch sử.
4. Snapshot nhận trạng thái `draft`, `reviewed` hoặc `locked`.

### Chỉ số chất lượng hiển thị trên dashboard

- Phần trăm giá trị tài sản đã được xác nhận trong chu kỳ.
- Giá trị đang dùng valuation stale.
- Số account chưa đối soát.
- Phần net worth không xác định được nguồn hoặc tỷ giá.

Không dùng một “điểm tin cậy” duy nhất nếu không giải thích được công thức. Các chỉ báo thành phần dễ kiểm tra hơn.

## 3. Hồ sơ khoản vay và đối tác

Khoản phải thu cần bối cảnh để quản lý được thực tế, nhưng dữ liệu này nhạy cảm. Giai đoạn đầu chỉ nên lưu contact tối thiểu và ghi chú có cấu trúc, không xây CRM hoặc chấm điểm tín dụng.

### Phạm vi hợp lý

- Liên kết loan với một `counterparty` có tên hiển thị, kênh liên hệ và nhãn quan hệ.
- Lưu điều khoản, mốc nhắc việc, ghi chú cuộc trao đổi và liên kết chứng từ.
- Có trạng thái vận hành: `active`, `due_soon`, `overdue`, `restructured`, `closed`, `disputed`.
- Tách rõ trạng thái vận hành khỏi nhận định về uy tín hay khả năng trả nợ.

### Không làm

- Không thu thập CCCD, địa chỉ hoặc dữ liệu nhạy cảm nếu không có nhu cầu rõ ràng.
- Không tự gửi tin nhắc người vay ở phiên bản đầu.
- Không suy luận credit score hoặc đề xuất hành động pháp lý.

## 4. Chứng từ và nguồn gốc dữ liệu

Người dùng asset-first thường cần trả lời “số này dựa vào đâu?”. Tính năng chứng từ nên là attachment có kiểm soát, không phải kho lưu trữ tệp tổng quát.

- Cho phép gắn file/link vào loan, asset, valuation, transaction hoặc snapshot.
- Lưu metadata: loại chứng từ, ngày hiệu lực, người tải lên, checksum và retention policy.
- Phân quyền xem/tải riêng; link tải phải ngắn hạn và được audit.
- OCR/search là enhancement sau; bản gốc và metadata vẫn là nguồn chính.

Không đưa giấy tờ định danh, password, OTP hay private key vào sản phẩm. Cảnh báo và chặn upload theo loại tệp/nội dung rủi ro khi khả thi.

## 5. Cảnh báo danh mục: concentration và thanh khoản

Cảnh báo chỉ có ý nghĩa nếu gắn với ngưỡng do người dùng chọn và luôn giải thích được.

| Cảnh báo | Công thức minh bạch | Ví dụ hành động |
|---|---|---|
| Concentration | Giá trị một asset/counterparty / net worth hoặc tổng asset cùng nhóm | Xem exposure, sửa phân loại hoặc điều chỉnh ngưỡng |
| Liquidity | Cash + expected inflows có độ tin cậy cao so với nghĩa vụ trong kỳ | Mở cash forecast và scenario |
| Valuation freshness | Giá trị asset dùng valuation quá ngưỡng | Cập nhật/confirm valuation |
| Currency exposure | Giá trị theo currency so với net worth | Kiểm tra tỷ giá/as-of |

Mặc định nên bắt đầu ở chế độ insight, không dùng ngôn ngữ “bán ngay”, “mua ngay” hoặc dự báo chắc chắn.

## 6. Mục tiêu tài sản và kịch bản

Mục tiêu giúp biến forecast thành kế hoạch nhưng không được làm giả định biến mất. Một goal có: giá trị mục tiêu, ngày mục tiêu, currency, nguồn đóng góp dự kiến và kịch bản được chọn.

Ví dụ: “Duy trì cash floor 200 triệu trong 90 ngày” có giá trị hơn “Tăng tài sản 20%” vì có thể liên kết trực tiếp với cash forecast. Độ lệch phải chỉ ra assumption nào thay đổi: payment chậm, chi phí lớn, valuation hoặc external cash flow.

## 7. Onboarding và import Excel

Đây là ưu tiên trải nghiệm cao vì người dùng mục tiêu thường đã theo dõi bằng bảng tính.

### MVP import

1. Cung cấp template riêng cho account, asset/valuation và loan/payment; không cố đoán mọi file Excel.
2. Cho người dùng map cột, xem preview và chọn cách xử lý duplicate.
3. Chạy validation trước khi ghi: decimal, currency, ngày hiệu lực, tham chiếu account/loan.
4. Ghi một import batch có thể xem, revert theo batch khi chưa bị bản ghi sau phụ thuộc.
5. Tạo audit theo từng row lỗi; không bỏ qua lỗi im lặng.

CSV trước, `.xlsx` sau nếu chi phí parser/UX không tương xứng. Không import trực tiếp thành số dư tổng nếu thiếu provenance; cần biết đó là opening balance hay transaction lịch sử.

## 8. Chia sẻ gia đình và tách cá nhân–kinh doanh

Đây là năng lực có giá trị nhưng là thay đổi trust model, không chỉ là thêm một nút “mời thành viên”. Thiết kế dữ liệu từ sớm với `user`, ownership và audit, nhưng chỉ mở UI sau khi có RBAC rõ ràng.

| Vai trò | Quyền gợi ý |
|---|---|
| Owner | Toàn quyền, export, mời thành viên, thay policy |
| Editor | Tạo/sửa bản ghi theo scope được cấp; không đổi quyền hoặc export toàn bộ |
| Viewer | Chỉ xem scope được cấp, không xem attachment nhạy cảm mặc định |
| Accountant | Xem/đối soát ledger; không xem ghi chú cá nhân nếu không được cấp |

Nên hỗ trợ tag hoặc portfolio `personal`/`business` trước. Tách user độc lập chỉ cần khi yêu cầu riêng tư, sổ sách hoặc quyền sở hữu thật sự khác nhau.

## Thứ tự đề xuất sau roadmap hiện tại

1. Đối soát + data quality ngay khi có Wealth snapshot.
2. Inbox việc cần làm cùng hoặc ngay sau Loan portfolio và Cash flow.
3. Import Excel khi cần onboarding người dùng thật, trước bank sync diện rộng.
4. Hồ sơ đối tác và chứng từ sau khi loan workflow ổn định.
5. Concentration/liquidity alert và goal sau khi forecast có dữ liệu đủ tin cậy.
6. Household sharing sau khi tenant isolation, RBAC và audit được kiểm thử thực tế.

## Quyết định cần chốt trước khi triển khai

1. Chu kỳ đối soát mặc định là tháng, quý hay do từng loại tài sản cấu hình?
2. Ai được xem tên/ghi chú/chứng từ của counterparty trong user chung?
3. Revert import batch có được phép khi batch đã ảnh hưởng snapshot đã `locked` không?
4. “Expected inflow có độ tin cậy cao” được xác định thủ công, theo trạng thái loan, hay cả hai?
5. Ngưỡng stale, cash floor và concentration mặc định theo user hay do từng portfolio/asset class?
