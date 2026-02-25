package services

import (
	"conciliacion-bancaria/db"
	"math"
	"strings"
	"time"
)

type MatchResult struct {
	MovimientoID int
	Estado       string // conciliado|diferencia|pendiente
	Observacion  string
}

// AutoMatch ejecuta la conciliación automática para todos los movimientos de una cartola.
// Lógica: busca en la misma cuenta un registro interno (custodia u honorario) con
// monto exacto en una ventana de ±3 días. Si encuentra uno, marca como conciliado.
// Si no, marca como diferencia.
func AutoMatch(cartolaID int) ([]MatchResult, error) {
	// Cargar movimientos de la cartola
	rows, err := db.DB.Query(
		`SELECT id, fecha, debe, haber FROM movimientos WHERE cartola_id = ? AND conciliado = 0`,
		cartolaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type movRow struct {
		ID    int
		Fecha string
		Debe  float64
		Haber float64
	}

	var movs []movRow
	for rows.Next() {
		var m movRow
		rows.Scan(&m.ID, &m.Fecha, &m.Debe, &m.Haber)
		movs = append(movs, m)
	}

	var results []MatchResult

	for _, m := range movs {
		monto := m.Haber
		if m.Debe > 0 {
			monto = m.Debe
		}

		matched, obs := findMatch(m.Fecha, monto)

		estado := "diferencia"
		if matched {
			estado = "conciliado"
			// Marcar movimiento como conciliado
			db.DB.Exec(`UPDATE movimientos SET conciliado = 1 WHERE id = ?`, m.ID)
		}

		results = append(results, MatchResult{
			MovimientoID: m.ID,
			Estado:       estado,
			Observacion:  obs,
		})
	}

	return results, nil
}

// findMatch busca un registro interno (honorario o ingreso_recurrente) que coincida.
func findMatch(fechaStr string, monto float64) (bool, string) {
	fecha, err := time.Parse("2006-01-02", fechaStr)
	if err != nil {
		return false, "fecha inválida"
	}

	desde := fecha.AddDate(0, 0, -3).Format("2006-01-02")
	hasta := fecha.AddDate(0, 0, 3).Format("2006-01-02")

	// Buscar en honorarios
	var honMonto float64
	var honConcepto string
	err = db.DB.QueryRow(
		`SELECT monto, concepto FROM honorarios
		 WHERE estado = 'pendiente' AND monto = ? AND fecha_emision BETWEEN ? AND ?
		 LIMIT 1`,
		monto, desde, hasta,
	).Scan(&honMonto, &honConcepto)

	if err == nil && floatEq(honMonto, monto) {
		return true, "coincide con honorario: " + truncate(honConcepto, 50)
	}

	// Buscar en ingresos recurrentes
	var ingMonto float64
	var ingDesc string
	err = db.DB.QueryRow(
		`SELECT monto, descripcion FROM ingresos_recurrentes
		 WHERE estado = 'esperado' AND monto = ? AND activo = 1
		 LIMIT 1`,
		monto,
	).Scan(&ingMonto, &ingDesc)

	if err == nil && floatEq(ingMonto, monto) {
		return true, "coincide con ingreso recurrente: " + truncate(ingDesc, 50)
	}

	return false, "sin coincidencia interna"
}

func floatEq(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
