package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDIANPrivateMigrationDeploymentContract(t *testing.T) {
	root := filepath.Clean("..")
	scriptPath := filepath.Join(root, "deploy", "scripts", "vps-migrate-private-dian.sh")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read DIAN private migration script: %v", err)
	}
	script := string(raw)
	required := []string{
		`EMPRESA_ID="${EMPRESA_ID:-}"`,
		`--empresa-id="$EMPRESA_ID"`,
		`--category=dian`,
		`--confirm=MIGRATE_PRIVATE_UPLOADS`,
		`find "$signature_dir" -mindepth 1 -maxdepth 1 -type f -name '*.pem'`,
		`chmod 0600`,
		`docker inspect "$BACKEND_CONTAINER"`,
		`readlink -f "$signature_dir"`,
	}
	for _, marker := range required {
		if !strings.Contains(script, marker) {
			t.Fatalf("DIAN private migration script missing safety marker %q", marker)
		}
	}
	if strings.Contains(script, "chown -R") || strings.Contains(script, "chmod -R") {
		t.Fatal("DIAN repair must not recursively change the full uploads tree")
	}
	if strings.Contains(script, `docker exec -i -u 0`) {
		t.Fatal("DIAN repair must not rely on root inside the cap-drop-all backend container")
	}

	dockerfileRaw, err := os.ReadFile(filepath.Join(root, "deploy", "docker", "backend.Dockerfile"))
	if err != nil {
		t.Fatalf("read backend Dockerfile: %v", err)
	}
	dockerfile := string(dockerfileRaw)
	if !strings.Contains(dockerfile, "go build -trimpath -ldflags=\"-s -w\" -o /out/pcs-migrate-private-uploads ./tools/migrate_private_uploads") {
		t.Fatal("private upload migration binary is not built into the backend image")
	}
	if !strings.Contains(dockerfile, "COPY --from=build /out/pcs-migrate-private-uploads /app/backend/pcs-migrate-private-uploads") {
		t.Fatal("private upload migration binary is not available to the backend runtime")
	}
}
