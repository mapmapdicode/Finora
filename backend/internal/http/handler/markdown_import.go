package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/importer"
)

// PreviewMarkdownImport parses a file without changing financial data. Commit
// is deliberately kept separate so the user sees every validation issue first.
func (h *WealthHandler) PreviewMarkdownImport(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body markdownImportPreviewRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	if len(body.Markdown) == 0 || len(body.Markdown) > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_IMPORT", "message": "file Markdown phải có kích thước từ 1 byte đến 5 MB"})
		return
	}
	month := strings.TrimSpace(body.Month)
	doc, issues := importer.Parse(body.Markdown)
	issues = append(issues, importer.ValidateMonth(doc, month)...)
	c.JSON(http.StatusOK, gin.H{
		"month": month, "overwrite": body.Overwrite, "canCommit": len(issues) == 0,
		"summary": gin.H{"accounts": len(doc.Accounts), "transactions": len(doc.Transactions), "loans": len(doc.Loans), "payments": len(doc.Payments)},
		"issues":  issues,
	})
}

// CommitMarkdownImport writes a previously previewed Markdown document. Source
// references make the operation safely repeatable: choosing overwrite allows a
// retry of an already imported source row without creating a duplicate. A full
// destructive replacement is intentionally not inferred from a checkbox.
func (h *WealthHandler) CommitMarkdownImport(c *gin.Context) {
	if !h.requireEditorRole(c) {
		return
	}
	var body markdownImportPreviewRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": err.Error()})
		return
	}
	if len(body.Markdown) == 0 || len(body.Markdown) > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_IMPORT", "message": "file Markdown phải có kích thước từ 1 byte đến 5 MB"})
		return
	}
	month := strings.TrimSpace(body.Month)
	doc, issues := importer.Parse(body.Markdown)
	issues = append(issues, importer.ValidateMonth(doc, month)...)
	if len(issues) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": "IMPORT_VALIDATION_FAILED", "message": "file có lỗi, chưa có dữ liệu nào được ghi", "issues": issues})
		return
	}
	result, err := h.commitMarkdownDocument(domain.ID(h.requireUserID(c)), doc, month, body.Overwrite)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "IMPORT_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"month": month, "result": result})
}

type markdownImportResult struct {
	AccountsCreated     int `json:"accountsCreated"`
	TransactionsCreated int `json:"transactionsCreated"`
	LoansCreated        int `json:"loansCreated"`
	PaymentsCreated     int `json:"paymentsCreated"`
	RowsSkipped         int `json:"rowsSkipped"`
}

