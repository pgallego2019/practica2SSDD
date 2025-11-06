package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Definición de una estructura para simular trabajos
type Trabajo struct {
	Vehiculo   *Vehiculo
	Incidencia *Incidencia
}

// Reenviar un trabajo a la cola
func reasignarTrabajo(chTrabajos chan Trabajo, v *Vehiculo, inc *Incidencia) {
	chTrabajos <- Trabajo{Vehiculo: v, Incidencia: inc}
}

// Inicia la goroutine de trabajo para un mecánico recién creado
func iniciarGoroutineMecanico(m *Mecanico, chTrabajos chan Trabajo, chResultados chan string, t *Taller) {
	if m == nil {
		return
	}
	if !m.Activo {
		return
	}
	go trabajoMecanico(m, chTrabajos, chResultados, t)
}

// Verifica si el mecánico puede atender la incidencia.
// Devuelve true si puede continuar, false si se reasigna o contrata otro mecánico.
func (t *Taller) verificarAsignacionMecanico(
	m *Mecanico,
	v *Vehiculo,
	inc *Incidencia,
	chResultados chan string,
	chTrabajos chan Trabajo,
) bool {

	// Si la incidencia ya está cerrada, no se procesa
	if inc.Estado == 2 {
		return false
	}

	// Si la especialidad coincide, puede atender sin problema
	if inc.Tipo == m.Especialidad {
		return true
	}

	// Si el vehículo ya tiene prioridad (>15s), puede atenderlo aunque no coincida
	t.updateTiempoTotalVehiculo(v)
	if v.TiempoTotal > 15 {
		return true
	}

	// Si ya está en proceso por otro mecánico y no hay prioridad, no reasignamos
	if inc.Estado == 1 && v.TiempoTotal <= 15 {
		return false
	}

	// Buscar otro mecánico disponible de la especialidad requerida
	var mecAdecuado *Mecanico
	for _, candidato := range t.Mecanicos {
		if candidato.Activo && candidato.Especialidad == inc.Tipo {
			mecAdecuado = candidato
			break
		}
	}

	// Si no hay mecánico disponible de esa especialidad, contratar uno nuevo y arrancarle la goroutine
	if mecAdecuado == nil {
		mecAdecuado = t.newMecanico(
			fmt.Sprintf("Auto-%s", inc.Tipo),
			string(inc.Tipo),
			1,
		)
		// lanzar su goroutine para que pueda procesar trabajos
		iniciarGoroutineMecanico(mecAdecuado, chTrabajos, chResultados, t)

		chResultados <- fmt.Sprintf("Contratado nuevo mecánico %s (%s) para incidencia %s",
			mecAdecuado.Nombre, mecAdecuado.Especialidad, inc.Tipo)
	}

	// Reasignar el trabajo a la cola UNA vez (si todavía no está cerrado)
	if inc.Estado != 2 {
		reasignarTrabajo(chTrabajos, v, inc)
	}
	return false
}

