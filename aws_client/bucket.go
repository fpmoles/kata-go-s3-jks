package aws_client

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
	"os"
)

type DownloadDetails struct {
	BucketName string
	Item       string
}

func DownloadFileFromBucket(dd DownloadDetails, config *aws.Config) (*os.File, error) {
	file, err := os.Create(dd.Item)
	if err != nil {
		log.Errorf("Error creating file with name: %v", dd.Item)
		return nil, err
	}
	sess, _ := session.NewSession(config)
	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(dd.BucketName),
			Key:    aws.String(dd.Item),
		})
	if err != nil {
		log.Errorf("Error downloading file from bucket: %v with name %v", dd.BucketName, dd.Item)
		return nil, err
	}

	log.Debugf("Downloaded", file.Name(), numBytes, "bytes")
	return file, nil
}

func UploadZipFile(dd DownloadDetails, config *aws.Config) {
	sess, _ := session.NewSession(config)

	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(dd.Item)
	if err != nil {
		log.Errorf("Error opening file handle for upload")
	}

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(dd.BucketName),
		Key:    aws.String(dd.Item),
		Body:   f,
	})
	if err != nil {
		log.Errorf("failed to upload file, %v", err)
		return
	}
	log.Debugf("Result of upload operation: %v", result.Location)
}

func GetDownloadDetails() DownloadDetails {
	fmt.Print("What is the bucket name?: ")
	var bucket string
	fmt.Scanln(&bucket)

	fmt.Print("What is the name of the Item?: ")
	var item string
	fmt.Scanln(&item)
	dd := DownloadDetails{
		BucketName: bucket,
		Item:       item,
	}
	return dd
}
