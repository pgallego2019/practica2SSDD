package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const MAX_PLAZAS = 8 // número máximo total de plazas en el taller

// ------------ TIPOS ENUMERADOS ------------

type Especialidad string

const (
	Mecanica   Especialidad = "mecanica"
	Electrica  Especialidad = "electrica"
	Carroceria Especialidad = "carroceria"
)

// ------------ DEFINICIÓN DE LAS ESTRUCTURAS ------------

type Cliente struct {
	ID        int
	Nombre    string
	Telefono  int
	Email     string
	Vehiculos []*Vehiculo
}

type Vehiculo struct {
	Matricula    string
	Marca        string
	Modelo       string
	FechaEntrada string
	FechaSalida  string
	Incidencias  []*Incidencia
	TiempoTotal  int
	Prioritario  bool
}

type Incidencia struct {
	ID              int
	Mecanicos       []*Mecanico
	Tipo            Especialidad
	Prioridad       string
	Descripcion     string
	Estado          int // 0 abierta, 1 en proceso, 2 cerrada
	TiempoAcumulado int
}

type Mecanico struct {
	ID           int
	Nombre       string
	Especialidad Especialidad
	AñosExp      int
	Activo       bool
}

type Plaza struct {
	ID          int
	Ocupada     bool
	VehiculoMat string
	MecanicoID  int
}

type Taller struct {
	Clientes         []*Cliente
	Vehiculos        []*Vehiculo
	Mecanicos        []*Mecanico
	Incidencias      []*Incidencia
	Plazas           []*Plaza
	nextClienteID    int // para que sea incremental y no al azar.
	nextIncidenciaID int
	nextMecanicoID   int
}

// ------------ FUNCIONES DE CREACIÓN ------------

func (t *Taller) newCliente(nombre string, tlf int, email string, vs []*Vehiculo) *Cliente {
	c := &Cliente{
		ID:        t.nextClienteID,
		Nombre:    nombre,
		Telefono:  tlf,
		Email:     email,
		Vehiculos: vs,
	}
	t.nextClienteID++
	t.Clientes = append(t.Clientes, c)
	return c
}

func (t *Taller) newVehiculo(mat string, mar string, mod string, fentrada string, fsalida string, ins []*Incidencia) *Vehiculo {
	v := &Vehiculo{
		Matricula:    mat,
		Marca:        mar,
		Modelo:       mod,
		FechaEntrada: fentrada,
		FechaSalida:  fsalida,
		Incidencias:  ins,
		TiempoTotal:  0,
		Prioritario:  false,
	}
	t.Vehiculos = append(t.Vehiculos, v)
	return v
}

func (t *Taller) newIncidencia(mat string, mecs []*Mecanico, tip string, p string, d string) (*Incidencia, error) {
	esp := Especialidad(strings.ToLower(tip))

	if esp != Mecanica && esp != Electrica && esp != Carroceria {
		return nil, fmt.Errorf("tipo de incidencia inválido (%s)", tip)
	}

	v := t.getVehiculo(mat)

	if v == nil {
		return nil, fmt.Errorf("vehículo con matrícula %s no encontrado", mat)
	}

	inc := &Incidencia{
		ID:              t.nextIncidenciaID,
		Mecanicos:       mecs,
		Tipo:            esp,
		Prioridad:       p,
		Descripcion:     d,
		Estado:          0,
		TiempoAcumulado: 0,
	}
	t.nextIncidenciaID++

	switch esp {
	case Mecanica:
		inc.TiempoAcumulado = 5
	case Electrica:
		inc.TiempoAcumulado = 7
	case Carroceria:
		inc.TiempoAcumulado = 11
	}

	t.Incidencias = append(t.Incidencias, inc)
	v.Incidencias = append(v.Incidencias, inc)
	t.updateTiempoTotalVehiculo(v)
	return inc, nil
}

