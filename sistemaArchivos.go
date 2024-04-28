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

/*
1.Struct utilizados para la administracion y creacion del sistema de archivos.

	-Struct para el super bloque.
	-Struct para la tabla de inodos.
	-Struct para el bloque de carpetas.
	-Struct para el contenido de un archivo.
	-Struct para el journaling.
*/
type SuperBloque struct {
	IdSistema            uint32
	NumeroInodos         uint32
	NumeroBloques        uint32
	NumeroBloquesLibres  uint32
	NumeroInodosLibres   uint32
	UltimaFechaMontado   [20]byte
	NumeroSistemaMontado uint32
	NumeroMagico         uint32
	TamanioInodo         uint32
	TamanioBloque        uint32
	PrimerInodoLibre     uint32
	PrimerBloqueLibre    uint32
	InicioBitMapsInodos  uint32
	InicioBitMapsBloques uint32
	InicioTablaInodos    uint32
	InicioTablaBloques   uint32
}

type TablaInodos struct {
	IdUsuario         uint32
	IdGrupo           uint32
	TamanioArchivo    uint32
	FechaLectura      [20]byte
	FechaCreacion     [20]byte
	FechaModificacion [20]byte
	Bloque            [15]int64
	Tipo              int64
	Permisos          uint32
}

type Contenido struct {
	Nombre    [25]byte
	Apuntador int64
}

type BloqueCarpeta struct {
	Contenidos [4]Contenido
}

type BloqueArchivos struct {
	Datos [150]byte
}

type Journaling struct {
	TipoOperacion string
	Tipo          byte
	Nombre        string
	Ruta          string
	Contenido     string
	Fecha         string
	Propietario   string
	Permisos      int
	Tamanio       int
}

var grupoActual = ""
var usuarioActual = ""
var contraseniaActual = ""
var sesionIniciada = false
var identificadorActual = ""


func crearSistemaArchivosEXT2(id string, tipoFormateo string) {

	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar un formateo..")
		return
	}

	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar un formateo..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	inicioParticion, tamanioParticion := obtenerInicioTamanio(nombreParticion, discoAux)


	archivo.Seek(inicioParticion, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))


	numeroEstructuras := obtenerNumeroEstructuras(tamanioParticion)

	bitMapInodos := inicioParticion + int64(unsafe.Sizeof(SuperBloque{}))
	bitMapBloques := inicioParticion + int64(unsafe.Sizeof(SuperBloque{})) + numeroEstructuras


	superBloque := SuperBloque{
		IdSistema:            2,
		NumeroInodos:         uint32(numeroEstructuras),
		NumeroBloques:        uint32(3 * numeroEstructuras),
		NumeroBloquesLibres:  uint32(3*numeroEstructuras) - 2,
		NumeroInodosLibres:   uint32(numeroEstructuras) - 2,
		NumeroSistemaMontado: 0,
		NumeroMagico:         0xEF53,
		TamanioInodo:         uint32(unsafe.Sizeof(TablaInodos{})),
		TamanioBloque:        uint32(unsafe.Sizeof(BloqueArchivos{})),
		PrimerInodoLibre:     uint32(bitMapInodos) + uint32(3*numeroEstructuras),
		PrimerBloqueLibre:    uint32(bitMapBloques) + uint32(3*numeroEstructuras),
		InicioBitMapsInodos:  uint32(bitMapInodos),
		InicioBitMapsBloques: uint32(bitMapBloques),
		InicioTablaInodos:    uint32(bitMapInodos) + uint32(numeroEstructuras) + uint32(3*numeroEstructuras),
		InicioTablaBloques:   uint32(bitMapBloques) + uint32(3*numeroEstructuras) + (uint32(numeroEstructuras) * uint32(unsafe.Sizeof(TablaInodos{}))),
	}
	copy(superBloque.UltimaFechaMontado[:], obtenerFecha())


	archivo.Seek(inicioParticion, 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, &superBloque)
	archivo.Write(buffer.Bytes())


	escribirBitMapInodo(archivo, uint32(numeroEstructuras), superBloque)

	escribirBitMapBloque(archivo, uint32(numeroEstructuras), superBloque)


	iNodoRoot := TablaInodos{
		IdUsuario:      1,
		IdGrupo:        1,
		TamanioArchivo: 0,
		Tipo:           int64(0),
		Permisos:       664,
	}
	copy(iNodoRoot.FechaCreacion[:], obtenerFecha())
	copy(iNodoRoot.FechaLectura[:], obtenerFecha())
	copy(iNodoRoot.FechaModificacion[:], obtenerFecha())
	for i := 0; i < 15; i++ {
		iNodoRoot.Bloque[i] = -1
	}

	posicionInodo := escribirInodo(archivo, superBloque, iNodoRoot)
	contenido := "1,G,root\n1,U,root,root,123\n"
	crearArchivo(archivo, superBloque, contenido, "users.txt", posicionInodo)

	fmt.Println("¡ Formateo del sistema EXT2 fue realizado exitosamente !")
}