func (h *WealthHandler) commitMarkdownDocument(userID domain.ID, doc importer.Document, month string, overwrite bool) (markdownImportResult, error) {
	result := markdownImportResult{}
	portfolio := h.getOrCreatePrimaryPortfolio(userID)
	if portfolio.ID == "" {
		return result, fmt.Errorf("không thể tạo không gian tài chính")
	}
	accounts := map[string]domain.Account{}
	resolveAccount := func(code string) (domain.Account, error) {
		code = strings.TrimSpace(code)
		if account, ok := accounts[code]; ok {
			return account, nil
		}
		ref, ok := h.store.GetImportReference(userID, "account", code)
		if !ok {
			return domain.Account{}, fmt.Errorf("không tìm thấy tài khoản có mã %q", code)
		}
		account, ok := h.store.GetAccount(ref.EntityID)
		if !ok || account.UserID != userID {
			return domain.Account{}, fmt.Errorf("tài khoản %q không còn tồn tại", code)
		}
		accounts[code] = *account
		return *account, nil
	}

	for _, row := range doc.Accounts {
		code := strings.TrimSpace(row.Code)
		if existing, ok := h.store.GetImportReference(userID, "account", code); ok {
			account, found := h.store.GetAccount(existing.EntityID)
			if !found || account.UserID != userID {
				return result, fmt.Errorf("tài khoản đã import %q không còn tồn tại", code)
			}
			accounts[code] = *account
			result.RowsSkipped++
			continue
		}
		name, kind, currency := strings.TrimSpace(row.Values["tên"]), strings.TrimSpace(row.Values["loại"]), strings.ToUpper(strings.TrimSpace(row.Values["đơn vị"]))
		if name == "" || kind == "" || currency == "" {
			return result, fmt.Errorf("dòng %d: tài khoản cần mã, tên, loại và đơn vị", row.Line)
		}
		account, err := h.store.CreateAccount(domain.Account{UserID: userID, PortfolioID: portfolio.ID, Name: name, Type: kind, Currency: currency})
		if err != nil {
			return result, fmt.Errorf("dòng %d: không tạo được tài khoản: %w", row.Line, err)
		}
		if _, err = h.store.UpsertImportReference(domain.ImportReference{UserID: userID, EntityType: "account", ExternalCode: code, EntityID: account.ID, ImportMonth: month}); err != nil {
			return result, err
		}
		accounts[code] = account
		result.AccountsCreated++
	}

	for _, row := range doc.Transactions {
		if _, exists := h.store.GetImportReference(userID, "transaction", row.Code); exists {
			if !overwrite {
				return result, fmt.Errorf("dòng %d: giao dịch %q đã được import", row.Line, row.Code)
			}
			result.RowsSkipped++
			continue
		}
		account, err := resolveAccount(row.Values["tài khoản"])
		if err != nil {
			return result, fmt.Errorf("dòng %d: %w", row.Line, err)
		}
		amount, err := markdownVND(row.Values["số tiền"], false)
		if err != nil {
			return result, fmt.Errorf("dòng %d: số tiền không hợp lệ", row.Line)
		}
		occurred, err := markdownDate(row.Values["ngày"])
		if err != nil {
			return result, fmt.Errorf("dòng %d: ngày không hợp lệ", row.Line)
		}
		kind := domain.TransactionType(strings.TrimSpace(row.Values["loại"]))
		if kind != domain.TransactionTypeIncome && kind != domain.TransactionTypeExpense {
			return result, fmt.Errorf("dòng %d: loại thu chi phải là income hoặc expense", row.Line)
		}
		name := strings.TrimSpace(row.Values["danh mục"])
		if name == "" {
			name = "Giao dịch import"
		}
		item, err := h.service.CreateTransaction(domain.Transaction{UserID: userID, AccountID: account.ID, PortfolioID: account.PortfolioID, Name: name, Type: kind, Amount: amount, Currency: account.Currency, Note: strings.TrimSpace(row.Values["ghi chú"]), Status: domain.TransactionStatusPosted, OccurredAt: occurred, Source: "markdown_import"})
		if err != nil {
			return result, fmt.Errorf("dòng %d: %w", row.Line, err)
		}
		if _, err = h.store.UpsertImportReference(domain.ImportReference{UserID: userID, EntityType: "transaction", ExternalCode: row.Code, EntityID: item.ID, ImportMonth: month}); err != nil {
			return result, err
		}
		result.TransactionsCreated++
	}

	customers := map[string]domain.ID{}
	for _, customer := range h.store.ListCustomers(userID, "", 1000) {
		customers[markdownKey(customer.Name)] = customer.ID
	}
	loans := map[string]domain.Loan{}
	for _, row := range doc.Loans {
		if existing, ok := h.store.GetImportReference(userID, "loan", row.Code); ok {
			loan, found := h.store.GetLoan(existing.EntityID)
			if !found || loan.UserID != userID {
				return result, fmt.Errorf("khoản vay đã import %q không còn tồn tại", row.Code)
			}
			loans[row.Code] = *loan
			result.RowsSkipped++
			continue
		}
		account, err := resolveAccount(row.Values["tài khoản"])
		if err != nil {
			return result, fmt.Errorf("dòng %d: %w", row.Line, err)
		}
		principal, err := markdownVND(row.Values["số tiền gốc"], false)
		if err != nil {
			return result, fmt.Errorf("dòng %d: số tiền gốc không hợp lệ", row.Line)
		}
		rate, err := markdownVND(row.Values["lãi/triệu/ngày"], true)
		if err != nil {
			return result, fmt.Errorf("dòng %d: lãi/triệu/ngày không hợp lệ", row.Line)
		}
		start, err := markdownDate(row.Values["ngày vay"])
		if err != nil {
			return result, fmt.Errorf("dòng %d: ngày vay không hợp lệ", row.Line)
		}
		due, err := markdownDate(row.Values["đáo hạn"])
		if err != nil {
			return result, fmt.Errorf("dòng %d: đáo hạn không hợp lệ", row.Line)
		}
		counterparty := strings.TrimSpace(row.Values["người vay/cho vay"])
		if counterparty == "" {
			return result, fmt.Errorf("dòng %d: thiếu người vay/cho vay", row.Line)
		}
		customerID := customers[markdownKey(counterparty)]
		if customerID == "" {
			customer, createErr := h.store.CreateCustomer(domain.Customer{UserID: userID, Name: counterparty})
			if createErr != nil {
				return result, fmt.Errorf("dòng %d: không tạo được đối tác: %w", row.Line, createErr)
			}
			customerID = customer.ID
			customers[markdownKey(counterparty)] = customerID
		}
		direction := domain.LoanDirection(strings.TrimSpace(row.Values["chiều tiền"]))
		if direction != domain.LoanDirectionReceivable && direction != domain.LoanDirectionPayable {
			return result, fmt.Errorf("dòng %d: chiều tiền phải là receivable hoặc payable", row.Line)
		}
		loan, err := h.store.CreateLoan(domain.Loan{UserID: userID, PortfolioID: account.PortfolioID, CustomerID: customerID, Counterparty: counterparty, Direction: direction, PrincipalInitial: principal, PrincipalBalance: principal, AnnualRate: "0", DailyRatePerMillion: rate, DayCountBasis: "actual/365", StartAt: start, DueAt: due, Status: domain.LoanStatusActive, SettlementAccountID: account.ID})
		if err != nil {
			return result, fmt.Errorf("dòng %d: không tạo được khoản vay: %w", row.Line, err)
		}
		if _, err = h.store.UpsertImportReference(domain.ImportReference{UserID: userID, EntityType: "loan", ExternalCode: row.Code, EntityID: loan.ID, ImportMonth: month}); err != nil {
			return result, err
		}
		loans[row.Code] = loan
		result.LoansCreated++
	}

	for _, row := range doc.Payments {
		if _, exists := h.store.GetImportReference(userID, "payment", row.Code); exists {
			if !overwrite {
				return result, fmt.Errorf("dòng %d: thanh toán %q đã được import", row.Line, row.Code)
			}
			result.RowsSkipped++
			continue
		}
		loan, ok := loans[row.Values["mã hợp đồng"]]
		if !ok {
			ref, found := h.store.GetImportReference(userID, "loan", row.Values["mã hợp đồng"])
			if !found {
				return result, fmt.Errorf("dòng %d: không tìm thấy khoản vay %q", row.Line, row.Values["mã hợp đồng"])
			}
			loanPtr, found := h.store.GetLoan(ref.EntityID)
			if !found {
				return result, fmt.Errorf("dòng %d: khoản vay %q không còn tồn tại", row.Line, row.Values["mã hợp đồng"])
			}
			loan = *loanPtr
		}
		account, err := resolveAccount(row.Values["tài khoản"])
		if err != nil {
			return result, fmt.Errorf("dòng %d: %w", row.Line, err)
		}
		principal, err := markdownVND(row.Values["tiền gốc"], true)
		if err != nil {
			return result, fmt.Errorf("dòng %d: tiền gốc không hợp lệ", row.Line)
		}
		interest, err := markdownVND(row.Values["tiền lãi"], true)
		if err != nil {
			return result, fmt.Errorf("dòng %d: tiền lãi không hợp lệ", row.Line)
		}
		fee, err := markdownVND(row.Values["phí"], true)
		if err != nil {
			return result, fmt.Errorf("dòng %d: phí không hợp lệ", row.Line)
		}
		waived, err := markdownVND(row.Values["miễn giảm"], true)
		if err != nil {
			return result, fmt.Errorf("dòng %d: miễn giảm không hợp lệ", row.Line)
		}
		occurred, err := markdownDate(row.Values["ngày"])
		if err != nil {
			return result, fmt.Errorf("dòng %d: ngày không hợp lệ", row.Line)
		}
		payment, err := h.service.CreateLoanPayment(string(loan.ID), domain.LoanPayment{UserID: userID, AccountID: account.ID, Principal: principal, Interest: interest, Fee: fee, Waived: waived, OccurredAt: occurred})
		if err != nil {
			return result, fmt.Errorf("dòng %d: %w", row.Line, err)
		}
		if _, err = h.store.UpsertImportReference(domain.ImportReference{UserID: userID, EntityType: "payment", ExternalCode: row.Code, EntityID: payment.ID, ImportMonth: month}); err != nil {
			return result, err
		}
		result.PaymentsCreated++
	}
	return result, nil
}

func markdownVND(value string, allowZero bool) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	multiplier := int64(1)
	if strings.HasSuffix(value, "tr") {
		value = strings.TrimSpace(strings.TrimSuffix(value, "tr"))
		multiplier = 1000000
	}
	if strings.HasSuffix(value, "k") {
		value = strings.TrimSpace(strings.TrimSuffix(value, "k"))
		multiplier = 1000
	}
	value = strings.NewReplacer(".", "", ",", "", " ", "").Replace(value)
	amount, err := strconv.ParseInt(value, 10, 64)
	if err != nil || amount < 0 || (!allowZero && amount == 0) {
		return "", fmt.Errorf("invalid VND")
	}
	return strconv.FormatInt(amount*multiplier, 10), nil
}

func markdownDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}
func markdownKey(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}
