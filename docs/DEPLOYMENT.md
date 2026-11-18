# Despliegue rápido

## Docker

```bash
cp .env.example .env
# Editar .env: secretos, pago, proveedores y conexiones

docker build -t nexo:0.1 .
docker run -d --name nexo \
  --env-file .env \
  -p 8090:8090 \
  -v nexo-data:/var/lib/nexo \
  --restart unless-stopped \
  nexo:0.1
```

Abrir `http://localhost:8090`.

## Binario local

```bash
export NEXO_ADDR=:8090
export NEXO_DATA_DIR="$HOME/.nexo"
export NEXO_WEB_DIR="./web"
go run ./cmd/nexo
```

## Antes de exponerlo a Internet

Esta entrega es adecuada para una demostración privada o pilotos controlados. No se debe presentar todavía como SaaS multiusuario seguro. Antes de desplegar públicamente hay que implementar y probar:

- cuentas y sesiones por usuario;
- aislamiento de datos por tenant;
- TLS detrás de un reverse proxy;
- cuota y límites por usuario/IP;
- escaneo de archivos, límites por duración y limpieza de temporales;
- almacenamiento de objetos y base de datos transaccional;
- Stripe Checkout y webhooks firmados;
- backups y restauración;
- política de retención y borrado;
- pruebas de carga y revisión de seguridad.

`NEXO_ACCESS_TOKEN` protege el API con una clave compartida, no sustituye autenticación multiusuario. Los enlaces de reproducción firmados caducan a los cinco minutos.
