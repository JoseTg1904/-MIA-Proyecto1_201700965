package main

import (
	"fmt"
	"os"
	"strconv"
	"unsafe"
)

type discoMontado struct {
	Path                string
	ID                  string
	ParticionesMontadas [100]particionMontada
}

type particionMontada struct {
	Nombre        [16]byte
	ID            int
	Perdida       bool
	InicioCopia   int64
	UsuarioActual string
	ContraActual  string
	GrupoActual   string
}

var discosMontados [26]discoMontado
var arregloLetras = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

var discoVacio = discoMontado{}
var particionVacia = particionMontada{}

func montarParticion(path, name string) {
	archivo := buscarDisco(path)
	if archivo == nil {
		fmt.Println("\033[1;31mEl archivo del disco aun no a sido creado\033[0m")
	}

	if validarExistenciaParticion(archivo, path, name) == false {
		fmt.Println("\033[1;31mLa particion no se encuentra en el disco\033[0m")
		return
	}

	//variable que almacena la cantidad de discos montados y una bandera para indicar si el
	//disco ya se encuentra montado
	contadorDiscosMontados := 0
	banderaDiscoMontado := false

	i := 0
	for i = 0; i < 26; i++ {
		if discosMontados[i] != discoVacio {
			contadorDiscosMontados++
			if discosMontados[i].Path == path {
				banderaDiscoMontado = true
				break
			}
		}
	}
	if i == 26 {
		i = 25
	}

	//insercion de la particion en un disco que ya esta montado
	if banderaDiscoMontado {
		//variable que indica si se a montado la particion o si ya se encuentra montada
		banderaParticionMontada := false
		banderaParticionDuplicada := false

		nombreAux := [16]byte{}
		copy(nombreAux[:], name)

		for j := 0; j < 100; j++ {
			if discosMontados[i].ParticionesMontadas[j] != particionVacia {
				if discosMontados[i].ParticionesMontadas[j].Nombre == nombreAux {
					banderaParticionDuplicada = true
					break
				}
			}
		}

		if banderaParticionDuplicada {
			fmt.Println("\033[1;31mLa particion ya se encuentra montada\033[0m")
			return
		}

		for j := 0; j < 100; j++ {
			if discosMontados[i].ParticionesMontadas[j] == particionVacia {
				particion := particionMontada{ID: j + 1,
					Perdida: false}
				copy(particion.Nombre[:], name)
				discosMontados[i].ParticionesMontadas[j] = particion
				banderaParticionMontada = true
				break
			}
		}

		if banderaParticionMontada {
			fmt.Println("\033[1;32mSe a montado la particion\033[0m")
			return
		}

		//insercion del disco y la particion
	} else {
		if contadorDiscosMontados == 26 {
			fmt.Println("\033[1;31mYa no se pueden montar mas discos\033[0m")
			return
		}

		k := 0
		for k = 0; k < 26; k++ {
			if discosMontados[k] == discoVacio {
				discosMontados[k] = discoMontado{Path: path, ID: arregloLetras[k]}
				break
			}
		}
		if k == 26 {
			k = 25
		}

		particion := particionMontada{ID: 1, Perdida: false}
		copy(particion.Nombre[:], name)
		discosMontados[k].ParticionesMontadas[0] = particion

		fmt.Println("\033[1;32mSe a montado la particion\033[0m")
	}
}

func validarExistenciaParticion(archivo *os.File, path, name string) bool {
	mbrAux := obtenerMBR(archivo)

	contExtendidas := 0
	posicionActualEBR := int64(0)
	banderaNombre := false

	nombre := [16]byte{}
	copy(nombre[:], name)

	for i := 0; i < 4; i++ {
		if mbrAux.Particiones[i].Inicio != -1 {
			if mbrAux.Particiones[i].Tipo == 'e' {
				contExtendidas++
				posicionActualEBR = mbrAux.Particiones[i].Inicio
			}
			if mbrAux.Particiones[i].Nombre == nombre {
				banderaNombre = true
			}
		}
	}

	if contExtendidas == 1 {
		_, banderaNombre = verificacionExistenciaLogica(posicionActualEBR, archivo, nombre)
	}

	return banderaNombre
}

func recorridoObtenerDiscoMontado(idDisco string) (int, discoMontado) {
	discoAux := discoMontado{}
	i := 0

	for i = 0; i < 26; i++ {
		if discosMontados[i].ID == idDisco {
			discoAux = discosMontados[i]
			break
		}
	}
	if i == 26 {
		i = 25
	}

	return i, discoAux
}

func recorridoObtenerParticionMontada(idParticion int, discoAux discoMontado) (int, particionMontada) {
	particionAux := particionMontada{}
	i := 0

	for i = 0; i < 100; i++ {
		if discoAux.ParticionesMontadas[i].ID == idParticion {
			particionAux = discoAux.ParticionesMontadas[i]
			break
		}
	}
	if i == 100 {
		i = 99
	}

	return i, particionAux
}