func (t *Taller) newMecanico(n string, e string, a int) *Mecanico {
	esp := Especialidad(strings.ToLower(e))

	if esp != Mecanica && esp != Electrica && esp != Carroceria {
		fmt.Printf("Especialidad inválida (%s). Debe ser 'mecanica', 'electrica' o 'carroceria'.\n", e)
		return nil
	}

	m := &Mecanico{
		ID:           t.nextMecanicoID,
		Nombre:       n,
		Especialidad: esp,
		AñosExp:      a,
		Activo:       true,
	}
	t.nextMecanicoID++
	t.Mecanicos = append(t.Mecanicos, m)

	// Control del máximo de plazas
	plazasDisponibles := MAX_PLAZAS - len(t.Plazas)
	if plazasDisponibles <= 0 {
		fmt.Printf("No se pueden crear nuevas plazas: límite máximo (%d) alcanzado\n", MAX_PLAZAS)
		return m
	}

	plazasACrear := 2
	if plazasACrear > plazasDisponibles {
		plazasACrear = plazasDisponibles
	}

	for i := 0; i < plazasACrear; i++ {
		plazaID := len(t.Plazas) + 1
		p := &Plaza{
			ID:         plazaID,
			Ocupada:    false,
			MecanicoID: m.ID,
		}
		t.Plazas = append(t.Plazas, p)
	}

	fmt.Printf("Mecánico %s creado (%s) — se añaden %d plazas (total: %d/%d)\n",
		m.Nombre, e, plazasACrear, len(t.Plazas), MAX_PLAZAS)

	return m
}

// ------------ FUNCIONES DE OBTENCIÓN ------------

