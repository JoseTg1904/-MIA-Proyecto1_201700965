package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
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
	Propietario      [10]byte
	IDGrupo          [10]byte
	Permisos         [3]byte
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
	Propietario              [10]byte
	IDGrupo                  [10]byte
	Permiso                  [3]byte
}

type bloqueDatos struct {
	Contenido [25]byte
}

type bitacora struct {
	TipoOperacion [8]byte
	Tipo          byte
	Path          [50]byte
	Contenido     [100]byte
	Fecha         [16]byte
	Tamanio       int16
}

func calcularNoEstructras(tamanioParticion uint32) (uint32, uint32, uint32) {
	/*
		Modelo matemacio:
		No Estructuras = (TamañoDeParticion - (2*TamañoDelSuperbloque)) /
		(27+TamArbolVirtual+TamDetalleDirectorio+(5*TamInodo+(20*TamBloque)+Bitacora))
	*/
	//no Inodos = 5 * no Estructuras
	//no Bloques = 20 * no Estructuras
	//para las demas estructuras son iguales al no Estructura

	numerador := uint32(tamanioParticion - (2 * uint32(unsafe.Sizeof(superBoot{}))))
	denominador := uint32(27 + uint32(unsafe.Sizeof(avd{})) + uint32(unsafe.Sizeof(detalleDirectorio{})) + (5*uint32(unsafe.Sizeof(iNodo{})) + (20 * uint32(unsafe.Sizeof(bloqueDatos{}))) + uint32(unsafe.Sizeof(bitacora{}))))
	division := uint32(numerador / denominador)
	return division, (division * 5), (division * 20)
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

	noEstructuras, noInodos, noBloques := calcularNoEstructras(tamanioParticion)

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

	//obteniendo el nombre del disco
	ruta := strings.Split(disco.Name(), "/")
	ruta = strings.Split(ruta[len(ruta)-1], ".")

	copy(super.Nombre[:], ruta[0])
	copy(super.Creacion[:], obtenerFecha())
	copy(super.UltimoMontaje[:], obtenerFecha())

	//rellenando los inicios de las secciones del formateo
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

	escribirEnBitMap(disco, 1, int64(super.InicioBitMapAVD))

	avdRaiz := avd{SubDirectorios: [6]int64{-1, -1, -1, -1, -1, -1},
		ApuntadorDD:      -1,
		ApuntadoExtraAVD: -1}
	copy(avdRaiz.Creacion[:], obtenerFecha())
	copy(avdRaiz.Nombre[:], "/")
	copy(avdRaiz.Propietario[:], "root")
	copy(avdRaiz.IDGrupo[:], "root")
	copy(avdRaiz.Permisos[:], "664")

	escribirAVD(disco, int64(super.InicioAVD), avdRaiz)

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

	contenidoDefecto := "1,G,root\n1,U,root,root,201700965\n"
	crearArchivo(id, "vacio", "/users.txt", contenidoDefecto, int64(len(contenidoDefecto)), 0)

	fmt.Println("\033[1;32mFormateo de unidad completado con exito\033[0m")

	disco.Close()
}

func verificarExistenciaAVD(disco *os.File, inicioAVD int64, nombreHijo string, insercion int64) {
	banderaEncontrado := false

	//copiando el string en un arreglo de bytes para poder comparar
	nombre := [20]byte{}
	copy(nombre[:], nombreHijo)

	//obteniendo el avd padre
	avdPadre := obtenerAVD(disco, inicioAVD)

	//recorrido para validar si la carpeta hija se encuentra en la carpeta padre
	i := 0
	for i = 0; i < 6; i++ {
		if avdPadre.SubDirectorios[i] != -1 {
			avdAux := obtenerAVD(disco, avdPadre.SubDirectorios[i])
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
		//no se encontro la carpeta hija pero la carpeta padre puede tener una anexa
		//verificar en ella
		if avdPadre.ApuntadoExtraAVD != -1 {
			posicionActualAVD = avdPadre.ApuntadoExtraAVD
			verificarExistenciaAVD(disco, avdPadre.ApuntadoExtraAVD, nombreHijo, insercion)
		} else {
			existenciaAVD = false
			//la carpeta buscada no existe, verificar si se desea insertar la carpeta en el sistema y crearla
			if insercion == 1 {
				crearAVDIndividual(disco, int64(super.InicioBitMapAVD), int64(super.InicioAVD), posicionActualAVD, obtenerBitLibreBitMap(disco, super.InicioBitMapAVD, super.NoAVD), nombreHijo)
			}
		}
	} else {
		//la carpeta buscada si existe entonces esa seria la carpeta actual para insertar archivos
		existenciaAVD = true
		posicionActualAVD = avdPadre.SubDirectorios[i]
	}

}

func obtenerSuperBoot(disco *os.File, inicioParticion int64) superBoot {
	superBootAux := superBoot{}
	contenido := make([]byte, int(unsafe.Sizeof(superBootAux)))
	disco.Seek(inicioParticion, 0)
	disco.Read(contenido)
	buffer := bytes.NewBuffer(contenido)
	binary.Read(buffer, binary.BigEndian, &superBootAux)
	return superBootAux
}

func obtenerDD(disco *os.File, posicionDD int64) detalleDirectorio {
	//variable que almacena el struct del avd,
	ddAux := detalleDirectorio{}

	//variable que almacena el contenido leido del disco
	contenido := make([]byte, int(unsafe.Sizeof(ddAux)))

	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionDD, 0)

	//obteniendo el contenido del archivo y asignandolo a la variable del contenido
	disco.Read(contenido)

	//asignando el contenido leido a un buffer para decodificar el binario
	bufferLectura := bytes.NewBuffer(contenido)

	//decodificando el contenido del buffer y asignandolo al struct
	binary.Read(bufferLectura, binary.BigEndian, &ddAux)

	return ddAux
}

func escribirDD(disco *os.File, posicionDD int64, dd detalleDirectorio) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionDD, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &dd)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

func obtenerBloque(disco *os.File, posicionBloque int64) bloqueDatos {
	//variable que almacena el struct del avd,
	bloqueAux := bloqueDatos{}

	//variable que almacena el contenido leido del disco
	contenido := make([]byte, int(unsafe.Sizeof(bloqueAux)))

	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionBloque, 0)

	//obteniendo el contenido del archivo y asignandolo a la variable del contenido
	disco.Read(contenido)

	//asignando el contenido leido a un buffer para decodificar el binario
	bufferLectura := bytes.NewBuffer(contenido)

	//decodificando el contenido del buffer y asignandolo al struct
	binary.Read(bufferLectura, binary.BigEndian, &bloqueAux)

	return bloqueAux
}

func escribirBloque(disco *os.File, posicionBloque int64, bloque bloqueDatos) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionBloque, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &bloque)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

func obtenerInodo(disco *os.File, posicionInodo int64) iNodo {
	//variable que almacena el struct del avd,
	inodoAux := iNodo{}

	//variable que almacena el contenido leido del disco
	contenido := make([]byte, int(unsafe.Sizeof(inodoAux)))

	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionInodo, 0)

	//obteniendo el contenido del archivo y asignandolo a la variable del contenido
	disco.Read(contenido)

	//asignando el contenido leido a un buffer para decodificar el binario
	bufferLectura := bytes.NewBuffer(contenido)

	//decodificando el contenido del buffer y asignandolo al struct
	binary.Read(bufferLectura, binary.BigEndian, &inodoAux)

	return inodoAux
}

func escribirInodo(disco *os.File, posicionInodo int64, inodo iNodo) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionInodo, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &inodo)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

func obtenerAVD(disco *os.File, posicionAVD int64) avd {
	//variable que almacena el struct del avd,
	avdAux := avd{}

	//variable que almacena el contenido leido del disco
	contenido := make([]byte, int(unsafe.Sizeof(avdAux)))

	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionAVD, 0)

	//obteniendo el contenido del archivo y asignandolo a la variable del contenido
	disco.Read(contenido)

	//asignando el contenido leido a un buffer para decodificar el binario
	bufferLectura := bytes.NewBuffer(contenido)

	//decodificando el contenido del buffer y asignandolo al struct
	binary.Read(bufferLectura, binary.BigEndian, &avdAux)

	return avdAux
}