func crearSistemaArchivosEXT3(id string, tipoFormateo string) {


	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar un formateo..")
		return
	}


	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar un formateo..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	inicioParticion, tamanioParticion := obtenerInicioTamanio(nombreParticion, discoAux)


	archivo.Seek(inicioParticion, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))


	numeroEstructuras := obtenerNumeroEstructuras(tamanioParticion)

	bitMapInodos := inicioParticion + int64(unsafe.Sizeof(SuperBloque{})) + (numeroEstructuras * int64(unsafe.Sizeof(Journaling{})))
	bitMapBloques := inicioParticion + int64(unsafe.Sizeof(SuperBloque{})) + numeroEstructuras + (numeroEstructuras * int64(unsafe.Sizeof(Journaling{})))


	superBloque := SuperBloque{
		IdSistema:            3,
		NumeroInodos:         uint32(numeroEstructuras),
		NumeroBloques:        uint32(3 * numeroEstructuras),
		NumeroBloquesLibres:  uint32(3*numeroEstructuras) - 2,
		NumeroInodosLibres:   uint32(numeroEstructuras) - 2,
		NumeroSistemaMontado: 0,
		NumeroMagico:         0xEF53,
		TamanioInodo:         uint32(unsafe.Sizeof(TablaInodos{})),
		TamanioBloque:        uint32(unsafe.Sizeof(BloqueArchivos{})),
		PrimerInodoLibre:     uint32(bitMapInodos) + uint32(3*numeroEstructuras),
		PrimerBloqueLibre:    uint32(bitMapBloques) + uint32(3*numeroEstructuras),
		InicioBitMapsInodos:  uint32(bitMapInodos),
		InicioBitMapsBloques: uint32(bitMapBloques),
		InicioTablaInodos:    uint32(bitMapInodos) + uint32(numeroEstructuras) + uint32(3*numeroEstructuras),
		InicioTablaBloques:   uint32(bitMapBloques) + uint32(3*numeroEstructuras) + (uint32(numeroEstructuras) * uint32(unsafe.Sizeof(TablaInodos{}))),
	}
	copy(superBloque.UltimaFechaMontado[:], obtenerFecha())


	archivo.Seek(inicioParticion, 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, &superBloque)
	archivo.Write(buffer.Bytes())


	limite := superBloque.InicioTablaInodos
	posicionJournaling := inicioParticion + int64(unsafe.Sizeof(SuperBloque{}))
	for i := posicionJournaling; i < int64(limite); i = (i + int64(unsafe.Sizeof(Journaling{}))) {
		archivo.Seek(int64(i), 0)
		binary.Write(buffer, binary.BigEndian, &superBloque)
		archivo.Write(buffer.Bytes())

	}


	escribirBitMapInodo(archivo, uint32(numeroEstructuras), superBloque)

	escribirBitMapBloque(archivo, uint32(numeroEstructuras), superBloque)


	iNodoRoot := TablaInodos{
		IdUsuario:      1,
		IdGrupo:        1,
		TamanioArchivo: 0,
		Tipo:           int64(0),
		Permisos:       664,
	}
	copy(iNodoRoot.FechaCreacion[:], obtenerFecha())
	copy(iNodoRoot.FechaLectura[:], obtenerFecha())
	copy(iNodoRoot.FechaModificacion[:], obtenerFecha())
	for i := 0; i < 15; i++ {
		iNodoRoot.Bloque[i] = -1
	}

	posicionInodo := escribirInodo(archivo, superBloque, iNodoRoot)
	contenido := "1,G,root\n1,U,root,root,123\n"
	crearArchivo(archivo, superBloque, contenido, "users.txt", posicionInodo)

	fmt.Println("¡ Formateo del sistema EXT3 fue realizado exitosamente !")
}



func escribirBitMapInodo(archivo *os.File, numeroEstructuras uint32, super SuperBloque) {

	posicion := super.InicioBitMapsInodos
	contador := 0

	for {

		if contador < int(numeroEstructuras) {
			escritura := bytes.NewBuffer([]byte{})
			archivo.Seek(int64(posicion), 0)
			binary.Write(escritura, binary.BigEndian, uint8(0))
			archivo.Write(escritura.Bytes())
			contador++
			posicion++

		} else {
			break
		}
	}
}


func escribirBitMapBloque(archivo *os.File, numeroEstructuras uint32, super SuperBloque) {

	posicion := int64(super.InicioBitMapsBloques)
	contador := int64(0)
	limiteBitMap := numeroEstructuras * 3

	for {

		if int(contador) < int(limiteBitMap) {
			escritura := bytes.NewBuffer([]byte{})
			archivo.Seek(posicion, 0)
			binary.Write(escritura, binary.BigEndian, uint8(0))
			archivo.Write(escritura.Bytes())
			contador++
			posicion++
		} else {
			break
		}
	}
}

func escribirInodo(disco *os.File, super SuperBloque, inodo TablaInodos) int64 {


	inicioBitMapInodos := int64(super.InicioBitMapsInodos)
	aux := uint8(1)

	contador := int64(0)

	for {
		if aux == uint8(1) {

			leerBitMap := make([]byte, int(unsafe.Sizeof(aux)))
			disco.Seek(inicioBitMapInodos+contador, 0)
			disco.Read(leerBitMap)
			buffer := bytes.NewBuffer(leerBitMap)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
			}
		} else {
			break
		}
	}
	
	
	inicioNuevoInodo := int64(super.InicioTablaInodos) + (contador * int64(unsafe.Sizeof(TablaInodos{})))
	disco.Seek(inicioNuevoInodo, 0)
	bufferInodo := bytes.NewBuffer([]byte{})
	binary.Write(bufferInodo, binary.BigEndian, &inodo)
	disco.Write(bufferInodo.Bytes())

	disco.Seek(inicioBitMapInodos+contador, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(1))
	disco.Write(buffer.Bytes())

	return inicioNuevoInodo
}

