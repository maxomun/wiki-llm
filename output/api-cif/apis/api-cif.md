# Api CIF

API para gestionar información de clientes

Version: `1.0.1`

## Tabla de contenidos

- [GET /clientes](#get-clientes)
- [GET /clientes/{id}](#get-clientes-id)
- [PATCH /clientes/{id}](#patch-clientes-id)
- [GET /clientes/{id}/ejecutivo](#get-clientes-id-ejecutivo)
- [GET /segmentos](#get-segmentos)

## `GET /clientes`

- OperationId: `BuscarPorRut`
- Summary: Busca cliente por RUT
- Description: Busca un cliente basado en el RUT proporcionado
- Tags: `ClienteHandler`

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
- Summary: Busca cliente por ID
- Description: Busca un cliente basado en el ID proporcionado
- Tags: `ClienteHandler`

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
- Summary: Actualiza datos de contacto del cliente
- Description: Permite actualizar el teléfono y/o correo de un cliente existente
- Tags: `ClienteHandler`

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
- Summary: Obtiene los datos del ejecutivo asociado a un cliente
- Description: Devuelve nombre, correo y teléfono del ejecutivo dado el ID del cliente
- Tags: `ClienteHandler`

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
- Summary: Obtiene el listado de segmentos
- Description: Devuelve los segmentos configurados en catálogo
- Tags: `Segmentos`

### Responses

| Status | Descripcion | Content Types | Schema |
|---|---|---|---|
| `200` | OK | application/json | `#/components/schemas/dto.SegmentoResponseDto` |
| `500` | Internal Server Error | application/json | `#/components/schemas/shared.MensajeDto-string` |

