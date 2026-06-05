package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type camarasWritePayload struct {
	Camara dbpkg.EmpresaCamara `json:"camara"`
}

type camaraTecnologiaCatalogo struct {
	Clave       string   `json:"clave"`
	Nombre      string   `json:"nombre"`
	Uso         string   `json:"uso"`
	Visores     []string `json:"visores"`
	Observacion string   `json:"observacion"`
}

func EmpresaCamarasHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := dbpkg.EnsureEmpresaCamarasSchema(dbEmp); err != nil {
			log.Printf("[camaras] ensure schema empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo preparar el modulo de camaras", http.StatusInternalServerError)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			handleEmpresaCamarasGET(w, r, dbEmp, empresaID, action)
		case http.MethodPost, http.MethodPut:
			handleEmpresaCamarasWrite(w, r, dbEmp, empresaID, action)
		case http.MethodDelete:
			handleEmpresaCamarasDelete(w, r, dbEmp, empresaID, action)
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleEmpresaCamarasGET(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	switch action {
	case "", "dashboard":
		includeInactive := strings.TrimSpace(r.URL.Query().Get("include_inactive")) == "1"
		camaras, err := dbpkg.ListEmpresaCamaras(dbEmp, empresaID, includeInactive)
		if err != nil {
			log.Printf("[camaras] dashboard empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudieron cargar las camaras", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":          true,
			"camaras":     camaras,
			"kpis":        buildEmpresaCamarasKPIs(camaras),
			"catalogo":    camarasTecnologiasCatalogo(),
			"visores":     camarasVisoresCatalogo(),
			"recomendado": "Para ver RTSP/ONVIF en navegador use un gateway HLS, WebRTC o MJPEG en la red local.",
		})
	case "catalogo":
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "tecnologias": camarasTecnologiasCatalogo(), "visores": camarasVisoresCatalogo()})
	case "camaras":
		includeInactive := strings.TrimSpace(r.URL.Query().Get("include_inactive")) == "1"
		camaras, err := dbpkg.ListEmpresaCamaras(dbEmp, empresaID, includeInactive)
		if err != nil {
			http.Error(w, "No se pudieron listar las camaras", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": camaras})
	case "camara":
		id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
		if id <= 0 {
			http.Error(w, "id requerido", http.StatusBadRequest)
			return
		}
		item, err := dbpkg.GetEmpresaCamara(dbEmp, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "camara no encontrada para esta empresa", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo cargar la camara", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "camara": item})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleEmpresaCamarasWrite(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	if action == "" {
		action = "camara"
	}
	if action != "camara" {
		http.Error(w, "accion no soportada", http.StatusBadRequest)
		return
	}
	var payload camarasWritePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	item := payload.Camara
	item.EmpresaID = empresaID
	item.UsuarioCreador = adminEmailFromRequest(r)
	if item.Estado == "" {
		item.Estado = "activo"
	}
	if !strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") && !item.Activa {
		item.Activa = true
	}
	if err := validateEmpresaCamaraURLs(item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := dbpkg.UpsertEmpresaCamara(dbEmp, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
}

func handleEmpresaCamarasDelete(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	if action == "" {
		action = "camara"
	}
	if action != "camara" {
		http.Error(w, "accion no soportada", http.StatusBadRequest)
		return
	}
	id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if id <= 0 {
		http.Error(w, "id requerido", http.StatusBadRequest)
		return
	}
	if err := dbpkg.DesactivarEmpresaCamara(dbEmp, empresaID, id, adminEmailFromRequest(r)); err != nil {
		http.Error(w, "No se pudo desactivar la camara", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
}

func buildEmpresaCamarasKPIs(camaras []dbpkg.EmpresaCamara) map[string]interface{} {
	out := map[string]interface{}{"total": len(camaras), "activas": 0, "con_estacion": 0, "visores_web": 0}
	protocolos := map[string]int{}
	for _, cam := range camaras {
		if cam.Activa && !strings.EqualFold(cam.Estado, "inactivo") {
			out["activas"] = out["activas"].(int) + 1
		}
		if cam.EstacionID > 0 || cam.CargarEnEstaciones {
			out["con_estacion"] = out["con_estacion"].(int) + 1
		}
		if cam.URLEmbed != "" || cam.URLSnapshot != "" || strings.EqualFold(cam.ProtocoloOrigen, dbpkg.CamaraProtocoloHLS) || strings.EqualFold(cam.ProtocoloOrigen, dbpkg.CamaraProtocoloWebRTC) || strings.EqualFold(cam.ProtocoloOrigen, dbpkg.CamaraProtocoloMJPEG) {
			out["visores_web"] = out["visores_web"].(int) + 1
		}
		protocolos[strings.ToLower(strings.TrimSpace(cam.ProtocoloOrigen))]++
	}
	out["protocolos"] = protocolos
	return out
}

func validateEmpresaCamaraURLs(item dbpkg.EmpresaCamara) error {
	values := []string{item.URLSnapshot, item.URLEmbed}
	for _, value := range values {
		clean := strings.ToLower(strings.TrimSpace(value))
		if clean == "" {
			continue
		}
		if strings.HasPrefix(clean, "javascript:") || strings.HasPrefix(clean, "data:") {
			return errors.New("url de camara no permitida")
		}
		if !(strings.HasPrefix(clean, "http://") || strings.HasPrefix(clean, "https://") || strings.HasPrefix(clean, "/")) {
			return errors.New("el visor web de la camara debe usar http, https o una ruta interna")
		}
	}
	stream := strings.ToLower(strings.TrimSpace(item.URLStream))
	if stream == "" {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(item.ProtocoloOrigen)) {
	case dbpkg.CamaraProtocoloRTSP:
		if !strings.HasPrefix(stream, "rtsp://") {
			return errors.New("la URL RTSP debe iniciar con rtsp://")
		}
	case dbpkg.CamaraProtocoloHLS, dbpkg.CamaraProtocoloWebRTC, dbpkg.CamaraProtocoloMJPEG, dbpkg.CamaraProtocoloIframe:
		if !(strings.HasPrefix(stream, "http://") || strings.HasPrefix(stream, "https://") || strings.HasPrefix(stream, "/")) {
			return errors.New("el stream web debe usar http, https o una ruta interna")
		}
	}
	return nil
}

func camarasVisoresCatalogo() []map[string]string {
	return []map[string]string{
		{"clave": "auto", "nombre": "Automático"},
		{"clave": "hls", "nombre": "HLS / m3u8"},
		{"clave": "webrtc", "nombre": "WebRTC baja latencia"},
		{"clave": "mjpeg", "nombre": "MJPEG"},
		{"clave": "iframe", "nombre": "Iframe / visor web DVR"},
		{"clave": "rtsp_gateway", "nombre": "RTSP mediante gateway"},
	}
}

func camarasTecnologiasCatalogo() []camaraTecnologiaCatalogo {
	return []camaraTecnologiaCatalogo{
		{Clave: dbpkg.CamaraProtocoloRTSP, Nombre: "RTSP", Uso: "Stream directo de DVR/NVR o camara IP.", Visores: []string{"rtsp_gateway", "hls", "webrtc"}, Observacion: "El navegador no reproduce RTSP directo; conviertalo a HLS, WebRTC o MJPEG con un gateway local."},
		{Clave: dbpkg.CamaraProtocoloONVIF, Nombre: "ONVIF Profile S/T", Uso: "Descubrimiento, metadatos y control PTZ en camaras compatibles.", Visores: []string{"rtsp_gateway", "hls", "webrtc"}, Observacion: "ONVIF normalmente entrega informacion y URLs RTSP; la visualizacion web requiere gateway."},
		{Clave: dbpkg.CamaraProtocoloHLS, Nombre: "HLS", Uso: "Video en tiempo real por archivos .m3u8 desde un gateway.", Visores: []string{"hls"}, Observacion: "Es estable para navegadores; la latencia suele ser mayor que WebRTC."},
		{Clave: dbpkg.CamaraProtocoloWebRTC, Nombre: "WebRTC", Uso: "Video de baja latencia para monitoreo operativo.", Visores: []string{"webrtc", "iframe"}, Observacion: "Recomendado cuando el gateway o DVR lo soporte."},
		{Clave: dbpkg.CamaraProtocoloMJPEG, Nombre: "MJPEG", Uso: "Stream HTTP simple por imagenes continuas.", Visores: []string{"mjpeg"}, Observacion: "Ligero de integrar, puede consumir mas ancho de banda."},
		{Clave: dbpkg.CamaraProtocoloIframe, Nombre: "Visor web DVR", Uso: "Inserta un visor web compatible publicado por el DVR o gateway.", Visores: []string{"iframe"}, Observacion: "Debe permitir iframe desde el mismo dominio o una URL confiable."},
	}
}
