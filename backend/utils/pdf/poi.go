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

func CreatePurchaseOrderInvoicePdf(poiStore types.PurchaseOrderStore, poi types.PurchaseOrderPdfPayload, prevFileName string) (string, error) {
	directory, err := filepath.Abs("static/pdf/purchase-order/")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(directory, 0744); err != nil {
		return "", err
	}

	pdf, err := initPurchaseOrderInvoicePdf()
	if err != nil {
		return "", err
	}

	err = createPurchaseOrderInvoiceHeader(pdf, poi.Supplier)
	if err != nil {
		return "", err
	}

	pdf.SetLineWidth(0.02)
	pdf.SetDashPattern([]float64{0.1, 0.1}, 0)
	pdf.SetY(2.45)
	pdf.Line(constants.POI_MARGIN, pdf.GetY(), (constants.POI_WIDTH - constants.POI_MARGIN), pdf.GetY())

	pdf.SetDashPattern([]float64{}, 0)

	pdf.SetY(pdf.GetY() + 0.2)

	err = createPurchaseOrderInvoiceInfo(pdf, poi)
	if err != nil {
		return "", err
	}

	startTableY := pdf.GetY() + 0.5

	startTableX, err := createPurchaseOrderInvoiceTableHeader(pdf, startTableY)
	if err != nil {
		return "", err
	}

	itemCount, err := createPurchaseOrderInvoiceData(pdf, startTableX, poi.MedicineLists)
	if err != nil {
		return "", err
	}

	startFooterY := 10.0

	if pdf.GetY() > startFooterY {
		pdf.AddPage()
	}

	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.SetLineWidth(0.02)

	if pdf.PageCount() > 1 {
		pdf.SetPage(1)

		pdf.Line(startTableX["number"], startTableY, startTableX["number"], (constants.POI_HEIGHT - constants.POI_MARGIN))
		pdf.Line(startTableX["item"], startTableY, startTableX["item"], (constants.POI_HEIGHT - constants.POI_MARGIN))
		pdf.Line(startTableX["orderQty"], startTableY, startTableX["orderQty"], (constants.POI_HEIGHT - constants.POI_MARGIN))
		pdf.Line(startTableX["receivedQty"], startTableY, startTableX["receivedQty"], (constants.POI_HEIGHT - constants.POI_MARGIN))
		pdf.Line(startTableX["unit"], startTableY, startTableX["unit"], (constants.POI_HEIGHT - constants.POI_MARGIN))
		pdf.Line(startTableX["end"], startTableY, startTableX["end"], (constants.POI_HEIGHT - constants.POI_MARGIN))

		for i := 1; i < (pdf.PageCount() - 1); i++ {
			pdf.SetPage(i + 1)
			pdf.Line(startTableX["number"], 0.5, startTableX["number"], (constants.POI_HEIGHT - constants.POI_MARGIN))
			pdf.Line(startTableX["item"], 0.5, startTableX["item"], (constants.POI_HEIGHT - constants.POI_MARGIN))
			pdf.Line(startTableX["orderQty"], 0.5, startTableX["orderQty"], (constants.POI_HEIGHT - constants.POI_MARGIN))
			pdf.Line(startTableX["receivedQty"], 0.5, startTableX["receivedQty"], (constants.POI_HEIGHT - constants.POI_MARGIN))
			pdf.Line(startTableX["unit"], 0.5, startTableX["unit"], (constants.POI_HEIGHT - constants.POI_MARGIN))
			pdf.Line(startTableX["end"], 0.5, startTableX["end"], (constants.POI_HEIGHT - constants.POI_MARGIN))
		}

		pdf.SetPage(pdf.PageCount())

		pdf.Line(startTableX["number"], 0.5, startTableX["number"], (startFooterY - 0.3))
		pdf.Line(startTableX["item"], 0.5, startTableX["item"], (startFooterY - 0.3))
		pdf.Line(startTableX["orderQty"], 0.5, startTableX["orderQty"], (startFooterY - 0.3))
		pdf.Line(startTableX["receivedQty"], 0.5, startTableX["receivedQty"], (startFooterY - 0.3))
		pdf.Line(startTableX["unit"], 0.5, startTableX["unit"], (startFooterY - 0.3))
		pdf.Line(startTableX["end"], 0.5, startTableX["end"], (startFooterY - 0.3))
	} else {
		pdf.Line(startTableX["number"], startTableY, startTableX["number"], (startFooterY - 0.3))
		pdf.Line(startTableX["item"], startTableY, startTableX["item"], (startFooterY - 0.3))
		pdf.Line(startTableX["orderQty"], startTableY, startTableX["orderQty"], (startFooterY - 0.3))
		pdf.Line(startTableX["receivedQty"], startTableY, startTableX["receivedQty"], (startFooterY - 0.3))
		pdf.Line(startTableX["unit"], startTableY, startTableX["unit"], (startFooterY - 0.3))
		pdf.Line(startTableX["end"], startTableY, startTableX["end"], (startFooterY - 0.3))
	}

	pdf.Line(startTableX["number"], (startFooterY - 0.3), startTableX["end"], (startFooterY - 0.3))

	pdf.SetDashPattern([]float64{}, 0)

	err = createPurchaseOrderInvoiceFooter(pdf, itemCount, startTableX, startFooterY)
	if err != nil {
		return "", err
	}

	fileName := prevFileName

	if prevFileName == "" {
		fileName := "poi-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
		isExist, err := poiStore.IsPdfUrlExist(fileName)
		if err != nil {
			return "", err
		}

		for isExist {
			fileName = "poi-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
			isExist, err = poiStore.IsPdfUrlExist(fileName)
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

func initPurchaseOrderInvoicePdf() (*fpdf.Fpdf, error) {
	s, _ := filepath.Abs("static/assets/font/")

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "cm",
		SizeStr:        "14x21",
		Size: fpdf.SizeType{
			Wd: constants.POI_WIDTH,
			Ht: constants.POI_HEIGHT,
		},
		FontDirStr: s,
	})

	pdf.SetMargins(0.2, 0.3, 0.2)
	pdf.SetAutoPageBreak(true, constants.POI_MARGIN)

	pdf.AddUTF8Font("Arial", constants.REGULAR, "Arial.TTF")
	pdf.AddUTF8Font("Arial", constants.BOLD, "ArialBD.TTF")
	pdf.AddUTF8Font("Arial", constants.ITALIC, "ArialI.TTF")
	pdf.AddUTF8Font("Calibri", constants.REGULAR, "Calibri.TTF")
	pdf.AddUTF8Font("Calibri", constants.BOLD, "CalibriBold.TTF")
	pdf.AddUTF8Font("Bree", constants.REGULAR, "bree-serif-regular.ttf")
	pdf.AddUTF8Font("Bree", constants.BOLD, "Bree Serif Bold.ttf")

	pdf.AddPage()

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error init purchase order pdf: %v", pdf.Error())
	}

	return pdf, nil
}

