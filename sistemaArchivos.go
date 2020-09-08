package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unsafe"
)

var super = superBoot{}

var existenciaAVD = false
var posicionActualAVD = int64(0)

var existenciaArchivo = false
var posicionActualDD = int64(0)

var posicionActualInodo = int64(0)

var contenidoArchivo = ""
var grupoActual = ""
var usuarioActual = ""
var contraActual = ""
var loggeado = false

type superBoot struct {
	Nombre              [20]byte
	NoAVD               uint32
	NoDD                uint32
	NoINodos            uint32
	NoBloques           uint32
	NoAVDLibres         uint32
	NoDDLibres          uint32
	NoINodosLibres      uint32
	NoBloquesLibres     uint32
	Creacion            [16]byte
	UltimoMontaje       [16]byte
	ContadorMontajes    uint16
	InicioBitMapAVD     uint32
	InicioAVD           uint32
	InicioBitMapDD      uint32
	InicioDD            uint32
	InicioBitMapINodo   uint32
	InicioINodo         uint32
	InicioBitMapBloques uint32
	InicioBLoques       uint32
	InicioLog           uint32
	TamanioAVD          uint16
	TamanioDD           uint16
	TamanioINodo        uint16
	TamanioBloque       uint16
	BitLibreAVD         uint32
	BitLibreDD          uint32
	BitLibreINodo       uint32
	BitLibreBloque      uint32
	Carnet              uint32
}

type avd struct {
	Creacion         [16]byte
	Nombre           [20]byte
	SubDirectorios   [6]int64
	ApuntadorDD      int64
	ApuntadoExtraAVD int64
	Propietario      [20]byte
	IDGrupo          [20]byte
	Permisos         [20]byte
}

type detalleDirectorio struct {
	ArregloArchivos  [5]estructuraInterndaDD
	ApuntadorExtraDD int64
}

type estructuraInterndaDD struct {
	NombreArchivo  [20]byte
	ApuntadorINodo int64
	Creacion       [16]byte
	Modificacion   [16]byte
}

type iNodo struct {
	NoINodo                  uint32
	TamanioArchivo           uint32
	ContadorBloquesAsignados uint8
	ApuntadroBloques         [4]int64
	ApuntadorExtraINodo      int64
	Propietario              [20]byte
	IDGrupo                  [20]byte
	Permiso                  [20]byte
}

type bloqueDatos struct {
	Contenido [25]byte
}

type bitacora struct {
	TipoOperacion [8]byte
	Tipo          byte
	Path          [50]byte
	Contenido     [75]byte
	Fecha         [16]byte
	Tamanio       int16
}

func calcularNoEstructras(tamanioParticion uint32) uint32 {
	super := superBoot{}
	avd := avd{}
	//interno := estructuraInterndaDD{}
	dd := detalleDirectorio{}
	inodo := iNodo{}
	datos := bloqueDatos{}
	bitacora := bitacora{}

	/*
		fmt.Println("Tamaño super boot: ", unsafe.Sizeof(super))
		fmt.Println("Tamaño avd: ", unsafe.Sizeof(avd))
		fmt.Println("Tamaño interno dd: ", unsafe.Sizeof(interno))
		fmt.Println("Tamaño dd: ", unsafe.Sizeof(dd))
		fmt.Println("Tamaño I-nodo: ", unsafe.Sizeof(inodo))
		fmt.Println("Tamaño datos: ", unsafe.Sizeof(datos))
		fmt.Println("Tamaño bitacora: ", unsafe.Sizeof(bitacora))
	*/

	//modelo matematico
	/*
		NumeroDeEstructuras: (TamañoDeParticion - (2*TamañoDelSuperbloque)) /
		(27+TamArbolVirtual+TamDetalleDirectorio+(5*TamInodo+(20*TamBloque)+Bitacora))
	*/

	numerador := uint32(tamanioParticion - (2 * uint32(unsafe.Sizeof(super))))
	denominador := uint32(27 + uint32(unsafe.Sizeof(avd)) + uint32(unsafe.Sizeof(dd)) + (5*uint32(unsafe.Sizeof(inodo)) + (20 * uint32(unsafe.Sizeof(datos))) + uint32(unsafe.Sizeof(bitacora))))
	return uint32(numerador / denominador)
}

