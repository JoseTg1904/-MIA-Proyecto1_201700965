package main

import (
	"unsafe"
)

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
	SubDirectorios   [6]uint32
	ApuntadorDD      uint32
	ApuntadoExtraAVD uint32
	Propietario      [20]byte
}

type detalleDirectorio struct {
	ArregloArchivos  [5]estructuraInterndaDD
	ApuntadorExtraDD uint32
}

type estructuraInterndaDD struct {
	NombreArchivo  [20]byte
	ApuntadorINodo uint32
	Creacion       [16]byte
	Modificacion   [16]byte
}

type iNodo struct {
	NoINodo                  uint32
	TamanioArchivo           uint32
	ContadorBloquesAsignados uint8
	ApuntadroBloques         [4]uint32
	ApuntadorExtraINodo      uint32
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
	//	interno := estructuraInterndaDD{}
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

	numerador := uint32(tamanioParticion - (uint32(2) * uint32(unsafe.Sizeof(super))))
	denominador := uint32(uint32(27) + uint32(unsafe.Sizeof(avd)) + uint32(unsafe.Sizeof(dd)) + (uint32(5)*uint32(unsafe.Sizeof(inodo)) + (uint32(20) * uint32(unsafe.Sizeof(datos))) + uint32(unsafe.Sizeof(bitacora))))
	return uint32(numerador / denominador)
}

func formateoSistema(tamanioParticion uint32, nombre string) {
	noEstructuras := calcularNoEstructras(tamanioParticion)
	noInodos := uint32(5 * noEstructuras)
	noBloques := uint32(20 * noEstructuras)

	super := superBoot{NoAVD: noEstructuras,
		NoDD:             noEstructuras,
		NoINodos:         noInodos,
		NoBloques:        noBloques,
		NoAVDLibres:      noEstructuras,
		NoDDLibres:       noEstructuras,
		NoINodosLibres:   noInodos,
		NoBloquesLibres:  noBloques,
		ContadorMontajes: 0,
		TamanioAVD:       uint16(unsafe.Sizeof(avd{})),
		TamanioDD:        uint16(unsafe.Sizeof(detalleDirectorio{})),
		TamanioBloque:    uint16(unsafe.Sizeof(bloqueDatos{})),
		TamanioINodo:     uint16(unsafe.Sizeof(iNodo{})),
		Carnet:           201700965}

	copy(super.Nombre[:], nombre)
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
}
