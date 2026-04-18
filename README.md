# wiki-llm

`wiki-llm` es una herramienta CLI en Go para generar documentacion tecnica en Markdown a partir de multiples fuentes de API (OpenAPI y Postman Collection).

## Objetivo del MVP

Construir un pipeline deterministico que:

1. Lea una fuente tecnica.
2. La transforme a un modelo interno.
3. Genere archivos Markdown consistentes.

En esta fase no se integra LLM. El estado actual es: **extractor + normalizacion + renderer Markdown base**.

## Estructura del proyecto

```text
wiki-llm/
  cmd/wiki-llm/main.go
  internal/
    config/
    discoverer/
    extractor/
    normalizer/
    promptbuilder/
    renderer/
    writer/
  templates/
  samples/
  output/
  docs/
  go.mod
  README.md
```

## Ejecutar localmente

Compilar:

```bash
go build ./...
```

Ejecutar tests:

```bash
go test ./...
```

Notas de validacion de CLI:

- `--source` debe apuntar a un archivo existente.
- `--source-type` soporta `auto`, `openapi` y `postman`.
- `--output` debe ser un directorio valido y escribible.

Ver ayuda global:

```bash
go run ./cmd/wiki-llm --help
```

Ver ayuda de `generate`:

```bash
go run ./cmd/wiki-llm generate --help
```

Ver ayuda de `generate api`:

```bash
go run ./cmd/wiki-llm generate api --help
```

Probar `generate api` con OpenAPI de ejemplo:

```bash
go run ./cmd/wiki-llm generate api --source ./docs/openapi.yaml --source-type openapi --output ./output/api-cif
```

Probar `generate api` con Postman Collection:

```bash
go run ./cmd/wiki-llm generate api --source ./docs/ob_api-cif.postman_collection.json --source-type postman --output ./output/postman-cif
```

Probar fusion de OpenAPI + Postman en una sola salida:

```bash
go run ./cmd/wiki-llm generate api \
  --source ./docs/openapi.yaml \
  --source ./docs/ob_api-cif.postman_collection.json \
  --output ./output/unified-cif
```

Archivos esperados en salida:

- `./output/api-cif/index.md`
- `./output/api-cif/apis/api-cif.md`
