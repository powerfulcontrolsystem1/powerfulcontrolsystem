// plan107_fixture_manifest creates a reviewable staging-fixture contract.
// It never opens a database and cannot create operational data.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	dbpkg "github.com/you/pos-backend/db"
)

func main() {
	var empresaID int64
	var runID string
	flag.Int64Var(&empresaID, "empresa-id", 0, "ID de empresa de staging (obligatorio)")
	flag.StringVar(&runID, "run-id", "P107-QA", "prefijo idempotente del run")
	flag.Parse()

	manifest, err := dbpkg.BuildPlan107FixtureManifest(empresaID, runID)
	if err != nil {
		log.Fatal(err)
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Fatalf("serializar manifiesto: %v", err)
	}
	if _, err := fmt.Fprintln(os.Stdout, string(encoded)); err != nil {
		log.Fatalf("escribir manifiesto: %v", err)
	}
}
