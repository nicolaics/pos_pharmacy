package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-pdf/fpdf"
	"github.com/nicolaics/pharmacon/config"
	"github.com/nicolaics/pharmacon/constants"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func CreateReceiptPdf(receipt types.ReceiptPdfPayload, invoiceStore types.InvoiceStore) (string, error) {
	directory, err := filepath.Abs("static/pdf/receipt/")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(directory, 0744); err != nil {
		return "", err
	}

	pdf, err := initReceiptPdf()
	if err != nil {
		return "", err
	}

	err = createReceiptHeader(pdf)
	if err != nil {
		return "", err
	}

	err = createReceiptData(pdf, receipt)
	if err != nil {
		return "", err
	}

	fileName := "r-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
	isExist, err := invoiceStore.IsPdfUrlExist(fileName, "receipt")
	if err != nil {
		return "", err
	}

	for isExist {
		fileName = "r-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
		isExist, err = invoiceStore.IsPdfUrlExist(fileName, "receipt")
		if err != nil {
			return "", err
		}
	}

	err = pdf.OutputFileAndClose(directory + "\\" + fileName)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func initReceiptPdf() (*fpdf.Fpdf, error) {
	s, _ := filepath.Abs("static/assets/font/")

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "cm",
		SizeStr:        "10.8x21.3",
		Size: fpdf.SizeType{
			Wd: constants.RECEIPT_WIDTH,
			Ht: constants.RECEIPT_HEIGHT,
		},
		FontDirStr: s,
	})

	pdf.SetMargins(0.3, 0.3, 0.3)
	pdf.SetAutoPageBreak(true, constants.RECEIPT_MARGIN)

	pdf.AddUTF8Font("Arial", constants.REGULAR, "Arial.TTF")
	pdf.AddUTF8Font("Arial", constants.BOLD, "ArialBD.TTF")
	pdf.AddUTF8Font("Arial", constants.ITALIC, "ArialI.TTF")
	pdf.AddUTF8Font("Bree", constants.REGULAR, "bree-serif-regular.ttf")
	pdf.AddUTF8Font("Bree", constants.BOLD, "Bree Serif Bold.ttf")
	pdf.AddUTF8Font("Calibri", constants.REGULAR, "Calibri.TTF")
	pdf.AddUTF8Font("Calibri", constants.BOLD, "CalibriBold.TTF")

	pdf.AddPage()

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error init receipt pdf: %v", pdf.Error())
	}

	return pdf, nil
}

func createReceiptHeader(pdf *fpdf.Fpdf) error {
	pdf.SetXY((constants.RECEIPT_MARGIN + 0.1), 0.3)

	pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
	pdf.SetDrawColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)

	pdf.TransformBegin()
	pdf.TransformRotate(90, 4.5, 4)

	pdf.Image(config.Envs.CompanyLogoURL, 0.2, 0.0, constants.RECEIPT_LOGO_WIDTH, constants.RECEIPT_LOGO_HEIGHT, false, "", 0, "")

	pdf.SetXY(2.0, 0.0)
	pdf.SetFont("Bree", constants.BOLD, 16)
	pdf.CellFormat(4.5, 0.6, config.Envs.CompanyName, "", 1, "C", false, 0, "")

	pdf.SetX(2.0)
	pdf.SetFont("Calibri", constants.REGULAR, 8)
	pdf.MultiCell(4.5, 0.3, config.Envs.CompanyAddress, "", "C", false)

	pdf.SetXY(2.0, (pdf.GetY() - 0.05))
	pdf.SetFont("Calibri", constants.REGULAR, 8)
	phoneNumber := fmt.Sprintf("No. Telp: %s", config.Envs.CompanyPhoneNumber)
	pdf.CellFormat(4.5, 0.3, phoneNumber, "", 1, "C", false, 0, "")

	pdf.SetX(2.0)
	pdf.SetFont("Calibri", constants.REGULAR, 8)
	whatsApp := fmt.Sprintf("WhatsApp: %s", config.Envs.CompanyWhatsAppNumber)
	pdf.CellFormat(4.5, 0.3, whatsApp, "", 1, "C", false, 0, "")

	pdf.SetDashPattern([]float64{0.3, 0.05}, 0)
	pdf.SetLineWidth(0.05)

	pdf.Line(-4.5, pdf.GetY()+0.3, constants.RECEIPT_HEIGHT, pdf.GetY()+0.3)

	pdf.TransformEnd()

	pdf.SetDashPattern([]float64{}, 0)

	if pdf.Error() != nil {
		return fmt.Errorf("error create receipt pdf header: %v", pdf.Error())
	}

	return nil
}

