param(
  [string]$Root = (Split-Path -Parent $PSScriptRoot)
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Add-Type -AssemblyName System.Drawing

$outDir = Join-Path $Root "web\img\portal-systems\realistic"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

$base = @{
  pos = "web\img\sistema punto de venta.png"
  restaurant = "web\img\sistema restaurante.png"
  motel = "web\img\sistema motel.png"
  sensor = "web\img\sistema independiente.png"
  salon = "web\img\sistema salon de belleza.png"
  taller = "web\img\sistema taller.png"
  finance = "web\img\sistema para prestamos de dinero.png"
  home = "web\img\Sistema para la vida.png"
  bar = "web\img\sistema bar.png"
  tecnico = "web\img\tecnico independiente.png"
  lavadero = "web\img\sistema lavadero de automovil.png"
  admin = "web\img\login-admin-real.png"
}

$items = @(
  @{slug="punto-venta"; title="Punto de venta"; base="pos"; lines=@("Venta", "Caja", "Factura")},
  @{slug="clientes-crm"; title="Clientes y CRM"; base="finance"; lines=@("Clientes", "Seguimiento", "Ventas")},
  @{slug="motel"; title="Motel"; base="motel"; lines=@("Estaciones", "Tarifas", "Domotica")},
  @{slug="restaurante"; title="Restaurante"; base="restaurant"; lines=@("Mesas", "Pedidos", "Cocina")},
  @{slug="control-sensor"; title="Control por sensor"; base="sensor"; lines=@("Sensores", "Alertas", "Monitoreo")},
  @{slug="hotel"; title="Hotel"; base="motel"; lines=@("Reservas", "Check-in", "Habitaciones")},
  @{slug="apartamentos-turisticos"; title="Apartamentos turisticos"; base="motel"; lines=@("Reservas", "Unidades", "Limpieza")},
  @{slug="parqueaderos-ticket-qr"; title="Parqueaderos QR"; base="pos"; lines=@("Ticket QR", "Entrada", "Salida")},
  @{slug="propiedad-horizontal"; title="Propiedad horizontal"; base="finance"; lines=@("Unidades", "Cuotas", "PQR")},
  @{slug="sistema-gimnasios"; title="Sistema para Gimnasios"; base="home"; lines=@("Membresias", "Accesos", "Clases")},
  @{slug="consultorios-odontologicos"; title="Consultorios odontologicos"; base="admin"; scene="dental"; lines=@("Pacientes", "Citas", "Historia")},
  @{slug="drogueria-farmacia"; title="Drogueria y farmacia"; base="pos"; scene="pharmacy"; lines=@("Lotes", "Inventario", "Ventas")},
  @{slug="alquileres-activos"; title="Alquileres de activos"; base="taller"; lines=@("Contratos", "Garantias", "Devolucion")},
  @{slug="domicilios-entregas"; title="Domicilios y entregas"; base="restaurant"; lines=@("Pedidos", "Rutas", "Entregas")},
  @{slug="taxi-system"; title="Taxi system"; base="sensor"; lines=@("Viajes", "Conductores", "Mapa")},
  @{slug="turnos-atencion"; title="Turnos de atencion"; base="finance"; lines=@("Filas", "Ventanillas", "Llamados")},
  @{slug="carnets-empresariales"; title="Carnets empresariales"; base="finance"; lines=@("Foto", "QR", "Impresion")},
  @{slug="produccion-mrp"; title="Produccion / MRP"; base="taller"; lines=@("Recetas", "Ordenes", "Materiales")},
  @{slug="compras-ocr-ia"; title="Compras con OCR e IA"; base="home"; lines=@("Soportes", "OCR", "Aprobacion")},
  @{slug="logistica-wms"; title="Logistica WMS"; base="lavadero"; lines=@("Picking", "Packing", "Rutas")},
  @{slug="bancos-pagos-masivos"; title="Bancos y pagos masivos"; base="finance"; lines=@("Bancos", "Lotes", "Conciliacion")},
  @{slug="gestion-documental-contratos"; title="Gestion documental"; base="finance"; lines=@("Contratos", "Vencimientos", "Firmas")},
  @{slug="tickets-ayuda-calidad"; title="Tickets y calidad"; base="finance"; lines=@("Soporte", "Prioridad", "Evidencias")},
  @{slug="tesoreria-presupuesto"; title="Tesoreria y presupuesto"; base="finance"; lines=@("Caja", "Presupuesto", "Flujo")},
  @{slug="facturacion-electronica-colombia"; title="Facturacion electronica"; base="pos"; lines=@("DIAN", "CUFE", "Envios")},
  @{slug="cierre-fiscal"; title="Cierre fiscal"; base="finance"; lines=@("Periodos", "Bloqueos", "Auditoria")},
  @{slug="centros-costo"; title="Centros de costo"; base="finance"; lines=@("Areas", "Rentabilidad", "Proyecto")},
  @{slug="activos-fijos-niif-fiscal"; title="Activos fijos NIIF/Fiscal"; base="taller"; lines=@("Activos", "Vida util", "Depreciacion")},
  @{slug="cobranza-profesional"; title="Cobranza profesional"; base="finance"; lines=@("Cartera", "Promesas", "Alertas")},
  @{slug="suite-contador"; title="Suite contador"; base="finance"; lines=@("NIIF", "Impuestos", "Reportes")},
  @{slug="certificados-tributarios"; title="Certificados tributarios"; base="finance"; lines=@("Terceros", "Retencion", "Descargas")},
  @{slug="aiu-construccion"; title="AIU construccion"; base="taller"; lines=@("Obra", "Contratos", "AIU")},
  @{slug="agencia-viajes"; title="Agencia de viajes"; base="home"; lines=@("Reservas", "Paquetes", "Vouchers")},
  @{slug="operador-turistico"; title="Operador turistico"; base="home"; lines=@("Tours", "Guias", "Check-in")},
  @{slug="eventos-boleteria"; title="Eventos y boleteria"; base="bar"; lines=@("Eventos", "Boletas QR", "Aforo")},
  @{slug="salon-spa"; title="Salon, barberia y spa"; base="salon"; lines=@("Citas", "Cabinas", "Comisiones")},
  @{slug="veterinaria-petshop"; title="Veterinaria y pet shop"; base="home"; lines=@("Mascotas", "Vacunas", "Productos")},
  @{slug="clinica-consultorios"; title="Clinica y consultorios"; base="admin"; scene="clinic"; lines=@("Citas", "Ordenes", "Pacientes")},
  @{slug="laboratorio-clinico"; title="Laboratorio clinico"; base="admin"; scene="lab"; lines=@("Muestras", "Resultados", "Calidad")},
  @{slug="guarderia-infantil"; title="Guarderia infantil"; base="home"; lines=@("Ninos", "Acudientes", "Novedades")},
  @{slug="lavanderia-tintoreria"; title="Lavanderia"; base="admin"; scene="laundry"; lines=@("Ordenes", "Prendas", "Entrega")},
  @{slug="taller-mecanico"; title="Taller mecanico"; base="taller"; scene="workshop"; lines=@("Ordenes", "Repuestos", "Garantias")},
  @{slug="transporte-carga-tms"; title="Transporte TMS"; base="sensor"; lines=@("Fletes", "Rutas", "Tracking")},
  @{slug="servicios-tecnicos"; title="Servicios tecnicos"; base="sensor"; scene="service"; lines=@("Ordenes", "Visitas", "Firmas")},
  @{slug="inmobiliaria-comercial"; title="Inmobiliaria"; base="finance"; lines=@("Propiedades", "Leads", "Contratos")},
  @{slug="seguridad-privada"; title="Seguridad privada"; base="sensor"; lines=@("Guardas", "Rondas QR", "Incidentes")},
  @{slug="club-deportivo"; title="Club deportivo"; base="home"; lines=@("Clases", "Pagos", "Asistencia")},
  @{slug="funeraria-exequial"; title="Funeraria exequial"; base="finance"; lines=@("Planes", "Servicios", "Documentos")},
  @{slug="parque-recreativo"; title="Parque recreativo"; base="bar"; lines=@("Entradas", "Aforo", "Atracciones")},
  @{slug="cooperativa-fondo"; title="Cooperativa"; base="finance"; lines=@("Asociados", "Aportes", "Creditos")},
  @{slug="capacitacion-empresarial"; title="Capacitacion empresarial"; base="finance"; lines=@("Cursos", "Asistencia", "Certificados")}
)

function Add-RoundedRectangle {
  param($Graphics, $Brush, [float]$X, [float]$Y, [float]$W, [float]$H, [float]$R)
  $path = New-Object System.Drawing.Drawing2D.GraphicsPath
  $d = $R * 2
  $path.AddArc($X, $Y, $d, $d, 180, 90)
  $path.AddArc($X + $W - $d, $Y, $d, $d, 270, 90)
  $path.AddArc($X + $W - $d, $Y + $H - $d, $d, $d, 0, 90)
  $path.AddArc($X, $Y + $H - $d, $d, $d, 90, 90)
  $path.CloseFigure()
  $Graphics.FillPath($Brush, $path)
  $path.Dispose()
}

function Add-Label {
  param($Graphics, [string]$Text, [float]$X, [float]$Y, [float]$W)
  $font = New-Object System.Drawing.Font "Arial", 10, ([System.Drawing.FontStyle]::Bold)
  $brush = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(245, 15, 23, 42))
  $sf = New-Object System.Drawing.StringFormat
  $sf.Alignment = [System.Drawing.StringAlignment]::Center
  $Graphics.DrawString($Text, $font, $brush, ([System.Drawing.RectangleF]::new($X, $Y, $W, 22)), $sf)
  $sf.Dispose()
  $brush.Dispose()
  $font.Dispose()
}

