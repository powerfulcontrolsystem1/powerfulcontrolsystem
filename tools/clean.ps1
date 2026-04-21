$files = Get-ChildItem -Path web -Recurse -Include *.html,*.js
foreach ($file in $files) {
    if ($file.Name -eq "patito_volando.html" -or $file.Name -match "Juegos") { continue }
    $content = Get-Content $file.FullName -Raw
    
    # Remplazos HTML de estilos fijos
    $newContent = $content -replace 'style="color:\s*#[0-9A-Fa-f]{3,8};?"', ''
    $newContent = $newContent -replace 'style="background-color:\s*#[0-9A-Fa-f]{3,8};?"', ''
    $newContent = $newContent -replace 'style="background:\s*#[0-9A-Fa-f]{3,8};?"', ''
    $newContent = $newContent -replace 'style="color:\s*rgba?\([^)]+\);?"', ''
    $newContent = $newContent -replace 'style="background(-color)?:\s*rgba?\([^)]+\);?"', ''
    $newContent = $newContent -replace 'style=""', ''
    
    # Remover margins/paddings quemados si el cliente lo desea centralizado
    $newContent = $newContent -replace 'style="margin-top:[^"]*"', ''
    $newContent = $newContent -replace 'style="padding:[^"]*"', ''

    # Reemplazar js de style.color = "#xx" a className
    $newContent = $newContent -replace '\.style\.color\s*=\s*''#ef5350''', '.classList.add("text-danger")'
    $newContent = $newContent -replace '\.style\.color\s*=\s*''#4caf50''', '.classList.add("text-success")'
    $newContent = $newContent -replace '\.style\.color\s*=\s*"(#[0-9A-Fa-f]{3,8})"', ''
    
    if ($content -ne $newContent) {
        Set-Content -Path $file.FullName -Value $newContent -Encoding UTF8
        Write-Host "Updated $($file.Name)"
    }
}
