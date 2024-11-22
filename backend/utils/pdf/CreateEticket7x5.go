package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/nicolaics/pharmacon/constants"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Eticket7X5 struct {
	Print       bool
	Qty         int
	MedicineQty int
	Number      int
	Dose        string
}

func CreateEticket7x5PDF(eticket types.EticketPDFReturnPayload, setNumber int, prescStore types.PrescriptionStore) (string, error) {
	directory, err := filepath.Abs("static/pdf/eticket/")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(directory, 0744); err != nil {
		return "", err
	}

	pdf, err := initEticket7x5Pdf()
	if err != nil {
		return "", err
	}

	err = createEtix7x5Data(pdf, eticket, setNumber)
	if err != nil {
		return "", err
	}

	fileName := "e-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
	isExist, err := prescStore.IsPDFUrlExist("eticket", fileName)
	if err != nil {
		return "", err
	}

	for isExist {
		fileName = "e-" + utils.GenerateRandomCodeAlphanumeric(8) + "-" + utils.GenerateRandomCodeAlphanumeric(8) + ".pdf"
		isExist, err = prescStore.IsPDFUrlExist("eticket", fileName)
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

func initEticket7x5Pdf() (*fpdf.Fpdf, error) {
	s, _ := filepath.Abs("static/assets/font/")

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "cm",
		SizeStr:        "5x7",
		Size: fpdf.SizeType{
			Wd: constants.ETIX_7X5_WIDTH,
			Ht: constants.ETIX_7X5_HEIGHT,
		},
		FontDirStr: s,
	})

	pdf.SetMargins(constants.ETIX_MARGIN, constants.ETIX_MARGIN, constants.ETIX_MARGIN)
	pdf.SetAutoPageBreak(false, constants.ETIX_MARGIN)

	pdf.AddUTF8Font("Arial", constants.REGULAR, "Arial.TTF")
	pdf.AddUTF8Font("Arial", constants.BOLD, "ArialBD.TTF")

	pdf.AddPage()

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error init eticket 7x5 pdf: %v", pdf.Error())
	}

	return pdf, nil
}

