package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func analyze(c *fiber.Ctx) error {
	body := c.Body()
	nombreArchivo := string(body)
	ruta := "-path=../" + nombreArchivo
	comando := [2]string{"execute",ruta}

	analizardor(comando[:])

	return c.Status(fiber.StatusOK).JSON(mensaje)
}

func getDisk(c *fiber.Ctx) error {

	type Disk struct {
		Path string `json:"path"`
	}

	var discos []Disk

	for _, disco := range listaDiscos {
		discos = append(discos, Disk{Path: disco})
	}

	return c.Status(fiber.StatusOK).JSON(discos)
}

func getPartition(c *fiber.Ctx) error {

	pathDisk := c.Params("idDisk")
	archivo := obtenerDisco("MIA/P1/" + pathDisk)
	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible crear una particion sin un disco..")
		return c.Status(fiber.StatusNotFound).JSON("Disco no existe, no es posible crear una particion sin un disco..")
	}

	discoAux := obtenerMBR(archivo)

	type Partition struct {
		Name string `json:"name"`
	}

	var partitions []Partition

	for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Inicio != -1 {
			partitions = append(partitions, Partition{Name: cadenaLimpia(discoAux.Particiones[i].Nombre[:])})
		}
	}

	return c.Status(fiber.StatusOK).JSON(partitions)
}

func verifyMount(c *fiber.Ctx) error {

	type Data struct {
		IdDisk string `json:"IdDisk"`
		IdPartition string `json:"IdPartition"`
	}

	var datos Data
	if err := c.BodyParser(&datos); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	
	nombreParticion := obtenerParticionMontada(datos.IdPartition)

	type Condition struct {
		Exist bool `json:"exist"`
	}
	var partitions Condition

	if cadenaLimpia(nombreParticion[:]) == "" {
		partitions.Exist = false
	} else {
		partitions.Exist = true
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.JSON(partitions)
}

func login(c *fiber.Ctx) error {
	type Usuario struct {
		User string `json:"user"`
		Password string `json:"password"`
		IdPartition string `json:"IdPartition"`
	}

	var iniciando Usuario
	if err := c.BodyParser(&iniciando); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	encontrado, validacion := iniciarSesion(iniciando.User, iniciando.Password, iniciando.IdPartition)
	fmt.Println("datos del usuario validacion ", validacion)

	if !encontrado {
		return c.Status(fiber.StatusBadRequest).JSON(encontrado)
	}

	return c.Status(fiber.StatusOK).JSON(encontrado)
}

func logout(c *fiber.Ctx) error {
	existeSesion, validacion := cerrarSesion()
	fmt.Println(validacion)
	return c.Status(fiber.StatusOK).JSON(existeSesion)
}

func obtenerImg(c *fiber.Ctx) error{

	downloadImage()
	type Image struct {
		URL string `json:"url"`
	}

	var pathPool []Image

	for _, path := range imagenes {
		pathPool = append(pathPool, Image{URL: path})
	}

	return c.Status(fiber.StatusOK).JSON(pathPool)
}

func getTxt(c *fiber.Ctx) error {

	idPartition := c.Params("id")
	fmt.Println("id ", idPartition)
	rutaObtenida := obtenerDiscoMontado(idPartition)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..\n")
		return c.Status(fiber.StatusNotFound).JSON("Particion no esta montado, no es posible realizar el reporte..")
	}

	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..\n")
		return c.Status(fiber.StatusNotFound).JSON("Disco no existe, no es posible realizar el reporte....")
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	return c.Status(fiber.StatusOK).JSON(contenido)
}

func iniciarServidor(){
	/* Creando una instancia del servidor */
	app := fiber.New()

	/* Cors  */
	app.Use(cors.New())
	app.Use(cors.New(cors.Config{
		/* Cambiar por ruta del otro servidor */
		/* AllowOrigins: "https://gofiber.io, https://gofiber.net", */
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	
	/* Rutas del servidor */
	app.Get("/hello", func(c *fiber.Ctx) error {
        return c.Status(fiber.StatusOK).JSON("Hola mundo desde Fiber y Golang!!")
    })

	app.Post("/login", login)
	app.Post("/analyze", analyze)
	app.Post("/verifyMount", verifyMount)
	app.Get("/logout", logout)
	app.Get("/getDisk", getDisk)
	app.Get("/getImg", obtenerImg)
	app.Get("/getTxt/:id", getTxt)
	app.Get("/getPartition/:idDisk", getPartition)

	/* Puerto donde escucha el servidor */
	app.Listen(":4000")
}