# Especificación de Requisitos de Software (SRS)
## Formato IEEE 830

**Proyecto:** conciliacion-bancaria
**Version:** v0.2.0
**Fecha:** 2026-02-25
**Estado:** Borrador — pendiente validación con clientes
**Clientes objetivo:** Alfaro & Madariaga (estudio jurídico) · Comunidad Virtual (empresa/comunidad)

## Control de cambios
| Version | Fecha | Cambio |
|---|---|---|
| v0.1.0 | 2026-02-25 | Borrador inicial enfocado en estudio jurídico. |
| v0.2.0 | 2026-02-25 | Generalización con perfiles de negocio configurables: `juridico` y `empresa`. Módulos de dominio condicionales según perfil. |

---

## 1. Introducción

### 1.1 Propósito
Definir los requisitos funcionales y no funcionales del sistema `conciliacion-bancaria`, una plataforma de gestión y conciliación financiera adaptable a distintos tipos de organización mediante **perfiles de negocio** configurables. El sistema cubre desde estudios jurídicos hasta empresas y comunidades con ingresos recurrentes.

### 1.2 Alcance
El sistema permite:
- Registrar y gestionar múltiples cuentas bancarias de la organización.
- Importar cartolas bancarias desde archivos del banco (TXT, CSV, Excel).
- Conciliar movimientos bancarios con registros internos automáticamente.
- Gestionar categorías de ingreso y egreso según el perfil activo.
- Detectar y alertar diferencias entre cartola y contabilidad interna.
- Generar reportes de posición financiera y pista de auditoría completa.
- Activar módulos de dominio específicos según el perfil configurado.

**Módulos de dominio por perfil:**

| Módulo | Perfil `juridico` | Perfil `empresa` |
|---|---|---|
| Fondos en custodia | Activo | Inactivo |
| Honorarios | Activo | Inactivo |
| Ingresos recurrentes | Inactivo | Activo |
| Categorías personalizadas | Inactivo | Activo |

### 1.3 Definiciones, acrónimos y abreviaturas
- `SRS`: Software Requirements Specification.
- `Cartola`: extracto bancario oficial con los movimientos de una cuenta.
- `Conciliación`: proceso de contrastar y hacer coincidir los registros internos con la cartola bancaria.
- `Perfil de negocio`: configuración que activa o desactiva módulos de dominio según el tipo de organización.
- `Fondos en custodia`: dineros de clientes retenidos temporalmente por el estudio jurídico para uso en causas. Deben mantenerse separados de los fondos propios. *(Solo perfil `juridico`)*
- `Honorarios pactados/cobrados`: montos acordados y recibidos por servicios jurídicos. *(Solo perfil `juridico`)*
- `Ingresos recurrentes`: cuotas de membresía, suscripciones u otros ingresos periódicos. *(Solo perfil `empresa`)*
- `Categoría`: etiqueta de clasificación de movimientos definida por el administrador. *(Solo perfil `empresa`)*
- `RIT`: Rol Interno del Tribunal — identificador de causa judicial. *(Solo perfil `juridico`)*
- `Diferencia`: transacción presente en la cartola pero no en el sistema, o viceversa.
- `RBAC`: Role-Based Access Control.
- `JWT`: JSON Web Token.

### 1.4 Referencias
- IEEE Std 830-1998.
- ISO/IEC 29148:2018 — Systems and software engineering — Requirements engineering.
- Repositorio: `clients/conciliacion-bancaria-v0.0.1`.

### 1.5 Visión general
Este documento cubre descripción general, interfaces, requisitos funcionales y no funcionales, reglas de negocio y criterios de aceptación del MVP para ambos perfiles.

---

## 2. Descripción general

### 2.1 Perspectiva del producto
Aplicación web tipo monorepo:
- `backend/`: Go + Fiber, API REST, autenticación JWT.
- `frontend/`: Vanilla JS, sin frameworks pesados.
- `base de datos`: PostgreSQL (VPS/cloud) o SQLite (instalación local).
- `configuración de perfil`: variable de entorno `APP_PROFILE=juridico|empresa` — activa o desactiva módulos en tiempo de arranque.

### 2.2 Funciones del producto

**Núcleo (ambos perfiles)**
- Gestión de cuentas bancarias.
- Importación y parseo de cartolas bancarias.
- Motor de conciliación automática con revisión manual de diferencias.
- Reportes de posición financiera por período y cuenta.
- Pista de auditoría completa.

