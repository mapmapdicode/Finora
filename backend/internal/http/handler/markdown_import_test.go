package handler

import (
	"testing"

	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/importer"
	"wealthos-backend/internal/service"
	"wealthos-backend/internal/storage"
)

func TestCommitMarkdownDocumentLinksLaterPaymentsBySourceCode(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewWealthHandler(store, service.NewWealthService(store, nil), nil)
	first, issues := importer.Parse(`# Finora Import v1

## Tài khoản
| mã | tên | loại | đơn vị |
|---|---|---|---|
| CASH | Tiền mặt | cash | VND |

## Thu chi
| mã | ngày | loại | tài khoản | số tiền | danh mục | ghi chú |
|---|---|---|---|---:|---|---|
| TX_01 | 2026-05-01 | income | CASH | 1000000 | Thu nhập | test |

## Khoản vay
| mã hợp đồng | người vay/cho vay | chiều tiền | tài khoản | số tiền gốc | lãi/triệu/ngày | ngày vay | đáo hạn | ghi chú |
|---|---|---|---|---:|---:|---|---|---|
| LOAN_01 | Anh A | receivable | CASH | 100000000 | 3000 | 2026-05-01 | 2026-12-31 | test |

## Thanh toán khoản vay
| mã thanh toán | mã hợp đồng | ngày | tài khoản | tiền gốc | tiền lãi | phí | miễn giảm | ghi chú |
|---|---|---|---|---:|---:|---:|---:|---|
| PAY_01 | LOAN_01 | 2026-05-02 | CASH | 0 | 300000 | 0 | 0 | test |`)
	if len(issues) != 0 {
		t.Fatalf("parse issues: %+v", issues)
	}
	userID := domain.ID("import-user")
	result, err := h.commitMarkdownDocument(userID, first, "2026-05", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.AccountsCreated != 1 || result.TransactionsCreated != 1 || result.LoansCreated != 1 || result.PaymentsCreated != 1 {
		t.Fatalf("unexpected first result: %+v", result)
	}

	second, issues := importer.Parse(`## Thanh toán khoản vay
| mã thanh toán | mã hợp đồng | ngày | tài khoản | tiền gốc | tiền lãi | phí | miễn giảm | ghi chú |
|---|---|---|---|---:|---:|---:|---:|---|
| PAY_02 | LOAN_01 | 2026-05-03 | CASH | 100000000 | 0 | 0 | 0 | tất toán |`)
	if len(issues) != 0 {
		t.Fatalf("parse issues: %+v", issues)
	}
	result, err = h.commitMarkdownDocument(userID, second, "2026-05", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.PaymentsCreated != 1 {
		t.Fatalf("unexpected second result: %+v", result)
	}
	loans := store.ListLoans(userID)
	if len(loans) != 1 || loans[0].Status != domain.LoanStatusClosed {
		t.Fatalf("loan was not settled: %+v", loans)
	}
}