func (t *Taller) getCliente(id int) *Cliente {
	for _, c := range t.Clientes {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (t *Taller) getVehiculo(mat string) *Vehiculo {
	for _, v := range t.Vehiculos {
		if v.Matricula == mat {
			return v
		}
	}
	return nil
}

func (t *Taller) getIncidencia(id int) *Incidencia {
	for _, inc := range t.Incidencias {
		if inc.ID == id {
			return inc
		}
	}
	return nil
}

func (t *Taller) getMecanico(id int) *Mecanico {
	for _, m := range t.Mecanicos {
		if m.ID == id {
			return m
		}
	}
	return nil
}

// ------------ FUNCIONES DE MODIFICACIÓN ------------

func (t *Taller) updateCliente(id int, nombre string, tlf int, email string) error {
	c := t.getCliente(id)
	if c == nil {
		return fmt.Errorf("cliente con ID %d no encontrado", id)
	}
	if nombre != "" {
		c.Nombre = nombre
	}
	if tlf != 0 {
		c.Telefono = tlf
	}
	if email != "" {
		c.Email = email
	}
	return nil
}

func (t *Taller) updateVehiculo(mat, marca, modelo, fEntrada, fSalida string) error {
	v := t.getVehiculo(mat)
	if v == nil {
		return fmt.Errorf("vehículo con matrícula %s no encontrado", mat)
	}
	if marca != "" {
		v.Marca = marca
	}
	if modelo != "" {
		v.Modelo = modelo
	}
	if fEntrada != "" {
		v.FechaEntrada = fEntrada
	}
	if fSalida != "" {
		v.FechaSalida = fSalida
	}
	return nil
}

func (t *Taller) updateTiempoTotalVehiculo(v *Vehiculo) {
	total := 0
	for _, inc := range v.Incidencias {
		if inc.Estado != 2 {
			total += inc.TiempoAcumulado
		}
	}
	v.TiempoTotal = total

	if v.TiempoTotal > 15 {
		v.Prioritario = true
	}
}

func (t *Taller) updateMecanico(id int, nombre, especialidad string, a int, activo bool) error {
	m := t.getMecanico(id)
	if m == nil {
		return fmt.Errorf("mecánico con ID %d no encontrado", id)
	}
	if nombre != "" {
		m.Nombre = nombre
	}
	if especialidad != "" {
		esp := Especialidad(strings.ToLower(especialidad))
		if esp != Mecanica && esp != Electrica && esp != Carroceria {
			return fmt.Errorf("especialidad inválida (%s): debe ser 'mecanica', 'electrica' o 'carroceria'", especialidad)
		}
		m.Especialidad = esp
	}
	if a != 0 {
		m.AñosExp = a
	}

	if !activo {
		for _, inc := range t.Incidencias {
			for _, mec := range inc.Mecanicos {
				if mec.ID == id && inc.Estado == 0 {
					return fmt.Errorf(
						"no se puede desactivar el mecánico ID %d: tiene una incidencia activa (ID %d)",
						id, inc.ID,
					)
				}
			}
		}
	}

	m.Activo = activo
	return nil
}

func (t *Taller) updateIncidencia(id int, tipo, prioridad, desc string, estado int) error {
	inc := t.getIncidencia(id)
	if inc == nil {
		return fmt.Errorf("incidencia con ID %d no encontrada", id)
	}
	if tipo != "" {
		esp := Especialidad(strings.ToLower(tipo))
		if esp != Mecanica && esp != Electrica && esp != Carroceria {
			return fmt.Errorf("tipo de incidencia inválido (%s)", tipo)
		}
		inc.Tipo = esp
	}
	if prioridad != "" {
		inc.Prioridad = prioridad
	}
	if desc != "" {
		inc.Descripcion = desc
	}
	if estado >= 0 && estado <= 2 {
		inc.Estado = estado
	}
	return nil
}

// ------------ FUNCIONES DE ELIMINACIÓN ------------

func (t *Taller) deleteCliente(id int) {
	for i, c := range t.Clientes {
		if c.ID == id {
			t.Clientes = append(t.Clientes[:i], t.Clientes[i+1:]...)
			break
		}
	}
}

func (t *Taller) deleteVehiculo(mat string) {
	for i, v := range t.Vehiculos {
		if v.Matricula == mat {
			t.Vehiculos = append(t.Vehiculos[:i], t.Vehiculos[i+1:]...)
			break
		}
	}
}

func (t *Taller) deleteIncidencia(id int) {
	for _, v := range t.Vehiculos {
		newIncs := []*Incidencia{}
		for _, inc := range v.Incidencias {
			if inc.ID != id {
				newIncs = append(newIncs, inc)
			}
		}
		v.Incidencias = newIncs
	}

	for i, inc := range t.Incidencias {
		if inc.ID == id {
			t.Incidencias = append(t.Incidencias[:i], t.Incidencias[i+1:]...)
			break
		}
	}
}

func (t *Taller) deleteMecanico(id int) error {
	for _, inc := range t.Incidencias {
		for _, mec := range inc.Mecanicos {
			if mec.ID == id {
				return fmt.Errorf("no se puede eliminar el mecánico ID %d: está asignado a una incidencia ID %d", id, inc.ID)
			}
		}
	}

	for i, m := range t.Mecanicos {
		if m.ID == id {
			t.Mecanicos = append(t.Mecanicos[:i], t.Mecanicos[i+1:]...)
			break
		}
	}
	return nil
}

// ---------- FUNCIONES DE MOSTRAR DATOS ----------

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func printCliente(c *Cliente) {
	fmt.Printf("Cliente %d - %s (%s)\n", c.ID, c.Nombre, c.Email)
	if len(c.Vehiculos) == 0 {
		fmt.Println("  Sin vehículos registrados")
		return
	}
	fmt.Println("  Vehículos:")
	for _, v := range c.Vehiculos {
		fmt.Printf("   - %s (%s %s)\n", v.Matricula, v.Marca, v.Modelo)
	}
}

func (t *Taller) showVehiculosCliente(id int) {
	for _, c := range t.Clientes {
		if c.ID == id {
			fmt.Printf("Vehículos del Cliente %s (ID %d):\n", c.Nombre, c.ID)
			if len(c.Vehiculos) == 0 {
				fmt.Println("\t(No tiene vehículos registrados)")
				return
			}
			for _, v := range c.Vehiculos {
				fmt.Printf("\t%s - %s %s\n", v.Matricula, v.Marca, v.Modelo)
			}
			return
		}
	}
	fmt.Println("Cliente no encontrado.")
}

func printVehiculo(v *Vehiculo) {
	fmt.Printf("Vehículo %s: %s %s (Tiempo estimado en reparar incidencias %d s)\n", v.Matricula, v.Marca, v.Modelo, v.TiempoTotal)
	if len(v.Incidencias) == 0 {
		fmt.Println("  Sin incidencias registradas")
		return
	}
	fmt.Println("  Incidencias:")
	for _, inc := range v.Incidencias {
		fmt.Printf("   - [%s] %s (%s) tiempo estimado de reparación %d s\n", estadoToString(inc.Estado), inc.Tipo, inc.Prioridad, inc.TiempoAcumulado)
	}
}

func (t *Taller) showIncidenciasVehiculo(mat string) {
	for _, v := range t.Vehiculos {
		if v.Matricula == mat {
			fmt.Printf("Vehículo: %s\n", v.Matricula)
			if len(v.Incidencias) == 0 {
				fmt.Println("\t(No tiene incidencias registradas)")
				return
			}
			for _, inc := range v.Incidencias {
				fmt.Printf("\tIncidencia ID %d: %s (Estado: %s)\n",
					inc.ID, inc.Descripcion, estadoToString(inc.Estado))
			}
			return
		}
	}
	fmt.Printf("No se encontró el vehículo con matrícula %s\n", mat)
}

func printIncidencia(i *Incidencia) {
	if i == nil {
		fmt.Println("Incidencia no encontrada.")
		return
	}

	fmt.Printf("ID: %d\n", i.ID)
	fmt.Printf("Tipo: %s\n", i.Tipo)
	fmt.Printf("Prioridad: %s\n", i.Prioridad)
	fmt.Printf("Descripción: %s\n", i.Descripcion)
	fmt.Printf("Tiempo acumulado: %d s\n", i.TiempoAcumulado)

	estadoStr := ""
	switch i.Estado {
	case 0:
		estadoStr = "Abierta"
	case 1:
		estadoStr = "En proceso"
	case 2:
		estadoStr = "Cerrada"
	default:
		estadoStr = "Desconocido"
	}
	fmt.Printf("Estado: %s\n", estadoStr)

	if len(i.Mecanicos) > 0 {
		fmt.Println("Mecánicos asignados:")
		for j, m := range i.Mecanicos {
			fmt.Printf("  %d.\n", j+1)
			printMecanico(m)
		}
	} else {
		fmt.Println("Mecánicos asignados: (ninguno)")
	}
}

func estadoToString(est int) string {
	switch est {
	case 0:
		return "Abierta"
	case 1:
		return "En proceso"
	case 2:
		return "Cerrada"
	default:
		return "Desconocido"
	}
}

func (t *Taller) showIncidenciasMecanico(id int) {
	fmt.Printf("Incidencias del Mecánico ID %d:\n", id)
	hay := false
	for _, inc := range t.Incidencias {
		for _, mec := range inc.Mecanicos {
			if mec.ID == id {
				fmt.Printf("  ID: %d | Tipo: %s | Prioridad: %s | Estado: %s\n",
					inc.ID, inc.Tipo, inc.Prioridad, estadoToString(inc.Estado))
				hay = true
			}
		}
	}
	if !hay {
		fmt.Println("  (no tiene incidencias asignadas)")
	}
}

func printMecanico(m *Mecanico) {
	if m == nil {
		fmt.Println("Mecánico no encontrado.")
		return
	}

	fmt.Printf("ID: %d\n", m.ID)
	fmt.Printf("Nombre: %s\n", m.Nombre)
	fmt.Printf("Especialidad: %s\n", string(m.Especialidad))
	fmt.Printf("Años de experiencia: %d\n", m.AñosExp)
	fmt.Printf("Activo: %t\n", m.Activo)
}

func (t *Taller) showMecanicosActivos() {
	fmt.Println("Mecánicos activos (sin incidencias asignadas):")
	hay := false

	for _, m := range t.Mecanicos {
		asignado := false

		for _, inc := range t.Incidencias {
			for _, mec := range inc.Mecanicos {
				if mec.ID == m.ID {
					asignado = true
					break
				}
			}
			if asignado {
				break
			}
		}

		if !asignado {
			fmt.Printf("ID: %d | Nombre: %s | Especialidad: %s | Años Exp: %d | Activo: %t\n",
				m.ID, m.Nombre, m.Especialidad, m.AñosExp, m.Activo)
			hay = true
		}
	}

	if !hay {
		fmt.Println("  (no hay mecánicos sin incidencias asignadas)")
	}
}

func printPlaza(p *Plaza) {
	if p == nil {
		fmt.Println("Plaza no encontrada.")
		return
	}

	fmt.Printf("ID: %d\n", p.ID)
	fmt.Printf("Ocupada: %t\n", p.Ocupada)
	if p.Ocupada {
		fmt.Printf("Vehículo matricula: %s\n", p.VehiculoMat)
		fmt.Printf("Mecánico asignado (ID): %d\n", p.MecanicoID)
	}
}

func (t *Taller) plazasOcupadas() []*Plaza {
	var ocupadas []*Plaza
	for _, p := range t.Plazas {
		if p.Ocupada {
			ocupadas = append(ocupadas, p)
		}
	}
	return ocupadas
}

// Verifica si un vehículo ha terminado todas sus incidencias y libera su plaza si corresponde
func (t *Taller) liberarPlaza(v *Vehiculo) {
	reparado := true
	for _, inc := range v.Incidencias {
		if inc.Estado != 2 {
			reparado = false
			break
		}
	}

	if !reparado {
		return
	}

	for _, p := range t.Plazas {
		if p.VehiculoMat == v.Matricula {
			p.Ocupada = false
			p.VehiculoMat = ""
			fmt.Printf("Vehículo %s finalizó todas las incidencias. Plaza %d liberada (%d/%d ocupadas)\n",
				v.Matricula, p.ID, len(t.plazasOcupadas()), len(t.Plazas))
			break
		}
	}
}

func printTaller(t *Taller) {
	if t == nil {
		fmt.Println("Taller no encontrado.")
		return
	}

	fmt.Println("=== Taller ===")

	// ---- CLIENTES ----
	fmt.Printf("Clientes (%d):\n", len(t.Clientes))
	if len(t.Clientes) > 0 {
		for i, c := range t.Clientes {
			fmt.Printf("  Cliente %d:\n", i+1)
			printCliente(c)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguno)")
	}

	// ---- VEHÍCULOS ----
	fmt.Printf("\nVehículos (%d):\n", len(t.Vehiculos))
	if len(t.Vehiculos) > 0 {
		for i, v := range t.Vehiculos {
			fmt.Printf("  Vehículo %d:\n", i+1)
			printVehiculo(v)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguno)")
	}

	// ---- MECÁNICOS ----
	fmt.Printf("\nMecánicos (%d):\n", len(t.Mecanicos))
	if len(t.Mecanicos) > 0 {
		for i, m := range t.Mecanicos {
			fmt.Printf("  Mecánico %d:\n", i+1)
			printMecanico(m)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguno)")
	}

	// ---- INCIDENCIAS ----
	fmt.Printf("\nIncidencias (%d):\n", len(t.Incidencias))
	if len(t.Incidencias) > 0 {
		for i, inc := range t.Incidencias {
			fmt.Printf("  Incidencia %d:\n", i+1)
			printIncidencia(inc)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguna)")
	}

	// ---- PLAZAS ----
	fmt.Printf("\nPlazas (%d):\n", len(t.Plazas))
	if len(t.Plazas) > 0 {
		for i, p := range t.Plazas {
			fmt.Printf("  Plaza %d:\n", i+1)
			printPlaza(p)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguna)")
	}

	fmt.Printf("\nPróximo ID cliente: %d\n", t.nextClienteID)
	fmt.Printf("Próximo ID incidencia: %d\n", t.nextIncidenciaID)
	fmt.Printf("Próximo ID mecánico: %d\n", t.nextMecanicoID)
}

func (t *Taller) admitirCliente(clienteID int, v *Vehiculo, mecanicoID int) error {
	// Verificar si el cliente existe
	cliente := t.getCliente(clienteID)
	if cliente == nil {
		return fmt.Errorf("cliente con ID %d no encontrado", clienteID)
	}

	// Verificar si el vehículo ya está asignado a alguna plaza
	for _, p := range t.Plazas {
		if p.VehiculoMat == v.Matricula {
			return fmt.Errorf("el vehículo %s ya está asignado a la plaza %d", v.Matricula, p.ID)
		}
	}

	// Buscar una plaza libre
	plazaLibre := -1
	for i, p := range t.Plazas {
		if !p.Ocupada {
			plazaLibre = i
			break
		}
	}
	if plazaLibre == -1 {
		return fmt.Errorf("no hay plazas disponibles para el vehículo %s", v.Matricula)
	}

	// Asegurar que el vehículo esté en el registro del taller
	existente := t.getVehiculo(v.Matricula)
	if existente == nil {
		t.Vehiculos = append(t.Vehiculos, v)
	}

	// Verificar si el cliente ya tiene el vehículo asignado
	for _, veh := range cliente.Vehiculos {
		if veh.Matricula == v.Matricula {
			return fmt.Errorf("el vehículo %s ya está asignado al cliente %s", v.Matricula, cliente.Nombre)
		}
	}

	// Asignar el vehículo al cliente
	cliente.Vehiculos = append(cliente.Vehiculos, v)

	// Asignar el vehículo a la plaza libre
	t.Plazas[plazaLibre].Ocupada = true
	t.Plazas[plazaLibre].VehiculoMat = v.Matricula
	t.Plazas[plazaLibre].MecanicoID = mecanicoID

	fmt.Printf("Vehículo %s asignado correctamente al cliente %s (plaza %d, mecánico %d)\n",
		v.Matricula, cliente.Nombre, t.Plazas[plazaLibre].ID, mecanicoID)

	return nil
}

// ---------- SUBMENÚS DE LAS ESTRUCTURAS ----------

func menuClientes(t *Taller) {
	for {
		fmt.Println("\n--- CLIENTES ---")
		fmt.Println("1. Crear cliente")
		fmt.Println("2. Mostrar todos los clientes")
		fmt.Println("3. Modificar cliente")
		fmt.Println("4. Eliminar cliente")
		fmt.Println("5. Listar vehículos de un cliente")
		fmt.Println("0. Volver al menú principal")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			var nombre, email string
			var tel int
			fmt.Print("Nombre: ")
			fmt.Scanln(&nombre)
			fmt.Print("Teléfono: ")
			fmt.Scanln(&tel)
			fmt.Print("Email: ")
			fmt.Scanln(&email)
			t.newCliente(nombre, tel, email, nil)
			fmt.Println("Cliente creado correctamente.")
		case 2:
			if len(t.Clientes) == 0 {
				fmt.Println("No hay clientes registrados.")
				break
			}
			for _, c := range t.Clientes {
				printCliente(c)
				fmt.Println("-----------------------------")
			}
		case 3:
			var id, tel int
			var nombre, email string
			fmt.Print("ID de cliente: ")
			fmt.Scanln(&id)
			fmt.Print("Nuevo nombre: ")
			fmt.Scanln(&nombre)
			fmt.Print("Nuevo teléfono: ")
			fmt.Scanln(&tel)
			fmt.Print("Nuevo email: ")
			fmt.Scanln(&email)
			if err := t.updateCliente(id, nombre, tel, email); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Cliente actualizado.")
			}
		case 4:
			var id int
			fmt.Print("ID de cliente: ")
			fmt.Scanln(&id)
			t.deleteCliente(id)
			fmt.Println("Cliente eliminado.")
		case 5:
			var id int
			fmt.Print("ID de cliente: ")
			fmt.Scanln(&id)
			t.showVehiculosCliente(id)
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}

func menuVehiculos(t *Taller) {
	for {
		fmt.Println("\n--- VEHÍCULOS ---")
		fmt.Println("1. Crear vehículo")
		fmt.Println("2. Mostrar todos los vehículos")
		fmt.Println("3. Modificar vehículo")
		fmt.Println("4. Eliminar vehículo")
		fmt.Println("5. Listar incidencias de un vehículo")
		fmt.Println("6. Asignar vehículo a plaza")
		fmt.Println("0. Volver")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			var mat, marca, modelo, fechaE string
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			fmt.Print("Marca: ")
			fmt.Scanln(&marca)
			fmt.Print("Modelo: ")
			fmt.Scanln(&modelo)
			fmt.Print("Fecha de entrada: ")
			fmt.Scanln(&fechaE)
			t.newVehiculo(mat, marca, modelo, fechaE, "", nil)
			fmt.Println("Vehículo creado.")
		case 2:
			if len(t.Vehiculos) == 0 {
				fmt.Println("No hay vehículos registrados.")
				break
			}
			for _, v := range t.Vehiculos {
				printVehiculo(v)
				fmt.Println("-----------------------------")
			}
		case 3:
			var mat, marca, modelo, fe, fs string
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			fmt.Print("Marca: ")
			fmt.Scanln(&marca)
			fmt.Print("Modelo: ")
			fmt.Scanln(&modelo)
			fmt.Print("Fecha entrada: ")
			fmt.Scanln(&fe)
			fmt.Print("Fecha salida: ")
			fmt.Scanln(&fs)
			fmt.Print("Fecha salida: ")
			fmt.Scanln(&fs)
			if err := t.updateVehiculo(mat, marca, modelo, fe, fs); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Vehículo actualizado.")
			}
		case 4:
			var mat string
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			t.deleteVehiculo(mat)
			fmt.Println("Vehículo eliminado.")
		case 5:
			var mat string
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			t.showIncidenciasVehiculo(mat)
		case 6:
			var mat string
			var mecID, clienteID int
			fmt.Print("Matrícula del vehículo: ")
			fmt.Scanln(&mat)
			fmt.Print("ID del cliente: ")
			fmt.Scanln(&clienteID)
			fmt.Print("ID del mecánico: ")
			fmt.Scanln(&mecID)

			v := t.getVehiculo(mat)
			if v == nil {
				fmt.Println("Vehículo no encontrado.")
				break
			}

			err := t.admitirCliente(clienteID, v, mecID)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Vehículo asignado correctamente al cliente.")
			}
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}

func menuIncidencias(t *Taller) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- INCIDENCIAS ---")
		fmt.Println("1. Crear incidencia")
		fmt.Println("2. Mostrar todas las incidencias")
		fmt.Println("3. Modificar incidencia")
		fmt.Println("4. Eliminar incidencia")
		fmt.Println("5. Cambiar estado de incidencia")
		fmt.Println("0. Volver")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			var mat, tipo, pri, desc string
			var mecID int
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			fmt.Print("Tipo: ")
			fmt.Scanln(&tipo)
			fmt.Print("Prioridad: ")
			fmt.Scanln(&pri)
			fmt.Print("Descripción: ")
			desc, _ = reader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			fmt.Print("ID Mecánico: ")
			fmt.Scanln(&mecID)
			mec := t.getMecanico(mecID)
			if mec == nil {
				fmt.Println("Mecánico no encontrado.")
				break
			}
			inc, err := t.newIncidencia(mat, []*Mecanico{mec}, tipo, pri, desc)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("Incidencia creada (ID: %d)\n", inc.ID)
			}
		case 2:
			if len(t.Incidencias) == 0 {
				fmt.Println("No hay incidencias registradas.")
				break
			}
			for _, inc := range t.Incidencias {
				printIncidencia(inc)
				fmt.Println("-----------------------------")
			}
		case 3:
			var id, estado int
			var tipo, pri, desc string
			fmt.Print("ID incidencia: ")
			fmt.Scanln(&id)
			fmt.Print("Tipo: ")
			fmt.Scanln(&tipo)
			fmt.Print("Prioridad: ")
			fmt.Scanln(&pri)
			fmt.Print("Descripción: ")
			desc, _ = reader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			fmt.Print("Estado (0 Abierta, 1 En proceso, 2 Cerrada): ")
			fmt.Scanln(&estado)
			if err := t.updateIncidencia(id, tipo, pri, desc, estado); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Incidencia actualizada.")
			}
		case 4:
			var id int
			fmt.Print("ID incidencia: ")
			fmt.Scanln(&id)
			t.deleteIncidencia(id)
			fmt.Println("Incidencia eliminada.")
		case 5:
			var id, estado int
			fmt.Print("ID incidencia: ")
			fmt.Scanln(&id)
			fmt.Print("Nuevo estado (0 Abierta, 1 En proceso, 2 Cerrada): ")
			fmt.Scanln(&estado)
			inc := t.getIncidencia(id)
			if inc == nil {
				fmt.Println("Incidencia no encontrada.")
				break
			}
			inc.Estado = estado
			fmt.Println("Estado actualizado.")
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}

