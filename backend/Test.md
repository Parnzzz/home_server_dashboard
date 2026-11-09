//POST

$body = @{
    user_id = "7aa871a4-09f2-4284-9d95-bba347f3766b"
    name = "My Home PC"
    hostname = $env:COMPUTERNAME
    os = "Windows 11"
} | ConvertTo-Json,k

Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:8080/api/v1/servers" `
    -ContentType "application/json" `
    -Body $body

//GET
    Invoke-RestMethod `
    -Method Get `
    -Uri "http://localhost:8080/api/v1/servers"