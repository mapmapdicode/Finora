# Finora Import v1

<!--
HƯỚNG DẪN
1. Chọn tháng trên màn Import. Tất cả ngày ở Thu chi, Khoản vay và Thanh toán
   khoản vay phải thuộc tháng đó, ví dụ 2026-08.
2. Giữ nguyên dòng tiêu đề và dòng |---| của mỗi bảng.
3. Mã là duy nhất trong từng phần.
4. Tiền VND chấp nhận: 5000000, 5.000.000, 5tr hoặc 500k. Không dùng số âm.
5. Loại thu chi: income (thu) hoặc expense (chi).
6. Chiều tiền: receivable (cho vay/phải thu), payable (đi vay/phải trả).
-->

## Tài khoản

| mã | tên | loại | tiền tệ |
|---|---|---|---|
| BANK_TP | Bank TP | bank | VND |
| CASH | Tiền mặt | cash | VND |

## Thu chi

| mã | ngày | loại | tài khoản | số tiền | danh mục | ghi chú |
|---|---|---|---|---:|---|---|
| TX_001 | 2026-08-01 | income | BANK_TP | 5.000.000 | Lương | Lương tháng 8 |
| TX_002 | 2026-08-02 | expense | CASH | 120.000 | Ăn uống | Cơm trưa |

## Khoản vay

| mã hợp đồng | người vay/cho vay | chiều tiền | tài khoản | số tiền gốc | lãi/triệu/ngày | ngày vay | đáo hạn | ghi chú |
|---|---|---|---|---:|---:|---|---|---|
| LOAN_001 | Nguyễn Văn A | receivable | BANK_TP | 200.000.000 | 3.000 | 2026-08-01 | 2026-11-01 | Cho A vay |

## Thanh toán khoản vay

| mã thanh toán | mã hợp đồng | ngày | tài khoản | tiền gốc | tiền lãi | phí | miễn giảm | ghi chú |
|---|---|---|---|---:|---:|---:|---:|---|
| PAY_001 | LOAN_001 | 2026-08-07 | BANK_TP | 0 | 600.000 | 0 | 0 | Thu lãi kỳ đầu |
