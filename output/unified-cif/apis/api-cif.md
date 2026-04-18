# Api CIF

API para gestionar información de clientes

Version: `1.0.1`

## Tabla de contenidos

- [GET /clientes](#get-clientes)
- [GET /clientes/{id}](#get-clientes-id)
- [PATCH /clientes/{id}](#patch-clientes-id)
- [GET /clientes/{id}/ejecutivo](#get-clientes-id-ejecutivo)
- [GET /segmentos](#get-segmentos)
- [POST /transferencias](#post-transferencias)
- [GET /transferencias/enviadas](#get-transferencias-enviadas)
- [POST /transferencias/enviadas](#post-transferencias-enviadas)
- [GET /transferencias/enviadas/{id}](#get-transferencias-enviadas-id)
- [GET /transferencias/recibidas](#get-transferencias-recibidas)
- [POST /transferencias/recibidas](#post-transferencias-recibidas)
- [GET /transferencias/recibidas/{id}](#get-transferencias-recibidas-id)

## `GET /clientes`

- OperationId: `BuscarPorRut`
- BasePath: `/banco/api-cif`
- Summary: Busca cliente por RUT
- Description: Busca un cliente basado en el RUT proporcionado
- Tags: `ClienteHandler`
- Sources: `openapi`

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `rut` | `query` | `false` | `string` | `-` | `-` | RUT del cliente |

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | application/json | `#/components/schemas/shared.MensajeDto-dto_ClienteDto` |
| `204` | No Content | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `400` | Bad Request | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `500` | Internal Server Error | application/json | `#/components/schemas/shared.MensajeDto-string` |

## `GET /clientes/{id}`

- OperationId: `BuscarPorId`
- BasePath: `/banco/api-cif`
- Summary: Busca cliente por ID
- Description: Busca un cliente basado en el ID proporcionado
- Tags: `ClienteHandler`
- Sources: `openapi`

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `id` | `path` | `true` | `integer` | `-` | `-` | ID del cliente |

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | application/json | `#/components/schemas/shared.MensajeDto-dto_ClienteDto` |
| `204` | No Content | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `400` | Bad Request | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `500` | Internal Server Error | application/json | `#/components/schemas/shared.MensajeDto-string` |

## `PATCH /clientes/{id}`

- OperationId: `UpdateContactoCliente`
- BasePath: `/banco/api-cif`
- Summary: Actualiza datos de contacto del cliente
- Description: Permite actualizar el teléfono y/o correo de un cliente existente
- Tags: `ClienteHandler`
- Sources: `openapi`

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `id` | `path` | `true` | `integer` | `-` | `-` | ID del cliente |

### Request Body

- Required: `true`
- Description: Datos a actualizar
- Content Types: `application/json`
- Schema: `#/components/schemas/dto.UpdateContactoClienteDto`

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `400` | Bad Request | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `404` | Not Found | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `500` | Internal Server Error | application/json | `#/components/schemas/shared.MensajeDto-string` |

## `GET /clientes/{id}/ejecutivo`

- OperationId: `GetEjecutivoPorCliente`
- BasePath: `/banco/api-cif`
- Summary: Obtiene los datos del ejecutivo asociado a un cliente
- Description: Devuelve nombre, correo y teléfono del ejecutivo dado el ID del cliente
- Tags: `ClienteHandler`
- Sources: `openapi`

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `id` | `path` | `true` | `integer` | `-` | `-` | ID del cliente |

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | application/json | `#/components/schemas/shared.MensajeDto-dto_EjecutivoDto` |
| `400` | Bad Request | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `404` | Not Found | application/json | `#/components/schemas/shared.MensajeDto-string` |
| `500` | Internal Server Error | application/json | `#/components/schemas/shared.MensajeDto-string` |

## `GET /segmentos`

- OperationId: `GetSegmentos`
- BasePath: `/banco/api-cif`
- Summary: Obtiene el listado de segmentos
- Description: Devuelve los segmentos configurados en catálogo
- Tags: `Segmentos`
- Sources: `openapi`

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | application/json | `#/components/schemas/dto.SegmentoResponseDto` |
| `500` | Internal Server Error | application/json | `#/components/schemas/shared.MensajeDto-string` |

## `POST /transferencias`

- OperationId: `POST_ejecutar_transferencia`
- BasePath: `/banco/api-cliente-tef/1.0`
- Summary: Ejecutar transferencia
- Tags: `APIM-DEV`
- Sources: `postman`

### Request Body

- Required: `true`
- Description: Body definido en Postman
- Content Types: `application/json`
- Schema: `raw`

## `GET /transferencias/enviadas`

- OperationId: `GET_deprecado_listar_transferencias_enviadas`
- BasePath: `/banco/api-cliente-tef/1.0`
- Summary: deprecado---Listar transferencias enviadas
- Tags: `APIM-DEV`
- Sources: `postman`

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `estado` | `query` | `false` | `string` | `-` | `-` | Procesado\|Rechazado Opcional |
| `fechaDesde` | `query` | `false` | `string` | `-` | `-` | Opcional |
| `fechaHasta` | `query` | `false` | `string` | `-` | `-` | Opcional |
| `origenCuenta` | `query` | `false` | `string` | `-` | `-` | Opcional |
| `rutOrigen` | `query` | `false` | `string` | `-` | `-` | Obligatorio |
| `tipo` | `query` | `false` | `string` | `-` | `-` | Todos\|EntreCuentas\|Terceros Opcional |

## `POST /transferencias/enviadas`

- OperationId: `POST_insertar_tef_enviada`
- BasePath: `/banco/api-cliente-tef/1.0`
- Summary: Insertar tef enviada
- Tags: `APIM-DEV`
- Sources: `postman`

### Request Body

- Required: `true`
- Description: Body definido en Postman
- Content Types: `application/json`
- Schema: `raw`

## `GET /transferencias/enviadas/{id}`

- OperationId: `GET_consulta_tef_enviada`
- BasePath: `/banco/api-cliente-tef/1.0`
- Summary: Consulta TEF enviada
- Tags: `APIM-DEV`
- Sources: `postman`

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `id` | `path` | `true` | `string` | `-` | `-` | - |

## `GET /transferencias/recibidas`

- OperationId: `GET_listar_transferencias_recibidas`
- BasePath: `/banco/api-cliente-tef/1.0`
- Summary: Listar transferencias recibidas
- Tags: `APIM-DEV`
- Sources: `postman`

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `destinoCuenta` | `query` | `false` | `string` | `-` | `-` | Opcional |
| `destinoRut` | `query` | `false` | `string` | `-` | `-` | Obligatorio |
| `fechaDesde` | `query` | `false` | `string` | `-` | `-` | Opcional |
| `fechaHasta` | `query` | `false` | `string` | `-` | `-` | Opcional |
| `numeroPagina` | `query` | `false` | `string` | `-` | `-` | Valor por defecto |
| `registrosPorPagina` | `query` | `false` | `string` | `-` | `-` | Valor por defecto |
| `tipo` | `query` | `false` | `string` | `-` | `-` | Todos\|EntreCuentas\|Terceros Opcional |

## `POST /transferencias/recibidas`

- OperationId: `POST_insertar_transferencia_recibida`
- BasePath: `/banco/api-cliente-tef/1.0`
- Summary: Insertar transferencia recibida
- Tags: `APIM-DEV`
- Sources: `postman`

### Request Body

- Required: `true`
- Description: Body definido en Postman
- Content Types: `application/json`
- Schema: `raw`

## `GET /transferencias/recibidas/{id}`

- OperationId: `GET_consulta_tef_recibida`
- BasePath: `/banco/api-cliente-tef/1.0`
- Summary: Consulta TEF recibida
- Tags: `APIM-DEV`
- Sources: `postman`

### Parametros

| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |
|---|---|---|---|---|---|---|
| `id` | `path` | `true` | `string` | `-` | `-` | - |

