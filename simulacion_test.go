// simulacion_test.go
package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// ---------- AUXILIARES ----------

func crearTallerConMecanicos(nombre string, plantilla map[Especialidad]int) *Taller {
	t := &Taller{}
	for esp, cantidad := range plantilla {
		for i := 0; i < cantidad; i++ {
			t.newMecanico(nombre+string(esp), string(esp), 1)
		}
	}
	return t
}

func generarVehiculosParaTest(t *Taller, numVehiculos int, numIncidencias int, tipos []Especialidad, r *rand.Rand) []*Vehiculo {
	var vehiculos []*Vehiculo
	for i := 1; i <= numVehiculos; i++ {
		v := t.newVehiculo(
			fmt.Sprintf("V-%02d", i),
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

// simulación controlada que devuelve estadísticas
func simularTallerConStats(t *Taller, vehiculos []*Vehiculo) map[string]int {
	chTrabajos := make(chan Trabajo, 100)
	chResultados := make(chan string, 100)
	done := make(chan struct{})

	// mapa para contar incidencias por mecánico
	stats := make(map[string]int)

	for _, m := range t.Mecanicos {
		go func(m *Mecanico) {
			for trabajo := range chTrabajos {
				inc := trabajo.Incidencia
				if inc.Estado == 2 {
					continue
				}
				inc.Estado = 2
				stats[m.Nombre]++
				chResultados <- fmt.Sprintf("%s terminó %s de vehículo %s", m.Nombre, inc.Tipo, trabajo.Vehiculo.Matricula)
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

	// imprimir resultados detallados
	fmt.Println("=== Resultados de la simulación ===")
	total := 0
	for msg := range chResultados {
		fmt.Println("->", msg)
		total++
	}
	fmt.Printf("Total de incidencias procesadas: %d\n", total)
	fmt.Println("Incidencias por mecánico:")
	for mec, n := range stats {
		fmt.Printf("  %s: %d\n", mec, n)
	}

	return stats
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
	vehiculos := generarVehiculosParaTest(taller, 4, 2, tipos, r)

	stats := simularTallerConStats(taller, vehiculos)

	expected := 4 * 2
	sum := 0
	for _, n := range stats {
		sum += n
	}
	if sum != expected {
		t.Errorf("Se esperaban %d incidencias, se obtuvieron %d", expected, sum)
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

	stats := simularTallerConStats(taller, vehiculos)

	expected := 4 * 1
	sum := 0
	for _, n := range stats {
		sum += n
	}
	if sum != expected {
		t.Errorf("Se esperaban %d incidencias, se obtuvieron %d", expected, sum)
	}
}

func TestSimulacionDistribucionMecanicos(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	tipos := []Especialidad{Mecanica, Electrica, Carroceria}

	taller1 := crearTallerConMecanicos("Mec", map[Especialidad]int{
		Mecanica:   3,
		Electrica:  1,
		Carroceria: 1,
	})
	vehiculos1 := generarVehiculosParaTest(taller1, 4, 1, tipos, r)
	stats1 := simularTallerConStats(taller1, vehiculos1)

	taller2 := crearTallerConMecanicos("Mec", map[Especialidad]int{
		Mecanica:   1,
		Electrica:  3,
		Carroceria: 3,
	})
	vehiculos2 := generarVehiculosParaTest(taller2, 4, 1, tipos, r)
	stats2 := simularTallerConStats(taller2, vehiculos2)

	sum1, sum2 := 0, 0
	for _, n := range stats1 {
		sum1 += n
	}
	for _, n := range stats2 {
		sum2 += n
	}

	if sum1 != sum2 {
		t.Errorf("Resultados difieren entre distribuciones: %d vs %d", sum1, sum2)
	}
}