**Perfil `juridico`**
- Control de fondos en custodia por cliente y causa (RIT).
- Gestión de honorarios por causa: pactados, cobrados y pendientes.

**Perfil `empresa`**
- Registro de ingresos recurrentes (cuotas, suscripciones, membresías).
- Categorías personalizadas de ingreso y egreso.
- Dashboard de ingresos por tipo de producto o plan.

### 2.3 Características de usuarios

| Rol | Descripción | Ambos perfiles |
|---|---|---|
| `ADMIN` | Acceso total: configuración, cuentas, usuarios, reportes. | Sí |
| `CONTADOR` | Conciliación, importación, reportes y módulos de dominio. | Sí |
| `OPERADOR` | Registro de movimientos de dominio (honorarios, cuotas, custodia). | Sí |
| `READONLY` | Solo visualización. Sin acciones. | Sí |

*En perfil `juridico`, `OPERADOR` equivale al rol del abogado que registra honorarios y custodia de sus causas propias.*

### 2.4 Restricciones
- PostgreSQL o SQLite como base de datos (configurable por entorno).
- Sin integración bancaria directa (Open Banking) en MVP — importación por archivo.
- Operación local o en VPS propio — sin dependencia de SaaS de terceros.
- Cumplimiento con Ley 19.628 de protección de datos personales (Chile).
- Un usuario OPERADOR no puede ver movimientos de dominio que no le pertenecen.

### 2.5 Supuestos y dependencias
- La organización tiene al menos una cuenta corriente bancaria en banco chileno.
- Las cartolas se obtienen desde el portal del banco en formato descargable.
- No se requiere integración con SII ni facturación electrónica en MVP.
- El perfil se configura una sola vez al instalar. No cambia en producción.

---

## 3. Requisitos específicos

### 3.1 Requisitos de interfaces externas

#### 3.1.1 Interfaz de usuario
- Login con email y contraseña.
- Dashboard con resumen de posición financiera (contenido varía según perfil).
- Vista de conciliación con panel de diferencias.
- Vista de módulos de dominio (custodia u honorarios, o ingresos y categorías).
- Vista de reportes exportables.

#### 3.1.2 Interfaz de software (API REST)

**Autenticación**
- `POST /auth/login`
- `POST /auth/logout`
- `GET  /auth/me`

**Cuentas bancarias**
- `GET    /cuentas`
- `POST   /cuentas`
- `PUT    /cuentas/:id`
- `DELETE /cuentas/:id`

**Cartolas**
- `POST   /cartolas/importar`
- `GET    /cartolas`
- `GET    /cartolas/:id/movimientos`
- `DELETE /cartolas/:id`

**Conciliación**
- `GET    /conciliacion`
- `POST   /conciliacion/ejecutar`
- `GET    /conciliacion/:id/diferencias`
- `PUT    /conciliacion/diferencias/:id/resolver`
- `GET    /conciliacion/:id/resumen`

**Módulo custodia** *(solo perfil `juridico`)*
- `GET    /custodia`
- `POST   /custodia`
- `PUT    /custodia/:id`
- `DELETE /custodia/:id`
- `GET    /custodia/cliente/:cliente_id`
- `GET    /custodia/causa/:rit`

**Módulo honorarios** *(solo perfil `juridico`)*
- `GET    /honorarios`
- `POST   /honorarios`
- `PUT    /honorarios/:id`
- `DELETE /honorarios/:id`
- `GET    /honorarios/causa/:rit`

**Módulo ingresos recurrentes** *(solo perfil `empresa`)*
- `GET    /ingresos`
- `POST   /ingresos`
- `PUT    /ingresos/:id`
- `DELETE /ingresos/:id`
- `GET    /ingresos/resumen`

**Módulo categorías** *(solo perfil `empresa`)*
- `GET    /categorias`
- `POST   /categorias`
- `PUT    /categorias/:id`
- `DELETE /categorias/:id`

**Clientes / Miembros**
- `GET    /entidades`
- `POST   /entidades`
- `PUT    /entidades/:id`
- `DELETE /entidades/:id`

*El recurso `/entidades` representa clientes en perfil `juridico` y miembros/empresas en perfil `empresa`.*

**Reportes**
- `GET    /reportes/posicion?desde=&hasta=`
- `GET    /reportes/dominio?desde=&hasta=` *(custodia+honorarios o ingresos+categorías)*
- `GET    /reportes/auditoria?desde=&hasta=`
- `POST   /reportes/exportar`

**Sistema**
- `GET    /health`
- `GET    /config/perfil`

