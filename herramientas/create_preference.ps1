$b = @{licencia_id=1; empresa_id=0; payer_email='sandbox_user@example.com'; payer_name='Sandbox Tester'} | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri 'http://localhost:8080/mercadopago/create_preference' -Body $b -ContentType 'application/json' | ConvertTo-Json -Depth 6