func createPurchaseOrderInvoiceHeader(pdf *fpdf.Fpdf, supplier types.SupplierInformationReturnPayload) error {
	var caser = cases.Title(language.Indonesian)

	pdf.Image(config.Envs.CompanyLogoURL, pdf.GetX(), pdf.GetY(), constants.POI_LOGO_WIDTH, constants.POI_LOGO_HEIGHT, false, "", 0, "")

	startBesideLogoX := constants.POI_MARGIN + constants.POI_LOGO_WIDTH + 0.1

	pdf.SetX(startBesideLogoX)
	companyName := strings.ToUpper(config.Envs.CompanyName)

	pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
	pdf.SetFont("Bree", constants.BOLD, 22)
	cellWidth := pdf.GetStringWidth(companyName) + constants.POI_MARGIN
	pdf.CellFormat(cellWidth, 0.65, companyName, "", 1, "L", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_HEADER_FONT_SZ)
	cellWidth = pdf.GetStringWidth(config.Envs.CompanyAddress) + constants.POI_MARGIN
	pdf.CellFormat(cellWidth, constants.POI_HEADER_HEIGHT, config.Envs.CompanyAddress, "", 1, "C", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_HEADER_FONT_SZ)
	phone := fmt.Sprintf("No. Telp: %s | WhatsApp: %s", config.Envs.CompanyPhoneNumber, config.Envs.CompanyWhatsAppNumber)
	cellWidth = pdf.GetStringWidth(phone) + constants.POI_MARGIN
	pdf.CellFormat(cellWidth, constants.POI_HEADER_HEIGHT, phone, "", 1, "L", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_HEADER_FONT_SZ)
	businessRegNumber := fmt.Sprintf("No. SIA: %s", config.Envs.BusinessRegistrationNumber)
	cellWidth = pdf.GetStringWidth(businessRegNumber) + constants.POI_MARGIN
	pdf.CellFormat(cellWidth, constants.POI_HEADER_HEIGHT, businessRegNumber, "", 1, "C", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_HEADER_FONT_SZ)
	pharmacist := fmt.Sprintf("Apoteker: %s", config.Envs.Pharmacist)
	cellWidth = pdf.GetStringWidth(pharmacist) + constants.POI_MARGIN
	pdf.CellFormat(cellWidth, constants.POI_HEADER_HEIGHT, pharmacist, "", 1, "C", false, 0, "")

	pdf.SetXY((constants.POI_WIDTH / 2), 0.3)
	pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.SetFont("Calibri", constants.BOLD, 20)
	pdf.CellFormat(0, 0.65, "Purchase Order", "", 1, "C", false, 0, "")

	startSupplierX := ((constants.POI_WIDTH / 2) - 0.5)
	startSupplierY := (pdf.GetY() + 0.1)

	pdf.SetXY(startSupplierX, startSupplierY)

	pdf.SetFont("Calibri", constants.BOLD, constants.POI_SUPPLIER_FONT_SZ)
	cellWidth = pdf.GetStringWidth("Pemasok") + 0.1
	pdf.CellFormat(cellWidth, constants.POI_STD_CELL_HEIGHT, "Pemasok", "", 0, "L", false, 0, "")

	pdf.SetFont("Calibri", constants.BOLD, constants.POI_SUPPLIER_FONT_SZ)
	cellWidth = pdf.GetStringWidth(":") + constants.POI_MARGIN
	pdf.CellFormat(cellWidth, constants.POI_STD_CELL_HEIGHT, ":", "", 0, "L", false, 0, "")

	startSupplierDataX := pdf.GetX()

	// uppercase the supplier name
	pdf.SetX(startSupplierDataX)
	pdf.SetFont("Arial", constants.REGULAR, constants.POI_SUPPLIER_FONT_SZ)
	supplierData := fmt.Sprintf("%s | T. %s", strings.ToUpper(supplier.Name), supplier.CompanyPhoneNumber)
	pdf.CellFormat(0, constants.POI_STD_CELL_HEIGHT, supplierData, "", 1, "L", false, 0, "")

	// Address
	pdf.SetX(startSupplierDataX)
	pdf.SetFont("Arial", constants.REGULAR, constants.POI_SUPPLIER_FONT_SZ)
	pdf.CellFormat(0, constants.POI_STD_CELL_HEIGHT, supplier.Address, "", 1, "L", false, 0, "")

	// Contact Person
	pdf.SetX(startSupplierDataX)
	pdf.SetFont("Arial", constants.REGULAR, constants.POI_SUPPLIER_FONT_SZ)
	contactPerson := fmt.Sprintf("CP. %s | CP. T. %s", caser.String(supplier.ContactPersonName), supplier.ContactPersonNumber)
	pdf.CellFormat(0, constants.POI_STD_CELL_HEIGHT, contactPerson, "", 1, "L", false, 0, "")

	pdf.SetLineWidth(0.02)
	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.RoundedRect(startSupplierX, startSupplierY, (constants.POI_WIDTH - constants.POI_MARGIN - startSupplierX), (pdf.GetY() - startSupplierY), 0.1, "1234", "D")

	if pdf.Error() != nil {
		return fmt.Errorf("error create invoice pdf header: %v", pdf.Error())
	}

	return nil
}

