package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// Este script en Go automatiza la instalación de RustDesk Server (hbbs y hbbr) en un VPS Ubuntu/Debian.
// Debe compilarse y ejecutarse con privilegios de root (sudo) en el VPS.
func main() {
	if os.Geteuid() != 0 {
		log.Fatalf("Este instalador debe ejecutarse como root (sudo).")
	}

	fmt.Println("Iniciando instalación de RustDesk Server...")

	commands := [][]string{
		// 1. Descargar el script de instalación oficial de RustDesk
		{"wget", "https://raw.githubusercontent.com/rustdesk/rustdesk-server/master/setup.sh", "-O", "setup.sh"},
		// 2. Dar permisos de ejecución
		{"chmod", "+x", "setup.sh"},
		// 3. Ejecutar el setup. El script de RustDesk instala hbbs (ID/Rendezvous server) y hbbr (Relay server),
		// configura systemd y abre los puertos con ufw (si está habilitado).
		{"./setup.sh"},
	}

	for _, cmdArgs := range commands {
		fmt.Printf("Ejecutando: %v\n", cmdArgs)
		// #nosec G204 -- commands is a closed installer-maintained list, not user input.
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Error ejecutando comando %v: %v", cmdArgs, err)
		}
	}

	fmt.Println("Instalación de RustDesk Server completada.")
	fmt.Println("Puertos necesarios abiertos: 21114-21119 TCP/UDP.")
	fmt.Println("Compruebe los servicios con: systemctl status rustdesk-hbbs && systemctl status rustdesk-hbbr")
}