func formateoSistema(id, tipo string) {

	disco, tamanioParticion, inicioParticion := obtenerDiscoMontado(id)

	if disco == nil {
		fmt.Println("El disco aun no se encuentra montado")
		return
	}

	disco.Seek(int64(inicioParticion), 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))
	for i := 0; i < int(tamanioParticion); i++ {
		disco.Write(buffer.Bytes())
	}

	noEstructuras := calcularNoEstructras(tamanioParticion)
	noInodos := uint32(5 * noEstructuras)
	noBloques := uint32(20 * noEstructuras)

	super := superBoot{NoAVD: noEstructuras,
		NoDD:             noEstructuras,
		NoINodos:         noInodos,
		NoBloques:        noBloques,
		NoAVDLibres:      noEstructuras - 1,
		NoDDLibres:       noEstructuras,
		NoINodosLibres:   noInodos,
		NoBloquesLibres:  noBloques,
		ContadorMontajes: 0,
		TamanioAVD:       uint16(unsafe.Sizeof(avd{})),
		TamanioDD:        uint16(unsafe.Sizeof(detalleDirectorio{})),
		TamanioBloque:    uint16(unsafe.Sizeof(bloqueDatos{})),
		TamanioINodo:     uint16(unsafe.Sizeof(iNodo{})),
		BitLibreAVD:      1,
		Carnet:           201700965}

	ruta := strings.Split(disco.Name(), "/")
	ruta = strings.Split(ruta[len(ruta)-1], ".")

	copy(super.Nombre[:], ruta[0])
	copy(super.Creacion[:], obtenerFecha())
	copy(super.UltimoMontaje[:], obtenerFecha())

	posicion := uint32(unsafe.Sizeof(superBoot{})) + tamanioParticion
	super.InicioBitMapAVD = posicion
	posicion += noEstructuras
	super.InicioAVD = posicion
	posicion += (noEstructuras * uint32(super.TamanioAVD))
	super.InicioBitMapDD = posicion
	posicion += noEstructuras
	super.InicioDD = posicion
	posicion += (noEstructuras * uint32(super.TamanioDD))
	super.InicioBitMapINodo = posicion
	posicion += noInodos
	super.InicioINodo = posicion
	posicion += (noInodos * uint32(super.TamanioINodo))
	super.InicioBitMapBloques = posicion
	posicion += noBloques
	super.InicioBLoques = posicion
	posicion += (noBloques * uint32(super.TamanioBloque))
	super.InicioLog = posicion

	disco.Seek(int64(inicioParticion), 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, &super)
	disco.Write(buffer.Bytes())

	disco.Seek(int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), 0)
	disco.Write(buffer.Bytes())

	disco.Seek(int64(super.InicioBitMapAVD), 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, uint8(1))
	disco.Write(buffer.Bytes())

	avd := avd{SubDirectorios: [6]int64{-1, -1, -1, -1, -1, -1},
		ApuntadorDD:      -1,
		ApuntadoExtraAVD: -1}
	copy(avd.Creacion[:], obtenerFecha())
	copy(avd.Nombre[:], "/")
	copy(avd.Propietario[:], "root")

	disco.Seek(int64(super.InicioAVD), 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, &avd)
	disco.Write(buffer.Bytes())

	posAux := int64(super.InicioLog)
	buffer.Reset()

	bitacoraAux := bitacora{}
	bitacoraAux.Tamanio = -1
	binary.Write(buffer, binary.BigEndian, &bitacoraAux)
	for i := 0; i < int(noEstructuras); i++ {
		disco.Seek(posAux, 0)
		disco.Write(buffer.Bytes())
		posAux += int64(unsafe.Sizeof(bitacora{}))
	}

	contenidoDefecto := "1, G, root\n1, U, root, root, 201700965\n"
	crearArchivo(id, "vacio", "/users.txt", contenidoDefecto, int64(len(contenidoDefecto)))

	fmt.Println("Formateo de unidad completado con exito")

	disco.Close()
}

func verificarExistenciaAVD(disco *os.File, inicioAVD int64, nombreHijo string) {
	banderaEncontrado := false

	nombre := [20]byte{}
	copy(nombre[:], nombreHijo)

	avdPadre := avd{}
	avdAux := avd{}

	//posicionando para leer la carpeta
	disco.Seek(inicioAVD, 0)

	contenido := make([]byte, int(unsafe.Sizeof(avd{})))
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	err = binary.Read(buffer, binary.BigEndian, &avdPadre)
	if err != nil {
	}

	i := 0
	for i = 0; i < 6; i++ {
		if avdPadre.SubDirectorios[i] != -1 {
			disco.Seek(avdPadre.SubDirectorios[i], 0)
			contenido := make([]byte, int(unsafe.Sizeof(avd{})))
			_, err := disco.Read(contenido)
			if err != nil {
				fmt.Println("Error en la lectura del disco")
			}
			buffer := bytes.NewBuffer(contenido)
			err = binary.Read(buffer, binary.BigEndian, &avdAux)
			if err != nil {
			}
			if avdAux.Nombre == nombre {
				banderaEncontrado = true
				break
			}
		}
	}

	if i == 6 {
		i = 5
	}

	if banderaEncontrado == false {
		if avdPadre.ApuntadoExtraAVD != -1 {
			verificarExistenciaAVD(disco, avdPadre.ApuntadoExtraAVD, nombreHijo)
		} else {
			existenciaAVD = false
		}
	} else {
		existenciaAVD = true
		posicionActualAVD = avdPadre.SubDirectorios[i]
	}

}

func obtenerSuperBoot(disco *os.File, inicioParticion int64) superBoot {
	superBootAux := superBoot{}
	contenido := make([]byte, int(unsafe.Sizeof(superBootAux)))
	disco.Seek(inicioParticion, 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &superBootAux)
	if a != nil {
	}
	return superBootAux
}

func obtenerCopiaSuperBoot(disco *os.File, inicioCopia int64) superBoot {
	superBootAux := superBoot{}
	contenido := make([]byte, int(unsafe.Sizeof(superBootAux)))
	disco.Seek(inicioCopia, 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &superBootAux)
	if a != nil {
	}
	return superBootAux
}

func simulacionPerdida(id string) {
	disco, _, inicio := obtenerDiscoMontado(id)
	if disco == nil {
		fmt.Println("El disco aun no se encuentra montado")
		return
	}

	cambiarEstadoPerdida(id, true)

	sb := obtenerSuperBoot(disco, int64(inicio))

	disco.Seek(int64(inicio), 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))

	for i := int(inicio); i < int(sb.InicioLog); i++ {
		disco.Write(buffer.Bytes())
	}

	disco.Close()

	fmt.Println("\nSe a perdido el sistema :c")
}

