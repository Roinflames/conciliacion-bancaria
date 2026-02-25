package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func Init(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		log.Fatalf("db: no se pudo abrir %s: %v", path, err)
	}
	DB.SetMaxOpenConns(1)

	if err = DB.Ping(); err != nil {
		log.Fatalf("db: ping falló: %v", err)
	}

	migrate()
	seedAdmin()
	log.Println("db: inicializada correctamente")
}

func migrate() {
	schema := `
	-- Auth
	CREATE TABLE IF NOT EXISTS usuarios (
		id            INTEGER  PRIMARY KEY AUTOINCREMENT,
		rut           TEXT     UNIQUE NOT NULL,
		nombre        TEXT     NOT NULL,
		email         TEXT     UNIQUE NOT NULL,
		password_hash TEXT     NOT NULL,
		rol           TEXT     NOT NULL DEFAULT 'READONLY',
		activo        INTEGER  NOT NULL DEFAULT 1,
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Cuentas bancarias
	CREATE TABLE IF NOT EXISTS cuentas_bancarias (
		id            INTEGER  PRIMARY KEY AUTOINCREMENT,
		banco         TEXT     NOT NULL,
		numero_cuenta TEXT     NOT NULL,
		tipo          TEXT     NOT NULL DEFAULT 'corriente',
		moneda        TEXT     NOT NULL DEFAULT 'CLP',
		alias         TEXT,
		activa        INTEGER  NOT NULL DEFAULT 1,
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Cartolas (una por importación)
	CREATE TABLE IF NOT EXISTS cartolas (
		id             INTEGER  PRIMARY KEY AUTOINCREMENT,
		cuenta_id      INTEGER  NOT NULL REFERENCES cuentas_bancarias(id),
		periodo_desde  DATE     NOT NULL,
		periodo_hasta  DATE     NOT NULL,
		archivo_nombre TEXT,
		importado_por  INTEGER  NOT NULL REFERENCES usuarios(id),
		importado_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Movimientos (filas de una cartola)
	CREATE TABLE IF NOT EXISTS movimientos (
		id          INTEGER  PRIMARY KEY AUTOINCREMENT,
		cartola_id  INTEGER  NOT NULL REFERENCES cartolas(id),
		fecha       DATE     NOT NULL,
		descripcion TEXT,
		referencia  TEXT,
		debe        REAL     NOT NULL DEFAULT 0,
		haber       REAL     NOT NULL DEFAULT 0,
		saldo       REAL,
		conciliado  INTEGER  NOT NULL DEFAULT 0,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_mov_cartola ON movimientos(cartola_id);
	CREATE INDEX IF NOT EXISTS idx_mov_fecha   ON movimientos(fecha);

	-- Ítems de conciliación
	CREATE TABLE IF NOT EXISTS conciliacion_items (
		id              INTEGER  PRIMARY KEY AUTOINCREMENT,
		cartola_id      INTEGER  NOT NULL REFERENCES cartolas(id),
		movimiento_id   INTEGER  REFERENCES movimientos(id),
		tipo            TEXT     NOT NULL DEFAULT 'pendiente',
		estado          TEXT     NOT NULL DEFAULT 'pendiente',
		observacion     TEXT,
		resuelto_por    INTEGER  REFERENCES usuarios(id),
		resuelto_at     DATETIME,
		created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Entidades (clientes o miembros según perfil)
	CREATE TABLE IF NOT EXISTS entidades (
		id         INTEGER  PRIMARY KEY AUTOINCREMENT,
		tipo       TEXT     NOT NULL DEFAULT 'cliente',
		rut        TEXT     UNIQUE,
		nombre     TEXT     NOT NULL,
		email      TEXT,
		telefono   TEXT,
		activo     INTEGER  NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ── MÓDULO JURÍDICO ──────────────────────────────────────────────────────

	CREATE TABLE IF NOT EXISTS custodia (
		id            INTEGER  PRIMARY KEY AUTOINCREMENT,
		entidad_id    INTEGER  NOT NULL REFERENCES entidades(id),
		descripcion   TEXT     NOT NULL,
		monto         REAL     NOT NULL,
		fecha_ingreso DATE     NOT NULL,
		fecha_devol   DATE,
		estado        TEXT     NOT NULL DEFAULT 'activo',
		cuenta_id     INTEGER  REFERENCES cuentas_bancarias(id),
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS honorarios (
		id            INTEGER  PRIMARY KEY AUTOINCREMENT,
		entidad_id    INTEGER  NOT NULL REFERENCES entidades(id),
		concepto      TEXT     NOT NULL,
		monto         REAL     NOT NULL,
		fecha_emision DATE     NOT NULL,
		fecha_pago    DATE,
		estado        TEXT     NOT NULL DEFAULT 'pendiente',
		movimiento_id INTEGER  REFERENCES movimientos(id),
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ── MÓDULO EMPRESA ───────────────────────────────────────────────────────

	CREATE TABLE IF NOT EXISTS categorias (
		id     INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT    UNIQUE NOT NULL,
		tipo   TEXT    NOT NULL DEFAULT 'ingreso',
		activa INTEGER NOT NULL DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS ingresos_recurrentes (
		id           INTEGER  PRIMARY KEY AUTOINCREMENT,
		entidad_id   INTEGER  NOT NULL REFERENCES entidades(id),
		categoria_id INTEGER  NOT NULL REFERENCES categorias(id),
		descripcion  TEXT     NOT NULL,
		monto        REAL     NOT NULL,
		periodicidad TEXT     NOT NULL DEFAULT 'mensual',
		dia_cobro    INTEGER,
		estado       TEXT     NOT NULL DEFAULT 'esperado',
		activo       INTEGER  NOT NULL DEFAULT 1,
		created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ── AUDIT LOG ────────────────────────────────────────────────────────────

	CREATE TABLE IF NOT EXISTS audit_log (
		id          INTEGER  PRIMARY KEY AUTOINCREMENT,
		usuario_id  INTEGER  REFERENCES usuarios(id),
		accion      TEXT     NOT NULL,
		tabla       TEXT     NOT NULL,
		registro_id TEXT     NOT NULL,
		payload     TEXT,
		ip          TEXT,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_audit_tabla   ON audit_log(tabla);
	CREATE INDEX IF NOT EXISTS idx_audit_usuario ON audit_log(usuario_id);
	CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at);
	`

	if _, err := DB.Exec(schema); err != nil {
		log.Fatalf("db: migración falló: %v", err)
	}
}

func seedAdmin() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM usuarios").Scan(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin1234"), 12)
	if err != nil {
		log.Fatalf("db: no se pudo hashear contraseña seed: %v", err)
	}

	_, err = DB.Exec(
		`INSERT INTO usuarios (rut, nombre, email, password_hash, rol) VALUES (?, ?, ?, ?, ?)`,
		"00000000-0", "Administrador", "admin@conciliacion.local", string(hash), "ADMIN",
	)
	if err != nil {
		log.Fatalf("db: no se pudo crear admin inicial: %v", err)
	}
	log.Println("db: usuario ADMIN creado — email: admin@conciliacion.local / pass: admin1234")
}

// Audit registra una operación en el log de auditoría.
func Audit(usuarioID int, accion, tabla, registroID, payload, ip string) {
	DB.Exec(
		`INSERT INTO audit_log (usuario_id, accion, tabla, registro_id, payload, ip, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		usuarioID, accion, tabla, registroID, payload, ip, time.Now(),
	)
}
