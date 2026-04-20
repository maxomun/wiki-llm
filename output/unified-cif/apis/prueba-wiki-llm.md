# prueba-wiki-llm

Version: `unknown`

BasePath principal: `/banco/api-cif/1.0`

Fuente principal del contrato: `openapi`

## Tabla de contenidos

- [GET /clientes](#get-clientes)
- [GET /clientes/{id}](#get-clientes-id)
- [PATCH /clientes/{id}](#patch-clientes-id)
- [GET /clientes/{id}/ejecutivo](#get-clientes-id-ejecutivo)
- [GET /health](#get-health)
- [GET /segmentos](#get-segmentos)

## `GET /clientes`

- OperationId: `BuscarPorRut`
- BasePath: `/banco/api-cif/1.0`
- Summary: Busca cliente por RUT
- Description: Busca un cliente basado en el RUT proporcionado
- Tags: `ClienteHandler`, `APIM DEV Copy`
- Sources: `openapi`, `postman`, `code`
- Confidence: `high`

### Implementacion interna

- Handler: `BuscarPorRut`
- Archivo: `/home/max/p/wiki-llm/proyectos-a-wikear/api-cif/infrastructure/router/oapi_router.gen.go`

### Dependencias

- Base de datos: `true`
- Tipo(s) BD: `sql`
- Queries detectadas:
  - `SELECT c.PK_cliente_id AS id, c.documento_id AS idDocumento, c.nombre, c.fecha_primer_ingreso_bco AS fechaIngreso, td.PK_tipo_documento_id AS [tipoDocumento.id], td.descripcion ...`
  - `rut`
- Mensajeria: `false`
- APIs externas: `false`

### Flujo resumido

- Enruta la solicitud al handler: BuscarPorRut
- Ejecuta una operacion de consulta sobre las entidades detectadas
- Ejecuta operaciones de base de datos (tipo(s): sql)
- Aplica consultas SQL detectadas (consulta, 2 query(s)): SELECT c.PK_cliente_id AS id, c.documento_id AS idDocumento, c.nombre, c.fecha_primer_ingreso_bco AS fechaIngreso, td.PK_tipo_documento_id AS [tipoDocumento.id], td.descripcion ... (+1 mas)
- Retorna respuesta al cliente

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `rut` | `query` | `false` | `string` | `-` | `-` | RUT del cliente |

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | - | `-` |
| `204` | No Content | - | `-` |
| `400` | Bad Request | - | `-` |
| `500` | Internal Server Error | - | `-` |

## `GET /clientes/{id}`

- OperationId: `BuscarPorId`
- BasePath: `/banco/api-cif/1.0`
- Summary: Busca cliente por ID
- Description: Busca un cliente basado en el ID proporcionado
- Tags: `ClienteHandler`, `APIM DEV Copy`
- Sources: `openapi`, `postman`, `code`
- Confidence: `high`

### Implementacion interna

- Handler: `BuscarPorId`
- Archivo: `/home/max/p/wiki-llm/proyectos-a-wikear/api-cif/infrastructure/router/oapi_router.gen.go`

### Dependencias

- Base de datos: `true`
- Tipo(s) BD: `sql`
- Queries detectadas:
  - `SELECT c.PK_cliente_id AS id, c.documento_id AS idDocumento, c.nombre, c.fecha_primer_ingreso_bco AS fechaIngreso, td.PK_tipo_documento_id AS [tipoDocumento.id], td.descripcion ...`
- Mensajeria: `false`
- APIs externas: `false`

### Flujo resumido

- Enruta la solicitud al handler: BuscarPorId
- Ejecuta una operacion de consulta sobre las entidades detectadas
- Ejecuta operaciones de base de datos (tipo(s): sql)
- Aplica consultas SQL detectadas (consulta, 1 query(s)): SELECT c.PK_cliente_id AS id, c.documento_id AS idDocumento, c.nombre, c.fecha_primer_ingreso_bco AS fechaIngreso, td.PK_tipo_documento_id AS [tipoDocumento.id], td.descripcion ...
- Retorna respuesta al cliente

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `id` | `path` | `true` | `string` | `-` | `-` | ID del cliente |

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | - | `-` |
| `204` | No Content | - | `-` |
| `400` | Bad Request | - | `-` |
| `500` | Internal Server Error | - | `-` |

## `PATCH /clientes/{id}`

- OperationId: `UpdateContactoCliente`
- BasePath: `/banco/api-cif/1.0`
- Summary: Actualiza datos de contacto del cliente
- Description: Permite actualizar el teléfono y/o correo de un cliente existente
- Tags: `ClienteHandler`, `APIM DEV Copy`
- Sources: `openapi`, `postman`, `code`
- Confidence: `high`

### Implementacion interna

- Handler: `UpdateContactoCliente`
- Archivo: `/home/max/p/wiki-llm/proyectos-a-wikear/api-cif/infrastructure/router/oapi_router.gen.go`

### Dependencias

- Base de datos: `true`
- Tipo(s) BD: `sql`
- Tablas: `cif.tb_correo`, `cif.tb_telefono`
- Queries detectadas:
  - `SELECT c.PK_cliente_id AS id, c.documento_id AS idDocumento, c.nombre, c.fecha_primer_ingreso_bco AS fechaIngreso, td.PK_tipo_documento_id AS [tipoDocumento.id], td.descripcion ...`
  - `UPDATE cif.tb_correo SET correo_electronico = @correo WHERE FK_cliente_id = @idCliente AND correo_electronico_principal = 1 AND vigencia = 1`
  - `UPDATE cif.tb_telefono SET numero_telefono = @telefono WHERE FK_cliente_id = @idCliente AND telefono_principal = 1 AND vigencia = 1`
- Mensajeria: `false`
- APIs externas: `false`

### Flujo resumido

- Enruta la solicitud al handler: UpdateContactoCliente
- Ejecuta una operacion de consulta y actualizacion sobre entidades: correo, telefono
- Ejecuta operaciones de base de datos (tipo(s): sql; tablas: cif.tb_correo, cif.tb_telefono)
- Aplica consultas SQL detectadas (consulta/actualizacion, 3 query(s)): SELECT c.PK_cliente_id AS id, c.documento_id AS idDocumento, c.nombre, c.fecha_primer_ingreso_bco AS fechaIngreso, td.PK_tipo_documento_id AS [tipoDocumento.id], td.descripcion ... (+2 mas)
- Retorna respuesta al cliente

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `request` | `body` | `true` | `object` | `-` | `#/definitions/dto.UpdateContactoClienteDto` | Datos a actualizar |
| `id` | `path` | `true` | `string` | `-` | `-` | ID del cliente |

### Request Body

- Required: `true`
- Description: Body definido en Postman
- Content Types: `application/json`
- Schema: `raw`
- Example: `{      "usuario": "",      "telefono": "44444444",      "correo": "test@t.com"  }`

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | - | `-` |
| `400` | Bad Request | - | `-` |
| `404` | Not Found | - | `-` |
| `500` | Internal Server Error | - | `-` |

## `GET /clientes/{id}/ejecutivo`

- OperationId: `GetEjecutivoPorCliente`
- BasePath: `/banco/api-cif/1.0`
- Summary: Obtiene los datos del ejecutivo asociado a un cliente
- Description: Devuelve nombre, correo y teléfono del ejecutivo dado el ID del cliente
- Tags: `ClienteHandler`, `APIM DEV Copy`
- Sources: `openapi`, `postman`, `code`
- Confidence: `high`

### Implementacion interna

- Handler: `GetEjecutivoPorCliente`
- Archivo: `/home/max/p/wiki-llm/proyectos-a-wikear/api-cif/infrastructure/router/oapi_router.gen.go`

### Dependencias

- Base de datos: `true`
- Tipo(s) BD: `sql`
- Tablas: `cif.tb_ejecutivo`, `cif.tb_relacion_cliente_ejecutivo`
- Queries detectadas:
  - `SELECT c.PK_cliente_id AS id, c.documento_id AS idDocumento, c.nombre, c.fecha_primer_ingreso_bco AS fechaIngreso, td.PK_tipo_documento_id AS [tipoDocumento.id], td.descripcion ...`
  - `SELECT eje.nombre_empleado AS nombre, eje.correo AS correo, eje.telefono AS telefono FROM cif.tb_relacion_cliente_ejecutivo rce LEFT JOIN cif.tb_ejecutivo eje ON eje.pk_codigo_e...`
- Mensajeria: `false`
- APIs externas: `false`

### Flujo resumido

- Enruta la solicitud al handler: GetEjecutivoPorCliente
- Ejecuta una operacion de consulta sobre entidades: ejecutivo, relacion_cliente_ejecutivo
- Ejecuta operaciones de base de datos (tipo(s): sql; tablas: cif.tb_ejecutivo, cif.tb_relacion_cliente_ejecutivo)
- Aplica consultas SQL detectadas (consulta, 2 query(s)): SELECT c.PK_cliente_id AS id, c.documento_id AS idDocumento, c.nombre, c.fecha_primer_ingreso_bco AS fechaIngreso, td.PK_tipo_documento_id AS [tipoDocumento.id], td.descripcion ... (+1 mas)
- Retorna respuesta al cliente

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `id` | `path` | `true` | `string` | `-` | `-` | ID del cliente |

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | - | `-` |
| `400` | Bad Request | - | `-` |
| `404` | Not Found | - | `-` |
| `500` | Internal Server Error | - | `-` |

## `GET /health`

- OperationId: `GET_health`
- BasePath: `/banco/api-cif/1.0`
- Summary: Health
- Description: Indicador de salud del servicio.
- Tags: `APIM DEV Copy`
- Sources: `postman`
- Confidence: `low`

## `GET /segmentos`

- OperationId: `GetSegmentos`
- BasePath: `/banco/api-cif/1.0`
- Summary: Obtiene el listado de segmentos
- Description: Devuelve los segmentos configurados en catálogo
- Tags: `Segmentos`, `APIM DEV Copy`
- Sources: `openapi`, `postman`, `code`
- Confidence: `high`

### Implementacion interna

- Handler: `GetSegmentos`
- Archivo: `/home/max/p/wiki-llm/proyectos-a-wikear/api-cif/infrastructure/router/oapi_router.gen.go`

### Dependencias

- Base de datos: `true`
- Tipo(s) BD: `sql`
- Tablas: `catalogo.tb_segmento_cli`
- Queries detectadas:
  - `SELECT id, descripcion FROM catalogo.tb_segmento_cli ORDER BY descripcion`
- Mensajeria: `false`
- APIs externas: `false`

### Flujo resumido

- Enruta la solicitud al handler: GetSegmentos
- Ejecuta una operacion de consulta sobre entidades: segmento_cli
- Consulta repositorios: repo.ListarSegmentos
- Ejecuta operaciones de base de datos (tipo(s): sql; tablas: catalogo.tb_segmento_cli)
- Aplica consultas SQL detectadas (consulta, 1 query(s)): SELECT id, descripcion FROM catalogo.tb_segmento_cli ORDER BY descripcion
- Retorna respuesta al cliente

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | - | `-` |
| `500` | Internal Server Error | - | `-` |

