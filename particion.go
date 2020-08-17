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
			contTotales += 1
		}
	}

	fmt.Println("particiones ", contTotales)

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

	if contTotales == 0 {
		particionNueva := particion{Estado: '0',
			Tipo:    []byte(tipo)[0],
			Inicio:  int64(unsafe.Sizeof(mbrAux) + 1),
			Tamanio: size}
		copy(particionNueva.Ajuste[:], fit)
		copy(particionNueva.Nombre[:], name)
		mbrAux.Particiones[0] = particionNueva
		escribirEnDisco(archivo, mbrAux)
		fmt.Println("Particion creada con exito")
		return
	}

	banderaExtendida := false
	if tipo == "e" {
		banderaExtendida = true
	}
	banderaInsercion := false
	if tipo == "p" || tipo == "e" {
		particionNueva := particion{Estado: '0',
			Tipo:    []byte(tipo)[0],
			Tamanio: size}
		copy(particionNueva.Ajuste[:], fit)
		copy(particionNueva.Nombre[:], name)
		for i := 0; i < 4; i++ {
			if mbrAux.Particiones[i] == particionVacia {
				if i == 0 {
					j := 1
					for j = 1; j < 4; j++ {
						if mbrAux.Particiones[j] != particionVacia {
							break
						}
					}
					if disponible := mbrAux.Particiones[j].Inicio - int64(unsafe.Sizeof(mbrAux)); disponible > size {
						particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
						mbrAux.Particiones[0] = particionNueva
						banderaInsercion = true
						break
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

	} else {
		inicio := int64(0)
		tamanio := int64(0)
		for i := 0; i < 4; i++ {
			if mbrAux.Particiones[i].Tipo == 'e' {
				inicio = mbrAux.Particiones[i].Inicio
				tamanio = mbrAux.Particiones[i].Tamanio
			}
		}

		fmt.Println(inicio, " ", tamanio)
	}

	if banderaExtendida == true && banderaInsercion == true {
		inicio := int64(0)
		for i := 0; i < 4; i++ {
			if mbrAux.Particiones[i].Tipo == 'e' {
				inicio = mbrAux.Particiones[i].Inicio
			}
		}
		ebr := logicaEBR{Estado: '0', Inicio: inicio, Tamanio: size, Siguiente: -1}
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
	return int64(-1)
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
