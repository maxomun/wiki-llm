# wiki-llm

`wiki-llm` es una herramienta CLI en Go para generar documentacion tecnica en Markdown a partir de fuentes de API, comenzando por OpenAPI.

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
go run ./cmd/wiki-llm generate api --source ./docs/openapi.yaml --output ./output/api-cif
```

Archivos esperados en salida:

- `./output/api-cif/index.md`
- `./output/api-cif/apis/api-cif.md`