func menuMecanicos(t *Taller) {
	for {
		fmt.Println("\n--- MECÁNICOS ---")
		fmt.Println("1. Crear mecánico")
		fmt.Println("2. Mostrar todos los mecánicos")
		fmt.Println("3. Modificar mecánico")
		fmt.Println("4. Eliminar mecánico")
		fmt.Println("5. Listar incidencias asignadas a un mecánico")
		fmt.Println("6. Listar mecánicos activos")
		fmt.Println("0. Volver")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			var nombre, esp string
			var exp int
			fmt.Print("Nombre: ")
			fmt.Scanln(&nombre)
			fmt.Print("Especialidad (mecanica / electrica / carroceria): ")
			fmt.Scanln(&esp)
			fmt.Print("Años de experiencia: ")
			fmt.Scanln(&exp)
			m := t.newMecanico(nombre, esp, exp)
			fmt.Printf("Mecánico creado (ID: %d)\n", m.ID)
		case 2:
			if len(t.Mecanicos) == 0 {
				fmt.Println("No hay mecánicos registrados.")
				break
			}
			for _, m := range t.Mecanicos {
				printMecanico(m)
				fmt.Println("-----------------------------")
			}
		case 3:
			var id, exp int
			var nombre, esp string
			var activo bool
			fmt.Print("ID mecánico: ")
			fmt.Scanln(&id)
			fmt.Print("Nuevo nombre: ")
			fmt.Scanln(&nombre)
			fmt.Print("Nueva especialidad (mecanica / electrica / carroceria): ")
			fmt.Scanln(&esp)
			fmt.Print("Años experiencia: ")
			fmt.Scanln(&exp)
			fmt.Print("Activo (1 sí / 0 no): ")
			var act int
			fmt.Scanln(&act)
			activo = act == 1
			if err := t.updateMecanico(id, nombre, esp, exp, activo); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Mecánico actualizado.")
			}
		case 4:
			var id int
			fmt.Print("ID mecánico: ")
			fmt.Scanln(&id)
			if err := t.deleteMecanico(id); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Mecánico eliminado.")
			}
		case 5:
			var id int
			fmt.Print("ID mecánico: ")
			fmt.Scanln(&id)
			t.showIncidenciasMecanico(id)
		case 6:
			t.showMecanicosActivos()
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}

