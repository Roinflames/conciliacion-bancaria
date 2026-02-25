package handlers

import (
	"conciliacion-bancaria/config"
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/middleware"
	"conciliacion-bancaria/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func ListEntidades(c *fiber.Ctx) error {
	rows, err := db.DB.Query(
		`SELECT id, tipo, COALESCE(rut,''), nombre, COALESCE(email,''), COALESCE(telefono,''), activo, created_at
		 FROM entidades ORDER BY nombre`,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var items []models.Entidad
	for rows.Next() {
		var e models.Entidad
		rows.Scan(&e.ID, &e.Tipo, &e.RUT, &e.Nombre, &e.Email, &e.Telefono, &e.Activo, &e.CreatedAt)
		items = append(items, e)
	}
	if items == nil {
		items = []models.Entidad{}
	}
	return c.JSON(fiber.Map{
		"tipo":      entidadLabel(),
		"entidades": items,
	})
}

func GetEntidad(c *fiber.Ctx) error {
	id := c.Params("id")
	var e models.Entidad
	err := db.DB.QueryRow(
		`SELECT id, tipo, COALESCE(rut,''), nombre, COALESCE(email,''), COALESCE(telefono,''), activo, created_at
		 FROM entidades WHERE id = ?`, id,
	).Scan(&e.ID, &e.Tipo, &e.RUT, &e.Nombre, &e.Email, &e.Telefono, &e.Activo, &e.CreatedAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": entidadLabel() + " no encontrado/a"})
	}
	return c.JSON(e)
}

func CreateEntidad(c *fiber.Ctx) error {
	var body models.Entidad
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cuerpo inválido"})
	}
	if body.Nombre == "" {
		return c.Status(400).JSON(fiber.Map{"error": "nombre es requerido"})
	}

	// El tipo lo fija el perfil, no el cliente
	body.Tipo = entidadTipo()

	res, err := db.DB.Exec(
		`INSERT INTO entidades (tipo, rut, nombre, email, telefono) VALUES (?,?,?,?,?)`,
		body.Tipo, body.RUT, body.Nombre, body.Email, body.Telefono,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "CREATE", "entidades", fmt.Sprint(id),
		fmt.Sprintf(`{"nombre":"%s"}`, body.Nombre), c.IP())

	return c.Status(201).JSON(fiber.Map{"id": id, "message": entidadLabel() + " creado/a"})
}

func UpdateEntidad(c *fiber.Ctx) error {
	id := c.Params("id")
	var body models.Entidad
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cuerpo inválido"})
	}

	db.DB.Exec(
		`UPDATE entidades SET rut=?, nombre=?, email=?, telefono=?, activo=? WHERE id=?`,
		body.RUT, body.Nombre, body.Email, body.Telefono, body.Activo, id,
	)

	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "UPDATE", "entidades", id, "", c.IP())
	return c.JSON(fiber.Map{"message": entidadLabel() + " actualizado/a"})
}

func DeleteEntidad(c *fiber.Ctx) error {
	id := c.Params("id")

	// Verificar dependencias
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM custodia WHERE entidad_id=?`, id).Scan(&count)
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "entidad tiene registros de custodia asociados"})
	}
	db.DB.QueryRow(`SELECT COUNT(*) FROM honorarios WHERE entidad_id=?`, id).Scan(&count)
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "entidad tiene honorarios asociados"})
	}

	db.DB.Exec(`DELETE FROM entidades WHERE id=?`, id)
	user := middleware.CurrentUser(c)
	db.Audit(user.UserID, "DELETE", "entidades", id, "", c.IP())
	return c.JSON(fiber.Map{"message": entidadLabel() + " eliminado/a"})
}

func entidadLabel() string {
	if config.IsJuridico() {
		return "cliente"
	}
	return "miembro"
}

func entidadTipo() string {
	if config.IsJuridico() {
		return "cliente"
	}
	return "miembro"
}
