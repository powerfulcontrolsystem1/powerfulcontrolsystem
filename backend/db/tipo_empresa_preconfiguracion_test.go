package db

import "testing"

func TestDefaultTipoEmpresaPreconfigIncludesPorteroRole(t *testing.T) {
	template := TipoEmpresaPreconfigTemplate{}
	roles := rolesFromTipoEmpresaPreconfigTemplate(template)
	if !stringSliceContainsForTest(roles, "portero") {
		t.Fatalf("roles por defecto deben incluir portero, got %v", roles)
	}
	if !stringSliceContainsForTest(roles, "contador") {
		t.Fatalf("roles por defecto deben incluir contador, got %v", roles)
	}
	if !stringSliceContainsForTest(roles, "empresario") {
		t.Fatalf("roles por defecto deben incluir empresario, got %v", roles)
	}
	if !stringSliceContainsForTest(roles, "servicio_limpieza") {
		t.Fatalf("roles por defecto deben incluir servicio_limpieza, got %v", roles)
	}
	for _, expected := range []string{"supervisor_sucursal", "vendedor", "recepcion", "jefe_bodega", "responsable_bodega", "recursos_humanos", "tecnico_solar"} {
		if !stringSliceContainsForTest(roles, expected) {
			t.Fatalf("roles por defecto deben incluir %s, got %v", expected, roles)
		}
	}

	permisos := permisosModuloPreconfigRol(77, "portero")
	if !rolPermisoModuloAllowedForTest(permisos, "ventas", "R") || !rolPermisoModuloAllowedForTest(permisos, "ventas", "A") {
		t.Fatalf("portero debe tener permisos R y A en ventas, got %+v", permisos)
	}
	for _, accion := range []string{"C", "U", "D"} {
		if rolPermisoModuloAllowedForTest(permisos, "ventas", accion) {
			t.Fatalf("portero no debe tener permiso %s en ventas", accion)
		}
	}

	permisosContador := permisosModuloPreconfigRol(88, "contador")
	if !rolPermisoModuloAllowedForTest(permisosContador, "finanzas", "R") || !rolPermisoModuloAllowedForTest(permisosContador, "facturacion", "R") {
		t.Fatalf("contador debe tener solo lectura en finanzas y facturacion, got %+v", permisosContador)
	}
	for _, item := range permisosContador {
		if item.Accion != "R" {
			t.Fatalf("contador no debe tener accion %s en modulo %s", item.Accion, item.Modulo)
		}
	}

	permisosEmpresario := permisosModuloPreconfigRol(99, "empresario")
	if len(permisosEmpresario) != 1 || !rolPermisoModuloAllowedForTest(permisosEmpresario, "reportes", "R") {
		t.Fatalf("empresario debe tener solo lectura de reportes, got %+v", permisosEmpresario)
	}

	permisosLimpieza := permisosModuloPreconfigRol(100, "servicio_limpieza")
	if len(permisosLimpieza) != 1 || !rolPermisoModuloAllowedForTest(permisosLimpieza, "ventas", "R") {
		t.Fatalf("servicio_limpieza debe tener solo lectura de ventas/estaciones, got %+v", permisosLimpieza)
	}

	permisosSolar := permisosModuloPreconfigRol(101, "tecnico_solar")
	if len(permisosSolar) != 1 || !rolPermisoModuloAllowedForTest(permisosSolar, "energia_solar", "R") {
		t.Fatalf("tecnico_solar debe tener solo lectura de energia_solar, got %+v", permisosSolar)
	}

	permisosBodega := permisosModuloPreconfigRol(102, "jefe_bodega")
	for _, accion := range []string{"R", "C", "U", "A"} {
		if !rolPermisoModuloAllowedForTest(permisosBodega, "inventario", accion) {
			t.Fatalf("jefe_bodega debe tener inventario:%s, got %+v", accion, permisosBodega)
		}
	}
	if rolPermisoModuloAllowedForTest(permisosBodega, "inventario", "D") {
		t.Fatalf("jefe_bodega no debe tener inventario:D, got %+v", permisosBodega)
	}

	permisosResponsableBodega := permisosModuloPreconfigRol(103, "responsable_bodega")
	for _, accion := range []string{"R", "C", "U", "A"} {
		if !rolPermisoModuloAllowedForTest(permisosResponsableBodega, "inventario", accion) {
			t.Fatalf("responsable_bodega debe tener inventario:%s, got %+v", accion, permisosResponsableBodega)
		}
	}
	if rolPermisoModuloAllowedForTest(permisosResponsableBodega, "inventario", "D") || !rolPermisoModuloAllowedForTest(permisosResponsableBodega, "compras", "R") {
		t.Fatalf("responsable_bodega debe administrar inventario sin eliminar y consultar compras, got %+v", permisosResponsableBodega)
	}
}

