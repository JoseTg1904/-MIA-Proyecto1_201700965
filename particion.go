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
		fmt.Println("\033[1;31mEl disco aun no a sido creado\033[0m")
		return
	}

	mbrAux := obtenerMBR(archivo)

	contExtendidas := 0
	contTotales := 0
	banderaNombre := false

	nombre := [16]byte{}
	copy(nombre[:], name)
	var tipoE byte
	tipoE = "e"[0]

	for i := 0; i < 4; i++ {
		if mbrAux.Particiones[i].Inicio != -1 {
			if mbrAux.Particiones[i].Tipo == tipoE {
				contExtendidas++
			}
			if mbrAux.Particiones[i].Nombre == nombre {
				banderaNombre = true
			}
			contTotales++
		}
	}

	if contExtendidas == 1 {
		posicionActualEBR := int64(0)
		for i := 0; i < 4; i++ {
			if particionAux := mbrAux.Particiones[i]; particionAux.Inicio != -1 {
				if particionAux.Tipo == 'e' {
					posicionActualEBR = mbrAux.Particiones[i].Inicio
				}
			}
		}

		for {
			ebrAux := obtnerEBR(archivo, posicionActualEBR)

			if ebrAux.Nombre == nombre {
				banderaNombre = true
				break
			}

			if ebrAux.Siguiente == -1 {
				break
			}
			posicionActualEBR = ebrAux.Siguiente
		}
	}

	if banderaNombre {
		fmt.Println("\033[1;31mLos nombres de las particiones no deben de repetirse\033[0m")
		return
	}

	if contExtendidas == 1 && tipo == "e" {
		fmt.Println("\033[1;31mSolo puede existir una particion extendida por disco\033[0m")
		return
	}

	if contExtendidas == 0 && tipo == "l" {
		fmt.Println("\033[1;31mNo pueden existir particiones logicas si no hay una extendida\033[0m")
		return
	}

	if contTotales == 4 && (tipo == "p" || tipo == "e") {
		fmt.Println("\033[1;31mSe a excedida el maximo de particiones primarias o extendidas\033[0m")
		return
	}

	size = obtenerTamanioParticion(size, unit)

	if size > (mbrAux.Tamanio - int64(unsafe.Sizeof(mbrAux))) {
		fmt.Println("\033[1;31mEl tamaño de la particion excede la capacidad del disco\033[0m")
		return
	}

	if contTotales == 0 && (tipo == "e" || tipo == "p") {
		particionNueva := particion{Estado: '0',
			Tipo:    tipo[0],
			Inicio:  int64(unsafe.Sizeof(mbrAux) + 1),
			Tamanio: size,
			Nombre:  nombre}
		copy(particionNueva.Ajuste[:], fit)
		mbrAux.Particiones[0] = particionNueva
		escribirEnDisco(archivo, mbrAux)
		fmt.Println("\033[1;32mParticion creada con exito\033[0m")
		return
	}

	banderaInsercion := false
	if tipo == "p" || tipo == "e" {
		particionNueva := particion{Estado: '0',
			Tipo:    tipo[0],
			Tamanio: size,
			Nombre:  nombre}
		copy(particionNueva.Ajuste[:], fit)
		i := 0
		for i = 0; i < 4; i++ {
			if mbrAux.Particiones[i].Inicio == -1 {
				if i == 0 {
					j := 1
					for j = 1; j < 4; j++ {
						if mbrAux.Particiones[j].Inicio != -1 {
							break
						}
					}
					if mbrAux.Particiones[j].Inicio != -1 {
						if disponible := mbrAux.Particiones[j].Inicio - int64(unsafe.Sizeof(mbrAux)); disponible >= size {
							particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
							mbrAux.Particiones[0] = particionNueva
							banderaInsercion = true
							escribirEnDisco(archivo, mbrAux)
							break
						}
					} else {
						if disponible := mbrAux.Tamanio - int64(unsafe.Sizeof(mbrAux)); disponible >= size {
							particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
							mbrAux.Particiones[0] = particionNueva
							escribirEnDisco(archivo, mbrAux)
							banderaInsercion = true
							break
						}
					}
				} else {
					j := i
					for j = i; j < 4; j++ {
						if mbrAux.Particiones[j].Inicio != -1 {
							break
						}
					}
					k := i - 1
					for {
						if mbrAux.Particiones[k].Inicio != -1 {
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

					if mbrAux.Particiones[j].Inicio != -1 && mbrAux.Particiones[k].Inicio != -1 {
						if disponible := mbrAux.Particiones[j].Inicio - (mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio); disponible >= size {
							particionNueva.Inicio = mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio + int64(1)
							mbrAux.Particiones[i] = particionNueva
							banderaInsercion = true
							escribirEnDisco(archivo, mbrAux)
							break
						}
					} else if mbrAux.Particiones[k].Inicio == -1 && mbrAux.Particiones[j].Inicio != -1 {
						if disponible := mbrAux.Particiones[j].Inicio - int64(unsafe.Sizeof(mbrAux)); disponible >= size {
							particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
							mbrAux.Particiones[0] = particionNueva
							banderaInsercion = true
							escribirEnDisco(archivo, mbrAux)
							break
						}
					} else if mbrAux.Particiones[k].Inicio != -1 && mbrAux.Particiones[j].Inicio == -1 {
						if disponible := mbrAux.Tamanio - (mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio); disponible >= size {
							particionNueva.Inicio = mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio + int64(1)
							mbrAux.Particiones[i] = particionNueva
							banderaInsercion = true
							escribirEnDisco(archivo, mbrAux)
							break
						}
					}
				}
			}
		}
	} else {
		//insercion de particion logica
		mbrAux := obtenerMBR(archivo)

		inicio := int64(0)
		tamanio := int64(0)

		for i := 0; i < 4; i++ {
			if mbrAux.Particiones[i].Inicio != -1 {
				if mbrAux.Particiones[i].Tipo == 'e' {
					inicio = mbrAux.Particiones[i].Inicio
					tamanio = mbrAux.Particiones[i].Inicio
					break
				}
			}
		}

		tamanio -= int64(unsafe.Sizeof(logicaEBR{}))

		if tamanio < (size + int64(unsafe.Sizeof(logicaEBR{}))) {
			fmt.Println("\033[1;31mEl tamaño de la particion sobrepasa el de la particon extendida\033[0m")
			return
		}

		ebrValidar := obtnerEBR(archivo, inicio)
		if ebrValidar.Tamanio == -1 {
			copy(ebrValidar.Ajuste[:], fit)
			copy(ebrValidar.Nombre[:], name)
			ebrValidar.Tamanio = size

			posicionSiguiente := int64(unsafe.Sizeof(logicaEBR{})+1) + inicio + size
			ebrValidar.Siguiente = posicionSiguiente

			ebrAdjunta := logicaEBR{Estado: '0', Inicio: posicionSiguiente, Tamanio: int64(-1), Siguiente: -1}
			copy(ebrAdjunta.Ajuste[:], "")
			copy(ebrAdjunta.Nombre[:], "")
			escrbirEBR(archivo, inicio, ebrValidar)
			escrbirEBR(archivo, posicionSiguiente, ebrAdjunta)

			fmt.Println("\033[1;32mSe a insertado la particion logica\033[0m")
			return
		} else {
			banderaInsercion := false
			tamanioEBR := int64(unsafe.Sizeof(logicaEBR{}))
			posicionActualEBR := inicio
			posicionSiguienteEBR := ebrValidar.Siguiente
			for {
				ebrActual := obtnerEBR(archivo, posicionActualEBR)
				ebrSiguiente := obtnerEBR(archivo, posicionSiguienteEBR)

				if ebrSiguiente.Tamanio == -1 {
					if tamanio >= (size + tamanioEBR) {
						banderaInsercion = true
						posicionSiguiente := tamanioEBR + 1 + size + ebrSiguiente.Inicio
						ebrSiguiente.Siguiente = posicionSiguiente
						ebrSiguiente.Tamanio = size
						ebrSiguiente.Nombre = [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
						ebrSiguiente.Ajuste = [2]byte{0, 0}
						copy(ebrSiguiente.Ajuste[:], fit)
						copy(ebrSiguiente.Nombre[:], name)
						ebrAdjunta := logicaEBR{Estado: '0', Inicio: posicionSiguiente, Tamanio: int64(-1), Siguiente: -1}
						copy(ebrAdjunta.Ajuste[:], "")
						copy(ebrAdjunta.Nombre[:], "")
						escrbirEBR(archivo, ebrSiguiente.Inicio, ebrSiguiente)
						escrbirEBR(archivo, posicionSiguiente, ebrAdjunta)
						break
					} else {
						break
					}
				}
				if val := ebrSiguiente.Inicio - (ebrActual.Inicio + tamanioEBR + ebrActual.Tamanio); val >= (size + tamanioEBR) {
					banderaInsercion = true
					posicionSiguiente := tamanioEBR + 1 + ebrActual.Tamanio + ebrActual.Inicio
					ebrActual.Siguiente = posicionSiguiente
					ebrAdjunta := logicaEBR{Estado: '0', Inicio: posicionSiguiente, Tamanio: size, Siguiente: ebrSiguiente.Inicio}
					copy(ebrAdjunta.Ajuste[:], fit)
					copy(ebrAdjunta.Nombre[:], name)
					escrbirEBR(archivo, ebrActual.Inicio, ebrActual)
					escrbirEBR(archivo, posicionSiguiente, ebrAdjunta)
					break
				}
				posicionActualEBR = posicionSiguienteEBR
				posicionSiguienteEBR = ebrSiguiente.Siguiente
			}

			if banderaInsercion == false {
				fmt.Println("\033[1;31mLa particion no pudo ser creada por razones de espacio\033[0m")
				return
			}

			fmt.Println("\033[1;32mSe a creado la particion logica\033[0m")
			return
		}
	}

	if tipo == "e" && banderaInsercion == true {
		inicio := int64(0)
		for i := 0; i < 4; i++ {
			if mbrAux.Particiones[i].Inicio != -1 {
				if mbrAux.Particiones[i].Tipo == 'e' {
					inicio = mbrAux.Particiones[i].Inicio
					break
				}
			}
		}
		ebr := logicaEBR{Estado: '0', Inicio: inicio, Tamanio: int64(-1), Siguiente: -1}
		copy(ebr.Ajuste[:], "")
		copy(ebr.Nombre[:], "")
		escribirEnDisco(archivo, mbrAux)
		escrbirEBR(archivo, inicio, ebr)
	}

	if banderaInsercion == false {
		fmt.Println("\033[1;31mLa particion no pudo ser creada\033[0m")
		return
	}

	fmt.Println("\033[1;32mSe a creado la particion exitosamente\033[0m")
}

func modificarTamanioParticion(size int64, unit, path, name string) {
	archivo := obtenerDisco(path)
	defer archivo.Close()
	if archivo == nil {
		fmt.Println("\033[1;31mEl disco aun no a sido creado\033[0m")
		return
	}

	tamanioMod := obtenerTamanioParticion(size, unit)

	mbrAux := obtenerMBR(archivo)

	nombre := [16]byte{}
	copy(nombre[:], name)

	banderaNombre := false

	i := 0
	for i = 0; i < 4; i++ {
		if mbrAux.Particiones[i].Inicio != -1 {
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
					fmt.Println("\033[1;31mLa cantida a agregar no puede ser aceptada por limitantes de espacio\033[0m")
					return
				}
			} else {

				j := i
				for j = i; j < 4; j++ {
					if mbrAux.Particiones[j].Inicio != -1 {
						break
					}
				}

				if j == 4 {
					j = 3
				}

				if mbrAux.Particiones[j].Inicio != -1 {
					if disponible := mbrAux.Particiones[j].Inicio - (mbrAux.Particiones[i].Inicio + mbrAux.Particiones[i].Tamanio); disponible >= tamanioMod {
						mbrAux.Tamanio += tamanioMod
						banderaCambio = true
					} else {
						fmt.Println("\033[1;31mLa cantidad a agregar no puede ser aceptada por limitantes de espacio\033[0m")
						return
					}
				} else {
					if disponible := mbrAux.Tamanio - (mbrAux.Particiones[i].Inicio + mbrAux.Particiones[i].Tamanio); disponible >= tamanioMod {
						mbrAux.Tamanio += tamanioMod
						banderaCambio = true
					} else {
						fmt.Println("\033[1;31mLa cantidad a agregar no puede ser aceptada por limitantes de espacio\033[0m")
						return
					}
				}
			}
		} else if tamanioMod < 0 {
			if (mbrAux.Particiones[i].Tamanio - tamanioMod) >= 0 {
				mbrAux.Tamanio -= tamanioMod
				banderaCambio = true
			} else {
				fmt.Println("\033[1;31mLa cantidad a quitar no puede ser aceptada, no pueden existir espacios negativos\033[0m")
				return
			}
		}

	} else {
		posicionActualEBR := int64(0)
		tamanioE := int64(0)
		for i = 0; i < 4; i++ {
			if mbrAux.Particiones[i].Inicio != -1 {
				if mbrAux.Particiones[i].Tipo == 'e' {
					posicionActualEBR = mbrAux.Particiones[i].Inicio
					tamanioE = mbrAux.Particiones[i].Tamanio
					break
				}
			}
		}

		if posicionActualEBR == 0 {
			fmt.Println("\033[1;31mLa particion a modificar no se encuentra en el sistema\033[0m")
			return
		}

		for {
			ebrAux := obtnerEBR(archivo, posicionActualEBR)

			if ebrAux.Nombre == nombre {
				banderaNombre = true
				break
			}

			if ebrAux.Siguiente == -1 {
				break
			}
			posicionActualEBR = ebrAux.Siguiente
		}

		if banderaNombre {
			ebrMod := obtnerEBR(archivo, posicionActualEBR)
			if tamanioMod > 0 {
				if ebrMod.Siguiente == -1 {
					if disponible := tamanioE - (ebrMod.Tamanio + ebrMod.Inicio + int64(unsafe.Sizeof(logicaEBR{}))); disponible >= tamanioMod {
						ebrMod.Tamanio += tamanioMod
						banderaCambio = true
						escrbirEBR(archivo, posicionActualEBR, ebrMod)
					} else {
						fmt.Println("\033[1;31mLa cantidad a agregar no puede ser aceptada por limitantes de espacio\033[0m")
						return
					}
				} else {
					ebrSiguiente := obtnerEBR(archivo, ebrMod.Siguiente)
					if disponible := ebrSiguiente.Inicio - (ebrMod.Inicio + ebrMod.Tamanio + int64(unsafe.Sizeof(logicaEBR{}))); disponible >= tamanioMod {
						ebrMod.Tamanio += tamanioMod
						banderaCambio = true
						escrbirEBR(archivo, posicionActualEBR, ebrMod)
					} else {
						fmt.Println("\033[1;31mLa cantidad a agregar no puede ser aceptada por limitantes de espacio\033[0m")
						return
					}
				}
			} else if tamanioMod < 0 {
				ebrMod := obtnerEBR(archivo, posicionActualEBR)
				if (ebrMod.Tamanio - tamanioMod) >= 0 {
					ebrMod.Tamanio -= tamanioMod
					banderaCambio = true
					escrbirEBR(archivo, posicionActualEBR, ebrMod)
				} else {
					fmt.Println("\033[1;31mLa cantidad a quitar no puede ser aceptada, no pueden existir espacion negativos\033[0m")
					return
				}
			}
		} else {
			fmt.Println("\033[1;31mLa particion a modificar no se encuentra en el sistema\033[0m")
			return
		}
	}

	if banderaCambio {
		escribirEnDisco(archivo, mbrAux)
		fmt.Println("\033[1;32mEl tamaño de la particion a sido modificado\033[0m")
	} else {
		fmt.Println("\033[1;31mNo se a podido modificar el tamaño de la particion\033[0m")
	}

}