func recuperarSistema(id string) {

	disco, _, _ := obtenerDiscoMontado(id)

	if disco == nil {
		fmt.Println("El disco no se encuentra montado")
		return
	}

	estado, inicioCopia := obtenerEstadoPerdida(id)

	if estado == false {
		fmt.Println("El sistema se encuentra funcionando correctamente")
		return
	}

	cambiarEstadoPerdida(id, false)
	sbCopia := obtenerCopiaSuperBoot(disco, inicioCopia)

	instrucciones := make([]bitacora, 0)

	posicionActual := int64(sbCopia.InicioLog)

	for {
		disco.Seek(posicionActual, 0)
		bitacoraVacia := bitacora{}
		contenido := make([]byte, int(unsafe.Sizeof(bitacora{})))
		_, err := disco.Read(contenido)
		if err != nil {
		}
		buffer := bytes.NewBuffer(contenido)
		a := binary.Read(buffer, binary.BigEndian, &bitacoraVacia)
		if a != nil {
		}
		if bitacoraVacia.Tamanio != -1 {
			disco.Seek(posicionActual, 0)
			bitacoraAux := bitacora{}
			contenidoBitacora := make([]byte, int64(unsafe.Sizeof(bitacoraAux)))
			_, err := disco.Read(contenidoBitacora)
			if err != nil {
			}
			bufferBitacora := bytes.NewBuffer(contenidoBitacora)
			a := binary.Read(bufferBitacora, binary.BigEndian, &bitacoraAux)
			if a != nil {
			}
			instrucciones = append(instrucciones, bitacoraAux)
			posicionActual += int64(unsafe.Sizeof(bitacora{}))
		} else {
			break
		}

	}

	formateoSistema(id, "full")

	for i := 0; i < len(instrucciones); i++ {
		instruccion := retornarStringLimpio(instrucciones[i].TipoOperacion[:])
		switch instruccion {
		case "mkfile":
			crearArchivo(id, "p", retornarStringLimpio(instrucciones[i].Path[:]), retornarStringLimpio(instrucciones[i].Contenido[:]), int64(instrucciones[i].Tamanio))
		case "mkdir":
			crearAVD(id, "p", retornarStringLimpio(instrucciones[i].Path[:]))
		}
	}

	fmt.Println("El sistema se a recuperado exitosamente c:")

}

func crearAVD(id, especial, ruta string) {
	listado := strings.Split(ruta, "/")

	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		fmt.Println("El disco aun no se encuentra montado")
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("La particion presento una perdida")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	posicionActualAVD = int64(super.InicioAVD)

	banderaSinCrear := false

	iteradorLog := 0

	if especial == "p" {
		for i := 1; i < len(listado); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i])
			if existenciaAVD == false {
				if super.NoAVDLibres > 0 {
					crearAVDIndividual(disco, int64(super.InicioBitMapAVD), int64(super.InicioAVD), posicionActualAVD, recorrerBitMapAVD(disco, super.InicioBitMapAVD, super.NoAVD), listado[i])
					super.NoAVDLibres--
					iteradorLog++
				} else {
					fmt.Println("a llegado a la capacidad maxima de creacion de carpetas")
				}
			}
		}
		if iteradorLog > 0 {
			agregarLog(disco, "mkdir", "1", ruta, "", 1)
		}
		disco.Seek(int64(inicio), 0)
		buffer := bytes.NewBuffer([]byte{})
		binary.Write(buffer, binary.BigEndian, &super)
		disco.Write(buffer.Bytes())

		disco.Seek(int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), 0)
		disco.Write(buffer.Bytes())
		disco.Close()
	} else {
		for i := 1; i < len(listado); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i])
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				} else {
					if super.NoAVDLibres > 0 {
						crearAVDIndividual(disco, int64(super.InicioBitMapAVD), int64(super.InicioAVD), posicionActualAVD, recorrerBitMapAVD(disco, super.InicioBitMapAVD, super.NoAVD), listado[i])
						super.NoAVDLibres--
					} else {
						fmt.Println("a llegado a la capacidad maxima de creacion de carpetas")
					}
				}
			}
		}
		if banderaSinCrear {
			fmt.Println("Una de las carpetas padre aun no se encuentra creada")
		} else {
			agregarLog(disco, "mkdir", "1", ruta, "", 1)
			disco.Seek(int64(inicio), 0)
			buffer := bytes.NewBuffer([]byte{})
			binary.Write(buffer, binary.BigEndian, &super)
			disco.Write(buffer.Bytes())
			disco.Seek(int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), 0)
			disco.Write(buffer.Bytes())
			disco.Close()
		}
	}
}

func crearAVDIndividual(disco *os.File, iniciobitAVD, inicioAVDS, posicionAVDPadre, bitHijo int64, nombre string) {
	carpetaAux := avd{}
	disco.Seek(posicionAVDPadre, 0)
	contenido := make([]byte, int(unsafe.Sizeof(carpetaAux)))
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &carpetaAux)
	if a != nil {
	}
	i := 0
	contadorOcupados := 0
	for i = 0; i < 6; i++ {
		if carpetaAux.SubDirectorios[i] == -1 {
			break
		} else {
			contadorOcupados++
		}
	}

	if i == 6 {
		i = 5
	}

	if contadorOcupados == 6 {
		if super.NoAVDLibres >= 2 {
			disco.Seek(posicionAVDPadre, 0)
			carpetaAux.ApuntadoExtraAVD = inicioAVDS + (bitHijo * int64(unsafe.Sizeof(avd{})))
			buffer.Reset()
			binary.Write(buffer, binary.BigEndian, &carpetaAux)
			disco.Write(buffer.Bytes())
			crearAVDAnexo(disco, iniciobitAVD, inicioAVDS, bitHijo, carpetaAux.Nombre, nombre)
		} else {
			fmt.Println("a llegado al maximo de inserciones de carpetas posibles")
		}
	} else {
		//creacion individual de la carpeta
		carpetaNueva := avd{SubDirectorios: [6]int64{-1, -1, -1, -1, -1, -1},
			ApuntadorDD:      -1,
			ApuntadoExtraAVD: -1}
		copy(carpetaNueva.Creacion[:], obtenerFecha())
		copy(carpetaNueva.Nombre[:], nombre)
		copy(carpetaNueva.Propietario[:], "root")

		super.BitLibreAVD++

		posicion := inicioAVDS + (bitHijo * int64(unsafe.Sizeof(avd{})))
		carpetaAux.SubDirectorios[i] = posicion

		disco.Seek(posicion, 0)

		buffer.Reset()
		binary.Write(buffer, binary.BigEndian, &carpetaNueva)
		disco.Write(buffer.Bytes())

		disco.Seek(iniciobitAVD+bitHijo, 0)

		buffer.Reset()
		binary.Write(buffer, binary.BigEndian, uint8(1))
		disco.Write(buffer.Bytes())

		disco.Seek(posicionAVDPadre, 0)
		buffer.Reset()
		binary.Write(buffer, binary.BigEndian, &carpetaAux)
		disco.Write(buffer.Bytes())
	}
}

