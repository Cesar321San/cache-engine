
# Características 

Operaciones básicas : SET, GET, DEL, EXPIRE  
Estrategia LRU: Least Recently Used para reemplazo de elementos  
Persistencia: Append-only log y snapshots en formato JSON  
APIs: CLI interactiva  
Concurrencia: Thread-safe con `sync.RWMutex`  
Auto-limpieza: Barrido periódico de claves expiradas  

# Instalación


cd cache-engine
go mod tidy


# Modo CLI (Interactivo)


go run ./cmd/cache-engine

# O Si se quiere establecer una limite de elementos 

go run ./cmd/cache-engine -max=3


# Comandos disponibles:

SET <key> <value>    - Establecer clave-valor
GET <key>            - Obtener valor
DEL <key>            - Eliminar clave
EXPIRE <key> <secs>  - Establecer expiración
SAVE [archivo]       - Guardar a log
LOAD [archivo]       - Cargar desde log
SNAPSHOT [archivo]   - Guardar snapshot
STATS                - Mostrar estadísticas
EXIT                 - Salir


# Ejemplo de sesión CLI:

cache> SET usuario:123 "Juan Pérez"
OK
cache> GET usuario:123
Juan Pérez
cache> EXPIRE usuario:123 60
OK
cache> STATS
Entradas en cache: 1
Límite máximo: 1000



