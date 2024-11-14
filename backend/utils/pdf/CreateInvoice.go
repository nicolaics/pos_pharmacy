package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nicolaics/pos_pharmacy/config"
	"github.com/nicolaics/pos_pharmacy/constants"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"

	"github.com/go-pdf/fpdf"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func CreateInvoicePDF(invoice types.InvoicePDFPayload, invoiceStore types.InvoiceStore, prevFileName string) (string, error) {
	directory, err := filepath.Abs("static/pdf/invoice/")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(directory, 0744); err != nil {
		return "", err
	}

	pdf, err := initInvoicePdf()
	if err != nil {
		return "", err
	}

	err = createInvoiceHeader(pdf, invoice)
	if err != nil {
		return "", err
	}

	startTableY := pdf.GetY() + 0.2

	startX, err := createInvoiceTableHeader(pdf, startTableY)
	if err != nil {
		return "", err
	}

	err = createInvoiceData(pdf, startX, invoice.MedicineLists)
	if err != nil {
		return "", err
	}

	startFooterY := 11.0

	if pdf.GetY() > startFooterY {
		pdf.AddPage()
	}

	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.SetLineWidth(0.02)

	if pdf.PageCount() > 1 {
		pdf.SetPage(1)

		pdf.Line(startX["item"], startTableY, startX["item"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
		pdf.Line(startX["qty"], startTableY, startX["qty"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
		pdf.Line(startX["unit"], startTableY, startX["unit"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
		pdf.Line(startX["price"], startTableY, startX["price"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
		pdf.Line(startX["discount"], startTableY, startX["discount"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
		pdf.Line(startX["subtotal"], startTableY, startX["subtotal"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))

		for i := 1; i < (pdf.PageCount() - 1); i++ {
			pdf.SetPage(i + 1)
			pdf.Line(startX["item"], 0.5, startX["item"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
			pdf.Line(startX["qty"], 0.5, startX["qty"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
			pdf.Line(startX["unit"], 0.5, startX["unit"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
			pdf.Line(startX["price"], 0.5, startX["price"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
			pdf.Line(startX["discount"], 0.5, startX["discount"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
			pdf.Line(startX["subtotal"], 0.5, startX["subtotal"], (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN))
		}

		pdf.SetPage(pdf.PageCount())
		pdf.Line(startX["item"], 0.5, startX["item"], (startFooterY - 0.5))
		pdf.Line(startX["qty"], 0.5, startX["qty"], (startFooterY - 0.5))
		pdf.Line(startX["unit"], 0.5, startX["unit"], (startFooterY - 0.5))
		pdf.Line(startX["price"], 0.5, startX["price"], (startFooterY - 0.5))
		pdf.Line(startX["discount"], 0.5, startX["discount"], (startFooterY - 0.5))
		pdf.Line(startX["subtotal"], 0.5, startX["subtotal"], (startFooterY - 0.5))
	} else {
		pdf.Line(startX["item"], startTableY, startX["item"], (startFooterY - 0.5))
		pdf.Line(startX["qty"], startTableY, startX["qty"], (startFooterY - 0.5))
		pdf.Line(startX["unit"], startTableY, startX["unit"], (startFooterY - 0.5))
		pdf.Line(startX["price"], startTableY, startX["price"], (startFooterY - 0.5))
		pdf.Line(startX["discount"], startTableY, startX["discount"], (startFooterY - 0.5))
		pdf.Line(startX["subtotal"], startTableY, startX["subtotal"], (startFooterY - 0.5))
	}

	pdf.SetDashPattern([]float64{0.05, 0.05}, 0)
	pdf.Line(0.05, (startFooterY - 0.3), (constants.INVOICE_WIDTH - 0.05), (startFooterY - 0.3))

	pdf.SetDashPattern([]float64{}, 0)

	err = createInvoiceFooter(pdf, startX, startFooterY, invoice)
	if err != nil {
		return "", err
	}

	fileName := prevFileName

	if prevFileName == "" {
		fileName := "i-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
		isExist, err := invoiceStore.IsPDFUrlExist(fileName)
		if err != nil {
			return "", err
		}

		for isExist {
			fileName = "i-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
			isExist, err = invoiceStore.IsPDFUrlExist(fileName)
			if err != nil {
				return "", err
			}
		}
	}

	err = pdf.OutputFileAndClose(directory + "\\" + fileName)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func initInvoicePdf() (*fpdf.Fpdf, error) {
	s, _ := filepath.Abs("static/assets/font/")

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "cm",
		SizeStr:        "10x15",
		Size: fpdf.SizeType{
			Wd: constants.INVOICE_WIDTH,
			Ht: constants.INVOICE_HEIGHT,
		},
		FontDirStr: s,
	})

	pdf.SetMargins(0.2, 0.3, 0.2)
	pdf.SetAutoPageBreak(true, constants.INVOICE_MARGIN)

	pdf.AddUTF8Font("Arial", constants.REGULAR, "Arial.TTF")
	pdf.AddUTF8Font("Arial", constants.BOLD, "ArialBD.TTF")
	pdf.AddUTF8Font("Arial", constants.ITALIC, "ArialI.TTF")
	pdf.AddUTF8Font("Calibri", constants.REGULAR, "Calibri.TTF")
	pdf.AddUTF8Font("Calibri", constants.BOLD, "CalibriBold.TTF")
	pdf.AddUTF8Font("Bree", constants.REGULAR, "bree-serif-regular.ttf")
	pdf.AddUTF8Font("Bree", constants.BOLD, "Bree Serif Bold.ttf")

	pdf.AddPage()

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error init invoice pdf: %v", pdf.Error())
	}

	return pdf, nil
}

func createInvoiceHeader(pdf *fpdf.Fpdf, invoice types.InvoicePDFPayload) error {
	pdf.SetXY((constants.INVOICE_MARGIN + 0.1), 0.3)

	pdf.SetFont("Bree", constants.BOLD, 20)
	cellWidth := pdf.GetStringWidth(config.Envs.CompanyName) + constants.INVOICE_MARGIN
	pdf.CellFormat(cellWidth, 0.6, config.Envs.CompanyName, "", 1, "L", false, 0, "")

	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_HEADER_FONT_SZ)
	pdf.MultiCell(cellWidth, constants.INVOICE_HEADER_HEIGHT, config.Envs.CompanyAddress, "", "C", false)

	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_HEADER_FONT_SZ)
	phoneNumber := fmt.Sprintf("No. Telp: %s", config.Envs.CompanyPhoneNumber)
	pdf.CellFormat(cellWidth, constants.INVOICE_HEADER_HEIGHT, phoneNumber, "", 1, "C", false, 0, "")

	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_HEADER_FONT_SZ)
	whatsApp := fmt.Sprintf("WhatsApp: %s", config.Envs.CompanyWhatsAppNumber)
	pdf.CellFormat(cellWidth, constants.INVOICE_HEADER_HEIGHT, whatsApp, "", 1, "C", false, 0, "")

	pdf.SetFont("Calibri", constants.REGULAR, 9)
	pdf.MultiCell(cellWidth, 0.34, config.Envs.CompanySlogan, "", "C", false)

	if pdf.Error() != nil {
		return fmt.Errorf("error create invoice pdf header: %v", pdf.Error())
	}

	err := createInvoiceInfo(pdf, invoice)
	if err != nil {
		return err
	}

	pdf.SetLineWidth(0.02)
	pdf.SetDashPattern([]float64{0.05, 0.05}, 0)
	pdf.Line(0.05, pdf.GetY(), (constants.INVOICE_WIDTH - 0.05), pdf.GetY())

	pdf.SetDashPattern([]float64{}, 0)

	return nil
}

func createInvoiceInfo(pdf *fpdf.Fpdf, invoice types.InvoicePDFPayload) error {
	var caser = cases.Title(language.Indonesian)

	startX := 5.3
	space := 0.1

	pdf.SetXY(startX, 0.3)

	// Number
	{
		pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(constants.INVOICE_INFO_TITLE_WIDTH, constants.INVOICE_STD_CELL_HEIGHT, "No.", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(0.2, constants.INVOICE_STD_CELL_HEIGHT, ":", "TB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(0, constants.INVOICE_STD_CELL_HEIGHT, strconv.Itoa(invoice.Number), "RTB", 1, "L", false, 0, "")
	}

	pdf.SetXY(startX, pdf.GetY()+space)

	// Date
	{
		pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(constants.INVOICE_INFO_TITLE_WIDTH, constants.INVOICE_STD_CELL_HEIGHT, "Tgl.", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(0.2, constants.INVOICE_STD_CELL_HEIGHT, ":", "TB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(0, constants.INVOICE_STD_CELL_HEIGHT, invoice.InvoiceDate.Format("02-01-2006"), "RTB", 1, "L", false, 0, "")
	}

	pdf.SetXY(startX, pdf.GetY()+space)

	// Cashier
	{
		pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(constants.INVOICE_INFO_TITLE_WIDTH, constants.INVOICE_STD_CELL_HEIGHT, "Kasir", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(0.2, constants.INVOICE_STD_CELL_HEIGHT, ":", "TB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(0, constants.INVOICE_STD_CELL_HEIGHT, caser.String(invoice.UserName), "RTB", 1, "L", false, 0, "")
	}

	pdf.SetXY(startX, pdf.GetY()+space)

	// Printed Time
	{
		pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat((constants.INVOICE_INFO_TITLE_WIDTH + 0.6), constants.INVOICE_STD_CELL_HEIGHT, "Tgl. Cetak", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(0.2, constants.INVOICE_STD_CELL_HEIGHT, ":", "TB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_STD_FONT_SZ)
		pdf.CellFormat(0, constants.INVOICE_STD_CELL_HEIGHT, time.Now().Format("02-01-2006  15:04"), "RTB", 1, "L", false, 0, "")
	}

	pdf.SetY(pdf.GetY() + 0.3)

	if pdf.Error() != nil {
		return fmt.Errorf("error create invoice info: %v", pdf.Error())
	}

	return nil
}

func createInvoiceTableHeader(pdf *fpdf.Fpdf, startTableY float64) (map[string]float64, error) {
	pdf.SetLineWidth(0.02)

	pdf.SetY(startTableY)

	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.INVOICE_NO_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, "No.", "B", 0, "C", false, 0, "")

	itemStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.INVOICE_ITEM_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, "Item", "B", 0, "C", false, 0, "")

	qtyStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.INVOICE_QTY_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, "Qty", "B", 0, "C", false, 0, "")

	unitStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.INVOICE_UNIT_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, "Unit", "B", 0, "C", false, 0, "")

	priceStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.INVOICE_PRICE_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, "Price", "B", 0, "C", false, 0, "")

	discStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.INVOICE_DISC_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, "%", "B", 0, "C", false, 0, "")

	subtotalStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.INVOICE_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.INVOICE_SUBTOTAL_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, "Subtotal", "B", 1, "C", false, 0, "")

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error create invoice table header: %v", pdf.Error())
	}

	startX := map[string]float64{
		"item":     itemStartX,
		"qty":      qtyStartX,
		"unit":     unitStartX,
		"price":    priceStartX,
		"discount": discStartX,
		"subtotal": subtotalStartX,
	}

	return startX, nil
}

func createInvoiceData(pdf *fpdf.Fpdf, startX map[string]float64, medicineLists []types.InvoiceMedicineListsPayload) error {
	var printer = message.NewPrinter(language.Indonesian)

	pdf.SetLineWidth(0.02)
	pdf.SetY(pdf.GetY() + 0.05)

	number := 1
	nextY := pdf.GetY()

	for _, medicine := range medicineLists {
		if (pdf.GetY() + (constants.INVOICE_TABLE_HEIGHT * 3)) > (constants.INVOICE_HEIGHT - constants.INVOICE_MARGIN) {
			pdf.AddPage()

			// change top margin into 0.5
			nextY = 0.5
		}
		startY := nextY

		pdf.SetXY(pdf.GetX(), startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.INVOICE_NO_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, strconv.Itoa(number), "", 0, "C", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_TABLE_DATA_FONT_SZ)
		pdf.MultiCell(constants.INVOICE_ITEM_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, strings.ToUpper(medicine.MedicineName), "", "L", false)

		nextY = pdf.GetY()

		pdf.SetXY(startX["qty"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_TABLE_DATA_FONT_SZ)
		qtyString := printer.Sprintf("%.1f", medicine.Qty)
		pdf.CellFormat(constants.INVOICE_QTY_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, qtyString, "", 0, "C", false, 0, "")

		pdf.SetXY(startX["unit"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.INVOICE_UNIT_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, strings.ToUpper(medicine.Unit), "", 0, "C", false, 0, "")

		pdf.SetXY(startX["price"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_TABLE_DATA_FONT_SZ)
		priceString := printer.Sprintf("Rp. %.1f", medicine.Price)
		pdf.CellFormat(constants.INVOICE_PRICE_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, priceString, "", 0, "C", false, 0, "")

		pdf.SetXY(startX["discount"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_TABLE_DATA_FONT_SZ)
		discountInPercentage := (medicine.Discount / medicine.Price) * 100
		discountString := printer.Sprintf("%.1f", discountInPercentage)
		pdf.CellFormat(constants.INVOICE_DISC_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, discountString, "", 0, "C", false, 0, "")

		pdf.SetXY(startX["subtotal"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_TABLE_DATA_FONT_SZ)
		subtotalString := printer.Sprintf("Rp. %.1f", medicine.Subtotal)
		pdf.CellFormat(constants.INVOICE_SUBTOTAL_COL_WIDTH, constants.INVOICE_TABLE_HEIGHT, subtotalString, "", 1, "C", false, 0, "")

		number++
	}

	if pdf.Error() != nil {
		return fmt.Errorf("error create invoice data: %v", pdf.Error())
	}

	return nil
}

func createInvoiceFooter(pdf *fpdf.Fpdf, startX map[string]float64, startFooterY float64, invoice types.InvoicePDFPayload) error {
	var printer = message.NewPrinter(language.Indonesian)

	pdf.SetLineWidth(0.02)
	pdf.SetDashPattern([]float64{}, 0)

	// Description
	{
		pdf.SetY(startFooterY)
		cellWidth := startX["unit"] - pdf.GetX() - 0.1
		pdf.SetFont("Calibri", constants.BOLD, constants.INVOICE_FOOTER_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, "Note:", "T", 1, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, 6)
		pdf.MultiCell(cellWidth, constants.INVOICE_STD_CELL_HEIGHT, invoice.Description, "", "L", false)

		pdf.Line(pdf.GetX(), startFooterY, pdf.GetX(), 14.0)
		pdf.Line(pdf.GetX(), 14.0, (startX["unit"] - 0.1), 14.0)
		pdf.Line((startX["unit"] - 0.1), startFooterY, (startX["unit"] - 0.1), 14.0)
	}

	pdf.SetXY(startX["unit"], startFooterY)

	cellWidth := startX["discount"] - pdf.GetX()

	// Subtotal
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.INVOICE_FOOTER_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, "Subtotal:", "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_FOOTER_FONT_SZ)
		subtotalString := printer.Sprintf("Rp. %.1f", invoice.Subtotal)
		pdf.CellFormat(0, constants.INVOICE_FOOTER_CELL_HEIGHT, subtotalString, "", 1, "L", false, 0, "")
	}

	pdf.SetX(startX["unit"])

	// Discount
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.INVOICE_FOOTER_FONT_SZ)
		discountPercentageString := printer.Sprintf("Discount (%.1f%%):", invoice.DiscountPercentage)
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, discountPercentageString, "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_FOOTER_FONT_SZ)
		discountString := printer.Sprintf("Rp. %.1f", invoice.Discount)
		pdf.CellFormat(0, constants.INVOICE_FOOTER_CELL_HEIGHT, discountString, "", 1, "L", false, 0, "")
	}

	pdf.SetX(startX["unit"])

	// Tax
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.INVOICE_FOOTER_FONT_SZ)
		taxPercentageString := printer.Sprintf("Tax (%.1f%%):", invoice.TaxPercentage)
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, taxPercentageString, "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_FOOTER_FONT_SZ)
		taxString := printer.Sprintf("Rp. %.1f", invoice.Tax)
		pdf.CellFormat(0, constants.INVOICE_FOOTER_CELL_HEIGHT, taxString, "", 1, "L", false, 0, "")
	}

	pdf.SetX(startX["unit"])

	// Total
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.INVOICE_FOOTER_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, "Total:", "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_FOOTER_FONT_SZ)
		totalString := printer.Sprintf("Rp. %.1f", invoice.TotalPrice)
		pdf.CellFormat(0, constants.INVOICE_FOOTER_CELL_HEIGHT, totalString, "", 1, "L", false, 0, "")
	}

	pdf.SetX(startX["unit"])

	// Paid Amount
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.INVOICE_FOOTER_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, "Paid:", "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_FOOTER_FONT_SZ)
		paidAmountString := printer.Sprintf("Rp. %.1f", invoice.PaidAmount)
		pdf.CellFormat(0, constants.INVOICE_FOOTER_CELL_HEIGHT, paidAmountString, "", 1, "L", false, 0, "")
	}

	pdf.SetX(startX["unit"])

	// Change Amount
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.INVOICE_FOOTER_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, "Change:", "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.INVOICE_FOOTER_FONT_SZ)
		changeAmountString := printer.Sprintf("Rp. %.1f", invoice.ChangeAmount)
		pdf.CellFormat(0, constants.INVOICE_FOOTER_CELL_HEIGHT, changeAmountString, "", 1, "L", false, 0, "")
	}

	// Disclaimer
	pdf.SetXY(0.95, (pdf.GetY() + 0.2))
	{
		pdf.SetTextColor(constants.RED_R, constants.RED_G, constants.RED_B)
		pdf.SetFont("Arial", constants.BOLD, 6)
		cellWidth = pdf.GetStringWidth("Barang yang sudah dibeli tidak dapat ditukar atau dikembalikan! ")
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, "Barang yang sudah dibeli tidak dapat ditukar atau dikembalikan! ", "LTB", 0, "L", false, 0, "")

		pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
		pdf.SetFont("Arial", constants.BOLD, 6)
		cellWidth = pdf.GetStringWidth("Terima Kasih!") + constants.INVOICE_MARGIN
		pdf.CellFormat(cellWidth, constants.INVOICE_FOOTER_CELL_HEIGHT, "Terima Kasih!", "RTB", 1, "L", false, 0, "")
	}

	if pdf.Error() != nil {
		return fmt.Errorf("error create invoice footer: %v", pdf.Error())
	}

	return nil
}
