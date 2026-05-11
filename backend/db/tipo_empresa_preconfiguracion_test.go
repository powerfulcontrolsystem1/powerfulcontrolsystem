package db

import "testing"

func TestDefaultTipoEmpresaPreconfigTemplatesCoverKnownBusinessTypes(t *testing.T) {
	tipos := []struct {
		nombre             string
		prefijo            string
		estaciones         int
		ventaDirecta       bool
		comisiones         bool
		nombreSingular     string
		permiteSinEstacion bool
	}{
		{nombre: "Restaurante", prefijo: "Mesa", estaciones: 8, ventaDirecta: true, nombreSingular: "Mesa"},
		{nombre: "Motel", prefijo: "Habitacion", estaciones: 10, nombreSingular: "Habitacion"},
		{nombre: "Hotel", prefijo: "Habitacion", estaciones: 12, nombreSingular: "Habitacion"},
		{nombre: "Bar", prefijo: "Mesa", estaciones: 10, nombreSingular: "Mesa"},
		{nombre: "Salon de belleza", prefijo: "Silla", estaciones: 6, ventaDirecta: true, comisiones: true, nombreSingular: "Silla"},
		{nombre: "Lavadero de autos", prefijo: "Bahia", estaciones: 6, ventaDirecta: true, comisiones: true, nombreSingular: "Bahia"},
		{nombre: "Pymes", prefijo: "Punto de venta", estaciones: 2, ventaDirecta: true, nombreSingular: "Punto de venta"},
		{nombre: "Tienda punto de venta", prefijo: "Punto de venta", estaciones: 1, ventaDirecta: true, nombreSingular: "Punto de venta"},
		{nombre: "Taller mecanico", prefijo: "Bahia", estaciones: 5, ventaDirecta: true, comisiones: true, nombreSingular: "Bahia"},
		{nombre: "Alquiler de herramientas y motos", prefijo: "Mostrador", estaciones: 2, ventaDirecta: true, nombreSingular: "Mostrador"},
		{nombre: "Constructora", prefijo: "Obra", estaciones: 6, ventaDirecta: true, comisiones: true, nombreSingular: "Obra"},
		{nombre: "Drogueria y farmacia", prefijo: "Caja", estaciones: 2, ventaDirecta: true, nombreSingular: "Caja"},
		{nombre: "Gimnasio", prefijo: "Zona", estaciones: 4, ventaDirecta: true, comisiones: true, nombreSingular: "Zona"},
		{nombre: "Odontologia", prefijo: "Consultorio", estaciones: 3, ventaDirecta: true, comisiones: true, nombreSingular: "Consultorio"},
		{nombre: "Manejo de turnos", prefijo: "Puesto", estaciones: 4, ventaDirecta: true, nombreSingular: "Puesto"},
		{nombre: "Vehiculos y flotas", prefijo: "Bahia", estaciones: 4, ventaDirecta: true, comisiones: true, nombreSingular: "Bahia"},
		{nombre: "Profesional independiente", prefijo: "Venta directa", estaciones: 0, ventaDirecta: true, nombreSingular: "Venta directa", permiteSinEstacion: true},
		{nombre: "Agencia de redes sociales", prefijo: "Cliente", estaciones: 4, ventaDirecta: true, nombreSingular: "Cliente"},
		{nombre: "Sensores y monitoreo", prefijo: "Acceso", estaciones: 4, ventaDirecta: true, nombreSingular: "Acceso"},
		{nombre: "Tipo personalizado", prefijo: "Estacion", estaciones: 4, ventaDirecta: true, nombreSingular: "Estacion"},
	}

	for _, tc := range tipos {
		t.Run(tc.nombre, func(t *testing.T) {
			preconfig := DefaultTipoEmpresaPreconfiguracion(123, tc.nombre)
			if !preconfig.Enabled {
				t.Fatalf("preconfiguracion default no quedo habilitada")
			}
			template, err := ParseTipoEmpresaPreconfigTemplate(preconfig.ConfigJSON)
			if err != nil {
				t.Fatalf("config json invalido: %v", err)
			}
			if template.Estaciones.Cantidad != tc.estaciones {
				t.Fatalf("estaciones guia incorrectas: got %d want %d", template.Estaciones.Cantidad, tc.estaciones)
			}
			if !tc.permiteSinEstacion && !template.Operacion.UsaEstaciones {
				t.Fatalf("debe usar estaciones")
			}
			if template.Estaciones.Prefijo != tc.prefijo {
				t.Fatalf("prefijo incorrecto: got %q want %q", template.Estaciones.Prefijo, tc.prefijo)
			}
			if template.Operacion.NombreEstacionSingular != tc.nombreSingular {
				t.Fatalf("nombre singular incorrecto: got %q want %q", template.Operacion.NombreEstacionSingular, tc.nombreSingular)
			}
			if template.Operacion.VentaDirectaEnabled != tc.ventaDirecta {
				t.Fatalf("venta directa incorrecta: got %v want %v", template.Operacion.VentaDirectaEnabled, tc.ventaDirecta)
			}
			if template.Operacion.ComisionesEnabled != tc.comisiones {
				t.Fatalf("comisiones incorrectas: got %v want %v", template.Operacion.ComisionesEnabled, tc.comisiones)
			}
			if len(template.Productos) < 3 {
				t.Fatalf("productos guia insuficientes: got %d want >= 3", len(template.Productos))
			}
			if len(template.Usuarios) < 3 {
				t.Fatalf("usuarios guia insuficientes: got %d want >= 3", len(template.Usuarios))
			}
			if len(rolesFromTipoEmpresaPreconfigTemplate(template)) < 3 {
				t.Fatalf("roles guia insuficientes: got %d want >= 3", len(rolesFromTipoEmpresaPreconfigTemplate(template)))
			}
			if !template.Asistente.Enabled {
				t.Fatalf("asistente IA guia deshabilitado")
			}
			if len(template.TareasGuia) == 0 {
				t.Fatalf("sin tareas guia")
			}
			switch tc.nombre {
			case "Motel":
				if len(template.Tarifas.Motel) == 0 {
					t.Fatalf("motel debe incluir tarifas guia")
				}
				if template.Modulos.ControlElectrico == nil || len(template.Modulos.ControlElectrico.Reles) == 0 {
					t.Fatalf("motel debe incluir aparatos guia de control electrico")
				}
			case "Hotel":
				if len(template.Tarifas.PorDia) == 0 {
					t.Fatalf("hotel debe incluir tarifas por dia guia")
				}
				if template.Modulos.ControlElectrico == nil || len(template.Modulos.ControlElectrico.Reles) == 0 {
					t.Fatalf("hotel debe incluir aparatos guia de control electrico")
				}
			case "Gimnasio":
				if template.Modulos.Gimnasio == nil || len(template.Modulos.Gimnasio.Planes) == 0 || len(template.Modulos.Gimnasio.Socios) == 0 {
					t.Fatalf("gimnasio debe incluir planes y socios guia")
				}
			case "Odontologia":
				if template.Modulos.Odontologia == nil || len(template.Modulos.Odontologia.Pacientes) == 0 || len(template.Modulos.Odontologia.Tratamientos) == 0 {
					t.Fatalf("odontologia debe incluir pacientes y tratamientos guia")
				}
			case "Manejo de turnos":
				if template.Modulos.TurnosAtencion == nil || len(template.Modulos.TurnosAtencion.Servicios) == 0 || len(template.Modulos.TurnosAtencion.Puestos) == 0 {
					t.Fatalf("turnos debe incluir servicios y puestos guia")
				}
			case "Vehiculos y flotas", "Taller mecanico", "Lavadero de autos":
				if template.Modulos.Vehiculos == nil || len(template.Modulos.Vehiculos.Registros) == 0 || len(template.Modulos.HojaVida) == 0 {
					t.Fatalf("%s debe incluir vehiculos y hoja de vida guia", tc.nombre)
				}
			case "Constructora":
				if template.Operacion.TipoNegocio != "constructora" {
					t.Fatalf("constructora debe quedar como tipo_negocio constructora, got %q", template.Operacion.TipoNegocio)
				}
				if template.Modulos.Vehiculos == nil || len(template.Modulos.Vehiculos.Registros) == 0 || len(template.Modulos.HojaVida) == 0 {
					t.Fatalf("constructora debe incluir maquinaria/flota y hoja de vida guia")
				}
				if !isTipoEmpresaConstructora("Construcción civil") {
					t.Fatalf("constructora debe reconocer construccion con tilde")
				}
			}
			if tc.nombre == "Drogueria y farmacia" {
				if template.Operacion.TipoNegocio != "drogueria_farmacia" {
					t.Fatalf("drogueria debe quedar como tipo_negocio drogueria_farmacia, got %q", template.Operacion.TipoNegocio)
				}
				if !isTipoEmpresaDrogueriaFarmacia("Farmacia especializada") {
					t.Fatalf("drogueria debe reconocer farmacia")
				}
				if len(template.TareasGuia) < 8 {
					t.Fatalf("drogueria debe incluir guia sanitaria amplia, got %d", len(template.TareasGuia))
				}
			}
			if tc.nombre == "Alquiler de herramientas y motos" {
				if template.Operacion.TipoNegocio != "alquileres" {
					t.Fatalf("alquileres debe quedar como tipo_negocio alquileres, got %q", template.Operacion.TipoNegocio)
				}
				if !isTipoEmpresaAlquilerObjetos("Renta de motos y herramientas") {
					t.Fatalf("alquileres debe reconocer renta de motos y herramientas")
				}
				if len(template.TareasGuia) < 8 {
					t.Fatalf("alquileres debe incluir guia profesional amplia, got %d", len(template.TareasGuia))
				}
			}
			raw, err := MarshalTipoEmpresaPreconfigTemplate(template)
			if err != nil {
				t.Fatalf("no serializa template normalizado: %v", err)
			}
			roundtrip, err := ParseTipoEmpresaPreconfigTemplate(raw)
			if err != nil {
				t.Fatalf("roundtrip invalido: %v", err)
			}
			if roundtrip.Estaciones.Cantidad != template.Estaciones.Cantidad {
				t.Fatalf("roundtrip cambio estaciones: got %d want %d", roundtrip.Estaciones.Cantidad, template.Estaciones.Cantidad)
			}
		})
	}
}