func menuPlazas(t *Taller) {
	for {
		fmt.Println("\n--- PLAZAS / ESTADO DEL TALLER ---")
		fmt.Println("1. Ver estado completo del taller")
		fmt.Println("2. Ver plazas ocupadas/libres")
		fmt.Println("0. Volver")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			printTaller(t)
		case 2:
			if len(t.Plazas) == 0 {
				fmt.Println("No hay plazas registradas.")
				break
			}
			for i := range t.Plazas {
				printPlaza(t.Plazas[i])
				fmt.Println("-----------------------------")
			}
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}

func main() {
	t := &Taller{}

	for {
		fmt.Println("\n===== GESTIÓN DE TALLER =====")
		fmt.Println("1. Clientes")
		fmt.Println("2. Vehículos")
		fmt.Println("3. Incidencias")
		fmt.Println("4. Mecánicos")
		fmt.Println("5. Plazas y estado del taller")
		fmt.Println("6. Limpiar pantalla")
		fmt.Println("7. Simulación concurrente (goroutines)")
		fmt.Println("0. Salir")
		fmt.Print("Seleccione una opción: ")

		var op int
		fmt.Scanln(&op)

		switch op {
		case 1:
			menuClientes(t)
		case 2:
			menuVehiculos(t)
		case 3:
			menuIncidencias(t)
		case 4:
			menuMecanicos(t)
		case 5:
			menuPlazas(t)
		case 6:
			clearScreen()
		case 7:
			simularTaller(t)
		case 0:
			fmt.Println("Saliendo del sistema...")
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}
