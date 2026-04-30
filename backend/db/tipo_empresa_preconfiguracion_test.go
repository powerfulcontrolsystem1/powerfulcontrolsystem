package db

import "testing"

func TestDefaultTipoEmpresaPreconfigTemplatesCoverKnownBusinessTypes(t *testing.T) {
	tipos := []string{
		"Restaurante",
		"Motel",
		"Hotel",
		"Bar",
		"Salon de belleza",
		"Lavadero de autos",
		"Tienda punto de venta",
		"Taller mecanico",
		"Profesional independiente",
		"Agencia de redes sociales",
		"Sensores y monitoreo",
		"Tipo personalizado",
	}

	for _, nombre := range tipos {
		t.Run(nombre, func(t *testing.T) {
			preconfig := DefaultTipoEmpresaPreconfiguracion(123, nombre)
			if !preconfig.Enabled {
				t.Fatalf("preconfiguracion default no quedo habilitada")
			}
			template, err := ParseTipoEmpresaPreconfigTemplate(preconfig.ConfigJSON)
			if err != nil {
				t.Fatalf("config json invalido: %v", err)
			}
			if template.Estaciones.Cantidad <= 0 {
				t.Fatalf("sin estaciones guia")
			}
			if template.Estaciones.Prefijo == "" {
				t.Fatalf("sin prefijo de estaciones")
			}
			if len(template.Productos) == 0 {
				t.Fatalf("sin productos guia")
			}
			if len(template.Usuarios) == 0 {
				t.Fatalf("sin usuarios guia")
			}
			if !template.Asistente.Enabled {
				t.Fatalf("asistente IA guia deshabilitado")
			}
			if len(template.TareasGuia) == 0 {
				t.Fatalf("sin tareas guia")
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
