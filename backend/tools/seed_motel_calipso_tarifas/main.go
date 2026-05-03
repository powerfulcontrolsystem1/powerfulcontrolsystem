package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type estacionesConfig struct {
	Cantidad   int `json:"cantidad"`
	Estaciones []struct {
		ID     int64  `json:"id"`
		Nombre string `json:"nombre"`
	} `json:"estaciones"`
}

func parseStationsConfig(raw string) (*estacionesConfig, error) {
	current := any(strings.TrimSpace(raw))
	for i := 0; i < 3; i++ {
		s, ok := current.(string)
		if !ok {
			break
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, nil
		}
		var decoded any
		if err := json.Unmarshal([]byte(s), &decoded); err != nil {
			return nil, err
		}
		current = decoded
	}
	m, ok := current.(map[string]any)
	if !ok || m == nil {
		return nil, nil
	}
	b, _ := json.Marshal(m)
	var out estacionesConfig
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out.Cantidad <= 0 {
		out.Cantidad = len(out.Estaciones)
	}
	if out.Cantidad <= 0 {
		out.Cantidad = 1
	}
	return &out, nil
}

func main() {
	var (
		empresaID         = flag.Int64("empresa_id", 7, "empresa_id (Calipso=7)")
		moneda            = flag.String("moneda", "COP", "moneda")
		minBase           = flag.Int("min_base", 120, "minutos_base")
		valBase           = flag.Float64("valor_base", 35000, "valor_base")
		minExtra          = flag.Int("min_extra", 60, "minutos_extra")
		valExtra          = flag.Float64("valor_extra", 20000, "valor_extra")
		valDia            = flag.Float64("valor_dia", 150000, "valor_dia")
		valMotelExpress   = flag.Float64("valor_motel_express", 50000, "valor base del plan motel express")
		valMotelNocturno  = flag.Float64("valor_motel_nocturno", 90000, "valor base del plan motel nocturno")
		servicio          = flag.String("servicio", "hospedaje", "servicio_nombre")
		checkIn           = flag.String("check_in", "15:00", "hora_check_in")
		checkOut          = flag.String("check_out", "12:00", "hora_check_out")
		usuarioCreador    = flag.String("usuario", "seed", "usuario_creador")
		overwriteExisting = flag.Bool("overwrite", false, "si true, actualiza tarifas existentes; si false solo crea si no existe")
	)
	flag.Parse()

	dsn := strings.TrimSpace(os.Getenv("DB_EMPRESAS_DSN"))
	if dsn == "" {
		log.Fatal("DB_EMPRESAS_DSN no está definido (debe apuntar a pcs_empresas en PostgreSQL)")
	}
	if *empresaID <= 0 {
		log.Fatal("empresa_id inválido")
	}

	driver := dbpkg.PostgresCompatDriverName()
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("sql.Open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db.Ping error: %v", err)
	}

	// Compat + schemas (tarifas y prefs)
	_ = dbpkg.EnsurePostgresRuntimeCompat(db)
	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(db); err != nil {
		log.Fatalf("EnsureEmpresaEstacionPrefsSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaTarifasPorMinutosSchema(db); err != nil {
		log.Fatalf("EnsureEmpresaTarifasPorMinutosSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaTarifasPorDiaSchema(db); err != nil {
		log.Fatalf("EnsureEmpresaTarifasPorDiaSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaTarifasMotelSchema(db); err != nil {
		log.Fatalf("EnsureEmpresaTarifasMotelSchema: %v", err)
	}

	pref, err := dbpkg.GetEmpresaEstacionPref(db, *empresaID, 0, "estaciones_config")
	if err != nil {
		log.Fatalf("GetEmpresaEstacionPref(estaciones_config): %v", err)
	}
	if pref == nil || strings.TrimSpace(pref.Valor) == "" {
		log.Fatalf("No existe estaciones_config para empresa_id=%d. Configura estaciones primero.", *empresaID)
	}

	cfg, err := parseStationsConfig(pref.Valor)
	if err != nil || cfg == nil {
		log.Fatalf("No se pudo parsear estaciones_config: %v", err)
	}

	// Construir mapa de nombres si vienen en el JSON.
	nombres := map[int64]string{}
	for _, st := range cfg.Estaciones {
		if st.ID > 0 {
			nombres[st.ID] = strings.TrimSpace(st.Nombre)
		}
	}

	createdMin := 0
	updatedMin := 0
	createdDia := 0
	updatedDia := 0
	createdMotel := 0
	updatedMotel := 0

	for stationID := int64(1); stationID <= int64(cfg.Cantidad); stationID++ {
		stationName := strings.TrimSpace(nombres[stationID])
		if stationName == "" {
			stationName = "Estación " + strconv.FormatInt(stationID, 10)
		}
		stationCode := fmt.Sprintf("EST-%d-%d", *empresaID, stationID)

		// Tarifas por minutos: una regla para todos los días (1..7).
		existingMin, err := dbpkg.ListEmpresaTarifasPorMinutos(db, *empresaID, dbpkg.EmpresaTarifaPorMinutosFilter{
			EstacionID:      stationID,
			DiaSemana:       0,
			IncludeInactive: true,
			Limit:           50,
		})
		if err != nil {
			log.Fatalf("ListEmpresaTarifasPorMinutos station=%d: %v", stationID, err)
		}
		if len(existingMin) == 0 {
			_, err := dbpkg.CreateEmpresaTarifaPorMinutos(db, dbpkg.EmpresaTarifaPorMinutos{
				EmpresaID:      *empresaID,
				EstacionID:     stationID,
				EstacionCodigo: stationCode,
				EstacionNombre: stationName,
				DiaSemanaDesde: 1,
				DiaSemanaHasta: 7,
				MinutosBase:    *minBase,
				ValorBase:      *valBase,
				MinutosExtra:   *minExtra,
				ValorExtra:     *valExtra,
				Moneda:         strings.ToUpper(strings.TrimSpace(*moneda)),
				Prioridad:      1,
				UsuarioCreador: strings.TrimSpace(*usuarioCreador),
				Estado:         "activo",
				Observaciones:  "seed motel calipso",
			})
			if err != nil {
				log.Fatalf("CreateEmpresaTarifaPorMinutos station=%d: %v", stationID, err)
			}
			createdMin++
		} else if *overwriteExisting {
			// actualizar la primera
			row := existingMin[0]
			row.MinutosBase = *minBase
			row.ValorBase = *valBase
			row.MinutosExtra = *minExtra
			row.ValorExtra = *valExtra
			row.Moneda = strings.ToUpper(strings.TrimSpace(*moneda))
			row.EstacionCodigo = stationCode
			row.EstacionNombre = stationName
			row.DiaSemanaDesde = 1
			row.DiaSemanaHasta = 7
			row.Prioridad = 1
			row.UsuarioCreador = strings.TrimSpace(*usuarioCreador)
			row.Estado = "activo"
			row.Observaciones = "seed motel calipso (overwrite)"
			if err := dbpkg.UpdateEmpresaTarifaPorMinutos(db, row); err != nil {
				log.Fatalf("UpdateEmpresaTarifaPorMinutos station=%d: %v", stationID, err)
			}
			updatedMin++
		}

		// Tarifas por día: una regla por estación.
		existingDia, err := dbpkg.ListEmpresaTarifasPorDia(db, *empresaID, dbpkg.EmpresaTarifaPorDiaFilter{
			EstacionID:      stationID,
			IncludeInactive: true,
			Limit:           10,
		})
		if err != nil {
			log.Fatalf("ListEmpresaTarifasPorDia station=%d: %v", stationID, err)
		}
		if len(existingDia) == 0 {
			_, err := dbpkg.CreateEmpresaTarifaPorDia(db, dbpkg.EmpresaTarifaPorDia{
				EmpresaID:              *empresaID,
				EstacionID:             stationID,
				EstacionCodigo:         stationCode,
				EstacionNombre:         stationName,
				ServicioNombre:         strings.TrimSpace(*servicio),
				ValorDia:               *valDia,
				HoraCheckIn:            strings.TrimSpace(*checkIn),
				HoraCheckOut:           strings.TrimSpace(*checkOut),
				Moneda:                 strings.ToUpper(strings.TrimSpace(*moneda)),
				Prioridad:              1,
				AplicarAutomaticamente: true,
				UsuarioCreador:         strings.TrimSpace(*usuarioCreador),
				Estado:                 "activo",
				Observaciones:          "seed motel calipso",
			})
			if err != nil {
				log.Fatalf("CreateEmpresaTarifaPorDia station=%d: %v", stationID, err)
			}
			createdDia++
		} else if *overwriteExisting {
			row := existingDia[0]
			row.ServicioNombre = strings.TrimSpace(*servicio)
			row.ValorDia = *valDia
			row.HoraCheckIn = strings.TrimSpace(*checkIn)
			row.HoraCheckOut = strings.TrimSpace(*checkOut)
			row.Moneda = strings.ToUpper(strings.TrimSpace(*moneda))
			row.EstacionCodigo = stationCode
			row.EstacionNombre = stationName
			row.Prioridad = 1
			row.AplicarAutomaticamente = true
			row.UsuarioCreador = strings.TrimSpace(*usuarioCreador)
			row.Estado = "activo"
			row.Observaciones = "seed motel calipso (overwrite)"
			if err := dbpkg.UpdateEmpresaTarifaPorDia(db, row); err != nil {
				log.Fatalf("UpdateEmpresaTarifaPorDia station=%d: %v", stationID, err)
			}
			updatedDia++
		}

		motelPlans := []dbpkg.EmpresaTarifaMotel{
			{
				NombrePlan:          "Express 2 horas",
				TipoPlan:            "express",
				CategoriaHabitacion: "estandar",
				HoraInicio:          "00:00",
				HoraFin:             "23:59",
				MinutosIncluidos:    *minBase,
				ValorBase:           *valMotelExpress,
				MinutosExtra:        *minExtra,
				ValorExtra:          *valExtra,
				ToleranciaMinutos:   10,
				Prioridad:           1,
			},
			{
				NombrePlan:          "Nocturno",
				TipoPlan:            "nocturno",
				CategoriaHabitacion: "estandar",
				HoraInicio:          "20:00",
				HoraFin:             "08:00",
				MinutosIncluidos:    720,
				ValorBase:           *valMotelNocturno,
				MinutosExtra:        *minExtra,
				ValorExtra:          *valExtra,
				ToleranciaMinutos:   15,
				Prioridad:           2,
			},
		}
		for _, plan := range motelPlans {
			existingMotel, err := dbpkg.ListEmpresaTarifasMotel(db, *empresaID, dbpkg.EmpresaTarifaMotelFilter{
				EstacionID:      stationID,
				TipoPlan:        plan.TipoPlan,
				IncludeInactive: true,
				Limit:           10,
			})
			if err != nil {
				log.Fatalf("ListEmpresaTarifasMotel station=%d tipo=%s: %v", stationID, plan.TipoPlan, err)
			}
			plan.EmpresaID = *empresaID
			plan.EstacionID = stationID
			plan.EstacionCodigo = stationCode
			plan.EstacionNombre = stationName
			plan.DiaSemanaDesde = 1
			plan.DiaSemanaHasta = 7
			plan.Moneda = strings.ToUpper(strings.TrimSpace(*moneda))
			plan.CobrarPorFraccion = true
			plan.AplicarAutomatico = true
			plan.UsuarioCreador = strings.TrimSpace(*usuarioCreador)
			plan.Estado = "activo"
			plan.Observaciones = "seed motel calipso"

			if len(existingMotel) == 0 {
				if _, err := dbpkg.CreateEmpresaTarifaMotel(db, plan); err != nil {
					log.Fatalf("CreateEmpresaTarifaMotel station=%d tipo=%s: %v", stationID, plan.TipoPlan, err)
				}
				createdMotel++
				continue
			}
			if *overwriteExisting {
				plan.ID = existingMotel[0].ID
				plan.Observaciones = "seed motel calipso (overwrite)"
				if err := dbpkg.UpdateEmpresaTarifaMotel(db, plan); err != nil {
					log.Fatalf("UpdateEmpresaTarifaMotel station=%d tipo=%s: %v", stationID, plan.TipoPlan, err)
				}
				updatedMotel++
			}
		}
	}

	log.Printf("OK seed empresa_id=%d estaciones=%d | minutos: creadas=%d actualizadas=%d | dia: creadas=%d actualizadas=%d | motel: creadas=%d actualizadas=%d",
		*empresaID, cfg.Cantidad, createdMin, updatedMin, createdDia, updatedDia, createdMotel, updatedMotel,
	)
}