func escribirBloqueCarpeta(disco *os.File, super SuperBloque, bloque BloqueCarpeta) int64 {

	inicioBitMapBloque := int64(super.InicioBitMapsBloques)
	aux := uint8(1)
	contador := int64(0)

	for {
		if aux == uint8(1) {
			leerBitMap := make([]byte, int(unsafe.Sizeof(aux)))
			disco.Seek(inicioBitMapBloque+contador, 0)
			disco.Read(leerBitMap)
			buffer := bytes.NewBuffer(leerBitMap)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
			}
		} else {
			break
		}
	}

	inicioNuevoBloque := int64(super.InicioTablaBloques) + (contador * int64(unsafe.Sizeof(BloqueCarpeta{})))
	disco.Seek(inicioNuevoBloque, 0)
	bufferBloque := bytes.NewBuffer([]byte{})
	binary.Write(bufferBloque, binary.BigEndian, &bloque)
	disco.Write(bufferBloque.Bytes())

	disco.Seek(inicioBitMapBloque+contador, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(1))
	disco.Write(buffer.Bytes())

	return inicioNuevoBloque
}


func escribirBloqueArchivo(disco *os.File, super SuperBloque, archivo BloqueArchivos) int64 {

	inicioBitMapBloque := int64(super.InicioBitMapsBloques)
	aux := uint8(1)
	contador := int64(0)

	for {
		if aux == uint8(1) {

			leerBitMap := make([]byte, int(unsafe.Sizeof(aux)))
			disco.Seek(inicioBitMapBloque+contador, 0)
			disco.Read(leerBitMap)
			buffer := bytes.NewBuffer(leerBitMap)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
				contador++
/* 				contador++
				contador++
				contador++
				contador++
				contador++ */
			}
		} else {
			break
		}
	}

	inicioNuevoArchivo := int64(super.InicioTablaBloques) + (contador * int64(unsafe.Sizeof(BloqueArchivos{})))
	disco.Seek(inicioNuevoArchivo, 0)
	bufferBloque := bytes.NewBuffer([]byte{})
	binary.Write(bufferBloque, binary.BigEndian, &archivo)
	disco.Write(bufferBloque.Bytes())

	disco.Seek(inicioBitMapBloque+contador, 0)
	bufferuno := bytes.NewBuffer([]byte{})
	binary.Write(bufferuno, binary.BigEndian, uint8(1))
	disco.Write(bufferuno.Bytes())

	return inicioNuevoArchivo
}


func crearArchivo(archivo *os.File, super SuperBloque, cadena string, nombreArchivo string, posicionInodo int64) {

	inodo := obtenerInodo(archivo, posicionInodo)
	if inodo.Tipo == int64(0) {
		if inodo.Bloque[0] == int64(-1) {
			nuevoBloque := BloqueCarpeta{}
			copy(nuevoBloque.Contenidos[0].Nombre[:], ".")
			nuevoBloque.Contenidos[0].Apuntador = 0
			copy(nuevoBloque.Contenidos[1].Nombre[:], "..")
			nuevoBloque.Contenidos[1].Apuntador = 0
			copy(nuevoBloque.Contenidos[2].Nombre[:], "")
			nuevoBloque.Contenidos[2].Apuntador = -1
			copy(nuevoBloque.Contenidos[3].Nombre[:], "")
			nuevoBloque.Contenidos[3].Apuntador = -1
			posicionBloque := escribirBloqueCarpeta(archivo, super, nuevoBloque)
			inodo.Bloque[0] = posicionBloque
			reescribirInodo(archivo, posicionInodo, inodo)
		}

		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != int64(-1) {
				posicionBloque := inodo.Bloque[i]
				bloqueCarpeta := obtenerBloqueCarpetas(archivo, posicionBloque)
				for j := 0; j < 4; j++ {
					if bloqueCarpeta.Contenidos[j].Apuntador == int64(-1) {
						nuevoInodo := TablaInodos{
							IdUsuario:      1,
							IdGrupo:        1,
							TamanioArchivo: uint32(len(cadena)),
							Tipo:           int64(1),
							Permisos:       664,
						}
						copy(nuevoInodo.FechaCreacion[:], obtenerFecha())
						copy(nuevoInodo.FechaLectura[:], obtenerFecha())
						copy(nuevoInodo.FechaModificacion[:], obtenerFecha())

						for k := 0; k < 15; k++ {
							if len(cadena) == 0 {
								nuevoInodo.Bloque[k] = -1
							} else {
								nuevoArchivo := BloqueArchivos{}

								copy(nuevoArchivo.Datos[:], cadena)
								cadena = ""
								posicionArchivo := escribirBloqueArchivo(archivo, super, nuevoArchivo)
								nuevoInodo.Bloque[k] = posicionArchivo
							}
						}

						posicionNuevoInodo := escribirInodo(archivo, super, nuevoInodo)
						bloqueCarpeta.Contenidos[j].Apuntador = posicionNuevoInodo
						copy(bloqueCarpeta.Contenidos[j].Nombre[:], nombreArchivo)
						reescribirBloque(archivo, inodo.Bloque[i], bloqueCarpeta)
						return
					}
				}

			} else {
				i--
			}
		}
	}
}


