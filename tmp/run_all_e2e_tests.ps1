
Set-Location D:\powerfulcontrolsystem\backend

$AllTestParams = @(
    'TestBackupsEmpresariales'
    'TestEmpresaAuditoria'
    'TestAuditoriaEmpresa'
    'TestDocumentos'
    'TestEmpresaVentas'
    'TestFacturacion'
    'TestEmpresaFacturacion'
    'TestReporte'
    'TestEmpresaReportes'
)

foreach ($testName in $AllTestParams) {
    Write-Host ""
    Write-Host "-------------------------------------------------"
    Write-Host "Running tests matching: $testName"
    Write-Host "-------------------------------------------------"
    go test ./handlers ./db -run $testName -count=1
}
