package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/you/pos-backend/db"
)

func PublicacionesRedSocialHandler(dbEmpresas *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
			offset, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("offset")))
			empresaID, _ := parseEmpresaIDQuery(r)
			var (
				pubs []db.PublicacionRedSocial
				err  error
			)
			if empresaID > 0 {
				pubs, err = db.GetPublicacionesRedSocialByEmpresa(dbEmpresas, int(empresaID), limit, offset)
			} else {
				pubs, err = db.GetPublicacionesRedSocialActivas(dbEmpresas, limit, offset)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pubs)
			return
		}
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func normalizeYoutubeURL(raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return ""
	}
	low := strings.ToLower(v)
	if !strings.Contains(low, "youtube.com") && !strings.Contains(low, "youtu.be") {
		return ""
	}
	return v
}

func isSafeHTTPURL(raw string) bool {
	v := strings.TrimSpace(raw)
	low := strings.ToLower(v)
	return strings.HasPrefix(low, "https://") || strings.HasPrefix(low, "http://")
}

func EmpresaPublicacionesRedSocialHandler(dbEmpresas *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := 0
		if ctxEmpresaID, ok := r.Context().Value("empresaID").(int64); ok && ctxEmpresaID > 0 {
			empresaID = int(ctxEmpresaID)
		} else if ctxEmpresaID, ok := r.Context().Value("empresa_id").(int); ok && ctxEmpresaID > 0 {
			empresaID = ctxEmpresaID
		} else if queryEmpresaID, err := parseEmpresaIDQuery(r); err == nil && queryEmpresaID > 0 {
			empresaID = int(queryEmpresaID)
		}
		if empresaID == 0 {
			http.Error(w, "Acceso denegado o empresa no seleccionada", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
			offset, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("offset")))
			pubs, err := db.GetPublicacionesRedSocialByEmpresa(dbEmpresas, empresaID, limit, offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(pubs)
			return
		}

		// Subida de imagen/video thumbnail desde dispositivo (celular/PC).
		// Guarda en web/uploads/red_social/empresa_<id>/ y devuelve URL pública.
		if r.Method == http.MethodPost && strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action"))) == "upload_media" {
			if err := r.ParseMultipartForm(12 << 20); err != nil {
				http.Error(w, "multipart invalido", http.StatusBadRequest)
				return
			}
			file, header, err := r.FormFile("archivo")
			if err != nil {
				file, header, err = r.FormFile("foto")
			}
			if err != nil {
				http.Error(w, "archivo es obligatorio", http.StatusBadRequest)
				return
			}
			defer file.Close()

			contentType := strings.ToLower(strings.TrimSpace(header.Header.Get("Content-Type")))
			if contentType == "" {
				contentType = strings.ToLower(strings.TrimSpace(r.FormValue("content_type")))
			}
			if !strings.HasPrefix(contentType, "image/") {
				http.Error(w, "solo imagenes (image/*)", http.StatusBadRequest)
				return
			}
			data, err := io.ReadAll(io.LimitReader(file, 8<<20))
			if err != nil || len(data) == 0 {
				http.Error(w, "no se pudo leer archivo", http.StatusBadRequest)
				return
			}

			ext := strings.ToLower(filepath.Ext(strings.TrimSpace(header.Filename)))
			if ext == "" {
				switch contentType {
				case "image/png":
					ext = ".png"
				case "image/webp":
					ext = ".webp"
				case "image/gif":
					ext = ".gif"
				default:
					ext = ".jpg"
				}
			}
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".gif" {
				http.Error(w, "extension no permitida", http.StatusBadRequest)
				return
			}

			webRoot := resolveWebRootDir()
			dir := filepath.Join(webRoot, "uploads", "red_social", fmt.Sprintf("empresa_%d", empresaID))
			if err := os.MkdirAll(dir, 0o700); err != nil {
				http.Error(w, "no se pudo preparar directorio", http.StatusInternalServerError)
				return
			}
			name := fmt.Sprintf("post_%d%s", time.Now().UnixNano(), ext)
			abs := filepath.Join(dir, name)
			if err := os.WriteFile(abs, data, 0o600); err != nil {
				http.Error(w, "no se pudo guardar archivo", http.StatusInternalServerError)
				return
			}
			publicURL := "/uploads/red_social/" + fmt.Sprintf("empresa_%d/", empresaID) + name
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "url": publicURL, "filename": name})
			return
		}

		if r.Method == http.MethodPost {
			var p db.PublicacionRedSocial
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			p.EmpresaID = empresaID
			if p.Estado == "" {
				p.Estado = "activo"
			}
			p.FotoURL = strings.TrimSpace(p.FotoURL)
			if p.FotoURL != "" && !strings.HasPrefix(p.FotoURL, "/uploads/") && !isSafeHTTPURL(p.FotoURL) {
				http.Error(w, "foto_url invalida", http.StatusBadRequest)
				return
			}
			p.YoutubeURL = normalizeYoutubeURL(p.YoutubeURL)
			if err := db.InsertPublicacionRedSocial(dbEmpresas, &p); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(p)
			return
		}

		if r.Method == http.MethodPut {
			idStr := strings.TrimPrefix(r.URL.Path, "/api/empresa/publicaciones/")
			id, _ := strconv.Atoi(idStr)
			var p db.PublicacionRedSocial
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			p.EmpresaID = empresaID
			p.ID = id
			p.FotoURL = strings.TrimSpace(p.FotoURL)
			if p.FotoURL != "" && !strings.HasPrefix(p.FotoURL, "/uploads/") && !isSafeHTTPURL(p.FotoURL) {
				http.Error(w, "foto_url invalida", http.StatusBadRequest)
				return
			}
			p.YoutubeURL = normalizeYoutubeURL(p.YoutubeURL)
			if err := db.UpdatePublicacionRedSocial(dbEmpresas, &p); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		if r.Method == http.MethodDelete {
			idStr := strings.TrimPrefix(r.URL.Path, "/api/empresa/publicaciones/")
			id, _ := strconv.Atoi(idStr)
			if err := db.DeletePublicacionRedSocial(dbEmpresas, id, empresaID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
