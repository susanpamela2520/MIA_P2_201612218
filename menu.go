package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var contador = 0;
var mensaje = "";


func inicio() {
	lector := bufio.NewReader(os.Stdin)
	mensaje += "\n******************* MIA PROYECTO 1 *******************"
	for {

		fmt.Print("Ingrese un comando : ")
		entrada, _ := lector.ReadString('\n')
		entrada = strings.ReplaceAll(entrada, "\n", "")
		entrada = strings.ReplaceAll(entrada, "\r", "")
		analizardor(strings.Split(entrada, " "))
	}
}


func analizardor(comando []string) {

	if strings.Contains(comando[0], "#") {
		mensaje += strings.Join(comando, " ")
	} else {

		cadena := strings.ToLower(comando[0])
		switch cadena {
		case "mkdisk":
			mensaje += "\nCREACION DE DISCO\n"
			comandoMkdisk(comando)
		case "rmdisk":
			mensaje += "\nELIMINACION DE DISCO\n"
			comandoRmdisk(comando)
		case "fdisk":
			mensaje += "\nCREACION PARTICIONES\n"
			comandoFdisk(comando)
		case "mount":
			mensaje += "\nMONTAR PARTICIONES\n"
			comandoMount(comando)
		case "unmount":
			mensaje += "\nDESMONTAR PARTICIONES\n"
			comandoUnmount(comando)
		case "mkfs":
			mensaje += "\nSISTEMA DE ARCHIVOS\n"
			comandoMkfs(comando)
		case "login":
			mensaje += "\nINICIAR SESION\n"
			comandoLogin(comando)
		case "logout":
			mensaje += "\nCERRAR SESION\n"
			comandoLogout(comando)
		case "mkgrp":
			mensaje += "\nCREAR GRUPOS\n"
			comandoMkgrp(comando)
		case "rmgrp":
			mensaje += "\nELIMINAR GRUPO\n"
			comandoRmgrp(comando)
		case "mkusr":
			mensaje += "\nCREAR USUARIOS\n"
			comandoMkuser(comando)
		case "rmusr":
			mensaje += "\nELIMINAR USUARIO\n"
			comandoRmusr(comando)
		case "mkdir":
			mensaje += "\nCREANDO CARPETAS\n"
			comandoMkdir(comando)
		case "rep":
			mensaje += "\nREPORTES\n"
			comandoReporte(comando)
		/* case "pause":
			mensaje += "\nPAUSE"
			pause() */
		case "execute":
			mensaje += "\nLEYENDO ARCHIVO\n"
			comandoExec(comando)
		case "exit":
			mensaje += "\nADIOS"
			exit()
		default:
			mensaje += "\nComando [" + cadena + "] no existe, vuelva a intentarlo..\n"

		}
	}
}

func comandoMkdisk(comando []string) {

	tamanio := int64(-1)
	ajuste := ""
	unidad := ""
	letra := obtenerLetra(contador)
	ruta := obtenerRuta("MIA/P1/" + letra + ".dsk")
	contador++;

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-size":
			tamanio = obtenerTamanio(valor[1])
		case "-fit":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "ff") == 0) || (strings.Compare(aux2, "bf") == 0) || (strings.Compare(aux2, "wf") == 0) {
				ajuste = aux2
			}

		case "-unit":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "k") == 0) || (strings.Compare(aux2, "m") == 0) {
				unidad = aux2
			}

		default:
			errores(aux1)
		}
	}

	if ajuste == "" {
		ajuste = "ff"
	}

	if unidad == "" {
		unidad = "m"
	}

	if tamanio > 0 && ruta != "" {
		crearDisco(tamanio, ajuste, unidad, ruta)
	} else {
		mensaje += "¡ Faltan parametros obligatorios [MKDISK], vuelva a intentarlo !\n"
	}
}

func comandoRmdisk(comando []string) {

	ruta := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-driveletter":
			ruta = obtenerRuta("MIA/P1/" + valor[1] + ".dsk")
		default:
			errores(aux1)
		}
	}

	if ruta != "" {
		eliminarDisco(ruta)
	} else {
		mensaje += "¡ Faltan parametros obligatorios [RMDISK], vuelva a intentarlo !\n"
	}
}

