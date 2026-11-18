# Arquitectura de NEXO 0.1

## Principio

NEXO es un organismo de aplicación, no un módulo nuevo dentro de PrismaTec-Core. La aplicación conserva el vocabulario audiovisual, la lógica de indexación y los invariantes del dominio. PrismaTec-Core proporciona primitivas de organismo, identidad, memoria, política y eventos mediante AIP.

## Límites

```text
Usuario / navegador
       │ HTTP
       ▼
NEXO Web + API (Go)
       ├── Dominio audiovisual
       ├── Evidencias y segmentos temporales
       ├── Búsqueda / exportación
       ├── Store local (MVP JSON)
       └── Adaptadores
             ├── PrismaTec-Core AIP v1
             └── Alset MiniNode (Mind / Zyrion / LispAI)
```

## Modelo mínimo

- Project: espacio de conocimiento y raíz de agrupación.
- Asset: fuente audiovisual autorizada.
- Segment: intervalo temporal y texto de evidencia.
- EvidenceID: identidad estable del fragmento dentro de NEXO.
- CoreOrganismID: referencia al organismo técnico creado en PrismaTec-Core.

La fuente original no se copia a PrismaTec-Core. El Core recibe una referencia de dominio mediante memoria. RootCID del Core no se usa como sustituto del ID del medio, de su hash criptográfico ni de una versión editorial.

## Búsqueda

0.1 implementa recuperación léxica sobre segmentos temporales. No se denomina búsqueda semántica ni inferencia. La siguiente iteración añadirá un proveedor de embeddings/LLM con resultados obligatoriamente anclados a EvidenceID y timestamps.

## Certeza y razonamiento

Mind puede producir inferencias trazables sobre hechos del dominio; Zyrion clasifica `1=True`, `0=False`, `2=Indeterminado`. El estado 2 activa una nueva profundidad de razonamiento, no una autorización. Toda salida debe conservar premisas, reglas y referencias de evidencia. La política del Core permanece separada.

## Estado de integración

| Capacidad | Estado en NEXO 0.1 |
|---|---|
| AIP info | Adaptador implementado; pendiente prueba de integración real |
| Crear/iniciar organismo | Adaptador implementado; pendiente prueba de integración real |
| memory.set | Adaptador implementado; pendiente prueba de integración real |
| Mind MiniNode | Proxy opcional; contrato del proveedor debe validarse |
| Zyrion MiniNode | Proxy opcional; contrato del proveedor debe validarse |
| LispAI MiniNode | Proxy opcional; contrato del proveedor debe validarse |
| Replicación/recovery | No activada desde NEXO |
| RootCID de contenido audiovisual | No reclamado; el Core RootCID identifica manifiesto del organismo |
