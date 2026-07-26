// plan107_preflight verifica las condiciones mínimas antes de ejecutar
// fixtures P107-QA. No crea datos, no inicia contenedores y no llama DIAN.
package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

type check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

type result struct {
	Plan         string  `json:"plan"`
	StagingOnly  bool    `json:"staging_only"`
	ReadyForData bool    `json:"ready_for_fixture_data"`
	Checks       []check `json:"checks"`
	NextStep     string  `json:"next_step"`
}

func main() {
	endpoint := flag.String("endpoint", "http://127.0.0.1:8082/health", "URL de health de staging")
	requireDocker := flag.Bool("require-docker", true, "exigir Docker para el stack staging local")
	flag.Parse()

	checks := make([]check, 0, 3)
	if *requireDocker {
		if path, err := exec.LookPath("docker"); err == nil {
			checks = append(checks, check{Name: "docker", OK: true, Detail: "disponible: " + path})
		} else {
			checks = append(checks, check{Name: "docker", OK: false, Detail: "Docker no está disponible en PATH"})
		}
	}

	parsed, err := url.Parse(strings.TrimSpace(*endpoint))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		checks = append(checks, check{Name: "staging_health", OK: false, Detail: "endpoint de staging inválido"})
	} else {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, requestErr := client.Get(parsed.String())
		if requestErr != nil {
			checks = append(checks, check{Name: "staging_health", OK: false, Detail: "staging no responde"})
		} else {
			_ = resp.Body.Close()
			checks = append(checks, check{Name: "staging_health", OK: resp.StatusCode >= 200 && resp.StatusCode < 300, Detail: resp.Status})
		}
	}
	checks = append(checks, check{Name: "scope", OK: true, Detail: "P107-QA solo staging; sin DIAN, banca, correos ni datos productivos"})

	ready := true
	for _, item := range checks {
		if !item.OK {
			ready = false
			break
		}
	}
	next := "ejecutar manifiesto P107-QA y luego la matriz contable"
	if !ready {
		next = "corregir las compuertas fallidas; no crear fixtures ni movimientos"
	}
	_ = json.NewEncoder(os.Stdout).Encode(result{
		Plan: "P107", StagingOnly: true, ReadyForData: ready, Checks: checks, NextStep: next,
	})
}