func comandoFdisk(comando []string) {

	tamanio := int64(-1)
	unidad := ""
	ruta := ""
	tipoParticion := ""
	ajuste := ""
	nombre := ""
	borra := false
	agregar := false
	tipoEliminar := ""
	valorAgregar := int64(-1)

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-size":
			tamanio = obtenerTamanio(valor[1])
		case "-unit":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "k") == 0) || (strings.Compare(aux2, "m") == 0) || (strings.Compare(aux2, "b") == 0) {
				unidad = aux2
			}

		case "-driveletter":
			ruta = obtenerRuta("MIA/P1/" + valor[1] + ".dsk")
			mensaje += "ruta " + ruta
		case "-type":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "p") == 0) || (strings.Compare(aux2, "e") == 0) || (strings.Compare(aux2, "l") == 0) {
				tipoParticion = aux2
			}

		case "-fit":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "ff") == 0) || (strings.Compare(aux2, "bf") == 0) || (strings.Compare(aux2, "wf") == 0) {
				ajuste = aux2
			}

		case "-name":
			nombre = valor[1]
		case "-delete":
			tipoEliminar = valor[1]
			borra = true
		case "-add":
			valorAgregar = obtenerTamanio(valor[1])
			agregar = true
		default:
			errores(aux1)
		}
	}

	if tamanio > 0 && ruta != "" && nombre != ""{
		if unidad == "" {
			unidad = "k"
		}
	
		if tipoParticion == "" {
			tipoParticion = "p"
		}
	
		if ajuste == "" {
			ajuste = "wf"
		}

		insertarParticion(tamanio, unidad, ruta, tipoParticion, ajuste, nombre)

	}else if nombre != "" && ruta != "" && borra {
		borrarParticion(tipoEliminar, ruta, nombre)
	} else if nombre != "" && ruta != "" && agregar{
		agregarParticion(ruta, nombre, valorAgregar, unidad)
	}else {
		mensaje += "¡ Faltan parametros obligatorios [FDISK], vuelva a intentarlo !\n"
	}
}

func comandoMount(comando []string) {

	ruta := ""
	nombre := ""
	letraDisco := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-driveletter":
			letraDisco = valor[1]
			ruta = obtenerRuta("MIA/P1/" + valor[1] + ".dsk")
		case "-name":
			nombre = valor[1]
		default:
			errores(aux1)
		}
	}

	if ruta != "" && nombre != "" {
		montarParticion(ruta, nombre, letraDisco)
	} else {
		mensaje += "¡ Faltan parametros obligatorios [MOUNT], vuelva a intentarlo !\n"
	}
}


func comandoUnmount(comando []string){
	identificador := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-id":
			identificador = valor[1]
		default:
			errores(aux1)
		}
	}

	if identificador != "" {
		desmotar(identificador)
	} else {
		mensaje += "¡ Faltan parametros obligatorios [UNMOUNT], vuelva a intentarlo !\n"
	}
}

func comandoMkfs(comando []string){


	identificador := ""
	tipoFormateo := ""
	tipoSistema := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-id":
			identificador = valor[1]
		case "-type":
			tipoFormateo = strings.ToLower(valor[1])
		case "-fs":
			tipoSistema = valor[1]
		default:
			errores(aux1)
		}
	}

	if tipoFormateo == "" {
		tipoFormateo = "full"
	}

	if tipoSistema == "" {
		tipoSistema = "2fs"
	}

	if identificador != "" && tipoSistema == "2fs"{
		crearSistemaArchivosEXT2(identificador, tipoFormateo)
	}else if identificador != "" && tipoSistema == "3fs"{
		crearSistemaArchivosEXT3(identificador, tipoFormateo)
	} else{
		mensaje += "¡ Faltan parametros obligatorios [MKFS], vuelva a intentarlo !\n"
	}
}

func comandoLogin(comando []string){

	usuario := ""
	contrasenia := ""
	id := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-user":
			usuario = valor[1]
		case "-pass":
			contrasenia = strings.ToLower(valor[1])
		case "-id":
			id = valor[1]
		default:
			errores(aux1)
		}
	}

	if id != "" && usuario != "" && contrasenia != "" {
		iniciarSesion(usuario, contrasenia, id)
	}else{
		mensaje += "¡ Faltan parametros obligatorios [LOGIN], vuelva a intentarlo !\n"
	}
}

func comandoLogout(comando []string){
	cerrarSesion()
}

