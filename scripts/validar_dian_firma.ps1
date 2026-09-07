param(
    [Parameter(Mandatory = $true)]
    [string]$XmlPath
)

$ErrorActionPreference = "Stop"

if (!(Test-Path -LiteralPath $XmlPath)) {
    throw "No existe el XML indicado: $XmlPath"
}

Add-Type -AssemblyName System.Security.Cryptography.Xml
$document = New-Object System.Xml.XmlDocument
$document.PreserveWhitespace = $true
$document.Load((Resolve-Path -LiteralPath $XmlPath).Path)
$namespaces = New-Object System.Xml.XmlNamespaceManager($document.NameTable)
$namespaces.AddNamespace("ds", "http://www.w3.org/2000/09/xmldsig#")
$signatureNode = $document.SelectSingleNode("//ds:Signature", $namespaces)

$valid = $false
$errorMessage = ""
if ($null -eq $signatureNode) {
    $errorMessage = "No se encontro ds:Signature"
} else {
    try {
        $signedXml = New-Object System.Security.Cryptography.Xml.SignedXml($document)
        $signedXml.LoadXml($signatureNode)
        $valid = $signedXml.CheckSignature()
        if (!$valid) {
            $errorMessage = "La firma XMLDSig no coincide con el documento o certificado incluido"
        }
    } catch {
        $errorMessage = $_.Exception.Message
    }
}

[pscustomobject]@{
    ok = $valid
    xml = (Resolve-Path -LiteralPath $XmlPath).Path
    firma_presente = ($null -ne $signatureNode)
    error = $errorMessage
} | ConvertTo-Json -Depth 4

if (!$valid) {
    exit 1
}
