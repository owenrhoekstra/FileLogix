package createRecord

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/datamatrix"
	"github.com/jung-kurt/gofpdf"
)

const (
	// Letter paper in mm
	pageWidth  = 215.9
	pageHeight = 279.4

	// Label dimensions in mm
	codeSize        = 10.0
	marginRight     = 7.0
	marginTop       = 7.0
	fontSize        = 4.0
	fontSpacing     = 0.0
	side_offset     = 1.0
	vertical_offset = 0.50
)

func GenerateLabel(documentID string) ([]byte, error) {
	// --- Generate Data Matrix ---
	dm, err := datamatrix.Encode(documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data matrix: %w", err)
	}

	scaled, err := barcode.Scale(dm, 100, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to scale data matrix: %w", err)
	}

	bounds := scaled.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, _ := scaled.At(x, y).RGBA()

			if r > 0 {
				gray.Set(x, y, color.Gray{255})
			} else {
				gray.Set(x, y, color.Gray{0})
			}
		}
	}

	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, gray); err != nil {
		return nil, fmt.Errorf("failed to encode data matrix as png: %w", err)
	}

	// --- Build PDF ---
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.AddPage()
	pdf.SetMargins(0, 0, 0)

	// Position: top-right, with margin
	codeX := pageWidth - marginRight - codeSize
	codeY := marginTop

	// "FileLogix" above the code if space allows
	labelY := codeY - fontSize/1.5 - vertical_offset
	pdf.SetFont("Helvetica", "B", fontSize*1.5)
	pdf.SetXY(codeX, labelY)
	pdf.CellFormat(codeSize, fontSize, "FileLogix", "", 0, "C", false, 0, "")

	// Data Matrix image
	pdf.RegisterImageOptionsReader(
		"datamatrix",
		gofpdf.ImageOptions{ImageType: "PNG"},
		&imgBuf,
	)
	pdf.ImageOptions("datamatrix", codeX, codeY, codeSize, codeSize, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	// UUID text around left, bottom, right sides
	pdf.SetFont("Helvetica", "", fontSize*1)

	// Truncate UUID for display — show first 8 chars on each side
	uuidLen := len(documentID)
	third := uuidLen / 3
	leftText := documentID[:third]
	bottomText := documentID[third : third*2]
	rightText := documentID[third*2:]

	// Left side (rotated 90°)
	pdf.TransformBegin()
	pdf.TransformRotate(-90, codeX, codeY+codeSize/2)
	pdf.SetXY(codeX-codeSize/2, codeY+codeSize/2-fontSize/2+side_offset)
	pdf.CellFormat(codeSize, fontSize, leftText, "", 0, "C", false, 0, "")
	pdf.TransformEnd()

	// Bottom
	pdf.SetXY(codeX, codeY+codeSize-vertical_offset*2.75)
	pdf.CellFormat(codeSize, fontSize, bottomText, "", 0, "C", false, 0, "")

	// Right side (rotated -90°)
	pdf.TransformBegin()
	pdf.TransformRotate(90, codeX+codeSize, codeY+codeSize/2)
	pdf.SetXY(codeX+codeSize-codeSize/2, codeY+codeSize/2-fontSize/2+side_offset)
	pdf.CellFormat(codeSize, fontSize, rightText, "", 0, "C", false, 0, "")
	pdf.TransformEnd()

	// --- Output ---
	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return out.Bytes(), nil
}
