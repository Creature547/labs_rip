$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
go mod tidy
if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed" }
go fmt ./...
if ($LASTEXITCODE -ne 0) { throw "go fmt failed" }
go test ./...
if ($LASTEXITCODE -ne 0) { throw "go test failed" }
go run ./cmd/awesomeProject
