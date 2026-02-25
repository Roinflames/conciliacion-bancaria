# conciliacion-bancaria

## Descripción
Módulo de conciliación bancaria para estudios jurídicos. Automatiza la conciliación de cuentas corrientes, fondos en custodia de clientes y honorarios.

## Cliente objetivo
Alfaro & Madariaga (en conversación)

## Estado
v0.0.1 — En definición. Sin desarrollo iniciado.

## Problema que resuelve
- Conciliación manual de cuentas corrientes del estudio
- Control de fondos en custodia por causa y cliente
- Seguimiento de honorarios pactados, cobrados y pendientes
- Detección de diferencias entre cartola bancaria y registros internos

## Stack propuesto
- Backend: Go + Fiber
- Base de datos: PostgreSQL
- Frontend: Vanilla JS
- Exportación: PDF

## Instalación y arranque

1. `git clone` y sitúate en el repo.
2. Instala dependencias de Go desde `backend` ejecutando `go mod download`.
3. Ajusta el perfil y puertos exportando `APP_PROFILE`, `PORT`, `JWT_SECRET` y `DB_PATH` o deja los valores por defecto.
4. Usa `scripts/start.sh` para compilar, iniciar el servidor y aprovechar la notificación opcional a Discord que el script ya maneja.

## Notificaciones a Discord

- El script local `scripts/start.sh` lee `~/.env_discord` si existe y usa la variable `DISCORD_WEBHOOK_URL` para avisarte cuando la compilación termina o falla.
- El workflow `.github/workflows/notify.yml` se dispara con cada `push` a `main` y publica un embed en el mismo webhook con autor, rama y mensaje del commit.
- Define el secreto `DISCORD_WEBHOOK_URL` en la configuración del repositorio para que CI pueda notificar correctamente.

## Próximos pasos
1. Reunión de levantamiento con cliente
2. Definición de requerimientos y flujos
3. SRS / propuesta técnica y económica
4. Inicio de desarrollo

## Historial
- v0.0.1 (2026-02-25): Registro inicial del proyecto en portafolio
