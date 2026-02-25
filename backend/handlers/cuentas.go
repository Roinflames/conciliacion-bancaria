package handlers

import (
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/middleware"
	"conciliacion-bancaria/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func ListCuentas(c *fiber.Ctx) error {
	rows, err := db.DB.Query(
		`SELECT id, banco, numero_cuenta, tipo, moneda, COALESCE(alias,''), activa, created_at
		 FROM cuentas_bancarias ORDER BY id`,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var cuentas []models.CuentaBancaria
	for rows.Next() {
		var cu models.CuentaBancaria
		rows.Scan(&cu.ID, &cu.Banco, &cu.NumeroCuenta, &cu.Tipo, &cu.Moneda, &cu.Alias, &cu.Activa, &cu.CreatedAt)
		cuentas = append(cuentas, cu)
	}
	if cuentas == nil {
		cuentas = []models.CuentaBancaria{}
	}
	return c.JSON(cuentas)
}

func GetCuenta(c *fiber.Ctx) error {
	id := c.Params("id")
	var cu models.CuentaBancaria
	err := db.DB.QueryRow(
		`SELECT id, banco, numero_cuenta, tipo, moneda, COALESCE(alias,''), activa, created_at
		 FROM cuentas_bancarias WHERE id = ?`, id,
	).Scan(&cu.ID, &cu.Banco, &cu.NumeroCuenta, &cu.Tipo, &cu.Moneda, &cu.Alias, &cu.Activa, &cu.CreatedAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "cuenta no encontrada"})
	}
	return c.JSON(cu)
}

func CreateCuenta(c *fiber.Ctx) error {
	var body models.CuentaBancaria
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cuerpo inválido"})
	}
	if body.Banco == "" || body.NumeroCuenta == "" {
		return c.Status(400).JSON(fiber.Map{"error": "banco y numero_cuenta son requeridos"})
	}
	if body.Moneda == "" {
		body.Moneda = "CLP"
	}

	res, err := db.DB.Exec(
		`INSERT INTO cuentas_bancarias (banco, numero_cuenta, tipo, moneda, alias) VALUES (?,?,?,?,?)`,
		body.Banco, body.NumeroCuenta, body.Tipo, body.Moneda, body.Alias,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "CREATE", "cuentas_bancarias", fmt.Sprint(id),
		fmt.Sprintf(`{"banco":"%s","numero":"%s"}`, body.Banco, body.NumeroCuenta), c.IP())

	return c.Status(201).JSON(fiber.Map{"id": id, "message": "cuenta creada"})
}

func UpdateCuenta(c *fiber.Ctx) error {
	id := c.Params("id")
	var body models.CuentaBancaria
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cuerpo inválido"})
	}

	_, err := db.DB.Exec(
		`UPDATE cuentas_bancarias SET banco=?, numero_cuenta=?, tipo=?, moneda=?, alias=?, activa=? WHERE id=?`,
		body.Banco, body.NumeroCuenta, body.Tipo, body.Moneda, body.Alias, body.Activa, id,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "UPDATE", "cuentas_bancarias", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "cuenta actualizada"})
}

func DeleteCuenta(c *fiber.Ctx) error {
	id := c.Params("id")

	// Verificar que no tenga cartolas asociadas
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM cartolas WHERE cuenta_id = ?`, id).Scan(&count)
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "la cuenta tiene cartolas importadas y no puede eliminarse"})
	}

	db.DB.Exec(`DELETE FROM cuentas_bancarias WHERE id = ?`, id)
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "DELETE", "cuentas_bancarias", id, "", c.IP())
	return c.JSON(fiber.Map{"message": "cuenta eliminada"})
}
