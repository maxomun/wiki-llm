# Api CIF

- Version: `1.0.1`
- Endpoints: `5`
- Fuente: `./docs/openapi.yaml`

## Documentacion

- [Detalle de la API](apis/api-cif.md)

## Endpoints

| Metodo | Path | OperationId |
|---|---|---|
| `GET` | `/clientes` | `BuscarPorRut` |
| `GET` | `/clientes/{id}` | `BuscarPorId` |
| `PATCH` | `/clientes/{id}` | `UpdateContactoCliente` |
| `GET` | `/clientes/{id}/ejecutivo` | `GetEjecutivoPorCliente` |
| `GET` | `/segmentos` | `GetSegmentos` |
