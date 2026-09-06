[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('Create', 'List', 'Lock', 'Unlock', 'Remove')]
    [string]$Action,

    [string]$Task,
    [string]$Path,
    [string]$Branch,
    [switch]$NoFetch
)

$ErrorActionPreference = 'Stop'

function Invoke-Git {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    & git @Arguments
    if ($LASTEXITCODE -ne 0) { throw "git $($Arguments -join ' ') fallo ($LASTEXITCODE)." }
}

function Get-RepositoryRoot {
    $root = (& git rev-parse --show-toplevel).Trim()
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($root)) {
        throw 'Ejecute este helper desde un checkout Git de PCS.'
    }
    return (Resolve-Path -LiteralPath $root).Path
}

function Get-WorktreeRoot {
    param([string]$RepositoryRoot)
    $parent = Split-Path -Parent $RepositoryRoot
    return (Join-Path $parent 'powerfulcontrolsystem.worktrees')
}

function Resolve-ManagedWorktree {
    param([string]$Candidate, [string]$WorktreeRoot)
    if ([string]::IsNullOrWhiteSpace($Candidate)) { throw 'Se requiere -Path.' }
    $resolved = (Resolve-Path -LiteralPath $Candidate -ErrorAction Stop).Path
    $managed = (Resolve-Path -LiteralPath $WorktreeRoot -ErrorAction Stop).Path
    $prefix = $managed.TrimEnd('\') + '\'
    if (-not $resolved.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "La ruta no pertenece al directorio administrado de worktrees: $resolved"
    }
    return $resolved
}

function Test-CleanWorktree {
    param([string]$WorktreePath)
    $status = @(& git -C $WorktreePath status --porcelain=v1)
    if ($LASTEXITCODE -ne 0) { throw "No se pudo leer el estado de $WorktreePath." }
    if ($status.Count -gt 0) { throw "El worktree no esta limpio: $WorktreePath" }
}

$repoRoot = Get-RepositoryRoot
$worktreeRoot = Get-WorktreeRoot -RepositoryRoot $repoRoot

switch ($Action) {
    'List' {
        Invoke-Git worktree list --porcelain
        break
    }
    'Create' {
        if ([string]::IsNullOrWhiteSpace($Task) -or $Task -notmatch '^[a-z0-9][a-z0-9-]{2,48}$') {
            throw 'Use -Task con un slug de 3 a 49 caracteres: minusculas, numeros y guiones.'
        }
        if (-not (Test-Path -LiteralPath $worktreeRoot)) {
            New-Item -ItemType Directory -Path $worktreeRoot | Out-Null
        }
        $managedRoot = (Resolve-Path -LiteralPath $worktreeRoot).Path
        if (-not $NoFetch) { Invoke-Git fetch origin }
        $date = Get-Date -Format 'yyyyMMdd'
        $id = [guid]::NewGuid().ToString('N').Substring(0, 8)
        $newBranch = "codex/$date-$Task-$id"
        $newPath = Join-Path $managedRoot "$date-$Task-$id"
        if (Test-Path -LiteralPath $newPath) { throw "La ruta ya existe: $newPath" }
        Invoke-Git worktree add -b $newBranch $newPath origin/main
        Invoke-Git worktree lock --reason "agent:$Task" $newPath
        $base = (& git -C $newPath rev-parse HEAD).Trim()
        [pscustomobject]@{ Worktree = $newPath; Branch = $newBranch; BaseSHA = $base; Locked = $true } | Format-List
        break
    }
    'Lock' {
        $managedPath = Resolve-ManagedWorktree -Candidate $Path -WorktreeRoot $worktreeRoot
        Invoke-Git worktree lock --reason ("agent:" + $(if ($Task) { $Task } else { 'manual' })) $managedPath
        break
    }
    'Unlock' {
        $managedPath = Resolve-ManagedWorktree -Candidate $Path -WorktreeRoot $worktreeRoot
        Invoke-Git worktree unlock $managedPath
        break
    }
    'Remove' {
        $managedPath = Resolve-ManagedWorktree -Candidate $Path -WorktreeRoot $worktreeRoot
        Test-CleanWorktree -WorktreePath $managedPath
        $head = (& git -C $managedPath rev-parse HEAD).Trim()
        $currentBranch = (& git -C $managedPath branch --show-current).Trim()
        if ([string]::IsNullOrWhiteSpace($currentBranch)) { throw 'No se retira un worktree detached sin revisión manual.' }

        & git merge-base --is-ancestor $head origin/main
        $integrated = ($LASTEXITCODE -eq 0)
        if (-not $integrated) {
            $merged = @(& gh pr list --state merged --head $currentBranch --limit 100 --json headRefOid | ConvertFrom-Json |
                Where-Object { $_.headRefOid -eq $head })
            $integrated = $merged.Count -gt 0
        }
        if (-not $integrated) { throw 'El HEAD no esta integrado en origin/main ni coincide con una PR fusionada.' }

        Invoke-Git worktree unlock $managedPath
        Invoke-Git worktree remove $managedPath
        Invoke-Git branch -D $currentBranch
        break
    }
}