func escribirAVD(disco *os.File, posicionAVD int64, carpeta avd) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionAVD, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &carpeta)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

func escribirSuperBoot(disco *os.File, posicionSuper int64, arranque superBoot) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionSuper, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &arranque)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

func escribirEnBitMap(disco *os.File, valor uint8, posicionBit int64) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionBit, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &valor)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

func obtenerBitLibreBitMap(disco *os.File, inicioBitMap, noEstructuras uint32) int64 {
	bitMap := make([]byte, noEstructuras)
	contenido := make([]byte, noEstructuras)

	disco.Seek(int64(inicioBitMap), 0)
	disco.Read(contenido)
	buffer := bytes.NewBuffer(contenido)
	binary.Read(buffer, binary.BigEndian, &bitMap)
	i := 0
	for i = 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			break
		}
	}

	if i == len(bitMap) {
		i = len(bitMap) - 1
	}

	return int64(i)
}

func simulacionPerdida(id string) {
	disco, _, inicio := obtenerDiscoMontado(id)
	if disco == nil {
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

	fmt.Println("\033[1;33mSe a perdido el sistema :C\033[0m")
}

func recuperarSistema(id string) {

	disco, _, _ := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, inicioCopia := obtenerEstadoPerdida(id)

	if estado == false {
		fmt.Println("\033[1;32mEl sistema se encuentra funcionando correctamente c:\033[0m")
		return
	}

	cambiarEstadoPerdida(id, false)
	sbCopia := obtenerSuperBoot(disco, inicioCopia)

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
			crearArchivo(id, "p", retornarStringLimpio(instrucciones[i].Path[:]), retornarStringLimpio(instrucciones[i].Contenido[:]), int64(instrucciones[i].Tamanio), 1)
		case "mkdir":
			crearAVD(id, "p", retornarStringLimpio(instrucciones[i].Path[:]), 1)
		case "mkgrp":
			crearGrupo(id, retornarStringLimpio(instrucciones[i].Contenido[:]))
		case "rmgrp":
			eliminarGrupo(id, retornarStringLimpio(instrucciones[i].Contenido[:]))
		case "mkusr":
			dividido := strings.Split(retornarStringLimpio(instrucciones[i].Contenido[:]), ",")
			sinSalto := strings.ReplaceAll(dividido[4], "\n", "")
			crearUsuario(id, dividido[3], sinSalto, dividido[2])
		case "rmusr":
			eliminarUsuario(id, retornarStringLimpio(instrucciones[i].Contenido[:]))
		case "edit":
			modificarContenidoArchivo(id, retornarStringLimpio(instrucciones[i].Path[:]), retornarStringLimpio(instrucciones[i].Contenido[:]), int64(instrucciones[i].Tamanio))
		case "ren":
			modificarNombre(id, retornarStringLimpio(instrucciones[i].Path[:]), retornarStringLimpio(instrucciones[i].Contenido[:]))
		case "rm":
			eliminarArchivosOCarpetas(id, retornarStringLimpio(instrucciones[i].Path[:]), "")
		case "mv":
			moverArchivoYCarpetas(id, retornarStringLimpio(instrucciones[i].Path[:]), retornarStringLimpio(instrucciones[i].Contenido[:]))
		}
	}

	fmt.Println("\033[1;32mEl sistema se a recuperado exitosamente c:\033[0m")
}

func crearAVD(id, especial, ruta string, insercionLog int) {
	listado := strings.Split(ruta, "/")

	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida :c\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	posicionActualAVD = int64(super.InicioAVD)

	banderaSinCrear := false

	existenciaAVD = false

	if especial == "p" {
		for i := 1; i < len(listado); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 1)
		}
		if insercionLog == 1 {
			agregarLog(disco, "mkdir", "1", ruta, "", 1)
		}

		super.BitLibreAVD = uint32(obtenerBitLibreBitMap(disco, super.InicioAVD, super.NoAVD))
		fmt.Println("\033[1;32mSe han creado las carpetas\033[0m")
		escribirSuperBoot(disco, int64(inicio), super)
		escribirSuperBoot(disco, int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), super)
		disco.Close()
	} else {
		for i := 1; i < len(listado); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				} else {
					if super.NoAVDLibres > 0 {
						crearAVDIndividual(disco, int64(super.InicioBitMapAVD), int64(super.InicioAVD), posicionActualAVD, obtenerBitLibreBitMap(disco, super.InicioBitMapAVD, super.NoAVD), listado[i])
						super.NoAVDLibres--
					} else {
						fmt.Println("\033[1;31mA llegado al maximo de creacion de carpetas\033[0m")
					}
				}
			}
		}
		if banderaSinCrear {
			fmt.Println("\033[1;31mUna de las carpetas padre aun no se encuentra creada\033[0m")
		} else {
			agregarLog(disco, "mkdir", "1", ruta, "", 1)
			fmt.Println("\033[1;32mSe han creado las carpetas\033[0m")
			super.BitLibreAVD = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapAVD, super.NoAVD))
			escribirSuperBoot(disco, int64(inicio), super)
			escribirSuperBoot(disco, int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), super)
			disco.Close()
		}
	}
}

func crearAVDIndividual(disco *os.File, iniciobitAVD, inicioAVDS, posicionAVDPadre, bitHijo int64, nombre string) {
	//variable que almacena la carpeta a validar
	carpetaAux := obtenerAVD(disco, posicionAVDPadre)

	//verificacion de los apuntadores hacia subdirectorios de la carpeta padre
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
		//verificando que se puedan crear almenos dos carpetas debido a que se crea una anexa
		//mas la original que se pretendia crear
		if super.NoAVDLibres >= 2 {
			//modificacion del puntero hacia carpeta anexa
			carpetaAux.ApuntadoExtraAVD = inicioAVDS + (bitHijo * int64(unsafe.Sizeof(avd{})))

			//escritura de la carpeta padre con el apuntador modificado hacia su carpeta anexa
			escribirAVD(disco, posicionAVDPadre, carpetaAux)

			crearAVDAnexo(disco, iniciobitAVD, inicioAVDS, bitHijo, carpetaAux.Nombre, nombre)
		} else {
			fmt.Println("\033[1;31mA llegado al maximo de instrucciones de carpetas posibles\033[0m")
		}
	} else {

		//creacion individual de la carpeta
		carpetaNueva := avd{SubDirectorios: [6]int64{-1, -1, -1, -1, -1, -1},
			ApuntadorDD:      -1,
			ApuntadoExtraAVD: -1}
		copy(carpetaNueva.Creacion[:], obtenerFecha())
		copy(carpetaNueva.Nombre[:], nombre)
		copy(carpetaNueva.Propietario[:], "root")
		copy(carpetaNueva.IDGrupo[:], "root")
		copy(carpetaNueva.Permisos[:], "664")
		super.BitLibreAVD++

		posicionActualAVD = inicioAVDS + (bitHijo * int64(unsafe.Sizeof(avd{})))

		carpetaAux.SubDirectorios[i] = posicionActualAVD

		escribirAVD(disco, posicionActualAVD, carpetaNueva)

		escribirEnBitMap(disco, 1, (iniciobitAVD + bitHijo))

		escribirAVD(disco, posicionAVDPadre, carpetaAux)
	}
}