func reescribirInodo(disco *os.File, posicionInodo int64, inodo TablaInodos) {
	
	disco.Seek(posicionInodo, 0)

	
	bufferEscritura := bytes.NewBuffer([]byte{})

	
	binary.Write(bufferEscritura, binary.BigEndian, &inodo)

	
	disco.Write(bufferEscritura.Bytes())
}

func reescribirBloque(disco *os.File, posicionBloque int64, bloque BloqueCarpeta) {
	
	disco.Seek(posicionBloque, 0)

	
	bufferEscritura := bytes.NewBuffer([]byte{})

	
	binary.Write(bufferEscritura, binary.BigEndian, &bloque)

	
	disco.Write(bufferEscritura.Bytes())
}


func reescribirBloqueArchivo(disco *os.File, posicionBloque int64, bloque BloqueArchivos) {
	
	disco.Seek(posicionBloque, 0)

	
	bufferEscritura := bytes.NewBuffer([]byte{})

	
	binary.Write(bufferEscritura, binary.BigEndian, &bloque)

	
	disco.Write(bufferEscritura.Bytes())
}


func obtenerInodo(disco *os.File, posicionInodo int64) TablaInodos {
	Inodo := TablaInodos{}

	contenido := make([]byte, int(unsafe.Sizeof(Inodo)))

	
	disco.Seek(posicionInodo, 0)

	
	disco.Read(contenido)


	bufferLectura := bytes.NewBuffer(contenido)


	binary.Read(bufferLectura, binary.BigEndian, &Inodo)

	return Inodo
}


func obtenerBloqueCarpetas(disco *os.File, posicionBloque int64) BloqueCarpeta {
	
	bloqueCarpeta := BloqueCarpeta{}

	
	contenido := make([]byte, int(unsafe.Sizeof(bloqueCarpeta)))

	
	disco.Seek(posicionBloque, 0)

	
	disco.Read(contenido)

	
	bufferLectura := bytes.NewBuffer(contenido)

	
	binary.Read(bufferLectura, binary.BigEndian, &bloqueCarpeta)

	return bloqueCarpeta
}


func obtenerBloqueArchivo(disco *os.File, posicionBloque int64) BloqueArchivos {
	
	bloqueArchivo := BloqueArchivos{}

	
	contenido := make([]byte, int(unsafe.Sizeof(bloqueArchivo)))

	
	disco.Seek(posicionBloque, 0)

	
	disco.Read(contenido)

	
	bufferLectura := bytes.NewBuffer(contenido)

	
	binary.Read(bufferLectura, binary.BigEndian, &bloqueArchivo)

	return bloqueArchivo
}


func iniciarSesion(usuario string, contrasenia string, id string) (bool, string) {

	error := ""
	encontrado := false

	if sesionIniciada {
		mensaje += "Ya existe una sesion iniciada no puede haber 2 sesiones al mismo tiempo....\n"
		return encontrado, error
	}

	rutaObtenida := obtenerDiscoMontado(id)
	nombreParticion := obtenerParticionMontada(id)
	nombreParticionAux := cadenaLimpia(nombreParticion[:])

	if rutaObtenida == "" || nombreParticionAux == "" {
		mensaje += "Particion o Disco no estan montados, no es posible realizar la accion..\n"
		return encontrado, error
	}


	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		mensaje += "Disco no existe, no es posible realizar la accion..\n"
		return encontrado, error
	}

	discoAux := obtenerMBR(archivo)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if letra[0] != "0" {
			if letra[1] == "U" {
				if letra[3] == usuario && letra[4] == contrasenia {
					usuarioActual = usuario
					contraseniaActual = contrasenia
					grupoActual = letra[2]
					identificadorActual = id
					sesionIniciada = true
					encontrado = true
				}
			}
		}
	}

	if !encontrado {
		mensaje += "Usuario o contraseña no existe, vuelva a intentarlo....\n"
	} else {
		mensaje += "¡ Bienvenido al sistema de archivos Usuario : [" + usuarioActual + "] !\n"
	}

	archivo.Close()

	return encontrado, error
}


func cerrarSesion() (bool, string) {

	validar := ""
	if !sesionIniciada {
		mensaje += "¡ No existe una sesion iniciada no es posible cerrar sesion !\n"
		return sesionIniciada, validar
	}
	usuarioActual = ""
	contraseniaActual = ""
	grupoActual = ""
	identificadorActual = ""
	sesionIniciada = false
	mensaje += "¡ Sesion cerrada exitosamente !\n"
	return sesionIniciada, validar
}


