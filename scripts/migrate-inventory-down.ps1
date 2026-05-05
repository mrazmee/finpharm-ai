$MigrationsPath = (Resolve-Path "$PSScriptRoot\..\services\inventory\migrations").Path
$DockerImage = "migrate/migrate:v4.17.0"
$DatabaseURL = "postgres://finpharm:finpharm@postgres:5432/inventory_db?sslmode=disable"

docker run --rm `
  --network finpharm-ai_default `
  -v "${MigrationsPath}:/migrations" `
  $DockerImage `
  -path=/migrations `
  -database "$DatabaseURL" `
  down 1

if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}