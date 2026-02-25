package handlers

import (
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/middleware"
	"conciliacion-bancaria/models"
	"conciliacion-bancaria/services"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func EjecutarConciliacion(c *fiber.Ctx) error {
	id := c.Params("id")

	results, err := services.AutoMatch(toInt(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Persistir resultados
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(
		`INSERT INTO conciliacion_items (cartola_id, movimiento_id, tipo, estado, observacion)
		 VALUES (?,?,?,?,?)`,
	)
	for _, r := range results {
		stmt.Exec(id, r.MovimientoID, "auto", r.Estado, r.Observacion)
	}
	stmt.Close()
	tx.Commit()

	// Calcular resumen
	resumen := buildResumen(toInt(id))

	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "CREATE", "conciliacion_items", id,
		fmt.Sprintf(`{"total":%d}`, len(results)), c.IP())

	return c.JSON(fiber.Map{
		"message": "conciliación ejecutada",
		"resumen": resumen,
	})
}

func GetDiferencias(c *fiber.Ctx) error {
	id := c.Params("id")
	rows, err := db.DB.Query(
		`SELECT ci.id, ci.cartola_id, ci.movimiento_id, ci.tipo, ci.estado,
		        COALESCE(ci.observacion,''), ci.resuelto_at,
		        m.fecha, m.descripcion, m.debe, m.haber
		 FROM conciliacion_items ci
		 LEFT JOIN movimientos m ON m.id = ci.movimiento_id
		 WHERE ci.cartola_id = ? AND ci.estado = 'diferencia'
		 ORDER BY ci.id`,
		id,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	type DiferenciaRow struct {
		models.ConciliacionItem
		Fecha       string  `json:"fecha"`
		Descripcion string  `json:"descripcion"`
		Debe        float64 `json:"debe"`
		Haber       float64 `json:"haber"`
	}

	var items []DiferenciaRow
	for rows.Next() {
		var d DiferenciaRow
		rows.Scan(&d.ID, &d.CartolaID, &d.MovimientoID, &d.Tipo, &d.Estado,
			&d.Observacion, &d.ResueltoAt, &d.Fecha, &d.Descripcion, &d.Debe, &d.Haber)
		items = append(items, d)
	}
	if items == nil {
		items = []DiferenciaRow{}
	}
	return c.JSON(items)
}

func ResolverDiferencia(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Observacion string `json:"observacion"`
	}
	c.BodyParser(&body)

	user := middleware.CurrentUser(c)
	_, err := db.DB.Exec(
		`UPDATE conciliacion_items
		 SET estado='conciliado', observacion=?, resuelto_por=?, resuelto_at=CURRENT_TIMESTAMP
		 WHERE id=?`,
		body.Observacion, user.UserID, id,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	db.Audit(user.UserID, "UPDATE", "conciliacion_items", id,
		fmt.Sprintf(`{"observacion":"%s"}`, body.Observacion), c.IP())

	return c.JSON(fiber.Map{"message": "diferencia resuelta"})
}

func GetResumenConciliacion(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(buildResumen(toInt(id)))
}

func buildResumen(cartolaID int) models.ResumenConciliacion {
	var total, conciliados, diferencias, pendientes int
	db.DB.QueryRow(`SELECT COUNT(*) FROM movimientos WHERE cartola_id=?`, cartolaID).Scan(&total)
	db.DB.QueryRow(`SELECT COUNT(*) FROM conciliacion_items WHERE cartola_id=? AND estado='conciliado'`, cartolaID).Scan(&conciliados)
	db.DB.QueryRow(`SELECT COUNT(*) FROM conciliacion_items WHERE cartola_id=? AND estado='diferencia'`, cartolaID).Scan(&diferencias)
	pendientes = total - conciliados - diferencias

	pct := 0.0
	if total > 0 {
		pct = float64(conciliados) / float64(total) * 100
	}

	return models.ResumenConciliacion{
		CartolaID:     cartolaID,
		TotalMov:      total,
		Conciliados:   conciliados,
		Diferencias:   diferencias,
		Pendientes:    pendientes,
		PctConciliado: pct,
	}
}

func toInt(s string) int {
	n := 0
	fmt.Sscan(s, &n)
	return n
}