// Goroutine
func trabajoMecanico(m *Mecanico, chTrabajos chan Trabajo, chResultados chan string, t *Taller) {
	for trabajo := range chTrabajos {
		v := trabajo.Vehiculo
		inc := trabajo.Incidencia

		// Verificar si este mecánico puede atender la incidencia o si la incidencia ya está cerrada, saltarla
		if !t.verificarAsignacionMecanico(m, v, inc, chResultados, chTrabajos) {
			continue
		}

		m.Activo = false

		fmt.Printf("Mecánico %s (%s) atendiendo vehículos %s [%s]\n",
			m.Nombre, m.Especialidad, v.Matricula, inc.Tipo)

		// Simulación según la especialidad
		var duracion int
		switch inc.Tipo {
		case Mecanica:
			duracion = 5
		case Electrica:
			duracion = 7
		case Carroceria:
			duracion = 11
		default:
			duracion = 5
		}

		time.Sleep(time.Duration(duracion) * time.Second)

		// Acumular tiempo de atención
		inc.TiempoAcumulado += duracion
		t.updateTiempoTotalVehiculo(v)
		//v.FechaSalida = time.Now().Format("2006-01-02 15:04:05")

		// Libero al mecánico
		m.Activo = true

		// Si tras el trabajo la suma total supera 15, intentamos añadir un mecánico extra (prioridad)
		if v.TiempoTotal > 15 {
			// Aviso (no lo hacemos infinitas veces porque v.Prioritario ya se puso true)
			chResultados <- fmt.Sprintf("Vehículo %s acumula %ds — necesita prioridad", v.Matricula, v.TiempoTotal)

			// Buscar otro mecánico libre (de cualquier especialidad diferente al actual)
			var mecLibre *Mecanico
			for _, candidato := range t.Mecanicos {
				if candidato.Activo && candidato.ID != m.ID {
					mecLibre = candidato
					break
				}
			}

			// Si no hay otro, contratar y arrancar su goroutine
			if mecLibre == nil {
				mecLibre = t.newMecanico(fmt.Sprintf("Auto-%s", inc.Tipo), string(inc.Tipo), 1)
				iniciarGoroutineMecanico(mecLibre, chTrabajos, chResultados, t)
				chResultados <- fmt.Sprintf("Contratado nuevo mecánico: %s (%s)", mecLibre.Nombre, mecLibre.Especialidad)
			}

			// Reasignar UNA vez para que otro mecánico (u el nuevo) coja la incidencia y ayude
			if inc.Estado != 2 {
				reasignarTrabajo(chTrabajos, v, inc)
			}

			// No cerramos la incidencia aquí (pues aún quedan trabajos por hacer)
			continue
		}
		// Si no necesita prioridad, marcar como cerrada
		inc.Estado = 2
		msg3 := fmt.Sprintf("Mecánico %s terminó incidencia del vehículo %s (%s) en %ds [Total %ds]",
			m.Nombre, v.Matricula, inc.Tipo, duracion, v.TiempoTotal)
		chResultados <- msg3
	}
}

// Goroutine generadora
func generadorVehículos(t *Taller, chTrabajos chan Trabajo) {
	tipos := []Especialidad{Mecanica, Electrica, Carroceria}

	for i := 1; i <= 4; i++ {
		v := t.newVehiculo(
			fmt.Sprintf("M-%03d", i),
			"Fiat",
			"500",
			time.Now().Format("2006-01-02 15:04:05"),
			"",
			nil,
		)

		// Cada vehículo tendrá entre 1 y 3 incidencias
		numInc := rand.Intn(3) + 1

		for j := 0; j < numInc; j++ {
			tipo := tipos[rand.Intn(len(tipos))]
			inc := &Incidencia{
				ID:              len(t.Incidencias) + 1,
				Tipo:            tipo,
				Prioridad:       "Alta",
				Descripcion:     fmt.Sprintf("Mantenimiento %s", tipo),
				Estado:          0,
				TiempoAcumulado: 0,
			}

			v.Incidencias = append(v.Incidencias, inc)
			t.Incidencias = append(t.Incidencias, inc)

			// Enviar cada incidencia como trabajo
			fmt.Printf("Llega vehículo %s con incidencia %s\n", v.Matricula, tipo)
			chTrabajos <- Trabajo{Vehiculo: v, Incidencia: inc}
		}

		time.Sleep(2 * time.Second) // simulando tiempo entre llegadas
	}
}

// Mostrar los resultados que van llegando
func imprimirResultados(chResultados chan string) {
	for msg := range chResultados {
		fmt.Println("->", msg)
	}
}

func simularTaller(t *Taller) {
	fmt.Println("\n=== SIMULACIÓN CONCURRENTE DEL TALLER ===")

	chTrabajos := make(chan Trabajo, 20)
	chResultados := make(chan string, 50)

	go imprimirResultados(chResultados)

	if len(t.Mecanicos) == 0 {
		fmt.Println("No hay mecánicos activos. Se crean tres de ejemplo.")
		t.newMecanico("Luis", "mecanica", 5)
		t.newMecanico("Ana", "electrica", 4)
		t.newMecanico("Carlos", "carroceria", 6)
	}

	for _, m := range t.Mecanicos {
		if m.Activo {
			go trabajoMecanico(m, chTrabajos, chResultados, t)
		}
	}

	go func() {
		generadorVehículos(t, chTrabajos)
		//close(chTrabajos)
	}()

	fmt.Println("(Simulando... espera unos segundos)")
	time.Sleep(20 * time.Second)

	close(chResultados)
	fmt.Println("\n=== Fin de la simulación ===")
}
