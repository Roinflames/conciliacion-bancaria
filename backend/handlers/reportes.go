package handlers

import (
	"conciliacion-bancaria/config"
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/models"

	"github.com/gofiber/fiber/v2"
)

func GetPosicion(c *fiber.Ctx) error {
	desde := c.Query("desde", "2000-01-01")
	hasta := c.Query("hasta", "2099-12-31")

	rows, err := db.DB.Query(
		`SELECT COALESCE(cb.alias, cb.banco), cb.banco,
		        COALESCE(SUM(m.haber),0) as ingresos,
		        COALESCE(SUM(m.debe),0)  as egresos
		 FROM cuentas_bancarias cb
		 LEFT JOIN cartolas ca ON ca.cuenta_id = cb.id
		 LEFT JOIN movimientos m ON m.cartola_id = ca.id
		     AND m.fecha BETWEEN ? AND ?
		 GROUP BY cb.id
		 ORDER BY cb.alias`,
		desde, hasta,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var posiciones []models.PosicionFinanciera
	for rows.Next() {
		var p models.PosicionFinanciera
		rows.Scan(&p.Cuenta, &p.Banco, &p.Ingresos, &p.Egresos)
		p.Saldo = p.Ingresos - p.Egresos
		posiciones = append(posiciones, p)
	}
	if posiciones == nil {
		posiciones = []models.PosicionFinanciera{}
	}
	return c.JSON(fiber.Map{
		"desde":      desde,
		"hasta":      hasta,
		"posiciones": posiciones,
	})
}

func GetAudit(c *fiber.Ctx) error {
	desde := c.Query("desde", "2000-01-01")
	hasta := c.Query("hasta", "2099-12-31")

	rows, err := db.DB.Query(
		`SELECT al.id, COALESCE(al.usuario_id,0), al.accion, al.tabla,
		        al.registro_id, COALESCE(al.payload,''), COALESCE(al.ip,''), al.created_at
		 FROM audit_log al
		 WHERE DATE(al.created_at) BETWEEN ? AND ?
		 ORDER BY al.created_at DESC
		 LIMIT 200`,
		desde, hasta,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var entries []models.AuditEntry
	for rows.Next() {
		var e models.AuditEntry
		rows.Scan(&e.ID, &e.UsuarioID, &e.Accion, &e.Tabla,
			&e.RegistroID, &e.Payload, &e.IP, &e.CreatedAt)
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []models.AuditEntry{}
	}
	return c.JSON(entries)
}

func GetConfigPerfil(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"perfil": config.Profile})
}