func crearAVDAnexo(disco *os.File, iniciobitAVD, inicioAVDS, bitAnexa int64, nombre [20]byte, nombreHija string) {
	carpetaNueva := avd{SubDirectorios: [6]int64{-1, -1, -1, -1, -1, -1},
		ApuntadorDD:      -1,
		ApuntadoExtraAVD: -1,
		Nombre:           nombre}
	copy(carpetaNueva.Creacion[:], obtenerFecha())
	copy(carpetaNueva.Propietario[:], "root")

	posicion := inicioAVDS + (bitAnexa * int64(unsafe.Sizeof(avd{})))

	disco.Seek(posicion, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, &carpetaNueva)
	disco.Write(buffer.Bytes())

	disco.Seek(iniciobitAVD+bitAnexa, 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, uint8(1))
	disco.Write(buffer.Bytes())

	super.BitLibreAVD++
	super.NoAVDLibres--

	crearAVDIndividual(disco, iniciobitAVD, inicioAVDS, posicion, bitAnexa+1, nombreHija)
}

func recorrerBitMapAVD(disco *os.File, inicioBitMap, noAVD uint32) int64 {
	bitMap := make([]byte, noAVD)
	contenido := make([]byte, noAVD)

	disco.Seek(int64(inicioBitMap), 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bitMap)
	if a != nil {
	}

	i := 0
	for i = 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			break
		}
	}

	return int64(i)
}

