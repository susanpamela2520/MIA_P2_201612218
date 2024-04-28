package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nfnt/resize"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

var imagenes []string;

func loadEnv()  {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}
}

func uploadImage(url string){
	loadEnv()

	fileName := filepath.Base(url)

	file, err := os.Open(url)
    if err != nil {
        fmt.Println("Error al abrir la imagen:", err)
        return
    }
    defer file.Close()

    img, _, err := image.Decode(file)
    if err != nil {
        fmt.Println("Error al decodificar la imagen:", err)
        return
    }

    img = resize.Resize(800, 600, img, resize.Lanczos3)

    var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
					fmt.Println("Error al codificar la imagen:", err)
					return
    }

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			""),
	})

	if err != nil {
		fmt.Println("error 1")
		return
	}

	clientS3 := s3.New(sess)

	_, err = clientS3.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("BUCKET_NAME")),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image"),
	})

	if err != nil {
		fmt.Println("error 2", err)
		return
	}
}

func downloadImage()  {

	loadEnv()
	
    sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})

	if err != nil {
		fmt.Println("Error al crear sesión de AWS:", err)
		return
	}
    svc := s3.New(sess)

    resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
        Bucket: aws.String(os.Getenv("BUCKET_NAME")),
    })

    if err != nil {
        fmt.Println("Error al listar objetos en S3:", err)
        return
    }

	imagenes = []string{}
	for _, objeto := range resp.Contents {
		valor := *objeto.Key
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(os.Getenv("BUCKET_NAME")),
			Key:    aws.String(valor),
		})

		urlPublica, err := req.Presign(3600 * time.Minute)
		if err != nil {
			fmt.Println("Error al obtener la URL pública del objeto en S3:", err)
			continue
		}

		imagenes = append(imagenes, urlPublica)
	}
}