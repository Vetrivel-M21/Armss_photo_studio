# Runs the Go backend API server on its own (port/env come from .env).
Set-Location "$PSScriptRoot\..\backend"
go run ./cmd/server