func createPurchaseOrderInvoiceInfo(pdf *fpdf.Fpdf, poi types.PurchaseOrderPdfPayload) error {
	var caser = cases.Title(language.Indonesian)

	space := 0.5

	// PO Number
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("PO No.:") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_INFO_HEIGHT, "PO No.:", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_STD_FONT_SZ)
		pdf.CellFormat(constants.POI_INFO_NUMBER_WIDTH, constants.POI_INFO_HEIGHT, strconv.Itoa(poi.Number), "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// PO Date
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Tgl. PO: ") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_INFO_HEIGHT, "Tgl. PO: ", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_STD_FONT_SZ)
		pdf.CellFormat(constants.POI_INFO_DATE_WIDTH, constants.POI_INFO_HEIGHT, poi.InvoiceDate.Format("02-01-2006"), "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// Terms
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Terms: ") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_INFO_HEIGHT, "Terms: ", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_STD_FONT_SZ)
		pdf.CellFormat(constants.POI_INFO_TERMS_WIDTH, constants.POI_INFO_HEIGHT, poi.Supplier.Terms, "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// Vendor is Taxable
	{
		vendorIsTaxable := "Yes"
		if !poi.Supplier.VendorIsTaxable {
			vendorIsTaxable = "No"
		}

		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Vendor is Taxable: ") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_INFO_HEIGHT, "Vendor is Taxable: ", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_STD_FONT_SZ)
		pdf.CellFormat(constants.POI_INFO_VENDOR_IS_TAX_WIDTH, constants.POI_INFO_HEIGHT, vendorIsTaxable, "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// Cashier
	{
		cashier := caser.String(poi.UserName)
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Dibuat Oleh: ") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_INFO_HEIGHT, "Dibuat Oleh: ", "LTB", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_STD_FONT_SZ)
		pdf.CellFormat(constants.POI_INFO_CASHIER_WIDTH, constants.POI_INFO_HEIGHT, cashier, "RTB", 1, "L", false, 0, "")
	}

	if pdf.Error() != nil {
		return fmt.Errorf("error create purchase order invoice info: %v", pdf.Error())
	}

	return nil
}

func createPurchaseOrderInvoiceTableHeader(pdf *fpdf.Fpdf, startTableY float64) (map[string]float64, error) {
	pdf.SetLineWidth(0.02)

	pdf.SetY(startTableY)

	numberStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.POI_NO_COL_WIDTH, constants.POI_TABLE_HEIGHT, "No.", "TB", 0, "C", false, 0, "")

	itemStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.POI_ITEM_COL_WIDTH, constants.POI_TABLE_HEIGHT, "Item", "TB", 0, "C", false, 0, "")

	orderQtyStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.POI_QTY_COL_WIDTH, constants.POI_TABLE_HEIGHT, "Order Qty", "TB", 0, "C", false, 0, "")

	receivedQtyStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.POI_QTY_COL_WIDTH, constants.POI_TABLE_HEIGHT, "Received Qty", "TB", 0, "C", false, 0, "")

	unitStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.REGULAR, constants.POI_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.POI_UNIT_COL_WIDTH, constants.POI_TABLE_HEIGHT, "Unit", "TB", 0, "C", false, 0, "")

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error create invoice table header: %v", pdf.Error())
	}

	startX := map[string]float64{
		"number":      numberStartX,
		"item":        itemStartX,
		"orderQty":    orderQtyStartX,
		"receivedQty": receivedQtyStartX,
		"unit":        unitStartX,
		"end":         pdf.GetX(),
	}

	pdf.Ln(-1)

	return startX, nil
}

