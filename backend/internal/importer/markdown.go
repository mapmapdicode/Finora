// Package importer parses the documented Finora Markdown import format.
package importer

import (
	"fmt"
	"strings"
	"time"
)

type Issue struct {
	Line    int    `json:"line"`
	Section string `json:"section"`
	Message string `json:"message"`
}

type Row struct {
	Code   string            `json:"code"`
	Line   int               `json:"line"`
	Values map[string]string `json:"values"`
}

type Document struct {
	Accounts     []Row `json:"accounts"`
	Transactions []Row `json:"transactions"`
	Loans        []Row `json:"loans"`
	Payments     []Row `json:"payments"`
}

var sectionNames = map[string]string{
	"tài khoản": "accounts", "thu chi": "transactions", "khoản vay": "loans", "thanh toán khoản vay": "payments",
}

// Parse accepts markdown tables under the four Vietnamese headings described
// in the downloadable template. Unknown headings are ignored so prose can be
// included in the file without affecting the import.
func Parse(markdown string) (Document, []Issue) {
	var doc Document
	var issues []Issue
	current := ""
	var headers []string
	for index, raw := range strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n") {
		lineNo := index + 1
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "## ") {
			current = sectionNames[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "## ")))]
			headers = nil
			continue
		}
		if current == "" || !strings.HasPrefix(line, "|") {
			continue
		}
		cells := splitTableRow(line)
		if len(cells) == 0 || isDivider(cells) {
			continue
		}
		if headers == nil {
			headers = cells
			continue
		}
		if len(cells) != len(headers) {
			issues = append(issues, Issue{Line: lineNo, Section: current, Message: "số cột không khớp dòng tiêu đề"})
			continue
		}
		values := make(map[string]string, len(headers))
		for i, header := range headers {
			values[normalizeHeader(header)] = strings.TrimSpace(cells[i])
		}
		code := values["mã"]
		switch current {
		case "loans":
			code = values["mã hợp đồng"]
		case "payments":
			code = values["mã thanh toán"]
		}
		if code == "" {
			issues = append(issues, Issue{Line: lineNo, Section: current, Message: "thiếu mã tham chiếu"})
			continue
		}
		row := Row{Code: code, Line: lineNo, Values: values}
		switch current {
		case "accounts":
			doc.Accounts = append(doc.Accounts, row)
		case "transactions":
			doc.Transactions = append(doc.Transactions, row)
		case "loans":
			doc.Loans = append(doc.Loans, row)
		case "payments":
			doc.Payments = append(doc.Payments, row)
		}
	}
	return doc, issues
}

func ValidateMonth(doc Document, month string) []Issue {
	if _, err := time.Parse("2006-01", month); err != nil {
		return []Issue{{Message: "tháng nhập phải có dạng YYYY-MM"}}
	}
	issues := []Issue{}
	seen := map[string]bool{}
	check := func(section string, rows []Row, dateKey string) {
		for _, row := range rows {
			key := section + ":" + row.Code
			if seen[key] {
				issues = append(issues, Issue{Line: row.Line, Section: section, Message: "mã tham chiếu bị trùng"})
			}
			seen[key] = true
			date := row.Values[dateKey]
			if date == "" {
				issues = append(issues, Issue{Line: row.Line, Section: section, Message: "thiếu ngày"})
				continue
			}
			parsed, err := time.Parse("2006-01-02", date)
			if err != nil || parsed.Format("2006-01") != month {
				issues = append(issues, Issue{Line: row.Line, Section: section, Message: fmt.Sprintf("ngày %q không thuộc tháng %s", date, month)})
			}
		}
	}
	check("transactions", doc.Transactions, "ngày")
	check("loans", doc.Loans, "ngày vay")
	check("payments", doc.Payments, "ngày")
	return issues
}

func splitTableRow(line string) []string {
	parts := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func isDivider(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		if strings.Trim(strings.Trim(cell, ":"), "-") != "" {
			return false
		}
	}
	return true
}

func normalizeHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	// Accept common Vietnamese variants found in exported/imported templates,
	// while exposing one canonical key to the commit layer.
	switch value {
	case "tên tài khoản":
		return "tên"
	case "loại hình":
		return "loại"
	case "tiền tệ", "currency":
		return "đơn vị"
	case "đối tác", "người vay", "người cho vay":
		return "người vay/cho vay"
	default:
		return value
	}
}
