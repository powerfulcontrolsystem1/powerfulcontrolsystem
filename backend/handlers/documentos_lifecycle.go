package handlers

import (
	"fmt"
	"sort"
	"strings"
)

type documentoTransition struct {
	Accion         string
	EstadoAnterior string
	EstadoNuevo    string
	Evento         string
}

type documentoTransitionRule struct {
	AccionCanonica string
	Evento         string
	EstadoNuevo    string
	EstadoDefault  string
	EstadosPrevios map[string]struct{}
}

func resolveFacturacionTransition(actionRaw, estadoActualRaw string) (documentoTransition, error) {
	rules := map[string]documentoTransitionRule{
		"emitir": {
			AccionCanonica: "emitir",
			Evento:         "factura_emitida",
			EstadoNuevo:    "emitida",
			EstadoDefault:  "borrador",
			EstadosPrevios: toAllowedSet("borrador", "pendiente_emision"),
		},
		"anular": {
			AccionCanonica: "anular",
			Evento:         "factura_anulada",
			EstadoNuevo:    "anulada",
			EstadoDefault:  "emitida",
			EstadosPrevios: toAllowedSet("emitida"),
		},
		"nota_credito": {
			AccionCanonica: "nota_credito",
			Evento:         "nota_credito_emitida",
			EstadoNuevo:    "ajustada",
			EstadoDefault:  "emitida",
			EstadosPrevios: toAllowedSet("emitida"),
		},
		"emitir_nota_credito": {
			AccionCanonica: "nota_credito",
			Evento:         "nota_credito_emitida",
			EstadoNuevo:    "ajustada",
			EstadoDefault:  "emitida",
			EstadosPrevios: toAllowedSet("emitida"),
		},
	}
	return resolveDocumentoTransition("facturacion", actionRaw, estadoActualRaw, rules)
}

func resolveComprasTransition(actionRaw, estadoActualRaw string) (documentoTransition, error) {
	rules := map[string]documentoTransitionRule{
		"emitir": {
			AccionCanonica: "emitir_orden",
			Evento:         "orden_compra_emitida",
			EstadoNuevo:    "emitida",
			EstadoDefault:  "borrador",
			EstadosPrevios: toAllowedSet("borrador", "pendiente_emision"),
		},
		"emitir_orden": {
			AccionCanonica: "emitir_orden",
			Evento:         "orden_compra_emitida",
			EstadoNuevo:    "emitida",
			EstadoDefault:  "borrador",
			EstadosPrevios: toAllowedSet("borrador", "pendiente_emision"),
		},
		"solicitar_aprobacion": {
			AccionCanonica: "solicitar_aprobacion",
			Evento:         "orden_compra_pendiente_aprobacion",
			EstadoNuevo:    "pendiente_aprobacion",
			EstadoDefault:  "borrador",
			EstadosPrevios: toAllowedSet("borrador", "pendiente_emision", "rechazada"),
		},
		"rechazar_compra": {
			AccionCanonica: "rechazar_compra",
			Evento:         "orden_compra_rechazada",
			EstadoNuevo:    "rechazada",
			EstadoDefault:  "pendiente_aprobacion",
			EstadosPrevios: toAllowedSet("pendiente_aprobacion"),
		},
		"recepcionar_parcial_compra": {
			AccionCanonica: "recepcionar_parcial_compra",
			Evento:         "compra_recepcion_parcial",
			EstadoNuevo:    "recepcion_parcial",
			EstadoDefault:  "emitida",
			EstadosPrevios: toAllowedSet("emitida", "recepcion_parcial"),
		},
		"recepcionar": {
			AccionCanonica: "recepcionar_compra",
			Evento:         "compra_recepcionada",
			EstadoNuevo:    "recepcionada",
			EstadoDefault:  "emitida",
			EstadosPrevios: toAllowedSet("emitida", "recepcion_parcial"),
		},
		"recepcionar_compra": {
			AccionCanonica: "recepcionar_compra",
			Evento:         "compra_recepcionada",
			EstadoNuevo:    "recepcionada",
			EstadoDefault:  "emitida",
			EstadosPrevios: toAllowedSet("emitida", "recepcion_parcial"),
		},
		"contabilizar": {
			AccionCanonica: "contabilizar_compra",
			Evento:         "compra_contabilizada",
			EstadoNuevo:    "contabilizada",
			EstadoDefault:  "recepcionada",
			EstadosPrevios: toAllowedSet("recepcionada"),
		},
		"contabilizar_compra": {
			AccionCanonica: "contabilizar_compra",
			Evento:         "compra_contabilizada",
			EstadoNuevo:    "contabilizada",
			EstadoDefault:  "recepcionada",
			EstadosPrevios: toAllowedSet("recepcionada"),
		},
	}
	return resolveDocumentoTransition("compras", actionRaw, estadoActualRaw, rules)
}

func resolveDocumentoTransition(scope, actionRaw, estadoActualRaw string, rules map[string]documentoTransitionRule) (documentoTransition, error) {
	action := normalizeDocumentoState(actionRaw)
	rule, ok := rules[action]
	if !ok {
		return documentoTransition{}, fmt.Errorf("accion no soportada para %s: %s", scope, strings.TrimSpace(actionRaw))
	}

	estadoAnterior := normalizeDocumentoState(estadoActualRaw)
	if estadoAnterior == "" {
		estadoAnterior = normalizeDocumentoState(rule.EstadoDefault)
	}

	if _, allowed := rule.EstadosPrevios[estadoAnterior]; !allowed {
		return documentoTransition{}, fmt.Errorf("transicion invalida para %s: accion=%s requiere estado_actual en [%s], recibido=%s", scope, rule.AccionCanonica, formatAllowedStates(rule.EstadosPrevios), estadoAnterior)
	}

	return documentoTransition{
		Accion:         rule.AccionCanonica,
		EstadoAnterior: estadoAnterior,
		EstadoNuevo:    normalizeDocumentoState(rule.EstadoNuevo),
		Evento:         rule.Evento,
	}, nil
}

func normalizeDocumentoState(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	return v
}

func toAllowedSet(states ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(states))
	for _, state := range states {
		normalized := normalizeDocumentoState(state)
		if normalized == "" {
			continue
		}
		set[normalized] = struct{}{}
	}
	return set
}

func formatAllowedStates(set map[string]struct{}) string {
	if len(set) == 0 {
		return ""
	}
	states := make([]string, 0, len(set))
	for state := range set {
		states = append(states, state)
	}
	sort.Strings(states)
	return strings.Join(states, ",")
}