func crearGrupo(nombreGrupo string) {

	if !sesionIniciada {
		mensaje += "Debe existir una sesion iniciada para crear un grupo, vuelva  intentarlo.\n"
		return
	}

	if usuarioActual != "root" {
		mensaje += "Solo el usuario [root] puede utilizar este comando.\n"
		return
	}

	if len(nombreGrupo) > 10 {
		mensaje += "El nombre del grupo no puede tener mas de 10 caracteres, vuelva a intentarlo.\n"
		return
	}

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		mensaje += "Particion no esta montado, no es posible realizar el reporte..\n"
		return
	}

	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		mensaje += "Disco no existe, no es posible realizar el reporte..\n"
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")
	grupoExiste := false
	idGrupo := 0

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if letra[0] != "0" {
			idGrupo, _ = strconv.Atoi(letra[0])
			if letra[1] == "G" {
				if letra[2] == nombreGrupo {
					grupoExiste = true
				}
			}
		}
	}

	if grupoExiste {
		mensaje += "¡ Grupo ya existe, ingrese un nuevo nombre !\n"
		return
	}

	idGrupo++
	contenido += strconv.Itoa(int(idGrupo)) + "," + "G" + "," + nombreGrupo + "\n"
	reescribirArchivo(archivo, super, int64(posicionInicial), contenido, "users.txt")
	mensaje += leerArchivo(archivo, super, int64(posicionInicial), "users.txt") + "\n"
	mensaje += "¡ Grupo creado exitosamente !\n"
	archivo.Close()
}

func eliminarGrupo(nombreGrupo string) {

	if !sesionIniciada {
		mensaje += "Debe existir una sesion iniciada para crear un grupo, vuelva  intentarlo.\n"
		return
	}

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		mensaje += "Particion no esta montado, no es posible realizar el reporte..\n"
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		mensaje += "Disco no existe, no es posible realizar el reporte..\n"
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")
	concatenarContenido := ""
	grupoExiste := false

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if letra[0] != "0" {
			if letra[2] == nombreGrupo {
				letra[0] = "0"
				grupoExiste = true
			}
		}

		if letra[1] == "U" {
			concatenarContenido += letra[0] + "," + letra[1] + "," + letra[2] + "," + letra[3] + "," + letra[4] + "\n"
		} else {
			concatenarContenido += letra[0] + "," + letra[1] + "," + letra[2] + "\n"
		}
	}

	if !grupoExiste {
		fmt.Println("El grupo no existe, vuelva a intentarlo.")
		return
	}

	reescribirArchivo(archivo, super, int64(posicionInicial), concatenarContenido, "users.txt")
mensaje += leerArchivo(archivo, super, int64(posicionInicial), "users.txt") + "\n"
	mensaje += "¡ Grupo eliminado exitosamente !\n"
	archivo.Close()
}

/* Metodo que crea un usuario. */
func crearUsuario(usuario string, contrasenia string, grupoPertenece string) {

	if !sesionIniciada {
		mensaje += "Debe existir una sesion iniciada para crear un grupo, vuelva  intentarlo.\n"
		return
	}

	if usuarioActual != "root" {
		mensaje += "Solo el usuario [root] puede utilizar este comando.\n"
		return
	}

	if len(contrasenia) > 10 && len(usuario) > 10 && len(grupoPertenece) > 10 {
		mensaje += "El nombre/grupo/contraseña no pueden tener mas de 10 caracteres, vuelva a intentarlo..\n"
		return
	}

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		mensaje += "Particion no esta montado, no es posible realizar el reporte..\n"
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		mensaje += "Disco no existe, no es posible realizar el reporte..\n"
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")
	grupoExiste := false
	usuarioRepetido := false
	idUsuario := 0

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if !grupoExiste {
			if letra[0] != "0" {
				if letra[1] == "G" {
					if !usuarioRepetido {
						if letra[2] == grupoPertenece {
							idUsuario = idUsuario + 1
							fmt.Println("ID ", idUsuario)
							grupoExiste = true
							contenido += strconv.Itoa(int(idUsuario)) + "," + "U" + "," + grupoPertenece + "," + usuario + "," + contrasenia + "\n"
						}
					}

				} else {
					if letra[3] == usuario {
						usuarioRepetido = true
					}
				}
			}
		}
	}

	if !grupoExiste {
		fmt.Println("Grupo no existe, no es posible crear un usuario, vuelva a intentarlo.")
		return
	}

	if usuarioRepetido {
		fmt.Println("EL usuario debe ser unico, vuelva a intentarlo..")
		return
	}

	reescribirArchivo(archivo, super, int64(posicionInicial), contenido, "users.txt")
	mensaje += leerArchivo(archivo, super, int64(posicionInicial), "users.txt") + "n"
	mensaje += "¡ Usuario creado exitosamente !\n"
	archivo.Close()
}

