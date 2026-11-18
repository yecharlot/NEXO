# NEXO

**Memoria audiovisual con evidencia temporal.** NEXO transforma vídeos, grabaciones, clases, entrevistas y otros materiales autorizados en fuentes consultables: cada fragmento conserva su procedencia y su posición temporal.

> Estado: MVP 0.1 funcional en fase de validación. No es todavía un SaaS multiusuario ni un sistema de IA autónomo. La primera entrega prioriza ingestión, transcripciones/subtítulos, búsqueda transversal y recuperación exacta de evidencias.

## Qué puede hacer esta versión

- Crear espacios de conocimiento.
- Cargar vídeo o audio (MP4, MOV, WEBM, MKV, MP3, WAV, M4A, OGG; máximo 512 MB).
- Transcripción automática opcional mediante FFmpeg + proveedor compatible con OpenAI; o adjuntar una transcripción con tiempos `HH:MM:SS-HH:MM:SS | texto`, o texto simple.
- Importar subtítulos SRT mediante API.
- Buscar términos en todas las evidencias de un espacio y volver al segundo exacto.
- Preguntar a NEXO (Pro) y obtener síntesis IA ancladas a segmentos recuperados, cuando se configure un proveedor LLM.
- Exportar CSV de evidencias con licencia Pro.
- Limitar el plan Free a tres fuentes por espacio y activar Pro mediante clave.
- Crear un organismo en PrismaTEC-Core por espacio y registrar referencias de medios en su memoria, si se configura AIP.
- Exponer adaptadores opcionales para Mind, Zyrion y LispAI del MiniNode Alset.

## Ejecutar

Requisitos: Go 1.23+ (para integración directa con la rama actual de PrismaTec-Core se recomienda Go 1.25+), navegador moderno.

```bash
go test ./...
go run ./cmd/nexo
```

Abrir `http://localhost:8090`.

## Configuración

```bash
export NEXO_ADDR=:8090
export NEXO_DATA_DIR=./data
export NEXO_WEB_DIR=./web

# Protección básica del API (recomendado en cualquier servidor accesible por red)
export NEXO_ACCESS_TOKEN='cambia-esto-por-un-secreto-largo'

# Enlace de cobro externo (Stripe Payment Link u otro proveedor)
export NEXO_PAYMENT_URL='https://buy.stripe.com/tu-enlace'
# Clave de licencia emitida manualmente después del pago
export NEXO_PRO_KEY='NEXO-PRO-clave-aleatoria'

# Transcripción automática (FFmpeg + endpoint compatible con OpenAI Audio Transcriptions)
export NEXO_TRANSCRIPTION_KEY='tu-api-key'
export NEXO_TRANSCRIPTION_MODEL='whisper-1'
# opcional: export NEXO_TRANSCRIPTION_URL='https://api.openai.com/v1/audio/transcriptions'

# Respuestas IA basadas en evidencias (endpoint compatible con OpenAI Chat Completions)
export NEXO_LLM_API_KEY='tu-api-key'
export NEXO_LLM_MODEL='gpt-4o-mini'
# opcional: export NEXO_LLM_URL='https://api.openai.com/v1/chat/completions'

# Integración remota con PrismaTec-Core AIP v1
export NEXO_PRISMATEC_URL='http://127.0.0.1:8080'
export NEXO_PRISMATEC_TOKEN='mismo-token-configurado-en-PrismaTec-Core'

# Opcional: MiniNode Alset/LISPAI Runtime para Mind, Zyrion y LispAI
export NEXO_MININODE_URL='http://127.0.0.1:5177'
```

No expongas NEXO a Internet sin HTTPS, autenticación por usuario, límites de carga, aislamiento de archivos y una revisión de seguridad. `NEXO_ACCESS_TOKEN` es una protección de despliegue de una sola clave, no un sistema multiusuario.

## Flujo de transcripción

En 0.1, la transcripción puede pegarse al cargar un medio o enviarse después:

```http
POST /api/assets/{asset_id}/transcript
Content-Type: application/json
{"text":"00:00:12-00:00:19 | Primera idea importante.\n00:01:03-00:01:15 | Segunda idea."}
```

También se acepta SRT:

```http
POST /api/assets/{asset_id}/captions
Content-Type: text/plain
```

La transcripción automática está disponible de forma opcional mediante FFmpeg y un endpoint compatible con OpenAI Audio Transcriptions. Los tiempos de transcripciones manuales sin marcas temporales se etiquetan como estimados; no se presentan como evidencia temporal exacta. La comprensión visual multimodal y la extracción automática de entidades/acontecimientos todavía no están implementadas.

## Integración con PrismaTec-Core

La integración es **remota por AIP v1**, no importa paquetes internos de PrismaTec-Core. Se apoya en el contrato público documentado en `docs/APPLICATION_INTEGRATION_CONTRACT.md` y en las rutas confirmadas en `docs/USAGE.md`:

- `GET /aip/v1/info`
- `POST /aip/v1/commands` con acciones `create`, `start` y `memory.set`

Un espacio NEXO intenta crear un organismo y guardar referencias a los medios en su memoria. Si el Core no está disponible, NEXO continúa en modo local y lo informa; no simula una integración exitosa.

## Mind, Zyrion y LispAI

Adaptador opcional hacia MiniNode Alset:

- `POST /api/cognition/mind` → `/api/mind/tick`
- `POST /api/cognition/zyrion` → `/api/zyrion`
- `POST /api/cognition/lispai` → `/api/lispai`

La salida cognitiva no concede autorización ni cambia evidencia fuente. Zyrion debe respetar el contrato acordado: `1=True`, `0=False`, `2=Indeterminado` (estado sumidero que activa otro nivel de razonamiento). No se activa inferencia automática hasta tener contratos y pruebas de cada proveedor.

## Monetización inicial

La versión inicial permite una venta piloto con:

- Free: tres fuentes por espacio y búsqueda con evidencia temporal.
- Pro: fuentes sin límite práctico del MVP y exportación CSV.
- Cobro: enlace externo configurado en `NEXO_PAYMENT_URL`.
- Activación: clave de licencia emitida después del pago y validada por `NEXO_PRO_KEY`.

Esto permite vender pilotos de forma manual, pero **no equivale a facturación SaaS automatizada**. Antes de abrir autoservicio se necesita autenticación multiusuario, Stripe Checkout + webhook verificado, gestión de suscripciones, cuotas por usuario y aislamiento de datos.

## Estructura

```text
cmd/nexo/                 Arranque HTTP
internal/domain/          Entidades y contratos de dominio
internal/store/           Persistencia JSON atómica (MVP)
internal/search/           Recuperación textual con referencias
internal/httpapi/          API + límites de acceso + exportación
internal/integration/      Adaptadores PrismaTec AIP y MiniNode
web/                      Interfaz web
 docs/                    Arquitectura, estado y roadmap
```

## Pruebas

```bash
go test ./...
```

## Licencia y contenido

NEXO almacena medios cargados por el usuario. El operador debe tener derechos o autorización para procesarlos. No redistribuir materiales protegidos de terceros. La exportación contiene metadatos y texto aportado/indexado; la licencia de uso del contenido original sigue perteneciendo a sus titulares.
