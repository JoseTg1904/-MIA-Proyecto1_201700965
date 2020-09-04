package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

type particion struct {
	Estado  byte
	Tipo    byte
	Ajuste  [2]byte
	Inicio  int64
	Tamanio int64
	Nombre  [16]byte
}

type logicaEBR struct {
	Estado    byte
	Ajuste    [2]byte
	Inicio    int64
	Tamanio   int64
	Siguiente int64
	Nombre    [16]byte
}

func crearParticion(size int64, unit, path, tipo, fit, name string) {

	archivo := obtenerDisco(path)
	defer archivo.Close()
	if archivo == nil {
		fmt.Println("El disco aun no a sido creado")
		return
	}

	mbrAux := obtenerMBR(archivo)
	particionVacia := particion{}

	contExtendidas := 0
	contTotales := 0
	banderaNombre := false

	nombre := [16]byte{}
	copy(nombre[:], name)

	for i := 0; i < 4; i++ {
		if particionAux := mbrAux.Particiones[i]; particionAux != particionVacia {
			if particionAux.Tipo == 'e' {
				contExtendidas++
			}
			if particionAux.Nombre == nombre {
				banderaNombre = true
			}
			contTotales++
		}
	}

	//falta validar el nombre de las logicas
	if banderaNombre {
		fmt.Println("Los nombres de las particiones no deben de repetirse")
		return
	}

	if contExtendidas == 1 && tipo == "e" {
		fmt.Println("Solo puede existir una particion extendida por disco")
		return
	}

	if contExtendidas == 0 && tipo == "l" {
		fmt.Println("No pueden existir particiones logicas si no hay una extendida")
		return
	}

	if contTotales == 4 && (tipo == "p" || tipo == "e") {
		fmt.Println("Se a excedido el maximo de particiones primarias o extendidas")
		return
	}

	size = obtenerTamanioParticion(size, unit)

	if size > (mbrAux.Tamanio - int64(unsafe.Sizeof(mbrAux))) {
		fmt.Println("El tamaño de la particion excede la capacidad del disco")
		return
	}

	if contTotales == 0 && (tipo == "e" || tipo == "p") {
		particionNueva := particion{Estado: '0',
			Tipo:    []byte(tipo)[0],
			Inicio:  int64(unsafe.Sizeof(mbrAux) + 1),
			Tamanio: size,
			Nombre:  nombre}
		copy(particionNueva.Ajuste[:], fit)
		mbrAux.Particiones[0] = particionNueva
		escribirEnDisco(archivo, mbrAux)
		fmt.Println("Particion creada con exito")
		return
	}

	banderaInsercion := false
	if tipo == "p" || tipo == "e" {
		particionNueva := particion{Estado: '0',
			Tipo:    []byte(tipo)[0],
			Tamanio: size,
			Nombre:  nombre}
		copy(particionNueva.Ajuste[:], fit)
		for i := 0; i < 4; i++ {
			if banderaInsercion == false {
				if mbrAux.Particiones[i] == particionVacia {
					if i == 0 {
						j := 1
						for j = 1; j < 4; j++ {
							if mbrAux.Particiones[j] != particionVacia {
								break
							} /* | 0 | | | |
							0	23	34	21	32
							*/
						}
						if mbrAux.Particiones[j] != particionVacia {
							if disponible := mbrAux.Particiones[j].Inicio - int64(unsafe.Sizeof(mbrAux)); disponible > size {
								particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
								mbrAux.Particiones[0] = particionNueva
								banderaInsercion = true
								break
							}
						} else {
							if disponible := mbrAux.Tamanio - int64(unsafe.Sizeof(mbrAux)); disponible > size {
								particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
								mbrAux.Particiones[0] = particionNueva
								banderaInsercion = true
								break
							}
						}
					} else {
						j := i
						for j = i; j < 4; j++ {
							if mbrAux.Particiones[j] != particionVacia {
								break
							}
						}
						k := i - 1
						for {
							if mbrAux.Particiones[k] != particionVacia {
								break
							}
							if k <= 0 {
								k = 0
								break
							}
							k--
						}

						if j == 4 {
							j = 3
						}

						if mbrAux.Particiones[j] != particionVacia && mbrAux.Particiones[k] != particionVacia {
							if disponible := mbrAux.Particiones[j].Inicio - (mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio); disponible > size {
								particionNueva.Inicio = mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio + int64(1)
								mbrAux.Particiones[i] = particionNueva
								banderaInsercion = true
								break
							}
						} else if mbrAux.Particiones[k] == particionVacia && mbrAux.Particiones[j] != particionVacia {
							if disponible := mbrAux.Particiones[j].Inicio - int64(unsafe.Sizeof(mbrAux)); disponible > size {
								particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
								mbrAux.Particiones[0] = particionNueva
								banderaInsercion = true
								break
							}
						} else if mbrAux.Particiones[k] != particionVacia && mbrAux.Particiones[j] == particionVacia {
							if disponible := mbrAux.Tamanio - (mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio); disponible > size {
								particionNueva.Inicio = mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio + int64(1)
								mbrAux.Particiones[i] = particionNueva
								banderaInsercion = true
								break
							}
						}
					}
				}
			}
		}

	} else {

		//insercion de particion logica
		/*
			inicio := int64(0)
			tamanio := int64(0)
			for i := 0; i < 4; i++ {
				if mbrAux.Particiones[i].Tipo == 'e' {
					inicio = mbrAux.Particiones[i].Inicio
					tamanio = mbrAux.Particiones[i].Tamanio
				}
			}

			fmt.Println(inicio, " ", tamanio)*/
	}

	if tipo == "e" && banderaInsercion == true {
		inicio := int64(0)
		for i := 0; i < 4; i++ {
			if mbrAux.Particiones[i].Tipo == 'e' {
				inicio = mbrAux.Particiones[i].Inicio
			}
		}
		ebr := logicaEBR{Estado: '0', Inicio: inicio, Tamanio: int64(0), Siguiente: -1}
		copy(ebr.Ajuste[:], fit)
		copy(ebr.Nombre[:], name)
		archivo.Seek(inicio, 0)
		buffer := bytes.NewBuffer([]byte{})
		binary.Write(buffer, binary.BigEndian, &ebr)
		archivo.Write(buffer.Bytes())
	}

	escribirEnDisco(archivo, mbrAux)
	fmt.Println("Particion creada")
}