/* Metodo que elimina un usuario dado el nombre del usuario. */
func eliminarUsuario(nombreUsuario string) {

	if !sesionIniciada {
		mensaje += "Debe existir una sesion iniciada para crear un grupo, vuelva  intentarlo.\n"
		return
	}

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		mensaje += "Particion no esta montado, no es posible realizar el reporte..\n"
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		mensaje += "Disco no existe, no es posible realizar el reporte..\n"
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")
	concatenarContenido := ""
	grupoExiste := false

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if letra[0] != "0" {
			if letra[1] == "U" {
				if letra[3] == nombreUsuario {
					letra[0] = "0"
					grupoExiste = true
				}
			}
		}

		if letra[1] == "U" {
			concatenarContenido += letra[0] + "," + letra[1] + "," + letra[2] + "," + letra[3] + "," + letra[4] + "\n"
		} else {
			concatenarContenido += letra[0] + "," + letra[1] + "," + letra[2] + "\n"
		}
	}

	if !grupoExiste {
		mensaje += "El usuario no existe, vuelva a intentarlo.\n"
		return
	}

	reescribirArchivo(archivo, super, int64(posicionInicial), concatenarContenido, "users.txt")
	mensaje += leerArchivo(archivo, super, int64(posicionInicial), "users.txt") + "\n"
	mensaje += "¡ Usuario eliminado exitosamente !\n"
	archivo.Close()
}

/* Funcion que retorna el contenido de un archivo. */
func leerArchivo(archivo *os.File, super SuperBloque, posicion int64, nombreArchivo string) string {

	contenido := ""
	inodo := obtenerInodo(archivo, posicion)
	if inodo.Tipo == int64(0) {
		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != int64(-1) {
				carpeta := obtenerBloqueCarpetas(archivo, inodo.Bloque[i])
				for j := 0; j < 4; j++ {
					nombreLimpio := cadenaLimpia(carpeta.Contenidos[j].Nombre[:])
					if carpeta.Contenidos[j].Apuntador != int64(-1) && nombreLimpio == nombreArchivo {
						inodoAux := obtenerInodo(archivo, carpeta.Contenidos[j].Apuntador)
						for k := 0; k < 15; k++ {
							if inodoAux.Bloque[k] != int64(-1) {
								archivo := obtenerBloqueArchivo(archivo, inodoAux.Bloque[k])
								contenido += cadenaLimpia(archivo.Datos[:])
							}
						}
						if inodoAux.Bloque[14] == -1 {
							return contenido
						}
					}
				}
			}
		}
	}
	return "null"
}

/* Funcion que analiza los parametros correspondiente para la creacion de carpetas */
func crearCarpeta(ruta string, padre bool) {

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte.")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte.")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	inicioPart, _ := obtenerInicioTamanio(nombreParticion, discoAux)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos
	contador := 0

	rutaLimpia := ""
	for i := 1; i < len(ruta); i++ {
		rutaLimpia += string(ruta[i])
	}

	ultimaCarpetaAlreves := ""
	for j := len(ruta) - 1; j >= 0; j-- {
		if string(ruta[j]) == "/" {
			break
		}
		contador++
		ultimaCarpetaAlreves += string(ruta[j])
	}
	
	ultimaCarpeta := ""
	for j := len(ultimaCarpetaAlreves) - 1; j >= 0; j-- {
		ultimaCarpeta += string(ultimaCarpetaAlreves[j])
	}
	
	nombrePadreAux := ""
	for k := len(ruta) - contador; k >= 0; k-- {
		nombrePadreAux += string(ruta[k])
	}
	
	rutaPadre := ""
	for l := len(nombrePadreAux) - 1; l > 0; l-- {
		rutaPadre += string(nombrePadreAux[l])
	}

	if padre {
		mkdirCOmplemento(archivo, super, rutaLimpia+"/", int64(posicionInicial), inicioPart)
	} else {

		mismaCarpeta := buscarCarpetaArchivo(archivo, int64(posicionInicial), rutaLimpia+"/", ultimaCarpeta)
		carpetaPadre := buscarInodo(archivo, int64(posicionInicial), rutaPadre)

		if mismaCarpeta == 1 {
			mensaje += "¡ No es posible crear la misma carpetas !\n"
			return
		} else if carpetaPadre == 1 {
			mensaje += "¡ No es posible crear las carpetas, no existen carpetas padres !\n"
			return
		} else {
			mkdirCOmplemento(archivo, super, rutaLimpia+"/", int64(posicionInicial), inicioPart)
		}
	}
}

/* 
	Metodo que de acuerdo a la ruta enviada busca la posicion de la carpeta si la encuentra retorna un valor de lo contrario sera
	cero y de ser asi entonces quiere decir que hay que crear una carpeta nueva.
 */
var pos int = 0
func mkdirCOmplemento(archivo *os.File, super SuperBloque, ruta string, posicionTablaInodos int64, inicioParticion int64) {

	nombre := ""
	for i := 0; i < len(ruta); i++ {
		if string(ruta[i]) == "/" {
			pos = int(i + 1)
			break
		}
		nombre += string(ruta[i])
	}

	retoRuta := ""
	for j := pos; j < len(ruta); j++ {
		retoRuta += string(ruta[j])
	}

	posicion := buscarInodo(archivo, posicionTablaInodos, nombre+"/")

	if len(nombre) > 0 {
		if posicion == 0 {
			crearDirectorio(archivo, super, nombre, posicionTablaInodos)
			otraPosicion := buscarInodo(archivo, posicionTablaInodos, nombre+"/")
			if len(nombre) > 0 {
				mkdirCOmplemento(archivo, super, retoRuta, otraPosicion, inicioParticion)
			}
			return
		} else {
			mkdirCOmplemento(archivo, super, retoRuta, posicion, inicioParticion)
		}
	}
}

