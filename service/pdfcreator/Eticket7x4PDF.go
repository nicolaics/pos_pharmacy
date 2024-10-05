package pdfcreator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/nicolaics/pos_pharmacy/constants"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

func CreateEticket7x4(eticket types.EticketPDFReturnPayload, setNumber int, prescStore types.PrescriptionStore, prevFileName string) (string, error) {
	directory := "../static/pdf/eticket"
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", err
	}

	s, _ := filepath.Abs("static/assets/font/")
	log.Println(s)

	pdf, err := initEticketPdf()
	if err != nil {
		return "", err
	}

    err = createEtix7x4Data(pdf, eticket, setNumber)
    if err != nil {
		return "", err
	}

	fileName := prevFileName
    
    if prevFileName != "" {
        fileName := "e-" + utils.GenerateRandomCodeAlphanumeric(6) + "-" + utils.GenerateRandomCodeAlphanumeric(6) + ".pdf"
        isExist, err := prescStore.IsPDFUrlExist("eticket", fileName)
        if err != nil {
            return "", err
        }

        for isExist {
            fileName = "e-" + utils.GenerateRandomCodeAlphanumeric(6) + "-" + utils.GenerateRandomCodeAlphanumeric(6) + ".pdf"
            isExist, err = prescStore.IsPDFUrlExist("eticket", fileName)
            if err != nil {
                return "", err
            }
        }
    }

	err = pdf.OutputFileAndClose(fileName)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func initEticketPdf() (*fpdf.Fpdf, error) {
	s, _ := filepath.Abs("static/assets/font/")

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "cm",
		SizeStr:        "4x7",
		Size: fpdf.SizeType{
			Wd: constants.ETIX_7X4_WIDTH,
			Ht: constants.ETIX_7X4_HEIGHT,
		},
		FontDirStr: s,
	})

	pdf.SetMargins(constants.ETIX_MARGIN, constants.ETIX_MARGIN, constants.ETIX_MARGIN)
	pdf.SetAutoPageBreak(false, constants.ETIX_MARGIN)

	pdf.AddUTF8Font("Arial", constants.REGULAR, "Arial.TTF")
	pdf.AddUTF8Font("Arial", constants.BOLD, "ArialBD.TTF")

	pdf.AddPage()

	if pdf.Error() != nil {
		return nil, fmt.Errorf("error init eticket 7x4 pdf: %v", pdf.Error())
	}

	return pdf, nil
}

