# Roadmap

## Fase 0.1 — MVP de evidencia audiovisual
- [x] Espacios de conocimiento.
- [x] Carga de audio/vídeo con límites.
- [x] Transcripción y subtítulos temporizados.
- [x] Búsqueda de evidencias y navegación temporal.
- [x] Exportación CSV Pro.
- [x] Enlace de pago externo y activación manual de licencia.
- [x] Adaptadores AIP y MiniNode opcionales.
- [x] CI inicial.

## Fase 0.2 — Memoria estructurada con procedencia (implementada en esta rama)
- [x] Modelo de entidades vinculadas a evidencias.
- [x] Modelo de afirmaciones con referencias temporales.
- [x] Modelo de relaciones entre entidades con evidencias.
- [x] Estados Zyrion `1/0/2`, con valor inicial indeterminado.
- [x] Validación de referencias dentro del mismo espacio.
- [x] Compatibilidad con bases JSON 0.1.
- [x] Pruebas de API y persistencia.
- [ ] Interfaz para crear/revisar anotaciones y explorar el grafo.

## Fase 0.3 — Explorador de conocimiento (implementado en esta rama)
- [x] Crear entidades desde la interfaz web.
- [x] Crear afirmaciones con selección explícita del estado Zyrion.
- [x] Crear relaciones entre entidades del mismo espacio.
- [x] Vincular anotaciones a evidencias existentes.
- [x] Abrir el audiovisual desde cada evidencia vinculada.
- [ ] Grafo visual interactivo con nodos y aristas.

## Fase 0.4 — Acontecimientos y extracción asistida (implementación inicial)
- [x] Modelo de acontecimientos con intervalo temporal y evidencia.
- [x] Extractor heurístico de señales desde transcripciones.
- [x] Propuestas pendientes separadas del conocimiento consolidado.
- [x] Revisión humana aceptar/rechazar con nota y fecha.
- [x] Procedencia del extractor y versión.
- [x] Interfaz de extracción, revisión y navegación a evidencia.
- [ ] Extracción de fotogramas, OCR y visión multimodal.
- [ ] Extractor semántico configurable mediante proveedor LLM, con evaluación y trazas.

## Fase 0.5 — Memoria transversal y razonamiento
- [ ] Comparación de ideas entre fuentes.
- [ ] Consulta de grafo y búsqueda semántica/vectorial.
- [ ] Detección de contradicciones sin colapsar incertidumbre.
- [ ] Mind/Zyrion con pruebas de reglas y profundidad fractal controlada.
- [ ] LispAI para reglas declarativas versionadas y sandboxed.

## Fase 0.6 — SaaS público
- [ ] Usuarios, organizaciones, roles y aislamiento multi-tenant.
- [ ] Stripe Checkout + webhook firmado + portal de suscripciones.
- [ ] Cuotas, límites de almacenamiento y procesamiento.
- [ ] Almacenamiento de objetos y base de datos transaccional.
- [ ] Cifrado, borrado, retención y auditoría.
- [ ] Observabilidad, backups y recuperación probada.