func crearAVDAnexo(disco *os.File, iniciobitAVD, inicioAVDS, bitAnexa int64, nombre [20]byte, nombreHija string) {
	//creando la carpeta anexa a la carpeta padre enviada
	carpetaNueva := avd{SubDirectorios: [6]int64{-1, -1, -1, -1, -1, -1},
		ApuntadorDD:      -1,
		ApuntadoExtraAVD: -1,
		Nombre:           nombre}
	copy(carpetaNueva.Creacion[:], obtenerFecha())
	copy(carpetaNueva.Propietario[:], "root")
	copy(carpetaNueva.IDGrupo[:], "root")
	copy(carpetaNueva.Permisos[:], "664")

	//obteniendo la posicion de la carpeta anexa
	posicion := inicioAVDS + (bitAnexa * int64(unsafe.Sizeof(avd{})))

	escribirAVD(disco, posicion, carpetaNueva)

	escribirEnBitMap(disco, 1, (iniciobitAVD + bitAnexa))

	//reduciendo la cantidad de avd libres
	super.NoAVDLibres--

	crearAVDIndividual(disco, iniciobitAVD, inicioAVDS, posicion, bitAnexa+1, nombreHija)
}

func crearArchivo(id, especial, ruta, cont string, size int64, insercionLog int) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida :c\033[0m")
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
		crearAVD(id, especial, rutaValidar, 1)
		disco, _, inicio = obtenerDiscoMontado(id)
	} else {
		posicionActualAVD = int64(super.InicioAVD)
		listado := strings.Split(rutaValidar, "/")
		for i := 1; i < len(listado); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
			if existenciaAVD == false {
				fmt.Println("\033[1;31mAlguna de las carpeatas padre no existe\033[0m")
				return
			}
		}
	}

	avdAux := obtenerAVD(disco, posicionActualAVD)

	if avdAux.ApuntadorDD == -1 {
		posicionDD := int64(super.InicioDD) + (obtenerBitLibreBitMap(disco, super.InicioBitMapDD, super.NoDD) * int64(unsafe.Sizeof(detalleDirectorio{})))
		ddAux := detalleDirectorio{}
		ddAux.ApuntadorExtraDD = -1

		if super.NoDDLibres == 0 {
			fmt.Println("\033[1;31mYa no se pueden crear mas detalles de directorio\033[0m")
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
			fmt.Println("\033[1;31mNo pueden crearse los inodos necesarios\033[0m")
			disco.Close()
			return
		}

		if super.NoBloquesLibres < uint32(noBloquesNecesarios) {
			fmt.Println("\033[1;31mNo pueden crearse los bloques necesarios\033[0m")
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

		bitInodo := obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos)
		posicionINodo := int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
		interno.ApuntadorINodo = posicionINodo
		ddAux.ArregloArchivos[0] = interno

		internoLimpio := estructuraInterndaDD{ApuntadorINodo: -1}
		copy(internoLimpio.NombreArchivo[:], "")
		ddAux.ArregloArchivos[1] = internoLimpio
		ddAux.ArregloArchivos[2] = internoLimpio
		ddAux.ArregloArchivos[3] = internoLimpio
		ddAux.ArregloArchivos[4] = internoLimpio

		j := 0
		iterador = 0
		for i := 0; i < noInodosNecesarios; i++ {
			inodoAux := iNodo{}
			inodoAux.TamanioArchivo = uint32(len(cont))
			inodoAux.NoINodo = uint32(bitInodo)
			inodoAux.ApuntadroBloques = [4]int64{-1, -1, -1, -1}
			copy(inodoAux.IDGrupo[:], "root")
			copy(inodoAux.Permiso[:], "664")
			copy(inodoAux.Propietario[:], "root")
			for j < noBloquesNecesarios {
				if iterador == 4 {
					iterador = 0
					break
				}
				bitBloque := obtenerBitLibreBitMap(disco, super.InicioBitMapBloques, super.NoBloques)
				posicionBloque := int64(super.InicioBLoques) + (bitBloque * int64(unsafe.Sizeof(bloqueDatos{})))
				bloque := bloqueDatos{}
				copy(bloque.Contenido[:], divisionContenido[j])

				escribirBloque(disco, posicionBloque, bloque)

				inodoAux.ApuntadroBloques[iterador] = posicionBloque

				escribirEnBitMap(disco, 1, int64(super.InicioBitMapBloques)+bitBloque)

				j++
				super.NoBloquesLibres = super.NoBloquesLibres - 1
				iterador++
			}
			if iterador == 0 {
				inodoAux.ContadorBloquesAsignados = uint8(4)
			} else {
				inodoAux.ContadorBloquesAsignados = uint8(iterador)
			}

			escribirEnBitMap(disco, 1, int64(super.InicioBitMapINodo)+bitInodo)

			if i == noInodosNecesarios-1 {
				inodoAux.ApuntadorExtraINodo = -1
			} else {
				inodoAux.ApuntadorExtraINodo = int64(super.InicioINodo) + (obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos) * int64(unsafe.Sizeof(iNodo{})))
			}

			escribirInodo(disco, posicionINodo, inodoAux)

			bitInodo = obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos)
			posicionINodo = int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
			super.NoINodosLibres = super.NoINodosLibres - 1
		}

		avdAux.ApuntadorDD = posicionDD

		escribirDD(disco, posicionDD, ddAux)

		escribirEnBitMap(disco, 1, obtenerBitLibreBitMap(disco, super.InicioBitMapDD, super.NoDD)+int64(super.InicioBitMapDD))

		escribirAVD(disco, posicionActualAVD, avdAux)

		super.NoDDLibres = super.NoDDLibres - 1
		super.BitLibreBloque = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapBloques, super.NoBloques))
		super.BitLibreDD = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapDD, super.NoDD))
		super.BitLibreINodo = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos))

		escribirSuperBoot(disco, int64(inicio), super)
		escribirSuperBoot(disco, int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), super)

		if insercionLog == 1 {
			agregarLog(disco, "mkfile", "0", ruta, contAux, size)
		}

		disco.Close()
		fmt.Println("\033[1;32mSe a creado el archivo exitosamente\033[0m")
	} else {
		banderaDDAnexo := false
		posicionDD := avdAux.ApuntadorDD
		nombreAux := [20]byte{}
		copy(nombreAux[:], rutaCarpeta[len(rutaCarpeta)-1])

		for {
			ddAux := obtenerDD(disco, posicionDD)

			banderaExistenciaArchivo := false

			i := 0
			for i = 0; i < 5; i++ {
				if ddAux.ArregloArchivos[i].ApuntadorINodo != -1 {
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
				internaLimpia := estructuraInterndaDD{ApuntadorINodo: -1}
				copy(internaLimpia.NombreArchivo[:], "")
				ddAux.ArregloArchivos[i] = internaLimpia
				escribirDD(disco, posicionDD, ddAux)
			}

			banderaDisponible := false

			i = 0
			for i = 0; i < 5; i++ {
				if ddAux.ArregloArchivos[i].ApuntadorINodo == -1 {
					banderaDisponible = true
					break
				}
			}

			if i == 5 {
				i = 4
			}

			if banderaDisponible == false {
				if ddAux.ApuntadorExtraDD != -1 {
					posicionDD = ddAux.ApuntadorExtraDD
				} else {
					if super.NoDDLibres > 0 {
						banderaDDAnexo = true
						posicionAux := int64(super.InicioDD) + (obtenerBitLibreBitMap(disco, super.InicioBitMapDD, super.NoDD) * int64(unsafe.Sizeof(detalleDirectorio{})))
						ddAux.ApuntadorExtraDD = posicionAux
						escribirDD(disco, posicionDD, ddAux)
						ddNuevo := detalleDirectorio{ApuntadorExtraDD: -1}
						internoLimpio := estructuraInterndaDD{ApuntadorINodo: -1}
						copy(internoLimpio.NombreArchivo[:], "")
						ddNuevo.ArregloArchivos[0] = internoLimpio
						ddNuevo.ArregloArchivos[1] = internoLimpio
						ddNuevo.ArregloArchivos[2] = internoLimpio
						ddNuevo.ArregloArchivos[3] = internoLimpio
						ddNuevo.ArregloArchivos[4] = internoLimpio
						escribirDD(disco, posicionAux, ddNuevo)
						posicionDD = posicionAux
						super.NoDDLibres--
						break
					} else {
						fmt.Println("\033[1;31mNo se pueden crear mas detalles de directorio\033[0m")
						return
					}
				}
			} else {
				break
			}
		}

		ddAux := obtenerDD(disco, posicionDD)

		if banderaDDAnexo {
			ddAux.ApuntadorExtraDD = -1
		}

		i := 0
		for i = 0; i < 5; i++ {
			if ddAux.ArregloArchivos[i].ApuntadorINodo == -1 {
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
		for k := 0; k < len(cont); k++ {
			if iterador == 25 {
				iterador = 0
				noBloquesNecesarios++
			}
			iterador++
		}

		noInodosNecesarios := 1
		iterador = 0
		for l := 0; l < noBloquesNecesarios; l++ {
			if iterador == 4 {
				iterador = 0
				noInodosNecesarios++
			}
			iterador++
		}

		if super.NoINodosLibres < uint32(noInodosNecesarios) {
			fmt.Println("\033[1;31mNo pueden crearse los inodos necesarios\033[0m")
			disco.Close()
			return
		}

		if super.NoBloquesLibres < uint32(noBloquesNecesarios) {
			fmt.Println("\033[1;31mNo pueden crearse los bloques necesarios\033[0m")
			disco.Close()
			return
		}

		inicioDivision := 0
		divisionContenido := make([]string, noBloquesNecesarios)
		for m := 0; m < noBloquesNecesarios; m++ {
			if m == noBloquesNecesarios-1 {
				divisionContenido[m] = cont[inicioDivision:len(cont)]
			} else {
				divisionContenido[m] = cont[inicioDivision : inicioDivision+25]
				inicioDivision += 25
			}
		}

		bitInodo := obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos)
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
			copy(inodoAux.IDGrupo[:], "root")
			copy(inodoAux.Permiso[:], "664")
			for j < noBloquesNecesarios {
				if iterador == 4 {
					iterador = 0
					break
				}
				bitBloque := obtenerBitLibreBitMap(disco, super.InicioBitMapBloques, super.NoBloques)
				posicionBloque := int64(super.InicioBLoques) + (bitBloque * int64(unsafe.Sizeof(bloqueDatos{})))
				bloque := bloqueDatos{}
				copy(bloque.Contenido[:], divisionContenido[j])

				escribirBloque(disco, posicionBloque, bloque)

				inodoAux.ApuntadroBloques[iterador] = posicionBloque

				escribirEnBitMap(disco, 1, int64(super.InicioBitMapBloques)+bitBloque)

				j++
				super.NoBloquesLibres--
				iterador++
			}
			if iterador == 0 {
				inodoAux.ContadorBloquesAsignados = uint8(4)
			} else {
				inodoAux.ContadorBloquesAsignados = uint8(iterador)
			}

			escribirEnBitMap(disco, 1, int64(super.InicioBitMapINodo)+bitInodo)

			if i == noInodosNecesarios-1 {
				inodoAux.ApuntadorExtraINodo = -1
			} else {
				inodoAux.ApuntadorExtraINodo = int64(super.InicioINodo) + (obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos) * int64(unsafe.Sizeof(iNodo{})))
			}

			escribirInodo(disco, posicionINodo, inodoAux)

			bitInodo = obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos)
			posicionINodo = int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
			super.NoINodosLibres--
		}

		super.BitLibreBloque = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapBloques, super.NoBloques))
		super.BitLibreINodo = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos))

		escribirDD(disco, posicionDD, ddAux)

		if banderaDDAnexo {
			escribirEnBitMap(disco, 1, obtenerBitLibreBitMap(disco, super.InicioBitMapDD, super.NoDD)+int64(super.InicioBitMapDD))
		}

		super.BitLibreDD = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapDD, super.NoDD))

		escribirSuperBoot(disco, int64(inicio), super)
		escribirSuperBoot(disco, int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), super)

		if insercionLog == 1 {
			agregarLog(disco, "mkfile", "0", ruta, contAux, size)
		}

		disco.Close()
		fmt.Println("\033[1;32mSe a creado el archivo exitosamente\033[0m")
	}
}

