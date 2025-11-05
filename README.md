# Practica 2 - Taller de Coches en GO 
#### Sistemas Distribuidos - GIT - URJC 2025

Con los mismos requisitos del Taller de coches de la práctica 1, se deberá hacer la implementación
usando goroutines y channels de GO. A diferencia de la práctica 1, aquí se podrá usar más de un
fichero.

Como hay varios archivos fuente, necesito inicializar el módulo Go
```
paula@840g3:~/SSDD/practica2SSDD$ go mod init practica2SSDD
paula@840g3:~/SSDD/practica2SSDD$ go run .
```
Para ejecutar los test:

```
paula@840g3:~/SSDD/practica2SSDD$ go test -v
```

## Explicación del diseño

### Estructuras de datos
El sistema mantiene las estructuras principales de la práctica anterior, con algunas modificaciones para incluir control de tiempo y concurrencia:

- Taller: estructura principal que agrupa listas de clientes, vehículos, mecánicos, incidencias y plazas de trabajo. No tiene modificaciones respecto a la práctica anterior.

- Mecánico: contiene ID, nombre, especialidad (mecanica, electrica o carroceria), años de experiencia y estado activo.

- Vehículo: incluye información básica y una lista de incidencias asociadas. No tiene modificaciones respecto a la práctica anterior.

- Incidencia: contiene tipo (mecanica, electrica o carroceria), prioridad, descripción, estado (abierta/en proceso/cerrada) y un nuevo campo TiempoAcumulado para medir el tiempo total de atención.

- Trabajo: nueva estructura introducida en simulacion.go, representa una unidad de trabajo que asocia un vehículo y una incidencia a ser procesada por un mecánico concurrentemente.

El **diagrama de clases** representa las nuevas estructuras


Vehículo – Trabajo: un vehículo puede ser atendido en múltiples ocasiones (0..*), pero cada trabajo pertenece a un único vehículo (1).

Incidencia – Trabajo: cada incidencia registrada genera exactamente un trabajo (1 a 1), que representa el proceso de atención concurrente.

Aunque en el diagrama de clases no se representa una relación directa entre Mecánico y Trabajo, en la simulación concurrente cada mecánico ejecuta la función trabajoMecanico, procesando diferentes trabajos de forma paralela. Esta relación de tipo “atiende” se modela dinámicamente mediante goroutines y canales en el módulo simulacion.go.

### Funciones principales
1. simularTaller(t *Taller)

Inicia la simulación concurrente del taller, que es una opción en el menú principal.
Para esto crea dos canales:

- chTrabajos: cola de espera de trabajos (vehículos que llegan).

- chResultados: canal de mensajes para registrar eventos del sistema.

Luego lanza goroutines:

- Una por cada mecánico activo (trabajoMecanico).

- Una para generar vehículos (generadorVehículos).

- Una para imprimir resultados (imprimirResultados).

Por último, espera 60 segundos antes de cerrar la simulación para que acaben todos los trabajos.

2. trabajoMecanico(m *Mecanico, chTrabajos, chResultados, t)

Cada mecánico ejecuta esta goroutine de forma independiente.
Lee trabajos del canal chTrabajos, simula la reparación con time.Sleep según su especialidad y acumula el tiempo en la incidencia.

Si una incidencia supera los 15 segundos acumulados, se le da prioridad. Para esto se intenta buscar un nuevo mecánico libre, pero si no hay se contrata un nuevo mecánico automáticamente y el trabajo se reenvía a la cola de espera.

3. generadorVehículos(t, chTrabajos)

Genera de forma periódica vehículos nuevos (cada 2 segundos) con distintos tipos de incidencia (mecánica, eléctrica, carrocería) y los envía al canal de trabajos.

4. imprimirResultados(chResultados)

Goroutine dedicada a mostrar en pantalla los mensajes que van llegando sobre eventos del sistema: inicio y fin de trabajos, reasignaciones, contrataciones, etc.

### Funcionamiento de la aplicación
Cuando desde el menú principal se elige la opción de simular taller, se ejecuta lo siguiente:

1. Si no hay mecánicos creados, se crean tres por defecto (uno de cada especialidad)

2. Los vehículos llegan de forma simulada y se insertan en una cola de espera.

3. Cada mecánico toma un trabajo de la cola, y lo procesa durante el tiempo necesario según la especialidad.

4. Si un vehículo acumula más de 15 segundos de atención, se le da prioridad, se busca uno libre (aunque no sea de la especialidad del trabajo) o se contrata un mecánico nuevo (este sí de la especialidad de trabajo) y se reasigna el trabajo.

5. Tras 60 segundos de simulación, los canales se cieran y se imprimen los resultados finales.

El **diagrama de flujo** representa el funcionamiento de la simulación.

