package config

import "os"

var (
	Profile   string // "juridico" | "empresa"
	Port      string
	JWTSecret string
	DBPath    string
)

func Load() {
	Profile = getEnv("APP_PROFILE", "juridico")
	Port = getEnv("PORT", "3032")
	JWTSecret = getEnv("JWT_SECRET", "dev-secret-change-in-production")
	DBPath = getEnv("DB_PATH", "./conciliacion.db")
}

func IsJuridico() bool { return Profile == "juridico" }
func IsEmpresa() bool  { return Profile == "empresa" }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