#### 3.1.3 Interfaz de comunicaciones
- HTTP/HTTPS para UI y API.
- Importación de archivos vía `multipart/form-data`.
- Exportación de reportes vía descarga de PDF y CSV.

### 3.2 Requisitos funcionales

#### Núcleo — ambos perfiles

**Autenticación y acceso**
- `RF-01`: El sistema debe permitir login con email y contraseña con autenticación JWT.
- `RF-02`: El sistema debe controlar el acceso a funcionalidades según el rol del usuario (RBAC).
- `RF-03`: El sistema debe permitir al ADMIN crear, editar y desactivar usuarios.

**Cuentas bancarias**
- `RF-04`: El sistema debe permitir registrar múltiples cuentas bancarias con tipo, banco, número y moneda.
- `RF-05`: El sistema debe exponer el saldo actual de cada cuenta calculado desde los movimientos registrados.

**Importación de cartolas**
- `RF-06`: El sistema debe permitir importar cartolas en formato TXT, CSV y Excel (.xlsx).
- `RF-07`: El sistema debe parsear automáticamente fecha, descripción, monto y saldo de cada movimiento.
- `RF-08`: El sistema debe detectar y rechazar cartolas duplicadas (misma cuenta, mismo período).
- `RF-09`: El sistema debe asociar cada cartola a una cuenta bancaria registrada.

**Conciliación**
- `RF-10`: El sistema debe ejecutar conciliación automática comparando movimientos de la cartola con registros internos del mismo período.
- `RF-11`: El sistema debe clasificar cada movimiento como `CONCILIADO`, `SOLO_CARTOLA` o `SOLO_SISTEMA`.
- `RF-12`: El sistema debe permitir resolver manualmente las diferencias, registrando usuario y justificación.
- `RF-13`: El sistema debe calcular saldo conciliado, diferencias totales y porcentaje de conciliación por período.
- `RF-14`: El sistema debe impedir cerrar un período con diferencias no resueltas sin confirmación explícita del ADMIN o CONTADOR.

**Reportes y auditoría**
- `RF-15`: El sistema debe generar reporte de posición financiera por período: ingresos, egresos y saldo por cuenta.
- `RF-16`: El sistema debe generar pista de auditoría con usuario, IP, fecha y acción de cada operación.
- `RF-17`: El sistema debe exportar todos los reportes en formato PDF y CSV.

#### Perfil `juridico`

**Fondos en custodia**
- `RF-18`: El sistema debe permitir registrar ingresos y egresos de fondos en custodia vinculados a una entidad (cliente) y opcionalmente a una causa (RIT).
- `RF-19`: El sistema debe mostrar el saldo de custodia por cliente y por causa en tiempo real.
- `RF-20`: El sistema debe alertar cuando el saldo de custodia de un cliente llegue a cero o sea negativo.
- `RF-21`: El sistema debe distinguir cuentas tipo `ESTUDIO` y tipo `CUSTODIA` e impedir mezclar fondos entre ellas en los reportes.

**Honorarios**
- `RF-22`: El sistema debe permitir registrar honorarios pactados por causa y cliente: monto, fecha y forma de pago.
- `RF-23`: El sistema debe registrar pagos de honorarios recibidos y calcular el saldo pendiente automáticamente.
- `RF-24`: El sistema debe alertar cuando un honorario pendiente supere N días sin pago (N configurable por ADMIN).
- `RF-25`: El sistema debe generar reporte de honorarios: pactados, cobrados y pendientes por período.

#### Perfil `empresa`

**Ingresos recurrentes**
- `RF-26`: El sistema debe permitir registrar ingresos recurrentes (cuotas, suscripciones, membresías) vinculados a una entidad (miembro o empresa).
- `RF-27`: El sistema debe detectar ingresos esperados no recibidos en el período y alertar al CONTADOR.
- `RF-28`: El sistema debe generar resumen de ingresos por tipo de plan o producto para el período seleccionado.

**Categorías**
- `RF-29`: El sistema debe permitir al ADMIN crear, editar y eliminar categorías de ingreso y egreso.
- `RF-30`: El sistema debe permitir clasificar cualquier movimiento interno con una categoría.
- `RF-31`: El sistema debe generar reporte de egresos e ingresos agrupados por categoría.

### 3.3 Requisitos no funcionales

**Seguridad**
- `RNF-01`: Toda contraseña debe almacenarse con hash bcrypt (costo mínimo 12).
- `RNF-02`: Los tokens JWT deben expirar en 8 horas. Refresh token de 30 días.
- `RNF-03`: Toda operación sensible debe registrarse en auditoría con usuario, IP y timestamp.
- `RNF-04`: Un usuario OPERADOR no puede acceder a movimientos de dominio que no le pertenecen.

