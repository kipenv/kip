# envshare — Checklist de lanzamiento

Todo lo que necesitas probar, corregir y configurar antes de lanzar.

---

## 1. Verificar que el proyecto compila limpio

```bash
cd ~/proyectos/envshares/envshare

# Compilar CLI
go build -o bin/kip ./cmd/kip/

# Compilar server (necesita CGO para SQLite)
CGO_ENABLED=1 go build -o bin/kip-server ./server/cmd/server/
```

**Consideraciones:**
- Si falla el server build con errores de SQLite, necesitas `gcc` instalado (`sudo apt install gcc`)
- El CLI compila sin CGO (puro Go), pero el server necesita CGO_ENABLED=1 por `go-sqlite3`
- goreleaser tiene `CGO_ENABLED=0` para el server — esto va a fallar con SQLite. Hay dos opciones:
  1. Cambiar SQLite por una alternativa pure-Go como `modernc.org/sqlite` (no necesita CGO)
  2. O usar cross-compilation con zig/musl en goreleaser (más complejo)
  3. **La más práctica:** distribuir el server solo como Docker image y el CLI como binario

---

## 2. Correr tests

```bash
# Todos los tests
go test -race -cover ./...

# Solo crypto (los más críticos)
go test -v ./internal/crypto/...

# Solo scanner
go test -v ./internal/scanner/...

# Con coverage report HTML
make test-coverage
# Abre coverage.html en el browser
```

**Consideraciones:**
- Los tests del server que usan Redis van a fallar si no tienes Redis corriendo. Verifica con `redis-cli ping`
- Si no tienes Redis local, levanta solo Redis: `docker run -d -p 6379:6379 redis:7-alpine`
- Revisa el % de coverage — para los badges, querrás saber el número exacto

---

## 3. Correr linter

```bash
# Instalar golangci-lint si no lo tienes
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Correr
make lint
```

**Consideraciones:**
- El CI va a correr exactamente este linter. Si pasa local, pasa en CI
- La config está en `.golangci.yml` — tiene `gosec` habilitado (busca vulnerabilidades de seguridad)
- Si hay warnings que no puedes/quieres arreglar ahora, puedes agregar `//nolint:nombre` en la línea o excluirlos en `.golangci.yml`

---

## 4. Probar Docker Compose producción

```bash
# Levantar
docker compose -f deploy/docker-compose.yml up --build

# En otra terminal, verificar health
curl http://localhost:8080/health
# Debe responder: {"status":"ok"}

# Probar push/pull con el CLI apuntando al server local
./bin/kip config set server http://localhost:8080
./bin/kip push test.env
# Copiar el link y hacer pull
./bin/kip pull <link>

# Probar revoke
./bin/kip push test.env --reads 5
# Copiar el ID del link
./bin/kip revoke <link>

# Verificar rate limiting (opcional)
# Hacer más de 30 requests en un minuto y ver que devuelve 429

# Bajar todo
docker compose -f deploy/docker-compose.yml down
```

**Consideraciones:**
- El docker-compose de producción NO expone Redis al host (a diferencia del dev). Esto es intencional — Redis solo es accesible internamente
- Si cambias el rate limit: `RATE_LIMIT=60 docker compose -f deploy/docker-compose.yml up`
- Si cambias el puerto: `PORT=3000 docker compose -f deploy/docker-compose.yml up`
- El healthcheck del server usa `wget` (viene en alpine). Si falla, verifica que el server arranca correctamente en los logs
- **Redis persistence:** `appendonly yes` está habilitado. Los secrets sobreviven un restart de Redis. Esto es correcto porque Redis TTL los elimina automáticamente

---

## 5. Probar goreleaser (dry run)

```bash
# Instalar goreleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Dry run (no publica nada, solo genera los binarios)
goreleaser release --snapshot --clean

# Los binarios generados están en dist/
ls dist/
```

**Consideraciones:**
- **PROBLEMA CONOCIDO:** El server build tiene `CGO_ENABLED=0` pero usa `go-sqlite3` que necesita CGO. Opciones:
  1. **Recomendada:** Quitar el build del server del goreleaser y distribuirlo solo como Docker image. Edita `.goreleaser.yml` y elimina el bloque `builds` con `id: server`
  2. **Alternativa:** Migrar a `modernc.org/sqlite` (pure Go, sin CGO). Es un cambio de import y listo, pero hay que testearlo bien