func stringSliceContainsForTest(items []string, expected string) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}

func rolPermisoModuloAllowedForTest(items []RolPermisoModulo, modulo, accion string) bool {
	for _, item := range items {
		if item.Modulo == modulo && item.Accion == accion && item.Permitido {
			return true
		}
	}
	return false
}

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
			if template.Modulos.EnergiaSolar == nil {
				t.Fatalf("preconfiguracion debe incluir energia solar como modulo opcional")
			}
			if template.Modulos.EnergiaSolar.Habilitado {
				t.Fatalf("energia solar debe quedar desactivada por defecto")
			}
			if len(template.Modulos.EnergiaSolar.Proveedores) < 4 || len(template.Modulos.EnergiaSolar.Baterias) < 5 {
				t.Fatalf("catalogo solar insuficiente: %+v", template.Modulos.EnergiaSolar)
			}
			if !template.AdaptacionNucleo.FuenteUnica ||
				!template.AdaptacionNucleo.UsuariosDesdeNucleo ||
				!template.AdaptacionNucleo.ProductosServiciosDesdeNucleo {
				t.Fatalf("adaptacion al nucleo incompleta: %+v", template.AdaptacionNucleo)
			}
			if template.Operacion.UsaEstaciones && !template.AdaptacionNucleo.EstacionesComoRecursosConfigurados {
				t.Fatalf("estaciones deben quedar como recursos configurables: %+v", template.AdaptacionNucleo)
			}
			if len(template.AdaptacionNucleo.UsuariosOperativos) < 3 {
				t.Fatalf("adaptacion debe declarar usuarios operativos del nucleo: %+v", template.AdaptacionNucleo)
			}
			if len(template.AdaptacionNucleo.ProductosServiciosGuia) < 3 {
				t.Fatalf("adaptacion debe declarar productos/servicios del nucleo: %+v", template.AdaptacionNucleo)
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
	for _, item := range NuevasPlantillasTipoEmpresaCatalog() {
		t.Run(item.Modulo, func(t *testing.T) {
			preconfig := DefaultTipoEmpresaPreconfiguracion(456, item.Nombre)
			if !preconfig.Enabled {
				t.Fatalf("preconfiguracion plantilla no quedo habilitada")
			}
			template, err := ParseTipoEmpresaPreconfigTemplate(preconfig.ConfigJSON)
			if err != nil {
				t.Fatalf("config plantilla invalida: %v", err)
			}
			if template.Operacion.TipoNegocio != item.Modulo {
				t.Fatalf("tipo_negocio plantilla incorrecto: got %q want %q", template.Operacion.TipoNegocio, item.Modulo)
			}
			if template.Estaciones.Cantidad < 3 {
				t.Fatalf("plantilla debe tener estaciones guia: got %d", template.Estaciones.Cantidad)
			}
			if len(template.Productos) < 3 {
				t.Fatalf("plantilla debe tener productos/servicios guia: got %d", len(template.Productos))
			}
			if len(template.Usuarios) < 3 {
				t.Fatalf("plantilla debe tener usuarios guia: got %d", len(template.Usuarios))
			}
			if len(template.TareasGuia) < 4 {
				t.Fatalf("plantilla debe tener tareas guia profesionales: got %d", len(template.TareasGuia))
			}
			if !template.AdaptacionNucleo.FuenteUnica ||
				!template.AdaptacionNucleo.UsuariosDesdeNucleo ||
				!template.AdaptacionNucleo.ProductosServiciosDesdeNucleo ||
				!template.AdaptacionNucleo.EstacionesComoRecursosConfigurados {
				t.Fatalf("plantilla debe adaptar usuarios/productos/estaciones al nucleo: %+v", template.AdaptacionNucleo)
			}
			if template.AdaptacionNucleo.EntidadEstacionSingular != item.StationPrefix {
				t.Fatalf("entidad estacion=%q want %q", template.AdaptacionNucleo.EntidadEstacionSingular, item.StationPrefix)
			}
			if len(template.AdaptacionNucleo.UsuariosOperativos) < len(item.Roles) {
				t.Fatalf("roles operativos no cubiertos por adaptacion: %+v", template.AdaptacionNucleo)
			}
			if template.IntegracionVertical == nil {
				t.Fatalf("plantilla debe quedar conectado a la matriz extendida de integracion")
			}
			if template.IntegracionVertical.Modulo != item.Modulo {
				t.Fatalf("integracion modulo=%q want %q", template.IntegracionVertical.Modulo, item.Modulo)
			}
			if template.IntegracionVertical.EstadoIntegracion != "plantilla_integrada_nucleo" {
				t.Fatalf("estado integracion=%q", template.IntegracionVertical.EstadoIntegracion)
			}
			if len(template.IntegracionVertical.TemplateActivates) == 0 ||
				len(template.IntegracionVertical.TablesTouched) == 0 ||
				len(template.IntegracionVertical.RequiredPermissions) == 0 ||
				len(template.IntegracionVertical.SaleFlow) == 0 ||
				len(template.IntegracionVertical.ReportsProduced) == 0 ||
				len(template.IntegracionVertical.FinancialCoreModules) == 0 ||
				len(template.IntegracionVertical.IncomeFlow) == 0 ||
				len(template.IntegracionVertical.ExpenseFlow) == 0 ||
				len(template.IntegracionVertical.FinancialTables) == 0 ||
				len(template.IntegracionVertical.FinancialReports) == 0 {
				t.Fatalf("integracion extendida incompleta: %+v", template.IntegracionVertical)
			}
		})
	}
}