func crearArchivo(id, especial, ruta, cont string, size int64) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		fmt.Println("El disco aun no se encuentra montado")
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("La particion presento una perdida")
		return
	}

	contAux := cont

	super = obtenerSuperBoot(disco, int64(inicio))

	if size == -1 {
		size = int64(len(cont))
	} else if size > int64(len(cont)) {
		iterador := 0
		for {
			if len(cont) == int(size) {
				break
			}
			cont += arregloLetras[iterador]
			iterador++
			if iterador == len(arregloLetras) {
				iterador = 0
			}
		}
	} else if size < int64(len(cont)) {
		cont = cont[:size]
	}

	rutaCarpeta := strings.Split(ruta, "/")

	rutaValidar := ""
	for i := 1; i < len(rutaCarpeta)-1; i++ {
		rutaValidar += "/" + rutaCarpeta[i]
	}

	if especial == "p" {
		crearAVD(id, especial, rutaValidar)
		disco, _, inicio = obtenerDiscoMontado(id)
	} else {
		posicionActualAVD = int64(super.InicioAVD)
		listado := strings.Split(rutaValidar, "/")
		for i := 1; i < len(listado); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i])
			if existenciaAVD == false {
				fmt.Println("Alguna de las carpetas padre no existe")
				return
			}
		}
	}

	avdAux := avd{}
	disco.Seek(posicionActualAVD, 0)
	contenido := make([]byte, int(unsafe.Sizeof(avd{})))
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	bufferLectura := bytes.NewBuffer(contenido)
	err = binary.Read(bufferLectura, binary.BigEndian, &avdAux)
	if err != nil {
	}

	if avdAux.ApuntadorDD == -1 {
		posicionDD := int64(super.InicioDD) + (recorrerBitMapDD(disco, super.InicioBitMapDD, super.NoDD) * int64(unsafe.Sizeof(detalleDirectorio{})))
		ddAux := detalleDirectorio{}
		ddAux.ApuntadorExtraDD = -1

		if super.NoDDLibres == 0 {
			fmt.Println("Ya no se pueden crear mas detalles de directorio")
			disco.Close()
			return
		}

		interno := estructuraInterndaDD{}
		copy(interno.Creacion[:], obtenerFecha())
		copy(interno.Modificacion[:], obtenerFecha())
		copy(interno.NombreArchivo[:], rutaCarpeta[len(rutaCarpeta)-1])

		noBloquesNecesarios := 1
		iterador := 0
		for i := 0; i < len(cont); i++ {
			if iterador == 25 {
				iterador = 0
				noBloquesNecesarios++
			}
			iterador++
		}

		noInodosNecesarios := 1
		iterador = 0
		for i := 0; i < noBloquesNecesarios; i++ {
			if iterador == 4 {
				iterador = 0
				noInodosNecesarios++
			}
			iterador++
		}

		if super.NoINodosLibres < uint32(noInodosNecesarios) {
			fmt.Println("No pueden crearse los inodos necesarios")
			disco.Close()
			return
		}

		if super.NoBloquesLibres < uint32(noBloquesNecesarios) {
			fmt.Print("\nNo pueden crearse los bloques necesarios\n")
			disco.Close()
			return
		}

		inicioDivision := 0
		divisionContenido := make([]string, noBloquesNecesarios)
		for i := 0; i < noBloquesNecesarios; i++ {
			if i == noBloquesNecesarios-1 {
				divisionContenido[i] = cont[inicioDivision:len(cont)]
			} else {
				divisionContenido[i] = cont[inicioDivision : inicioDivision+25]
				inicioDivision += 25
			}
		}

		bitInodo := recorrerBitMapInodo(disco, super.InicioBitMapINodo, super.NoINodos)
		posicionINodo := int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
		interno.ApuntadorINodo = posicionINodo
		ddAux.ArregloArchivos[0] = interno

		j := 0
		iterador = 0
		for i := 0; i < noInodosNecesarios; i++ {
			inodoAux := iNodo{}
			inodoAux.TamanioArchivo = uint32(len(cont))
			inodoAux.NoINodo = uint32(bitInodo)
			inodoAux.ApuntadroBloques = [4]int64{-1, -1, -1, -1}
			copy(inodoAux.Propietario[:], "root")
			bufferEscritura := bytes.NewBuffer([]byte{})
			for j < noBloquesNecesarios {
				if iterador == 4 {
					iterador = 0
					break
				}
				bitBloque := recorrerBitMapBloques(disco, super.InicioBitMapBloques, super.NoBloques)
				posicionBloque := int64(super.InicioBLoques) + (bitBloque * int64(unsafe.Sizeof(bloqueDatos{})))
				bloque := bloqueDatos{}
				copy(bloque.Contenido[:], divisionContenido[j])
				disco.Seek(posicionBloque, 0)
				bufferEscritura.Reset()
				binary.Write(bufferEscritura, binary.BigEndian, &bloque)
				disco.Write(bufferEscritura.Bytes())
				inodoAux.ApuntadroBloques[iterador] = posicionBloque
				bufferEscritura.Reset()
				disco.Seek(int64(super.InicioBitMapBloques)+bitBloque, 0)
				binary.Write(bufferEscritura, binary.BigEndian, uint8(1))
				disco.Write(bufferEscritura.Bytes())
				j++
				super.NoBloquesLibres = super.NoBloquesLibres - 1
				iterador++
			}
			if iterador == 0 {
				inodoAux.ContadorBloquesAsignados = uint8(4)
			} else {
				inodoAux.ContadorBloquesAsignados = uint8(iterador)
			}
			bufferEscritura.Reset()
			disco.Seek(int64(super.InicioBitMapINodo)+bitInodo, 0)
			binary.Write(bufferEscritura, binary.BigEndian, uint8(1))
			disco.Write(bufferEscritura.Bytes())
			if i == noInodosNecesarios-1 {
				inodoAux.ApuntadorExtraINodo = -1
			} else {
				inodoAux.ApuntadorExtraINodo = int64(super.InicioINodo) + (recorrerBitMapInodo(disco, super.InicioBitMapINodo, super.NoINodos) * int64(unsafe.Sizeof(iNodo{})))
			}
			disco.Seek(posicionINodo, 0)
			bufferEscritura.Reset()
			binary.Write(bufferEscritura, binary.BigEndian, &inodoAux)
			disco.Write(bufferEscritura.Bytes())
			bitInodo = recorrerBitMapInodo(disco, super.InicioBitMapINodo, super.NoINodos)
			posicionINodo = int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
			super.NoINodosLibres = super.NoINodosLibres - 1
		}

		avdAux.ApuntadorDD = posicionDD

		bufferEscritura := bytes.NewBuffer([]byte{})
		disco.Seek(posicionDD, 0)
		bufferEscritura.Reset()
		binary.Write(bufferEscritura, binary.BigEndian, &ddAux)
		disco.Write(bufferEscritura.Bytes())

		disco.Seek(recorrerBitMapDD(disco, super.InicioBitMapDD, super.NoDD)+int64(super.InicioBitMapDD), 0)
		bufferEscritura.Reset()
		binary.Write(bufferEscritura, binary.BigEndian, uint8(1))
		disco.Write(bufferEscritura.Bytes())

		disco.Seek(posicionActualAVD, 0)
		bufferEscritura.Reset()
		binary.Write(bufferEscritura, binary.BigEndian, &avdAux)
		disco.Write(bufferEscritura.Bytes())

		super.NoDDLibres = super.NoDDLibres - 1
		super.BitLibreBloque = uint32(recorrerBitMapBloques(disco, super.InicioBitMapBloques, super.NoBloques))
		super.BitLibreDD = uint32(recorrerBitMapDD(disco, super.InicioBitMapDD, super.NoDD))
		super.BitLibreINodo = uint32(recorrerBitMapInodo(disco, super.InicioBitMapINodo, super.NoINodos))

		bufferEscritura.Reset()
		disco.Seek(int64(inicio), 0)
		binary.Write(bufferEscritura, binary.BigEndian, &super)
		disco.Write(bufferEscritura.Bytes())
		disco.Seek(int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), 0)
		disco.Write(bufferEscritura.Bytes())

		agregarLog(disco, "mkfile", "0", ruta, contAux, size)

		disco.Close()
		fmt.Print("\nSe a creado el archivo exitosamente\n")
		return
	} else {
		posicionDD := avdAux.ApuntadorDD
		disco.Seek(posicionDD, 0)
		ddAux := detalleDirectorio{}
		contenido := make([]byte, int(unsafe.Sizeof(detalleDirectorio{})))
		_, err := disco.Read(contenido)
		if err != nil {
			fmt.Println("Error en la lectura del disco")
		}
		bufferDD := bytes.NewBuffer(contenido)
		err = binary.Read(bufferDD, binary.BigEndian, &ddAux)
		if err != nil {
		}

		internaVacia := estructuraInterndaDD{}

		banderaExistenciaArchivo := false
		var nombreAux [20]byte
		copy(nombreAux[:], rutaCarpeta[len(rutaCarpeta)-1])
		i := 0
		for i = 0; i < 5; i++ {
			if ddAux.ArregloArchivos[i] != internaVacia {
				if ddAux.ArregloArchivos[i].NombreArchivo == nombreAux {
					banderaExistenciaArchivo = true
					break
				}
			}
		}

		if i == 5 {
			i = 4
		}

		if banderaExistenciaArchivo {
			borrarInodos(disco, ddAux.ArregloArchivos[i].ApuntadorINodo)
			ddAux.ArregloArchivos[i] = internaVacia
		}

		banderaDisponible := false

		if super.NoDDLibres == 0 {
			fmt.Println("Ya no se pueden crear mas detalles de directorio")
			return
		}

		i = 0
		for i = 0; i < 5; i++ {
			if ddAux.ArregloArchivos[i] == internaVacia {
				banderaDisponible = true
				break
			}
		}

		if i == 5 {
			i = 4
		}

		banderaDDAnexo := false

		if banderaDisponible == false {
			if super.NoDD > 0 {
				banderaDDAnexo = true
				posicionAux := int64(super.InicioDD) + (recorrerBitMapDD(disco, super.InicioBitMapDD, super.NoDD) * int64(unsafe.Sizeof(detalleDirectorio{})))
				ddAux.ApuntadorExtraDD = posicionAux
				disco.Seek(posicionDD, 0)
				bufferDDPadre := bytes.NewBuffer([]byte{})
				binary.Write(bufferDDPadre, binary.BigEndian, &ddAux)
				disco.Write(bufferDDPadre.Bytes())
				posicionDD = posicionAux
				super.NoDDLibres--
			} else {
				fmt.Println("No se pueden crear mas detalles de directorio")
				return
			}
		}

		disco.Seek(posicionDD, 0)
		ddAux = detalleDirectorio{}
		contenido = make([]byte, int(unsafe.Sizeof(detalleDirectorio{})))
		_, err = disco.Read(contenido)
		if err != nil {
			fmt.Println("Error en la lectura del disco")
		}
		bufferDD.Reset()
		bufferDD = bytes.NewBuffer(contenido)
		err = binary.Read(bufferDD, binary.BigEndian, &ddAux)
		if err != nil {
		}
		if banderaDDAnexo {
			ddAux.ApuntadorExtraDD = -1
		}

		banderaDisponible = false

		i = 0
		for i = 0; i < 5; i++ {
			if ddAux.ArregloArchivos[i] == internaVacia {
				banderaDisponible = true
				break
			}
		}

		if i == 5 {
			i = 4
		}

		interno := estructuraInterndaDD{}
		copy(interno.Creacion[:], obtenerFecha())
		copy(interno.Modificacion[:], obtenerFecha())
		copy(interno.NombreArchivo[:], rutaCarpeta[len(rutaCarpeta)-1])

		noBloquesNecesarios := 1
		iterador := 0
		for i := 0; i < len(cont); i++ {
			if iterador == 25 {
				iterador = 0
				noBloquesNecesarios++
			}
			iterador++
		}

		noInodosNecesarios := 1
		iterador = 0
		for i := 0; i < noBloquesNecesarios; i++ {
			if iterador == 4 {
				iterador = 0
				noInodosNecesarios++
			}
			iterador++
		}

		if super.NoINodosLibres < uint32(noInodosNecesarios) {
			fmt.Println("No pueden crearse los inodos necesarios")
			disco.Close()
			return
		}

		if super.NoBloquesLibres < uint32(noBloquesNecesarios) {
			fmt.Println("No pueden crearse los bloques necesarios")
			disco.Close()
			return
		}

		inicioDivision := 0
		divisionContenido := make([]string, noBloquesNecesarios)
		for i := 0; i < noBloquesNecesarios; i++ {
			if i == noBloquesNecesarios-1 {
				divisionContenido[i] = cont[inicioDivision:len(cont)]
			} else {
				divisionContenido[i] = cont[inicioDivision : inicioDivision+25]
				inicioDivision += 25
			}
		}

		bitInodo := recorrerBitMapInodo(disco, super.InicioBitMapINodo, super.NoINodos)
		posicionINodo := int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
		interno.ApuntadorINodo = posicionINodo
		ddAux.ArregloArchivos[i] = interno

		j := 0
		iterador = 0
		for i := 0; i < noInodosNecesarios; i++ {
			inodoAux := iNodo{}
			inodoAux.TamanioArchivo = uint32(len(cont))
			inodoAux.NoINodo = uint32(bitInodo)
			inodoAux.ApuntadroBloques = [4]int64{-1, -1, -1, -1}
			copy(inodoAux.Propietario[:], "root")
			bufferEscritura := bytes.NewBuffer([]byte{})
			for j < noBloquesNecesarios {
				if iterador == 4 {
					iterador = 0
					break
				}
				bitBloque := recorrerBitMapBloques(disco, super.InicioBitMapBloques, super.NoBloques)
				posicionBloque := int64(super.InicioBLoques) + (bitBloque * int64(unsafe.Sizeof(bloqueDatos{})))
				bloque := bloqueDatos{}
				copy(bloque.Contenido[:], divisionContenido[j])
				disco.Seek(posicionBloque, 0)
				bufferEscritura.Reset()
				binary.Write(bufferEscritura, binary.BigEndian, &bloque)
				disco.Write(bufferEscritura.Bytes())
				inodoAux.ApuntadroBloques[iterador] = posicionBloque
				bufferEscritura.Reset()
				disco.Seek(int64(super.InicioBitMapBloques)+bitBloque, 0)
				binary.Write(bufferEscritura, binary.BigEndian, uint8(1))
				disco.Write(bufferEscritura.Bytes())
				j++
				super.NoBloquesLibres--
				iterador++
			}
			if iterador == 0 {
				inodoAux.ContadorBloquesAsignados = uint8(4)
			} else {
				inodoAux.ContadorBloquesAsignados = uint8(iterador)
			}
			bufferEscritura.Reset()
			disco.Seek(int64(super.InicioBitMapINodo)+bitInodo, 0)
			binary.Write(bufferEscritura, binary.BigEndian, uint8(1))
			disco.Write(bufferEscritura.Bytes())
			if i == noInodosNecesarios-1 {
				inodoAux.ApuntadorExtraINodo = -1
			} else {
				inodoAux.ApuntadorExtraINodo = int64(super.InicioINodo) + (recorrerBitMapInodo(disco, super.InicioBitMapINodo, super.NoINodos) * int64(unsafe.Sizeof(iNodo{})))
			}
			disco.Seek(posicionINodo, 0)
			bufferEscritura.Reset()
			binary.Write(bufferEscritura, binary.BigEndian, &inodoAux)
			disco.Write(bufferEscritura.Bytes())
			bitInodo = recorrerBitMapInodo(disco, super.InicioBitMapINodo, super.NoINodos)
			posicionINodo = int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
			super.NoINodosLibres--
		}

		super.BitLibreBloque = uint32(recorrerBitMapBloques(disco, super.InicioBitMapBloques, super.NoBloques))
		super.BitLibreINodo = uint32(recorrerBitMapInodo(disco, super.InicioBitMapINodo, super.NoINodos))

		bufferEscritura := bytes.NewBuffer([]byte{})
		disco.Seek(posicionDD, 0)
		binary.Write(bufferEscritura, binary.BigEndian, &ddAux)
		disco.Write(bufferEscritura.Bytes())

		if banderaDDAnexo {
			disco.Seek(recorrerBitMapDD(disco, super.InicioBitMapDD, super.NoDD)+int64(super.InicioBitMapDD), 0)
			bufferEscritura.Reset()
			binary.Write(bufferEscritura, binary.BigEndian, uint8(1))
			disco.Write(bufferEscritura.Bytes())
		}
		super.BitLibreDD = uint32(recorrerBitMapDD(disco, super.InicioBitMapDD, super.NoDD))
		disco.Seek(int64(inicio), 0)
		bufferEscritura.Reset()
		binary.Write(bufferEscritura, binary.BigEndian, &super)
		disco.Write(bufferEscritura.Bytes())
		disco.Seek(int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), 0)
		disco.Write(bufferEscritura.Bytes())
		agregarLog(disco, "mkfile", "0", ruta, contAux, size)

		disco.Close()

		fmt.Print("\nSe a creado el archivo exitosamente\n")
		return
	}
}

