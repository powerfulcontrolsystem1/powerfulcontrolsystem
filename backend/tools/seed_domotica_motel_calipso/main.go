package main

import (
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func main() {
	var empresaID = flag.Int64("empresa_id", 0, "empresa_id; si queda en 0 busca Motel Calipso")
	var usuario = flag.String("usuario", "codex.domotica.qa", "usuario creador")
	flag.Parse()

	dsn := strings.TrimSpace(os.Getenv("DB_EMPRESAS_DSN"))
	if dsn == "" {
		log.Fatal("DB_EMPRESAS_DSN no esta definido")
	}
	db, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	_ = dbpkg.EnsurePostgresRuntimeCompat(db)
	if err := dbpkg.EnsureEmpresaControlElectricoSchema(db); err != nil {
		log.Fatal(err)
	}
	if *empresaID <= 0 {
		id, err := findEmpresa(db, "motel calipso")
		if err != nil {
			log.Fatal(err)
		}
		*empresaID = id
	}
	empresa, _ := dbpkg.GetEmpresaByScopeID(db, *empresaID)
	nombreEmpresa := fmt.Sprintf("Empresa %d", *empresaID)
	if empresa != nil && strings.TrimSpace(empresa.Nombre) != "" {
		nombreEmpresa = empresa.Nombre
	}

	cfg := dbpkg.EmpresaControlElectricoConfig{
		EmpresaID:          *empresaID,
		Habilitado:         true,
		RaspberryIP:        "http://127.0.0.1:8123",
		RaspberryPort:      8123,
		APIPath:            "/api/services/switch/turn_on",
		TimeoutMS:          1200,
		AutoSyncEstaciones: true,
		FailSafeOnError:    true,
		UsuarioCreador:     *usuario,
		Estado:             "activo",
		Observaciones:      "QA domotica empresarial: Home Assistant, Shelly, consumo, alarmas y agenda.",
	}
	if _, err := dbpkg.UpsertEmpresaControlElectricoConfig(db, &cfg); err != nil {
		log.Fatal(err)
	}
	haID, err := dbpkg.UpsertEmpresaControlElectricoRaspberry(db, &dbpkg.EmpresaControlElectricoRaspberry{
		EmpresaID:       *empresaID,
		Codigo:          "home_assistant_principal",
		Nombre:          "Home Assistant principal",
		TipoControlador: "home_assistant",
		Proveedor:       "home_assistant",
		BaseURL:         "http://127.0.0.1:8123",
		RaspberryIP:     "http://127.0.0.1:8123",
		RaspberryPort:   8123,
		APIPath:         "/api",
		TimeoutMS:       1200,
		UsuarioCreador:  *usuario,
		Estado:          "activo",
		Observaciones:   "Gateway recomendado para Siri/HomeKit Bridge, Matter, Hue, Tuya y Zigbee2MQTT.",
	})
	if err != nil {
		log.Fatal(err)
	}
	shellyID, err := dbpkg.UpsertEmpresaControlElectricoRaspberry(db, &dbpkg.EmpresaControlElectricoRaspberry{
		EmpresaID:       *empresaID,
		Codigo:          "shelly_demo",
		Nombre:          "Shelly Plus demo",
		TipoControlador: "shelly_rpc",
		Proveedor:       "shelly",
		BaseURL:         "http://127.0.0.1:8088",
		RaspberryIP:     "127.0.0.1",
		RaspberryPort:   8088,
		APIPath:         "/rpc/Switch.Set",
		TimeoutMS:       1200,
		UsuarioCreador:  *usuario,
		Estado:          "activo",
		Observaciones:   "Controlador directo Shelly RPC para cargas con medicion.",
	})
	if err != nil {
		log.Fatal(err)
	}

	reles := []dbpkg.EmpresaControlElectricoRele{
		rele(*empresaID, haID, 1, "estacion_1_luz", "luces", "Lampara Estacion 1", "homekit_siri", "Apple/Home Assistant", "HomeKit Bridge", "switch.lampara_estacion_1", 10, true, 60, "18:00", "23:50"),
		rele(*empresaID, shellyID, 2, "estacion_2_bomba", "motobomba", "Motobomba Estacion 2", "shelly_rpc", "Shelly", "Plus 1PM", "", 0, true, 650, "08:00", "08:15"),
		rele(*empresaID, haID, 3, "estacion_3_aire", "aire", "Aire Estacion 3", "home_assistant", "Tuya", "Smart Life AC", "switch.aire_estacion_3", 12, false, 1200, "20:00", "06:00"),
	}
	created := 0
	var aireReleID int64
	for i := range reles {
		id, err := dbpkg.UpsertEmpresaControlElectricoRele(db, &reles[i])
		if err != nil {
			log.Fatal(err)
		}
		created++
		if reles[i].SalidaCodigo == "estacion_3_aire" {
			aireReleID = id
		}
		state := "off"
		if i < 2 {
			state = "on"
		}
		_ = dbpkg.UpdateEmpresaControlElectricoReleRuntime(db, *empresaID, id, state, "qa_seed", "")
		_, _ = dbpkg.InsertEmpresaControlElectricoLectura(db, dbpkg.EmpresaControlElectricoLectura{
			EmpresaID:  *empresaID,
			EstacionID: reles[i].EstacionID,
			ReleID:     id,
			Origen:     "qa_seed",
			Estado:     state,
			ConsumoW:   reles[i].PotenciaW,
			ConsumoKWh: float64(i+1) * 0.42,
			VoltajeV:   120,
			CorrienteA: reles[i].PotenciaW / 120,
		})
		_, _ = dbpkg.InsertEmpresaControlElectricoEvento(db, dbpkg.EmpresaControlElectricoEvento{
			EmpresaID:      *empresaID,
			EstacionID:     reles[i].EstacionID,
			ReleID:         id,
			Comando:        "qa_seed_estado",
			EstadoObjetivo: state,
			Resultado:      "ok",
			Actor:          *usuario,
			Origen:         "qa_domotica",
			MetadataJSON:   `{"nota":"prueba sin controlador fisico real"}`,
		})
		if img, err := ensureDemoImage(*empresaID, nombreEmpresa, id); err == nil {
			_ = dbpkg.UpdateEmpresaControlElectricoReleImagen(db, *empresaID, id, img)
		}
	}
	_, _ = dbpkg.UpsertEmpresaControlElectricoRegla(db, &dbpkg.EmpresaControlElectricoRegla{
		EmpresaID:        *empresaID,
		Nombre:           "Puerta abierta apaga aire",
		SensorCodigo:     "binary_sensor.puerta_estacion_3",
		SensorTipo:       "contacto",
		Condicion:        "igual",
		Valor:            "open",
		Accion:           "apagar",
		EstacionID:       3,
		ReleID:           aireReleID,
		AlarmaHabilitada: true,
		Severidad:        "critica",
		Mensaje:          "Si la puerta queda abierta, generar alarma y apagar carga asociada.",
		UsuarioCreador:   *usuario,
		Estado:           "activo",
	})
	fmt.Printf("OK domotica QA empresa_id=%d empresa=%q controladores=2 aparatos=%d lecturas=3 reglas=1 fecha=%s\n", *empresaID, nombreEmpresa, created, time.Now().Format("2006-01-02 15:04:05"))
}

func rele(empresaID, raspberryID, estacionID int64, salida, tipo, nombre, integracion, fabricante, modelo, entity string, gpio int, schedule bool, watts float64, on, off string) dbpkg.EmpresaControlElectricoRele {
	return dbpkg.EmpresaControlElectricoRele{
		EmpresaID: empresaID, RaspberryID: raspberryID, EstacionID: estacionID,
		EstacionCodigo: fmt.Sprintf("EST-%d", estacionID), EstacionNombre: fmt.Sprintf("Estacion %d", estacionID),
		SalidaCodigo: salida, TipoCarga: tipo, IntegracionTipo: integracion, Fabricante: fabricante, Modelo: modelo,
		EntityID: entity, DeviceID: "0", Capability: "switch", GPIOPin: gpio, RelayName: nombre, ActiveHigh: true,
		Modo: "seguimiento_estacion", ProgramacionHabilitada: schedule, HoraEncendido: on, HoraApagado: off,
		ProgramacionDias: "todos", ProgramacionTimezone: "America/Bogota", MonitoreoHabilitado: true, PotenciaW: watts,
		SensorConsumoEntityID: strings.Replace(entity, "switch.", "sensor.", 1) + "_power",
		UsuarioCreador:        "codex.domotica.qa", Estado: "activo", Observaciones: "QA domotica Motel Calipso",
	}
}

func findEmpresa(db *sql.DB, contains string) (int64, error) {
	empresas, err := dbpkg.GetEmpresas(db)
	if err != nil {
		return 0, err
	}
	contains = strings.ToLower(strings.TrimSpace(contains))
	for _, e := range empresas {
		if strings.Contains(strings.ToLower(e.Nombre), contains) {
			if e.EmpresaID > 0 {
				return e.EmpresaID, nil
			}
			return e.ID, nil
		}
	}
	return 0, fmt.Errorf("no se encontro empresa que contenga %q", contains)
}

func ensureDemoImage(empresaID int64, empresaNombre string, releID int64) (string, error) {
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
	folder := "empresa_" + strconvFormat(empresaID) + "_" + sanitizeFolder(empresaNombre)
	dir := filepath.Join("..", "web", "uploads", "empresas", folder, "imagenes", "domotica")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("rele_%d_demo.png", releID)
	if err := os.WriteFile(filepath.Join(dir, name), png, 0o644); err != nil {
		return "", err
	}
	return "/uploads/empresas/" + folder + "/imagenes/domotica/" + name, nil
}

func sanitizeFolder(raw string) string {
	return sanitizeFolderASCII(raw)
}

func sanitizeFolderASCII(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	var out []rune
	for _, r := range value {
		switch r {
		case 0x00e1, 0x00e0, 0x00e2, 0x00e3, 0x00e4:
			r = 'a'
		case 0x00e9, 0x00e8, 0x00ea, 0x00eb:
			r = 'e'
		case 0x00ed, 0x00ec, 0x00ee, 0x00ef:
			r = 'i'
		case 0x00f3, 0x00f2, 0x00f4, 0x00f5, 0x00f6:
			r = 'o'
		case 0x00fa, 0x00f9, 0x00fb, 0x00fc:
			r = 'u'
		case 0x00f1:
			r = 'n'
		case ' ', '-', '.', '/', '\\':
			r = '_'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			out = append(out, r)
		}
	}
	clean := strings.Trim(string(out), "_")
	if clean == "" {
		clean = "empresa"
	}
	if len(clean) > 60 {
		clean = clean[:60]
	}
	return clean
}

func strconvFormat(v int64) string {
	return fmt.Sprintf("%d", v)
}
