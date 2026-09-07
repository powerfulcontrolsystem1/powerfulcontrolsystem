param(
    [switch]$NoLogo
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Resolve-Path (Join-Path $ScriptDir "..")
$LogDir = Join-Path $ScriptDir "logs"
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
$LogPath = Join-Path $LogDir ("consola_de_recuperacion_{0}.log" -f (Get-Date -Format "yyyyMMdd_HHmmss"))

Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
Add-Type -AssemblyName Microsoft.VisualBasic

function Write-RecoveryLog {
    param([string]$Text)
    $line = "[{0}] {1}" -f (Get-Date -Format "HH:mm:ss"), $Text
    Add-Content -LiteralPath $LogPath -Value $line -Encoding UTF8
    if ($script:OutputBox) {
        $script:OutputBox.AppendText($line + [Environment]::NewLine)
        $script:OutputBox.SelectionStart = $script:OutputBox.TextLength
        $script:OutputBox.ScrollToCaret()
    }
}

function Set-RecoveryProgress {
    param([int]$Value, [string]$Status)
    $value = [Math]::Max(0, [Math]::Min(100, $Value))
    if ($script:ProgressBar) { $script:ProgressBar.Value = $value }
    if ($script:StatusLabel) { $script:StatusLabel.Text = ("{0}% - {1}" -f $value, $Status) }
    [System.Windows.Forms.Application]::DoEvents()
}

function Invoke-RecoveryScript {
    param(
        [string]$Title,
        [string]$ScriptPath,
        [string[]]$Arguments = @()
    )
    if (-not (Test-Path -LiteralPath $ScriptPath)) {
        Write-RecoveryLog "No existe: $ScriptPath"
        Set-RecoveryProgress 0 "Script no encontrado"
        return
    }
    Set-RecoveryProgress 5 $Title
    Write-RecoveryLog "Iniciando: $Title"
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "powershell.exe"
    $psi.WorkingDirectory = $RepoRoot.Path
    $escapedScript = "'" + ($ScriptPath -replace "'", "''") + "'"
    $escapedArgs = ($Arguments | ForEach-Object { "'" + ($_ -replace "'", "''") + "'" }) -join " "
    $psi.Arguments = "-NoProfile -ExecutionPolicy Bypass -File $escapedScript $escapedArgs"
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.UseShellExecute = $false
    $psi.CreateNoWindow = $true
    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $psi
    [void]$process.Start()
    while (-not $process.HasExited) {
        while (-not $process.StandardOutput.EndOfStream) { Write-RecoveryLog $process.StandardOutput.ReadLine() }
        while (-not $process.StandardError.EndOfStream) { Write-RecoveryLog ("ERROR: " + $process.StandardError.ReadLine()) }
        Set-RecoveryProgress ([Math]::Min(95, $script:ProgressBar.Value + 1)) $Title
        Start-Sleep -Milliseconds 250
    }
    while (-not $process.StandardOutput.EndOfStream) { Write-RecoveryLog $process.StandardOutput.ReadLine() }
    while (-not $process.StandardError.EndOfStream) { Write-RecoveryLog ("ERROR: " + $process.StandardError.ReadLine()) }
    if ($process.ExitCode -eq 0) {
        Set-RecoveryProgress 100 "$Title terminado"
        Write-RecoveryLog "Terminado correctamente: $Title"
    } else {
        Set-RecoveryProgress 0 "$Title fallo"
        Write-RecoveryLog "Fallo $Title con codigo $($process.ExitCode)"
    }
}

function Select-BackupAndRestore {
    $dialog = New-Object System.Windows.Forms.OpenFileDialog
    $dialog.Title = "Selecciona el backup .tar.gz del VPS"
    $dialog.Filter = "Backups VPS (*.tar.gz)|*.tar.gz|Todos los archivos (*.*)|*.*"
    $dialog.InitialDirectory = "D:\Backup vps PCS"
    if ($dialog.ShowDialog() -ne [System.Windows.Forms.DialogResult]::OK) { return }
    $targetHost = [Microsoft.VisualBasic.Interaction]::InputBox("IP o host del VPS nuevo:", "Restaurar a VPS nuevo", "")
    $targetHost = [string]$targetHost
    if ([string]::IsNullOrWhiteSpace($targetHost)) {
        Write-RecoveryLog "Restauracion cancelada: no se indico VPS destino."
        return
    }
    Invoke-RecoveryScript -Title "Preparar restauracion en VPS nuevo" -ScriptPath (Join-Path $ScriptDir "crear_backup_vps.ps1") -Arguments @("-Restore", "-BackupPath", $dialog.FileName, "-TargetHost", $targetHost.Trim())
}

$form = New-Object System.Windows.Forms.Form
$form.Text = "Consola de recuperacion PCS"
$form.StartPosition = "CenterScreen"
$form.Size = New-Object System.Drawing.Size(760, 520)
$form.MinimumSize = New-Object System.Drawing.Size(680, 440)

$title = New-Object System.Windows.Forms.Label
$title.Text = "Consola de recuperacion"
$title.Font = New-Object System.Drawing.Font("Segoe UI", 15, [System.Drawing.FontStyle]::Bold)
$title.AutoSize = $true
$title.Location = New-Object System.Drawing.Point(18, 16)
$form.Controls.Add($title)

$subtitle = New-Object System.Windows.Forms.Label
$subtitle.Text = "Operaciones locales para backup, restauracion, sincronizacion y reinicio del sistema PCS."
$subtitle.AutoSize = $true
$subtitle.Location = New-Object System.Drawing.Point(20, 50)
$form.Controls.Add($subtitle)

$buttonsPanel = New-Object System.Windows.Forms.FlowLayoutPanel
$buttonsPanel.Location = New-Object System.Drawing.Point(20, 82)
$buttonsPanel.Size = New-Object System.Drawing.Size(700, 98)
$buttonsPanel.Anchor = "Top,Left,Right"
$buttonsPanel.WrapContents = $true
$buttonsPanel.AutoScroll = $true
$form.Controls.Add($buttonsPanel)

function Add-RecoveryButton {
    param([string]$Text, [scriptblock]$Click)
    $button = New-Object System.Windows.Forms.Button
    $button.Text = $Text
    $button.Width = 220
    $button.Height = 38
    $button.Margin = New-Object System.Windows.Forms.Padding(4)
    $button.Add_Click($Click)
    $buttonsPanel.Controls.Add($button) | Out-Null
}

Add-RecoveryButton "Crear backup VPS" { Invoke-RecoveryScript -Title "Crear backup VPS" -ScriptPath (Join-Path $ScriptDir "crear_backup_vps.ps1") }
Add-RecoveryButton "Restaurar a VPS nuevo" { Select-BackupAndRestore }
Add-RecoveryButton "Ejecutar sync_to_vps" { Invoke-RecoveryScript -Title "sync_to_vps" -ScriptPath (Join-Path $ScriptDir "sync_to_vps.ps1") }
Add-RecoveryButton "Actualizar repositorio" { Invoke-RecoveryScript -Title "actualizar_repositorio" -ScriptPath (Join-Path $ScriptDir "actualizar_repositorio.ps1") }
Add-RecoveryButton "Ejecutar rs" { Invoke-RecoveryScript -Title "rs" -ScriptPath (Join-Path $ScriptDir "rs.ps1") }

$script:ProgressBar = New-Object System.Windows.Forms.ProgressBar
$script:ProgressBar.Location = New-Object System.Drawing.Point(20, 190)
$script:ProgressBar.Size = New-Object System.Drawing.Size(700, 20)
$script:ProgressBar.Anchor = "Top,Left,Right"
$form.Controls.Add($script:ProgressBar)

$script:StatusLabel = New-Object System.Windows.Forms.Label
$script:StatusLabel.Text = "0% - Lista"
$script:StatusLabel.AutoSize = $true
$script:StatusLabel.Location = New-Object System.Drawing.Point(20, 216)
$form.Controls.Add($script:StatusLabel)

$script:OutputBox = New-Object System.Windows.Forms.TextBox
$script:OutputBox.Multiline = $true
$script:OutputBox.ScrollBars = "Vertical"
$script:OutputBox.ReadOnly = $true
$script:OutputBox.Font = New-Object System.Drawing.Font("Consolas", 9)
$script:OutputBox.Location = New-Object System.Drawing.Point(20, 242)
$script:OutputBox.Size = New-Object System.Drawing.Size(700, 220)
$script:OutputBox.Anchor = "Top,Bottom,Left,Right"
$form.Controls.Add($script:OutputBox)

Write-RecoveryLog "Consola lista. Log: $LogPath"
[void]$form.ShowDialog()
