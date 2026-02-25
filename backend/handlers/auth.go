package handlers

import (
	"conciliacion-bancaria/config"
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/middleware"
	"conciliacion-bancaria/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var body loginInput
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cuerpo inválido"})
	}

	var u models.Usuario
	err := db.DB.QueryRow(
		`SELECT id, email, password_hash, nombre, rol, activo FROM usuarios WHERE email = ?`,
		body.Email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Nombre, &u.Rol, &u.Activo)

	if err != nil || !u.Activo {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "credenciales inválidas"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "credenciales inválidas"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"email":   u.Email,
		"rol":     u.Rol,
		"exp":     time.Now().Add(8 * time.Hour).Unix(),
	})

	signed, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error generando token"})
	}

	return c.JSON(fiber.Map{
		"token": signed,
		"user": fiber.Map{
			"id":     u.ID,
			"nombre": u.Nombre,
			"email":  u.Email,
			"rol":    u.Rol,
		},
	})
}

func Me(c *fiber.Ctx) error {
	claims := middleware.CurrentUser(c)

	var u models.Usuario
	err := db.DB.QueryRow(
		`SELECT id, rut, nombre, email, rol, activo, created_at FROM usuarios WHERE id = ?`,
		claims.UserID,
	).Scan(&u.ID, &u.RUT, &u.Nombre, &u.Email, &u.Rol, &u.Activo, &u.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "usuario no encontrado"})
	}

	return c.JSON(u)
}

func Logout(c *fiber.Ctx) error {
	// JWT es stateless; el cliente descarta el token.
	return c.JSON(fiber.Map{"message": "sesión cerrada"})
}
