package pdf

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/nicolaics/pharmacon/config"
	"github.com/nicolaics/pharmacon/constants"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func CreateProductionPdf(prod types.ProductionPdfPayload, prevFileName string, productionStore types.ProductionStore) (string, error) {
	directory := "static/pdf/production/"
	if err := os.MkdirAll(directory, 0744); err != nil {
		return "", err
	}

	pdf, err := initProductionPdf()
	if err != nil {
		return "", err
	}

	err = createProductionHeader(pdf, prod.Description)
	if err != nil {
		return "", err
	}

	pdf.SetLineWidth(0.02)
	pdf.SetDashPattern([]float64{0.1, 0.1}, 0)
	pdf.SetY(2.45)
	pdf.Line(constants.PRODUCTION_MARGIN, pdf.GetY(), (constants.PRODUCTION_WIDTH - constants.PRODUCTION_MARGIN), pdf.GetY())

	pdf.SetDashPattern([]float64{}, 0)

	pdf.SetY(pdf.GetY() + 0.2)

	err = createProductionInfo(pdf, prod)
	if err != nil {
		return "", err
	}

	startTableY := pdf.GetY() + 0.5

	startTableX, err := createProductionTableHeader(pdf, startTableY)
	if err != nil {
		return "", err
	}

	_, err = createProductionData(pdf, startTableX, prod.MedicineLists)
	if err != nil {
		return "", err
	}

	startFooterY := 12.5

	if pdf.GetY() > startFooterY {
		pdf.AddPage()
	}

	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.SetLineWidth(0.02)

	if pdf.PageCount() > 1 {
		pdf.SetPage(1)

		pdf.Line(startTableX["number"], startTableY, startTableX["number"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
		pdf.Line(startTableX["barcode"], startTableY, startTableX["barcode"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
		pdf.Line(startTableX["item"], startTableY, startTableX["item"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
		pdf.Line(startTableX["qty"], startTableY, startTableX["qty"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
		pdf.Line(startTableX["unit"], startTableY, startTableX["unit"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
		pdf.Line(startTableX["cost"], startTableY, startTableX["cost"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
		pdf.Line(startTableX["end"], startTableY, startTableX["end"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))

		for i := 1; i < (pdf.PageCount() - 1); i++ {
			pdf.SetPage(i + 1)
			pdf.Line(startTableX["number"], 0.5, startTableX["number"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
			pdf.Line(startTableX["barcode"], 0.5, startTableX["barcode"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
			pdf.Line(startTableX["item"], 0.5, startTableX["item"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
			pdf.Line(startTableX["qty"], 0.5, startTableX["qty"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
			pdf.Line(startTableX["unit"], 0.5, startTableX["unit"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
			pdf.Line(startTableX["cost"], 0.5, startTableX["cost"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
			pdf.Line(startTableX["end"], 0.5, startTableX["end"], (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN))
		}

		pdf.SetPage(pdf.PageCount())

		pdf.Line(startTableX["number"], 0.5, startTableX["number"], (startFooterY - 0.5))
		pdf.Line(startTableX["barcode"], 0.5, startTableX["barcode"], (startFooterY - 0.5))
		pdf.Line(startTableX["item"], 0.5, startTableX["item"], (startFooterY - 0.5))
		pdf.Line(startTableX["qty"], 0.5, startTableX["qty"], (startFooterY - 0.5))
		pdf.Line(startTableX["unit"], 0.5, startTableX["unit"], (startFooterY - 0.5))
		pdf.Line(startTableX["cost"], 0.5, startTableX["cost"], (startFooterY - 0.5))
		pdf.Line(startTableX["end"], 0.5, startTableX["end"], (startFooterY - 0.5))
	} else {
		pdf.Line(startTableX["number"], startTableY, startTableX["number"], (startFooterY - 0.5))
		pdf.Line(startTableX["barcode"], startTableY, startTableX["barcode"], (startFooterY - 0.5))
		pdf.Line(startTableX["item"], startTableY, startTableX["item"], (startFooterY - 0.5))
		pdf.Line(startTableX["qty"], startTableY, startTableX["qty"], (startFooterY - 0.5))
		pdf.Line(startTableX["unit"], startTableY, startTableX["unit"], (startFooterY - 0.5))
		pdf.Line(startTableX["cost"], startTableY, startTableX["cost"], (startFooterY - 0.5))
		pdf.Line(startTableX["end"], startTableY, startTableX["end"], (startFooterY - 0.5))
	}

	pdf.Line(startTableX["number"], (startFooterY - 0.5), startTableX["end"], (startFooterY - 0.5))

	pdf.SetDashPattern([]float64{}, 0)

	err = createProductionFooter(pdf, startTableX, startFooterY, prod.TotalCost)
	if err != nil {
		return "", err
	}

	fileName := prevFileName

	if prevFileName == "" {
		fileName = "prod-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
		isExist, err := productionStore.IsPdfUrlExist(fileName)
		if err != nil {
			return "", err
		}

		for isExist {
			fileName = "prod-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
			isExist, err = productionStore.IsPdfUrlExist(fileName)
			if err != nil {
				return "", err
			}
		}
	}

	err = pdf.OutputFileAndClose(directory + fileName)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func initProductionPdf() (*fpdf.Fpdf, error) {
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "cm",
		SizeStr:        "14x21",
		Size: fpdf.SizeType{
			Wd: constants.PRODUCTION_WIDTH,
			Ht: constants.PRODUCTION_HEIGHT,
		},
		FontDirStr: "static/assets/font/",
	})

	pdf.SetMargins(0.2, 0.3, 0.2)
	pdf.SetAutoPageBreak(true, constants.PRODUCTION_MARGIN)

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

func createProductionHeader(pdf *fpdf.Fpdf, description string) error {
	pdf.Image(config.Envs.CompanyLogoURL, pdf.GetX(), pdf.GetY(), constants.PRODUCTION_LOGO_WIDTH, constants.PRODUCTION_LOGO_HEIGHT, false, "", 0, "")

	startBesideLogoX := constants.PRODUCTION_MARGIN + constants.PRODUCTION_LOGO_WIDTH + 0.1

	pdf.SetX(startBesideLogoX)

	pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
	pdf.SetFont("Bree", constants.BOLD, 22)
	cellWidth := pdf.GetStringWidth(config.Envs.CompanyName) + constants.PRODUCTION_MARGIN
	pdf.CellFormat(cellWidth, 0.65, config.Envs.CompanyName, "", 1, "L", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.PRODUCTION_HEADER_FONT_SZ)
	cellWidth = pdf.GetStringWidth(config.Envs.CompanyAddress) + constants.PRODUCTION_MARGIN
	pdf.CellFormat(cellWidth, constants.PRODUCTION_HEADER_HEIGHT, config.Envs.CompanyAddress, "", 1, "C", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.PRODUCTION_HEADER_FONT_SZ)
	contact := fmt.Sprintf("No. Telp: %s | WhatsApp: %s", config.Envs.CompanyPhoneNumber, config.Envs.CompanyWhatsAppNumber)
	cellWidth = pdf.GetStringWidth(contact) + constants.PRODUCTION_MARGIN
	pdf.CellFormat(cellWidth, constants.PRODUCTION_HEADER_HEIGHT, contact, "", 1, "L", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.PRODUCTION_HEADER_FONT_SZ)
	businessReg := fmt.Sprintf("No. SIA: %s", config.Envs.BusinessRegistrationNumber)
	cellWidth = pdf.GetStringWidth(businessReg) + constants.PRODUCTION_MARGIN
	pdf.CellFormat(cellWidth, constants.PRODUCTION_HEADER_HEIGHT, businessReg, "", 1, "C", false, 0, "")

	pdf.SetX(startBesideLogoX)
	pdf.SetFont("Calibri", constants.REGULAR, constants.PRODUCTION_HEADER_FONT_SZ)
	pharmacist := fmt.Sprintf("Apoteker: %s", config.Envs.Pharmacist)
	cellWidth = pdf.GetStringWidth(pharmacist) + constants.PRODUCTION_MARGIN
	pdf.CellFormat(cellWidth, constants.PRODUCTION_HEADER_HEIGHT, pharmacist, "", 1, "C", false, 0, "")

	pdf.SetXY((constants.PRODUCTION_WIDTH / 2), 0.3)
	pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.SetFont("Calibri", constants.BOLD, 20)
	pdf.CellFormat(0, 0.65, "Job Costing", "", 1, "C", false, 0, "")

	startDescX := ((constants.PRODUCTION_WIDTH / 2) - 0.5)
	startDescY := (pdf.GetY() + 0.1)

	pdf.SetXY(startDescX, startDescY)

	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_STD_FONT_SZ)
	cellWidth = pdf.GetStringWidth("Deskripsi")
	pdf.CellFormat(cellWidth, constants.PRODUCTION_STD_CELL_HEIGHT, "Deskripsi", "", 0, "L", false, 0, "")

	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_STD_FONT_SZ)
	cellWidth = pdf.GetStringWidth(":") + constants.PRODUCTION_MARGIN
	pdf.CellFormat(cellWidth, constants.PRODUCTION_STD_CELL_HEIGHT, ":", "", 0, "L", false, 0, "")

	startDescDataX := pdf.GetX()
	newY := pdf.GetY() + constants.PRODUCTION_STD_CELL_HEIGHT*4

	pdf.SetX(startDescDataX)
	pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_STD_FONT_SZ)
	pdf.MultiCell(0, constants.PRODUCTION_STD_CELL_HEIGHT, description, "", "L", false)

	pdf.SetLineWidth(0.02)
	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.RoundedRect(startDescX, startDescY, (constants.PRODUCTION_WIDTH - constants.PRODUCTION_MARGIN - startDescX), (newY - pdf.GetY()), 0.1, "1234", "D")

	pdf.SetY(newY)

	if pdf.Error() != nil {
		return fmt.Errorf("error create production pdf header: %v", pdf.Error())
	}

	return nil
}

func createProductionInfo(pdf *fpdf.Fpdf, prod types.ProductionPdfPayload) error {
	var caser = cases.Title(language.Indonesian)

	space := 0.5

	normalWidth := (constants.PRODUCTION_WIDTH - (2 * constants.PRODUCTION_MARGIN) - (2 * space)) / 3

	// Production Number
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("No.:") + constants.PRODUCTION_MARGIN
		pdf.CellFormat(cellWidth, constants.PRODUCTION_INFO_HEIGHT, "No.:", "LTB", 0, "L", false, 0, "")

		cellWidth = normalWidth - cellWidth

		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_STD_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.PRODUCTION_INFO_HEIGHT, strconv.Itoa(prod.Number), "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// Production Date
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Tgl.: ") + constants.PRODUCTION_MARGIN
		pdf.CellFormat(cellWidth, constants.PRODUCTION_INFO_HEIGHT, "Tgl.: ", "LTB", 0, "L", false, 0, "")

		cellWidth = normalWidth - cellWidth
		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_STD_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.PRODUCTION_INFO_HEIGHT, prod.ProductionDate.Format("02-01-2006"), "RTB", 0, "L", false, 0, "")
	}

	pdf.SetX(pdf.GetX() + space)

	// Cashier
	{
		pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_STD_FONT_SZ)
		cellWidth := pdf.GetStringWidth("Dibuat Oleh: ") + constants.PRODUCTION_MARGIN
		pdf.CellFormat(cellWidth, constants.PRODUCTION_INFO_HEIGHT, "Dibuat Oleh: ", "LTB", 0, "L", false, 0, "")

		cellWidth = normalWidth - cellWidth

		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_STD_FONT_SZ)
		pdf.CellFormat(cellWidth, constants.PRODUCTION_INFO_HEIGHT, caser.String(prod.UserName), "RTB", 1, "L", false, 0, "")
	}

	if pdf.Error() != nil {
		return fmt.Errorf("error create production info: %v", pdf.Error())
	}

	return nil
}

func createProductionTableHeader(pdf *fpdf.Fpdf, startTableY float64) (map[string]float64, error) {
	pdf.SetLineWidth(0.02)

	pdf.SetY(startTableY)

	numberStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PRODUCTION_NO_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, "No.", "TB", 0, "C", false, 0, "")

	barcodeStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PRODUCTION_BARCODE_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, "Barcode", "TB", 0, "C", false, 0, "")

	itemStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PRODUCTION_ITEM_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, "Item", "TB", 0, "C", false, 0, "")

	qtyStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PRODUCTION_QTY_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, "Qty", "TB", 0, "C", false, 0, "")

	unitStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PRODUCTION_UNIT_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, "Unit", "TB", 0, "C", false, 0, "")

	costStartX := pdf.GetX()
	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_TABLE_HEADER_FONT_SZ)
	pdf.CellFormat(constants.PRODUCTION_COST_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, "Cost (Rp.)", "TB", 0, "C", false, 0, "")

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error create production table header: %v", pdf.Error())
	}

	startX := map[string]float64{
		"number":  numberStartX,
		"barcode": barcodeStartX,
		"item":    itemStartX,
		"qty":     qtyStartX,
		"unit":    unitStartX,
		"cost":    costStartX,
		"end":     pdf.GetX(),
	}

	pdf.Ln(-1)

	return startX, nil
}

func createProductionData(pdf *fpdf.Fpdf, startTableX map[string]float64, medicineLists []types.ProductionMedicineListPayload) (int, error) {
	var printer = message.NewPrinter(language.Indonesian)

	pdf.SetLineWidth(0.02)
	pdf.SetY(pdf.GetY() + 0.05)

	number := 1
	nextY := pdf.GetY()

	for _, medicine := range medicineLists {
		if (pdf.GetY() + (constants.PRODUCTION_TABLE_HEIGHT * 2)) > (constants.PRODUCTION_HEIGHT - constants.PRODUCTION_MARGIN) {
			pdf.AddPage()

			// change top margin into 0.5
			nextY = 0.5
		}
		startY := nextY

		pdf.SetXY(pdf.GetX(), startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.PRODUCTION_NO_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, strconv.Itoa(number), "", 0, "C", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.PRODUCTION_BARCODE_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, medicine.MedicineBarcode, "", 0, "C", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_TABLE_DATA_FONT_SZ)
		pdf.MultiCell(constants.PRODUCTION_ITEM_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, strings.ToUpper(medicine.MedicineName), "", "L", false)

		nextY = pdf.GetY()

		pdf.SetXY(startTableX["qty"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_TABLE_DATA_FONT_SZ)
		qtyString := printer.Sprintf("%.1f", medicine.Qty)
		pdf.CellFormat(constants.PRODUCTION_QTY_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, qtyString, "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["unit"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_TABLE_DATA_FONT_SZ)
		pdf.CellFormat(constants.PRODUCTION_UNIT_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, strings.ToUpper(medicine.Unit), "", 0, "C", false, 0, "")

		pdf.SetXY(startTableX["cost"], startY)
		pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_TABLE_DATA_FONT_SZ)
		costString := printer.Sprintf("%.1f", medicine.Cost)
		pdf.CellFormat(constants.PRODUCTION_COST_COL_WIDTH, constants.PRODUCTION_TABLE_HEIGHT, costString, "", 1, "R", false, 0, "")

		number++
	}

	if pdf.Error() != nil {
		return 0, fmt.Errorf("error create production info: %v", pdf.Error())
	}

	return (number - 1), nil
}

func createProductionFooter(pdf *fpdf.Fpdf, startTableX map[string]float64, startFooterY float64, totalCost float64) error {
	var printer = message.NewPrinter(language.Indonesian)

	pdf.SetLineWidth(0.02)
	pdf.SetDashPattern([]float64{}, 0)

	pdf.SetXY(startTableX["qty"], (startFooterY - 0.2))

	cellWidth := startTableX["cost"] - startTableX["qty"]

	pdf.SetFont("Calibri", constants.BOLD, constants.PRODUCTION_STD_FONT_SZ)
	pdf.CellFormat(cellWidth, constants.PRODUCTION_FOOTER_CELL_HEIGHT, "Total Cost (Rp.):", "", 0, "RM", false, 0, "")

	pdf.SetX(startTableX["cost"])
	pdf.SetFont("Arial", constants.REGULAR, constants.PRODUCTION_STD_FONT_SZ)
	totalCostString := printer.Sprintf("%.1f", totalCost)
	pdf.CellFormat(0, constants.PRODUCTION_FOOTER_CELL_HEIGHT, totalCostString, "", 1, "CM", false, 0, "")

	pdf.SetLineWidth(0.02)
	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.RoundedRect(startTableX["qty"], (startFooterY - 0.2),
		(constants.PRODUCTION_WIDTH - constants.PRODUCTION_MARGIN - startTableX["qty"]),
		constants.PRODUCTION_FOOTER_CELL_HEIGHT,
		0, "1234", "D")

	if pdf.Error() != nil {
		return fmt.Errorf("error create production footer: %v", pdf.Error())
	}

	return nil
}
