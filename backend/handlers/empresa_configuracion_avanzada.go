package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaConfiguracionAvanzadaHandler expone la configuración avanzada por empresa
// para preparación de facturación electrónica en Colombia.
func EmpresaConfiguracionAvanzadaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, empresaID)
			if err != nil {
				log.Printf("[empresa_config_avanzada] get empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo cargar la configuración avanzada", http.StatusInternalServerError)
				return
			}
			hydrateEmpresaDefaults(dbEmp, cfg)
			writeJSON(w, http.StatusOK, cfg)
			return

		case http.MethodPost, http.MethodPut:
			// Leer cuerpo y decodificar tanto a mapa (para detectar keys presentes)
			// como a la estructura tipada. Luego fusionar sobre la configuración existente
			// para soportar parches parciales sin sobrescribir valores no enviados.
			body, err := io.ReadAll(r.Body)
			if err != nil || len(body) == 0 {
				http.Error(w, "JSON invalido o cuerpo vacio", http.StatusBadRequest)
				return
			}

			var raw map[string]interface{}
			if err := json.Unmarshal(body, &raw); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}

			var payload dbpkg.EmpresaConfiguracionAvanzada
			_ = json.Unmarshal(body, &payload) // ignore error; usamos raw map para presencia de keys

			// Si empresa_id no viene en el payload, intentar obtenerlo de la query
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
					raw["empresa_id"] = empresaID
				} else if v, ok := raw["empresa_id"]; ok {
					// Si viene como numero en raw (float64), convertir
					switch vv := v.(type) {
					case float64:
						payload.EmpresaID = int64(vv)
					case string:
						// intentar parsear string numerico
						if parsed, perr := strconv.ParseInt(strings.TrimSpace(vv), 10, 64); perr == nil {
							payload.EmpresaID = parsed
						}
					}
				}
			}

			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}

			// Obtener configuración existente y fusionar con los campos enviados
			existingCfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, payload.EmpresaID)
			if err != nil {
				log.Printf("[empresa_config_avanzada] get existing empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, "No se pudo cargar la configuración existente", http.StatusInternalServerError)
				return
			}

			// serializar existing a map, sobreescribir con raw y volver a struct
			existBytes, _ := json.Marshal(existingCfg)
			var existMap map[string]interface{}
			_ = json.Unmarshal(existBytes, &existMap)
			for k, v := range raw {
				existMap[k] = v
			}
			mergedBytes, merr := json.Marshal(existMap)
			if merr != nil {
				log.Printf("[empresa_config_avanzada] marshal merged error: %v", merr)
				http.Error(w, "Error interno al procesar payload", http.StatusInternalServerError)
				return
			}

			var merged dbpkg.EmpresaConfiguracionAvanzada
			if err := json.Unmarshal(mergedBytes, &merged); err != nil {
				log.Printf("[empresa_config_avanzada] unmarshal merged error: %v", err)
				http.Error(w, "Error interno al procesar payload", http.StatusInternalServerError)
				return
			}

			merged.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, merged)
			if err != nil {
				log.Printf("[empresa_config_avanzada] upsert empresa_id=%d error: %v", merged.EmpresaID, err)
				http.Error(w, "No se pudo guardar la configuración avanzada", http.StatusInternalServerError)
				return
			}

			cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, merged.EmpresaID)
			if err != nil {
				log.Printf("[empresa_config_avanzada] get after upsert empresa_id=%d error: %v", merged.EmpresaID, err)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			}
			hydrateEmpresaDefaults(dbEmp, cfg)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"configuracion": cfg,
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func EmpresaConfiguracionAvanzadaLogoUploadHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(6 << 20); err != nil {
			http.Error(w, "payload multipart invalido", http.StatusBadRequest)
			return
		}
		empresaID, err := parseInt64Form(r, "empresa_id")
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("logo")
		if err != nil {
			http.Error(w, "logo es obligatorio", http.StatusBadRequest)
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(strings.TrimSpace(header.Filename)))
		allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true}
		if !allowed[ext] {
			http.Error(w, "extension de logo no permitida", http.StatusBadRequest)
			return
		}
		const maxLogoBytes = 5 << 20
		if header.Size > maxLogoBytes {
			http.Error(w, "el logo supera 5 MB", http.StatusBadRequest)
			return
		}

		webRoot := resolveWebRootDir()
		dir := filepath.Join(webRoot, "uploads", "empresa_logos", fmt.Sprintf("empresa_%d", empresaID))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			http.Error(w, "no se pudo preparar el directorio de logos", http.StatusInternalServerError)
			return
		}

		fileName := fmt.Sprintf("logo_%d%s", time.Now().UnixNano(), ext)
		absPath := filepath.Join(dir, fileName)
		out, err := os.Create(absPath)
		if err != nil {
			http.Error(w, "no se pudo crear el archivo de logo", http.StatusInternalServerError)
			return
		}
		written, copyErr := io.Copy(out, io.LimitReader(file, maxLogoBytes+1))
		closeErr := out.Close()
		if copyErr != nil {
			_ = os.Remove(absPath)
			http.Error(w, "no se pudo guardar el logo", http.StatusInternalServerError)
			return
		}
		if written > maxLogoBytes {
			_ = os.Remove(absPath)
			http.Error(w, "el logo supera 5 MB", http.StatusBadRequest)
			return
		}
		if closeErr != nil {
			_ = os.Remove(absPath)
			http.Error(w, "no se pudo cerrar el archivo de logo", http.StatusInternalServerError)
			return
		}

		logoURL := "/uploads/empresa_logos/empresa_" + strconv.FormatInt(empresaID, 10) + "/" + fileName
		cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, empresaID)
		if err != nil {
			log.Printf("[empresa_config_avanzada_logo] get empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "no se pudo cargar la configuracion", http.StatusInternalServerError)
			return
		}
		cfg.LogoURL = logoURL
		cfg.MostrarLogo = true
		cfg.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
		if _, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, *cfg); err != nil {
			log.Printf("[empresa_config_avanzada_logo] upsert empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "no se pudo guardar el logo en la configuracion", http.StatusInternalServerError)
			return
		}

		stored, _ := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, empresaID)
		hydrateEmpresaDefaults(dbEmp, stored)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":            true,
			"empresa_id":    empresaID,
			"logo_url":      logoURL,
			"configuracion": stored,
		})
	}
}

func hydrateEmpresaDefaults(dbEmp *sql.DB, cfg *dbpkg.EmpresaConfiguracionAvanzada) {
	if cfg == nil || cfg.EmpresaID <= 0 {
		return
	}
	if strings.TrimSpace(cfg.RazonSocial) != "" && strings.TrimSpace(cfg.NIT) != "" {
		return
	}

	var nombre string
	var nit string
	err := dbpkg.QueryRowCompat(dbEmp, `SELECT COALESCE(nombre, ''), COALESCE(nit, '') FROM empresas WHERE id = ? LIMIT 1`, cfg.EmpresaID).Scan(&nombre, &nit)
	if err != nil {
		return
	}
	if strings.TrimSpace(cfg.RazonSocial) == "" {
		cfg.RazonSocial = strings.TrimSpace(nombre)
	}
	if strings.TrimSpace(cfg.NIT) == "" {
		cfg.NIT = strings.TrimSpace(nit)
	}
}
