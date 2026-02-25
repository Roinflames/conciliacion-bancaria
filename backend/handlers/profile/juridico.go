package profile

import (
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/middleware"
	"conciliacion-bancaria/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// ── CUSTODIA ─────────────────────────────────────────────────────────────────

func ListCustodia(c *fiber.Ctx) error {
	rows, err := db.DB.Query(
		`SELECT cu.id, cu.entidad_id, e.nombre, cu.descripcion, cu.monto,
		        cu.fecha_ingreso, COALESCE(cu.fecha_devol,''), cu.estado,
		        cu.cuenta_id, cu.created_at
		 FROM custodia cu
		 JOIN entidades e ON e.id = cu.entidad_id
		 ORDER BY cu.fecha_ingreso DESC`,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var items []models.Custodia
	for rows.Next() {
		var cu models.Custodia
		rows.Scan(&cu.ID, &cu.EntidadID, &cu.EntidadNombre, &cu.Descripcion,
			&cu.Monto, &cu.FechaIngreso, &cu.FechaDevol, &cu.Estado,
			&cu.CuentaID, &cu.CreatedAt)
		items = append(items, cu)
	}
	if items == nil {
		items = []models.Custodia{}
	}
	return c.JSON(items)
}

func CreateCustodia(c *fiber.Ctx) error {
	var body models.Custodia
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cuerpo inválido"})
	}
	if body.EntidadID == 0 || body.Monto == 0 || body.FechaIngreso == "" {
		return c.Status(400).JSON(fiber.Map{"error": "entidad_id, monto y fecha_ingreso son requeridos"})
	}

	res, err := db.DB.Exec(
		`INSERT INTO custodia (entidad_id, descripcion, monto, fecha_ingreso, fecha_devol, estado, cuenta_id)
		 VALUES (?,?,?,?,?,?,?)`,
		body.EntidadID, body.Descripcion, body.Monto, body.FechaIngreso,
		nullStr(body.FechaDevol), body.Estado, body.CuentaID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "CREATE", "custodia", fmt.Sprint(id),
		fmt.Sprintf(`{"monto":%.2f}`, body.Monto), c.IP())

	return c.Status(201).JSON(fiber.Map{"id": id, "message": "custodia registrada"})
}

func UpdateCustodia(c *fiber.Ctx) error {
	id := c.Params("id")
	var body models.Custodia
	c.BodyParser(&body)

	db.DB.Exec(
		`UPDATE custodia SET descripcion=?, monto=?, fecha_devol=?, estado=? WHERE id=?`,
		body.Descripcion, body.Monto, nullStr(body.FechaDevol), body.Estado, id,
	)

	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "UPDATE", "custodia", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "custodia actualizada"})
}

func DeleteCustodia(c *fiber.Ctx) error {
	id := c.Params("id")
	db.DB.Exec(`DELETE FROM custodia WHERE id=?`, id)
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "DELETE", "custodia", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "custodia eliminada"})
}

func SaldoCustodiaCliente(c *fiber.Ctx) error {
	entidadID := c.Params("entidad_id")
	var saldo float64
	db.DB.QueryRow(
		`SELECT COALESCE(SUM(CASE WHEN estado='activo' THEN monto ELSE -monto END),0)
		 FROM custodia WHERE entidad_id=?`, entidadID,
	).Scan(&saldo)
	return c.JSON(fiber.Map{"entidad_id": entidadID, "saldo_custodia": saldo})
}

// ── HONORARIOS ───────────────────────────────────────────────────────────────

func ListHonorarios(c *fiber.Ctx) error {
	rows, err := db.DB.Query(
		`SELECT h.id, h.entidad_id, e.nombre, h.concepto, h.monto,
		        h.fecha_emision, COALESCE(h.fecha_pago,''), h.estado,
		        h.movimiento_id, h.created_at
		 FROM honorarios h
		 JOIN entidades e ON e.id = h.entidad_id
		 ORDER BY h.fecha_emision DESC`,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var items []models.Honorario
	for rows.Next() {
		var h models.Honorario
		rows.Scan(&h.ID, &h.EntidadID, &h.EntidadNombre, &h.Concepto, &h.Monto,
			&h.FechaEmision, &h.FechaPago, &h.Estado, &h.MovimientoID, &h.CreatedAt)
		items = append(items, h)
	}
	if items == nil {
		items = []models.Honorario{}
	}
	return c.JSON(items)
}

func CreateHonorario(c *fiber.Ctx) error {
	var body models.Honorario
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cuerpo inválido"})
	}
	if body.EntidadID == 0 || body.Monto == 0 || body.FechaEmision == "" {
		return c.Status(400).JSON(fiber.Map{"error": "entidad_id, monto y fecha_emision son requeridos"})
	}
	if body.Estado == "" {
		body.Estado = "pendiente"
	}

	res, err := db.DB.Exec(
		`INSERT INTO honorarios (entidad_id, concepto, monto, fecha_emision, fecha_pago, estado, movimiento_id)
		 VALUES (?,?,?,?,?,?,?)`,
		body.EntidadID, body.Concepto, body.Monto, body.FechaEmision,
		nullStr(body.FechaPago), body.Estado, body.MovimientoID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "CREATE", "honorarios", fmt.Sprint(id),
		fmt.Sprintf(`{"monto":%.2f,"estado":"%s"}`, body.Monto, body.Estado), c.IP())

	return c.Status(201).JSON(fiber.Map{"id": id, "message": "honorario registrado"})
}

func UpdateHonorario(c *fiber.Ctx) error {
	id := c.Params("id")
	var body models.Honorario
	c.BodyParser(&body)

	db.DB.Exec(
		`UPDATE honorarios SET concepto=?, monto=?, fecha_pago=?, estado=?, movimiento_id=? WHERE id=?`,
		body.Concepto, body.Monto, nullStr(body.FechaPago), body.Estado, body.MovimientoID, id,
	)

	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "UPDATE", "honorarios", id,
		fmt.Sprintf(`{"estado":"%s"}`, body.Estado), c.IP())
	return c.JSON(fiber.Map{"message": "honorario actualizado"})
}

func DeleteHonorario(c *fiber.Ctx) error {
	id := c.Params("id")
	db.DB.Exec(`DELETE FROM honorarios WHERE id=?`, id)
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "DELETE", "honorarios", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "honorario eliminado"})
}

func ResumenHonorarios(c *fiber.Ctx) error {
	var pactado, cobrado, pendiente float64
	db.DB.QueryRow(`SELECT COALESCE(SUM(monto),0) FROM honorarios`).Scan(&pactado)
	db.DB.QueryRow(`SELECT COALESCE(SUM(monto),0) FROM honorarios WHERE estado='pagado'`).Scan(&cobrado)
	pendiente = pactado - cobrado

	return c.JSON(fiber.Map{
		"pactado":   pactado,
		"cobrado":   cobrado,
		"pendiente": pendiente,
	})
}

// nullStr convierte string vacío a nil para SQLite.
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
