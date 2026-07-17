# WealthOS — Hệ điều hành tài sản cá nhân

## 1. Vấn đề cần giải quyết

Người dùng không chỉ cần biết đã chi bao nhiêu, mà cần biết tại một thời điểm xác định:

- Tổng tài sản, tổng nợ và **tài sản ròng** là bao nhiêu.
- Bao nhiêu vốn đang sinh lời, lãi/ngày, lãi/tháng và lãi đã tích lũy nhưng chưa thu.
- Khoản nào sắp thu gốc, thu lãi; tiền mặt có đủ cho nghĩa vụ và cơ hội đầu tư hay không.
- Tài sản ròng tăng/giảm vì dòng tiền, lãi, định giá hay tỷ giá.
- Nếu giữ một tập giả định rõ ràng, tài sản và tiền mặt sẽ ở đâu vào cuối năm hoặc một mốc tương lai.

Dữ liệu thường rải giữa Zalo, Excel, sổ tay và ứng dụng ngân hàng nên không có nguồn chân lý thống nhất để theo dõi danh mục hay ra quyết định.

## 2. Mục tiêu sản phẩm

### Quản lý tài sản

- Theo dõi tiền mặt, tài khoản ngân hàng, bất động sản và tài sản nhập thủ công.
- Lưu lịch sử định giá, nguồn định giá và ngày hiệu lực; không ghi đè giá trị cũ.

### Quản lý vốn sinh lời

- Theo dõi danh mục khoản cho vay/phải thu: dư gốc, lãi suất, cơ sở tính ngày, lịch thu và trạng thái rủi ro.
- Tính lãi/ngày, lãi kỳ và lãi cộng dồn; tách rõ **lãi đã phát sinh**, **lãi đã thu** và **gốc đã thu**.

### Quản lý dòng tiền

- Ghi nhận thu nhập, chi tiêu, chuyển tiền và các lịch tiền vào/ra.
- Cung cấp ngân sách cho người dùng phổ thông, nhưng không để ngân sách che khuất tài sản ròng.

### Dự báo tài sản

- Dự báo tiền mặt, tài sản ròng và tăng trưởng vốn theo kịch bản có thể xem/sửa.
- Mô phỏng các sự kiện đầu tư, khoản thu gốc/lãi, định giá lại và chi phí lớn.

## 3. Đối tượng sử dụng

| Nhóm | Nhu cầu chính |
|---|---|
| Người quản lý chi tiêu cá nhân | Thu–chi, ngân sách và tiền khả dụng |
| Gia đình | Tài sản chung, dòng tiền chung và phân quyền riêng tư |
| Người cho vay vốn | Dư gốc, lãi, lịch thu và danh mục vốn đang hoạt động |
| Nhà đầu tư đất | Danh mục bất động sản, giá mua, định giá hiện tại, lãi/lỗ |
| Chủ hộ kinh doanh | Dòng tiền cá nhân/chủ hộ và tài sản vận hành |

## 4. Triết lý sản phẩm: bốn cấp độ

| Cấp độ | Câu hỏi người dùng trả lời được | Năng lực WealthOS |
|---|---|---|
| 1. Thu–chi | Tiền đi đâu? | Sổ giao dịch, phân loại, ngân sách |
| 2. Dòng tiền | Tiền đến từ đâu, sắp vào/ra khi nào? | Lịch thu–chi, recurring rule, cash forecast |
| 3. Tài sản | Mình đang sở hữu bao nhiêu? | Net worth, tiền mặt, khoản phải thu, bất động sản, nợ |
| 4. Tăng trưởng | Tài sản tăng với tốc độ nào và vì sao? | Return attribution, lãi tích lũy, định giá và mô phỏng |

Người dùng phổ thông có thể dừng ở cấp 1–2. Sản phẩm chỉ phát huy đầy đủ ở cấp 3–4 khi có tài sản và dữ liệu định giá đáng tin cậy.

## 5. Module cốt lõi

| Module | Giá trị chính |
|---|---|
| Asset Dashboard | Tổng tài sản, nợ phải trả, tài sản ròng, cơ cấu và thay đổi theo thời gian |
| Loan Portfolio | Dư gốc, lãi/ngày, lãi cộng dồn, lịch thu gốc/lãi, trạng thái khoản vay |
| Cash Flow | Thu nhập, chi tiêu, dòng tiền ròng và khoản sắp đến |
| Budget | Hạn mức tháng, tiến độ và cảnh báo vượt mức |
| Property Management | Giá mua, giá trị hiện tại, lịch sử định giá, lãi/lỗ chưa thực hiện |
| Forecast Engine | Tiền mặt, tài sản ròng và kịch bản tăng trưởng có thể giải thích |

## 6. Dashboard mặc định

Trang đầu tiên là **TÀI SẢN RÒNG**, không phải “Chi tiêu tháng này”. Dashboard phải luôn có ngày chốt số liệu và nút xem công thức.

```text
TÀI SẢN RÒNG (as of 17/07/2026)        1,17 tỷ VND

Khoản phải thu / cho vay                 1,01 tỷ VND
Ngân hàng                                  160 triệu VND
Bất động sản                         [theo định giá gần nhất]
Nợ phải trả                          [nếu có]

Lãi phát sinh hôm nay                [tính từ từng khoản vay]
Dòng tiền tháng                      [thu - chi - trả nợ]
```

`1,17 tỷ VND` là cách biểu thị nhất quán của 1.170.000.000 VND; hệ thống không dùng cách ghi mơ hồ “1.170 tỷ”. Lãi/ngày chỉ hiển thị sau khi từng khoản vay có dư gốc, lãi suất và cơ sở tính ngày hợp lệ.

## 7. Thước đo thành công

### Người dùng phổ thông

- Ghi giao dịch đều đặn, xem dòng tiền và sử dụng ngân sách.

### Người có tài sản

- Cập nhật/đối soát tài sản ròng định kỳ.
- Theo dõi danh mục vốn, lịch thu và tỷ lệ vốn sinh lời.
- Hiểu tăng trưởng tài sản đến từ lãi, dòng tiền mới hay biến động định giá.

Hai chỉ số sản phẩm trung tâm là **Net Worth** và **Net Worth Growth Rate**; định nghĩa và cách tránh hiểu sai được quy định tại [06-calculation-engine.md](06-calculation-engine.md).

Tham khảo thị trường thu–chi/budgeting: [13-nghien-cuu-thi-truong-quan-ly-thu-nhap-chi-tieu-toan-cau.md](13-nghien-cuu-thi-truong-quan-ly-thu-nhap-chi-tieu-toan-cau.md). Quyết định định vị WealthOS: [14-wealthos-product-strategy.md](14-wealthos-product-strategy.md).