func TestDefaultTipoEmpresaPreconfigTemplatesCoverNewVerticalCatalog(t *testing.T) {
	for _, item := range NuevosVerticalesTipoEmpresaCatalog() {
		t.Run(item.Modulo, func(t *testing.T) {
			preconfig := DefaultTipoEmpresaPreconfiguracion(456, item.Nombre)
			if !preconfig.Enabled {
				t.Fatalf("preconfiguracion vertical no quedo habilitada")
			}
			template, err := ParseTipoEmpresaPreconfigTemplate(preconfig.ConfigJSON)
			if err != nil {
				t.Fatalf("config vertical invalida: %v", err)
			}
			if template.Operacion.TipoNegocio != item.Modulo {
				t.Fatalf("tipo_negocio vertical incorrecto: got %q want %q", template.Operacion.TipoNegocio, item.Modulo)
			}
			if template.Estaciones.Cantidad < 3 {
				t.Fatalf("vertical debe tener estaciones guia: got %d", template.Estaciones.Cantidad)
			}
			if len(template.Productos) < 3 {
				t.Fatalf("vertical debe tener productos/servicios guia: got %d", len(template.Productos))
			}
			if len(template.Usuarios) < 3 {
				t.Fatalf("vertical debe tener usuarios guia: got %d", len(template.Usuarios))
			}
			if len(template.TareasGuia) < 4 {
				t.Fatalf("vertical debe tener tareas guia profesionales: got %d", len(template.TareasGuia))
			}
		})
	}
}