- El CLI sí compila con `CGO_ENABLED=0` sin problemas
- `--snapshot` genera versión 0.0.0-SNAPSHOT. En CI con un tag `v1.0.0`, usa la versión real
- Verifica que los archivos en `dist/` tienen nombre correcto: `envshare_VERSION_OS_ARCH.tar.gz`

---

## 6. Probar install.sh

```bash
# Ver qué haría sin instalar (leer el script)
cat install.sh

# Para testear de verdad necesitas un release publicado en GitHub
# Por ahora, verifica la lógica:
#   - Detecta tu OS correctamente
#   - Detecta tu ARCH correctamente
#   - El URL que construye es correcto

# Test manual de detección:
uname -s  # Debe dar Linux o Darwin
uname -m  # Debe dar x86_64 o arm64/aarch64
```

**Consideraciones:**
- El script NO funciona hasta que haya al menos un release en GitHub (necesita la API de releases para obtener la versión)
- Cuando hagas el primer release:
  ```bash
  git tag v0.1.0
  git push origin v0.1.0
  # El GitHub Action de release corre goreleaser automáticamente
  # Luego testea:
  curl -fsSL https://raw.githubusercontent.com/kipenv/kip/main/install.sh | sh
  ```
- El script instala en `/usr/local/bin` — si no tienes permisos, usa `sudo`
- En macOS puede haber warnings de Gatekeeper. Los usuarios tendrían que ir a System Preferences > Security para permitir el binario

---

## 7. Verificar GitHub Actions CI

```bash
# El CI corre automáticamente cuando hagas push a main o abras un PR
# Para verificar que la config es correcta antes de pushear:

# 1. Revisa que go.mod tiene la versión correcta de Go
head -3 go.mod

# 2. Verifica que los tests pasan localmente
go test -race ./...

# 3. Verifica que el linter pasa
golangci-lint run ./...
```

**Consideraciones:**
- El CI usa `go-version-file: go.mod` — toma la versión de Go de tu go.mod automáticamente
- El CI levanta Redis como service container para los tests de integración
- Si el CI falla por SQLite/CGO: el step `Install dependencies` instala `gcc` en Ubuntu. Verifica que es suficiente
- Los badges del README apuntan a `kipenv/kip` — si tu repo está en otro org/user, actualiza los URLs en el README
- **golangci-lint-action@v6** siempre usa la última versión del linter. Si necesitas fijar una versión, agrega `version: v1.62.0` (o la que quieras)

---

## 8. Verificar badges en README

Los badges se van a mostrar como "no found" o "failing" hasta que:

1. **CI badge:** Necesitas al menos un push a `main` después de crear el workflow
2. **Release badge:** Necesitas al menos un tag `v*` pusheado

```bash
# Después de pushear todo a main, verifica los badges en:
# https://github.com/kipenv/kip
```

**Consideraciones:**
- Si los badges no se actualizan, es cache de GitHub. Agrega `?branch=main&event=push` al URL del badge para forzar refresh
- Si quieres badge de coverage, necesitas integrar con Codecov o Coveralls. Actualmente no está configurado, pero el CI genera `coverage.out`. Para agregar Codecov:
  ```yaml
  # Agregar al final del job test en ci.yml:
  - name: Upload to Codecov
    uses: codecov/codecov-action@v4
    with:
      file: coverage.out
  ```
  Y agregar el badge al README:
  ```
  [![codecov](https://codecov.io/gh/kipenv/kip/branch/main/graph/badge.svg)](https://codecov.io/gh/kipenv/kip)
  ```

---

## 9. Dominio y deploy (cuando estés listo)

### 9a. Comprar dominio

- **kipenv.dev** (si está disponible) — los `.dev` son HTTPS-only por defecto, perfecto para una herramienta de seguridad
- Alternativas: `envshare.io`, `envshare.sh`, `getenvshare.com`
- Registradores baratos: Namecheap, Cloudflare Registrar, Porkbun

### 9b. Deploy del server

**Opción A: Fly.io (recomendada para empezar)**
```bash
# Instalar flyctl
curl -L https://fly.io/install.sh | sh

# Desde la raíz del proyecto
fly launch
# Selecciona región cercana (Miami para Venezuela)

# Agregar Redis
fly redis create

# Deploy
fly deploy --dockerfile deploy/Dockerfile
```

