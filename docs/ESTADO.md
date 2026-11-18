# Estado de la implementación NEXO 0.1

## Implementado en el árbol de trabajo

- Backend Go con API HTTP.
- Persistencia JSON local con escritura temporal y rename.
- Espacios, carga de archivos y metadatos.
- Transcripciones manuales con timestamps y subtítulos SRT.
- Transcripción automática opcional por FFmpeg + API compatible con OpenAI, segmentada en bloques de 10 minutos.
- Búsqueda léxica con retorno de evidencia y tiempo exacto.
- Respuesta IA Pro opcional basada en evidencias recuperadas.
- Exportación CSV Pro.
- Plan Free limitado a tres medios por espacio; Pro por licencia manual.
- Enlace externo de pago configurable.
- Enlaces de reproducción temporales firmados para streaming/range.
- Adaptador AIP v1 para crear/iniciar organismos y escribir referencias en memoria.
- Proxy opcional para Mind, Zyrion y LispAI vía MiniNode.
- Dockerfile y CI de Go.

## Verificación ejecutada en este entorno

- `go test ./...` — PASS.
- `go vet ./...` — PASS.
- `go build -o /tmp/nexo ./cmd/nexo` — PASS.
- `node --check web/app.js` — PASS.
- Arranque HTTP y `GET /api/health` — PASS.
- Servir interfaz estática — HTTP 200.

## No implementado / no afirmado

- Comprensión visual multimodal de fotogramas.
- Extracción automática de objetos/personas/escenas.
- Búsqueda semántica real por embeddings.
- Grafo transversal de conceptos/relaciones.
- Resolución de contradicciones mediante Mind/Zyrion.
- Ejecución LispAI para reglas de dominio automáticas.
- Autenticación multiusuario, Stripe webhooks y SaaS público.
- Replicación/recovery de organismos desde la aplicación.

## Siguiente bloque de trabajo

1. Validar AIP contra un nodo PrismaTec-Core real y el MiniNode real.
2. Añadir importación directa SRT/VTT desde interfaz.
3. Añadir embeddings/reranking opcionales con evidencias obligatorias.
4. Crear extractor de keyframes y proveedor multimodal con consentimiento.
5. Implementar grafo de conocimiento con procedencia y estados Zyrion.
6. Migrar JSON a SQLite/PostgreSQL antes de multiusuario.
