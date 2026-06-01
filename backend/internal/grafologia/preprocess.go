package grafologia

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

type PreprocessResult struct {
	Width      int                  `json:"width"`
	Height     int                  `json:"height"`
	Threshold  int                  `json:"threshold"`
	Lines      []PreprocessLine     `json:"lines"`
	InkBox     PreprocessBox        `json:"ink_box"`
	Steps      []PreprocessStepInfo `json:"steps"`
	ImageBytes map[string][]byte    `json:"-"`
	ImageURLs  map[string]string    `json:"image_urls,omitempty"`
	Quality    PreprocessQuality    `json:"quality"`
}

type PreprocessLine struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type PreprocessBox struct {
	MinX int `json:"min_x"`
	MinY int `json:"min_y"`
	MaxX int `json:"max_x"`
	MaxY int `json:"max_y"`
}

type PreprocessStepInfo struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PreprocessQuality struct {
	Contrast          float64 `json:"contrast"`
	InkDensity        float64 `json:"ink_density"`
	Sharpness         float64 `json:"sharpness"`
	CropSuggested     bool    `json:"crop_suggested"`
	LightingWarning   bool    `json:"lighting_warning"`
	ResolutionWarning bool    `json:"resolution_warning"`
}

func GeneratePreprocessArtifacts(data []byte) (PreprocessResult, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return PreprocessResult{}, fmt.Errorf("no se pudo leer la imagen para preprocesamiento: %w", err)
	}
	ws := buildWorkspace(img)
	if ws.width == 0 || ws.height == 0 {
		return PreprocessResult{}, fmt.Errorf("imagen invalida")
	}
	grayImg := renderGrayImage(ws)
	binaryImg := renderBinaryImage(ws)
	edgesImg, sharpness := renderEdgesImage(ws)
	linesImg := renderLineOverlay(ws, img)

	lines := make([]PreprocessLine, 0, len(ws.lines))
	for _, line := range ws.lines {
		lines = append(lines, PreprocessLine{Start: line.Start, End: line.End})
	}
	inkDensity := round4(float64(ws.inkCount) / math.Max(1, float64(ws.width*ws.height)))
	contrast := round2(imageContrast(ws.gray))
	result := PreprocessResult{
		Width:     ws.width,
		Height:    ws.height,
		Threshold: int(ws.threshold),
		Lines:     lines,
		InkBox:    PreprocessBox{MinX: ws.box.MinX, MinY: ws.box.MinY, MaxX: ws.box.MaxX, MaxY: ws.box.MaxY},
		Steps: []PreprocessStepInfo{
			{Key: "grayscale", Name: "Escala de grises", Description: "Normaliza la imagen para medir intensidad del trazo."},
			{Key: "binary", Name: "Binarización Otsu", Description: "Separa tinta y fondo con umbral automático."},
			{Key: "edges", Name: "Bordes", Description: "Calcula gradientes tipo Sobel para detectar cambios de trazo."},
			{Key: "lines", Name: "Segmentación de líneas", Description: "Marca bandas de escritura y caja envolvente de tinta."},
		},
		ImageBytes: map[string][]byte{
			"grayscale": mustPNG(grayImg),
			"binary":    mustPNG(binaryImg),
			"edges":     mustPNG(edgesImg),
			"lines":     mustPNG(linesImg),
		},
		Quality: PreprocessQuality{
			Contrast:          contrast,
			InkDensity:        inkDensity,
			Sharpness:         round2(sharpness),
			CropSuggested:     ws.box.MinX > ws.width/5 || ws.box.MinY > ws.height/5,
			LightingWarning:   contrast < 18,
			ResolutionWarning: ws.width < 700 || ws.height < 450,
		},
	}
	return result, nil
}

func renderGrayImage(ws analysisWorkspace) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, ws.width, ws.height))
	copy(img.Pix, ws.gray)
	return img
}

func renderBinaryImage(ws analysisWorkspace) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, ws.width, ws.height))
	for i, isInk := range ws.ink {
		if isInk {
			img.Pix[i] = 0
		} else {
			img.Pix[i] = 255
		}
	}
	return img
}

func renderEdgesImage(ws analysisWorkspace) (*image.Gray, float64) {
	img := image.NewGray(image.Rect(0, 0, ws.width, ws.height))
	var sum float64
	var count float64
	for y := 1; y < ws.height-1; y++ {
		for x := 1; x < ws.width-1; x++ {
			gx := -int(ws.gray[(y-1)*ws.width+x-1]) + int(ws.gray[(y-1)*ws.width+x+1]) -
				2*int(ws.gray[y*ws.width+x-1]) + 2*int(ws.gray[y*ws.width+x+1]) -
				int(ws.gray[(y+1)*ws.width+x-1]) + int(ws.gray[(y+1)*ws.width+x+1])
			gy := -int(ws.gray[(y-1)*ws.width+x-1]) - 2*int(ws.gray[(y-1)*ws.width+x]) - int(ws.gray[(y-1)*ws.width+x+1]) +
				int(ws.gray[(y+1)*ws.width+x-1]) + 2*int(ws.gray[(y+1)*ws.width+x]) + int(ws.gray[(y+1)*ws.width+x+1])
			mag := math.Sqrt(float64(gx*gx + gy*gy))
			if mag > 255 {
				mag = 255
			}
			img.Pix[y*ws.width+x] = uint8(mag)
			sum += mag
			count++
		}
	}
	return img, sum / math.Max(1, count)
}

func renderLineOverlay(ws analysisWorkspace, src image.Image) *image.RGBA {
	bounds := src.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, ws.width, ws.height))
	draw.Draw(out, out.Bounds(), src, bounds.Min, draw.Src)
	red := color.RGBA{R: 220, A: 255}
	green := color.RGBA{G: 170, A: 255}
	for _, line := range ws.lines {
		drawHorizontal(out, line.Start, green)
		drawHorizontal(out, line.End, green)
	}
	if ws.inkCount > 0 {
		drawRect(out, ws.box.MinX, ws.box.MinY, ws.box.MaxX, ws.box.MaxY, red)
	}
	return out
}

func drawHorizontal(img *image.RGBA, y int, c color.RGBA) {
	if y < 0 || y >= img.Bounds().Dy() {
		return
	}
	for x := 0; x < img.Bounds().Dx(); x++ {
		img.SetRGBA(x, y, c)
	}
}

func drawRect(img *image.RGBA, minX, minY, maxX, maxY int, c color.RGBA) {
	for x := minX; x <= maxX; x++ {
		if minY >= 0 && minY < img.Bounds().Dy() {
			img.SetRGBA(x, minY, c)
		}
		if maxY >= 0 && maxY < img.Bounds().Dy() {
			img.SetRGBA(x, maxY, c)
		}
	}
	for y := minY; y <= maxY; y++ {
		if minX >= 0 && minX < img.Bounds().Dx() {
			img.SetRGBA(minX, y, c)
		}
		if maxX >= 0 && maxX < img.Bounds().Dx() {
			img.SetRGBA(maxX, y, c)
		}
	}
}

func mustPNG(img image.Image) []byte {
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
