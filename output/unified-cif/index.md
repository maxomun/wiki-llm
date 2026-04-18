# Api CIF

- Version: `1.0.1`
- Endpoints: `12`
- Fuente: `./docs/openapi.yaml, ./docs/ob_api-cif.postman_collection.json`

## Documentacion

- [Detalle de la API](apis/api-cif.md)

## Endpoints

| Metodo | BasePath | Path | OperationId | Sources |
|---|---|---|---|---|
| `GET` | `/banco/api-cif` | `/clientes` | `BuscarPorRut` | `openapi` |
| `GET` | `/banco/api-cif` | `/clientes/{id}` | `BuscarPorId` | `openapi` |
| `PATCH` | `/banco/api-cif` | `/clientes/{id}` | `UpdateContactoCliente` | `openapi` |
| `GET` | `/banco/api-cif` | `/clientes/{id}/ejecutivo` | `GetEjecutivoPorCliente` | `openapi` |
| `GET` | `/banco/api-cif` | `/segmentos` | `GetSegmentos` | `openapi` |
| `POST` | `/banco/api-cliente-tef/1.0` | `/transferencias` | `POST_ejecutar_transferencia` | `postman` |
| `GET` | `/banco/api-cliente-tef/1.0` | `/transferencias/enviadas` | `GET_deprecado_listar_transferencias_enviadas` | `postman` |
| `POST` | `/banco/api-cliente-tef/1.0` | `/transferencias/enviadas` | `POST_insertar_tef_enviada` | `postman` |
| `GET` | `/banco/api-cliente-tef/1.0` | `/transferencias/enviadas/{id}` | `GET_consulta_tef_enviada` | `postman` |
| `GET` | `/banco/api-cliente-tef/1.0` | `/transferencias/recibidas` | `GET_listar_transferencias_recibidas` | `postman` |
| `POST` | `/banco/api-cliente-tef/1.0` | `/transferencias/recibidas` | `POST_insertar_transferencia_recibida` | `postman` |
| `GET` | `/banco/api-cliente-tef/1.0` | `/transferencias/recibidas/{id}` | `GET_consulta_tef_recibida` | `postman` |
