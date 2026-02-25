package routes

import (
	"conciliacion-bancaria/config"
	"conciliacion-bancaria/handlers"
	"conciliacion-bancaria/handlers/profile"
	"conciliacion-bancaria/middleware"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	api := app.Group("/api")

	// ── Auth (público) ────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Post("/logout", middleware.Auth, handlers.Logout)
	auth.Get("/me", middleware.Auth, handlers.Me)

	// ── Config ────────────────────────────────────────────────────────────────
	api.Get("/config/perfil", handlers.GetConfigPerfil)

	// ── Cuentas bancarias ─────────────────────────────────────────────────────
	cuentas := api.Group("/cuentas", middleware.Auth)
	cuentas.Get("/", handlers.ListCuentas)
	cuentas.Get("/:id", handlers.GetCuenta)
	cuentas.Post("/", middleware.RequireRole("ADMIN", "CONTADOR"), handlers.CreateCuenta)
	cuentas.Put("/:id", middleware.RequireRole("ADMIN", "CONTADOR"), handlers.UpdateCuenta)
	cuentas.Delete("/:id", middleware.RequireRole("ADMIN"), handlers.DeleteCuenta)

	// ── Cartolas ──────────────────────────────────────────────────────────────
	cartolas := api.Group("/cartolas", middleware.Auth)
	cartolas.Get("/", handlers.ListCartolas)
	cartolas.Get("/:id/movimientos", handlers.GetMovimientos)
	cartolas.Post("/importar", middleware.RequireRole("ADMIN", "CONTADOR", "OPERADOR"), handlers.ImportCartola)
	cartolas.Delete("/:id", middleware.RequireRole("ADMIN"), handlers.DeleteCartola)

	// ── Conciliación ──────────────────────────────────────────────────────────
	conc := api.Group("/conciliacion", middleware.Auth)
	conc.Post("/:id/auto", middleware.RequireRole("ADMIN", "CONTADOR", "OPERADOR"), handlers.EjecutarConciliacion)
	conc.Get("/:id/diferencias", handlers.GetDiferencias)
	conc.Get("/:id/resumen", handlers.GetResumenConciliacion)
	conc.Put("/item/:id/resolver", middleware.RequireRole("ADMIN", "CONTADOR"), handlers.ResolverDiferencia)

	// ── Entidades ─────────────────────────────────────────────────────────────
	entidades := api.Group("/entidades", middleware.Auth)
	entidades.Get("/", handlers.ListEntidades)
	entidades.Get("/:id", handlers.GetEntidad)
	entidades.Post("/", middleware.RequireRole("ADMIN", "CONTADOR"), handlers.CreateEntidad)
	entidades.Put("/:id", middleware.RequireRole("ADMIN", "CONTADOR"), handlers.UpdateEntidad)
	entidades.Delete("/:id", middleware.RequireRole("ADMIN"), handlers.DeleteEntidad)

	// ── Reportes ──────────────────────────────────────────────────────────────
	rep := api.Group("/reportes", middleware.Auth)
	rep.Get("/posicion", handlers.GetPosicion)
	rep.Get("/auditoria", middleware.RequireRole("ADMIN"), handlers.GetAudit)

	// ── Módulo Jurídico ───────────────────────────────────────────────────────
	if config.IsJuridico() {
		jur := api.Group("/juridico", middleware.Auth)

		custodia := jur.Group("/custodia")
		custodia.Get("/", profile.ListCustodia)
		custodia.Post("/", middleware.RequireRole("ADMIN", "CONTADOR", "OPERADOR"), profile.CreateCustodia)
		custodia.Put("/:id", middleware.RequireRole("ADMIN", "CONTADOR", "OPERADOR"), profile.UpdateCustodia)
		custodia.Delete("/:id", middleware.RequireRole("ADMIN"), profile.DeleteCustodia)
		custodia.Get("/cliente/:entidad_id", profile.SaldoCustodiaCliente)

		hon := jur.Group("/honorarios")
		hon.Get("/", profile.ListHonorarios)
		hon.Get("/resumen", profile.ResumenHonorarios)
		hon.Post("/", middleware.RequireRole("ADMIN", "CONTADOR", "OPERADOR"), profile.CreateHonorario)
		hon.Put("/:id", middleware.RequireRole("ADMIN", "CONTADOR", "OPERADOR"), profile.UpdateHonorario)
		hon.Delete("/:id", middleware.RequireRole("ADMIN"), profile.DeleteHonorario)
	}

	// ── Módulo Empresa ────────────────────────────────────────────────────────
	if config.IsEmpresa() {
		emp := api.Group("/empresa", middleware.Auth)

		cat := emp.Group("/categorias")
		cat.Get("/", profile.ListCategorias)
		cat.Post("/", middleware.RequireRole("ADMIN"), profile.CreateCategoria)
		cat.Put("/:id", middleware.RequireRole("ADMIN"), profile.UpdateCategoria)
		cat.Delete("/:id", middleware.RequireRole("ADMIN"), profile.DeleteCategoria)

		ing := emp.Group("/ingresos")
		ing.Get("/", profile.ListIngresos)
		ing.Get("/resumen", profile.ResumenIngresos)
		ing.Post("/", middleware.RequireRole("ADMIN", "CONTADOR", "OPERADOR"), profile.CreateIngreso)
		ing.Put("/:id", middleware.RequireRole("ADMIN", "CONTADOR", "OPERADOR"), profile.UpdateIngreso)
		ing.Delete("/:id", middleware.RequireRole("ADMIN"), profile.DeleteIngreso)
	}
}