func TestNuevasPlantillasProduccionMasivaSeleccionaVeinte(t *testing.T) {
	selected := NuevasPlantillasProduccionMasivaSeleccionados()
	if len(selected) != 20 {
		t.Fatalf("plantillas produccion masiva len=%d, want 20: %v", len(selected), selected)
	}

	seen := map[string]bool{}
	for i, modulo := range selected {
		if modulo == "" {
			t.Fatalf("modulo vacio en prioridad %d", i+1)
		}
		if seen[modulo] {
			t.Fatalf("modulo duplicado en seleccion: %s", modulo)
		}
		seen[modulo] = true
		if rank := NuevoVerticalProduccionMasivaRank(modulo); rank != i+1 {
			t.Fatalf("rank %s=%d want %d", modulo, rank, i+1)
		}
		plantilla := GetEmpresaModuloColombiaPlantilla(modulo)
		preconfig := DefaultTipoEmpresaPreconfiguracion(789, plantilla.Titulo)
		template, err := ParseTipoEmpresaPreconfigTemplate(preconfig.ConfigJSON)
		if err != nil {
			t.Fatalf("preconfig %s invalida: %v", modulo, err)
		}
		if template.IntegracionVertical == nil || !template.IntegracionVertical.ProduccionMasiva {
			t.Fatalf("%s debe quedar marcado como produccion masiva: %+v", modulo, template.IntegracionVertical)
		}
		if template.IntegracionVertical.Decision != "integrar_v1_produccion_masiva" {
			t.Fatalf("%s decision=%q", modulo, template.IntegracionVertical.Decision)
		}
	}

	masivos := 0
	for _, item := range NuevasPlantillasTipoEmpresaCatalog() {
		integracion := BuildTipoEmpresaPreconfigIntegracionVertical(item.Modulo)
		if integracion == nil {
			t.Fatalf("sin integracion para %s", item.Modulo)
		}
		if !integracion.ProduccionMasiva {
			t.Fatalf("%s debe estar promovido a produccion masiva", item.Modulo)
		}
		if integracion.Decision != "integrar_v1_produccion_masiva" {
			t.Fatalf("%s decision=%q", item.Modulo, integracion.Decision)
		}
		if len(integracion.IncomeFlow) == 0 || len(integracion.ExpenseFlow) == 0 || len(integracion.FinancialTables) == 0 {
			t.Fatalf("%s debe declarar ingresos/egresos del nucleo financiero: %+v", item.Modulo, integracion)
		}
		masivos++
	}
	if masivos != 20 {
		t.Fatalf("plantillas masivos=%d, want 20", masivos)
	}
}

func TestIntegracionVerticalClasicaNoSeMarcaComoDiferidaV1(t *testing.T) {
	for _, modulo := range []string{"gimnasio", "odontologia", "drogueria_farmacia", "alquileres", "constructora"} {
		t.Run(modulo, func(t *testing.T) {
			integracion := BuildTipoEmpresaPreconfigIntegracionVertical(modulo)
			if integracion == nil {
				t.Fatalf("sin integracion clasica para %s", modulo)
			}
			if integracion.Decision != "mantener_como_plantilla" {
				t.Fatalf("%s decision=%q, want mantener_como_plantilla", modulo, integracion.Decision)
			}
			if integracion.ProduccionMasiva {
				t.Fatalf("%s no debe entrar en ranking de nuevas plantillas masivos", modulo)
			}
			if len(integracion.IncomeFlow) == 0 || len(integracion.ExpenseFlow) == 0 || len(integracion.FinancialTables) == 0 {
				t.Fatalf("%s debe declarar ingresos/egresos del nucleo financiero: %+v", modulo, integracion)
			}
		})
	}
}
