package records

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	documentName := r.FormValue("documentName")
	documentDate := r.FormValue("documentDate")
	documentSensitivity := r.FormValue("documentSensitivity")
	documentTypes := r.Form["documentType"]
	sensitivity := documentSensitivity == "true"
	files := r.MultipartForm.File["photos"]

	fmt.Println("----- NEW RECORD -----")
	fmt.Println("Name:", documentName)
	fmt.Println("Date:", documentDate)
	fmt.Println("Sensitivity:", sensitivity)
	fmt.Println("Types:", documentTypes)
	fmt.Println("File count:", len(files))

	id := uuid.New().String()
	fmt.Println("ID:", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func PrintLabel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; "+
			"script-src 'self'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data:; "+
			"connect-src 'self'; "+
			"object-src 'self'; "+
			"base-uri 'self'; "+
			"frame-ancestors 'self';")
	w.Header().Set("Content-Encoding", "identity")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline")

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	stream := "q\n" +
		"1 1 1 rg\n" +
		"0 0 612 792 re f\n" +
		"Q\n" +
		"q\n" +
		"0 0 0 RG\n" +
		"1 w\n" +
		"468 750 108 20 re S\n" +
		"BT\n" +
		"/F1 9 Tf\n" +
		"472 755 Td\n" +
		"(ID: " + id + ") Tj\n" +
		"ET\n"

	var buf bytes.Buffer
	offsets := make([]int, 6)

	buf.WriteString("%PDF-1.4\n")

	offsets[1] = buf.Len()
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	offsets[2] = buf.Len()
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	offsets[3] = buf.Len()
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n")

	offsets[4] = buf.Len()
	buf.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(stream), stream))

	offsets[5] = buf.Len()
	buf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString("0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset))

    fmt.Println("=== PrintLabel called, id:", id)
    fmt.Println("=== PDF size:", buf.Len(), "bytes")

    w.Write(buf.Bytes())
}

func TypeOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	options := map[string]interface{}{
		"documentTypes": []map[string]string{
			{
				"documentLabel":      "Invoice",
				"documentLabelValue": "invoice",
			},
		},
	}

	json.NewEncoder(w).Encode(options)
}