**Disponibilidad**
- `RNF-05`: El sistema debe exponer `GET /health` para monitoreo.
- `RNF-06`: El backend debe soportar reinicio automático (PM2 o systemd).

**Rendimiento**
- `RNF-07`: Importación de cartola con hasta 500 movimientos: menos de 5 segundos.
- `RNF-08`: Conciliación de un mes completo: menos de 10 segundos.

**Portabilidad**
- `RNF-09`: El sistema debe funcionar en local (SQLite) y en VPS Linux (PostgreSQL) cambiando solo variables de entorno.
- `RNF-10`: El perfil de negocio se configura con `APP_PROFILE=juridico|empresa` al arrancar.
- `RNF-11`: El frontend debe funcionar en Chrome, Firefox y Edge (últimas 2 versiones).

**Mantenibilidad**
- `RNF-12`: El backend debe tener logs estructurados con nivel INFO, WARN y ERROR.
- `RNF-13`: Las migraciones de base de datos deben estar versionadas y ser reproducibles.

---

## 4. Reglas de negocio

| ID | Aplica a | Regla |
|---|---|---|
| `RN-01` | `juridico` | Los fondos en custodia nunca deben mezclarse con los fondos propios del estudio. Deben operar en cuentas bancarias distintas. |
| `RN-02` | Ambos | Un período conciliado no puede reabrirse sin rol ADMIN y justificación registrada. |
| `RN-03` | `juridico` | Un movimiento de custodia siempre debe tener cliente asociado. La causa (RIT) es opcional. |
| `RN-04` | Ambos | El saldo inicial de una cuenta debe ingresarse manualmente al crear la cuenta. |
| `RN-05` | Ambos | No se puede eliminar una cuenta bancaria con cartolas o movimientos de dominio asociados. |
| `RN-06` | `juridico` | Honorarios en UF o USD deben registrar el valor de conversión al momento del pago. |
| `RN-07` | `empresa` | Un ingreso recurrente puede marcarse como `ESPERADO` o `RECIBIDO`. Solo los `RECIBIDO` impactan el saldo. |

---

## 5. Criterios de aceptación del MVP

| ID | Perfil | Criterio |
|---|---|---|
| `CA-01` | Ambos | Importar una cartola en CSV y registrar todos los movimientos correctamente. |
| `CA-02` | Ambos | Conciliación automática identifica coincidencias y diferencias para un mes completo. |
| `CA-03` | Ambos | Reporte de posición financiera muestra saldos correctos para el período seleccionado. |
| `CA-04` | Ambos | Exportación PDF de cualquier reporte se genera sin errores. |
| `CA-05` | Ambos | Un usuario OPERADOR no puede ver datos de dominio que no le pertenecen. |
| `CA-06` | Ambos | Pista de auditoría registra correctamente todas las operaciones de una sesión de prueba. |
| `CA-07` | `juridico` | El saldo de custodia de un cliente se actualiza correctamente tras un ingreso y un egreso. |
| `CA-08` | `juridico` | El reporte de honorarios muestra correctamente pactados, cobrados y pendientes. |
| `CA-09` | `empresa` | Un ingreso recurrente marcado como `RECIBIDO` impacta el saldo de la cuenta correctamente. |
| `CA-10` | `empresa` | El reporte de ingresos agrupa correctamente por tipo de plan para el período seleccionado. |

---

## 6. Fuera de alcance (MVP)

- Integración bancaria directa (Open Banking / API del banco).
- Emisión de facturas o boletas electrónicas (SII).
- Contabilidad completa (libro mayor, balance general).
- Aplicación móvil nativa.
- Integración automática con PJUD Scraping (fase posterior).
- Multi-empresa (varias razones sociales en una misma instalación).
- Cambio de perfil de negocio en caliente (requiere reinstalación).

---

## 7. Trazabilidad de versiones

| Version | Fecha | Cambios | Requisitos impactados |
|---|---|---|---|
| v0.1.0 | 2026-02-25 | Borrador inicial para estudio jurídico | RF-01..RF-27, RNF-01..RNF-13 |
| v0.2.0 | 2026-02-25 | Generalización con perfiles `juridico` y `empresa`; `/entidades` unifica clientes/miembros; RF-26..RF-31 nuevos para perfil empresa | RF-26..RF-31, RNF-10, CA-09..CA-10 |
