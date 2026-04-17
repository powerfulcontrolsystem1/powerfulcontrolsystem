//go:build tools
// +build tools

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	id := flag.Int64("id", 1, "empresa id")
	flag.Parse()

	db, err := sql.Open("sqlite", "empresas.db")
	if err != nil {
		log.Fatalf("open empresas db: %v", err)
	}
	defer db.Close()

	var usuarioCreador sql.NullString
	var nombre sql.NullString
	if err := db.QueryRow("SELECT COALESCE(usuario_creador,''), COALESCE(nombre,'') FROM empresas WHERE id = ? LIMIT 1", *id).Scan(&usuarioCreador, &nombre); err != nil {
		log.Fatalf("query empresa: %v", err)
	}
	fmt.Printf("empresa_id=%d nombre=%s usuario_creador=%s\n", *id, nombre.String, usuarioCreador.String)
}