func borrarAVD(disco *os.File, posicionAVD int64) {
	bitAVD := (posicionAVD - int64(super.InicioAVD)) / int64(unsafe.Sizeof(avd{}))

	escribirEnBitMap(disco, 0, int64(super.InicioBitMapAVD)+bitAVD)

	avdAux := obtenerAVD(disco, posicionAVD)

	if avdAux.ApuntadoExtraAVD != -1 {
		borrarAVD(disco, avdAux.ApuntadoExtraAVD)
	}

	for i := 0; i < 6; i++ {
		if avdAux.SubDirectorios[i] != -1 {
			borrarAVD(disco, avdAux.SubDirectorios[i])
		}
	}

	if avdAux.ApuntadorDD != -1 {
		borrarDD(disco, avdAux.ApuntadorDD)
	}

	disco.Seek(posicionAVD, 0)
	bufferCeros := bytes.NewBuffer([]byte{})
	binary.Write(bufferCeros, binary.BigEndian, uint8(0))
	for i := 0; i < int(unsafe.Sizeof(avd{})); i++ {
		disco.Write(bufferCeros.Bytes())
	}
	super.NoAVDLibres++
}

func borrarDD(disco *os.File, posicionDD int64) {
	bitDD := (posicionDD - int64(super.InicioDD)) / int64(unsafe.Sizeof(detalleDirectorio{}))

	escribirEnBitMap(disco, 0, int64(super.InicioBitMapDD)+bitDD)

	ddAux := obtenerDD(disco, posicionDD)

	if ddAux.ApuntadorExtraDD != -1 {
		borrarDD(disco, ddAux.ApuntadorExtraDD)
	}

	for i := 0; i < 5; i++ {
		if ddAux.ArregloArchivos[i].ApuntadorINodo != -1 {
			borrarInodos(disco, ddAux.ArregloArchivos[i].ApuntadorINodo)
		}
	}

	disco.Seek(posicionDD, 0)
	bufferCeros := bytes.NewBuffer([]byte{})
	binary.Write(bufferCeros, binary.BigEndian, uint8(0))
	for i := 0; i < int(unsafe.Sizeof(detalleDirectorio{})); i++ {
		disco.Write(bufferCeros.Bytes())
	}
	super.NoDDLibres++
}

func borrarInodos(disco *os.File, posicionINodo int64) {
	bitINodo := (posicionINodo - int64(super.InicioINodo)) / int64(unsafe.Sizeof(iNodo{}))

	escribirEnBitMap(disco, 0, int64(super.InicioBitMapINodo)+bitINodo)

	inodoAux := obtenerInodo(disco, posicionINodo)

	if inodoAux.ApuntadorExtraINodo != -1 {
		borrarInodos(disco, inodoAux.ApuntadorExtraINodo)
	}

	for i := 0; i < 4; i++ {
		if inodoAux.ApuntadroBloques[i] != -1 {
			borrarBloques(disco, inodoAux.ApuntadroBloques[i])
		}
	}

	disco.Seek(posicionINodo, 0)
	bufferCeros := bytes.NewBuffer([]byte{})
	binary.Write(bufferCeros, binary.BigEndian, uint8(0))
	for i := 0; i < int(unsafe.Sizeof(iNodo{})); i++ {
		disco.Write(bufferCeros.Bytes())
	}
	super.NoINodosLibres++
}

