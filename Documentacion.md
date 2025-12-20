# Documentación Técnica: Custom Cache Engine

## 📋 Tabla de Contenidos

1. [Alcance del Proyecto](#alcance-del-proyecto)
2. [Requerimientos Funcionales](#requerimientos-funcionales)
3. [Requerimientos Técnicos](#requerimientos-técnicos)
4. [Estructura del Proyecto](#estructura-del-proyecto)
5. [Referencia de Código](#referencia-de-código)

---

## Alcance del Proyecto

### Almacenamiento Clave-Valor

**¿Qué es un almacenamiento clave-valor?**
Imagina un diccionario: buscas una palabra (la "clave") y encuentras su definición (el "valor"). Nuestro cache funciona igual: guardas datos con un nombre único y luego puedes recuperarlos usando ese nombre.

**Archivo:** `internal/cache/cache.go`

```go
// Línea 8-13: Estructura de una entrada en el cache
// Esto define CÓMO se guarda cada dato en el cache

type CacheEntry struct {
    Value      interface{} // El dato que guardamos (puede ser texto, número, etc.)
    ExpiresAt  int64       // Cuándo expira (como la fecha de vencimiento de un producto)
    LastAccess int64       // Cuándo fue la última vez que alguien usó este dato (para LRU)
}

// Línea 17: El "almacén" principal donde guardamos todo
// Es un mapa: cada clave (string) apunta a una entrada (CacheEntry)

data map[string]*CacheEntry // Como un diccionario: palabra → definición
```

**¿Por qué `interface{}`?** 
En Go, `interface{}` significa "cualquier tipo de dato". Es como tener una caja que puede guardar cualquier cosa: texto, números, listas, etc. Esto hace que nuestro cache sea flexible, igual que Redis que puede guardar diferentes tipos de datos.

**Ejemplo en la vida real:**
- Clave: `"usuario:123"` → Valor: `"Juan Pérez"` (texto)
- Clave: `"contador"` → Valor: `42` (número)
- Clave: `"activo"` → Valor: `true` (verdadero/falso)

---

### Expiración

**¿Qué es la expiración?**
Es como la fecha de vencimiento de un producto en el supermercado. Después de cierto tiempo, el dato "vence" y se elimina automáticamente del cache. Esto es útil para datos temporales como sesiones de usuario o tokens.

**Archivo:** `internal/cache/cache.go`

```go
// Línea 110-131: Método EXPIRE
// Este método le pone una "fecha de vencimiento" a un dato

func (c *CacheEngine) Expire(key string, seconds int) bool {
    // "key" es el nombre del dato
    // "seconds" es cuántos segundos debe vivir el dato
    
    expiresAt := time.Now().Unix() + int64(seconds)
    // ↑ Calculamos: hora actual + segundos = fecha de vencimiento
    // Ejemplo: si son las 3:00pm y seconds=60, expira a las 3:01pm
    
    entry.ExpiresAt = expiresAt
    // ↑ Guardamos esa fecha en el dato
}
```

**¿Cómo funciona la expiración?**
El sistema usa DOS métodos para eliminar datos expirados:

1. **Expiración Pasiva (Lazy):** Cuando alguien pide un dato con GET, revisamos si ya venció. Si venció, lo eliminamos y decimos "no existe".

2. **Barrido Activo (Active Sweep):** Cada cierto tiempo, un proceso revisa TODOS los datos y elimina los vencidos (aunque nadie los haya pedido).

---

### Persistencia

**¿Qué es la persistencia?**
Es guardar los datos en un archivo en el disco duro. Si apagas el programa, los datos no se pierden porque están guardados en el archivo. Cuando vuelvas a iniciar, puedes cargarlos.

**Archivo:** `internal/persistence/persistence.go`

```go
// Línea 24-50: Función para guardar una operación en el archivo
// Cada vez que haces SET, DEL o EXPIRE, se guarda en el archivo

func LogOperation(filename, operation, key string, value interface{}, expiresAt int64) error {
    
    // Paso 1: Abrir el archivo en modo "agregar al final"
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    // ↑ os.O_APPEND significa "agregar al final, no borrar lo que hay"
    // ↑ os.O_CREATE significa "si el archivo no existe, créalo"
    // ↑ Es como escribir en un cuaderno: siempre en la siguiente línea, nunca borras
    
    // Paso 2: Convertir los datos a formato JSON
    encoder := json.NewEncoder(file)
    // ↑ JSON es un formato de texto que las computadoras entienden fácilmente
    // ↑ Es como traducir los datos a un idioma universal
    
    // Paso 3: Escribir en el archivo
    encoder.Encode(logEntry)
    // ↑ Esto escribe los datos en el archivo
}
```

**¿Por qué "append-only"?**
"Append-only" significa "solo agregar". Nunca borramos ni modificamos lo que ya escribimos en el archivo. Es como un libro de contabilidad: solo agregas líneas, nunca las tachas. Esto tiene ventajas:
- Si algo falla, no pierdes datos anteriores
- Puedes ver el historial completo de operaciones
- Es como Redis AOF (Append-Only File)

---

### Estrategias de Reemplazo (LRU)

**¿Qué es LRU?**
LRU significa "Least Recently Used" (Menos Recientemente Usado). Imagina que tienes un estante con espacio para solo 5 libros. Cuando quieres agregar un libro nuevo pero el estante está lleno, ¿cuál quitas? LRU dice: "quita el libro que hace más tiempo que nadie tocó".

**Archivo:** `internal/cache/cache.go`

```go
// Línea 134-151: Algoritmo LRU
// Esta función decide QUÉ dato eliminar cuando el cache está lleno

func (c *CacheEngine) evictLRU() {
    var oldestKey string          // Aquí guardaremos el nombre del dato más viejo
    var oldestTime int64 = time.Now().UnixNano()  // Hora actual como referencia

    // Paso 1: Revisar TODOS los datos buscando el más viejo
    for key, entry := range c.data {
        // ↑ "for" es como decir "para cada dato en el cache, haz esto:"
        
        if entry.LastAccess < oldestTime {
            // ↑ Si este dato fue accedido ANTES que el más viejo que encontramos...
            oldestTime = entry.LastAccess  // ...este se convierte en el más viejo
            oldestKey = key                 // ...y guardamos su nombre
        }
    }

    // Paso 2: Eliminar el dato más viejo
    if oldestKey != "" {
        delete(c.data, oldestKey)
        // ↑ "delete" elimina el dato del cache
    }
}
```

**¿Cómo sabe el sistema cuándo fue el último acceso?**
Cada vez que alguien usa un dato (con GET), actualizamos su "LastAccess":

```go
// Línea 86: Esto ocurre CADA VEZ que alguien pide un dato
entry.LastAccess = time.Now().UnixNano()
// ↑ Es como poner un sello de "última vez que lo usaron"
```

---

## Requerimientos Funcionales

### 1. Operaciones Obligatorias: SET, GET, DEL, EXPIRE

Estas son las 4 operaciones básicas que todo cache debe tener:

| Operación | ¿Qué hace? | Ejemplo |
|-----------|------------|---------|
| **SET** | Guarda un dato | `SET nombre "Juan"` |
| **GET** | Recupera un dato | `GET nombre` → `"Juan"` |
| **DEL** | Elimina un dato | `DEL nombre` |
| **EXPIRE** | Pone fecha de vencimiento | `EXPIRE nombre 60` (vence en 60 segundos) |

**Archivo:** `internal/cache/cache.go`

#### SET (Guardar un dato)
```go
// Línea 42-66
func (c *CacheEngine) Set(key string, value interface{}) {
    // "key" = nombre del dato (ej: "usuario:123")
    // "value" = el dato a guardar (ej: "Juan Pérez")
    
    c.mu.Lock()  
    // ↑ "Lock" significa "cerrar con llave"
    // ↑ Es como poner un cartel de "no molestar" mientras trabajamos
    // ↑ Esto evita que dos procesos modifiquen el cache al mismo tiempo
    
    // Verificar si tenemos espacio
    if len(c.data) >= c.maxEntries {
        // ↑ Si ya tenemos el máximo de datos permitidos...
        c.evictLRU()  // ...eliminamos el menos usado
    }

    now := time.Now().UnixNano()
    // ↑ Obtenemos la hora exacta actual (en nanosegundos)
    
    c.data[key] = &CacheEntry{
        Value:      value,      // El dato que queremos guardar
        ExpiresAt:  0,          // 0 significa "nunca expira"
        LastAccess: now,        // Marcamos que lo usamos AHORA
    }
    // ↑ Creamos una nueva entrada y la guardamos en el mapa
    
    c.mu.Unlock()
    // ↑ "Unlock" = quitar el cartel de "no molestar"
}
```

#### GET (Recuperar un dato)
```go
// Línea 68-88
func (c *CacheEngine) Get(key string) (interface{}, bool) {
    // Retorna DOS cosas:
    // 1. El valor (si existe)
    // 2. true/false indicando si lo encontró
    
    c.mu.Lock()
    defer c.mu.Unlock()
    // ↑ "defer" significa "haz esto CUANDO la función termine"
    // ↑ Es un truco para no olvidar hacer Unlock
    
    entry, exists := c.data[key]
    // ↑ Buscamos el dato en el mapa
    // ↑ "exists" será true si lo encontró, false si no
    
    if !exists {
        return nil, false  // No existe, retornamos "nada" y "no encontrado"
    }
    
    // Verificar si expiró (fecha de vencimiento)
    now := time.Now().Unix()
    if entry.ExpiresAt > 0 && entry.ExpiresAt <= now {
        // ↑ Si tiene fecha de expiración Y ya pasó esa fecha...
        delete(c.data, key)  // ...lo eliminamos
        return nil, false     // ...y decimos que no existe
    }
    
    // Actualizar "último acceso" para LRU
    entry.LastAccess = time.Now().UnixNano()
    // ↑ Es como decir "alguien usó este dato justo ahora"
    
    return entry.Value, true  // Retornamos el valor y "sí encontrado"
}
```

---

### 2. Implementación LRU

**Explicación simple de LRU:**

Imagina que tienes una lista de 5 amigos que puedes invitar a tu fiesta (el cache tiene límite). Cuando llega un nuevo amigo pero la lista está llena, ¿a quién quitas? LRU dice: "al que hace más tiempo que no te habla".

```go
// Línea 134-151
func (c *CacheEngine) evictLRU() {
    var oldestKey string  
    // ↑ Variable para guardar el nombre del "amigo que hace más tiempo no te habla"
    
    var oldestTime int64 = time.Now().UnixNano()  
    // ↑ Empezamos con la hora actual como referencia

    // Revisamos a todos los "amigos" (datos en el cache)
    for key, entry := range c.data {
        if entry.LastAccess < oldestTime {
            // ↑ "Si este amigo te habló ANTES que el más antiguo que encontraste..."
            oldestTime = entry.LastAccess  // ↑ ...ahora ESTE es el más antiguo
            oldestKey = key                 // ↑ ...y guardamos su nombre
        }
    }

    // Eliminar al "amigo" que hace más tiempo no te habla
    if oldestKey != "" {
        delete(c.data, oldestKey)
    }
}
```

---

### 3. Persistencia Append-Only Log

**Explicación simple:**
Es como llevar un diario. Cada acción que haces (guardar, borrar, etc.) se escribe en el diario. Si mañana quieres recordar qué hiciste, lees el diario desde el principio.

**Archivo:** `internal/persistence/persistence.go`

```go
// Línea 11-18: La estructura de cada "línea del diario"
type LogEntry struct {
    Operation string      `json:"operation"`  // ¿Qué hiciste? (SET, DEL, EXPIRE)
    Key       string      `json:"key"`        // ¿A qué dato?
    Value     interface{} `json:"value,omitempty"`      // ¿Qué valor? (solo para SET)
    ExpiresAt int64       `json:"expires_at,omitempty"` // ¿Cuándo expira?
    Timestamp int64       `json:"timestamp"`  // ¿A qué hora lo hiciste?
}
// Los `json:"..."` son etiquetas que dicen cómo se llama cada campo en el archivo JSON
```

**Ejemplo de cómo se ve el archivo de log:**
```json
{"operation":"SET","key":"usuario:1","value":"Juan","timestamp":1766111011}
{"operation":"SET","key":"usuario:2","value":"Maria","timestamp":1766111029}
{"operation":"DEL","key":"usuario:1","timestamp":1766111050}
```
- Línea 1: A las 17:30 guardamos "Juan" con el nombre "usuario:1"
- Línea 2: A las 17:31 guardamos "Maria" con el nombre "usuario:2"
- Línea 3: A las 17:32 borramos "usuario:1"

---

### 4. API CLI

**¿Qué es CLI?**
CLI significa "Command Line Interface" (Interfaz de Línea de Comandos). Es la pantalla negra donde escribes comandos, como cuando usas `cmd` en Windows.

**Archivo:** `internal/api/cli/cli.go`

```go
// Línea 14-166: La función principal que ejecuta el CLI
func Run(cacheEngine *cache.CacheEngine) {
    reader := bufio.NewReader(os.Stdin)
    // ↑ Esto nos permite leer lo que el usuario escribe en el teclado
    
    // Mostramos el menú de ayuda
    fmt.Println("=== Custom Cache Engine CLI ===")
    fmt.Println("Comandos disponibles:")
    fmt.Println("  SET <key> <value>  - Guardar un dato")
    // ... más comandos ...
    
    // Bucle infinito: esperamos comandos del usuario
    for {
        // ↑ "for" sin condición = "repite esto para siempre"
        
        fmt.Print("cache> ")  // Mostramos el prompt
        input, _ := reader.ReadString('\n')  // Leemos lo que escribe el usuario
        
        // Separamos el comando de sus argumentos
        parts := strings.Fields(input)
        // ↑ "strings.Fields" separa por espacios
        // ↑ Ejemplo: "SET nombre Juan" → ["SET", "nombre", "Juan"]
        
        command := strings.ToUpper(parts[0])
        // ↑ Convertimos a mayúsculas para que "set" y "SET" funcionen igual
        
        // Ejecutamos el comando según lo que escribió
        switch command {
            case "SET":
                // ... código para SET ...
            case "GET":
                // ... código para GET ...
            // ... etc ...
        }
    }
}
```

**Ejemplo de uso en la terminal:**
```
cache> SET usuario:1 "Juan Perez"
OK
cache> GET usuario:1
Juan Perez
cache> EXPIRE usuario:1 60
OK
cache> STATS
Entradas en cache: 1
Límite máximo: 1000
cache> EXIT
Cerrando cache engine...
```

---

### 5. Barrido Periódico de Claves Expiradas

**Explicación simple:**
Imagina un supermercado. Hay dos formas de quitar productos vencidos:
1. **Cuando un cliente intenta comprarlo:** El cajero revisa la fecha y si venció, lo quita.
2. **Cada noche, un empleado revisa TODO el supermercado:** Busca productos vencidos y los quita, aunque nadie los haya intentado comprar.

Nuestro cache hace ambas cosas.

**¿Qué es un Goroutine?**
Un goroutine es como un "trabajador invisible" que hace tareas en segundo plano. Mientras tú usas el cache normalmente, hay un goroutine que cada segundo revisa si hay datos vencidos. Es como tener un empleado que trabaja sin que lo veas.

**Archivo:** `internal/cache/cache.go`

```go
// Línea 37: Cuando creamos el cache, iniciamos el "trabajador invisible"
go cache.periodicCleanup()
// ↑ "go" es la palabra mágica que crea un goroutine
// ↑ Es como decir "haz esto en segundo plano, no me esperes"
// ↑ El programa continúa inmediatamente, sin esperar a que termine

// Línea 153-166: Esto es lo que hace el "trabajador invisible"
func (c *CacheEngine) periodicCleanup() {
    ticker := time.NewTicker(1 * time.Second)
    // ↑ Creamos un "despertador" que suena cada 1 segundo
    
    defer ticker.Stop()
    // ↑ Cuando esta función termine, apagar el despertador
    
    for {
        // ↑ Bucle infinito: el trabajador nunca para (hasta que le digamos)
        
        select {
        // ↑ "select" es como esperar a que pase una de varias cosas
        
        case <-ticker.C:
            // ↑ "Se activó el despertador" (pasó 1 segundo)
            c.cleanExpired()  // Limpiamos los datos vencidos
            
        case <-c.stopClean:
            // ↑ "Alguien nos dijo que paremos" (el cache se está cerrando)
            return  // Terminamos la función
        }
    }
}

// Línea 168-179: La limpieza de datos vencidos
func (c *CacheEngine) cleanExpired() {
    c.mu.Lock()
    defer c.mu.Unlock()
    // ↑ Ponemos el cartel de "no molestar" mientras limpiamos
    
    now := time.Now().Unix()
    // ↑ ¿Qué hora es ahora?
    
    for key, entry := range c.data {
        // ↑ Para cada dato en el cache...
        
        if entry.ExpiresAt > 0 && entry.ExpiresAt <= now {
            // ↑ Si tiene fecha de vencimiento Y ya pasó esa fecha...
            delete(c.data, key)  // ↑ ...lo eliminamos
        }
    }
}
```

**¿Cómo paramos al "trabajador invisible"?**
```go
// Línea 20: Canal para comunicarnos con el goroutine
stopClean chan bool
// ↑ Un "canal" es como un walkie-talkie para hablar entre goroutines

// Línea 193-196: Cuando cerramos el cache
func (c *CacheEngine) Close() {
    close(c.stopClean)
    // ↑ "close" cierra el canal
    // ↑ Es como decirle al trabajador "ya puedes irte a casa"
}
```

---

## Requerimientos Técnicos

### 1. Implementación con RWMutex

**¿Qué es un Mutex?**
Imagina un baño público con UN solo cubículo. Si alguien está adentro, los demás deben esperar. Un Mutex es igual: cuando alguien está modificando el cache, los demás deben esperar.

**¿Qué es RWMutex?**
Es un Mutex más inteligente. Permite que VARIOS lean al mismo tiempo (como varias personas viendo un cuadro en un museo), pero solo UNO puede escribir (como el pintor que está pintando).

**Archivo:** `internal/cache/cache.go`

```go
// Línea 18: Declaramos el "candado"
mu sync.RWMutex
// ↑ "mu" es nuestro candado
// ↑ "RWMutex" significa "Read-Write Mutex" (Mutex de Lectura-Escritura)

// Cuando vamos a ESCRIBIR (modificar datos):
func (c *CacheEngine) Set(key string, value interface{}) {
    c.mu.Lock()     // ← "Me encierro con llave, nadie más puede entrar"
    // ... hacemos cambios ...
    c.mu.Unlock()   // ← "Ya terminé, otros pueden entrar"
}

// Cuando solo vamos a LEER (sin modificar):
func (c *CacheEngine) Size() int {
    c.mu.RLock()    // ← "Entro a leer, otros lectores pueden entrar también"
    defer c.mu.RUnlock()  // ← "Cuando termine, aviso que ya no estoy leyendo"
    return len(c.data)
}
```

**¿Por qué es importante?**
Sin Mutex, pueden pasar cosas raras. Imagina:
- Proceso A: "Voy a guardar el valor 5 en 'contador'"
- Proceso B: "Voy a guardar el valor 10 en 'contador'"
Si ambos escriben al mismo tiempo, puede quedar 5, 10, o incluso un valor corrupto. Con Mutex, uno espera al otro.

---

### 2. Módulo de Expiración Concurrente

**Explicación simple:**
"Concurrente" significa "al mismo tiempo". Mientras tú usas el cache (guardando y leyendo datos), hay otro proceso corriendo en paralelo que se encarga de limpiar los datos vencidos. Ambos trabajan al mismo tiempo sin estorbarse.

```go
// Línea 37: Iniciamos el "limpiador" en paralelo
go cache.periodicCleanup()
// ↑ "go" crea un goroutine (proceso paralelo)
// ↑ Es como contratar a un empleado que limpia mientras otros trabajan

// Línea 20: Creamos un "walkie-talkie" para comunicarnos con el limpiador
stopClean chan bool
// ↑ "chan" significa "channel" (canal)
// ↑ Los canales permiten que los goroutines se comuniquen

// Cuando queremos que el limpiador pare:
func (c *CacheEngine) Close() {
    close(c.stopClean)
    // ↑ Cerramos el canal = le decimos "ya para"
}
```

**Ciclo de vida:**
1. `NewCacheEngine()` → Crea el cache e inicia el goroutine limpiador
2. El goroutine corre en segundo plano, limpiando cada segundo
3. `Close()` → Envía señal de "para" al goroutine
4. El goroutine termina limpiamente

---

### 3. Serialización JSON

**¿Qué es serialización?**
Es convertir datos de la memoria de la computadora a un formato que puedas guardar en un archivo o enviar por internet. "Deserialización" es lo opuesto: leer el archivo y convertirlo de vuelta a datos en memoria.

**¿Qué es JSON?**
JSON (JavaScript Object Notation) es un formato de texto muy popular. Es fácil de leer para humanos y computadoras.

**Archivo:** `internal/persistence/persistence.go`

```go
// Línea 5: Importamos la librería de JSON
import "encoding/json"
// ↑ Go viene con esta librería incluida

// Para ESCRIBIR JSON (guardar):
encoder := json.NewEncoder(file)
// ↑ Creamos un "escritor" de JSON

encoder.Encode(logEntry)
// ↑ Convertimos el dato a JSON y lo escribimos en el archivo
// ↑ logEntry = {"operation":"SET","key":"usuario"...}

// Para LEER JSON (cargar):
decoder := json.NewDecoder(file)
// ↑ Creamos un "lector" de JSON

decoder.Decode(&logEntry)
// ↑ Leemos el JSON del archivo y lo convertimos a un dato en memoria
// ↑ El "&" significa "pon el resultado aquí"
```

**Ejemplo:**
```
Dato en memoria:
  Operation: "SET"
  Key: "usuario:1"
  Value: "Juan"

Después de serializar (JSON):
  {"operation":"SET","key":"usuario:1","value":"Juan"}
```

---

### 4. Benchmarks

**¿Qué son benchmarks?**
Son pruebas de rendimiento. Miden qué tan rápido funciona el código. Es como cronometrar a un corredor para ver en cuánto tiempo hace 100 metros.

**Archivo:** `internal/cache/benchmark_test.go`

```go
// Línea 8-18: Benchmark para SET
func BenchmarkSet(b *testing.B) {
    // ↑ "b *testing.B" es el "cronómetro" que Go nos da
    
    cache := NewCacheEngine(10000)  // Creamos un cache
    defer cache.Close()              // Al terminar, lo cerramos
    
    b.ResetTimer()
    // ↑ Reiniciamos el cronómetro (no contamos el tiempo de crear el cache)
    
    for i := 0; i < b.N; i++ {
        // ↑ "b.N" es un número que Go decide
        // ↑ Go repite esto muchas veces para medir con precisión
        
        key := fmt.Sprintf("key%d", i)
        // ↑ Creamos una clave única: "key0", "key1", "key2"...
        
        cache.Set(key, i)
        // ↑ Guardamos el dato (esto es lo que estamos midiendo)
    }
}
```

**Cómo ejecutar los benchmarks:**
```bash
cd internal/cache
go test -bench=. -benchmem
```

**Cómo leer los resultados:**
```
BenchmarkSet-12    1000000    85323 ns/op    64 B/op    3 allocs/op
```
- `BenchmarkSet-12`: Nombre del test, usando 12 procesadores
- `1000000`: Se ejecutó 1 millón de veces
- `85323 ns/op`: Cada operación tardó ~85 nanosegundos
- `64 B/op`: Cada operación usó 64 bytes de memoria
- `3 allocs/op`: Cada operación hizo 3 asignaciones de memoria

---

### 5. Pruebas para LRU, Expiración y Persistencia

**¿Qué son las pruebas unitarias?**
Son pequeños tests que verifican que cada parte del código funciona correctamente. Es como revisar que cada ingrediente esté bueno antes de cocinar.

**Archivo:** `internal/cache/cache_test.go`

#### Prueba de LRU
```go
// Línea 79-122
func TestLRUEviction(t *testing.T) {
    // ↑ "t *testing.T" es el "juez" que decide si la prueba pasó o falló
    
    cache := NewCacheEngine(3)  // Cache con espacio para solo 3 datos
    defer cache.Close()
    
    // Guardamos 3 datos (el cache se llena)
    cache.Set("key1", "value1")
    time.Sleep(100 * time.Millisecond)  // Esperamos un poquito
    cache.Set("key2", "value2")
    time.Sleep(100 * time.Millisecond)
    cache.Set("key3", "value3")
    time.Sleep(100 * time.Millisecond)
    
    // Accedemos a key1 y key2 (las "refrescamos")
    cache.Get("key1")
    time.Sleep(100 * time.Millisecond)
    cache.Get("key2")
    time.Sleep(100 * time.Millisecond)
    
    // Ahora key3 es la menos usada
    
    // Agregamos key4 (el cache está lleno, debe eliminar algo)
    cache.Set("key4", "value4")
    
    // Verificamos que key3 fue eliminada (era la menos usada)
    _, exists := cache.Get("key3")
    if exists {
        t.Error("key3 debería haber sido expulsada por LRU")
        // ↑ Si key3 todavía existe, la prueba FALLA
    }
    
    // Verificamos que las demás siguen existiendo
    _, exists = cache.Get("key1")
    if !exists {
        t.Error("key1 debería existir")
    }
}
```

#### Prueba de Expiración
```go
// Línea 55-77
func TestExpire(t *testing.T) {
    cache := NewCacheEngine(10)
    defer cache.Close()
    
    cache.Set("key1", "value1")        // Guardamos un dato
    cache.Expire("key1", 1)            // Le ponemos 1 segundo de vida
    
    // Verificamos que existe ANTES de que expire
    _, exists := cache.Get("key1")
    if !exists {
        t.Error("La clave debería existir antes de expirar")
    }
    
    time.Sleep(2 * time.Second)        // Esperamos 2 segundos
    
    // Verificamos que YA NO existe (expiró)
    _, exists = cache.Get("key1")
    if exists {
        t.Error("La clave debería haber expirado")
    }
}
```

**Cómo ejecutar las pruebas:**
```bash
go test -v ./internal/cache/
```

**Resultado esperado:**
```
=== RUN   TestSetAndGet
--- PASS: TestSetAndGet (0.00s)
=== RUN   TestExpire
--- PASS: TestExpire (2.00s)
=== RUN   TestLRUEviction
--- PASS: TestLRUEviction (0.50s)
...
PASS
```

---

## Estructura del Proyecto

```
cache-engine/
├── cmd/
│   └── cache-engine/
│       └── main.go              # Punto de entrada (donde empieza el programa)
├── internal/
│   ├── cache/
│   │   ├── cache.go             # El corazón: LRU, expiración, mutex
│   │   ├── cache_test.go        # Pruebas (¿funciona bien?)
│   │   └── benchmark_test.go    # Rendimiento (¿es rápido?)
│   ├── persistence/
│   │   └── persistence.go       # Guardar/cargar a disco
│   └── api/
│       └── cli/
│           └── cli.go           # Interfaz para el usuario
├── go.mod                       # Configuración del proyecto
└── README.md                    # Documentación general
```

---

## Referencia de Código

### Resumen de Archivos

| Archivo | Líneas | ¿Qué hace? |
|---------|--------|------------|
| `cache.go` | 240 | El motor: guarda datos, LRU, expiración |
| `persistence.go` | 194 | Guardar y cargar datos del disco |
| `cli.go` | 173 | La terminal donde escribes comandos |
| `cache_test.go` | 203 | Pruebas para verificar que todo funciona |
| `benchmark_test.go` | 149 | Medir qué tan rápido es |

### Requisitos Cumplidos ✅

| Requisito | ✅ | ¿Dónde está? |
|-----------|---|--------------|
| SET, GET, DEL, EXPIRE | ✅ | `cache.go` líneas 42-132 |
| LRU | ✅ | `cache.go` líneas 134-151 |
| Append-only log | ✅ | `persistence.go` líneas 24-50 |
| API CLI | ✅ | `cli.go` |
| Barrido periódico | ✅ | `cache.go` líneas 153-179 |
| RWMutex | ✅ | `cache.go` línea 18 |
| Expiración concurrente | ✅ | `cache.go` líneas 37, 153-179 (goroutine) |
| Serialización JSON | ✅ | `persistence.go` líneas 44-46 |
| Benchmarks | ✅ | `benchmark_test.go` (8 tests de rendimiento) |
| Tests LRU/Expiración | ✅ | `cache_test.go` (8 pruebas unitarias) |

---

## Glosario de Términos

| Término | Significado Simple |
|---------|---------------------|
| **Cache** | Memoria rápida para guardar datos temporalmente |
| **Mutex** | Candado que evita que dos procesos modifiquen algo al mismo tiempo |
| **Goroutine** | Trabajador invisible que hace tareas en segundo plano |
| **Channel** | Walkie-talkie para que los goroutines se comuniquen |
| **LRU** | Estrategia que elimina lo que menos se usa |
| **TTL** | Tiempo de vida de un dato antes de expirar |
| **JSON** | Formato de texto para guardar datos |
| **Append-only** | Solo agregar al final, nunca borrar |
| **CLI** | Terminal donde escribes comandos |
| **Benchmark** | Prueba de velocidad |