func createEtix7x5Data(pdf *fpdf.Fpdf, eticket types.EticketPDFReturnPayload, setNumber int) error {
	caser := cases.Title(language.Indonesian)

	pdf.SetLineWidth(0.02)
	pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
	pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())

	// Number
	{
		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
		number := fmt.Sprintf("No.  %d-%d", eticket.Number, setNumber)
		pdf.CellFormat(0, constants.ETIX_7X5_STD_CELL_HEIGHT, number, "", 1, "L", false, 0, "")
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())

	// Date and Time
	{
		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
		dateTime := fmt.Sprintf("Tgl.  %s", time.Now().Format("02-01-2006"))
		pdf.CellFormat(3.0, constants.ETIX_7X5_STD_CELL_HEIGHT, dateTime, "", 0, "L", false, 0, "")

		pdf.SetFont("Arial", constants.REGULAR, (constants.ETIX_7X5_STD_FONT_SZ - 2))
		pdf.CellFormat(0, constants.ETIX_7X5_STD_CELL_HEIGHT, time.Now().Format("15:04"), "", 1, "LM", false, 0, "")
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())

	// Name
	{
		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
		pdf.CellFormat(1.1, (constants.ETIX_7X5_STD_CELL_HEIGHT * 3), "Nama:", "", 0, "LM", false, 0, "")

		nameSplit := strings.Split(eticket.PatientName, " ")
		startName := pdf.GetX()
		pdf.SetFont("Arial", constants.BOLD, (constants.ETIX_7X5_STD_FONT_SZ + 2))

		// fmt.Println(constants.ETIX_7X5_WIDTH - (2 * constants.ETIX_MARGIN) - startName)

		if len(nameSplit) == 1 {
			pdf.CellFormat(0, (0.6 * 3), eticket.PatientName, "", 1, "CM", false, 0, "")
		} else if len(nameSplit) == 2 {
			pdf.CellFormat(0, ((0.6 * 3) / 2), nameSplit[0], "", 1, "CB", false, 0, "")

			pdf.SetX(startName)
			pdf.CellFormat(0, ((0.6 * 3) / 2), nameSplit[1], "", 1, "CT", false, 0, "")
		} else if len(nameSplit) > 2 {
			pdf.CellFormat(0, 0.6, nameSplit[0], "", 1, "CB", false, 0, "")

			pdf.SetX(startName)
			pdf.CellFormat(0, 0.6, nameSplit[1], "", 1, "CM", false, 0, "")

			pdf.SetX(startName)
			pdf.CellFormat(0, 0.6, nameSplit[2], "", 1, "CT", false, 0, "")
		}
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())

	// Usage
	{
		eticket.SetUsage = caser.String(eticket.SetUsage)

		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
		pdf.CellFormat(0, constants.ETIX_7X5_STD_CELL_HEIGHT, eticket.SetUsage, "", 1, "CM", false, 0, "")
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())

	// Dose
	{
		eticket.SetUnit = caser.String(eticket.SetUnit)

		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)

		slashIdx := strings.Index(eticket.SetUnit, "/")
		if slashIdx != -1 {
			pdf.SetX(0.7)
			pdf.CellFormat(1.15, constants.ETIX_7X5_STD_CELL_HEIGHT, "Sehari", "", 0, "L", false, 0, "")

			doseSplit := strings.Split(eticket.Dose, "/")
			doseSplit[1] = strings.TrimSpace(doseSplit[1])

			aDay := strings.Split(doseSplit[0], "x")
			aDay[0] = strings.TrimSpace(aDay[0])
			aDay[1] = strings.TrimSpace(aDay[1])

			cellWidth := pdf.GetStringWidth(fmt.Sprintf("%s x ", aDay[0]))
			pdf.CellFormat(cellWidth, constants.ETIX_7X5_STD_CELL_HEIGHT, fmt.Sprintf("%s x ", aDay[0]), "", 0, "L", false, 0, "")

			pdf.SubWrite(constants.PRESC_STD_CELL_HEIGHT, aDay[1], 5, 0.4, 0, "")

			pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
			pdf.CellFormat(pdf.GetStringWidth("/"), constants.ETIX_7X5_STD_CELL_HEIGHT, "/", "", 0, "L", false, 0, "")

			pdf.SubWrite(constants.PRESC_STD_CELL_HEIGHT, doseSplit[1], 5, -4.8, 0, "")

			pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
			pdf.SetX(pdf.GetX() + 0.1)
			pdf.CellFormat(0, constants.ETIX_7X5_STD_CELL_HEIGHT, eticket.SetUnit, "", 1, "L", false, 0, "")
		} else {
			doseTxt := fmt.Sprintf("Sehari %s %s", strings.ToLower(eticket.Dose), strings.ToLower(eticket.SetUnit))

			pdf.CellFormat(0, constants.ETIX_7X5_STD_CELL_HEIGHT, doseTxt, "", 1, "C", false, 0, "")
		}
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())

	// Consume Time
	{
		var consumeTime string
		if eticket.ConsumeTime == "ac" {
			consumeTime = "Sebelum Makan"
		} else {
			consumeTime = "Setelah Makan"
		}

		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
		pdf.CellFormat(0, constants.ETIX_7X5_STD_CELL_HEIGHT, consumeTime, "", 1, "C", false, 0, "")
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())

	// Must Finish
	{
		if eticket.MustFinish {
			pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
			pdf.CellFormat(0, constants.ETIX_7X5_STD_CELL_HEIGHT, "HABISKAN", "", 1, "C", false, 0, "")

			pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())
		}
	}

	// Qty
	{
		qtyTxt := fmt.Sprintf("Qty: %.0f", eticket.MedicineQty)
		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X5_STD_FONT_SZ)
		pdf.CellFormat(0, constants.ETIX_7X5_STD_CELL_HEIGHT, qtyTxt, "", 1, "C", false, 0, "")
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X5_WIDTH, pdf.GetY())

	if pdf.Error() != nil {
		return fmt.Errorf("error create eticket 7x5 data: %v", pdf.Error())
	}

	return nil
}