/* FUncion que busca la posicion de un archivo o carpeta dada su ruta, retorna cero si no la encuentra. */
var posicionEFE int = 0
func buscarInodo(archivo *os.File, posicion int64, ruta string) int64 {

	inodo := obtenerInodo(archivo, posicion)

	nombre := ""
	for i := 0; i < len(ruta); i++ {
		if string(ruta[i]) == "/" {
			posicionEFE = int(i + 1)
			break
		}
		nombre += string(ruta[i])
	}

	retoRuta := ""
	for j := posicionEFE; j < len(ruta); j++ {
		retoRuta += string(ruta[j])
	}

	if inodo.Tipo == 0 {
		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != -1 {
				carpeta := obtenerBloqueCarpetas(archivo, inodo.Bloque[i])
				for j := 0; j < 4; j++ {
					nombreLimpio := cadenaLimpia(carpeta.Contenidos[j].Nombre[:])
					if nombreLimpio == nombre {
						if len(retoRuta) == 0 {
							return carpeta.Contenidos[j].Apuntador
						} else {

							buscarInodo(archivo, carpeta.Contenidos[j].Apuntador, retoRuta)
						}
					}
				}
			}
		}
		if inodo.Bloque[14] == -1 {
			return 0
		}
	}
	return 0
}

/* Funcion que crea el inodo y los bloques correspondiente para la creacion de una carpeta. */
func crearDirectorio(archivo *os.File, super SuperBloque, nombreCarpeta string, posicionInodo int64) {

	inodo := obtenerInodo(archivo, posicionInodo)
	if inodo.Tipo == int64(0) {
		if inodo.Bloque[0] == int64(-1) {
			nuevoBloque := BloqueCarpeta{}
			copy(nuevoBloque.Contenidos[0].Nombre[:], ".")
			nuevoBloque.Contenidos[0].Apuntador = 0
			copy(nuevoBloque.Contenidos[1].Nombre[:], "..")
			nuevoBloque.Contenidos[1].Apuntador = 0
			copy(nuevoBloque.Contenidos[2].Nombre[:], "")
			nuevoBloque.Contenidos[2].Apuntador = -1
			copy(nuevoBloque.Contenidos[3].Nombre[:], "")
			nuevoBloque.Contenidos[3].Apuntador = -1
			posicionBloque := escribirBloqueCarpeta(archivo, super, nuevoBloque)
			inodo.Bloque[0] = posicionBloque
			reescribirInodo(archivo, posicionInodo, inodo)
		}

		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != int64(-1) {
				posicionBloque := inodo.Bloque[i]
				bloqueCarpeta := obtenerBloqueCarpetas(archivo, posicionBloque)
				for j := 0; j < 4; j++ {
					if bloqueCarpeta.Contenidos[j].Apuntador == int64(-1) {
						nuevoInodo := TablaInodos{
							IdUsuario:      1,
							IdGrupo:        1,
							TamanioArchivo: 0,
							Tipo:           int64(0),
							Permisos:       664,
						}
						copy(nuevoInodo.FechaCreacion[:], obtenerFecha())
						copy(nuevoInodo.FechaLectura[:], obtenerFecha())
						copy(nuevoInodo.FechaModificacion[:], obtenerFecha())
						for k := 0; k < 15; k++ {
							nuevoInodo.Bloque[k] = -1
						}
						posicionNuevoInodo := escribirInodo(archivo, super, nuevoInodo)
						bloqueCarpeta.Contenidos[j].Apuntador = posicionNuevoInodo
						copy(bloqueCarpeta.Contenidos[j].Nombre[:], nombreCarpeta)
						reescribirBloque(archivo, inodo.Bloque[i], bloqueCarpeta)
						return
					}
				}

			} else {
				nuevoBloque := BloqueCarpeta{}
				copy(nuevoBloque.Contenidos[0].Nombre[:], "")
				nuevoBloque.Contenidos[0].Apuntador = -1
				copy(nuevoBloque.Contenidos[1].Nombre[:], "")
				nuevoBloque.Contenidos[1].Apuntador = -1
				copy(nuevoBloque.Contenidos[2].Nombre[:], "")
				nuevoBloque.Contenidos[2].Apuntador = -1
				copy(nuevoBloque.Contenidos[3].Nombre[:], "")
				nuevoBloque.Contenidos[3].Apuntador = -1
				posicionBloque := escribirBloqueCarpeta(archivo, super, nuevoBloque)
				inodo.Bloque[i] = posicionBloque
				reescribirInodo(archivo, posicionInodo, inodo)
				i--
			}
		}
	}
}

/* Busca una carpeta o archivo dada una ruta, si la encuentra retornara 1. */
var posicionBuscar int = 0
func buscarCarpetaArchivo(archivo *os.File, posicion int64, ruta string, nombreCarpeta string) int64 {

	inodo := obtenerInodo(archivo, posicion)

	nombre := ""
	for i := 0; i < len(ruta); i++ {
		if string(ruta[i]) == "/" {
			posicionBuscar = int(i + 1)
			break
		}
		nombre += string(ruta[i])
	}

	retoRuta := ""
	for j := posicionBuscar; j < len(ruta); j++ {
		retoRuta += string(ruta[j])
	}

	if inodo.Tipo == 0 {
		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != -1 {
				carpeta := obtenerBloqueCarpetas(archivo, inodo.Bloque[i])
				for j := 0; j < 4; j++ {
					nombreLimpio := cadenaLimpia(carpeta.Contenidos[j].Nombre[:])
					if nombreLimpio == nombre {
						if len(retoRuta) == 0 {
							return 1
						} else {
							return buscarCarpetaArchivo(archivo, carpeta.Contenidos[j].Apuntador, retoRuta, nombreCarpeta)
						}
					}
				}
			}
		}
		if inodo.Bloque[14] == -1 {
			return 0
		}
	}
	return 0
}