func borrarBloques(disco *os.File, posicionBloque int64) {
	bitBloque := (posicionBloque - int64(super.InicioBLoques)) / int64(unsafe.Sizeof(bloqueDatos{}))

	escribirEnBitMap(disco, 0, int64(super.InicioBitMapBloques)+bitBloque)

	disco.Seek(posicionBloque, 0)
	bufferCeros := bytes.NewBuffer([]byte{})
	binary.Write(bufferCeros, binary.BigEndian, uint8(0))
	for i := 0; i < int(unsafe.Sizeof(bloqueDatos{})); i++ {
		disco.Write(bufferCeros.Bytes())
	}
	super.NoBloquesLibres++
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
		disco.Read(content)
		buffer := bytes.NewBuffer(content)
		binary.Read(buffer, binary.BigEndian, &logAux)
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
	ddAux := obtenerDD(disco, posicionActualDD)

	name := [20]byte{}
	copy(name[:], nombre)

	existencia := false
	for i := 0; i < 5; i++ {
		if ddAux.ArregloArchivos[i].ApuntadorINodo != -1 {
			if ddAux.ArregloArchivos[i].NombreArchivo == name {
				existencia = true
				posicionActualInodo = ddAux.ArregloArchivos[i].ApuntadorINodo
				break
			}
		}
	}

	if existencia == false {
		if ddAux.ApuntadorExtraDD != -1 {
			busquedaArchivoDD(disco, ddAux.ApuntadorExtraDD, nombre)
		}
	}
}

func busquedaArchivoBloques(disco *os.File, posicionActualBloque int64) {
	bloqueAux := obtenerBloque(disco, posicionActualBloque)
	contenidoArchivo += retornarStringLimpio(bloqueAux.Contenido[:])
}

func busquedaArchivoInodo(disco *os.File, posicionActualInodo int64) {
	inodoAux := obtenerInodo(disco, posicionActualInodo)

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
	posicionActualInodo = 0
	posicionActualAVD = int64(super.InicioAVD)

	listado := strings.Split(path, "/")
	for i := 1; i < len(listado)-1; i++ {
		verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
		if existenciaAVD == false {
			fmt.Println("\033[1;31mAlguna de las carpetas padres a las cual desea acceder no existe\033[0m")
			return contenidoArchivo
		}
	}

	avdAux := obtenerAVD(disco, posicionActualAVD)

	if avdAux.ApuntadorDD == -1 {
		fmt.Println("\033[1;31mLa carpeta no tiene ningun archivo asociado\033[0m")
		return contenidoArchivo
	}
	busquedaArchivoDD(disco, avdAux.ApuntadorDD, listado[len(listado)-1])
	if posicionActualInodo == 0 {
		fmt.Println("\033[1;31mEl archivo no existe en la carpeta\033[0m")
		return contenidoArchivo
	}

	busquedaArchivoInodo(disco, posicionActualInodo)
	return contenidoArchivo
}

func cerrarSesion() {
	if loggeado == false {
		fmt.Println("\033[1;31mDebe de iniciar sesion para poder cerrala\033[0m")
		return
	}

	loggeado = false
	usuarioActual = ""
	grupoActual = ""
	contraActual = ""
	fmt.Println("\033[1;32mSe a cerrado la sesion\033[0m")
}