func modificarTamanioParticion(size int64, unit, path, name string) {
	archivo := obtenerDisco(path)
	defer archivo.Close()
	if archivo == nil {
		fmt.Println("El disco aun no a sido creado")
		return
	}

	tamanioMod := obtenerTamanioParticion(size, unit)

	mbrAux := obtenerMBR(archivo)
	particionVacia := particion{}

	nombre := [16]byte{}
	copy(nombre[:], name)

	banderaNombre := false

	i := 0
	for i = 0; i < 4; i++ {
		if mbrAux.Particiones[i] != particionVacia {
			if mbrAux.Particiones[i].Nombre == nombre {
				banderaNombre = true
				break
			}
		}
	}

	if i == 4 {
		i = 3
	}

	banderaCambio := false

	if banderaNombre {
		if tamanioMod > 0 {
			if i == 3 {
				if disponible := mbrAux.Tamanio - (mbrAux.Particiones[i].Inicio + mbrAux.Particiones[i].Tamanio); disponible >= tamanioMod {
					mbrAux.Tamanio += tamanioMod
					banderaCambio = true
				} else {
					fmt.Println("La cantidad a agregar no puede ser aceptada por limitantes de espacio")
					return
				}
			} else {
				if disponible := mbrAux.Particiones[i+1].Inicio - (mbrAux.Particiones[i].Inicio + mbrAux.Particiones[i].Tamanio); disponible >= tamanioMod {
					mbrAux.Tamanio += tamanioMod
					banderaCambio = true
				} else {
					fmt.Println("La cantidad a agregar no puede ser aceptada por limitantes de espacio")
					return
				}
			}
		} else if tamanioMod < 0 {
			if (mbrAux.Particiones[i].Tamanio - tamanioMod) >= 0 {
				mbrAux.Tamanio -= tamanioMod
				banderaCambio = true
			} else {
				fmt.Println("La cantidad a quitar no puede ser aceptada, no pueden existir espacios negativos")
				return
			}
		}

	} else {
		//validar las logicas
	}

	if banderaCambio {
		escribirEnDisco(archivo, mbrAux)
		fmt.Println("EL tamaño de la particion a sido modificado")
	} else {

	}

}

func eliminarParticion(path, name, tipoELiminacion string) {

	archivo := obtenerDisco(path)
	defer archivo.Close()
	if archivo == nil {
		fmt.Println("El disco aun no a sido creado")
		return
	}

	mbrAux := obtenerMBR(archivo)
	particionVacia := particion{}

	banderaEncontrado := false

	nombre := [16]byte{}
	copy(nombre[:], name)

	i := 0
	for i = 0; i < 4; i++ {
		if mbrAux.Particiones[i] != particionVacia {
			if mbrAux.Particiones[i].Nombre == nombre {
				banderaEncontrado = true
				break
			}
		}
	}

	if i == 4 {
		i = 3
	}

	if banderaEncontrado {
		if tipoELiminacion == "full" {
			archivo.Seek(mbrAux.Particiones[i].Inicio, 0)
			buffer := bytes.NewBuffer([]byte{})
			binary.Write(buffer, binary.BigEndian, int8(0))
			for k := mbrAux.Particiones[i].Inicio; k <= (mbrAux.Particiones[i].Inicio + mbrAux.Particiones[i].Tamanio); k++ {
				archivo.Write(buffer.Bytes())
			}
			mbrAux.Particiones[i] = particionVacia
			fmt.Println("Eliminacion de particion completa, realizada con exito")
			return
		} else if tipoELiminacion == "fast" {
			mbrAux.Particiones[i] = particionVacia
			fmt.Println("Eliminacion de particion rapida, realizada con exito")
			return
		}
	} else {
		//buscar en las logicas
	}
}

func escribirEnDisco(archivo *os.File, mbr discoMBR) {
	archivo.Seek(0, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, &mbr)
	archivo.Write(buffer.Bytes())
}

func obtenerTamanioParticion(size int64, unit string) int64 {
	if unit == "b" {
		return size
	} else if unit == "k" {
		size = size * 1024
		return size
	} else if unit == "m" {
		size = size * 1048576
		return size
	}
	return int64(0)
}

func obtenerMBR(archivo *os.File) discoMBR {
	mbrAux := discoMBR{}
	contenido := make([]byte, int(unsafe.Sizeof(mbrAux)))
	archivo.Seek(0, 0)
	_, err := archivo.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &mbrAux)
	if a != nil {
	}
	return mbrAux
}

func obtenerDisco(path string) *os.File {
	if _, err := os.Stat(path); err == nil {
		archivo, _ := os.OpenFile(path, os.O_RDWR, 0644)
		return archivo
	}
	return nil
}
