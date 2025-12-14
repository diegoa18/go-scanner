# Documento de Diseño: Orchestrator Engine

## Contexto

`go-scanner` es un framework de reconocimiento de red modular orientado a ciberseguridad. El proyecto debe soportar múltiples tipos de escaneo (TCP Connect, SYN, UDP) con control granular sobre el nivel de intrusión.

## Problema

**Estado Anterior**: Toda la lógica de ejecución vivía en el CLI:

- Decisiones de si ejecutar probes activos
- Filtrado de tipos de probes permitidos
- Coordinación de Scanner → Service Detection → Probing

**Consecuencias**:

- ❌ No escalaba: cada nuevo scanner duplicaría código
- ❌ Riesgo de seguridad: fácil olvidar validaciones de "activo vs pasivo"
- ❌ Testing complejo: lógica mezclada con parsing de flags

## Solución: Patrón Orchestrator

### Componentes

#### 1. `ScanPolicy` (Policy Object)

**Propósito**: Encapsular las reglas de negocio del escaneo.

**Campos**:

- `Timeout`, `Concurrency`: Parámetros técnicos
- `ServiceDetection`: Habilitar detección pasiva de servicios
- `ActiveProbing`: Gate principal para probes activos
- `AllowedProbes`: Whitelist de tipos de probes (seguridad)

**Rationale**:

- Separa "intención del usuario" (CLI) de "lógica de ejecución" (Engine)
- Validable independientemente
- Reutilizable entre diferentes interfaces (CLI, API, GUI)

#### 2. `Engine` (Orchestrator)

**Responsabilidades**:

1. Coordinar el pipeline completo
2. Aplicar la `ScanPolicy` en cada paso
3. Enriquecer resultados sin bloquear el flujo

**Arquitectura**:

```
Scanner Base (interfaz)
    ↓
processResult() [por cada puerto]
    ↓
    ├─ Service Detection (si Policy.ServiceDetection)
    └─ Active Probing (si Policy.ActiveProbing && servicio permitido)
```

**Decisiones de Diseño**:

**a) Concurrencia Interna**:

- Enriquecimiento paralelo (1 gorutina por resultado)
- Evita bloqueo del scanner base
- `WaitGroup` para sincronización limpia

**b) Context Support**:

- Preparado para cancelación (Ctrl+C, timeouts globales)
- Actualmente usa `context.Background()` (mejora futura)

**c) Whitelist Estricta**:

- Si `AllowedProbes` está vacía → **no ejecutar probes** (seguro por defecto)
- El CLI debe configurar explícitamente los tipos permitidos

### Flujo de Ejecución

```
1. CLI parsea flags
2. CLI construye ScanPolicy a partir de flags
3. CLI instancia Scanner Base (ej: TCPConnectScanner)
4. CLI crea Engine(policy, target, ports, scanner)
5. Engine.Run(ctx) retorna canal de resultados
6. Por cada resultado crudo del scanner:
   a. Engine verifica si puerto está abierto
   b. Engine detecta servicio (pasivo)
   c. Engine consulta Probe Registry
   d. Engine verifica si probe está en AllowedProbes
   e. Engine ejecuta probe (si permitido)
   f. Engine enriquece resultado con info del probe
7. CLI consume canal y reporta
```

## Ventajas Arquitectónicas

### Extensibilidad

- Nuevos scanners: solo implementar `scanner.Scanner` interface
- Nuevos probes: registrar en `probe.Registry`
- Nuevas policies: agregar campos a `ScanPolicy` (ej: `StealthMode`)

### Seguridad

- Decisiones centralizadas en un solo lugar (Engine)
- Whitelist explícita (imposible ejecutar probes no autorizados)
- Pasivo por defecto

### Testabilidad

- Engine testeable con mock scanners
- Policy testeable independientemente
- CLI reducido a "glue code" (fácil de mantener)

### Escalabilidad Futura

Preparado para:

- **Múltiples scanners simultáneos** (TCP + UDP en paralelo)
- **Políticas complejas** (ej: "aggressive", "stealth", "custom")
- **Output Pipelines** (JSON, XML, base de datos)
- **Integración con APIs** (RESTful, gRPC)

## Alternativas Consideradas

### ❌ Strategy Pattern en Scanner

- **Problema**: Acoplaría lógica de probing al scanner base
- **Razón de rechazo**: Scanner debe ser "tonto" (solo escanear)

### ❌ Pipeline Fluent API

- **Ejemplo**: `scanner.Scan().Detect().Probe().Report()`
- **Problema**: Complejidad innecesaria para el caso de uso actual
- **Razón de rechazo**: Overengineering

### ✅ Orchestrator Simple (elegido)

- **Razón**: Balance entre simplicidad y extensibilidad
- **Ventaja**: Fácil de entender y mantener

## Consideraciones de Performance

### Cuello de Botella Potencial

- Enriquecimiento paralelo podría saturar con miles de puertos
- **Solución futura**: Worker Pool con límite de gorutinas

### Timeouts

- Probes usan timeout independiente (3s default)
- Evita timeouts heredados del scanner base (que puede ser <100ms)

## Mantenimiento

### Agregar Nuevo Scanner

1. Implementar `scanner.Scanner` interface
2. Registrar en CLI como nuevo subcomando
3. **No modificar Engine** (ya soporta cualquier Scanner)

### Agregar Nueva Policy

1. Agregar campo a `ScanPolicy`
2. Leer en `Engine.processResult()`
3. Actualizar CLI para exponer como flag

## Compatibilidad

✅ **Backward Compatible**: Flags del CLI no cambiaron  
✅ **Output Idéntico**: Report format se mantiene  
✅ **Comportamiento**: Pasivo por defecto, activo opt-in
