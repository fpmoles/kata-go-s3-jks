package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(getLogLevel())
}

func main() {
	config := &aws.Config{
		Region: aws.String("us-east-2"),
	}
	dd:=getDownloadDetails()
	file, err:= downloadFileFromBucket(dd, config)
	if err != nil{
		log.Errorf("Error while downloading file")
		os.Exit(1)
	}
	if file == nil{
		fmt.Print("File is nil")
	}else{
		fmt.Print("we have a file")
	}

}

type DownloadDetails struct {
	bucketName string
	item string

}

func downloadFileFromBucket(dd *DownloadDetails, config *aws.Config) (*os.File, error) {
	file, err := os.Create(dd.item)
	if err != nil {
		log.Errorf("Error creating file with name: %v", dd.item)
		return nil, err
	}
	sess, _ := session.NewSession(config)
	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(dd.bucketName),
			Key:    aws.String(dd.item),
		})
	if err != nil {
		log.Errorf("Error downloading file from bucket: %v with name %v", dd.bucketName, dd.item)
		return nil, err
	}

	log.Debugf("Downloaded", file.Name(), numBytes, "bytes")
	return file, nil
}

func getLogLevel() log.Level {
	value := os.Getenv("LOG_LEVEL")
	if len(value) == 0 {
		return log.WarnLevel
	}
	level, err := log.ParseLevel(value)
	if err != nil {
		return log.WarnLevel
	}
	return level
}

func getDownloadDetails() *DownloadDetails{
	fmt.Print("What is the bucket name?: ")
	var bucket string
	fmt.Scanln(&bucket)

	fmt.Print("What is the name of the item?: ")
	var item string
	fmt.Scanln(&item)
	dd:= DownloadDetails{
		bucketName: bucket,
		item: item,
	}
	return &dd
}
