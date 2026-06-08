param(
    [Parameter(Mandatory = $true)]
    [string]$XmlPath,

    [string]$ToolboxRoot = ""
)

$ErrorActionPreference = "Stop"

if (!(Test-Path -LiteralPath $XmlPath)) {
    throw "No existe el XML indicado: $XmlPath"
}

$repoRoot = Split-Path -Parent $PSScriptRoot
if ([string]::IsNullOrWhiteSpace($ToolboxRoot)) {
    $candidateRoot = Join-Path $repoRoot "documentos\referencias\dian\2026-06-08\Caja-de-herramientas-FE-V19-V2026"
    $xsdRoot = Get-ChildItem -LiteralPath $candidateRoot -Recurse -Directory -ErrorAction SilentlyContinue |
        Where-Object { $_.FullName -like "*\XSD\maindoc" } |
        Select-Object -First 1
    if ($null -eq $xsdRoot) {
        throw "No se encontro XSD\maindoc dentro de $candidateRoot"
    }
    $ToolboxRoot = Split-Path -Parent $xsdRoot.FullName
}

$xmlText = Get-Content -LiteralPath $XmlPath -Raw
if ($xmlText -match "<Invoice(\s|>)") {
    $schemaName = "UBL-Invoice-2.1.xsd"
    $namespace = "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
} elseif ($xmlText -match "<CreditNote(\s|>)") {
    $schemaName = "UBL-CreditNote-2.1.xsd"
    $namespace = "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2"
} elseif ($xmlText -match "<DebitNote(\s|>)") {
    $schemaName = "UBL-DebitNote-2.1.xsd"
    $namespace = "urn:oasis:names:specification:ubl:schema:xsd:DebitNote-2"
} else {
    throw "La raiz del XML no es Invoice, CreditNote ni DebitNote"
}

$schemaPath = Join-Path $ToolboxRoot "maindoc\$schemaName"
if (!(Test-Path -LiteralPath $schemaPath)) {
    throw "No existe el XSD DIAN esperado: $schemaPath"
}

$messages = New-Object System.Collections.Generic.List[string]
$schemas = New-Object System.Xml.Schema.XmlSchemaSet
$schemas.XmlResolver = New-Object System.Xml.XmlUrlResolver
[void]$schemas.Add($namespace, $schemaPath)
$schemas.Compile()

$settings = New-Object System.Xml.XmlReaderSettings
$settings.ValidationType = [System.Xml.ValidationType]::Schema
$settings.Schemas = $schemas
$settings.XmlResolver = New-Object System.Xml.XmlUrlResolver
$settings.add_ValidationEventHandler({
    param($sender, $eventArgs)
    $messages.Add(("{0}: {1}" -f $eventArgs.Severity, $eventArgs.Message))
})

$reader = [System.Xml.XmlReader]::Create((Resolve-Path -LiteralPath $XmlPath).Path, $settings)
try {
    while ($reader.Read()) { }
} finally {
    $reader.Close()
}

[pscustomobject]@{
    ok = ($messages.Count -eq 0)
    xml = (Resolve-Path -LiteralPath $XmlPath).Path
    schema = (Resolve-Path -LiteralPath $schemaPath).Path
    errores = @($messages)
} | ConvertTo-Json -Depth 4
