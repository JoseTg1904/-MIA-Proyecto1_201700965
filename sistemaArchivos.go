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

var posicionActualDD = int64(0)

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
}

type bloqueDatos struct {
	Contenido [25]byte
}

type bitacora struct {
	TipoOperacion [20]byte
	Tipo          byte
	Nombre        [20]byte
	Contenido     [25]byte
	Fecha         [16]byte
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

	fmt.Println("estructuras:", noEstructuras, " Inodos:", noInodos, " bloques: ", noBloques)
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

	fmt.Println("Formateo de unidad completado con exito")
	//escribir el usuario.txt
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
			fmt.Println(i, ". nombre del interno", string(avdAux.Nombre[:]))
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
		fmt.Println("estoy lleno")
		if avdPadre.ApuntadoExtraAVD != -1 {
			fmt.Println("simon :'v")
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

func crearAVD(id, especial, ruta string) {
	listado := strings.Split(ruta, "/")
	listado[0] = "/"
	disco, _, inicio := obtenerDiscoMontado(id)

	super = obtenerSuperBoot(disco, int64(inicio))

	posicionActualAVD = int64(super.InicioAVD)

	banderaSinCrear := false

	if especial == "p" {
		for i := 1; i < len(listado); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i])
			if existenciaAVD == false {
				if super.NoAVDLibres > 0 {
					crearAVDIndividual(disco, int64(super.InicioBitMapAVD), int64(super.InicioAVD), posicionActualAVD, recorrerBitMapAVD(disco, super.InicioBitMapAVD, super.NoAVD), listado[i])
					super.NoAVDLibres--
				} else {
					fmt.Println("a llegado a la capacidad maxima de creacion de carpetas")
				}
			}
		}
		disco.Seek(int64(inicio), 0)
		buffer := bytes.NewBuffer([]byte{})
		binary.Write(buffer, binary.BigEndian, &super)
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
			disco.Seek(int64(inicio), 0)
			buffer := bytes.NewBuffer([]byte{})
			binary.Write(buffer, binary.BigEndian, &super)
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

func crearArchivo() {

}

func recorrerBitMapDD(disco *os.File, inicioBitMap, noDD uint32) int64 {
	bitMap := make([]byte, noDD)
	contenido := make([]byte, noDD)

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
