package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/jung-kurt/gofpdf"
)

func hexToRGB(hexColor string) (r, g, b int) {
	if hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	// Parse hex values
	if len(hexColor) == 6 {
		rInt, _ := strconv.ParseInt(hexColor[0:2], 16, 0)
		gInt, _ := strconv.ParseInt(hexColor[2:4], 16, 0)
		bInt, _ := strconv.ParseInt(hexColor[4:6], 16, 0)
		r, g, b = int(rInt), int(gInt), int(bInt)
	}

	return r, g, b
}

func GeneratePDF(input InputUser) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.SetFont("Arial", "", 10)

	r, g, b := hexToRGB(input.Culoare)

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 18)
		pdf.SetTextColor(r, g, b)
		pdf.CellFormat(190, 10, input.Sender.Nume, "", 0, "C", false, 0, "")
		pdf.Ln(15)

		pdf.SetDrawColor(r, g, b)
		pdf.SetLineWidth(0.5)
		pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(5)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(128, 128, 128)

		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	pdf.SetTextColor(0, 0, 0)

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "INVOICE", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(95, 8, fmt.Sprintf("Invoice Number: %s", input.Seria), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Date: %s", input.Data.Format("2006-01-02")), "", 1, "R", false, 0, "")
	pdf.Ln(5)

	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(95, 8, "FROM:", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 8, "TO:", "1", 1, "L", true, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 10)

	senderHeight := 40.0
	receiverHeight := 40.0

	pdf.MultiCell(95, 6, fmt.Sprintf("%s\n%s\nEmail: %s\nCUICNP: %s",
		input.Sender.Nume,
		input.Sender.Adresa,
		input.Sender.Email,
		input.Sender.CUICNP), "LR", "L", false)

	pdf.SetY(pdf.GetY() - senderHeight + 16)
	pdf.SetX(105)

	pdf.MultiCell(95, 6, fmt.Sprintf("%s\n%s\nEmail: %s\nCUICNP: %s",
		input.Receiver.Nume,
		input.Receiver.Adresa,
		input.Receiver.Email,
		input.Receiver.CUICNP), "LR", "L", false)

	maxHeight := senderHeight
	if receiverHeight > senderHeight {
		maxHeight = receiverHeight
	}
	pdf.SetY(pdf.GetY() + (maxHeight - receiverHeight) + 5)

	pdf.Ln(10)
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 12)

	colWidth1 := 90.0 // Product name
	colWidth2 := 40.0 // Quantity
	colWidth3 := 60.0 // Price
	rowHeight := 8.0

	pdf.CellFormat(colWidth1, rowHeight, "Product", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colWidth2, rowHeight, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidth3, rowHeight, "Price", "1", 1, "R", true, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 10)

	total := float32(0)
	for i, prod := range input.Produse {
		if i%2 == 1 {
			pdf.SetFillColor(240, 240, 240)
			pdf.CellFormat(colWidth1, rowHeight, prod.Nume, "1", 0, "L", true, 0, "")
			pdf.CellFormat(colWidth2, rowHeight, prod.Cantitate, "1", 0, "C", true, 0, "")
			pdf.CellFormat(colWidth3, rowHeight, fmt.Sprintf("%.2f", prod.Pret), "1", 1, "R", true, 0, "")
		} else {
			pdf.CellFormat(colWidth1, rowHeight, prod.Nume, "1", 0, "L", false, 0, "")
			pdf.CellFormat(colWidth2, rowHeight, prod.Cantitate, "1", 0, "C", false, 0, "")
			pdf.CellFormat(colWidth3, rowHeight, fmt.Sprintf("%.2f", prod.Pret), "1", 1, "R", false, 0, "")
		}

		// Calculate total
		qty := 1.0
		if qtyFloat, err := strconv.ParseFloat(prod.Cantitate, 32); err == nil {
			qty = qtyFloat
		}
		total += prod.Pret * float32(qty)
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(colWidth1+colWidth2, rowHeight, "Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(colWidth3, rowHeight, fmt.Sprintf("%.2f", total), "1", 1, "R", false, 0, "")

	pdf.Ln(10)
	pdf.SetDrawColor(r, g, b)
	pdf.SetLineWidth(0.5)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.MultiCell(190, 5, "Thank you for your business. Payment is due within 30 days of receipt of this invoice. "+
		"Please make all checks payable to "+input.Sender.Nume+".", "", "L", false)

	// PAYMENT INFORMATION (optional)
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(190, 8, "Payment Information", "", 1, "L", false, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(190, 5, "Bank transfer to the account details provided separately or as agreed in our contract terms.", "", "L", false)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
