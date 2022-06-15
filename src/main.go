package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	bucketname         = "marfiles"
	targetDir          = "/tmp/"
	batch_size, _      = strconv.Atoi(getenv("BatchSize", "1"))
	initial_workers, _ = strconv.Atoi(getenv("InitialWorker", "1"))
	Sleep_Period, _    = strconv.Atoi(getenv("SleepPeriod", "1"))
)

func main() {
	godotenv.Load(".env")

	modellist := make(map[string]int, 20)
	if len(os.Args) > 1 {
		targetDir = os.Args[1]
	}
	log.Printf("TargetPath : %s", targetDir)

	endpoint := getenv("S3Url", "minio.url.com")
	accessKeyID := getenv("S3AccessKey", "XXXX")
	secretAccessKey := getenv("S3SecretKey", "XXXX")
	useSSL := true

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	opts := minio.ListObjectsOptions{
		Recursive: true,
	}

	for true {

		for object := range minioClient.ListObjects(context.Background(), bucketname, opts) {
			if object.Err != nil {
				fmt.Println(object.Err)
				return
			}
			if val, ok := modellist[object.Key]; ok {
				if val != int(object.Size) {
					// Model has change in Size -> New Version ;)
					DownloadModel(minioClient, object.Key)
					RegisterNewModel("http://localhost:8081/models", object.Key)
					modellist[object.Key] = int(object.Size)
					log.Printf("Model %s has changed. Adding....", object.Key)
				}
			} else {
				// Model does not exist - DOwnload And Register
				DownloadModel(minioClient, object.Key)
				RegisterNewModel("http://localhost:8081/models", object.Key)
				modellist[object.Key] = int(object.Size)
				log.Printf("Model %s is new. Adding....", object.Key)
			}

		}
		log.Printf("Sleeping for %d Minutes", Sleep_Period)
		time.Sleep(time.Duration(Sleep_Period) * time.Minute)
	}
	return
}

func DownloadModel(minioClient *minio.Client, filename string) {
	target := fmt.Sprintf("%s/%s", targetDir, filename)
	if err := minioClient.FGetObject(context.Background(), bucketname, filename,
		target, minio.GetObjectOptions{}); err != nil {
		log.Fatalln(err)
	}
	err := os.Chmod(target, 0777)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("File %s to %s downloaded.", filename, target)
}

func RegisterNewModel(url string, modelname string) error {
	url = fmt.Sprintf("%s?url=%s&batchsize=%d&initial_workers=%d", url, modelname, batch_size, initial_workers)
	log.Printf("Calling Url: %s", url)
	req, err := http.NewRequest("POST", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	resbody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(resbody))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Result: %v", resp)

	return nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
