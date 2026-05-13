# Manual operativo del modulo de seguridad VPS

Fecha de actualizacion: 2026-04-16

## Objetivo

Este modulo permite auditar un VPS Linux Ubuntu desde el panel super y tambien desde consola, con historial de ejecuciones, comparacion contra el escaneo anterior y exportes en JSON, TXT, HTML, CSV, PDF y XLS.

## Componentes principales

- Panel super: `/super/seguridad.html`.
- API protegida: `/super/api/security/vps/config`, `/run`, `/status`, `/history`, `/report`, `/compare`.
- Configuracion persistente: `backend/secure/vps_security_config.json`.
- Historial y reportes: `backend/logs/vps_security/runs/<scan_id>/`.
- CLI Go: `backend/tools/vps_security_scan/main.go`.
- Scripts Linux:
  - `scripts/install_vps_security_tools.sh`
  - `scripts/run_vps_security_scan.sh`
  - `scripts/install_vps_security_cron.sh`

## Herramientas evaluadas

- Lynis: hardening y auditoria del sistema.
- Nmap: deteccion de puertos y servicios expuestos.
- Trivy: alternativa ligera a OpenVAS para vulnerabilidades y misconfiguracion.
- Chequeos propios: firewall, SSH, Nginx, permisos criticos, servicios y actualizaciones pendientes.

## Instalacion en Ubuntu

1. Dar permisos de ejecucion a los scripts:

```bash
chmod +x scripts/install_vps_security_tools.sh scripts/run_vps_security_scan.sh scripts/install_vps_security_cron.sh
```

2. Instalar dependencias base en el VPS:

```bash
sudo ./scripts/install_vps_security_tools.sh
```

3. Revisar o ajustar la configuracion inicial generada en:

```bash
backend/secure/vps_security_config.json
```

4. Opcional: compilar el binario dedicado para no depender de `go run`:

```bash
cd backend
go build -o ./bin/vps_security_scan_linux_amd64 ./tools/vps_security_scan
```

## Ejecucion manual

Desde la raiz del repositorio:

```bash
./scripts/run_vps_security_scan.sh --trigger manual --triggered-by super_admin
```

Con override de host, puertos o perfil:

```bash
./scripts/run_vps_security_scan.sh --target 127.0.0.1 --ports 49222,80,443,8080 --profile full --trigger manual --triggered-by super_admin
```

Tambien puede ejecutarse directamente la CLI:

```bash
cd backend
go run ./tools/vps_security_scan --config ./secure/vps_security_config.json --target 127.0.0.1 --ports 49222,80,443 --profile quick
```

## Programacion por cron

Instalar una ejecucion diaria a las 02:00:

```bash
sudo ./scripts/install_vps_security_cron.sh "0 2 * * *"
```

La salida del cron quedara en:

```bash
backend/logs/vps_security/cron.log
```

## Ejemplo de salida de consola

```text
SCAN_ID=vps-20260416-021530-48321
GENERATED_AT=2026-04-16T02:15:34Z
TARGET=127.0.0.1
PROFILE=full
HEALTH=warning
TOTAL_FINDINGS=12
REPORT_JSON=/opt/powerfulcontrolsystem/backend/logs/vps_security/runs/vps-20260416-021530-48321/reports/report.json
REPORT_TXT=/opt/powerfulcontrolsystem/backend/logs/vps_security/runs/vps-20260416-021530-48321/reports/report.txt
REPORT_HTML=/opt/powerfulcontrolsystem/backend/logs/vps_security/runs/vps-20260416-021530-48321/reports/report.html
REPORT_PDF=/opt/powerfulcontrolsystem/backend/logs/vps_security/runs/vps-20260416-021530-48321/reports/report.pdf
REPORT_XLS=/opt/powerfulcontrolsystem/backend/logs/vps_security/runs/vps-20260416-021530-48321/reports/report.xls
```

## Ejemplo resumido de reporte JSON

```json
{
  "scan_id": "vps-20260416-021530-48321",
  "target_host": "127.0.0.1",
  "profile": "full",
  "summary": {
    "critical": 1,
    "high": 2,
    "medium": 4,
    "low": 5,
    "total_findings": 12,
    "health": "warning",
    "open_ports": [49222, 80, 443]
  },
  "findings": [
    {
      "tool": "lynis",
      "severity": "HIGH",
      "title": "PermitRootLogin habilitado",
      "recommendation": "Deshabilitar acceso root directo por SSH"
    },
    {
      "tool": "nmap",
      "severity": "MEDIUM",
      "title": "Puerto expuesto",
      "port": 8080,
      "service": "http-proxy"
    }
  ]
}
```

## Flujo desde panel super

1. Abrir `/super/seguridad.html`.
2. Ajustar host, puertos, perfil y herramientas activas.
3. Guardar configuracion.
4. Ejecutar escaneo.
5. Revisar resumen, hallazgos, historial y comparacion.
6. Exportar el mismo reporte en el formato requerido.

## Notas operativas

- `localhost:8080` es solo para desarrollo local y no requiere DNS.
- Para produccion, el VPS debe aceptar `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com` en Nginx o reverse proxy, aunque el backend ya redirige `www` al dominio raiz como host canonico.
- El modulo no crea tablas nuevas: guarda configuracion y reportes en filesystem del backend.
- Si alguna herramienta no existe en el VPS, el reporte registra ese faltante como hallazgo en lugar de abortar todo el escaneo.
