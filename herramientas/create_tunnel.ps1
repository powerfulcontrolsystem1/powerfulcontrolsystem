$b = @{name='e2e_8080'; addr='8080'; proto='http'} | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri 'http://127.0.0.1:4040/api/tunnels' -Body $b -ContentType 'application/json' | ConvertTo-Json -Depth 6