func createPurchaseOrderInvoiceData(pdf *fpdf.Fpdf, startTableX map[string]float64, medicineLists []types.PurchaseOrderMedicineListPayload) (int, error) {
	var printer = message.NewPrinter(language.Indonesian)

	pdf.SetLineWidth(0.02)
	pdf.SetY(pdf.GetY() + 0.05)

	number := 1
	nextY := pdf.GetY()

	for _, medicine := range medicineLists {
		if (pdf.GetY() + (constants.POI_TABLE_HEIGHT * 2)) > (constants.POI_HEIGHT - constants.POI_MARGIN) {
			pdf.AddPage()

			// change top margin into 0.5
			nextY = 0.5
		}
		startY := nextY

		pdf.SetXY(pdf.GetX(), startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.POI_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.POI_NO_COL_WIDTH, constants.POI_TABLE_HEIGHT, strconv.Itoa(number), "", 0, "C", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_TABLE_DATA_FONT_SZ)
		pdf.MultiCell(constants.POI_ITEM_COL_WIDTH, constants.POI_TABLE_HEIGHT, strings.ToUpper(medicine.MedicineName), "", "L", false)

		nextY = pdf.GetY()

		pdf.SetXY(startTableX["orderQty"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.POI_TABLE_DATA_FONT_SZ)
		qtyString := printer.Sprintf("%.1f", medicine.OrderQty)
		pdf.CellFormat(constants.POI_QTY_COL_WIDTH, constants.POI_TABLE_HEIGHT, qtyString, "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["receivedQty"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.POI_TABLE_DATA_FONT_SZ)
		qtyString = printer.Sprintf("%.1f", medicine.ReceivedQty)
		pdf.CellFormat(constants.POI_QTY_COL_WIDTH, constants.POI_TABLE_HEIGHT, qtyString, "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["unit"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.POI_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.POI_UNIT_COL_WIDTH, constants.POI_TABLE_HEIGHT, strings.ToUpper(medicine.Unit), "", 1, "C", false, 0, "")

		number++
	}

	if pdf.Error() != nil {
		return 0, fmt.Errorf("error create purchase order invoice info: %v", pdf.Error())
	}

	return (number - 1), nil
}

func createPurchaseOrderInvoiceFooter(pdf *fpdf.Fpdf, itemCount int, startTableX map[string]float64, startFooterY float64) error {
	pdf.SetLineWidth(0.02)
	pdf.SetDashPattern([]float64{}, 0)

	// Total Item
	{
		pdf.SetY(startFooterY)
		cellWidth := pdf.GetStringWidth("Total Item: ") + constants.POI_MARGIN
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, "Total Item: ", "TBL", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_STD_FONT_SZ)
		cellWidth = pdf.GetStringWidth(strconv.Itoa(itemCount)) + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, strconv.Itoa(itemCount), "TBR", 1, "L", false, 0, "")
	}

	pdf.SetY(pdf.GetY() + 0.8)

	startPharmacistBoxX := pdf.GetX()
	startPharmacistBoxY := pdf.GetY()

	// Pharmacist
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Apoteker") + 0.05
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, "Apoteker", "", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth = pdf.GetStringWidth(":") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, ":", "", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_STD_FONT_SZ)
		cellWidth = startTableX["orderQty"] - constants.POI_MARGIN - 2.0
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, config.Envs.Pharmacist, "", 1, "L", false, 0, "")
	}

	// Pharmacist License Number
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Apoteker") + 0.05
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, "No. SIPA", "", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth = pdf.GetStringWidth(":") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, ":", "", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.POI_STD_FONT_SZ)
		cellWidth = startTableX["orderQty"] - constants.POI_MARGIN - 2.0
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, config.Envs.PharmacistLicenseNumber, "", 1, "L", false, 0, "")
	}

	pdf.RoundedRect(startPharmacistBoxX, startPharmacistBoxY, (startTableX["orderQty"] - constants.POI_MARGIN - 2.0), (pdf.GetY() - startPharmacistBoxY), 0.1, "1234", "D")

	startSignX := 12.2

	pdf.SetXY(startSignX, startFooterY)

	// Prepared By
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Disiapkan Oleh:") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, "Disiapkan Oleh:", "", 1, "L", false, 0, "")

		pdf.SetX(startSignX)
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth = pdf.GetStringWidth("Tgl:") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, "Tgl:", "", 1, "L", false, 0, "")
	}

	pdf.Rect(startSignX, startFooterY, (constants.POI_WIDTH - startSignX - constants.POI_MARGIN), 1.7, "D")

	pdf.SetXY(startSignX, (startFooterY + 2.0))

	// Approved By
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Disetujui Oleh:") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, "Disetujui Oleh:", "", 1, "L", false, 0, "")

		pdf.SetX(startSignX)
		pdf.SetFont("Calibri", constants.BOLD, constants.POI_STD_FONT_SZ)
		cellWidth = pdf.GetStringWidth("Tgl:") + constants.POI_MARGIN
		pdf.CellFormat(cellWidth, constants.POI_FOOTER_CELL_HEIGHT, "Tgl:", "", 0, "L", false, 0, "")
	}

	pdf.Rect(startSignX, (startFooterY + 2.0), (constants.POI_WIDTH - startSignX - constants.POI_MARGIN), 1.7, "D")

	if pdf.Error() != nil {
		return fmt.Errorf("error create purchase order invoice footer: %v", pdf.Error())
	}

	return nil
}