func borrarInodos(disco *os.File, posicionINodo int64) {
	bitINodo := (posicionINodo - int64(super.InicioINodo)) / int64(unsafe.Sizeof(iNodo{}))
	disco.Seek(int64(super.InicioBitMapINodo)+bitINodo, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))
	disco.Write(buffer.Bytes())

	disco.Seek(posicionINodo, 0)
	inodoAux := iNodo{}
	contenido := make([]byte, int(unsafe.Sizeof(iNodo{})))
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer.Reset()
	buffer = bytes.NewBuffer(contenido)
	err = binary.Read(buffer, binary.BigEndian, &inodoAux)
	if err != nil {
	}

	if inodoAux.ApuntadorExtraINodo != -1 {
		borrarInodos(disco, inodoAux.ApuntadorExtraINodo)
	}

	for i := 0; i < 3; i++ {
		if inodoAux.ApuntadroBloques[i] != -1 {
			borrarBloques(disco, inodoAux.ApuntadroBloques[i])
		}
	}

	disco.Seek(posicionINodo, 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, uint8(0))
	for i := 0; i < int(unsafe.Sizeof(iNodo{})); i++ {
		disco.Write(buffer.Bytes())
	}
	super.NoINodosLibres++
}

func borrarBloques(disco *os.File, posicionBloque int64) {
	bitBloque := (posicionBloque - int64(super.InicioBLoques)) / int64(unsafe.Sizeof(bloqueDatos{}))
	disco.Seek(int64(super.InicioBitMapBloques)+bitBloque, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))
	disco.Write(buffer.Bytes())

	disco.Seek(posicionBloque, 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, uint8(0))
	for i := 0; i < int(unsafe.Sizeof(bloqueDatos{})); i++ {
		disco.Write(buffer.Bytes())
	}
	super.NoBloquesLibres++
}

