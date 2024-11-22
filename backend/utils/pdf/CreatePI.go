package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"strconv"
	"strings"

	"github.com/nicolaics/pharmacon/config"
	"github.com/nicolaics/pharmacon/constants"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"

	"github.com/go-pdf/fpdf"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func CreatePurchaseInvoicePDF(piStore types.PurchaseInvoiceStore, purchaseInvoice types.PurchaseInvoicePDFPayload, prevFileName string) (string, error) {
	directory, err := filepath.Abs("static/pdf/purchase-invoice/")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(directory, 0744); err != nil {
		return "", err
	}

	pdf, err := initPurchaseInvoicePdf()
	if err != nil {
		return "", err
	}

	err = createPurchaseInvoiceHeader(pdf, purchaseInvoice)
	if err != nil {
		return "", err
	}

	pdf.SetLineWidth(0.02)
	pdf.SetDashPattern([]float64{0.1, 0.1}, 0)
	pdf.SetY(2.45)
	pdf.Line(constants.PI_MARGIN, pdf.GetY(), (constants.PI_WIDTH - constants.PI_MARGIN), pdf.GetY())

	pdf.SetDashPattern([]float64{}, 0)

	pdf.SetY(pdf.GetY() + 0.2)

	err = createPurchaseInvoiceInfo(pdf, purchaseInvoice)
	if err != nil {
		return "", err
	}

	startTableY := pdf.GetY() + 0.5

	startTableX, err := createPurchaseInvoiceTableHeader(pdf, startTableY)
	if err != nil {
		return "", err
	}

	_, err = createPurchaseInvoiceData(pdf, startTableX, purchaseInvoice.MedicineLists)
	if err != nil {
		return "", err
	}

	startFooterY := 11.5

	if pdf.GetY() > startFooterY {
		pdf.AddPage()
	}

	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.SetLineWidth(0.02)

	if pdf.PageCount() > 1 {
		pdf.SetPage(1)

		pdf.Line(startTableX["number"], startTableY, startTableX["number"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		pdf.Line(startTableX["item"], startTableY, startTableX["item"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		pdf.Line(startTableX["qty"], startTableY, startTableX["qty"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		pdf.Line(startTableX["unit"], startTableY, startTableX["unit"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		pdf.Line(startTableX["price"], startTableY, startTableX["price"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		pdf.Line(startTableX["discount"], startTableY, startTableX["discount"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		pdf.Line(startTableX["tax"], startTableY, startTableX["tax"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		pdf.Line(startTableX["subtotal"], startTableY, startTableX["subtotal"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		pdf.Line(startTableX["end"], startTableY, startTableX["end"], (constants.PI_HEIGHT - constants.PI_MARGIN))

		for i := 1; i < (pdf.PageCount() - 1); i++ {
			pdf.SetPage(i + 1)
			pdf.Line(startTableX["number"], 0.5, startTableX["number"], (constants.PI_HEIGHT - constants.PI_MARGIN))
			pdf.Line(startTableX["item"], 0.5, startTableX["item"], (constants.PI_HEIGHT - constants.PI_MARGIN))
			pdf.Line(startTableX["qty"], 0.5, startTableX["qty"], (constants.PI_HEIGHT - constants.PI_MARGIN))
			pdf.Line(startTableX["unit"], 0.5, startTableX["unit"], (constants.PI_HEIGHT - constants.PI_MARGIN))
			pdf.Line(startTableX["price"], 0.5, startTableX["price"], (constants.PI_HEIGHT - constants.PI_MARGIN))
			pdf.Line(startTableX["discount"], 0.5, startTableX["discount"], (constants.PI_HEIGHT - constants.PI_MARGIN))
			pdf.Line(startTableX["tax"], 0.5, startTableX["tax"], (constants.PI_HEIGHT - constants.PI_MARGIN))
			pdf.Line(startTableX["subtotal"], 0.5, startTableX["subtotal"], (constants.PI_HEIGHT - constants.PI_MARGIN))
			pdf.Line(startTableX["end"], 0.5, startTableX["end"], (constants.PI_HEIGHT - constants.PI_MARGIN))
		}

		pdf.SetPage(pdf.PageCount())

		pdf.Line(startTableX["number"], 0.5, startTableX["number"], (startFooterY - 0.5))
		pdf.Line(startTableX["item"], 0.5, startTableX["item"], (startFooterY - 0.5))
		pdf.Line(startTableX["qty"], 0.5, startTableX["qty"], (startFooterY - 0.5))
		pdf.Line(startTableX["unit"], 0.5, startTableX["unit"], (startFooterY - 0.5))
		pdf.Line(startTableX["price"], 0.5, startTableX["price"], (startFooterY - 0.5))
		pdf.Line(startTableX["discount"], 0.5, startTableX["discount"], (startFooterY - 0.5))
		pdf.Line(startTableX["tax"], 0.5, startTableX["tax"], (startFooterY - 0.5))
		pdf.Line(startTableX["subtotal"], 0.5, startTableX["subtotal"], (startFooterY - 0.5))
		pdf.Line(startTableX["end"], 0.5, startTableX["end"], (startFooterY - 0.5))
	} else {
		pdf.Line(startTableX["number"], startTableY, startTableX["number"], (startFooterY - 0.5))
		pdf.Line(startTableX["item"], startTableY, startTableX["item"], (startFooterY - 0.5))
		pdf.Line(startTableX["qty"], startTableY, startTableX["qty"], (startFooterY - 0.5))
		pdf.Line(startTableX["unit"], startTableY, startTableX["unit"], (startFooterY - 0.5))
		pdf.Line(startTableX["price"], startTableY, startTableX["price"], (startFooterY - 0.5))
		pdf.Line(startTableX["discount"], startTableY, startTableX["discount"], (startFooterY - 0.5))
		pdf.Line(startTableX["tax"], startTableY, startTableX["tax"], (startFooterY - 0.5))
		pdf.Line(startTableX["subtotal"], startTableY, startTableX["subtotal"], (startFooterY - 0.5))
		pdf.Line(startTableX["end"], startTableY, startTableX["end"], (startFooterY - 0.5))
	}

	pdf.Line(startTableX["number"], (startFooterY - 0.5), startTableX["end"], (startFooterY - 0.5))

	pdf.SetDashPattern([]float64{}, 0)

	err = createPurchaseInvoiceFooter(pdf, startTableX, startFooterY, purchaseInvoice)
	if err != nil {
		return "", err
	}

	fileName := prevFileName

	if prevFileName == "" {
		fileName := "pi-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
		isExist, err := piStore.IsPDFUrlExist(fileName)
		if err != nil {
			return "", err
		}

		for isExist {
			fileName = "pi-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
			isExist, err = piStore.IsPDFUrlExist(fileName)
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

func initPurchaseInvoicePdf() (*fpdf.Fpdf, error) {
	s, _ := filepath.Abs("static/assets/font/")

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "cm",
		SizeStr:        "14x21",
		Size: fpdf.SizeType{
			Wd: constants.PI_WIDTH,
			Ht: constants.PI_HEIGHT,
		},
		FontDirStr: s,
	})

	pdf.SetMargins(0.2, 0.3, 0.2)
	pdf.SetAutoPageBreak(true, constants.PI_MARGIN)

	pdf.AddUTF8Font("Arial", constants.REGULAR, "Arial.TTF")
	pdf.AddUTF8Font("Arial", constants.BOLD, "ArialBD.TTF")
	pdf.AddUTF8Font("Arial", constants.ITALIC, "ArialI.TTF")
	pdf.AddUTF8Font("Calibri", constants.REGULAR, "Calibri.TTF")
	pdf.AddUTF8Font("Calibri", constants.BOLD, "CalibriBold.TTF")
	pdf.AddUTF8Font("Bree", constants.REGULAR, "bree-serif-regular.ttf")
	pdf.AddUTF8Font("Bree", constants.BOLD, "Bree Serif Bold.ttf")

	pdf.AddPage()

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error init purchase invoice pdf: %v", pdf.Error())
	}

	return pdf, nil
}

func createPurchaseInvoiceHeader(pdf *fpdf.Fpdf, purchaseInvoice types.PurchaseInvoicePDFPayload) error {
	var caser = cases.Title(language.Indonesian)

	pdf.Image(config.Envs.CompanyLogoURL, pdf.GetX(), pdf.GetY(), constants.PI_LOGO_WIDTH, constants.PI_LOGO_HEIGHT, false, "", 0, "")

	startBesideLogoX := constants.PI_MARGIN + constants.PI_LOGO_WIDTH + 0.1

	pdf.SetX(startBesideLogoX)
	pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
	pdf.SetFont("Bree", constants.BOLD, 22)
	cellWidth := pdf.GetStringWidth(config.Envs.CompanyName) + constants.PI_MARGIN
	pdf.CellFormat(cellWidth, 0.65, config.Envs.CompanyName, "", 1, "L", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_HEADER_FONT_SZ)
	cellWidth = pdf.GetStringWidth(config.Envs.CompanyAddress) + constants.PI_MARGIN
	pdf.CellFormat(cellWidth, constants.PI_HEADER_HEIGHT, config.Envs.CompanyAddress, "", 1, "C", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_HEADER_FONT_SZ)
	phoneNumber := fmt.Sprintf("No. Telp: %s | WhatsApp: %s", config.Envs.CompanyPhoneNumber, config.Envs.CompanyWhatsAppNumber)
	cellWidth = pdf.GetStringWidth(phoneNumber) + constants.PI_MARGIN
	pdf.CellFormat(cellWidth, constants.PI_HEADER_HEIGHT, phoneNumber, "", 1, "L", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_HEADER_FONT_SZ)
	cellWidth = pdf.GetStringWidth(config.Envs.BusinessRegistrationNumber) + constants.PI_MARGIN
	pdf.CellFormat(cellWidth, constants.PI_HEADER_HEIGHT, config.Envs.BusinessRegistrationNumber, "", 1, "C", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_HEADER_FONT_SZ)
	pharmacist := fmt.Sprintf("Apoteker: %s", config.Envs.Pharmacist)
	cellWidth = pdf.GetStringWidth(pharmacist) + constants.PI_MARGIN
	pdf.CellFormat(cellWidth, constants.PI_HEADER_HEIGHT, pharmacist, "", 1, "C", false, 0, "")

	pdf.SetXY((constants.PI_WIDTH / 2), 0.3)
	pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.SetFont("Calibri", constants.BOLD, 20)
	pdf.CellFormat(0, 0.65, "Purchase Invoice", "", 1, "C", false, 0, "")

	startSupplierX := ((constants.PI_WIDTH / 2) - 0.5)
	startSupplierY := (pdf.GetY() + 0.1)

	pdf.SetXY(startSupplierX, startSupplierY)

	pdf.SetFont("Calibri", constants.BOLD, constants.PI_SUPPLIER_FONT_SZ)
	cellWidth = pdf.GetStringWidth("Pemasok") + 0.1
	pdf.CellFormat(cellWidth, constants.PI_STD_CELL_HEIGHT, "Pemasok", "", 0, "L", false, 0, "")

	pdf.SetFont("Calibri", constants.BOLD, constants.PI_SUPPLIER_FONT_SZ)
	cellWidth = pdf.GetStringWidth(":") + constants.PI_MARGIN
	pdf.CellFormat(cellWidth, constants.PI_STD_CELL_HEIGHT, ":", "", 0, "L", false, 0, "")

	startSupplierDataX := pdf.GetX()

	// uppercase the supplier name
	pdf.SetX(startSupplierDataX)
	pdf.SetFont("Arial", constants.REGULAR, constants.PI_SUPPLIER_FONT_SZ)
	supplier := fmt.Sprintf("%s | T. %s", strings.ToUpper(purchaseInvoice.Supplier.Name), purchaseInvoice.Supplier.CompanyPhoneNumber)
	pdf.CellFormat(0, constants.PI_STD_CELL_HEIGHT, supplier, "", 1, "L", false, 0, "")

	// Address
	pdf.SetX(startSupplierDataX)
	pdf.SetFont("Arial", constants.REGULAR, constants.PI_SUPPLIER_FONT_SZ)
	pdf.CellFormat(0, constants.PI_STD_CELL_HEIGHT, purchaseInvoice.Supplier.Address, "", 1, "L", false, 0, "")

	// Contact Person
	pdf.SetX(startSupplierDataX)
	pdf.SetFont("Arial", constants.REGULAR, constants.PI_SUPPLIER_FONT_SZ)
	contactPerson := fmt.Sprintf("CP. %s | CP.T. %s", caser.String(purchaseInvoice.Supplier.ContactPersonName), purchaseInvoice.Supplier.ContactPersonNumber)
	pdf.CellFormat(0, constants.PI_STD_CELL_HEIGHT, contactPerson, "", 1, "L", false, 0, "")

	pdf.SetLineWidth(0.02)
	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.RoundedRect(startSupplierX, startSupplierY, (constants.PI_WIDTH - constants.PI_MARGIN - startSupplierX), (pdf.GetY() - startSupplierY), 0.1, "1234", "D")

	if pdf.Error() != nil {
		return fmt.Errorf("error create purchase invoice pdf header: %v", pdf.Error())
	}

	return nil
}

func createPurchaseInvoiceInfo(pdf *fpdf.Fpdf, pi types.PurchaseInvoicePDFPayload) error {
	var caser = cases.Title(language.Indonesian)

	var vendorIsTaxable string
	if pi.Supplier.VendorIsTaxable {
		vendorIsTaxable = "Yes"
	} else {
		vendorIsTaxable = "No"
	}

	space := 0.5

	// PO Number
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("No.:") + constants.PI_MARGIN
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, "No.:", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		pdf.CellFormat(constants.PI_INFO_NUMBER_WIDTH, constants.PI_INFO_HEIGHT, strconv.Itoa(pi.Number), "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// PO Date
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Tgl.: ") + constants.PI_MARGIN
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, "Tgl.: ", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		pdf.CellFormat(constants.PI_INFO_DATE_WIDTH, constants.PI_INFO_HEIGHT, pi.InvoiceDate.Format("02-01-2006"), "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// Terms
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Terms: ") + constants.PI_MARGIN
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, "Terms: ", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		pdf.CellFormat(constants.PI_INFO_TERMS_WIDTH, constants.PI_INFO_HEIGHT, pi.Supplier.Terms, "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// Vendor is Taxable
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Vendor is Taxable: ") + constants.PI_MARGIN
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, "Vendor is Taxable: ", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		pdf.CellFormat(constants.PI_INFO_VENDOR_IS_TAX_WIDTH, constants.PI_INFO_HEIGHT, vendorIsTaxable, "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// Cashier
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Dibuat Oleh: ") + constants.PI_MARGIN
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, "Dibuat Oleh: ", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		pdf.CellFormat(constants.PI_INFO_CASHIER_WIDTH, constants.PI_INFO_HEIGHT, caser.String(pi.UserName), "RTB", 1, "L", false, 0, "")
	}

	if pdf.Error() != nil {
		return fmt.Errorf("error create purchase invoice info: %v", pdf.Error())
	}

	return nil
}

func createPurchaseInvoiceTableHeader(pdf *fpdf.Fpdf, startTableY float64) (map[string]float64, error) {
	pdf.SetLineWidth(0.02)

	pdf.SetY(startTableY)

	numberStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PI_NO_COL_WIDTH, constants.PI_TABLE_HEIGHT, "No.", "TB", 0, "C", false, 0, "")

	itemStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PI_ITEM_COL_WIDTH, constants.PI_TABLE_HEIGHT, "Item", "TB", 0, "C", false, 0, "")

	qtyStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PI_QTY_COL_WIDTH, constants.PI_TABLE_HEIGHT, "Qty", "TB", 0, "C", false, 0, "")

	unitStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PI_UNIT_COL_WIDTH, constants.PI_TABLE_HEIGHT, "Unit", "TB", 0, "C", false, 0, "")

	priceStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PI_PRICE_COL_WIDTH, constants.PI_TABLE_HEIGHT, "Price @", "TB", 0, "C", false, 0, "")

	discStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PI_DISC_COL_WIDTH, constants.PI_TABLE_HEIGHT, "%", "TB", 0, "C", false, 0, "")

	taxStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PI_TAX_COL_WIDTH, constants.PI_TABLE_HEIGHT, "Tax", "TB", 0, "C", false, 0, "")

	subtotalStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.PI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PI_SUBTOTAL_COL_WIDTH, constants.PI_TABLE_HEIGHT, "Subtotal", "TB", 0, "C", false, 0, "")

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error create purchase invoice table header: %v", pdf.Error())
	}

	startX := map[string]float64{
		"number":   numberStartX,
		"item":     itemStartX,
		"qty":      qtyStartX,
		"unit":     unitStartX,
		"price":    priceStartX,
		"discount": discStartX,
		"tax":      taxStartX,
		"subtotal": subtotalStartX,
		"end":      pdf.GetX(),
	}

	pdf.Ln(-1)

	return startX, nil
}

func createPurchaseInvoiceData(pdf *fpdf.Fpdf, startTableX map[string]float64, medicineLists []types.PurchaseMedicineListPayload) (int, error) {
	var printer = message.NewPrinter(language.Indonesian)

	pdf.SetLineWidth(0.02)
	pdf.SetY(pdf.GetY() + 0.05)

	number := 1
	nextY := pdf.GetY()

	for _, medicine := range medicineLists {
		if (pdf.GetY() + (constants.PI_TABLE_HEIGHT * 2)) > (constants.PI_HEIGHT - constants.PI_MARGIN) {
			pdf.AddPage()

			// change top margin into 0.5
			nextY = 0.5
		}
		startY := nextY

		pdf.SetXY(pdf.GetX(), startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PI_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.PI_NO_COL_WIDTH, constants.PI_TABLE_HEIGHT, strconv.Itoa(number), "", 0, "C", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_TABLE_DATA_FONT_SZ)
		pdf.MultiCell(constants.PI_ITEM_COL_WIDTH, constants.PI_TABLE_HEIGHT, strings.ToUpper(medicine.MedicineName), "", "L", false)

		nextY = pdf.GetY()

		pdf.SetXY(startTableX["qty"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PI_TABLE_DATA_FONT_SZ)
		qtyString := printer.Sprintf("%.1f", medicine.Qty)
		pdf.CellFormat(constants.PI_QTY_COL_WIDTH, constants.PI_TABLE_HEIGHT, qtyString, "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["unit"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PI_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.PI_UNIT_COL_WIDTH, constants.PI_TABLE_HEIGHT, strings.ToUpper(medicine.Unit), "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["price"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PI_TABLE_DATA_FONT_SZ)
		priceString := printer.Sprintf("Rp. %.1f", medicine.Price)
		pdf.CellFormat(constants.PI_PRICE_COL_WIDTH, constants.PI_TABLE_HEIGHT, priceString, "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["discount"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PI_TABLE_DATA_FONT_SZ)
		discountString := printer.Sprintf("%.1f", medicine.DiscountPercentage)
		pdf.CellFormat(constants.PI_DISC_COL_WIDTH, constants.PI_TABLE_HEIGHT, discountString, "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["tax"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PI_TABLE_DATA_FONT_SZ)
		taxString := printer.Sprintf("%.1f", medicine.TaxPercentage)
		pdf.CellFormat(constants.PI_DISC_COL_WIDTH, constants.PI_TABLE_HEIGHT, taxString, "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["subtotal"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PI_TABLE_DATA_FONT_SZ)
		subtotalString := printer.Sprintf("Rp. %.1f", medicine.Subtotal)
		pdf.CellFormat(constants.PI_SUBTOTAL_COL_WIDTH, constants.PI_TABLE_HEIGHT, subtotalString, "", 1, "C", false, 0, "")

		number++
	}

	if pdf.Error() != nil {
		return 0, fmt.Errorf("error create purchase invoice data: %v", pdf.Error())
	}

	return (number - 1), nil
}

func createPurchaseInvoiceFooter(pdf *fpdf.Fpdf, startTableX map[string]float64, startFooterY float64, pi types.PurchaseInvoicePDFPayload) error {
	var printer = message.NewPrinter(language.Indonesian)

	pdf.SetLineWidth(0.02)
	pdf.SetDashPattern([]float64{}, 0)

	pdf.SetY(startFooterY)

	// Purchase Order Number
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("No. PO: ") + constants.PI_MARGIN
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, "No. PO: ", "LT", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		cellWidth = startTableX["qty"] - pdf.GetX()
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, strconv.Itoa(pi.PurchaseOrderNumber), "RT", 1, "L", false, 0, "")
	}

	// Purchase Order Date
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Tgl. PO: ") + constants.PI_MARGIN
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, "Tgl. PO: ", "LB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		cellWidth = startTableX["qty"] - pdf.GetX()
		pdf.CellFormat(cellWidth, constants.PI_INFO_HEIGHT, pi.PurchaseOrderDate.Format("02-01-2006"), "RB", 1, "L", false, 0, "")
	}

	pdf.SetXY(startTableX["discount"], startFooterY)

	cellWidth := startTableX["subtotal"] - startTableX["discount"]

	// Subtotal
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.PI_FOOTER_CELL_HEIGHT, "Subtotal:", "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		subtotalString := printer.Sprintf("Rp. %.1f", pi.Subtotal)
		pdf.CellFormat(0, constants.PI_FOOTER_CELL_HEIGHT, subtotalString, "", 1, "L", false, 0, "")
	}

	pdf.SetX(startTableX["discount"])

	// Discount
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		discountPercentageString := printer.Sprintf("Discount (%.1f%%):", pi.DiscountPercentage)
		pdf.CellFormat(cellWidth, constants.PI_FOOTER_CELL_HEIGHT, discountPercentageString, "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		discountString := printer.Sprintf("Rp. %.1f", pi.DiscountAmount)
		pdf.CellFormat(0, constants.PI_FOOTER_CELL_HEIGHT, discountString, "", 1, "L", false, 0, "")
	}

	pdf.SetX(startTableX["discount"])

	// Tax
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		taxPercentageString := printer.Sprintf("Tax (%.1f%%):", pi.TaxPercentage)
		pdf.CellFormat(cellWidth, constants.PI_FOOTER_CELL_HEIGHT, taxPercentageString, "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		taxString := printer.Sprintf("Rp. %.1f", pi.TaxAmount)
		pdf.CellFormat(0, constants.PI_FOOTER_CELL_HEIGHT, taxString, "", 1, "L", false, 0, "")
	}

	pdf.SetXY(startTableX["discount"], (pdf.GetY() + 0.05))
	pdf.Line(pdf.GetX(), pdf.GetY(), (constants.PI_WIDTH - constants.PI_MARGIN), pdf.GetY())
	pdf.SetXY(startTableX["discount"], (pdf.GetY() + 0.05))

	// Total
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.PI_FOOTER_CELL_HEIGHT, "Total:", "", 0, "R", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
		totalString := printer.Sprintf("Rp. %.1f", pi.TotalPrice)
		pdf.CellFormat(0, constants.PI_FOOTER_CELL_HEIGHT, totalString, "", 1, "L", false, 0, "")
	}

	// pdf.SetX(startTableX["discount"])

	// // Paid Amount
	// {
	// 	pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
	// 	pdf.CellFormat(cellWidth, constants.PI_FOOTER_CELL_HEIGHT, "Paid:", "", 0, "R", false, 0, "")

	// 	pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
	// 	paidAmountString := printer.Sprintf("Rp. %.1f", paidAmount)
	// 	pdf.CellFormat(0, constants.PI_FOOTER_CELL_HEIGHT, paidAmountString, "", 1, "L", false, 0, "")
	// }

	// pdf.SetX(startTableX["unit"])

	// // Change Amount
	// {
	// 	pdf.SetFont("Calibri", constants.BOLD, constants.PI_STD_FONT_SZ)
	// 	pdf.CellFormat(cellWidth, constants.PI_FOOTER_CELL_HEIGHT, "Change:", "", 0, "R", false, 0, "")

	// 	pdf.SetFont("Arial", constants.REGULAR, constants.PI_STD_FONT_SZ)
	// 	changeAmountString := printer.Sprintf("Rp. %.1f", changeAmount)
	// 	pdf.CellFormat(0, constants.PI_FOOTER_CELL_HEIGHT, changeAmountString, "", 1, "L", false, 0, "")
	// }

	if pdf.Error() != nil {
		return fmt.Errorf("error create purchase invoice footer: %v", pdf.Error())
	}

	return nil
}
