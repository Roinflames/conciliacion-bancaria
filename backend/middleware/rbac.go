package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// RequireRole devuelve un middleware que verifica que el usuario tenga uno de los roles indicados.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *fiber.Ctx) error {
		user := CurrentUser(c)
		if !allowed[user.Rol] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "no tienes permisos para esta acción",
			})
		}
		return c.Next()
	}
}