func iniciarSesion(id, usr, pwd string) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	if loggeado == true {
		fmt.Println("\033[1;31mYa existe una sesion iniciada en el sistema\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	lineas := strings.Split(obtenerContenidoArchivo(disco, "/users.txt"), "\n")

	banderaEncontado := false

	for i := 0; i < len(lineas)-1; i++ {
		linea := strings.Split(lineas[i], ",")
		if linea[0] != "0" {
			if linea[1] == "U" {
				if linea[3] == usr && linea[4] == pwd {
					usuarioActual = usr
					contraActual = pwd
					grupoActual = linea[2]
					banderaEncontado = true
					loggeado = true
				}
			}
		}
	}

	disco.Close()
	if banderaEncontado {
		fmt.Println("\033[1;32mA iniciado sesion exitosamente\033[0m")
	} else {
		fmt.Println("\033[1;31mSus credenciales no se han encontrado en el sistema\033[0m")
	}

}

func crearGrupo(id, name string) {

	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	if loggeado == false {
		fmt.Println("\033[1;31mDebe de iniciar sesion para poder realizar estas acciones\033[0m")
		return
	}

	if grupoActual != "root" {
		fmt.Println("\033[1;31mUnicamente los usaurios root pueden crear grupos\033[0m")
		return
	}

	if len(name) > 10 {
		fmt.Println("\033[1;31mNo pude sobrepasar el maximo de 10 caracteres para el nombre del grupo\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	contenidoUsuarios := obtenerContenidoArchivo(disco, "/users.txt")

	lineas := strings.Split(contenidoUsuarios, "\n")

	banderaEncontado := false

	noGrupo := ""

	for i := 0; i < len(lineas)-1; i++ {
		linea := strings.Split(lineas[i], ",")
		if linea[0] != "0" {
			noGrupo = linea[0]
			if linea[1] == "G" {
				if linea[2] == name {
					banderaEncontado = true
					break
				}
			}
		}
	}

	if banderaEncontado {
		fmt.Println("\033[1;31mEl grupo a crear ya existe dentro del sistema\033[0m")
		return
	}

	noGrupoConvertido, _ := strconv.Atoi(noGrupo)
	noGrupoConvertido++
	noGrupo = strconv.Itoa(noGrupoConvertido)
	contenidoUsuarios += noGrupo + ",G," + name + "\n"
	crearArchivo(id, "vacio", "/users.txt", contenidoUsuarios, int64(len(contenidoUsuarios)), 0)
	agregarLog(disco, "mkgrp", "0", "/users.txt", name, int64(len(name)))
	disco.Close()
	fmt.Println("\033[1;32mSe a agregado el nuevo grupo\033[0m")
}

func eliminarGrupo(id, name string) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	if loggeado == false {
		fmt.Println("\033[1;31mDebe de iniciar sesion para poder realizar estas acciones\033[0m")
		return
	}

	if grupoActual != "root" {
		fmt.Println("\033[1;31mUnicamente los usuarios root pueden eliminar grupos\033[0m")
		return
	}

	if name == "root" {
		fmt.Println("\033[1;31mEl grupo root no puede ser eliminado\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	contenidoUsuarios := obtenerContenidoArchivo(disco, "/users.txt")
	lineas := strings.Split(contenidoUsuarios, "\n")

	banderaEncontado := false

	contenidoAux := ""

	for i := 0; i < len(lineas)-1; i++ {
		linea := strings.Split(lineas[i], ",")
		if linea[0] != "0" {
			if linea[2] == name {
				linea[0] = "0"
				banderaEncontado = true
			}
		}
		if linea[1] == "U" {
			contenidoAux += linea[0] + "," + linea[1] + "," + linea[2] + "," + linea[3] + "," + linea[4] + "\n"
		} else {
			contenidoAux += linea[0] + "," + linea[1] + "," + linea[2] + "\n"
		}
	}

	if banderaEncontado == false {
		fmt.Println("\033[1;31mEl grupo a eliminar no se encuentra en el sistema\033[0m")
		return
	}

	crearArchivo(id, "vacio", "/users.txt", contenidoAux, int64(len(contenidoAux)), 0)
	agregarLog(disco, "rmgrp", "0", "/users.txt", name, int64(len(name)))
	disco.Close()
	fmt.Println("\033[1;32mSe a eliminado el grupo\033[0m")
}

func crearUsuario(id, usr, pwd, grupo string) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	if loggeado == false {
		fmt.Println("\033[1;31mDebe de iniciar sesion para poder realizar estas acciones\033[0m")
		return
	}

	if grupoActual != "root" {
		fmt.Println("\033[1;31mUnicamente los usuarios root pueden crear usuarios\033[0m")
		return
	}

	if len(grupo) > 10 || len(usr) > 10 || len(pwd) > 10 {
		fmt.Println("\033[1;31mNo pueden sobrebapasar el maximo de 10 caracteres para los parametros\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	contenidoUsuarios := obtenerContenidoArchivo(disco, "/users.txt")

	lineas := strings.Split(contenidoUsuarios, "\n")

	banderaEncontado := false

	banderaRepetido := false

	nuevo := ""

	contenidoAux := ""

	for i := 0; i < len(lineas)-1; i++ {
		linea := strings.Split(lineas[i], ",")
		contenidoAux += lineas[i] + "\n"
		if banderaEncontado == false {
			if linea[0] != "0" {
				if linea[1] == "G" {
					if banderaRepetido == false {
						if linea[2] == grupo {
							nuevo = linea[0] + ",U," + grupo + "," + usr + "," + pwd + "\n"
							contenidoAux += nuevo
							banderaEncontado = true
						}
					}
				} else {
					if linea[3] == usr {
						banderaRepetido = true
						break
					}
				}
			}
		}
	}

	if banderaRepetido {
		fmt.Println("\033[1;31mEl usuario ya se encuentra en el sistema\033[0m")
		return
	}

	if banderaEncontado == false {
		fmt.Println("\033[1;31mEl grupo no se encuentra en el sistema\033[0m")
		return
	}

	crearArchivo(id, "vacio", "/users.txt", contenidoAux, int64(len(contenidoAux)), 0)
	agregarLog(disco, "mkusr", "0", "/users.txt", nuevo, int64(len(nuevo)))
	disco.Close()
	fmt.Println("\033[1;32mSe a agregado el nuevo usuario\033[0m")
}

func eliminarUsuario(id, usr string) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	if loggeado == false {
		fmt.Println("\033[1;31mDebe de iniciar sesion para poder realizar estas acciones\033[0m")
		return
	}

	if grupoActual != "root" {
		fmt.Println("\033[1;31mUnicamente los usuarios root pueden eliminar usuarios\033[0m")
		return
	}

	if usr == "root" {
		fmt.Println("\033[1;31mEl usuario root no puede ser eliminado\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	contenidoUsuarios := obtenerContenidoArchivo(disco, "/users.txt")
	lineas := strings.Split(contenidoUsuarios, "\n")

	banderaEncontado := false

	contenidoAux := ""

	for i := 0; i < len(lineas)-1; i++ {
		linea := strings.Split(lineas[i], ",")
		if linea[0] != "0" {
			if linea[1] == "U" {
				if linea[3] == usr {
					linea[0] = "0"
					banderaEncontado = true
				}
			}
		}
		if linea[1] == "U" {
			contenidoAux += linea[0] + "," + linea[1] + "," + linea[2] + "," + linea[3] + "," + linea[4] + "\n"
		} else {
			contenidoAux += linea[0] + "," + linea[1] + "," + linea[2] + "\n"
		}
	}

	if banderaEncontado == false {
		fmt.Println("\033[1;31mEl usuario a eliminar no se encuentra en el sistema\033[0m")
		return
	}

	crearArchivo(id, "vacio", "/users.txt", contenidoAux, int64(len(contenidoAux)), 0)
	agregarLog(disco, "rmusr", "0", "/users.txt", usr, int64(len(usr)))
	disco.Close()
	fmt.Println("\033[1;32mSe a eliminado el usuario\033[0m")
}

func mostrarContenidoArchivo(id, path string) {
	contenidoArchivo = ""
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	obtenerContenidoArchivo(disco, path)

	fmt.Println("\033[1;33mEl contenido del archivo", path, "es: \033[0m")
	fmt.Println("\033[1;33m" + contenidoArchivo + "\033[0m")
	disco.Close()
}

func modificarContenidoArchivo(id, ruta, cont string, size int64) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
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

	posicionActualAVD = int64(super.InicioAVD)
	listado := strings.Split(rutaValidar, "/")
	for i := 1; i < len(listado); i++ {
		verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
		if existenciaAVD == false {
			fmt.Println("\033[1;31mAlguna de las carpetas padre no existe\033[0m")
			disco.Close()
			return
		}
	}

	avdAux := obtenerAVD(disco, posicionActualAVD)

	if avdAux.ApuntadorDD == -1 {
		fmt.Println("\033[1;31mLa carpeta no contiene ningun archivo\033[0m")
		disco.Close()
		return
	}

	posicionDD := avdAux.ApuntadorDD
	nombreAux := [20]byte{}
	copy(nombreAux[:], rutaCarpeta[len(rutaCarpeta)-1])
	banderaExistenciaArchivo := false

	for {
		ddAux := obtenerDD(disco, posicionDD)

		for i := 0; i < 5; i++ {
			if ddAux.ArregloArchivos[i].ApuntadorINodo != -1 {
				if ddAux.ArregloArchivos[i].NombreArchivo == nombreAux {
					borrarInodos(disco, ddAux.ArregloArchivos[i].ApuntadorINodo)
					internaLimpia := estructuraInterndaDD{ApuntadorINodo: -1}
					internaLimpia.NombreArchivo = [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
					ddAux.ArregloArchivos[i] = internaLimpia
					escribirDD(disco, posicionDD, ddAux)
					banderaExistenciaArchivo = true
					break
				}
			}
		}

		if ddAux.ApuntadorExtraDD == -1 {
			break
		}
		posicionDD = ddAux.ApuntadorExtraDD
	}

	if banderaExistenciaArchivo == false {
		fmt.Println("\033[1;31mLa carpeta no contiene el archivo a modificar\033[0m")
		return
	}

	ddMod := obtenerDD(disco, posicionDD)
	h := 0
	for h = 0; h < 5; h++ {
		if ddMod.ArregloArchivos[h].ApuntadorINodo == -1 {
			break
		}
	}

	if h == 5 {
		h = 4
	}

	noBloquesNecesarios := 1
	iterador := 0
	for k := 0; k < len(cont); k++ {
		if iterador == 25 {
			iterador = 0
			noBloquesNecesarios++
		}
		iterador++
	}

	noInodosNecesarios := 1
	iterador = 0
	for l := 0; l < noBloquesNecesarios; l++ {
		if iterador == 4 {
			iterador = 0
			noInodosNecesarios++
		}
		iterador++
	}

	if super.NoINodosLibres < uint32(noInodosNecesarios) {
		fmt.Println("\033[1;31mNo pueden crearse los inodos necesarios\033[0m")
		disco.Close()
		return
	}

	if super.NoBloquesLibres < uint32(noBloquesNecesarios) {
		fmt.Println("\033[1;31mNo pueden crearse los bloques necesarios\033[0m")
		disco.Close()
		return
	}

	inicioDivision := 0
	divisionContenido := make([]string, noBloquesNecesarios)
	for m := 0; m < noBloquesNecesarios; m++ {
		if m == noBloquesNecesarios-1 {
			divisionContenido[m] = cont[inicioDivision:len(cont)]
		} else {
			divisionContenido[m] = cont[inicioDivision : inicioDivision+25]
			inicioDivision += 25
		}
	}

	bitInodo := obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos)
	posicionINodo := int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
	interno := estructuraInterndaDD{}
	copy(interno.Creacion[:], obtenerFecha())
	copy(interno.Modificacion[:], obtenerFecha())
	copy(interno.NombreArchivo[:], rutaCarpeta[len(rutaCarpeta)-1])
	interno.ApuntadorINodo = posicionINodo
	ddMod.ArregloArchivos[h] = interno
	escribirDD(disco, posicionDD, ddMod)

	j := 0
	iterador = 0
	for i := 0; i < noInodosNecesarios; i++ {
		inodoAux := iNodo{}
		inodoAux.TamanioArchivo = uint32(len(cont))
		inodoAux.NoINodo = uint32(bitInodo)
		inodoAux.ApuntadroBloques = [4]int64{-1, -1, -1, -1}
		copy(inodoAux.Propietario[:], "root")
		copy(inodoAux.IDGrupo[:], "root")
		copy(inodoAux.Permiso[:], "664")
		for j < noBloquesNecesarios {
			if iterador == 4 {
				iterador = 0
				break
			}
			bitBloque := obtenerBitLibreBitMap(disco, super.InicioBitMapBloques, super.NoBloques)
			posicionBloque := int64(super.InicioBLoques) + (bitBloque * int64(unsafe.Sizeof(bloqueDatos{})))
			bloque := bloqueDatos{}
			copy(bloque.Contenido[:], divisionContenido[j])

			escribirBloque(disco, posicionBloque, bloque)

			inodoAux.ApuntadroBloques[iterador] = posicionBloque

			escribirEnBitMap(disco, 1, int64(super.InicioBitMapBloques)+bitBloque)

			j++
			super.NoBloquesLibres--
			iterador++
		}
		if iterador == 0 {
			inodoAux.ContadorBloquesAsignados = uint8(4)
		} else {
			inodoAux.ContadorBloquesAsignados = uint8(iterador)
		}

		escribirEnBitMap(disco, 1, int64(super.InicioBitMapINodo)+bitInodo)

		if i == noInodosNecesarios-1 {
			inodoAux.ApuntadorExtraINodo = -1
		} else {
			inodoAux.ApuntadorExtraINodo = int64(super.InicioINodo) + (obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos) * int64(unsafe.Sizeof(iNodo{})))
		}

		escribirInodo(disco, posicionINodo, inodoAux)

		bitInodo = obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos)
		posicionINodo = int64(super.InicioINodo) + (bitInodo * int64(unsafe.Sizeof(iNodo{})))
		super.NoINodosLibres--
	}

	super.BitLibreBloque = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapBloques, super.NoBloques))
	super.BitLibreINodo = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapINodo, super.NoINodos))
	super.BitLibreDD = uint32(obtenerBitLibreBitMap(disco, super.InicioBitMapDD, super.NoDD))

	escribirSuperBoot(disco, int64(inicio), super)
	escribirSuperBoot(disco, int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), super)
	agregarLog(disco, "edit", "0", ruta, contAux, size)
	disco.Close()

	fmt.Println("\033[1;32mSe a editado el archivo exitosamente\033[0m")
}