function Add-DomainScene {
  param($Graphics, $Item)
  $scene = ""
  if ($Item.ContainsKey("scene")) {
    $scene = [string]$Item.scene
  }
  if ([string]::IsNullOrWhiteSpace($scene)) { return }

  $panel = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(190, 248, 250, 252))
  $line = New-Object System.Drawing.Pen ([System.Drawing.Color]::FromArgb(210, 100, 116, 139)), 3
  $green = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(245, 15, 118, 72))
  $blue = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(235, 14, 116, 144))
  $red = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(235, 220, 38, 38))
  $yellow = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(235, 234, 179, 8))
  $dark = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(225, 15, 23, 42))
  $white = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(245, 255, 255, 255))

  switch ($scene) {
    "lab" {
      Add-RoundedRectangle $Graphics $panel 58 86 470 330 22
      Add-RoundedRectangle $Graphics $white 92 136 150 210 16
      Add-Label $Graphics "ORDENES" 103 154 128
      Add-RoundedRectangle $Graphics $blue 118 204 22 118 10
      Add-RoundedRectangle $Graphics $green 154 184 22 138 10
      Add-RoundedRectangle $Graphics $yellow 190 225 22 97 10
      $Graphics.DrawLine($line, 330, 157, 330, 286)
      $Graphics.DrawLine($line, 285, 286, 390, 286)
      $Graphics.DrawEllipse($line, 298, 118, 64, 48)
      Add-RoundedRectangle $Graphics $dark 296 312 118 26 10
      Add-Label $Graphics "MUESTRAS / RESULTADOS" 278 356 170
    }
    "clinic" {
      Add-RoundedRectangle $Graphics $panel 70 94 440 310 22
      Add-RoundedRectangle $Graphics $white 102 135 145 190 14
      Add-Label $Graphics "AGENDA CLINICA" 112 158 120
      $Graphics.FillEllipse($blue, 330, 138, 92, 92)
      Add-RoundedRectangle $Graphics $green 354 244 130 34 12
      Add-RoundedRectangle $Graphics $dark 324 296 150 36 12
      Add-Label $Graphics "PACIENTE / ORDEN" 324 348 150
    }
    "dental" {
      Add-RoundedRectangle $Graphics $panel 70 94 440 310 22
      Add-RoundedRectangle $Graphics $white 108 134 165 205 16
      Add-Label $Graphics "ODONTOGRAMA" 120 154 140
      $Graphics.DrawArc($line, 140, 202, 78, 96, 20, 320)
      $Graphics.DrawLine($line, 368, 142, 318, 284)
      $Graphics.DrawEllipse($line, 338, 116, 76, 58)
      Add-RoundedRectangle $Graphics $green 320 306 150 32 12
      Add-Label $Graphics "CITAS / TRATAMIENTO" 306 356 180
    }
    "pharmacy" {
      Add-RoundedRectangle $Graphics $panel 58 88 480 322 22
      for ($i = 0; $i -lt 4; $i++) {
        Add-RoundedRectangle $Graphics $white (86 + ($i * 98)) 135 74 145 12
        Add-RoundedRectangle $Graphics $green (103 + ($i * 98)) 162 38 74 10
      }
      Add-RoundedRectangle $Graphics $dark 118 306 330 46 14
      Add-Label $Graphics "LOTES / INVENTARIO / VENTAS" 150 318 270
    }
    "laundry" {
      Add-RoundedRectangle $Graphics $panel 54 84 485 328 22
      Add-RoundedRectangle $Graphics $white 88 122 126 186 18
      $Graphics.FillEllipse($blue, 108, 145, 86, 86)
      Add-RoundedRectangle $Graphics $green 105 320 110 32 14
      Add-RoundedRectangle $Graphics $white 258 124 210 210 18
      $Graphics.DrawArc($line, 284, 152, 62, 74, 210, 220)
      $Graphics.DrawArc($line, 374, 152, 62, 74, 110, 220)
      Add-Label $Graphics "PRENDAS / ENTREGA" 278 354 170
    }
    "workshop" {
      Add-RoundedRectangle $Graphics $panel 48 88 500 328 22
      Add-RoundedRectangle $Graphics $dark 90 250 310 62 20
      $Graphics.FillEllipse($white, 118, 292, 62, 62)
      $Graphics.FillEllipse($white, 316, 292, 62, 62)
      Add-RoundedRectangle $Graphics $yellow 150 186 230 58 18
      $toolPen = New-Object System.Drawing.Pen ([System.Drawing.Color]::FromArgb(235, 255, 255, 255)), 9
      $Graphics.DrawLine($toolPen, 420, 142, 482, 204)
      $Graphics.DrawLine($toolPen, 482, 142, 420, 204)
      $toolPen.Dispose()
      Add-Label $Graphics "ORDEN / REPUESTO / GARANTIA" 130 360 280
    }
    "service" {
      Add-RoundedRectangle $Graphics $panel 52 86 490 324 22
      Add-RoundedRectangle $Graphics $white 100 132 160 214 16
      Add-Label $Graphics "EQUIPO CLIENTE" 116 152 128
      Add-RoundedRectangle $Graphics $blue 128 205 102 78 16
      $toolPen = New-Object System.Drawing.Pen ([System.Drawing.Color]::FromArgb(240, 15, 23, 42)), 8
      $Graphics.DrawLine($toolPen, 348, 148, 454, 254)
      $Graphics.DrawLine($toolPen, 454, 148, 348, 254)
      $toolPen.Dispose()
      Add-RoundedRectangle $Graphics $green 318 296 166 38 12
      Add-Label $Graphics "VISITA / FIRMA / GARANTIA" 300 354 210
    }
  }

  $panel.Dispose()
  $line.Dispose()
  $green.Dispose()
  $blue.Dispose()
  $red.Dispose()
  $yellow.Dispose()
  $dark.Dispose()
  $white.Dispose()
}

