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
// Devuelve true si puede continuar, false si se debe reasignar o contrata otro mecánico.
func (t *Taller) verificarAsignacionMecanico(
	m *Mecanico,
	v *Vehiculo,
	inc *Incidencia,
) bool {
	if inc.Estado == 1 {
		return false
	}
	// Si la especialidad coincide o el vehículo es prioritario, puede atenderlo
	if inc.Tipo == m.Especialidad || v.Prioritario {
		return true
	}
	return false
}

// Goroutine de cada mecánico
func trabajoMecanico(m *Mecanico, chTrabajos chan Trabajo, chResultados chan string, t *Taller) {
	for trabajo := range chTrabajos {
		v := trabajo.Vehiculo
		inc := trabajo.Incidencia

		// Si la incidencia ya está cerrada, saltarla
		if inc.Estado == 2 {
			continue
		}

		// Verificar si este mecánico puede atender la incidencia
		if !t.verificarAsignacionMecanico(m, v, inc) {
			// Buscar otro mecánico disponible de la especialidad correcta
			var mecLibre *Mecanico
			for _, candidato := range t.Mecanicos {
				if candidato.Activo && candidato.Especialidad == inc.Tipo && candidato.ID != m.ID {
					mecLibre = candidato
					break
				}
			}

			// Si no hay, contratamos uno nuevo
			if mecLibre == nil {
				mecLibre = t.newMecanico(fmt.Sprintf("Auto-%s", inc.Tipo), string(inc.Tipo), 1)
				iniciarGoroutineMecanico(mecLibre, chTrabajos, chResultados, t)
				chResultados <- fmt.Sprintf("No había mecánicos disponibles (%s) — contratado nuevo: %s",
					inc.Tipo, mecLibre.Nombre)
			}

			reasignarTrabajo(chTrabajos, v, inc)
			continue
		}

		m.Activo = false
		inc.Estado = 1

		fmt.Printf("Mecánico %s (%s) atendiendo vehículos %s [%s]\n", m.Nombre, m.Especialidad, v.Matricula, inc.Tipo)

		duracion := inc.TiempoAcumulado
		time.Sleep(time.Duration(duracion) * time.Second)

		inc.Estado = 2
		v.TiempoTotal += duracion
		t.updateTiempoTotalVehiculo(v)
		m.Activo = true

		// Reportar resultado final
		if v.TiempoTotal == 0 {
			chResultados <- fmt.Sprintf(
				"Mecánico %s terminó incidencia del vehículo %s (%s) en %ds.\nEl vehículo %s está reparado",
				m.Nombre, v.Matricula, inc.Tipo, duracion, v.Matricula)
			t.liberarPlaza(v)
		} else {
			chResultados <- fmt.Sprintf(
				"Mecánico %s terminó incidencia del vehículo %s (%s) en %ds [Tiempo restante del vehículo %ds]",
				m.Nombre, v.Matricula, inc.Tipo, duracion, v.TiempoTotal)
		}
	}
}

// Goroutine generadora de vehículos e incidencias para alimentar el canal de trabajos
func generadorVehículos(t *Taller, chTrabajos chan Trabajo, nvehiculos int) {
	tipos := []Especialidad{Mecanica, Electrica, Carroceria}

	for i := 1; i <= nvehiculos; i++ {
		v := t.newVehiculo(
			fmt.Sprintf("M-%03d", i),
			"Fiat",
			"500",
			time.Now().Format("2006-01-02 15:04:05"),
			"",
			nil,
		)

		// Buscar plaza libre
		plazaLibre := -1
		for i, p := range t.Plazas {
			if !p.Ocupada {
				plazaLibre = i
				break
			}
		}

		if plazaLibre == -1 {
			fmt.Printf("Vehículo %s rechazado: no hay plazas disponibles (%d/%d)\n",
				v.Matricula, len(t.plazasOcupadas()), len(t.Plazas))
			continue
		}

		// Ocupar la plaza
		t.Plazas[plazaLibre].Ocupada = true
		t.Plazas[plazaLibre].VehiculoMat = v.Matricula
		fmt.Printf("Vehículo %s ocupa plaza %d (%d/%d ocupadas)\n",
			v.Matricula, t.Plazas[plazaLibre].ID, len(t.plazasOcupadas()), len(t.Plazas))

		// Cada vehículo tendrá entre 1 y 3 incidencias
		numInc := rand.Intn(3) + 1

		for j := 0; j < numInc; j++ {
			tipo := tipos[rand.Intn(len(tipos))]

			inc, err := t.newIncidencia(
				v.Matricula,
				nil,
				string(tipo),
				"Alta",
				fmt.Sprintf("Mantenimiento %s", tipo),
			)
			if err != nil {
				fmt.Println("Error creando incidencia:", err)
				continue
			}

			fmt.Printf("Llega vehículo %s con incidencia %s (tiempo estimado %d s)\n",
				v.Matricula, inc.Tipo, inc.TiempoAcumulado)

			chTrabajos <- Trabajo{Vehiculo: v, Incidencia: inc}
		}

		t.updateTiempoTotalVehiculo(v)
		fmt.Printf("El vehículo %s necesitará %d segundos en total\n", v.Matricula, v.TiempoTotal)
		if v.Prioritario {
			fmt.Printf("El vehículo %s tiene prioridad\n", v.Matricula)
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

//NOTA: no cierro los canales para evitar panic writing on closed channel

// Función principal de simulación concurrente
func simularTaller(t *Taller) {
	fmt.Println("\n=== SIMULACIÓN CONCURRENTE DEL TALLER ===")

	var numVehiculos int
	fmt.Print("Introduce el número de vehículos a generar: ")
	_, err := fmt.Scan(&numVehiculos)
	if err != nil || numVehiculos <= 0 {
		numVehiculos = 5 // valor por defecto
		fmt.Println("Entrada inválida, se generarán 5 vehículos por defecto.")
	}

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
		generadorVehículos(t, chTrabajos, numVehiculos)
		//close(chTrabajos)
	}()

	fmt.Println("(Simulando... espera unos segundos)")
	time.Sleep(60 * time.Second)

	//close(chResultados)
	fmt.Println("\n=== Fin de la simulación ===")
}
