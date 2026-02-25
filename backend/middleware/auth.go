package middleware

import (
	"conciliacion-bancaria/config"
	"conciliacion-bancaria/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token requerido"})
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token inválido"})
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "claims inválidos"})
	}

	claims := models.Claims{
		UserID: int(mapClaims["user_id"].(float64)),
		Email:  mapClaims["email"].(string),
		Rol:    mapClaims["rol"].(string),
	}

	c.Locals("user", claims)
	return c.Next()
}

// CurrentUser extrae los claims del contexto Fiber.
func CurrentUser(c *fiber.Ctx) models.Claims {
	return c.Locals("user").(models.Claims)
}
