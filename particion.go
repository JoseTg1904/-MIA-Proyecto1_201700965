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
	//obtencion del archivo que simula el disco
	archivo := obtenerDisco(path)
	defer archivo.Close()

	//verificacion de existencia del archivo
	if archivo == nil {
		fmt.Println("\033[1;31mEl disco aun no a sido creado\033[0m")
		return
	}

	mbrAux := obtenerMBR(archivo)

	//variables para validar la cantidad de particiones extendidas, particiones totales,
	//el inicio de las particiones extendidas si es que existe y
	//repeticion del nombre
	contExtendidas := 0
	contTotales := 0
	posicionActualEBR := int64(0)
	banderaNombre := false

	//convertiendo de string a un arreglo de bytes el nombre de la particion
	nombre := [16]byte{}
	copy(nombre[:], name)

	//validando la cantidad total de particiones, el nombre de la particion a crear y
	//la cantidad de extendidas
	for i := 0; i < 4; i++ {
		if mbrAux.Particiones[i].Inicio != -1 {
			if mbrAux.Particiones[i].Tipo == 'e' {
				contExtendidas++
				posicionActualEBR = mbrAux.Particiones[i].Inicio
			}
			if mbrAux.Particiones[i].Nombre == nombre {
				banderaNombre = true
			}
			contTotales++
		}
	}

	//validacion del nombre de la particion a insertar en las logicas
	if contExtendidas == 1 {
		_, banderaNombre = verificacionExistenciaLogica(posicionActualEBR, archivo, nombre)
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

	//variable para validar la insercion de la particion
	banderaInsercion := false

	//variable para validar que exista espacio para insertar la particion
	banderaEspacio := false

	//instancia de la particion a insertar
	particionNueva := particion{Estado: '0',
		Tipo:    tipo[0],
		Tamanio: size,
		Nombre:  nombre}
	copy(particionNueva.Ajuste[:], fit)

	//verificando si la particion a insertar es extendida o primaria
	if tipo == "p" || tipo == "e" {
		i := 0
		for i = 0; i < 4; i++ {
			//verificando que espacio se encuentra vacio en la tabla de particiones del mbr
			if mbrAux.Particiones[i].Inicio == -1 {
				if i == 0 {
					//buscando una particion contigua a la derecha
					j := 1
					for j = 1; j < 4; j++ {
						if mbrAux.Particiones[j].Inicio != -1 {
							break
						}
					}
					if j == 4 {
						j = 3
					}

					if mbrAux.Particiones[j].Inicio != -1 {
						//verificando que exista espacio suficiente entre el inicio del disco
						//y la particion contigua a la derecha encontrada
						if disponible := mbrAux.Particiones[j].Inicio - int64(unsafe.Sizeof(mbrAux)); disponible >= size {
							banderaEspacio = true
						}
					} else {
						//verificando que exista espacio suficiente entre el inicio del disco
						//y el final del disco
						if disponible := mbrAux.Tamanio - int64(unsafe.Sizeof(mbrAux)); disponible >= size {
							banderaEspacio = true
						}
					}

					//si existe el espacio realizar la insercion de la nueva particion
					if banderaEspacio {
						particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
						mbrAux.Particiones[0] = particionNueva
						banderaInsercion = true
						break
					}
				} else {
					//busqueda de una particion contigua a la derecha
					j := i
					for j = i; j < 4; j++ {
						if mbrAux.Particiones[j].Inicio != -1 {
							break
						}
					}
					if j == 4 {
						j = 3
					}

					//busqueda de una particion contigua a la izquierda
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

					//verificando espacio libre si existen dos particiones contiguas
					if mbrAux.Particiones[j].Inicio != -1 && mbrAux.Particiones[k].Inicio != -1 {
						if disponible := mbrAux.Particiones[j].Inicio - (mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio); disponible >= size {
							particionNueva.Inicio = mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio + int64(1)
							mbrAux.Particiones[i] = particionNueva
							banderaInsercion = true
							break
						}
						//verificando espacio libre si solo existe una particion contigua a la derecha
					} else if mbrAux.Particiones[k].Inicio == -1 && mbrAux.Particiones[j].Inicio != -1 {
						if disponible := mbrAux.Particiones[j].Inicio - int64(unsafe.Sizeof(mbrAux)); disponible >= size {
							particionNueva.Inicio = int64(unsafe.Sizeof(mbrAux) + 1)
							mbrAux.Particiones[0] = particionNueva
							banderaInsercion = true
							break
						}
						//verificando si existe espacio libre si solo existe una particion contigua a la izquierda
					} else if mbrAux.Particiones[k].Inicio != -1 && mbrAux.Particiones[j].Inicio == -1 {
						if disponible := mbrAux.Tamanio - (mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio); disponible >= size {
							particionNueva.Inicio = mbrAux.Particiones[k].Inicio + mbrAux.Particiones[k].Tamanio + int64(1)
							mbrAux.Particiones[i] = particionNueva
							banderaInsercion = true
							break
						}
					}
				}
			}
		}

		//insercion de particion logica
	} else {
		//variables que almacenan el inicio y el tamaño de la particion extendida
		inicio, tamanio := obtenerInicioTamanioExtendida(mbrAux)

		//quitandole al tamaño total de la particion extendida el tamaño de un EBR
		tamanio -= int64(unsafe.Sizeof(logicaEBR{}))

		if tamanio < (size + int64(unsafe.Sizeof(logicaEBR{}))) {
			fmt.Println("\033[1;31mEl tamaño de la particion sobrepasa el de la particion extendida\033[0m")
			return
		}

		//obteniendo el ebr inicial para validar donde se puede insertar la particion nueva
		ebrValidar := obtnerEBR(archivo, inicio)

		if ebrValidar.Tamanio == -1 {
			//obtencion de la nueva posicion del ebr siguiente
			posicionSiguiente := int64(1) + inicio + size

			//llenado de los nuevos datos del ebr inicial
			copy(ebrValidar.Ajuste[:], fit)
			ebrValidar.Nombre = nombre
			ebrValidar.Tamanio = size
			ebrValidar.Siguiente = posicionSiguiente

			//instancia y llenado de datos del ebr adjunto a la crecion de una particion
			ebrAdjunta := logicaEBR{Estado: '0',
				Inicio:    posicionSiguiente,
				Tamanio:   int64(-1),
				Siguiente: int64(-1)}
			copy(ebrAdjunta.Ajuste[:], "")
			copy(ebrAdjunta.Nombre[:], "")

			escrbirEBR(archivo, inicio, ebrValidar)
			escrbirEBR(archivo, posicionSiguiente, ebrAdjunta)

			fmt.Println("\033[1;32mSe a insertado la particion logica\033[0m")
		} else {
			//variables que almacenan el tamaño de un ebr, la posicion del ebr actual y
			//la del siguiente ebr
			tamanioEBR := int64(unsafe.Sizeof(logicaEBR{}))
			posicionActualEBR := inicio
			posicionSiguienteEBR := ebrValidar.Siguiente

			//ciclo infinito para encontrar el espacio disponible para insertar la particion logica
			for {
				//obtencion del ebr actual y siguiente
				ebrActual := obtnerEBR(archivo, posicionActualEBR)
				ebrSiguiente := obtnerEBR(archivo, posicionSiguienteEBR)

				//verificando si el espacio entre dos ebr es suficiente para poder insertar la logica
				if val := ebrSiguiente.Inicio - (ebrActual.Inicio + tamanioEBR + ebrActual.Tamanio); val >= (size + tamanioEBR) {
					//posicion del nuevo ebr
					posicionSiguiente := int64(1) + ebrActual.Tamanio + ebrActual.Inicio

					//apuntando el nuevo actual hacia el nuevo ebr
					ebrActual.Siguiente = posicionSiguiente

					//instancia y llenado de datos del nuevo ebr
					ebrAdjunta := logicaEBR{Estado: '0',
						Inicio:    posicionSiguiente,
						Tamanio:   size,
						Siguiente: ebrSiguiente.Inicio,
						Nombre:    nombre}
					copy(ebrAdjunta.Ajuste[:], fit)

					//escritura del ebr modificado y el nuevo
					escrbirEBR(archivo, ebrActual.Inicio, ebrActual)
					escrbirEBR(archivo, posicionSiguiente, ebrAdjunta)

					banderaInsercion = true
					break
				}

				//verificando si el espacio entre el ultimo ebr y el final de la particion extendida es
				//suficiente para insertar la logica
				if ebrSiguiente.Tamanio == -1 {
					if val := tamanio - (ebrSiguiente.Inicio + tamanioEBR); val >= (size + tamanioEBR) {
						//calculando la nueva posicion del ebr siguiente
						posicionSiguiente := int64(1) + size + ebrSiguiente.Inicio

						//llenado de los nuevos datos del ebr
						ebrSiguiente.Siguiente = posicionSiguiente
						ebrSiguiente.Tamanio = size
						ebrSiguiente.Nombre = nombre
						copy(ebrSiguiente.Ajuste[:], fit)

						//instancia y llenado de los datos del ebr adjunto
						ebrAdjunta := logicaEBR{Estado: '0',
							Inicio:    posicionSiguiente,
							Tamanio:   int64(-1),
							Siguiente: int64(-1)}
						copy(ebrAdjunta.Ajuste[:], "")
						copy(ebrAdjunta.Nombre[:], "")

						escrbirEBR(archivo, ebrSiguiente.Inicio, ebrSiguiente)
						escrbirEBR(archivo, posicionSiguiente, ebrAdjunta)

						banderaInsercion = true
						break
					} else {
						break
					}
				}

				//movimiento de las posiciones de los ebr para modificar los ebr a analizar
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

	//creacion del ebr adjunto a la creacion de la particion extendida
	if tipo == "e" && banderaInsercion == true {
		//instancia y llenado de datos del ebr adjunto
		ebr := logicaEBR{Estado: '0',
			Inicio:    posicionActualEBR,
			Tamanio:   int64(-1),
			Siguiente: int64(-1)}
		copy(ebr.Ajuste[:], "")
		copy(ebr.Nombre[:], "")

		escrbirEBR(archivo, posicionActualEBR, ebr)
	}

	//validacion de la insercion
	if banderaInsercion == false {
		fmt.Println("\033[1;31mLa particion no pudo ser creada\033[0m")
		return
	}

	escribirEnDisco(archivo, mbrAux)

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

	banderaNombre, i := verificacionExistenciaPrimariaExtendida(mbrAux, nombre)

	banderaCambio := false

	if banderaNombre {
		if tamanioMod > 0 {
			//buscando la particion contigua a la derecha de la que se busca aumentar su tamaño
			j := i
			for j = i; j < 4; j++ {
				if mbrAux.Particiones[j].Inicio != -1 {
					break
				}
			}
			if j == 4 {
				j = 3
			}

			//verificacion del espacio disponible con la particion contigua
			if mbrAux.Particiones[j].Inicio != -1 && (j != i) {
				if disponible := mbrAux.Particiones[j].Inicio - (mbrAux.Particiones[i].Inicio + mbrAux.Particiones[i].Tamanio); disponible >= tamanioMod {
					mbrAux.Particiones[i].Tamanio += tamanioMod
					banderaCambio = true
				} else {
					fmt.Println("\033[1;31mLa cantidad a agregar no puede ser aceptada por limitantes de espacio\033[0m")
					return
				}
				//verificacion del espacio disponible con el tamaño total del disco
			} else {
				if disponible := mbrAux.Tamanio - (mbrAux.Particiones[i].Inicio + mbrAux.Particiones[i].Tamanio); disponible >= tamanioMod {
					mbrAux.Particiones[i].Tamanio += tamanioMod
					banderaCambio = true
				} else {
					fmt.Println("\033[1;31mLa cantidad a agregar no puede ser aceptada por limitantes de espacio\033[0m")
					return
				}
			}
		} else if tamanioMod < 0 {
			//verificando que la reduccion de espacio no quede negativa
			if (mbrAux.Particiones[i].Tamanio - tamanioMod) >= 0 {
				mbrAux.Particiones[i].Tamanio -= tamanioMod
				banderaCambio = true
			} else {
				fmt.Println("\033[1;31mLa cantidad a quitar no puede ser aceptada, no pueden existir espacios negativos\033[0m")
				return
			}
		}
		//verificando si se encuentra en las logicas la particion a modificar
	} else {
		posicionActualEBR, _ := obtenerInicioTamanioExtendida(mbrAux)

		if posicionActualEBR == 0 {
			fmt.Println("\033[1;31mLa particion a modificar no se encuentra en el sistema\033[0m")
			return
		}

		posicionActualEBR, banderaNombre = verificacionExistenciaLogica(posicionActualEBR, archivo, nombre)

		if banderaNombre {
			ebrMod := obtnerEBR(archivo, posicionActualEBR)

			if tamanioMod > 0 {
				ebrSiguiente := obtnerEBR(archivo, ebrMod.Siguiente)

				//verificando el espacio disponible entre el ebr siguiente y el que se quiere modificar
				if disponible := ebrSiguiente.Inicio - (ebrMod.Inicio + ebrMod.Tamanio); disponible >= tamanioMod {
					ebrMod.Tamanio += tamanioMod
					banderaCambio = true
					escrbirEBR(archivo, posicionActualEBR, ebrMod)
				} else {
					fmt.Println("\033[1;31mLa cantidad a agregar no puede ser aceptada por limitantes de espacio\033[0m")
					return
				}
			} else if tamanioMod < 0 {
				//verificando que el espacio a modificar no sea negativo
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

func verificacionExistenciaLogica(posicionEBR int64, disco *os.File, nombre [16]byte) (int64, bool) {
	bandera := false

	for {
		ebrAux := obtnerEBR(disco, posicionEBR)

		if ebrAux.Nombre == nombre {
			bandera = true
			break
		}

		if ebrAux.Siguiente == -1 {
			break
		}
		posicionEBR = ebrAux.Siguiente
	}

	return posicionEBR, bandera
}

func verificacionExistenciaPrimariaExtendida(mbrAux discoMBR, nombre [16]byte) (bool, int) {
	bandera := false
	i := 0

	for i = 0; i < 4; i++ {
		if mbrAux.Particiones[i].Inicio != -1 {
			if mbrAux.Particiones[i].Nombre == nombre {
				bandera = true
				break
			}
		}
	}
	if i == 4 {
		i = 3
	}

	return bandera, i
}

func eliminarParticion(path, name, tipoELiminacion string) {
	archivo := obtenerDisco(path)
	defer archivo.Close()

	if archivo == nil {
		fmt.Println("\033[1;31mEl disco aun no a sido creado\033[0m")
		return
	}

	mbrAux := obtenerMBR(archivo)

	nombre := [16]byte{}
	copy(nombre[:], name)

	banderaEncontrado, i := verificacionExistenciaPrimariaExtendida(mbrAux, nombre)

	//proceso de eliminacion para primarias o extendidas
	if banderaEncontrado {
		//si es eliminacion full se borra todo el contenido de la particion
		//y se borra de la tabla de particiones
		if tipoELiminacion == "full" {
			archivo.Seek(mbrAux.Particiones[i].Inicio, 0)
			buffer := bytes.NewBuffer([]byte{})
			binary.Write(buffer, binary.BigEndian, int8(0))

			for k := mbrAux.Particiones[i].Inicio; k <= mbrAux.Particiones[i].Tamanio; k++ {
				archivo.Write(buffer.Bytes())
			}

			fmt.Println("\033[1;32mEliminacion de particion completa, realizada con exito\033[0m")
			//si es eliminacion fast solo se elimina de la tabla de particiones
		} else if tipoELiminacion == "fast" {
			fmt.Println("\033[1;32mEliminacion de particion rapida, realizada con exito\033[0m")
		}
		//proceso que comparten ambas eliminaciones que es eliminar de la tabla de particiones
		//y el desmontaje de la unidad si lo estuviera
		mbrAux.Particiones[i] = particion{Inicio: -1}
		desmontarParticionEliminada(path, name)
		escribirEnDisco(archivo, mbrAux)

		//proceso de eliminacion para logicas
	} else {
		//variable que almacena el inicio de la particion extendida que contiene las logicas
		inicioExtendida := int64(0)

		//obtencion de la posicion del ebr inicial
		posicionActualEBR, _ := obtenerInicioTamanioExtendida(mbrAux)
		inicioExtendida = posicionActualEBR

		if posicionActualEBR == 0 {
			fmt.Println("\033[1;31mLa particion a eliminar no se encuentra en el sistema\033[0m")
			return
		}

		//modificacion de la posicion del ebr inicial hacia el ebr a eliminar
		posicionActualEBR, banderaNombre := verificacionExistenciaLogica(posicionActualEBR, archivo, nombre)

		if banderaNombre {
			if tipoELiminacion == "full" {
				eliminacionParticionLogica(inicioExtendida, posicionActualEBR, archivo, path, name, nombre)
				fmt.Println("\033[1;32mEliminacion de particion completa, realzada con exito\033[0m")
			} else if tipoELiminacion == "fast" {
				eliminacionParticionLogica(inicioExtendida, posicionActualEBR, archivo, path, name, nombre)
				fmt.Println("\033[1;32mEliminacion de particion rapida, realizada con exito\033[0m")
			}
		} else {
			fmt.Println("\033[1;31mLa particion no se a encontrado en el sistema\033[0m")
			return
		}
	}
}

func eliminacionParticionLogica(inicioExtendida, posicionActualEBR int64, archivo *os.File, path, name string, nombre [16]byte) {
	//validacion de eliminacion del ebr inicial de la particion extendida
	if posicionActualEBR == inicioExtendida {
		//obtencion del ebr inicial
		ebrEliminar := obtnerEBR(archivo, posicionActualEBR)

		//creacion del buffer para escribir ceros en la particion logica
		buffer := bytes.NewBuffer([]byte{})
		binary.Write(buffer, binary.BigEndian, int8(0))

		//posicionandose en el inicio de la particion logica
		archivo.Seek(ebrEliminar.Inicio+int64(unsafe.Sizeof(logicaEBR{})), 0)

		//borrando el contenido de la particion logica
		for k := ebrEliminar.Inicio + int64(unsafe.Sizeof(logicaEBR{})); k <= ebrEliminar.Tamanio; k++ {
			archivo.Write(buffer.Bytes())
		}

		//modificando a valores por defecto el ebr inicial
		ebrEliminar.Tamanio = -1
		ebrEliminar.Nombre = [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

		escrbirEBR(archivo, posicionActualEBR, ebrEliminar)

		//eliminacion de
	} else {
		//variable que almacena la posicion del ebr a eliminar
		posicionSiguiente := int64(0)

		//ciclo infinito que se cumple hasta que se encuentra el ebr a eliminar
		//con este se busca dejar la posicion del ebr anterior al que
		//se desea eliminar en la variable inicioExtendida
		for {
			ebrAux := obtnerEBR(archivo, inicioExtendida)

			posicionSiguiente = ebrAux.Siguiente
			ebrSiguiente := obtnerEBR(archivo, posicionSiguiente)

			if ebrSiguiente.Nombre == nombre {
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
	}
	desmontarParticionEliminada(path, name)
}

func obtenerInicioTamanioExtendida(mbrAux discoMBR) (int64, int64) {
	inicio := int64(0)
	tamanio := int64(0)

	for i := 0; i < 4; i++ {
		if mbrAux.Particiones[i].Inicio != -1 {
			if mbrAux.Particiones[i].Inicio == 'e' {
				inicio = mbrAux.Particiones[i].Inicio
				tamanio = mbrAux.Particiones[i].Tamanio
			}
		}
	}

	return inicio, tamanio
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
		return size * 1024
	} else if unit == "m" {
		return size * 1048576
	}
	return int64(0)
}

func obtenerMBR(archivo *os.File) discoMBR {
	mbrAux := discoMBR{}
	contenido := make([]byte, int(unsafe.Sizeof(mbrAux)))
	archivo.Seek(0, 0)
	archivo.Read(contenido)
	buffer := bytes.NewBuffer(contenido)
	binary.Read(buffer, binary.BigEndian, &mbrAux)
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