**Opción B: VPS (DigitalOcean, Hetzner, etc.)**
```bash
# En el VPS:
git clone https://github.com/kipenv/kip.git
cd envshare
docker compose -f deploy/docker-compose.yml up -d

# Configurar reverse proxy (Caddy es el más simple):
sudo apt install caddy
# Editar /etc/caddy/Caddyfile:
# kipenv.dev {
#     reverse_proxy localhost:8080
# }
sudo systemctl restart caddy
# Caddy genera certificados HTTPS automáticamente
```

**Consideraciones:**
- Fly.io tiene free tier (3 shared VMs). Suficiente para arrancar
- Si usas VPS, Hetzner es el más barato (~€4/mes). DigitalOcean ~$6/mes
- **Redis en producción:** Fly.io tiene Redis managed. En VPS, el docker-compose ya incluye Redis
- **Backups:** Los secrets son efímeros (se auto-destruyen). SQLite para teams sí necesita backup. Agrega un cron: `0 */6 * * * cp /path/to/kip.db /path/to/backup/`

### 9c. Deploy de la landing page

```bash
# Opción A: Vercel (gratis)
cd web
npx vercel

# Opción B: Servida por el Go server (agregar file server al main.go)
# Opción C: Cloudflare Pages (gratis)
```

**Consideraciones:**
- Si usas Vercel/Cloudflare Pages, la landing y el API están en dominios distintos. El CORS middleware ya está configurado, pero necesitas ajustar los origins permitidos
- Si sirves todo desde Go, es un solo dominio pero necesitas build estático de Astro primero (`cd web && npm run build`)

---

## 10. Homebrew tap (después del primer release)

```bash
# Crear repo para el tap
# GitHub: kipenv/homebrew-tap

# Crear la fórmula (goreleaser puede hacerlo automáticamente)
# Agregar a .goreleaser.yml:
#
# brews:
#   - repository:
#       owner: kipenv
#       name: homebrew-tap
#     homepage: https://kipenv.dev
#     description: Share encrypted .env files with self-destructing links
#     license: MIT
#     install: |
#       bin.install "envshare"

# Después del release, los usuarios instalan con:
# brew tap kipenv/tap
# brew install envshare
```

**Consideraciones:**
- goreleaser puede crear la fórmula automáticamente si agregas el bloque `brews` a `.goreleaser.yml`
- Necesitas un GitHub token con permisos de escritura en el repo `homebrew-tap`
- Para que funcione `brew install envshare` sin `tap`, necesitas enviar la fórmula al core de Homebrew — requiere que el proyecto tenga tracción (notability guidelines)

---

## 11. Antes de hacer el primer release

Checklist final:

- [ ] Todos los tests pasan: `go test -race ./...`
- [ ] Linter pasa: `make lint`
- [ ] Docker Compose producción funciona: push + pull + revoke
- [ ] goreleaser dry run pasa (al menos el CLI)
- [ ] README no tiene placeholders/TODOs visibles
- [ ] GIF demo se ve bien en GitHub
- [ ] Links del README apuntan al repo correcto (`kipenv/kip`)
- [ ] `go install github.com/kipenv/kip@latest` funciona (después de push a main)
- [ ] Decidir si el server se distribuye como binario o solo Docker
- [ ] SECURITY.md tiene info de contacto real para reportar vulnerabilidades

```bash
# Primer release:
git tag v0.1.0
git push origin v0.1.0
# GitHub Action genera los binarios automáticamente

# Verificar:
# https://github.com/kipenv/kip/releases
```

---

## 12. Re-grabar GIF demo con dominio real

Cuando tengas dominio y server desplegado:

```bash
# 1. Actualizar fake-envshare (si sigues usando simulación)
#    o grabar con el CLI real apuntando al server de producción

# 2. Actualizar demo.tape si cambió algo

# 3. Grabar
cd ~/proyectos/envshares/envshare
vhs assets/demo.tape

# 4. Verificar el GIF
# Abrir assets/demo.gif y revisar que se ve limpio

# 5. Commit y push
```

---

## Orden recomendado

1. **Hoy:** Compilar, tests, linter (pasos 1-3)
2. **Hoy:** Docker Compose producción (paso 4)
3. **Hoy:** goreleaser dry run (paso 5) — decidir qué hacer con CGO/SQLite
4. **Cuando tengas tiempo:** Dominio + deploy (paso 9)
5. **Después del deploy:** Primer release + install.sh + badges (pasos 6-8, 11)
6. **Después del release:** Homebrew tap (paso 10)
7. **Antes de marketing:** Re-grabar GIF (paso 12)