func recorrerBitMapDD(disco *os.File, inicioBitMap, noDD uint32) int64 {
	bitMap := make([]byte, noDD)
	contenido := make([]byte, noDD)

	disco.Seek(int64(inicioBitMap), 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco bit map dd")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bitMap)
	if a != nil {
	}

	i := 0
	for i = 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			break
		}
	}

	return int64(i)
}

func recorrerBitMapBloques(disco *os.File, inicioBitMap, noBloques uint32) int64 {
	bitMap := make([]byte, noBloques)
	contenido := make([]byte, noBloques)

	disco.Seek(int64(inicioBitMap), 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco bit map bloques")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bitMap)
	if a != nil {
	}

	i := 0
	for i = 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			break
		}
	}

	return int64(i)
}

func recorrerBitMapInodo(disco *os.File, inicioBitMap, noInodo uint32) int64 {
	bitMap := make([]byte, noInodo)
	contenido := make([]byte, noInodo)

	disco.Seek(int64(inicioBitMap), 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco inodos")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bitMap)
	if a != nil {
	}

	i := 0
	for i = 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			break
		}
	}

	return int64(i)
}

func agregarLog(disco *os.File, operacion, tipo, path, contenido string, size int64) {
	bitacoraAux := bitacora{Tamanio: int16(size),
		Tipo: tipo[0]}
	copy(bitacoraAux.TipoOperacion[:], operacion)
	copy(bitacoraAux.Path[:], path)
	copy(bitacoraAux.Contenido[:], contenido)
	copy(bitacoraAux.Fecha[:], obtenerFecha())

	posicionLog := int64(super.InicioLog)
	for i := 0; i < int(super.NoAVD); i++ {
		disco.Seek(posicionLog, 0)
		logAux := bitacora{}
		content := make([]byte, int(unsafe.Sizeof(bitacora{})))
		_, err := disco.Read(content)
		if err != nil {
			fmt.Println("Error en la lectura del disco")
		}
		buffer := bytes.NewBuffer(content)
		a := binary.Read(buffer, binary.BigEndian, &logAux)
		if a != nil {
		}
		if logAux.Tamanio == -1 {
			break
		}
		posicionLog += int64(unsafe.Sizeof(bitacora{}))
	}

	disco.Seek(posicionLog, 0)

	bufferEscritura := bytes.NewBuffer([]byte{})
	binary.Write(bufferEscritura, binary.BigEndian, &bitacoraAux)
	disco.Write(bufferEscritura.Bytes())
}

