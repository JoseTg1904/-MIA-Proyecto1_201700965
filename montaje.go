package main

import (
	"fmt"
	"os"
	"strconv"
)

type discoMontado struct {
	Path                string
	ID                  string
	ParticionesMontadas [100]particionMontada
}

type particionMontada struct {
	Nombre [16]byte
	ID     int
}

var discosMontados [26]discoMontado
var arregloLetras = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

var discoVacio = discoMontado{}
var particionVacia = particionMontada{}

func montarParticion(path, name string) {
	archivo := buscarDisco(path)
	if archivo == nil {
		fmt.Println("El archivo del disco aun no a sido creado")
	}

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

	if banderaDiscoMontado {

		banderaParticionMontada := false
		contadorParticionesMontadas := 0

		for j := 0; j < 100; j++ {
			if discosMontados[i].ParticionesMontadas[j] != particionVacia {
				contadorParticionesMontadas++
			} else {
				particion := particionMontada{ID: i + 1}
				copy(particion.Nombre[:], name)
				discosMontados[i].ParticionesMontadas[j] = particion
				banderaParticionMontada = true
				break
			}
		}

		if banderaParticionMontada {
			fmt.Println("Se a montado la particion")
			return
		}

		if contadorParticionesMontadas == 100 {
			fmt.Println("Ya no se pueden montar mas particiones")
			return
		}

	} else {
		if contadorDiscosMontados == 26 {
			fmt.Println("Ya no se pueden montar mas discos")
		}

		k := 0
		for k = 0; k < 26; k++ {
			if discosMontados[k] == discoVacio {
				discosMontados[k] = discoMontado{Path: path, ID: arregloLetras[i]}
				break
			}
		}

		if k == 26 {
			k = 25
		}

		particion := particionMontada{ID: 1}
		copy(particion.Nombre[:], name)
		discosMontados[k].ParticionesMontadas[0] = particion

		fmt.Println("Se a montado la particion")
	}

}

func desmontarParticion(id string) {
	disco := string(id[2])
	particion, _ := strconv.Atoi(id[2:len(id)])
	discoAux := discoMontado{}

	banderaDesmontaje := false

	for i := 0; i < 26; i++ {
		if discosMontados[i].ID == disco {
			discoAux = discosMontados[i]
			break
		}
	}

	if discoAux != discoVacio {
		for i := 0; i < 100; i++ {
			if discoAux.ParticionesMontadas[i].ID == particion {
				discoAux.ParticionesMontadas[i] = particionVacia
				banderaDesmontaje = true
			}
		}

		if banderaDesmontaje {
			fmt.Println("La particion a sido desmontada")
			return
		} else {
			fmt.Println("La particion no se a encontrado")
			return
		}

	} else {
		fmt.Println("El disco aun no se encuentra montado")
		return
	}
}

func mostrarParticionesMontadas() {
	for i := 0; i < 26; i++ {
		if discosMontados[i] != discoVacio {
			for j := 0; j < 100; j++ {
				if discosMontados[i].ParticionesMontadas[j] != particionVacia {
					fmt.Println("id -> vd", discosMontados[i], discosMontados[i].ParticionesMontadas[j], " path -> ", discosMontados[i].Path, " name -> ", discosMontados[i].ParticionesMontadas[j])
				}
			}
		}
	}
}

//archivo del disco, tamaño de la particion, inicio de la particion
func obtenerDiscoMontado(id string) (*os.File, uint32, uint32) {
	disco := string(id[2])
	particion, _ := strconv.Atoi(id[2:len(id)])

	discoAux := discoMontado{}
	particionAux := particionMontada{}

	for i := 0; i < 26; i++ {
		if discosMontados[i].ID == disco {
			discoAux = discosMontados[i]
			break
		}
	}

	if discoAux != discoVacio {
		for i := 0; i < 100; i++ {
			if discoAux.ParticionesMontadas[i].ID == particion {
				particionAux = discoAux.ParticionesMontadas[i]
				break
			}
		}

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

			return nil, tamanio, inicio
		} else {
			fmt.Println("La particion no se a encontrado")
			return nil, 0, 0
		}

	} else {
		fmt.Println("El disco aun no se encuentra montado")
		return nil, 0, 0
	}
}

func buscarDisco(path string) *os.File {
	if _, err := os.Stat(path); err == nil {
		archivo, _ := os.OpenFile(path, os.O_RDWR, 0644)
		return archivo
	}
	return nil
}
