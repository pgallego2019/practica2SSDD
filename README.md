# Practica 2 - Taller de Coches en GO 
#### Sistemas Distribuidos - GIT - URJC 2025

Con los mismos requisitos del Taller de coches de la práctica 1, se deberá hacer la implementación usando goroutines y channels de GO. A diferencia de la práctica 1, aquí se podrá usar más de un fichero.

Como hay varios archivos fuente, es necesario inicializar el módulo Go
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

- Mecánico: contiene ID, nombre, especialidad (mecanica, electrica o carroceria), años de experiencia y estado activo (si está trabajando o no).

- Vehículo: incluye información básica y una lista de incidencias asociadas. Ahora también incluye un campo tiempoAcumulado que se corresponde con el tiempoAcumulado de sus incidencias y un campo Prioritario para marcarlo cuando se trabaja en él.

- Incidencia: contiene tipo (mecanica, electrica o carroceria), prioridad, descripción, estado (abierta/en proceso/cerrada) y un nuevo campo TiempoAcumulado para medir el tiempo total de atención según especialidad.

- Trabajo: nueva estructura introducida en simulacion.go, representa una unidad de trabajo que asocia un vehículo y una incidencia a ser procesada por un mecánico concurrentemente.

El **diagrama de clases** representa las nuevas estructuras

![diagrama de clases](https://github.com/pgallego2019/practica2SSDD/blob/main/diagramas/diagramaspractica2ssdd-Diagrama%20de%20clases.drawio.png)

Las relaciones entre clases se explican como:

- Vehículo –> Trabajo: Un vehículo puede ser atendido en múltiples ocasiones (0..*), pero cada trabajo pertenece a un único vehículo (1).

- Incidencia –> Trabajo: Cada incidencia registrada genera exactamente un trabajo (1 a 1), que representa el proceso de atención concurrente.

Aunque en el diagrama de clases no se representa una relación directa entre Mecánico y Trabajo, en la simulación concurrente cada mecánico ejecuta la función trabajoMecanico, procesando diferentes trabajos de forma paralela. Esta relación de tipo “atiende” se modela dinámicamente mediante goroutines y canales en el módulo simulacion.go.

### Funciones principales y funcionamiento de la aplicación
1. _**simularTaller(t *Taller)**_: Inicia la simulación concurrente del taller, que es una opción en el menú principal. Para esto crea dos canales:

- chTrabajos: cola de espera de trabajos (vehículos que llegan).

- chResultados: canal de mensajes para registrar eventos del sistema.

Luego lanza goroutines:

- Una por cada mecánico activo (trabajoMecanico).

- Una para generar vehículos (generadorVehículos).

- Una para imprimir resultados (imprimirResultados).

2. _**verificarAsignacionMecanico(m *Mecanico, v *Vehiculo, inc *Incidencia, chResultados chan string, chTrabajos chan Trabajo,) bool)**_: Función auxiliar de control que determina si un mecánico puede atender una incidencia determinada. Devuelve true si el mecánico puede continuar con la reparación y false si la incidencia debe ser reasignada o atendida por otro mecánico. Su comportamiento se resume así:

- Verificación de estado: Si la incidencia ya está cerrada (Estado == 2), no se procesa.

- Coincidencia de especialidad: Si la especialidad del mecánico coincide con el tipo de incidencia (m.Especialidad == inc.Tipo), el mecánico puede atenderla directamente.

- Prioridad por tiempo acumulado: Si el vehículo supera los 15 segundos de atención total (v.TiempoTotal > 15), la incidencia obtiene prioridad. En este caso, cualquier mecánico disponible puede atenderla, incluso si su especialidad no coincide.

- Reasignación de trabajo: Si la incidencia está en proceso por otro mecánico y aún no ha alcanzado prioridad, se omite el trabajo para evitar duplicidad de procesamiento concurrente.

- Contratación dinámica: Si no hay mecánicos disponibles de la especialidad requerida, la función crea un nuevo mecánico con newMecanico(), lanza su goroutine con iniciarGoroutineMecanico() y registra el evento en chResultados.

- Reenvío de trabajo: Cuando se contrata un nuevo mecánico o se encuentra uno más adecuado, el trabajo se reenvía a la cola chTrabajos mediante reasignarTrabajo() para su futura atención.