func createReceiptData(pdf *fpdf.Fpdf, receipt types.ReceiptPdfPayload) error {
	var printer = message.NewPrinter(language.Indonesian)
	var caser = cases.Title(language.Indonesian)

	pdf.SetDrawColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
	pdf.SetDashPattern([]float64{0.02, 0.02}, 0)

	pdf.SetXY(3.0, 0.5)

	// Number
	{
		pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(constants.RECEIPT_CELL_WIDTH_TITLE, constants.RECEIPT_CELL_HEIGHT_TITLE, "No.", "", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0.3, constants.RECEIPT_CELL_HEIGHT_TITLE, ":", "", 0, "L", false, 0, "")

		pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
		pdf.SetFont("Arial", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0, constants.RECEIPT_CELL_HEIGHT_BODY, strconv.Itoa(receipt.Number), "B", 1, "LT", false, 0, "")
	}

	// Patient
	{
		pdf.SetXY(3.0, (pdf.GetY() + 0.5))
		pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(constants.RECEIPT_CELL_WIDTH_TITLE, constants.RECEIPT_CELL_HEIGHT_TITLE, "Sudah Terima Dari", "", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0.3, constants.RECEIPT_CELL_HEIGHT_TITLE, ":", "", 0, "L", false, 0, "")

		pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
		pdf.SetFont("Arial", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0, constants.RECEIPT_CELL_HEIGHT_BODY, caser.String(receipt.Patient), "B", 1, "LT", false, 0, "")
	}

	// Received Amount String
	{
		pdf.SetXY(3.0, (pdf.GetY() + 0.5))
		pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(constants.RECEIPT_CELL_WIDTH_TITLE, constants.RECEIPT_CELL_HEIGHT_TITLE, "Banyaknya Uang", "", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0.3, constants.RECEIPT_CELL_HEIGHT_TITLE, ":", "", 0, "L", false, 0, "")

		pdf.SetFillColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetAlpha(0.1, "Screen")
		pdf.Rect(pdf.GetX(), (pdf.GetY() - 0.1), (constants.RECEIPT_HEIGHT - constants.RECEIPT_MARGIN - pdf.GetX()), (constants.RECEIPT_CELL_HEIGHT_BODY + 0.2), "F")

		pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
		pdf.SetAlpha(1.0, "")
		pdf.SetFont("Arial", constants.REGULAR, (constants.RECEIPT_DATA_FONT_SZ - 2))
		pdf.CellFormat(0, constants.RECEIPT_CELL_HEIGHT_BODY, caser.String(receipt.ReceivedAmountString), "", 1, "LM", false, 0, "")
	}

	// Prescription Details String
	{
		pdf.SetXY(3.0, (pdf.GetY() + 0.5))
		pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0, constants.RECEIPT_CELL_HEIGHT_TITLE, "Untuk Pembayaran Obat-obatan menurut Resep:", "", 1, "L", false, 0, "")
	}

	// Doctor and Prescription Number
	{
		pdf.SetXY(3.0, (pdf.GetY() + 0.5))
		pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0.6, constants.RECEIPT_CELL_HEIGHT_TITLE, "Dr.", "", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0.3, constants.RECEIPT_CELL_HEIGHT_TITLE, ":", "", 0, "L", false, 0, "")

		pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
		pdf.SetFont("Arial", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(9, constants.RECEIPT_CELL_HEIGHT_BODY, caser.String(receipt.Doctor), "B", 0, "LT", false, 0, "")

		pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0.7, constants.RECEIPT_CELL_HEIGHT_TITLE, "No.", "", 0, "L", false, 0, "")

		pdf.SetFont("Calibri", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0.3, constants.RECEIPT_CELL_HEIGHT_TITLE, ":", "", 0, "L", false, 0, "")

		pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
		pdf.SetFont("Arial", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		pdf.CellFormat(0, constants.RECEIPT_CELL_HEIGHT_BODY, strconv.Itoa(receipt.PrescriptionNumber), "B", 1, "LT", false, 0, "")
	}

	// Date
	{
		pdf.SetXY(14.5, 6.5)

		pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
		pdf.SetFont("Arial", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)

		text := fmt.Sprintf("Jakarta, %s", receipt.Date.Format("02-01-2006"))

		pdf.CellFormat(0, constants.RECEIPT_CELL_HEIGHT_BODY, text, "", 0, "L", false, 0, "")
	}

	// Received Amount
	{
		pdf.SetXY(3.0, 9.0)

		pdf.SetTextColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetFont("Calibri", constants.REGULAR, (constants.RECEIPT_DATA_FONT_SZ + 2))
		pdf.CellFormat(2.4, 0.5, "Jumlah Rp. ", "", 0, "L", false, 0, "")

		pdf.SetFillColor(constants.GREEN_R, constants.GREEN_G, constants.GREEN_B)
		pdf.SetAlpha(0.1, "Screen")
		pdf.Rect(pdf.GetX(), (pdf.GetY() - 0.1), 7.0, (constants.RECEIPT_CELL_HEIGHT_BODY + 0.2), "F")

		pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
		pdf.SetAlpha(1.0, "")
		pdf.SetFont("Arial", constants.REGULAR, constants.RECEIPT_DATA_FONT_SZ)
		receivedAmount := printer.Sprintf("%.1f", receipt.ReceivedAmount)
		pdf.CellFormat(7, constants.RECEIPT_CELL_HEIGHT_BODY, receivedAmount, "", 1, "LM", false, 0, "")
	}

	if pdf.Error() != nil {
		return fmt.Errorf("error create receipt data: %v", pdf.Error())
	}

	return nil
}
