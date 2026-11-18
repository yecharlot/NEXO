# Estado de la implementación NEXO 0.2

## Implementado

- Backend Go con API HTTP.
- Persistencia JSON local con escritura temporal y rename.
- Compatibilidad de lectura con bases 0.1: inicialización de mapas nuevos sin migración destructiva.
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
- Memoria estructurada manual: entidades, afirmaciones y relaciones.
- Validación de pertenencia de evidencias y entidades al mismo espacio.
- Estado Zyrion ternario en afirmaciones, con valor inicial 2 (Indeterminado).
- Dockerfile y CI de Go.

## Verificación ejecutada en el entorno de trabajo

- `go test ./... -count=1` — PASS.
- `go vet ./...` — PASS.
- `node --check web/app.js` — PASS.
- GitHub Actions de NEXO 0.1 — PASS en el commit base.

La CI de la nueva rama debe ejecutarse después de publicar los cambios; no se declara validada hasta que GitHub Actions termine.

## No implementado / no afirmado

- Comprensión visual multimodal de fotogramas.
- Extracción automática de objetos/personas/escenas/entidades/afirmaciones/relaciones.
- Búsqueda semántica real por embeddings.
- Consulta y exploración interactiva del grafo.
- Resolución de contradicciones mediante Mind/Zyrion.
- Ejecución LispAI para reglas de dominio automáticas.
- Autenticación multiusuario, Stripe webhooks y SaaS público.
- Replicación/recovery de organismos desde la aplicación.
- Validación de AIP contra un Core real y contratos MiniNode reales.

## Próximo bloque

1. Añadir importación de anotaciones y visualización del grafo en interfaz.
2. Diseñar extractor multimodal desacoplado que proponga, no confirme, entidades y afirmaciones.
3. Añadir revisión humana y auditoría de cambios de estado.
4. Validar Mind/Zyrion/LispAI con contratos reales y pruebas de integración.
5. Migrar JSON a SQLite/PostgreSQL antes de multiusuario.
6. Implementar identidad, aislamiento tenant, pagos y cuotas antes de SaaS público.
