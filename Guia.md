# Guía de NEXO 0.2

## Propósito

NEXO convierte material audiovisual propio o autorizado en evidencia recuperable. El valor inicial no es resumir vídeos: es localizar un fragmento útil, mostrar su origen y volver al instante exacto sin revisar manualmente horas de grabación.

## Inicio rápido

1. Instalar Go 1.23 o superior.
2. (Opcional) Instalar FFmpeg para transcripción automática.
3. Copiar `.env.example` a `.env` y configurar secretos/proveedores.
4. Ejecutar `go test ./...`.
5. Ejecutar `go run ./cmd/nexo`.
6. Abrir `http://localhost:8090`.
7. Crear un espacio y cargar un vídeo, audio o grabación.
8. Pegar una transcripción temporizada o configurar `NEXO_TRANSCRIPTION_KEY` para transcripción automática.
9. Buscar frases y conceptos; abrir el resultado para reproducir el original en el segundo indicado.

## Formato de transcripción manual

Usar una línea por fragmento:

```text
00:00:12-00:00:18 | La primera afirmación.
00:01:03-00:01:15 | La segunda afirmación.
```

Si se pega texto sin marcas de tiempo, NEXO calcula intervalos aproximados y los etiqueta como `estimated`. No deben citarse como tiempos exactos.

## Funciones por plan

- Free: hasta tres fuentes por espacio, búsqueda literal y navegación a evidencia.
- Pro: exportación CSV y preguntas IA basadas en evidencias recuperadas.
- El cobro se configura mediante un enlace externo. La activación Pro es manual mediante una clave de licencia.

## Integración PrismaTEC-Core

Configurar `NEXO_PRISMATEC_URL` y, si el nodo AIP está protegido, `NEXO_PRISMATEC_TOKEN`. Al crear un espacio NEXO:

1. Se envía `POST /aip/v1/commands` con `action=create`.
2. Se inicia el organismo con `action=start`.
3. Se almacena el `OrganismID` devuelto por el Core.
4. Al cargar un medio, se registra una referencia en memoria con `action=memory.set`.

NEXO no importa paquetes internos del Core. La integración es remota mediante AIP v1. Si el Core no responde, el espacio sigue funcionando localmente y se conserva el estado de error de integración.

## Mind, Zyrion y LispAI

Configurar `NEXO_MININODE_URL` para reenviar peticiones a MiniNode:

- `POST /api/cognition/mind` con `{ "text": "..." }`.
- `POST /api/cognition/zyrion` con `{ "env": {}, "labels": [] }`.
- `POST /api/cognition/lispai` con `{ "cmd": "..." }`.

La interfaz HTTP existe como puente; no se activa una inferencia de negocio automática. La integración cognitiva debe preservar evidencia, reglas y procedencia. Zyrion: `1=True`, `0=False`, `2=Indeterminado`, con el estado 2 como sumidero que abre otro nivel de razonamiento. Ninguna inferencia sustituye a la política de autorización del Core.

## Seguridad y operación

- `NEXO_ACCESS_TOKEN` activa protección por token compartido para `/api/*`.
- La reproducción utiliza enlaces firmados que caducan a los cinco minutos.
- Los medios se guardan en `NEXO_DATA_DIR/media` y la base JSON en `NEXO_DATA_DIR/nexo.json`.
- No desplegar en Internet como SaaS multiusuario: faltan usuarios, sesiones, aislamiento tenant, cuotas, Stripe webhook, revisión de seguridad y recuperación operativa.
- Procesar solo contenido propio o con autorización.

## Pruebas

```bash
go test ./... -count=1
go vet ./...
go build ./cmd/nexo
node --check web/app.js
```

## Memoria estructurada (NEXO 0.2)

Después de indexar una transcripción, se pueden registrar entidades, afirmaciones y relaciones mediante la API. Cada anotación debe apuntar a uno o más `evidence_ids` existentes en el mismo espacio. La consulta agregada es:

```http
GET /api/projects/{project_id}/knowledge
```

Crear entidad:

```http
POST /api/projects/{project_id}/knowledge/entities
Content-Type: application/json
{"name":"Investigadora","type":"person","description":"Participante de la entrevista","evidence_ids":["ev_..."],"confidence":0.95}
```

Crear afirmación:

```http
POST /api/projects/{project_id}/knowledge/claims
Content-Type: application/json
{"text":"La investigadora describe el origen del proyecto.","kind":"statement","entity_ids":["ent_..."],"evidence_ids":["ev_..."],"confidence":0.9}
```

Una afirmación nueva toma `zyrion_state=2` si no se especifica. Los valores admitidos son `1=True`, `0=False` y `2=Indeterminado`. El estado es metadato explícito de la anotación, no una conclusión automática de Mind ni un permiso del Core.

Crear relación:

```http
POST /api/projects/{project_id}/knowledge/relations
Content-Type: application/json
{"subject_entity_id":"ent_...","predicate":"participa_en","object_entity_id":"ent_...","evidence_ids":["ev_..."],"confidence":0.85}
```

La API rechaza evidencias inexistentes o pertenecientes a otro espacio, entidades de otro espacio, relaciones autorreferentes y valores de confianza fuera de 0..1. Esta fase crea una memoria estructurada manual y verificable; todavía no extrae entidades ni relaciones automáticamente.
