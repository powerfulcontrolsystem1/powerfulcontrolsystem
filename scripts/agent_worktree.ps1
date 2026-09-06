[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('Create', 'List', 'Lock', 'Unlock', 'Remove')]
    [string]$Action,
    [string]$Task,
    [string]$Path,
    [switch]$NoFetch
)

$ErrorActionPreference = 'Stop'

function Invoke-Git {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    & git @Arguments
    if ($LASTEXITCODE -ne 0) { throw "git $($Arguments -join ' ') fallo ($LASTEXITCODE)." }
}

function Get-WorktreeRoot {
    $repoRoot = (& git rev-parse --show-toplevel).Trim()
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($repoRoot)) { throw 'Ejecute desde un checkout Git de PCS.' }
    return (Join-Path (Split-Path -Parent (Resolve-Path -LiteralPath $repoRoot).Path) 'powerfulcontrolsystem.worktrees')
}

function Resolve-ManagedWorktree {
    param([string]$Candidate, [string]$WorktreeRoot)
    if ([string]::IsNullOrWhiteSpace($Candidate)) { throw 'Se requiere -Path.' }
    $resolved = (Resolve-Path -LiteralPath $Candidate -ErrorAction Stop).Path
    $managed = (Resolve-Path -LiteralPath $WorktreeRoot -ErrorAction Stop).Path
    if (-not $resolved.StartsWith($managed.TrimEnd('\') + '\', [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "La ruta no pertenece al directorio administrado de worktrees: $resolved"
    }
    return $resolved
}

$worktreeRoot = Get-WorktreeRoot

switch ($Action) {
    'List' { Invoke-Git worktree list --porcelain; break }
    'Create' {
        if ([string]::IsNullOrWhiteSpace($Task) -or $Task -notmatch '^[a-z0-9][a-z0-9-]{2,48}$') {
            throw 'Use -Task con un slug de 3 a 49 caracteres: minusculas, numeros y guiones.'
        }
        if (-not (Test-Path -LiteralPath $worktreeRoot)) { New-Item -ItemType Directory -Path $worktreeRoot | Out-Null }
        $managed = (Resolve-Path -LiteralPath $worktreeRoot).Path
        if (-not $NoFetch) { Invoke-Git fetch origin }
        $suffix = "$(Get-Date -Format 'yyyyMMdd')-$Task-$([guid]::NewGuid().ToString('N').Substring(0, 8))"
        $branch = "codex/$suffix"
        $newPath = Join-Path $managed $suffix
        if (Test-Path -LiteralPath $newPath) { throw "La ruta ya existe: $newPath" }
        Invoke-Git worktree add -b $branch $newPath origin/main
        Invoke-Git worktree lock --reason "agent:$Task" $newPath
        [pscustomobject]@{ Worktree = $newPath; Branch = $branch; BaseSHA = (& git -C $newPath rev-parse HEAD).Trim(); Locked = $true } | Format-List
        break
    }
    'Lock' {
        $managedPath = Resolve-ManagedWorktree -Candidate $Path -WorktreeRoot $worktreeRoot
        Invoke-Git worktree lock --reason "agent:manual" $managedPath
        break
    }
    'Unlock' {
        $managedPath = Resolve-ManagedWorktree -Candidate $Path -WorktreeRoot $worktreeRoot
        Invoke-Git worktree unlock $managedPath
        break
    }
    'Remove' {
        $managedPath = Resolve-ManagedWorktree -Candidate $Path -WorktreeRoot $worktreeRoot
        $status = @(& git -C $managedPath status --porcelain=v1)
        if ($LASTEXITCODE -ne 0 -or $status.Count -gt 0) { throw "El worktree no esta limpio: $managedPath" }
        $head = (& git -C $managedPath rev-parse HEAD).Trim()
        $branch = (& git -C $managedPath branch --show-current).Trim()
        if ([string]::IsNullOrWhiteSpace($branch)) { throw 'No se retira un worktree detached sin revisión manual.' }
        & git merge-base --is-ancestor $head origin/main
        $integrated = ($LASTEXITCODE -eq 0)
        if (-not $integrated) {
            $merged = @(& gh pr list --state merged --head $branch --limit 100 --json headRefOid | ConvertFrom-Json | Where-Object { $_.headRefOid -eq $head })
            $integrated = $merged.Count -gt 0
        }
        if (-not $integrated) { throw 'El HEAD no esta integrado en origin/main ni coincide con una PR fusionada.' }
        Invoke-Git worktree unlock $managedPath
        Invoke-Git worktree remove $managedPath
        Invoke-Git branch -D $branch
        break
    }
}
