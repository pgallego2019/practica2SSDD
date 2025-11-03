package main

import (
	"fmt"
	"time"
)

// Definición de una estructura para simular trabajos
type Trabajo struct {
	Vehiculo	*Vehiculo
	Incidencia	*Incidencia
}

// Reenviar un trabajo a la cola
func reasignarTrabajo(chTrabajos chan Trabajo, v *Vehiculo, inc *Incidencia) {
	chTrabajos <- Trabajo{Vehiculo: v, Incidencia: inc}
}

// Goroutine
func trabajoMecanico(m *Mecanico, chTrabajos chan Trabajo, chResultados chan string, t *Taller) {
	for trabajo := range chTrabajos {
		v := trabajo.Vehiculo
		inc := trabajo.Incidencia

		fmt.Printf("Mecánico %s (%s) atendiendo vehículos %s [%s]\n",
			m.Nombre, m.Especialidad, v.Matricula, inc.Tipo)

		// Simulación según la especialidad
		var duracion int
		switch m.Especialidad {
		case "mecanica":
			duracion = 5
		case "electrica":
			duracion = 7
		case "carroceria":
			duracion = 11
		default: // para que no falle pongo un default
			duracion = 5
		}

		time.Sleep(time.Duration(duracion) * time.Second)

		// Acumular tiempo de atención
		inc.Estado = 1 // En proceso
		v.FechaSalida = time.Now().Format("2006-01-02 15:04:05")

		// Acumular tiempo total de atención
		inc.TiempoAcumulado += duracion

		// Si supera los 15 segundos, asignar prioridad
		if inc.TiempoAcumulado > 15 {
			msg := fmt.Sprintf("Vehículo %s acumula %ds — necesita prioridad", v.Matricula, inc.TiempoAcumulado )
			chResultados <- msg

			// Buscar otro mecánico libre
			mecLibre := buscarMecanicoLibre(t, m.Especialidad)
			if mecLibre == nil {
				// Si no hay mecánicos disponibles, se contrata uno nuevo
				mecLibre = t.newMecanico(fmt.Sprintf("Auto-%s", m.Especialidad), m.Especialidad, 1)
				msg2 := fmt.Sprintf("Contratado nuevo mecánico: %s (%s)", mecLibre.Nombre, mecLibre.Especialidad)
				chResultados <- msg2
			}

			// Reasignar el vehículo
			go reasignarTrabajo(chTrabajos, v, inc)
		} else {
			inc.Estado = 2 // Cerrada
			msg := fmt.Sprintf("Mecánico %s terminó vehículo %s (%s) en %ds",
				m.Nombre, v.Matricula, m.Especialidad, duracion)
			chResultados <- msg
		}
	}
}

// auxiliar para buscar un mecánico libre
func buscarMecanicoLibre(t *Taller, especialidad string) *Mecanico {
	for _, m := range t.Mecanicos {
		if m.Activo && m.Especialidad == especialidad {
			return m
		}
	}
	return nil
}

// Goroutine generadora
func generadorVehículos(t *Taller, chTrabajos chan Trabajo) {
	tipos := []string{"mecanica", "electrica", "carroceria"}

	for i := 1; i <= 10; i++ {
		tipo := tipos[i%3]

		v := t.newVehiculo(
			fmt.Sprintf("M-%03d", i),
			"Fiat",
			"500",
			time.Now().Format("2006-01-02 15:04:05"),
			"",
			nil,
		)

		inc := &Incidencia{
			ID:          i,
			Tipo:        tipo,
			Prioridad:   "Alta",
			Descripcion: "Mantenimiento",
			Estado:          0,
			TiempoAcumulado: 0,
		}
		v.Incidencias = append(v.Incidencias, inc)
		t.Incidencias = append(t.Incidencias, inc)

		fmt.Printf("Llega vehículo %s (%s)\n", v.Matricula, tipo)
		chTrabajos <- Trabajo{Vehiculo: v, Incidencia: inc}

		time.Sleep(2 * time.Second) // simulando tiempo entre llegadas
	}
	close(chTrabajos)
}

// Mostrar los resultados que van llegando
func imprimirResultados(chResultados chan string) {
	for msg := range chResultados {
		fmt.Println("->", msg)
	}
}

func simularTaller(t *Taller) {
	fmt.Println("\n=== SIMULACIÓN CONCURRENTE DEL TALLER ===")
	
	// Creo una cola de espera
	chTrabajos := make(chan Trabajo, 20);
	// Creo una cola para imprimir los resultados
	chResultados := make(chan string);
	
	// Lanzar goroutine para mostrar resultados
	go imprimirResultados(chResultados)
	
	// Por si no hay mecánicos
	if len(t.Mecanicos) == 0 {
		fmt.Println("No hay mecánicos activos. Se crean tres de ejemplo.")
		t.newMecanico("Luis", "mecanica", 5)
		t.newMecanico("Ana", "electrica", 4)
		t.newMecanico("Carlos", "carroceria", 6)
	}
	
	// goroutine de mecanicos activos
	for _, m := range t.Mecanicos {
		if m.Activo {
			go trabajoMecanico(m, chTrabajos, chResultados, t)
		}
	}
	
	// Tengo que alimentar la cola generando vehículos
	go generadorVehículos(t, chTrabajos)
	
	// Tengo que esperar a que acaben
	time.Sleep(60 * time.Second)
	close(chResultados)

	fmt.Println("\n=== Fin de la simulación ===")
}