func busquedaArchivoDD(disco *os.File, posicionActualDD int64, nombre string) {
	ddAux := detalleDirectorio{}
	internoVacio := estructuraInterndaDD{}

	contenido := make([]byte, int(unsafe.Sizeof(ddAux)))
	disco.Seek(posicionActualDD, 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &ddAux)
	if a != nil {
	}

	name := [20]byte{}
	copy(name[:], nombre)

	for i := 0; i < 5; i++ {
		if ddAux.ArregloArchivos[i] != internoVacio {
			if ddAux.ArregloArchivos[i].NombreArchivo == name {
				existenciaArchivo = true
				posicionActualInodo = ddAux.ArregloArchivos[i].ApuntadorINodo
				break
			}
		}
	}

	if existenciaArchivo == false {
		if ddAux.ApuntadorExtraDD != -1 {
			busquedaArchivoDD(disco, ddAux.ApuntadorExtraDD, nombre)
		}
	}
}

func busquedaArchivoBloques(disco *os.File, posicionActualBloque int64) {
	bloqueAux := bloqueDatos{}

	contenido := make([]byte, int(unsafe.Sizeof(bloqueAux)))
	disco.Seek(posicionActualBloque, 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bloqueAux)
	if a != nil {
	}

	contenidoArchivo += retornarStringLimpio(bloqueAux.Contenido[:])
}

func busquedaArchivoInodo(disco *os.File, posicionActualInodo int64) {
	inodoAux := iNodo{}

	contenido := make([]byte, int(unsafe.Sizeof(inodoAux)))
	disco.Seek(posicionActualInodo, 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &inodoAux)
	if a != nil {
	}

	for i := 0; i < 4; i++ {
		if inodoAux.ApuntadroBloques[i] != -1 {
			busquedaArchivoBloques(disco, inodoAux.ApuntadroBloques[i])
		}
	}

	if inodoAux.ApuntadorExtraINodo != -1 {
		busquedaArchivoInodo(disco, inodoAux.ApuntadorExtraINodo)
	}
}

func obtenerContenidoArchivo(disco *os.File, path string) string {

	contenidoArchivo = ""

	posicionActualAVD = int64(super.InicioAVD)

	listado := strings.Split(path, "/")
	for i := 1; i < len(listado)-1; i++ {
		verificarExistenciaAVD(disco, posicionActualAVD, listado[i])
		if existenciaAVD == false {
			fmt.Println("Alguna de las carpetas padres a las cual desea acceder no existe")
			return ""
		}
	}

	avdAux := avd{}
	contenido := make([]byte, int(unsafe.Sizeof(avdAux)))
	disco.Seek(posicionActualAVD, 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &avdAux)
	if a != nil {
	}

	if avdAux.ApuntadorDD == -1 {
		fmt.Println("La carpeta no tiene ningun archivo asociado")
		return ""
	}

	busquedaArchivoDD(disco, avdAux.ApuntadorDD, listado[len(listado)-1])

	if posicionActualInodo == 0 {
		fmt.Println("El archivo no existe en la carpeta")
		return ""
	}

	busquedaArchivoInodo(disco, posicionActualInodo)

	return contenidoArchivo
}

func cerrarSesion() {
	if loggeado == false {
		fmt.Println("Debe de iniciar sesion para poder cerrarla")
		return
	}

	loggeado = false
	usuarioActual = ""
	grupoActual = ""
	contraActual = ""
	fmt.Println("Se a cerrado la sesion exitosamente")
}

func iniciarSesion(id, usr, pwd string) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		fmt.Println("El disco no se encuentra montado")
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("La particion presento una perdida")
		return
	}

	if loggeado == true {
		fmt.Println("Ya existe una sesion iniciada en el sistema")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	lineas := strings.Split(obtenerContenidoArchivo(disco, "/users.txt"), "\n")

	banderaEncontado := false

	for i := 0; i < len(lineas)-1; i++ {
		linea := strings.Split(lineas[i], ",")
		if val := strings.Trim(linea[0], " "); val != "0" {
			if strings.Trim(linea[1], " ") == "U" {
				if strings.Trim(linea[3], " ") == usr && strings.Trim(linea[4], " ") == pwd {
					usuarioActual = usr
					contraActual = pwd
					grupoActual = strings.Trim(linea[2], " ")
					banderaEncontado = true
					loggeado = true
				}
			}
		}
	}

	if banderaEncontado {
		fmt.Println("A inicio sesion exitosamente")
	} else {
		fmt.Println("Sus credenciales no se han econtrado en el sistema")
	}

}