func modificarNombre(id, path, name string) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	posicionActualAVD = int64(super.InicioAVD)

	banderaSinCrear := false

	listado := strings.Split(path, "/")

	nombreLimpio := [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	if strings.Contains(listado[len(listado)-1], ".") {
		for i := 1; i < len(listado)-1; i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				}
			}
		}

		if banderaSinCrear {
			fmt.Println("\033[1;31mUna de las carpetas padre aun no se encuentra creada\033[0m")
			return
		}

		avdAux := obtenerAVD(disco, posicionActualAVD)

		if avdAux.ApuntadorDD == -1 {
			fmt.Println("\033[1;31mLa carpeta no contiene archivos\033[0m")
			return
		}

		posicionDD := avdAux.ApuntadorDD
		banderaEncontado := false

		nameAux := [20]byte{}
		copy(nameAux[:], listado[len(listado)-1])

		ddAux := detalleDirectorio{}

		for {
			ddAux = obtenerDD(disco, posicionDD)
			for i := 0; i < 5; i++ {
				if ddAux.ArregloArchivos[i].NombreArchivo == nameAux {
					banderaEncontado = true
					ddAux.ArregloArchivos[i].NombreArchivo = nombreLimpio
					copy(ddAux.ArregloArchivos[i].NombreArchivo[:], name)
					break
				}
			}
			if ddAux.ApuntadorExtraDD == -1 {
				break
			}
			posicionDD = ddAux.ApuntadorExtraDD
		}

		if banderaEncontado == false {
			fmt.Println("\033[1;31mEl archivo no se encuentra en la carpeta\033[0m")
			return
		}

		escribirDD(disco, posicionDD, ddAux)

		agregarLog(disco, "ren", "0", path, name, int64(len(name)))

		fmt.Println("\033[1;32mSe a modificado el nombre del archivo\033[0m")
	} else {
		for i := 1; i < len(listado); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				}
			}
		}

		if banderaSinCrear {
			fmt.Println("\033[1;31mUna de las carpetas aun no se encuentra creada\033[0m")
			return
		}

		avdAux := obtenerAVD(disco, posicionActualAVD)
		avdAux.Nombre = nombreLimpio
		copy(avdAux.Nombre[:], name)

		if avdAux.ApuntadoExtraAVD != -1 {
			posicionRecursiva := avdAux.ApuntadoExtraAVD
			for {
				avdRecursivo := obtenerAVD(disco, posicionRecursiva)
				avdRecursivo.Nombre = nombreLimpio
				copy(avdRecursivo.Nombre[:], name)
				escribirAVD(disco, posicionRecursiva, avdRecursivo)
				if avdRecursivo.ApuntadoExtraAVD == -1 {
					break
				}
				posicionRecursiva = avdAux.ApuntadoExtraAVD
			}
		}

		escribirAVD(disco, posicionActualAVD, avdAux)

		agregarLog(disco, "ren", "1", path, name, 1)
		disco.Close()
		fmt.Println("\033[1;32mSe a modificado el nombre de la carpeta exitosamente\033[0m")
	}
}

func eliminarArchivosOCarpetas(id, path, especial string) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	posicionActualAVD = int64(super.InicioAVD)

	banderaSinCrear := false

	listado := strings.Split(path, "/")

	if strings.Contains(listado[len(listado)-1], ".") {
		for i := 1; i < len(listado)-1; i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				}
			}
		}

		if banderaSinCrear {
			fmt.Println("\033[1;31mUna de las carpetas padres no existe\033[0m")
			return
		}

		avdAux := obtenerAVD(disco, posicionActualAVD)

		if avdAux.ApuntadorDD == -1 {
			fmt.Println("\033[1;31mLa carpeta no cuenta con ningun archivo\033[0m")
			return
		}

		nombreAux := [20]byte{}
		copy(nombreAux[:], listado[len(listado)-1])
		posicionDD := avdAux.ApuntadorDD
		banderaEncontrado := false
		i := 0
		for {
			ddAux := obtenerDD(disco, posicionDD)

			for i = 0; i < 5; i++ {
				if ddAux.ArregloArchivos[i].ApuntadorINodo != -1 {
					if ddAux.ArregloArchivos[i].NombreArchivo == nombreAux {
						banderaEncontrado = true
						break
					}
				}
			}

			if i == 5 {
				i = 4
			}

			if banderaEncontrado {
				break
			}
			if ddAux.ApuntadorExtraDD == -1 {
				break
			} else {
				posicionDD = ddAux.ApuntadorExtraDD
				i = 0
			}
		}

		if banderaEncontrado == false {
			fmt.Println("\033[1;31mEl archivo a elimianr no se encuentra en la ruta especificada\033[0m")
			return
		}

		ddAux := obtenerDD(disco, posicionDD)
		internoLimpio := estructuraInterndaDD{ApuntadorINodo: -1}
		copy(internoLimpio.NombreArchivo[:], "")
		posicionInodoBorrar := ddAux.ArregloArchivos[i].ApuntadorINodo
		ddAux.ArregloArchivos[i] = internoLimpio
		escribirDD(disco, posicionDD, ddAux)
		borrarInodos(disco, posicionInodoBorrar)
		agregarLog(disco, "rm", "0", path, "", 1)
		escribirSuperBoot(disco, int64(inicio), super)
		escribirSuperBoot(disco, int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), super)
		fmt.Println("\033[1;32mEl archivo se a eliminado exitosamente\033[0m")
		disco.Close()

	} else {
		for i := 1; i < len(listado)-1; i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
			if existenciaAVD == false {
				banderaSinCrear = true
				break
			}
		}

		if banderaSinCrear {
			fmt.Println("\033[1;31mUna de las carpetas aun no se encuentra creada\033[0m")
			return
		} else {
			nombreAux := [20]byte{}
			copy(nombreAux[:], listado[len(listado)-1])
			posicionEliminar := int64(0)
			for {
				avdAux := obtenerAVD(disco, posicionActualAVD)
				for i := 0; i < 6; i++ {
					if avdAux.SubDirectorios[i] != -1 {
						avdEliminar := obtenerAVD(disco, avdAux.SubDirectorios[i])
						if avdEliminar.Nombre == nombreAux {
							posicionEliminar = avdAux.SubDirectorios[i]
							avdAux.SubDirectorios[i] = -1
							escribirAVD(disco, posicionActualAVD, avdAux)
							break
						}
					}
				}

				if avdAux.ApuntadoExtraAVD == -1 {
					break
				}

				posicionActualAVD = avdAux.ApuntadoExtraAVD
			}

			if posicionEliminar == 0 {
				fmt.Println("\033[1;31mLa carpeta a eliminar no se encuentra en la ruta especificada\033[0m")
				return
			}

			borrarAVD(disco, posicionEliminar)

			agregarLog(disco, "rm", "1", path, "", 1)
			escribirSuperBoot(disco, int64(inicio), super)
			escribirSuperBoot(disco, int64(super.InicioLog)+(int64(super.NoAVD)*int64(unsafe.Sizeof(bitacora{}))), super)
			disco.Close()
			fmt.Println("\033[1;32mSe a eliminado la carpeta y todo su contenido exitosamente\033[0m")
		}
	}
}

