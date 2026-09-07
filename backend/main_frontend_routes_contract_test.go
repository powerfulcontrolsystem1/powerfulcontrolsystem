package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestFrontendEndpointsAreRegisteredInBackend(t *testing.T) {
	backendRoutes, err := collectBackendRoutes("main.go")
	if err != nil {
		t.Fatalf("collect backend routes: %v", err)
	}

	frontendEndpoints, err := collectFrontendEndpoints(filepath.Join("..", "web"))
	if err != nil {
		t.Fatalf("collect frontend endpoints: %v", err)
	}

	missing := make([]string, 0)
	for endpoint, references := range frontendEndpoints {
		if backendRouteExists(endpoint, backendRoutes) {
			continue
		}
		sort.Strings(references)
		missing = append(missing, fmt.Sprintf("%s referenced by %s", endpoint, strings.Join(references, ", ")))
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("frontend endpoints without backend route:\n%s", strings.Join(missing, "\n"))
	}
}

func collectBackendRoutes(filePath string) ([]string, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	routePattern := regexp.MustCompile(`http\.HandleFunc\("([^"]+)"`)
	matches := routePattern.FindAllStringSubmatch(string(raw), -1)
	routes := make([]string, 0, len(matches))
	for _, match := range matches {
		routes = append(routes, match[1])
	}
	return routes, nil
}

func collectFrontendEndpoints(webDir string) (map[string][]string, error) {
	endpoints := make(map[string][]string)
	endpointPattern := regexp.MustCompile("[`'\"]((?:/api/|/super/api/|/epayco/|/wompi/|/nequi/|/generate\\b|/download\\b)[^`'\"?#)\\s]*)")

	err := filepath.WalkDir(webDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules", "vendor":
				return filepath.SkipDir
			default:
				return nil
			}
		}

		extension := strings.ToLower(filepath.Ext(path))
		if extension != ".html" && extension != ".js" {
			return nil
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		reference := filepath.ToSlash(path)
		for _, match := range endpointPattern.FindAllStringSubmatch(string(raw), -1) {
			endpoint := normalizeFrontendEndpoint(match[1])
			if endpoint == "" {
				continue
			}
			if !containsString(endpoints[endpoint], reference) {
				endpoints[endpoint] = append(endpoints[endpoint], reference)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

func normalizeFrontendEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	if strings.Contains(endpoint, "+") || strings.Contains(endpoint, "${") {
		return ""
	}
	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		return ""
	}
	return endpoint
}

func backendRouteExists(endpoint string, routes []string) bool {
	for _, route := range routes {
		if endpoint == route {
			return true
		}
		if strings.HasSuffix(route, "/") && strings.HasPrefix(endpoint, route) {
			return true
		}
	}
	return false
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
