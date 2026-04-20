# prueba-wiki-llm

- Version: `unknown`
- BasePath principal: `/banco/api-cif/1.0`
- Fuente principal del contrato: `openapi`
- Endpoints: `6`
- Fuente: `/home/max/p/wiki-llm/proyectos-a-wikear/api-cif/docs/swagger.json, ./docs/prueba-wiki-llm.postman_collection.json`

## Documentacion

- [Detalle de la API](apis/prueba-wiki-llm.md)

## Endpoints

| Metodo | BasePath | Path | OperationId | Sources |
|---|---|---|---|---|
| `GET` | `/banco/api-cif/1.0` | `/clientes` | `BuscarPorRut` | `openapi, postman, code` |
| `GET` | `/banco/api-cif/1.0` | `/clientes/{id}` | `BuscarPorId` | `openapi, postman, code` |
| `PATCH` | `/banco/api-cif/1.0` | `/clientes/{id}` | `UpdateContactoCliente` | `openapi, postman, code` |
| `GET` | `/banco/api-cif/1.0` | `/clientes/{id}/ejecutivo` | `GetEjecutivoPorCliente` | `openapi, postman, code` |
| `GET` | `/banco/api-cif/1.0` | `/health` | `GET_health` | `postman` |
| `GET` | `/banco/api-cif/1.0` | `/segmentos` | `GetSegmentos` | `openapi, postman, code` |
