package services

import (
	"conciliacion-bancaria/models"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ParseCSV parsea una cartola en formato CSV.
// Formato esperado (BancoEstado / BCI / Santander):
// Fecha;Descripcion;Referencia;Debe;Haber;Saldo
func ParseCSV(r io.Reader) ([]models.Movimiento, error) {
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.LazyQuotes = true

	var movimientos []models.Movimiento
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error leyendo CSV línea %d: %w", lineNum, err)
		}
		lineNum++

		// Saltar encabezado
		if lineNum == 1 {
			continue
		}
		if len(record) < 4 {
			continue
		}

		mov, err := parseRow(record)
		if err != nil {
			continue // fila inválida se omite
		}
		movimientos = append(movimientos, mov)
	}

	return movimientos, nil
}

// ParseTXT detecta automáticamente el formato del TXT:
// - Ancho fijo banco chileno (líneas de 121 chars con +000) → ParseBancoFW
// - Separado por tabulaciones → ParseCSV
func ParseTXT(r io.Reader) ([]models.Movimiento, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	content := string(data)

	// Detección: formato ancho fijo si la primera línea mide ~121 y contiene +000
	firstLine := strings.SplitN(content, "\n", 2)[0]
	firstLine = strings.TrimRight(firstLine, "\r")
	if len(firstLine) >= 100 && strings.Contains(firstLine, "+000") {
		return ParseBancoFW(strings.NewReader(content))
	}

	replaced := strings.ReplaceAll(content, "\t", ";")
	return ParseCSV(strings.NewReader(replaced))
}

// ParseBancoFW parsea cartolas en formato de ancho fijo de bancos chilenos.
//
// Estructura de cada línea (121 chars):
//   - Pos  0-12 (13): código de cuenta/sucursal
//   - Pos 13-18  (6): fecha YYMMDD
//   - Pos 19-??     : relleno de ceros hasta el campo de monto
//   - 22 dígitos    : monto (cero-relleno izquierda) inmediatamente antes de "+000"
//   - "+000"        : separador
//   - Pos +0 a +44  : descripción (45 chars, rellena con espacios a la derecha)
//   - Pos +45       : tipo  S=sistema  C=cargo/egreso  A=abono/ingreso
//   - Pos +46 a +53 : fecha YYYYMMDD (referencia)
//   - Pos +54 a +58 : zeros de cierre
//
// Las filas de tipo "S" (sistema: retenciones, saldo contable) se omiten.
func ParseBancoFW(r io.Reader) ([]models.Movimiento, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Regex para capturar el monto: exactamente 22 dígitos antes de +000
	amountRe := regexp.MustCompile(`(\d{22})\+000`)

	var movimientos []models.Movimiento

	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(rawLine, "\r")
		if len(line) < 100 {
			continue
		}

		// Fecha desde pos 13-18 (YYMMDD)
		if len(line) < 19 {
			continue
		}
		yy, mm, dd := "20"+line[13:15], line[15:17], line[17:19]
		fecha := yy + "-" + mm + "-" + dd

		// Monto
		m := amountRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		monto, _ := strconv.ParseFloat(strings.TrimLeft(m[1], "0"), 64)
		if monto == 0 {
			monto = 0
		}

		// Posición después de +000
		idx := strings.Index(line, "+000")
		if idx < 0 || len(line) < idx+50 {
			continue
		}
		after := line[idx+4:]
		if len(after) < 46 {
			continue
		}

		desc := strings.TrimSpace(after[:45])
		tipo := string(after[45])

		// Omitir filas de sistema (retenciones, saldo contable, etc.)
		if tipo == "S" {
			continue
		}

		var debe, haber float64
		switch tipo {
		case "C":
			debe = monto
		case "A":
			haber = monto
		default:
			continue
		}

		movimientos = append(movimientos, models.Movimiento{
			Fecha:       fecha,
			Descripcion: desc,
			Debe:        debe,
			Haber:       haber,
		})
	}

	if len(movimientos) == 0 {
		return nil, fmt.Errorf("no se encontraron movimientos en el archivo")
	}

	return movimientos, nil
}

// ParseExcel parsea la primera hoja de un archivo .xlsx.
func ParseExcel(r io.Reader) ([]models.Movimiento, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el Excel: %w", err)
	}

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("el archivo no tiene hojas")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	var movimientos []models.Movimiento
	for i, row := range rows {
		if i == 0 {
			continue // encabezado
		}
		if len(row) < 4 {
			continue
		}
		mov, err := parseRow(row)
		if err != nil {
			continue
		}
		movimientos = append(movimientos, mov)
	}

	return movimientos, nil
}

// parseRow convierte una fila [fecha, descripcion, referencia, debe, haber, saldo?] en Movimiento.
func parseRow(cols []string) (models.Movimiento, error) {
	clean := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
		return s
	}

	fecha, err := parseDate(strings.TrimSpace(cols[0]))
	if err != nil {
		return models.Movimiento{}, fmt.Errorf("fecha inválida: %s", cols[0])
	}

	descripcion := strings.TrimSpace(cols[1])
	referencia := ""
	debeIdx, haberIdx, saldoIdx := 2, 3, 4

	// Si hay 6+ columnas, col[2] es referencia
	if len(cols) >= 6 {
		referencia = strings.TrimSpace(cols[2])
		debeIdx, haberIdx, saldoIdx = 3, 4, 5
	}

	debe, _ := strconv.ParseFloat(clean(cols[debeIdx]), 64)
	haber, _ := strconv.ParseFloat(clean(cols[haberIdx]), 64)
	saldo := 0.0
	if len(cols) > saldoIdx {
		saldo, _ = strconv.ParseFloat(clean(cols[saldoIdx]), 64)
	}

	return models.Movimiento{
		Fecha:       fecha,
		Descripcion: descripcion,
		Referencia:  referencia,
		Debe:        debe,
		Haber:       haber,
		Saldo:       saldo,
	}, nil
}

// parseDate intenta parsear fecha en múltiples formatos chilenos.
func parseDate(s string) (string, error) {
	formats := []string{
		"02/01/2006", "02-01-2006",
		"2006-01-02", "2006/01/02",
		"02/01/06", "2/1/2006",
	}
	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("formato de fecha no reconocido: %s", s)
}
