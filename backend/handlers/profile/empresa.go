package profile

import (
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/middleware"
	"conciliacion-bancaria/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// ── CATEGORÍAS ────────────────────────────────────────────────────────────────

func ListCategorias(c *fiber.Ctx) error {
	rows, err := db.DB.Query(`SELECT id, nombre, tipo, activa FROM categorias ORDER BY tipo, nombre`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var items []models.Categoria
	for rows.Next() {
		var cat models.Categoria
		rows.Scan(&cat.ID, &cat.Nombre, &cat.Tipo, &cat.Activa)
		items = append(items, cat)
	}
	if items == nil {
		items = []models.Categoria{}
	}
	return c.JSON(items)
}

func CreateCategoria(c *fiber.Ctx) error {
	var body models.Categoria
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cuerpo inválido"})
	}
	if body.Nombre == "" {
		return c.Status(400).JSON(fiber.Map{"error": "nombre es requerido"})
	}

	res, err := db.DB.Exec(`INSERT INTO categorias (nombre, tipo) VALUES (?,?)`, body.Nombre, body.Tipo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	id, _ := res.LastInsertId()
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "CREATE", "categorias", fmt.Sprint(id),
		fmt.Sprintf(`{"nombre":"%s","tipo":"%s"}`, body.Nombre, body.Tipo), c.IP())

	return c.Status(201).JSON(fiber.Map{"id": id, "message": "categoría creada"})
}

func UpdateCategoria(c *fiber.Ctx) error {
	id := c.Params("id")
	var body models.Categoria
	c.BodyParser(&body)
	db.DB.Exec(`UPDATE categorias SET nombre=?, tipo=?, activa=? WHERE id=?`, body.Nombre, body.Tipo, body.Activa, id)
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "UPDATE", "categorias", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "categoría actualizada"})
}

func DeleteCategoria(c *fiber.Ctx) error {
	id := c.Params("id")
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM ingresos_recurrentes WHERE categoria_id=?`, id).Scan(&count)
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "categoría tiene ingresos asociados"})
	}
	db.DB.Exec(`DELETE FROM categorias WHERE id=?`, id)
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "DELETE", "categorias", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "categoría eliminada"})
}

// ── INGRESOS RECURRENTES ──────────────────────────────────────────────────────

func ListIngresos(c *fiber.Ctx) error {
	rows, err := db.DB.Query(
		`SELECT ir.id, ir.entidad_id, e.nombre, ir.categoria_id, cat.nombre,
		        ir.descripcion, ir.monto, ir.periodicidad,
		        COALESCE(ir.dia_cobro,0), ir.estado, ir.activo, ir.created_at
		 FROM ingresos_recurrentes ir
		 JOIN entidades e ON e.id = ir.entidad_id
		 JOIN categorias cat ON cat.id = ir.categoria_id
		 ORDER BY e.nombre`,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var items []models.IngresoRecurrente
	for rows.Next() {
		var ir models.IngresoRecurrente
		rows.Scan(&ir.ID, &ir.EntidadID, &ir.EntidadNombre,
			&ir.CategoriaID, &ir.CategoriaNombre,
			&ir.Descripcion, &ir.Monto, &ir.Periodicidad,
			&ir.DiaCobro, &ir.Estado, &ir.Activo, &ir.CreatedAt)
		items = append(items, ir)
	}
	if items == nil {
		items = []models.IngresoRecurrente{}
	}
	return c.JSON(items)
}

func CreateIngreso(c *fiber.Ctx) error {
	var body models.IngresoRecurrente
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cuerpo inválido"})
	}
	if body.EntidadID == 0 || body.CategoriaID == 0 || body.Monto == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "entidad_id, categoria_id y monto son requeridos"})
	}
	if body.Estado == "" {
		body.Estado = "esperado"
	}

	res, err := db.DB.Exec(
		`INSERT INTO ingresos_recurrentes (entidad_id, categoria_id, descripcion, monto, periodicidad, dia_cobro, estado)
		 VALUES (?,?,?,?,?,?,?)`,
		body.EntidadID, body.CategoriaID, body.Descripcion, body.Monto,
		body.Periodicidad, body.DiaCobro, body.Estado,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "CREATE", "ingresos_recurrentes", fmt.Sprint(id),
		fmt.Sprintf(`{"monto":%.2f,"estado":"%s"}`, body.Monto, body.Estado), c.IP())

	return c.Status(201).JSON(fiber.Map{"id": id, "message": "ingreso registrado"})
}

func UpdateIngreso(c *fiber.Ctx) error {
	id := c.Params("id")
	var body models.IngresoRecurrente
	c.BodyParser(&body)

	db.DB.Exec(
		`UPDATE ingresos_recurrentes SET descripcion=?, monto=?, periodicidad=?, dia_cobro=?, estado=?, activo=? WHERE id=?`,
		body.Descripcion, body.Monto, body.Periodicidad, body.DiaCobro, body.Estado, body.Activo, id,
	)

	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "UPDATE", "ingresos_recurrentes", id,
		fmt.Sprintf(`{"estado":"%s"}`, body.Estado), c.IP())
	return c.JSON(fiber.Map{"message": "ingreso actualizado"})
}

func DeleteIngreso(c *fiber.Ctx) error {
	id := c.Params("id")
	db.DB.Exec(`DELETE FROM ingresos_recurrentes WHERE id=?`, id)
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "DELETE", "ingresos_recurrentes", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "ingreso eliminado"})
}

func ResumenIngresos(c *fiber.Ctx) error {
	var esperado, recibido float64
	db.DB.QueryRow(`SELECT COALESCE(SUM(monto),0) FROM ingresos_recurrentes WHERE activo=1`).Scan(&esperado)
	db.DB.QueryRow(`SELECT COALESCE(SUM(monto),0) FROM ingresos_recurrentes WHERE estado='recibido' AND activo=1`).Scan(&recibido)

	return c.JSON(fiber.Map{
		"esperado": esperado,
		"recibido": recibido,
		"pendiente": esperado - recibido,
	})
}
