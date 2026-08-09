package importer

import "testing"

func TestParseAndValidateMonth(t *testing.T) {
	input := "## Thu chi\n| mã | ngày | loại | tài khoản | số tiền |\n|---|---|---|---|---:|\n| TX_1 | 2026-08-01 | income | BANK | 5.000.000 |\n"
	doc, issues := Parse(input)
	if len(issues) != 0 || len(doc.Transactions) != 1 {
		t.Fatalf("unexpected parse result: %#v, %#v", doc, issues)
	}
	if got := ValidateMonth(doc, "2026-08"); len(got) != 0 {
		t.Fatalf("expected no month issue: %#v", got)
	}
	if got := ValidateMonth(doc, "2026-09"); len(got) != 1 {
		t.Fatalf("expected month issue: %#v", got)
	}
}

func TestParseAccountCurrencyHeaderAlias(t *testing.T) {
	doc, issues := Parse("## Tài khoản\n| mã | tên | loại | tiền tệ |\n|---|---|---|---|\n| CASH | Tiền mặt | cash | VND |\n")
	if len(issues) != 0 || len(doc.Accounts) != 1 {
		t.Fatalf("unexpected parse result: %#v, %#v", doc, issues)
	}
	if got := doc.Accounts[0].Values["đơn vị"]; got != "VND" {
		t.Fatalf("currency alias was not normalized: %q", got)
	}
}