func moverArchivoYCarpetas(id, path, destino string) {
	disco, _, inicio := obtenerDiscoMontado(id)

	if disco == nil {
		return
	}

	estado, _ := obtenerEstadoPerdida(id)

	if estado {
		fmt.Println("\033[1;31mLa particion presento una perdida\033[0m")
		return
	}

	super = obtenerSuperBoot(disco, int64(inicio))

	posicionActualAVD = int64(super.InicioAVD)

	banderaSinCrear := false

	listado := strings.Split(path, "/")

	if strings.Contains(listado[len(listado)-1], ".") {
		//se va a mover el archivo
		for i := 1; i < len(listado)-1; i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				}
			}
		}

		if banderaSinCrear {
			fmt.Println("\033[1;31mUna de las carpetas padre aun no se encuentra creada\033[0m")
			disco.Close()
			return
		}

		avdAux := obtenerAVD(disco, posicionActualAVD)

		if avdAux.ApuntadorDD == -1 {
			fmt.Println("\033[1;31mLa carpeta no tiene archivos\033[0m")
			disco.Close()
			return
		}

		banderaSinCrear = false
		posicionActualAVD = int64(super.InicioAVD)
		dest := strings.Split(destino, "/")
		for i := 1; i < len(dest); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, dest[i], 0)
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				}
			}
		}

		if banderaSinCrear {
			fmt.Println("\033[1;31mLa ruta a la que se desea mover el archivo no existe\033[0m")
			disco.Close()
			return
		}

		nombreAux := [20]byte{}
		copy(nombreAux[:], listado[len(listado)-1])

		posicionMover := int64(0)
		posicionDD := avdAux.ApuntadorDD
		for {
			ddAux := obtenerDD(disco, posicionDD)
			for i := 0; i < 5; i++ {
				if ddAux.ArregloArchivos[i].ApuntadorINodo != -1 {
					if ddAux.ArregloArchivos[i].NombreArchivo == nombreAux {
						posicionMover = ddAux.ArregloArchivos[i].ApuntadorINodo
						internoLimpio := estructuraInterndaDD{ApuntadorINodo: -1}
						copy(internoLimpio.NombreArchivo[:], "")
						ddAux.ArregloArchivos[i] = internoLimpio
						escribirDD(disco, posicionDD, ddAux)
						break
					}
				}
			}

			if ddAux.ApuntadorExtraDD == -1 {
				break
			}
			posicionDD = ddAux.ApuntadorExtraDD
		}

		if posicionMover == 0 {
			fmt.Println("\033[1;31mEl archivo no se encuentra en la ruta especificad\033[0m")
			disco.Close()
			return
		}

		avdMover := obtenerAVD(disco, posicionActualAVD)
		posicionDD = avdMover.ApuntadorDD

		if avdMover.ApuntadorDD == -1 {
			internoLimpio := estructuraInterndaDD{ApuntadorINodo: -1}
			copy(internoLimpio.NombreArchivo[:], "")
			ddNuevo := detalleDirectorio{ApuntadorExtraDD: -1}
			ddNuevo.ArregloArchivos[0] = internoLimpio
			ddNuevo.ArregloArchivos[1] = internoLimpio
			ddNuevo.ArregloArchivos[2] = internoLimpio
			ddNuevo.ArregloArchivos[3] = internoLimpio
			ddNuevo.ArregloArchivos[4] = internoLimpio
			posicionDD = int64(super.InicioDD) + (obtenerBitLibreBitMap(disco, super.InicioBitMapDD, super.NoDD) * int64(unsafe.Sizeof(detalleDirectorio{})))
			escribirDD(disco, posicionDD, ddNuevo)
		}

		for {
			ddLlenar := obtenerDD(disco, posicionDD)
			for i := 0; i < 5; i++ {
				if ddLlenar.ArregloArchivos[i].ApuntadorINodo == -1 {
					ddLlenar.ArregloArchivos[i].ApuntadorINodo = posicionMover
					ddLlenar.ArregloArchivos[i].NombreArchivo = nombreAux
					copy(ddLlenar.ArregloArchivos[i].Creacion[:], obtenerFecha())
					copy(ddLlenar.ArregloArchivos[i].Modificacion[:], obtenerFecha())
					escribirDD(disco, posicionDD, ddLlenar)
					break
				}
			}

			if ddLlenar.ApuntadorExtraDD == -1 {
				break
			}

			posicionDD = ddLlenar.ApuntadorExtraDD
		}
		//agregar a la bitacora
		agregarLog(disco, "mv", "0", path, destino, 1)
		disco.Close()
		fmt.Println("\033[1;32mSe a movido el archivo exitosamente\033[0m")
	} else {
		//se va a mover la carpeta
		for i := 1; i < len(listado)-1; i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, listado[i], 0)
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				}
			}
		}

		posicionPadre := posicionActualAVD
		if banderaSinCrear {
			fmt.Println("\033[1;31mUna de las carpetas padre aun no se encuentra creada\033[0m")
			disco.Close()
			return
		}

		banderaSinCrear = false
		posicionActualAVD = int64(super.InicioAVD)
		dest := strings.Split(destino, "/")
		for i := 1; i < len(dest); i++ {
			verificarExistenciaAVD(disco, posicionActualAVD, dest[i], 0)
			if existenciaAVD == false {
				if i != len(listado)-1 {
					banderaSinCrear = true
					break
				}
			}
		}

		if banderaSinCrear {
			fmt.Println("\033[1;31mLa ruta a la que se desea mover el archivo no existe\033[0m")
			disco.Close()
			return
		}

		nombreAux := [20]byte{}
		copy(nombreAux[:], listado[len(listado)-1])

		posicionMover := int64(0)

		for {
			avdAux := obtenerAVD(disco, posicionPadre)
			for i := 0; i < 6; i++ {
				if avdAux.SubDirectorios[i] != -1 {
					avdValidar := obtenerAVD(disco, avdAux.SubDirectorios[i])
					if avdValidar.Nombre == nombreAux {
						posicionMover = avdAux.SubDirectorios[i]
						avdAux.SubDirectorios[i] = -1
						escribirAVD(disco, posicionPadre, avdAux)
						break
					}
				}
			}

			if avdAux.ApuntadoExtraAVD == -1 {
				break
			}
			posicionPadre = avdAux.ApuntadoExtraAVD
		}

		if posicionMover == 0 {
			fmt.Println("\033[1;31mLa carpeta a mover no se encuentra en la ruta especificada\033[0m")
			disco.Close()
			return
		}

		for {
			avdMover := obtenerAVD(disco, posicionActualAVD)
			for i := 0; i < 6; i++ {
				if avdMover.SubDirectorios[i] == -1 {
					avdMover.SubDirectorios[i] = posicionMover
					escribirAVD(disco, posicionActualAVD, avdMover)
					break
				}
			}

			if avdMover.ApuntadoExtraAVD == -1 {
				break
			}
			posicionActualAVD = avdMover.ApuntadoExtraAVD
		}

		agregarLog(disco, "mv", "1", path, destino, 1)
		disco.Close()
		fmt.Println("\033[1;32mSe a movido la carpeta exitosamente\033[0m")

	}
}
