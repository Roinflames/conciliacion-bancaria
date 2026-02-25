package handlers

import (
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/middleware"
	"conciliacion-bancaria/models"
	"conciliacion-bancaria/services"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ListCartolas(c *fiber.Ctx) error {
	rows, err := db.DB.Query(
		`SELECT c.id, c.cuenta_id, c.periodo_desde, c.periodo_hasta,
		        COALESCE(c.archivo_nombre,''), c.importado_por, c.importado_at,
		        COUNT(m.id) as total
		 FROM cartolas c
		 LEFT JOIN movimientos m ON m.cartola_id = c.id
		 GROUP BY c.id
		 ORDER BY c.id DESC`,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var cartolas []models.Cartola
	for rows.Next() {
		var ca models.Cartola
		rows.Scan(&ca.ID, &ca.CuentaID, &ca.PeriodoDesde, &ca.PeriodoHasta,
			&ca.ArchivoNombre, &ca.ImportadoPor, &ca.ImportadoAt, &ca.TotalMov)
		cartolas = append(cartolas, ca)
	}
	if cartolas == nil {
		cartolas = []models.Cartola{}
	}
	return c.JSON(cartolas)
}

func GetMovimientos(c *fiber.Ctx) error {
	id := c.Params("id")
	rows, err := db.DB.Query(
		`SELECT id, cartola_id, fecha, COALESCE(descripcion,''), COALESCE(referencia,''),
		        debe, haber, saldo, conciliado, created_at
		 FROM movimientos WHERE cartola_id = ? ORDER BY fecha, id`,
		id,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var movs []models.Movimiento
	for rows.Next() {
		var m models.Movimiento
		rows.Scan(&m.ID, &m.CartolaID, &m.Fecha, &m.Descripcion, &m.Referencia,
			&m.Debe, &m.Haber, &m.Saldo, &m.Conciliado, &m.CreatedAt)
		movs = append(movs, m)
	}
	if movs == nil {
		movs = []models.Movimiento{}
	}
	return c.JSON(movs)
}

func ImportCartola(c *fiber.Ctx) error {
	file, err := c.FormFile("archivo")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "archivo requerido (campo: archivo)"})
	}

	cuentaID := c.FormValue("cuenta_id")
	periodoDesde := c.FormValue("periodo_desde")
	periodoHasta := c.FormValue("periodo_hasta")

	if cuentaID == "" || periodoDesde == "" || periodoHasta == "" {
		return c.Status(400).JSON(fiber.Map{"error": "cuenta_id, periodo_desde y periodo_hasta son requeridos"})
	}

	// Verificar que no exista cartola duplicada
	var dup int
	db.DB.QueryRow(
		`SELECT COUNT(*) FROM cartolas WHERE cuenta_id=? AND periodo_desde=? AND periodo_hasta=?`,
		cuentaID, periodoDesde, periodoHasta,
	).Scan(&dup)
	if dup > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "ya existe una cartola para este período y cuenta"})
	}

	// Abrir y parsear
	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "no se pudo leer el archivo"})
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var movimientos []models.Movimiento

	switch ext {
	case ".csv":
		movimientos, err = services.ParseCSV(f)
	case ".txt":
		movimientos, err = services.ParseTXT(f)
	case ".xlsx":
		movimientos, err = services.ParseExcel(f)
	default:
		return c.Status(400).JSON(fiber.Map{"error": "formato no soportado (csv, txt, xlsx)"})
	}

	if err != nil {
		return c.Status(422).JSON(fiber.Map{"error": "error parseando archivo: " + err.Error()})
	}

	if len(movimientos) == 0 {
		return c.Status(422).JSON(fiber.Map{"error": "el archivo no contiene movimientos válidos"})
	}

	// Insertar cartola
	user := middleware.CurrentUser(c)
	res, err := db.DB.Exec(
		`INSERT INTO cartolas (cuenta_id, periodo_desde, periodo_hasta, archivo_nombre, importado_por)
		 VALUES (?,?,?,?,?)`,
		cuentaID, periodoDesde, periodoHasta, file.Filename, user.UserID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	cartolaID, _ := res.LastInsertId()

	// Insertar movimientos en lote
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(
		`INSERT INTO movimientos (cartola_id, fecha, descripcion, referencia, debe, haber, saldo)
		 VALUES (?,?,?,?,?,?,?)`,
	)
	for _, m := range movimientos {
		stmt.Exec(cartolaID, m.Fecha, m.Descripcion, m.Referencia, m.Debe, m.Haber, m.Saldo)
	}
	stmt.Close()
	tx.Commit()

	db.Audit(user.UserID, "CREATE", "cartolas", fmt.Sprint(cartolaID),
		fmt.Sprintf(`{"archivo":"%s","movimientos":%d}`, file.Filename, len(movimientos)), c.IP())

	return c.Status(201).JSON(fiber.Map{
		"cartola_id":  cartolaID,
		"movimientos": len(movimientos),
		"message":     "cartola importada correctamente",
	})
}

func DeleteCartola(c *fiber.Ctx) error {
	id := c.Params("id")
	db.DB.Exec(`DELETE FROM movimientos WHERE cartola_id = ?`, id)
	db.DB.Exec(`DELETE FROM cartolas WHERE id = ?`, id)
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "DELETE", "cartolas", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "cartola eliminada"})
}
