package main

import (
	"fmt"
	"os"
	"strconv"
)

type ParticionMontada struct{
	Numero int64
    Estado int64
    Nombre[20]byte
}

type Montaje struct{
    Ruta string
	Letra string
    Estado int64
    Particiones [26]ParticionMontada
}

var discosMontados [50]Montaje
func montarParticion(ruta string, nombre string, letraDisco string){
	
	existeParticion := false
	existeDiscoMontado := false
	posicion := 0

	archivo, _ := os.OpenFile(ruta, os.O_RDWR, 0644)
	if archivo == nil {
		mensaje += "Disco no existe, no es posible montar una particion sin un disco..\n"
		return
	}

	defer archivo.Close()


	discoAux := obtenerMBR(archivo)
	nombreAux := [20]byte{}
	copy(nombreAux[:], nombre)

	 for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Nombre == nombreAux{
			existeParticion = true
		}
	}

	if !existeParticion {
        mensaje += "La particion con el nombre [" + nombre + "] no existe.\n"
        return
    }

	for i := 0; i < 50; i++ {
		if discosMontados[i].Estado == 1 && discosMontados[i].Ruta == ruta {
			existeDiscoMontado = true
			posicion = i
			break
		}
	}

	if !existeDiscoMontado {
		for i := 0; i < 50; i++ {
			if discosMontados[i].Estado == 0 {
				discosMontados[i].Estado = 1
				discosMontados[i].Letra = letraDisco
				discosMontados[i].Ruta = ruta
				existeDiscoMontado = true
				posicion = i
				break
			}
		}
	}

	if existeDiscoMontado {
		for i := 0; i < 26; i++ {
			if discosMontados[posicion].Particiones[i].Nombre == nombreAux {
				mensaje += "¡ Particion ya se encuentra montada, no se puede volver a montar la misma particion !\n"
				return
			}
		}
	}

	if existeDiscoMontado {
		for i := 0; i < 26; i++ {
			if discosMontados[posicion].Particiones[i].Estado == 0 {
				discosMontados[posicion].Particiones[i].Nombre = nombreAux
				discosMontados[posicion].Particiones[i].Estado = 1
				discosMontados[posicion].Particiones[i].Numero = int64(i + 1)
				mensaje += "\n¡ Particion Montada exitosamente !\n"
				break
			}
		}
	}

	mostrarMontaje()
}

func desmotar(id string){
	existeDisco := obtenerDiscoMontado(id)
	encontrado :=  false
	if existeDisco != ""{
		idAux := string(id[0])
		for i:= 0; i < 50; i++{
			if discosMontados[i].Letra == idAux{
				numero, err := strconv.ParseInt(string(id[1]), 10, 64)
				if err != nil {
					fmt.Println("Error al convertir el string a int64")
					return
				}
				for j:= 0; j < 26; j++{
					if discosMontados[i].Particiones[j].Numero == numero{
						discosMontados[i].Particiones[j] = ParticionMontada{}
						mensaje += "Particion desmotada exitosamente\n"
						encontrado = true
						break
					}else{
						mensaje += "Particion no esta motada, intentelo nuevamente\n"
						break
					}
				} 
			}
		}
	}else{
		mensaje += "¡ El disco no a sido montado, intentelo nuevamente !\n"
	}
	if(encontrado){
		mostrarMontaje()
	}else{
		mensaje += "Particion no esta montada, intente montar una.\n"
	}
}

func mostrarMontaje(){

	mensaje += "Reporte de Disco y Particiones montadas..\n"
	for i := 0; i < 50; i++ {
		if discosMontados[i].Estado == 1 {
			mensaje += "*-*-*-*-*-*-*-*-*-*-*-*-*-* " + strconv.Itoa(int(i)) + " -*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-\n"
			mensaje += "Letra del disco : " + discosMontados[i].Letra + "\n"
			mensaje += "Ruta del disco : " + discosMontados[i].Ruta + "\n"
			for j := 0; j < 26; j++ {
				if discosMontados[i].Particiones[j].Estado == 1 {
					mensaje += "\tNombre de la particion : " + cadenaLimpia(discosMontados[i].Particiones[j].Nombre[:]) + "\n"
					mensaje += "\nNumero de la particion : " + strconv.Itoa(int(discosMontados[i].Particiones[j].Numero)) + "\n"
				}
			}
			mensaje += "*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-\n"
		}
	}
}

