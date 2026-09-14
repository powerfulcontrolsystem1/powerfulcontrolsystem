package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type weatherRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn weatherRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func TestEmpresaWeatherHandlerProxiesAllowedQueries(t *testing.T) {
	previousClient := weatherHTTPClient
	weatherHTTPClient = &http.Client{Transport: weatherRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Host {
		case "api.open-meteo.com":
			if r.URL.Query().Get("latitude") != "4.711" || r.URL.Query().Get("longitude") != "-74.0721" {
				t.Fatalf("unexpected forecast coordinates: %s", r.URL.RawQuery)
			}
			return weatherResponse(`{"current":{"temperature_2m":18.4,"weather_code":2}}`), nil
		case "geocoding-api.open-meteo.com":
			if r.URL.Query().Get("name") != "Bogota, Colombia" || r.URL.Query().Get("count") != "1" {
				t.Fatalf("unexpected geocoding query: %s", r.URL.RawQuery)
			}
			return weatherResponse(`{"results":[{"name":"Bogota","latitude":4.711,"longitude":-74.0721}]}`), nil
		default:
			t.Fatalf("unexpected weather upstream: %s", r.URL.String())
			return nil, nil
		}
	})}
	t.Cleanup(func() { weatherHTTPClient = previousClient })

	for _, testCase := range []struct {
		url      string
		contains string
	}{
		{"/api/empresa/clima?mode=forecast&latitude=4.711&longitude=-74.0721", `"temperature_2m":18.4`},
		{"/api/empresa/clima?mode=geocoding&name=Bogota%2C+Colombia", `"name":"Bogota"`},
	} {
		recorder := httptest.NewRecorder()
		EmpresaWeatherHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, testCase.url, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", testCase.url, recorder.Code, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), testCase.contains) {
			t.Fatalf("%s response missing %q: %s", testCase.url, testCase.contains, recorder.Body.String())
		}
	}
}

func TestEmpresaWeatherHandlerRejectsInvalidInputWithoutCallingProvider(t *testing.T) {
	previousClient := weatherHTTPClient
	weatherHTTPClient = &http.Client{Transport: weatherRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("provider must not be called for invalid input")
		return nil, nil
	})}
	t.Cleanup(func() { weatherHTTPClient = previousClient })

	for _, target := range []string{
		"/api/empresa/clima?mode=forecast&latitude=91&longitude=-74",
		"/api/empresa/clima?mode=forecast&latitude=4&longitude=not-a-number",
		"/api/empresa/clima?mode=geocoding&name=",
		"/api/empresa/clima?mode=unknown",
	} {
		recorder := httptest.NewRecorder()
		EmpresaWeatherHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s returned %d, want 400", target, recorder.Code)
		}
	}
}

func weatherResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