Esta función actúa como mecanismo de decisión y equilibrio de carga dentro del sistema concurrente, evitando bloqueos, distribuyendo eficientemente los trabajos y asegurando que las incidencias prioritarias se atiendan con rapidez.

3. _**trabajoMecanico(m *Mecanico, chTrabajos, chResultados, t)**_: Cada mecánico ejecuta esta goroutine de forma independiente. Lee trabajos del canal chTrabajos, revisa si la incidencia está cerrada (para saltar ese trabajo), revisa si el mecánico puede trabajar en esa incidencia  (_verificarAsignacionMecanico_) y si puede trabajar, simula la reparación con time.Sleep según su especialidad y acumula el tiempo en la incidencia y en el vehículo. Si una incidencia supera los 15 segundos acumulados, se le da prioridad. Para esto se intenta buscar un nuevo mecánico libre, pero si no hay se contrata un nuevo mecánico automáticamente y el trabajo se reenvía a la cola de espera.

4. _**generadorVehículos(t, chTrabajos)**_: Genera de forma periódica vehículos nuevos (cada 2 segundos) con distintos tipos de incidencia (mecánica, eléctrica, carrocería) y los envía al canal de trabajos.

5. _**imprimirResultados(chResultados)**_: Goroutine dedicada a mostrar en pantalla los mensajes que van llegando sobre eventos del sistema: inicio y fin de trabajos, reasignaciones, contrataciones, etc.

#### Representación dinámica: Diagrama de secuencia

En el **diagrama de secuencia** mostramos la interacción entre los principales componentes del sistema durante la simulación del taller. Representa cómo se comunican las goroutines a través de los canales (chTrabajos y chResultados) para gestionar concurrentemente los trabajos de reparación.

