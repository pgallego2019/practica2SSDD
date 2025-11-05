// simulacion_test.go
package main

import (
	"math/rand"
	"testing"
	"time"
)

// ---------- HELPERS ----------

// Crear taller con mecánicos según un mapa de especialidades
func crearTallerConMecanicos(nombre string, plantilla map[Especialidad]int) *Taller {
	t := &Taller{}
	for esp, cantidad := range plantilla {
		for i := 0; i < cantidad; i++ {
			t.newMecanico(nombre+string(esp), string(esp), 1)
		}
	}
	return t
}

// Generar vehículos con incidencias controladas
func generarVehiculosParaTest(t *Taller, numVehiculos int, numIncidencias int, tipos []Especialidad, r *rand.Rand) []*Vehiculo {
	var vehiculos []*Vehiculo
	for i := 1; i <= numVehiculos; i++ {
		v := t.newVehiculo(
			string(rune('A'+i-1)), // Matrícula simple A, B, C...
			"Marca",
			"Modelo",
			time.Now().Format("2006-01-02 15:04:05"),
			"",
			nil,
		)
		for j := 0; j < numIncidencias; j++ {
			tipo := tipos[r.Intn(len(tipos))]
			inc := &Incidencia{
				ID:              len(t.Incidencias) + 1,
				Tipo:            tipo,
				Prioridad:       "Alta",
				Descripcion:     "Mantenimiento",
				Estado:          0,
				TiempoAcumulado: 0,
			}
			v.Incidencias = append(v.Incidencias, inc)
			t.Incidencias = append(t.Incidencias, inc)
		}
		vehiculos = append(vehiculos, v)
	}
	return vehiculos
}

// Simulación “controlada” para test sin sleeps ni prints
func simularTallerTest(t *Taller, vehiculos []*Vehiculo) []string {
	var resultados []string
	chTrabajos := make(chan Trabajo, 100)
	chResultados := make(chan string, 100)

	done := make(chan struct{})

	// lanzar mecánicos
	for _, m := range t.Mecanicos {
		go func(m *Mecanico) {
			for trabajo := range chTrabajos {
				//v := trabajo.Vehiculo
				inc := trabajo.Incidencia
				if inc.Estado == 2 {
					continue
				}
				inc.Estado = 2
				chResultados <- m.Nombre + " terminó " + string(inc.Tipo)
			}
			done <- struct{}{}
		}(m)
	}

	// enviar trabajos
	go func() {
		for _, v := range vehiculos {
			for _, inc := range v.Incidencias {
				chTrabajos <- Trabajo{Vehiculo: v, Incidencia: inc}
			}
		}
		close(chTrabajos)
	}()

	// esperar a que todos los mecánicos terminen
	for range t.Mecanicos {
		<-done
	}
	close(chResultados)

	// recoger resultados
	for msg := range chResultados {
		resultados = append(resultados, msg)
	}

	return resultados
}

// ---------- TESTS ----------

func TestSimulacionDuplicarIncidencias(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	taller := crearTallerConMecanicos("Mec", map[Especialidad]int{
		Mecanica:   1,
		Electrica:  1,
		Carroceria: 1,
	})

	tipos := []Especialidad{Mecanica, Electrica, Carroceria}
	vehiculos := generarVehiculosParaTest(taller, 4, 2, tipos, r) // duplicamos incidencias

	resultados := simularTallerTest(taller, vehiculos)

	if len(resultados) != 4*2 {
		t.Errorf("Se esperaban %d resultados, se obtuvieron %d", 4*2, len(resultados))
	}
}

func TestSimulacionDuplicarMecanicos(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	taller := crearTallerConMecanicos("Mec", map[Especialidad]int{
		Mecanica:   2,
		Electrica:  2,
		Carroceria: 2,
	})

	tipos := []Especialidad{Mecanica, Electrica, Carroceria}
	vehiculos := generarVehiculosParaTest(taller, 4, 1, tipos, r)

	resultados := simularTallerTest(taller, vehiculos)

	if len(resultados) != 4*1 {
		t.Errorf("Se esperaban %d resultados, se obtuvieron %d", 4*1, len(resultados))
	}
}

func TestSimulacionDistribucionMecanicos(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	tipos := []Especialidad{Mecanica, Electrica, Carroceria}

	// Caso 1: 3 mecánica, 1 eléctrica, 1 carrocería
	taller1 := crearTallerConMecanicos("Mec", map[Especialidad]int{
		Mecanica:   3,
		Electrica:  1,
		Carroceria: 1,
	})
	vehiculos1 := generarVehiculosParaTest(taller1, 4, 1, tipos, r)
	resultados1 := simularTallerTest(taller1, vehiculos1)

	// Caso 2: 1 mecánica, 3 eléctrica, 3 carrocería
	taller2 := crearTallerConMecanicos("Mec", map[Especialidad]int{
		Mecanica:   1,
		Electrica:  3,
		Carroceria: 3,
	})
	vehiculos2 := generarVehiculosParaTest(taller2, 4, 1, tipos, r)
	resultados2 := simularTallerTest(taller2, vehiculos2)

	if len(resultados1) != len(resultados2) {
		t.Errorf("Resultados difieren entre distribuciones: %d vs %d", len(resultados1), len(resultados2))
	}
}
