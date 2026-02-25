package models

import "time"

// ── Auth ─────────────────────────────────────────────────────────────────────

type Usuario struct {
	ID           int       `json:"id"`
	RUT          string    `json:"rut"`
	Nombre       string    `json:"nombre"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Rol          string    `json:"rol"`
	Activo       bool      `json:"activo"`
	CreatedAt    time.Time `json:"created_at"`
}

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Rol    string `json:"rol"`
}

// ── Cuentas ──────────────────────────────────────────────────────────────────

type CuentaBancaria struct {
	ID           int       `json:"id"`
	Banco        string    `json:"banco"`
	NumeroCuenta string    `json:"numero_cuenta"`
	Tipo         string    `json:"tipo"`
	Moneda       string    `json:"moneda"`
	Alias        string    `json:"alias"`
	Activa       bool      `json:"activa"`
	CreatedAt    time.Time `json:"created_at"`
}

// ── Cartolas ─────────────────────────────────────────────────────────────────

type Cartola struct {
	ID             int       `json:"id"`
	CuentaID       int       `json:"cuenta_id"`
	PeriodoDesde   string    `json:"periodo_desde"`
	PeriodoHasta   string    `json:"periodo_hasta"`
	ArchivoNombre  string    `json:"archivo_nombre"`
	ImportadoPor   int       `json:"importado_por"`
	ImportadoAt    time.Time `json:"importado_at"`
	TotalMov       int       `json:"total_movimientos,omitempty"`
}

type Movimiento struct {
	ID          int       `json:"id"`
	CartolaID   int       `json:"cartola_id"`
	Fecha       string    `json:"fecha"`
	Descripcion string    `json:"descripcion"`
	Referencia  string    `json:"referencia"`
	Debe        float64   `json:"debe"`
	Haber       float64   `json:"haber"`
	Saldo       float64   `json:"saldo"`
	Conciliado  bool      `json:"conciliado"`
	CreatedAt   time.Time `json:"created_at"`
}

// ── Conciliación ─────────────────────────────────────────────────────────────

type ConciliacionItem struct {
	ID            int       `json:"id"`
	CartolaID     int       `json:"cartola_id"`
	MovimientoID  *int      `json:"movimiento_id"`
	Tipo          string    `json:"tipo"`
	Estado        string    `json:"estado"` // pendiente|conciliado|diferencia
	Observacion   string    `json:"observacion"`
	ResueltoAt    *time.Time `json:"resuelto_at"`
}

type ResumenConciliacion struct {
	CartolaID      int     `json:"cartola_id"`
	TotalMov       int     `json:"total_movimientos"`
	Conciliados    int     `json:"conciliados"`
	Diferencias    int     `json:"diferencias"`
	Pendientes     int     `json:"pendientes"`
	PctConciliado  float64 `json:"pct_conciliado"`
}

// ── Entidades ─────────────────────────────────────────────────────────────────

type Entidad struct {
	ID        int       `json:"id"`
	Tipo      string    `json:"tipo"`
	RUT       string    `json:"rut"`
	Nombre    string    `json:"nombre"`
	Email     string    `json:"email"`
	Telefono  string    `json:"telefono"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Módulo Jurídico ───────────────────────────────────────────────────────────

type Custodia struct {
	ID           int       `json:"id"`
	EntidadID    int       `json:"entidad_id"`
	EntidadNombre string   `json:"entidad_nombre,omitempty"`
	Descripcion  string    `json:"descripcion"`
	Monto        float64   `json:"monto"`
	FechaIngreso string    `json:"fecha_ingreso"`
	FechaDevol   string    `json:"fecha_devolucion"`
	Estado       string    `json:"estado"`
	CuentaID     *int      `json:"cuenta_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type Honorario struct {
	ID           int       `json:"id"`
	EntidadID    int       `json:"entidad_id"`
	EntidadNombre string   `json:"entidad_nombre,omitempty"`
	Concepto     string    `json:"concepto"`
	Monto        float64   `json:"monto"`
	FechaEmision string    `json:"fecha_emision"`
	FechaPago    string    `json:"fecha_pago"`
	Estado       string    `json:"estado"` // pendiente|pagado|anulado
	MovimientoID *int      `json:"movimiento_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// ── Módulo Empresa ────────────────────────────────────────────────────────────

type Categoria struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Tipo   string `json:"tipo"` // ingreso|egreso
	Activa bool   `json:"activa"`
}

type IngresoRecurrente struct {
	ID           int       `json:"id"`
	EntidadID    int       `json:"entidad_id"`
	EntidadNombre string   `json:"entidad_nombre,omitempty"`
	CategoriaID  int       `json:"categoria_id"`
	CategoriaNombre string `json:"categoria_nombre,omitempty"`
	Descripcion  string    `json:"descripcion"`
	Monto        float64   `json:"monto"`
	Periodicidad string    `json:"periodicidad"`
	DiaCobro     int       `json:"dia_cobro"`
	Estado       string    `json:"estado"` // esperado|recibido
	Activo       bool      `json:"activo"`
	CreatedAt    time.Time `json:"created_at"`
}

// ── Reportes ─────────────────────────────────────────────────────────────────

type PosicionFinanciera struct {
	Cuenta   string  `json:"cuenta"`
	Banco    string  `json:"banco"`
	Ingresos float64 `json:"ingresos"`
	Egresos  float64 `json:"egresos"`
	Saldo    float64 `json:"saldo"`
}

type AuditEntry struct {
	ID         int       `json:"id"`
	UsuarioID  int       `json:"usuario_id"`
	Accion     string    `json:"accion"`
	Tabla      string    `json:"tabla"`
	RegistroID string    `json:"registro_id"`
	Payload    string    `json:"payload"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}