func eliminarParticion(path, name, tipoELiminacion string) {

	archivo := obtenerDisco(path)
	defer archivo.Close()
	if archivo == nil {
		fmt.Println("\033[1;31mEl disco aun no a sido creado\033[0m")
		return
	}

	mbrAux := obtenerMBR(archivo)

	banderaEncontrado := false

	nombre := [16]byte{}
	copy(nombre[:], name)

	i := 0
	for i = 0; i < 4; i++ {
		if mbrAux.Particiones[i].Inicio != -1 {
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
			mbrAux.Particiones[i] = particion{Inicio: -1}
			desmontarParticionEliminada(path, name)
			escribirEnDisco(archivo, mbrAux)
			fmt.Println("\033[1;32mEliminacion de particion completa, realizada con exito\033[0m")
			return
		} else if tipoELiminacion == "fast" {
			mbrAux.Particiones[i] = particion{Inicio: -1}
			desmontarParticionEliminada(path, name)
			escribirEnDisco(archivo, mbrAux)
			fmt.Println("\033[1;32mEliminacion de particion rapida, realizada con exito\033[0m")
			return
		}
	} else {
		posicionActualEBR := int64(0)
		inicioExtendida := int64(0)

		for i := 0; i < 4; i++ {
			if particionAux := mbrAux.Particiones[i]; particionAux.Inicio != -1 {
				if particionAux.Tipo == 'e' {
					posicionActualEBR = mbrAux.Particiones[i].Inicio
					inicioExtendida = posicionActualEBR
				}
			}
		}

		if posicionActualEBR == 0 {
			fmt.Println("\033[1;31mLa particion a eliminar no se encuentra en el sistema\033[0m")
			return
		}

		banderaNombre := false
		for {
			ebrAux := obtnerEBR(archivo, posicionActualEBR)

			if ebrAux.Nombre == nombre {
				banderaNombre = true
				break
			}

			if ebrAux.Siguiente == -1 {
				break
			}
			posicionActualEBR = ebrAux.Siguiente
		}

		if banderaNombre {
			if tipoELiminacion == "full" {
				if posicionActualEBR == inicioExtendida {
					ebrEliminar := obtnerEBR(archivo, posicionActualEBR)
					buffer := bytes.NewBuffer([]byte{})
					binary.Write(buffer, binary.BigEndian, int8(0))
					for k := ebrEliminar.Inicio + int64(unsafe.Sizeof(logicaEBR{})); k <= (ebrEliminar.Tamanio); k++ {
						archivo.Write(buffer.Bytes())
					}
					ebrEliminar.Tamanio = -1
					ebrEliminar.Nombre = [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
					ebrEliminar.Ajuste = [2]byte{0, 0}
					escrbirEBR(archivo, posicionActualEBR, ebrEliminar)
					desmontarParticionEliminada(path, name)
					fmt.Println("\033[1;32mEliminacion de particion completa, realizada con exito\033[0m")
				} else {
					posicionSiguiente := int64(0)
					for {
						ebrAux := obtnerEBR(archivo, inicioExtendida)
						posicionSiguiente = ebrAux.Siguiente
						ebrSiguiente := obtnerEBR(archivo, posicionSiguiente)
						if ebrSiguiente.Nombre == nombre {
							banderaNombre = true
							break
						}
						inicioExtendida = ebrAux.Siguiente
					}
					ebrEliminar := obtnerEBR(archivo, posicionActualEBR)
					ebrAnterior := obtnerEBR(archivo, inicioExtendida)
					ebrAnterior.Siguiente = ebrEliminar.Siguiente
					escrbirEBR(archivo, inicioExtendida, ebrAnterior)
					buffer := bytes.NewBuffer([]byte{})
					binary.Write(buffer, binary.BigEndian, int8(0))
					for k := ebrEliminar.Inicio; k <= (ebrEliminar.Tamanio); k++ {
						archivo.Write(buffer.Bytes())
					}
					desmontarParticionEliminada(path, name)
					fmt.Println("\033[1;32mEliminacion de particion completa, realzada con exito\033[0m")
					return
				}
			} else if tipoELiminacion == "fast" {
				if posicionActualEBR == inicioExtendida {
					ebrEliminar := obtnerEBR(archivo, posicionActualEBR)
					buffer := bytes.NewBuffer([]byte{})
					binary.Write(buffer, binary.BigEndian, int8(0))
					for k := ebrEliminar.Inicio + int64(unsafe.Sizeof(logicaEBR{})); k <= (ebrEliminar.Tamanio); k++ {
						archivo.Write(buffer.Bytes())
					}
					ebrEliminar.Tamanio = -1
					ebrEliminar.Nombre = [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
					ebrEliminar.Ajuste = [2]byte{0, 0}
					escrbirEBR(archivo, posicionActualEBR, ebrEliminar)
					desmontarParticionEliminada(path, name)
					fmt.Println("\033[1;32mEliminacion de particion rapida, realizada con exito\033[0m")
				} else {
					posicionSiguiente := int64(0)
					for {
						ebrAux := obtnerEBR(archivo, inicioExtendida)
						posicionSiguiente = ebrAux.Siguiente
						ebrSiguiente := obtnerEBR(archivo, posicionSiguiente)
						if ebrSiguiente.Nombre == nombre {
							banderaNombre = true
							break
						}
						inicioExtendida = ebrAux.Siguiente
					}
					ebrEliminar := obtnerEBR(archivo, posicionActualEBR)
					ebrAnterior := obtnerEBR(archivo, inicioExtendida)
					ebrAnterior.Siguiente = ebrEliminar.Siguiente
					escrbirEBR(archivo, inicioExtendida, ebrAnterior)
					buffer := bytes.NewBuffer([]byte{})
					binary.Write(buffer, binary.BigEndian, int8(0))
					for k := ebrEliminar.Inicio; k <= (ebrEliminar.Tamanio); k++ {
						archivo.Write(buffer.Bytes())
					}
					desmontarParticionEliminada(path, name)
					fmt.Println("\033[1;32mEliminacion de particion rapida, relizada con exito\033[0m")
					return
				}
			}
		} else {
			fmt.Println("\033[1;31mLa particion no se a encontrado en el sistema\033[0m")
			return
		}
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

func escrbirEBR(disco *os.File, poicionEBR int64, ebr logicaEBR) {
	disco.Seek(poicionEBR, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, &ebr)
	disco.Write(buffer.Bytes())
}

func obtnerEBR(disco *os.File, poisicionEBR int64) logicaEBR {
	ebrAux := logicaEBR{}
	contenido := make([]byte, int(unsafe.Sizeof(ebrAux)))
	disco.Seek(poisicionEBR, 0)
	disco.Read(contenido)
	buffer := bytes.NewBuffer(contenido)
	binary.Read(buffer, binary.BigEndian, &ebrAux)
	return ebrAux
}
