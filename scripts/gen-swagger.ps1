$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$server = Join-Path $root 'server'

Push-Location $server
try {
  go install github.com/swaggo/swag/cmd/swag@latest
  swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
} finally {
  Pop-Location
}