function New-PortalPhoto {
  param($Item)
  $srcPath = Join-Path $Root $base[$Item.base]
  if (-not (Test-Path -LiteralPath $srcPath)) {
    throw "Base image missing: $srcPath"
  }

  $src = [System.Drawing.Image]::FromFile($srcPath)
  $w = 1200
  $h = 760
  $bmp = New-Object System.Drawing.Bitmap $w, $h, ([System.Drawing.Imaging.PixelFormat]::Format24bppRgb)
  $g = [System.Drawing.Graphics]::FromImage($bmp)
  $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
  $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
  $g.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
  $g.TextRenderingHint = [System.Drawing.Text.TextRenderingHint]::AntiAliasGridFit

  $scale = [Math]::Max($w / $src.Width, $h / $src.Height)
  $dw = [int]($src.Width * $scale)
  $dh = [int]($src.Height * $scale)
  $dx = [int](($w - $dw) / 2)
  $dy = [int](($h - $dh) / 2)
  $g.DrawImage($src, $dx, $dy, $dw, $dh)

  $overlay = New-Object System.Drawing.Drawing2D.LinearGradientBrush ([System.Drawing.Rectangle]::new(0, 0, $w, $h)), ([System.Drawing.Color]::FromArgb(115, 4, 17, 28)), ([System.Drawing.Color]::FromArgb(10, 0, 0, 0)), 0
  $g.FillRectangle($overlay, 0, 0, $w, $h)
  $overlay.Dispose()

  Add-DomainScene $g $Item

  $screenX = 720
  $screenY = 120
  $screenW = 405
  $screenH = 300
  if ($Item.base -in @("restaurant", "bar", "salon", "taller", "pos")) {
    $screenX = 655
    $screenY = 105
  }

  $shadow = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(85, 0, 0, 0))
  Add-RoundedRectangle $g $shadow ($screenX + 10) ($screenY + 14) $screenW $screenH 22
  $frame = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(235, 14, 25, 39))
  Add-RoundedRectangle $g $frame $screenX $screenY $screenW $screenH 22
  $screen = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(248, 248, 250, 252))
  Add-RoundedRectangle $g $screen ($screenX + 18) ($screenY + 22) ($screenW - 36) ($screenH - 44) 13

  $green = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(255, 15, 118, 72))
  $white = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::White)
  $dark = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(255, 15, 23, 42))
  $muted = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(255, 71, 85, 105))
  $fontSmallBold = New-Object System.Drawing.Font "Arial", 13, ([System.Drawing.FontStyle]::Bold)
  $fontTiny = New-Object System.Drawing.Font "Arial", 10, ([System.Drawing.FontStyle]::Regular)
  $fontMed = New-Object System.Drawing.Font "Arial", 16, ([System.Drawing.FontStyle]::Bold)

  Add-RoundedRectangle $g $green ($screenX + 34) ($screenY + 40) 145 34 8
  $g.DrawString("PCS", $fontSmallBold, $white, $screenX + 48, $screenY + 47)
  $g.DrawString($Item.title, $fontMed, $dark, $screenX + 38, $screenY + 92)
  $y = $screenY + 138
  foreach ($line in $Item.lines) {
    $rowBrush = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(255, 229, 234, 242))
    $g.FillRectangle($rowBrush, $screenX + 38, $y, 275, 24)
    $rowBrush.Dispose()
    $g.DrawString($line, $fontTiny, $muted, $screenX + 50, $y + 5)
    $okBrush = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(255, 34, 197, 94))
    $g.FillRectangle($okBrush, $screenX + 325, $y + 6, 26, 12)
    $okBrush.Dispose()
    $y += 39
  }
  Add-RoundedRectangle $g $green ($screenX + 38) ($screenY + $screenH - 68) 165 34 10
  $g.DrawString("Abrir modulo", $fontTiny, $white, $screenX + 68, $screenY + $screenH - 58)

  $titlePanel = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(218, 7, 21, 37))
  Add-RoundedRectangle $g $titlePanel 46 560 520 130 22
  $fontTitle = New-Object System.Drawing.Font "Arial", 26, ([System.Drawing.FontStyle]::Bold)
  $fontSub = New-Object System.Drawing.Font "Arial", 14, ([System.Drawing.FontStyle]::Regular)
  $softWhite = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(235, 203, 213, 225))
  $g.DrawString($Item.title, $fontTitle, $white, 74, 585)
  $g.DrawString("Trabajador operando Powerful Control System", $fontSub, $softWhite, 76, 632)
  Add-RoundedRectangle $g $green 76 660 148 20 10
  $g.DrawString("Sistema en uso", (New-Object System.Drawing.Font "Arial", 9, ([System.Drawing.FontStyle]::Bold)), $white, 103, 663)

  $logoBrush = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(230, 255, 255, 255))
  Add-RoundedRectangle $g $logoBrush 998 30 150 42 18
  $g.DrawString("PCS", (New-Object System.Drawing.Font "Arial", 18, ([System.Drawing.FontStyle]::Bold)), $green, 1037, 39)

  $out = Join-Path $outDir ($Item.slug + ".jpg")
  $codec = [System.Drawing.Imaging.ImageCodecInfo]::GetImageEncoders() | Where-Object { $_.MimeType -eq "image/jpeg" }
  $params = New-Object System.Drawing.Imaging.EncoderParameters 1
  $params.Param[0] = New-Object System.Drawing.Imaging.EncoderParameter ([System.Drawing.Imaging.Encoder]::Quality), 86L
  $bmp.Save($out, $codec, $params)

  $params.Dispose()
  $fontSmallBold.Dispose()
  $fontTiny.Dispose()
  $fontMed.Dispose()
  $fontTitle.Dispose()
  $fontSub.Dispose()
  $softWhite.Dispose()
  $logoBrush.Dispose()
  $titlePanel.Dispose()
  $green.Dispose()
  $white.Dispose()
  $dark.Dispose()
  $muted.Dispose()
  $screen.Dispose()
  $frame.Dispose()
  $shadow.Dispose()
  $g.Dispose()
  $bmp.Dispose()
  $src.Dispose()
}

foreach ($item in $items) {
  New-PortalPhoto -Item $item
}

Write-Host "generated=$($items.Count) dir=$outDir"
