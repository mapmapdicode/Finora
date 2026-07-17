# Engine tính toán tài sản

## Nguyên tắc

Engine đọc dữ liệu đã hạch toán/định giá và tạo kết quả dẫn xuất. Giao dịch, lịch sử định giá và bút toán lãi là nguồn chân lý; cache phải tái tạo được. Dùng decimal cố định theo tiền tệ, không dùng `float`.

Mọi kết quả hiển thị phải mang `as_of_at`, tiền tệ báo cáo, nguồn tỷ giá và trạng thái dữ liệu (`reconciled`, `estimated`, `stale`).

## Công thức chính

| Đại lượng | Công thức | Ghi chú |
|---|---|---|
| Số dư tài khoản | `số dư đầu kỳ + income - expense ± transfer` | Chỉ `posted` được tính vào actual |
| Dư gốc khoản cho vay | `gốc ban đầu - tổng thu gốc hợp lệ` | Không giảm bởi thu lãi |
| Lãi phát sinh ngày | `dư gốc đầu ngày × lãi suất năm / day_count_basis` | 360, 365 hoặc actual/actual phải là điều khoản của loan |
| Lãi cộng dồn chưa thu | `tổng accrual - tổng interest payment` | Không cộng vào tiền mặt trước khi thực thu |
| Giá trị bất động sản | `định giá gần nhất có effective_at ≤ as_of_at` | Hiển thị tuổi dữ liệu/nguồn định giá |
| Tài sản ròng | `cash + receivables + property + other_assets - liabilities` | Tách lãi đã phát sinh khi người dùng chọn phương pháp accrual |
| Dòng tiền ròng | `thu tiền mặt - chi tiền mặt - giải ngân/rút vốn` | Không nhầm với biến động định giá |

## Net Worth Growth Rate

Không dùng đơn giản `(NW cuối - NW đầu) / NW đầu` nếu người dùng vừa nộp thêm vốn, rút vốn hoặc nhận tài sản. WealthOS phải hiển thị cả hai:

- **Thay đổi tài sản ròng:** `NW_end - NW_start` — câu trả lời cho “giàu hơn bao nhiêu?”.
- **Tăng trưởng từ hiệu quả danh mục:** `((NW_end - net_external_cash_flow) / NW_start) - 1` — ước tính đơn giản cho kỳ khi external cash flow được phân loại rõ.

Với nhiều dòng vốn trong kỳ, dùng tiền-weighted return hoặc time-weighted return và hiển thị rõ phương pháp; không so sánh hai loại tỷ suất trên cùng biểu đồ mà không gắn nhãn.

## Attribution: vì sao tài sản thay đổi

`Δ Net Worth = external cash flow + lãi/thu nhập + chi tiêu/chi phí + thay đổi định giá + thay đổi tỷ giá + điều chỉnh`.

Dashboard phải cho drill-down theo các thành phần trên. Ví dụ, thu gốc từ khoản cho vay làm giảm `receivable` và tăng `cash`, nên **không làm thay đổi tài sản ròng**; nó chỉ thay đổi cơ cấu thanh khoản.

## Quy tắc thời gian, làm tròn và đối soát

- Tính báo cáo theo timezone workspace; lưu thời điểm gốc UTC và `effective_at` cho valuation.
- Làm tròn ở biên hiển thị, không làm tròn sau từng phép cộng trung gian.
- Khi điều chỉnh quá khứ, tính lại aggregate/forecast từ ngày hiệu lực.
- Nếu số dư cache lệch sổ cái hoặc valuation đã quá hạn, dashboard phải gắn nhãn dữ liệu cần đối soát, không trình bày như số chính xác.