func desmontarParticion(id string) {
	//descomposicion del string que correspone a una particion montada en la letra que correspone
	//al disco y el numero que corresponde a la particion
	disco := string(id[2])
	particion, _ := strconv.Atoi(id[3:len(id)])

	banderaDesmontaje := false

	i, discoAux := recorridoObtenerDiscoMontado(disco)

	if discoAux != discoVacio {
		for j := 0; j < 100; j++ {
			if discosMontados[i].ParticionesMontadas[j] != particionVacia {
				if discosMontados[i].ParticionesMontadas[j].ID == particion {
					discosMontados[i].ParticionesMontadas[j] = particionMontada{}
					banderaDesmontaje = true
					break
				}
			}
		}

		if banderaDesmontaje {
			fmt.Println("\033[1;32mLa particion a sido desmontado\033[0m")
		} else {
			fmt.Println("\033[1;31mLa particion no se a encontrado\033[0m")
		}

	} else {
		fmt.Println("\033[1;31mEl disco aun no se encuentra montado\033[0m")
	}
}

func mostrarParticionesMontadas() {
	for i := 0; i < 26; i++ {
		if discosMontados[i] != discoVacio {
			for j := 0; j < 100; j++ {
				if discosMontados[i].ParticionesMontadas[j] != particionVacia {
					fmt.Println("id -> vd"+discosMontados[i].ID+
						strconv.Itoa(discosMontados[i].ParticionesMontadas[j].ID), " path ->",
						discosMontados[i].Path, " name ->",
						string(discosMontados[i].ParticionesMontadas[j].Nombre[:]))
				}
			}
		}
	}
}

//archivo del disco, tamaño de la particion, inicio de la particion
func obtenerDiscoMontado(id string) (*os.File, uint32, uint32) {
	disco := string(id[2])
	particion, _ := strconv.Atoi(id[3:len(id)])

	_, discoAux := recorridoObtenerDiscoMontado(disco)

	if discoAux != discoVacio {
		_, particionAux := recorridoObtenerParticionMontada(particion, discoAux)

		if particionAux != particionVacia {
			archivo := buscarDisco(discoAux.Path)
			mbrAux := obtenerMBR(archivo)

			tamanio, inicio := uint32(0), uint32(0)

			for i := 0; i < 4; i++ {
				if mbrAux.Particiones[i].Nombre == particionAux.Nombre {
					tamanio = uint32(mbrAux.Particiones[i].Tamanio)
					inicio = uint32(mbrAux.Particiones[i].Inicio)
					break
				}
			}

			return archivo, tamanio, inicio
		}
		fmt.Println("\033[1;31mLa particion no a sido encontrado\033[0m")
		return nil, 0, 0
	}
	fmt.Println("\033[1;31mEl disco aun no se encuentrada montado\033[0m")
	return nil, 0, 0
}

func buscarDisco(path string) *os.File {
	if _, err := os.Stat(path); err == nil {
		archivo, _ := os.OpenFile(path, os.O_RDWR, 0644)
		return archivo
	}
	return nil
}

func desmontarDiscoEliminado(path string) {
	for i := 0; i < 26; i++ {
		if discosMontados[i].Path == path {
			discosMontados[i] = discoMontado{}
			break
		}
	}
}

func desmontarParticionEliminada(path, nombre string) {
	discoAux := discoMontado{}

	nombreAux := [16]byte{}
	copy(nombreAux[:], nombre)

	i := 0
	for i = 0; i < 26; i++ {
		if discosMontados[i].Path == path {
			discoAux = discosMontados[i]
			break
		}
	}
	if i == 26 {
		i = 25
	}

	if discoAux != discoVacio {
		for j := 0; j < 100; j++ {
			if discosMontados[i].ParticionesMontadas[j] != particionVacia {
				if discosMontados[i].ParticionesMontadas[j].Nombre == nombreAux {
					discosMontados[i].ParticionesMontadas[j] = particionVacia
					break
				}
			}
		}
	}
}

func cambiarEstadoPerdida(id string, estado bool) {
	disco := string(id[2])
	particion, _ := strconv.Atoi(id[3:len(id)])

	j, discoAux := recorridoObtenerDiscoMontado(disco)

	if discoAux != discoVacio {
		i, particionAux := recorridoObtenerParticionMontada(particion, discoAux)

		if particionAux != particionVacia {
			discosMontados[j].ParticionesMontadas[i].Perdida = estado
			if estado {
				archivo, _, inicio := obtenerDiscoMontado(id)
				sb := obtenerSuperBoot(archivo, int64(inicio))
				discosMontados[j].ParticionesMontadas[i].InicioCopia = int64(sb.InicioLog) + (int64(sb.NoAVD) * int64(unsafe.Sizeof(bitacora{})))
			}
		}
	}
}

//estado y posicion de la copia del super boot
func obtenerEstadoPerdida(id string) (bool, int64) {
	disco := string(id[2])
	particion, _ := strconv.Atoi(id[3:len(id)])

	j, discoAux := recorridoObtenerDiscoMontado(disco)

	if discoAux != discoVacio {
		i, particionAux := recorridoObtenerParticionMontada(particion, discoAux)

		if particionAux != particionVacia {
			return discosMontados[j].ParticionesMontadas[i].Perdida, discosMontados[j].ParticionesMontadas[i].InicioCopia
		}
	}

	return false, 0
}
