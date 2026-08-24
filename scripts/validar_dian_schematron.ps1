param(
    [Parameter(Mandatory = $true)]
    [string]$XmlPath,

    [Parameter(Mandatory = $true)]
    [string]$SaxonJar,

    [string]$ToolboxRoot = "",

    [string]$JavaExecutable = "java",

    [string]$OutputPath = ""
)

$ErrorActionPreference = "Stop"

foreach ($requiredPath in @($XmlPath, $SaxonJar)) {
    if (!(Test-Path -LiteralPath $requiredPath -PathType Leaf)) {
        throw "No existe el archivo requerido: $requiredPath"
    }
}

$repoRoot = Split-Path -Parent $PSScriptRoot
if ([string]::IsNullOrWhiteSpace($ToolboxRoot)) {
    $candidateRoot = Join-Path $repoRoot "documentos\referencias\dian\2026-06-08\Caja-de-herramientas-FE-V19-V2026"
    $xslFile = Get-ChildItem -LiteralPath $candidateRoot -Recurse -File -Filter "DIAN-UBL21-model-compiled.xsl" -ErrorAction SilentlyContinue |
        Select-Object -First 1
    if ($null -eq $xslFile) {
        throw "No se encontro DIAN-UBL21-model-compiled.xsl dentro de $candidateRoot"
    }
    $ToolboxRoot = Split-Path -Parent $xslFile.DirectoryName
}

$xslPath = Join-Path $ToolboxRoot "XSL\DIAN-UBL21-model-compiled.xsl"
if (!(Test-Path -LiteralPath $xslPath -PathType Leaf)) {
    throw "No existe el Schematron DIAN compilado: $xslPath"
}

if ([string]::IsNullOrWhiteSpace($OutputPath)) {
    $OutputPath = "$XmlPath.schematron.xml"
}

$saxonDirectory = Split-Path -Parent (Resolve-Path -LiteralPath $SaxonJar).Path
$classPath = (Resolve-Path -LiteralPath $SaxonJar).Path
$saxonLib = Join-Path $saxonDirectory "lib\*"
if (Test-Path -LiteralPath (Join-Path $saxonDirectory "lib") -PathType Container) {
    $classPath = "$classPath;$saxonLib"
}

$messages = @(& $JavaExecutable -cp $classPath net.sf.saxon.Transform `
    "-s:$((Resolve-Path -LiteralPath $XmlPath).Path)" `
    "-xsl:$((Resolve-Path -LiteralPath $xslPath).Path)" `
    "-o:$OutputPath" 2>&1)
$exitCode = $LASTEXITCODE
$fatalMessages = @($messages | Where-Object { "$_" -like "*Fatal:*" })

[pscustomobject]@{
    ok = ($exitCode -eq 0 -and $fatalMessages.Count -eq 0)
    xml = (Resolve-Path -LiteralPath $XmlPath).Path
    schematron = (Resolve-Path -LiteralPath $xslPath).Path
    salida = $OutputPath
    codigo_saxon = $exitCode
    mensajes = @($messages | ForEach-Object { "$_" })
    fatales = @($fatalMessages | ForEach-Object { "$_" })
} | ConvertTo-Json -Depth 6

if ($exitCode -ne 0 -or $fatalMessages.Count -gt 0) {
    exit 1
}
