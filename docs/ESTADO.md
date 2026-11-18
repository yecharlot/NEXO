# Estado de la implementación NEXO 0.4

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
- Interfaz web para crear y consultar entidades, afirmaciones y relaciones.
- Selectores de evidencia y entidades basados en el contenido del espacio activo.
- Navegación desde cada anotación al fragmento audiovisual original.
- Validación de pertenencia de evidencias y entidades al mismo espacio.
- Estado Zyrion ternario en afirmaciones, con valor inicial 2 (Indeterminado).
- Dockerfile y CI de Go.

## Verificación ejecutada en el entorno de trabajo

- `go test ./... -count=1` — PASS.
- `git diff --check` — PASS.
- `go vet ./...` — PASS.
- `node --check web/app.js` — PASS.
- GitHub Actions de NEXO 0.1 — PASS en el commit base.

La CI de la nueva rama debe ejecutarse después de publicar los cambios; no se declara validada hasta que GitHub Actions termine.

## No implementado / no afirmado

- Comprensión visual multimodal de fotogramas.
- Extracción automática de objetos/personas/escenas/entidades/afirmaciones/relaciones.
- Búsqueda semántica real por embeddings.
- Visualización gráfica de nodos y aristas (la fase 0.3 ofrece exploración en listas estructuradas, no un grafo visual).
- Resolución de contradicciones mediante Mind/Zyrion.
- Ejecución LispAI para reglas de dominio automáticas.
- Autenticación multiusuario, Stripe webhooks y SaaS público.
- Replicación/recovery de organismos desde la aplicación.
- Validación de AIP contra un Core real y contratos MiniNode reales.

## Fase 0.4 — Acontecimientos y extracción asistida

- Modelo temporal de acontecimientos con intervalo, fuente, evidencia, procedencia y estado.
- Extractor heurístico léxico sobre segmentos de transcripción; no utiliza un LLM ni simula visión artificial.
- Propuestas pendientes separadas del conocimiento consolidado.
- Revisión humana aceptar/rechazar con nota, responsable genérico y fecha.
- Interfaz para generar propuestas, abrir evidencia y revisar decisiones.
- Corrección de actualización de fuentes en el frontend y seguimiento del procesamiento automático.
- Búsqueda robusta ante respuestas sin coincidencias (lista vacía, no `null`).
- CI ampliada con comprobación sintáctica de `web/app.js`.
- Compatibilidad con JSON anterior mediante inicialización del nuevo mapa al abrir la base.

## Límites explícitos

- La extracción puede producir falsos positivos y falsos negativos.
- La aceptación es una decisión editorial, no una verificación independiente de verdad.
- No se crean entidades, afirmaciones ni relaciones automáticamente.
- No hay extracción visual, OCR ni fotogramas en esta fase.

## Próximo bloque

1. Incorporar acontecimientos temporales como entidad de dominio explícita.
2. Diseñar extractor desacoplado que proponga, no confirme, entidades, afirmaciones y acontecimientos.
3. Añadir revisión humana y auditoría de cambios de estado.
4. Validar Mind/Zyrion/LispAI con contratos reales y pruebas de integración.
5. Migrar JSON a SQLite/PostgreSQL antes de multiusuario.
6. Implementar identidad, aislamiento tenant, pagos y cuotas antes de SaaS público.
