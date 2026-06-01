package grafologia

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestAnalyzeImageBytesProducesCoreMetrics(t *testing.T) {
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
	result, err := AnalyzeImageBytes(buf.Bytes(), "texto de prueba")
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