func createEtix7x4Data(pdf *fpdf.Fpdf, eticket types.EticketPDFReturnPayload, setNumber int) error {
	pdf.SetLineWidth(0.02)
    pdf.SetDrawColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)
    pdf.SetTextColor(constants.BLACK_R, constants.BLACK_G, constants.BLACK_B)

    pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())

    // Number
    {
        pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)
        number := fmt.Sprintf("No.  %d-%d", eticket.Number, setNumber)
        pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, number, "", 1, "L", false, 0, "")
    }

    pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())

    // Date and Time
    {
        pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)
        dateTime := fmt.Sprintf("Tgl.  %s", time.Now().Format("02-01-2006"))
        pdf.CellFormat(2.8, constants.ETIX_7X4_STD_CELL_HEIGHT, dateTime, "", 0, "L", false, 0, "")

        pdf.SetFont("Arial", constants.REGULAR, (constants.ETIX_7X4_STD_FONT_SZ - 2))
        pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, time.Now().Format("15:04"), "", 1, "LM", false, 0, "")
    }

    pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())

    // Name
    {
        pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)
        pdf.CellFormat(1.1, (constants.ETIX_7X4_STD_CELL_HEIGHT * 3), "Nama:", "", 0, "LM", false, 0, "")

		nameSplit := strings.Split(eticket.PatientName, " ")
		startName := pdf.GetX()
		pdf.SetFont("Arial", constants.BOLD, (constants.ETIX_7X4_STD_FONT_SZ + 2))

		if len(nameSplit) == 1 {
			pdf.CellFormat(0, (constants.ETIX_7X4_STD_CELL_HEIGHT * 3), eticket.PatientName, "", 1, "CM", false, 0, "")
		} else if len(nameSplit) == 2 {
			pdf.CellFormat(0, ((constants.ETIX_7X4_STD_CELL_HEIGHT * 3) / 2), nameSplit[0], "", 1, "CB", false, 0, "")

			pdf.SetX(startName)
			pdf.CellFormat(0, ((constants.ETIX_7X4_STD_CELL_HEIGHT * 3) / 2), nameSplit[1], "", 1, "CT", false, 0, "")
		} else if len(nameSplit) > 2 {
			pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, nameSplit[0], "", 1, "CB", false, 0, "")

			pdf.SetX(startName)
			pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, nameSplit[1], "", 1, "CM", false, 0, "")

			pdf.SetX(startName)
			pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, nameSplit[2], "", 1, "CT", false, 0, "")
		}
    }
    
    pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())

    // Usage
    {

		usageSplit := strings.Split(eticket.SetUsage, " ")
		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)

		if len(usageSplit) > 2 {
			pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, (usageSplit[0] + " " + usageSplit[1]), "", 1, "CB", false, 0, "")
			pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, usageSplit[2], "", 1, "CT", false, 0, "")
		} else {
			pdf.CellFormat(0, (constants.ETIX_7X4_STD_CELL_HEIGHT * 2), eticket.SetUsage, "", 1, "CM", false, 0, "")
		}
    }

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())

	// Dose
	{
		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)

		slashIdx := strings.Index(eticket.Dose, "/")
		if slashIdx != -1 {
			pdf.SetX(0.2)
			pdf.CellFormat(1.15, constants.ETIX_7X4_STD_CELL_HEIGHT, "Sehari", "", 0, "L", false, 0, "")

			doseSplit := strings.Split(eticket.Dose, "/")
			doseSplit[1] = strings.TrimSpace(doseSplit[1])

			aDay := strings.Split(doseSplit[0], "x")
			aDay[0] = strings.TrimSpace(aDay[0])
			aDay[1] = strings.TrimSpace(aDay[1])
			
			cellWidth := pdf.GetStringWidth(fmt.Sprintf("%s x ", aDay[0]))
			pdf.CellFormat(cellWidth, constants.ETIX_7X4_STD_CELL_HEIGHT, fmt.Sprintf("%s x ", aDay[0]), "", 0, "L", false, 0, "")

			pdf.SubWrite(constants.PRESC_STD_CELL_HEIGHT, aDay[1], 4, 1, 0, "")

			pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)
			pdf.CellFormat(pdf.GetStringWidth("/"), constants.ETIX_7X4_STD_CELL_HEIGHT, "/", "", 0, "L", false, 0, "")
			
			pdf.SubWrite(constants.PRESC_STD_CELL_HEIGHT, doseSplit[1], 4, -3.5, 0, "")

			pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)
			pdf.SetX(pdf.GetX() + 0.1)
			pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, eticket.SetUnit, "", 1, "L", false, 0, "")
		} else {
			doseTxt := fmt.Sprintf("Sehari %s %s", strings.ToLower(eticket.Dose), strings.ToLower(eticket.SetUnit))

			pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, doseTxt, "", 1, "C", false, 0, "")
		}
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())

	// Consume Time
	{
		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)
		pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, eticket.ConsumeTime, "", 1, "C", false, 0, "")
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())

	// Must Finish
	{
		if eticket.MustFinish {
			pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)
			pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, "HABISKAN", "", 1, "C", false, 0, "")

			pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())
		}
	}

	// Qty
	{
		qtyTxt := fmt.Sprintf("Qty: %.0f", eticket.MedicineQty)
		pdf.SetFont("Arial", constants.REGULAR, constants.ETIX_7X4_STD_FONT_SZ)
		pdf.CellFormat(0, constants.ETIX_7X4_STD_CELL_HEIGHT, qtyTxt, "", 1, "C", false, 0, "")
	}

	pdf.Line(0, pdf.GetY(), constants.ETIX_7X4_WIDTH, pdf.GetY())

    if pdf.Error() != nil {
        return fmt.Errorf("error create eticket 7x4 data: %v", pdf.Error())
    }

	return nil
}
