# Arquitectura de NEXO 0.2

## Principio

NEXO es un organismo de aplicación, no un módulo nuevo dentro de PrismaTEC-Core. La aplicación conserva el vocabulario audiovisual, la lógica de indexación y los invariantes del dominio. PrismaTEC-Core proporciona primitivas de organismo, identidad, memoria, política y eventos mediante AIP.

## Límites

```text
Usuario / navegador
       │ HTTP
       ▼
NEXO Web + API (Go)
       ├── Dominio audiovisual
       ├── Fuentes y segmentos temporales
       ├── Memoria estructurada
       │     ├── Entidades
       │     ├── Afirmaciones
       │     └── Relaciones
       ├── Búsqueda / exportación
       ├── Store local (JSON MVP)
       └── Adaptadores remotos
             ├── PrismaTec-Core AIP v1
             └── Alset MiniNode (Mind / Zyrion / LispAI)
```

## Modelo de dominio

- Project: espacio de conocimiento y raíz de agrupación.
- Asset: fuente audiovisual autorizada.
- Segment: intervalo temporal y texto de evidencia; cada segmento tiene `EvidenceID` estable.
- Entity: persona, organización, lugar, concepto u otro elemento explícito, asociado a evidencias.
- Claim: afirmación textual asociada a evidencias y opcionalmente a entidades.
- Relation: relación dirigida entre dos entidades, asociada a evidencias.
- CoreOrganismID: referencia al organismo técnico creado en PrismaTec-Core.

La fuente original no se copia a PrismaTec-Core. El Core recibe una referencia de dominio mediante memoria. RootCID del Core no se usa como sustituto del ID del medio, de su hash criptográfico ni de una versión editorial.

## Invariantes de procedencia

1. Toda entidad, afirmación y relación debe referenciar al menos un segmento existente del mismo Project.
2. Las relaciones conectan dos entidades distintas del mismo Project.
3. La confianza declarada se limita al intervalo 0..1 y no equivale a una probabilidad calibrada.
4. El origen de esta fase es `manual`; no se simula extracción automática.
5. Las afirmaciones nuevas comienzan con Zyrion `2` (Indeterminado). Estados admitidos: `1=True`, `0=False`, `2=Indeterminado`.
6. Ningún estado cognitivo concede autorización. Las políticas del Core siguen separadas.

## API de memoria estructurada

- `GET /api/projects/{id}/knowledge`: devuelve entidades, afirmaciones y relaciones del espacio.
- `POST /api/projects/{id}/knowledge/entities`: registra una entidad.
- `POST /api/projects/{id}/knowledge/claims`: registra una afirmación.
- `POST /api/projects/{id}/knowledge/relations`: registra una relación.

Las validaciones se realizan en el límite HTTP contra un snapshot del mismo espacio antes de persistir. La persistencia JSON mantiene compatibilidad con bases 0.1: los mapas nuevos se inicializan al abrir una base antigua.

## Búsqueda

0.1/0.2 implementa recuperación léxica sobre segmentos temporales. La memoria estructurada ya conserva relaciones y procedencia, pero todavía no hay consulta de grafo, búsqueda vectorial ni inferencia semántica. No se denomina búsqueda semántica ni inferencia.

## Certeza y razonamiento

Mind puede producir inferencias trazables sobre hechos del dominio; Zyrion clasifica `1=True`, `0=False`, `2=Indeterminado`. El estado 2 activa una nueva profundidad de razonamiento, no una autorización. Toda salida futura debe conservar premisas, reglas y referencias de evidencia. En 0.2 NEXO solo almacena el estado y no invoca automáticamente Mind/Zyrion para evaluar afirmaciones.

## Estado de integración

| Capacidad | Estado en NEXO 0.2 |
|---|---|
| AIP info | Adaptador implementado; pendiente prueba de integración real |
| Crear/iniciar organismo | Adaptador implementado; pendiente prueba de integración real |
| memory.set | Adaptador implementado; pendiente prueba de integración real |
| Mind MiniNode | Proxy opcional; contrato del proveedor debe validarse |
| Zyrion MiniNode | Proxy opcional; contrato del proveedor debe validarse |
| LispAI MiniNode | Proxy opcional; contrato del proveedor debe validarse |
| Entidades/afirmaciones/relaciones | API manual implementada con evidencia obligatoria |
| Extracción automática multimodal | No implementada |
| Consulta vectorial/grafo | No implementada |
| Replicación/recovery | No activada desde NEXO |
| RootCID de contenido audiovisual | No reclamado; el Core RootCID identifica manifiesto del organismo |