## Test realizados
El objetivo de esta sección es analizar el comportamiento del taller bajo distintas condiciones de carga y distribución de mecánicos, simulando el procesamiento de incidencias de vehículos de manera concurrente. Se comparan tres escenarios principales:

	1. Duplicación del número de incidencias por vehículo.

	2. Duplicación de la plantilla de mecánicos.

	3. Distribución desigual de mecánicos según especialidad.

Creamos un nuevo archivo simulacion_test.go para los test. Para evitar los test concurrentes no usamos time.Sleep ni cerramos los canales en medio de la simulación.

Además, del código propuesto como ejemplo en el enunciado (capítulo 8 de "The Go Programming Language") sacamos:

1. No depender de las impresiones en la shell, sino devolver resultados estructurados para poder testearlos.

2. Se crean varios casos (p.ej. plantillas de mecánicos) y se comprueba el resultado esperado usando Errorf

3. Fijar la semilla del generador de números aleatorios para que la simulación sea más determinista durante los test y así usar el mismo patrón de incidencias (reproducibilidad)

Las funciones para ajustar los tests al enunciado son:

- TestSimulacionDuplicarIncidencias: Se generan 4 vehículos con 2 incidencias cada uno, frente a un escenario base de una incidencia. Validamos que se hayan procesado 8 incidencias.

- TestSimulacionDuplicarMecanicos: Se crean 6 mecánicos (2 de cada especialidad), frente a un escenario con 3 mecánicos. Validamos que todas las incidencias (una por vehículo) se procesan.

- TestSimulacionDistribucionMecanicos: Hay dos casos
	- Caso 1: 3 mecánicos de mecánica, 1 mecánico de eléctrica, 1 mecánico de carrocería.
	- Caso 2: 1 mecánico de mecánica, 3 mecánicos de eléctrica, 3 mecánicos de carrocería.
	Aquí se usan distintas distribuciones de especiales para ver el número de incidencias procesadas. Se valida que ninguna distribución deja trabajos sin completar.
	
### Métricas obtenidas y análisis
Las métricas registradas son:
- Número total de incidencias procesadas.

- Incidencias atendidas por cada mecánico.

- Distribución de incidencias por tipo.

#### Escenarios

**Duplicación de incidencias por vehículo**
- Escenario: 4 vehículos, 2 incidencias por vehículo (duplicación respecto al caso base de 1 incidencia).
- Plantilla: 1 mecánico por especialidad.
Resultados:

| Mecánico | Incidencias atendidas |
|:-------- |:--------:|
| MecMecanica   | 3   |
| MecElectrica   | 3   |
| MecCarroceria   | 2   |
| Total   | 8   |

Al duplicar el número de incidencias, los mecánicos existentes pudieron atender todas las incidencias, pero algunos mecánicos alcanzaron más carga de trabajo, lo que podría afectar el tiempo total de atención en un escenario real con duraciones simuladas.

**Duplicación de plantilla de mecánicos**
- Escenario: 4 vehículos, 1 incidencia por vehículo.
- Plantilla: 2 mecánicos por especialidad (6 en total).
Resultados:

| Mecánico | Incidencias atendidas |
|:-------- |:--------:|
| MecMecanica1   | 1   |
| MecMecanica2  | 0   |
| MecElectrica1   | 1   |
| MecElectrica2   | 0  |
| MecCarroceria1   | 1   |
| MecCarroceria2   | 0   |
| Total   | 3  |

La duplicación de la plantilla permite repartir mejor la carga, aunque en este escenario con pocas incidencias algunos mecánicos no llegan a atender ninguna incidencia

**Distribución desigual de mecánicos**
- Caso 1: 3 mecánica, 1 eléctrica, 1 carrocería

- Caso 2: 1 mecánica, 3 eléctrica, 3 carrocería

- Escenario: 4 vehículos, 1 incidencia por vehículo.

|Distribución | Total incidencias procesadas |Observaciones |
|:-------- |:--------:|:--------:|
| Caso 1   | 4   | Mayor carga para mecánicos de mecánica, eléctrica y carrocería menos cargados. |
| Caso 2  | 4   | Mayor distribución en eléctrica y carrocería, mecánico de mecánica sobrecargado. |

El balance de mecánicos afecta directamente la eficiencia y la carga por trabajador.

Se observa que distribuir la plantilla según la demanda de especialidades permite optimizar el flujo de trabajo y evitar cuellos de botella.

#### Conclusiones
- Duplicación de incidencias: El sistema puede manejar un aumento de carga limitado, pero la carga de cada mecánico aumenta proporcionalmente.

- Duplicación de plantilla: Incrementar la cantidad de mecánicos reduce la carga individual, mejorando la capacidad de atención simultánea.

- Distribución de especialidades: La asignación estratégica de mecánicos según tipo de incidencia es crucial para evitar sobrecarga y optimizar la eficiencia del taller.