![diagrama de secuencia](https://github.com/pgallego2019/practica2SSDD/blob/main/diagramas/diagramaspractica2ssdd-Diagrama%20de%20secuencia.drawio.png)

La secuencia se puede explicar como: 
1. Inicio de la simulación: La función simularTaller() crea los canales y lanza las goroutines: una por cada mecánico (trabajoMecanico), una para generar vehículos (generadorVehículos) y otra para imprimir los resultados (imprimirResultados).

2. Generación de vehículos: generadorVehículos crea periódicamente instancias de Vehículo con incidencias aleatorias.
Cada incidencia se encapsula en un objeto Trabajo y se envía al canal chTrabajos.

3. Procesamiento de trabajos: Cada goroutine de trabajoMecanico escucha el canal chTrabajos. Cuando recibe un trabajo, el mecánico simula la reparación con time.Sleep, acumula el tiempo en la incidencia y envía un mensaje de progreso por chResultados.

4. Reasignación o contratación: Si una incidencia supera los 15 segundos de atención acumulada, se marca como prioritaria. trabajoMecanico intenta reasignarla a un mecánico disponible; si no hay, simularTaller crea un nuevo mecánico y reenvía el trabajo al canal.

5. Registro de eventos: La goroutine imprimirResultados escucha continuamente el canal chResultados y muestra los eventos en la consola (inicio y fin de trabajos, reasignaciones, contrataciones, etc.).

6. Finalización: Tras 60 segundos, simularTaller cierra los canales y espera a que todas las goroutines finalicen su ejecución antes de imprimir el resumen.

#### Representación general: Diagrama de flujo

Mientras el diagrama de secuencia muestra las interacciones entre procesos concurrentes, el **diagrama de flujo** ofrece una visión global del funcionamiento de la simulación, desde la inicialización hasta el cierre de la ejecución.

![diagrama de flujo](https://github.com/pgallego2019/practica2SSDD/blob/main/diagramas/diagramaspractica2ssdd-Diagrama%20de%20flujo.drawio.png)

Este diagrama refleja el ciclo completo del sistema:

1. Inicialización de datos y creación de mecánicos.

2. Inicio de la simulación y generación de vehículos.

3. Procesamiento concurrente de incidencias por los mecánicos.

4. Control de prioridades y contratación dinámica.

5. Registro y visualización de resultados.

6. Finalización y cierre controlado de los canales.

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

#### · Duplicación de incidencias por vehículo
- Escenario: 4 vehículos, 2 incidencias por vehículo (duplicación respecto al caso base de 1 incidencia).
- Plantilla: 1 mecánico por especialidad.
Resultados:

´´´
=== RUN   TestSimulacionDuplicarIncidencias
=== Resultados de la simulación ===
-> Mecmecanica terminó carroceria de vehículo V-01
-> Mecmecanica terminó carroceria de vehículo V-02
-> Mecmecanica terminó mecanica de vehículo V-02
-> Mecmecanica terminó electrica de vehículo V-03
-> Mecmecanica terminó electrica de vehículo V-03
-> Mecmecanica terminó mecanica de vehículo V-04
-> Mecmecanica terminó carroceria de vehículo V-04
-> Mecelectrica terminó carroceria de vehículo V-01
Total de incidencias procesadas: 8
Incidencias por mecánico:
  Mecmecanica: 7
  Mecelectrica: 1
--- PASS: TestSimulacionDuplicarIncidencias (0.00s)
´´

Al duplicar el número de incidencias, los mecánicos existentes pudieron atender todas las incidencias, pero algunos mecánicos alcanzaron más carga de trabajo, lo que podría afectar el tiempo total de atención en un escenario real con duraciones simuladas.

#### · Duplicación de plantilla de mecánicos
- Escenario: 4 vehículos, 1 incidencia por vehículo.
- Plantilla: 2 mecánicos por especialidad (6 en total).
Resultados:

´´´
=== RUN   TestSimulacionDuplicarMecanicos
=== Resultados de la simulación ===
-> Mecmecanica terminó carroceria de vehículo V-01
-> Mecmecanica terminó carroceria de vehículo V-03
-> Mecmecanica terminó mecanica de vehículo V-04
-> Mecelectrica terminó carroceria de vehículo V-02
Total de incidencias procesadas: 4
Incidencias por mecánico:
  Mecmecanica: 3
  Mecelectrica: 1
--- PASS: TestSimulacionDuplicarMecanicos (0.00s)
´´´

La duplicación de la plantilla permite repartir mejor la carga, aunque en este escenario con pocas incidencias algunos mecánicos no llegan a atender ninguna incidencia

#### · Distribución desigual de mecánicos
- Caso 1: 3 mecánica, 1 eléctrica, 1 carrocería

- Caso 2: 1 mecánica, 3 eléctrica, 3 carrocería

- Escenario: 4 vehículos, 1 incidencia por vehículo.

```
=== RUN   TestSimulacionDistribucionMecanicos
=== Resultados de la simulación ===
-> Mecmecanica terminó carroceria de vehículo V-01
-> Mecmecanica terminó carroceria de vehículo V-02
-> Mecmecanica terminó carroceria de vehículo V-03
-> Mecmecanica terminó mecanica de vehículo V-04
Total de incidencias procesadas: 4
Incidencias por mecánico:
  Mecmecanica: 4
=== Resultados de la simulación ===
-> Mecelectrica terminó electrica de vehículo V-01
-> Mecelectrica terminó mecanica de vehículo V-03
-> Meccarroceria terminó carroceria de vehículo V-04
-> Mecmecanica terminó electrica de vehículo V-02
Total de incidencias procesadas: 4
Incidencias por mecánico:
  Mecelectrica: 2
  Meccarroceria: 1
  Mecmecanica: 1
--- PASS: TestSimulacionDistribucionMecanicos (0.00s)
```

El balance de mecánicos afecta directamente la eficiencia y la carga por trabajador.

Se observa que distribuir la plantilla según la demanda de especialidades permite optimizar el flujo de trabajo y evitar cuellos de botella.

#### Conclusiones
- Duplicación de incidencias: El sistema puede manejar un aumento de carga limitado, pero la carga de cada mecánico aumenta proporcionalmente.

- Duplicación de plantilla: Incrementar la cantidad de mecánicos reduce la carga individual, mejorando la capacidad de atención simultánea.

- Distribución de especialidades: La asignación estratégica de mecánicos según tipo de incidencia es crucial para evitar sobrecarga y optimizar la eficiencia del taller.

En conjunto, la práctica demuestra cómo la concurrencia en Go permite diseñar sistemas escalables y eficientes mediante el uso coordinado de goroutines y canales, mejorando significativamente la capacidad de respuesta del taller frente al enfoque secuencial de la práctica anterior.

