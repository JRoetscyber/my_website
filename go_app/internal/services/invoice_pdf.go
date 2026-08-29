package services

import (
	"bytes"
	"fmt"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/go-pdf/fpdf"
)

type LineItem struct {
	Description string
	Quantity    int
	UnitPrice   float64
	Total       float64
}

func GenerateInvoicePDF(settings *models.InvoiceSettings, invoiceNum, clientName, clientEmail, clientAddress string, items []LineItem, vatRate float64) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Colors
	brandR, brandG, brandB := 20, 20, 20
	textR, textG, textB := 60, 60, 60

	// Header
	pdf.SetTextColor(brandR, brandG, brandB)
	pdf.SetFont("Helvetica", "B", 20)
	bizName := settings.BizName
	if bizName == "" {
		bizName = "JO4 DEV"
	}
	pdf.CellFormat(100, 10, bizName, "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(180, 40, 40)
	pdf.CellFormat(80, 10, "INVOICE", "", 1, "R", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(textR, textG, textB)
	pdf.CellFormat(100, 5, settings.BizAddress, "", 0, "L", false, 0, "")
	pdf.CellFormat(80, 5, fmt.Sprintf("Invoice #: %s", invoiceNum), "", 1, "R", false, 0, "")

	pdf.CellFormat(100, 5, fmt.Sprintf("Email: %s", settings.BizEmail), "", 0, "L", false, 0, "")
	pdf.CellFormat(80, 5, fmt.Sprintf("Phone: %s", settings.BizPhone), "", 1, "R", false, 0, "")

	pdf.Ln(10)

	// Bill To Box
	pdf.SetFillColor(245, 247, 250)
	pdf.Rect(15, pdf.GetY(), 180, 25, "F")
	pdf.SetXY(18, pdf.GetY()+3)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(brandR, brandG, brandB)
	pdf.Cell(50, 5, "BILLED TO:")
	pdf.Ln(5)
	pdf.SetX(18)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(textR, textG, textB)
	pdf.Cell(100, 4, clientName)
	pdf.Ln(4)
	pdf.SetX(18)
	pdf.Cell(100, 4, clientEmail)
	if clientAddress != "" {
		pdf.Ln(4)
		pdf.SetX(18)
		pdf.Cell(100, 4, clientAddress)
	}

	pdf.SetY(pdf.GetY() + 15)

	// Items Table Header
	pdf.SetFillColor(30, 30, 30)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(100, 8, " Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Unit Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(30, 8, "Total", "1", 1, "R", true, 0, "")

	// Items
	pdf.SetTextColor(textR, textG, textB)
	pdf.SetFont("Helvetica", "", 9)
	var subtotal float64
	for _, it := range items {
		lineTotal := float64(it.Quantity) * it.UnitPrice
		subtotal += lineTotal
		pdf.CellFormat(100, 7, fmt.Sprintf(" %s", it.Description), "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 7, fmt.Sprintf("%d", it.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 7, fmt.Sprintf("R %.2f", it.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("R %.2f", lineTotal), "1", 1, "R", false, 0, "")
	}

	// Totals
	vatAmount := subtotal * (vatRate / 100.0)
	grandTotal := subtotal + vatAmount

	pdf.Ln(3)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(150, 6, "Subtotal:", "", 0, "R", false, 0, "")
	pdf.CellFormat(30, 6, fmt.Sprintf("R %.2f", subtotal), "", 1, "R", false, 0, "")

	if vatRate > 0 {
		pdf.CellFormat(150, 6, fmt.Sprintf("VAT (%.1f%%):", vatRate), "", 0, "R", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("R %.2f", vatAmount), "", 1, "R", false, 0, "")
	}

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(180, 40, 40)
	pdf.CellFormat(150, 8, "TOTAL DUE:", "", 0, "R", false, 0, "")
	pdf.CellFormat(30, 8, fmt.Sprintf("R %.2f", grandTotal), "", 1, "R", false, 0, "")

	// Banking Details & Terms
	pdf.Ln(10)
	pdf.SetTextColor(brandR, brandG, brandB)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(100, 5, "Banking Details:")
	pdf.Ln(5)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(textR, textG, textB)
	if settings.BankName != "" {
		pdf.Cell(180, 4, fmt.Sprintf("Bank: %s", settings.BankName))
		pdf.Ln(4)
	}
	if settings.AccountHolder != "" {
		pdf.Cell(180, 4, fmt.Sprintf("Account Holder: %s", settings.AccountHolder))
		pdf.Ln(4)
	}
	if settings.AccountNumber != "" {
		pdf.Cell(180, 4, fmt.Sprintf("Account Number: %s", settings.AccountNumber))
		pdf.Ln(4)
	}
	if settings.BranchCode != "" {
		pdf.Cell(180, 4, fmt.Sprintf("Branch Code: %s", settings.BranchCode))
		pdf.Ln(4)
	}

	if settings.PaymentTerms != "" {
		pdf.Ln(4)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.MultiCell(180, 4, fmt.Sprintf("Terms: %s", settings.PaymentTerms), "", "L", false)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
