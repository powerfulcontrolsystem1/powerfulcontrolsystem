package grafologia

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestAnalyzeImageBytesProducesCoreMetrics(t *testing.T) {
	data := syntheticHandwritingPNG(t)
	result, err := AnalyzeImageBytes(data, "texto de prueba")
	if err != nil {
		t.Fatalf("AnalyzeImageBytes returned error: %v", err)
	}
	if result.Version != EngineVersion {
		t.Fatalf("unexpected version %q", result.Version)
	}
	if len(result.Metrics) != 10 {
		t.Fatalf("expected 10 metrics, got %d", len(result.Metrics))
	}
	if len(result.Traits) != 12 {
		t.Fatalf("expected 12 traits, got %d", len(result.Traits))
	}
	if result.Image.LinesDetected < 3 {
		t.Fatalf("expected multiple lines, got %d", result.Image.LinesDetected)
	}
	if result.GlobalTrust <= 0 {
		t.Fatalf("expected positive trust, got %v", result.GlobalTrust)
	}
}

func TestPreprocessArtifactsAndReportExports(t *testing.T) {
	data := syntheticHandwritingPNG(t)
	result, err := AnalyzeImageBytes(data, "")
	if err != nil {
		t.Fatalf("AnalyzeImageBytes returned error: %v", err)
	}
	preprocess, err := GeneratePreprocessArtifacts(data)
	if err != nil {
		t.Fatalf("GeneratePreprocessArtifacts returned error: %v", err)
	}
	if len(preprocess.ImageBytes) != 4 {
		t.Fatalf("expected 4 image artifacts, got %d", len(preprocess.ImageBytes))
	}
	for key, payload := range preprocess.ImageBytes {
		if len(payload) < 32 {
			t.Fatalf("artifact %s too small", key)
		}
		if !bytes.HasPrefix(payload, []byte{0x89, 'P', 'N', 'G'}) {
			t.Fatalf("artifact %s is not png", key)
		}
	}
	preprocess.ImageURLs = map[string]string{"binary": "/uploads/binary.png", "lines": "/uploads/lines.png"}
	preprocess.ImageBytes = nil
	result.Preprocess = &preprocess
	if pdf := RenderPDFReport("Prueba", result); !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("expected pdf header")
	}
	if csv := RenderCSVReport(result); !bytes.Contains([]byte(csv), []byte("metrica")) || !bytes.Contains([]byte(csv), []byte("interpretacion")) {
		t.Fatalf("csv missing expected sections: %s", csv)
	}
	if txt := RenderTextReport("Prueba", result); !bytes.Contains([]byte(txt), []byte("RESUMEN GENERAL")) {
		t.Fatalf("text report missing summary")
	}
	if doc := RenderWordReport("Prueba", result); !bytes.Contains(doc, []byte("urn:schemas-microsoft-com:office:word")) {
		t.Fatalf("word report missing office namespace")
	}
}

func syntheticHandwritingPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 900, 520))
	for y := 0; y < 520; y++ {
		for x := 0; x < 900; x++ {
			img.Set(x, y, color.White)
		}
	}
	for line := 0; line < 4; line++ {
		yBase := 80 + line*85
		for word := 0; word < 5; word++ {
			xBase := 80 + word*145
			for i := 0; i < 80; i++ {
				x := xBase + i
				y := yBase + i/8
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 2; dx++ {
						img.Set(x+dx, y+dy, color.Black)
					}
				}
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	return buf.Bytes()
}