func comandoMkgrp (comando[]string){
	
	grupo := ""
	
	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-name":
			grupo = valor[1]
		default:
			errores(aux1)
		}
	}

	if grupo != "" {
		crearGrupo(grupo)
	}else{
		mensaje += "¡ Faltan parametros obligatorios [MKGRP], vuelva a intentarlo !\n"
	}
}

func comandoRmgrp (comando[]string){
	
	grupo := ""
	
	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-name":
			grupo = valor[1]
		default:
			errores(aux1)
		}
	}

	if grupo != "" {
		eliminarGrupo(grupo)
	}else{
		mensaje += "¡ Faltan parametros obligatorios [RMGRP], vuelva a intentarlo !\n"
	}
}

func comandoMkuser(comando []string){

	usuario := ""
	contrasenia := ""
	grupoPertenece := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-user":
			usuario = valor[1]
		case "-pass":
			contrasenia = strings.ToLower(valor[1])
		case "-grp":
			grupoPertenece = valor[1]
		default:
			errores(aux1)
		}
	}

	if grupoPertenece != "" && usuario != "" && contrasenia != "" {
		crearUsuario(usuario, contrasenia, grupoPertenece)
	}else{
		mensaje += "¡ Faltan parametros obligatorios [MKUSR], vuelva a intentarlo !\n"
	}
}

func comandoRmusr(comando[]string){

	usuario := ""
	
	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-user":
			usuario = valor[1]
		default:
			errores(aux1)
		}
	}

	if usuario != "" {
		eliminarUsuario(usuario)
	}else{
		mensaje += "¡ Faltan parametros obligatorios [RMUSR], vuelva a intentarlo !\n"
	}
}

func comandoMkdir(comando[]string){

	ruta := ""
	padre := false

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-path":
			ruta = valor[1]
		case "-r":
			padre = true
		default:
			errores(aux1)
		}
	}

	if ruta != "" {
		crearCarpeta(ruta, padre)
	}else{
		mensaje += "¡ Faltan parametros obligatorios [MKDIR], vuelva a intentarlo !\n"
	}
}

func comandoReporte(comando []string){

	ruta := ""
	identificador := ""
	nombre := ""
	rutaOpcional := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-name":
			nombre = valor[1]
		case "-path":
			ruta = obtenerRuta(valor[1])
		case "-id":
			identificador = valor[1]
		case "-ruta":
			rutaOpcional = valor[1]
		default:
			errores(aux1)
		}
	}

	if nombre != "" && ruta != "" && identificador != "" {
		analizarReporte(nombre, ruta, identificador, rutaOpcional)
	}else{
		mensaje += "¡ Faltan parametros obligatorios [REP], vuelva a intentarlo !\n"
	}
}


func comandoExec(comando []string) {

	ruta := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")
		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-path":
			ruta = obtenerRuta(valor[1])
		default:
			errores(aux1)
		}
	}

	if ruta != "" {
		leerArchivoEntrada(ruta)
	}
}

func leerArchivoEntrada(ruta string) {

	archivo, error := os.Open(ruta)

	if error != nil {
		mensaje += "¡ Error, el archivo no existe, vuelva a intentarlo !\n"
		inicio()
		archivo.Close()
	}

	scanner := bufio.NewScanner(archivo)

	for scanner.Scan() {

		linea := scanner.Text()

		concatenar := linea

		if concatenar != "" {

			if strings.Contains(string(concatenar[0]), "#") {
				mensaje += concatenar + "\n"
				concatenar = ""
			} else {
				mensaje += concatenar
				analizardor(strings.Split(concatenar, " "))
				concatenar = ""
			}

		}
	}
}

func obtenerRuta(valor string) string {

	ruta := ""
	if strings.Contains(valor, "\"") {
		ruta = strings.ReplaceAll(valor, "\"", "")
	} else {
		ruta = valor
	}

	return ruta
}

func obtenerTamanio(valor string) int64 {

	tamanio, _ := strconv.Atoi(valor)
	if tamanio > 0 {
		return int64(tamanio)
	}
	return -1
}

func errores(comando string) {
	if comando != "" {
		mensaje += "El comando ["+comando+"] no es reconocido, ingrese un comando valido.\n"
	}
}

func pause() {
	mensaje += "Presione cualquier tecla para continuar .....\n"
	tecla := ""
	fmt.Scanln(&tecla)
}

func exit() {
	mensaje += "¡ Finalizacion del programa realizado exitosamente !\n"
	os.Exit(0)
}