/* Metodo que reescribe el archivo de users.txt  */
func reescribirArchivo(archivo *os.File, super SuperBloque, posicion int64, contenido string, nombreArchivo string) {

	inodo := obtenerInodo(archivo, posicion)
	if inodo.Tipo == int64(0) {
		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != int64(-1) {
				carpeta := obtenerBloqueCarpetas(archivo, inodo.Bloque[i])
				for j := 0; j < 4; j++ {
					nombreLimpio := cadenaLimpia(carpeta.Contenidos[j].Nombre[:])
					if carpeta.Contenidos[j].Apuntador != int64(-1) && nombreLimpio == nombreArchivo {
						inodoAux := obtenerInodo(archivo, carpeta.Contenidos[j].Apuntador)
						for k := 0; k < 15; k++ {
							if inodoAux.Bloque[k] != int64(-1) {
								archivoUsr := obtenerBloqueArchivo(archivo, inodoAux.Bloque[k])
								copy(archivoUsr.Datos[:], contenido)
								reescribirBloqueArchivo(archivo, inodoAux.Bloque[k], archivoUsr)
							}
						}
						reescribirInodo(archivo, carpeta.Contenidos[j].Apuntador, inodoAux)
					}
				}
			}
		}
	}
}

/* Funcion que obtiene el numero de bloques y carpetas que contendra nuestro sistema de archivos. */
func obtenerNumeroEstructuras(tamanioParticion int64) int64 {

	tamanioSuperBloque := unsafe.Sizeof(SuperBloque{})
	tamanioInodo := unsafe.Sizeof(TablaInodos{})
	tamanioBloque := unsafe.Sizeof(BloqueArchivos{})

	numerador := tamanioParticion - int64(tamanioSuperBloque)
	denominador := 4 + tamanioInodo + (3 * tamanioBloque)

	return numerador / int64(denominador)
}


func createdJson(id string)  {
	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)

	iNodo := TablaInodos{}
	posicionInodo := int64(super.InicioTablaInodos) + (int64(0) * int64(unsafe.Sizeof(TablaInodos{})))
	iNodo = obtenerInodo(archivo, posicionInodo)

	grafo := "{\n\"nombre\": \"/\",\n"
	r := recorrerJson(super, archivo, iNodo)
	r = r[:len(r)-1]
	grafo += "\"hijos1\": [\n " + r + "]},"
	fmt.Println("GRAFOOOOOOOOOOOOOOOOOOO \n", grafo)
	

}

func recorrerJson(superAux SuperBloque, archivo *os.File, iNodo TablaInodos) string {
	grafo := ""
	for i := 0; i < 15; i++{
	  if iNodo.Bloque[i] != -1 {
		if iNodo.Tipo == int64(0) {
		  bloqueCarpeta := BloqueCarpeta{}
		  contenidoCarpeta := make([]byte, int(unsafe.Sizeof(bloqueCarpeta)))
		  archivo.Seek(iNodo.Bloque[i], 0)
		  archivo.Read(contenidoCarpeta)
		  buffer2 := bytes.NewBuffer(contenidoCarpeta)
		  binary.Read(buffer2, binary.BigEndian, &bloqueCarpeta)
  
		  for j := 0; j < 4; j++ {
			fmt.Println("NOMBRES " , cadenaLimpia(bloqueCarpeta.Contenidos[j].Nombre[:]))
			if cadenaLimpia(bloqueCarpeta.Contenidos[j].Nombre[:]) == "." || cadenaLimpia(bloqueCarpeta.Contenidos[j].Nombre[:]) == ".." {
  
			}else{
			  puntoIndex := strings.Index(cadenaLimpia(bloqueCarpeta.Contenidos[j].Nombre[:]), ".")
						if puntoIndex == -1 {
							iNodoNuevo := TablaInodos{}
							contenido := make([]byte, int(unsafe.Sizeof(iNodoNuevo)))
							archivo.Seek(bloqueCarpeta.Contenidos[j].Apuntador, 0)
							archivo.Read(contenido)
							buffer := bytes.NewBuffer(contenido)
							binary.Read(buffer, binary.BigEndian, &iNodoNuevo)
							grafo += "{\"nombre1\":\"" + cadenaLimpia(bloqueCarpeta.Contenidos[j].Nombre[:]) + "\",\n"
							grafo += recorrerJson(superAux, archivo, iNodoNuevo) + "},"
			  } else {
							grafo += "{\"nombre2\":\"" + cadenaLimpia(bloqueCarpeta.Contenidos[j].Nombre[:]) + "\",\n"
			  }
			}
		  }
		  return grafo
		}else if iNodo.Tipo == int64(1){
		  return grafo
		}
	  }
	}
	return ""
  }