func obtenerDiscoMontado(id string) string{

	idAux := string(id[0])
	rutaObtenida := ""
	for i := 0; i < 50; i++ {
		if discosMontados[i].Letra ==  idAux{
			rutaObtenida = discosMontados[i].Ruta
			break
		}
	}
	return rutaObtenida
}

func obtenerParticionMontada(id string) [20]byte{

	rutaDisco := obtenerDiscoMontado(id)
	numeroParticion := obtenerNumero(int64(id[1]))
	nombreParticion := [20]byte{}

	for i := 0; i < 50; i++ {
		if discosMontados[i].Ruta == rutaDisco {
			for j := 0; j < 26; j++ {
				if discosMontados[i].Particiones[j].Numero == numeroParticion {
					copy(nombreParticion[:], discosMontados[i].Particiones[j].Nombre[:])
				}
			}
		}
	}

	return nombreParticion
}

func obtenerInicioTamanio(nombreParticion [20]byte, discoAux MBR) (int64, int64){

	inicio := int64(-1)
	tamanio := int64(-1)

	
	for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Nombre == nombreParticion {
			inicio = discoAux.Particiones[i].Inicio
			tamanio = discoAux.Particiones[i].Tamanio
			break
		}
	}

	return inicio, tamanio
}

func obtenerLetra(numero  int) string{

    letra := ""
    switch (numero) {
    case 0:
        letra = "a"
    case 1:
        letra = "b"
        
    case 2:
        letra = "c"
        
    case 3:
        letra = "d"
        
    case 4:
        letra = "e"
        
    case 5:
        letra = "f"
        
    case 6:
        letra = "g"
        
    case 7:
        letra = "h"
        
    case 8:
        letra = "i"
        
    case 9:
        letra = "j"
        
    case 10:
        letra = "k"
        
    case 11:
        letra = "l"
        
    case 12:
        letra = "m"
        
    case 13:
        letra = "n"
        
    case 14:
        letra = "o"
        
    case 15:
        letra = "p"
        
    case 16:
        letra = "k"
        
    case 17:
        letra = "r"
        
    case 18:
        letra = "s"
        
    case 19:
        letra = "t"
        
    case 20:
        letra = "u"
        
    case 21:
        letra = "v"
        
    case 22:
        letra = "w"
        
    case 23:
        letra = "x"
        
    case 24:
        letra = "y"

    case 25:
        letra = "z"
    }

    return letra
}

func obtenerNumero(valor int64) int64{

	numero := int64(0)
	switch valor {
	case 49:
		numero = int64(1)
	case 50:
		numero = int64(2)
	case 51:
		numero = int64(3)
	case 52:
		numero = int64(4)
	case 53:
		numero = int64(5)
	case 54:
		numero = int64(6)
	case 55:
		numero = int64(7)
	case 56:
		numero = int64(8)
	case 57:
		numero = int64(9)
	}
	return numero
}

func obtenerLetraMontada(numero  int64) string{

    letra := ""
    switch (numero) {
    case 97:
        letra = "a"
        
    case 98:
        letra = "b"
        
    case 99:
        letra = "c"
        
    case 100:
        letra = "d"
        
    case 101:
        letra = "e"
        
    case 102:
        letra = "f"
        
    case 103:
        letra = "g"
        
    case 104:
        letra = "h"
        
    case 105:
        letra = "i"
        
    case 106:
        letra = "j"
        
    case 107:
        letra = "k"
        
    case 108:
        letra = "l"
        
    case 109:
        letra = "m"
        
    case 110:
        letra = "n"
        
    case 111:
        letra = "o"
        
    case 112:
        letra = "p"
        
    case 113:
        letra = "k"
        
    case 114:
        letra = "r"
        
    case 115:
        letra = "s"
        
    case 116:
        letra = "t"
        
    case 117:
        letra = "u"
        
    case 118:
        letra = "v"
        
    case 119:
        letra = "w"
        
    case 120:
        letra = "x"
        
    case 121:
        letra = "y"

    case 122:
        letra = "z"
    }

    return letra
}