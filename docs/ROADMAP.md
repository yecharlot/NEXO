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

## Fase 0.3 — Extracción asistida multimodal
- [ ] Worker de procesamiento con extracción de audio y keyframes.
- [ ] OCR y visión por fotogramas con modelos multimodales.
- [ ] Propuestas automáticas de entidades, acontecimientos y afirmaciones.
- [ ] Revisión humana obligatoria antes de consolidar conocimiento.
- [ ] Confianza, procedencia del modelo y trazas de procesamiento.

## Fase 0.4 — Memoria transversal y razonamiento
- [ ] Comparación de ideas entre fuentes.
- [ ] Consulta de grafo y búsqueda semántica/vectorial.
- [ ] Detección de contradicciones sin colapsar incertidumbre.
- [ ] Mind/Zyrion con pruebas de reglas y profundidad fractal controlada.
- [ ] LispAI para reglas declarativas versionadas y sandboxed.

## Fase 0.5 — SaaS público
- [ ] Usuarios, organizaciones, roles y aislamiento multi-tenant.
- [ ] Stripe Checkout + webhook firmado + portal de suscripciones.
- [ ] Cuotas, límites de almacenamiento y procesamiento.
- [ ] Almacenamiento de objetos y base de datos transaccional.
- [ ] Cifrado, borrado, retención y auditoría.
- [ ] Observabilidad, backups y recuperación probada.
