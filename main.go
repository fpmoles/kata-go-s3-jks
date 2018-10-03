package main

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/fpmoles/kata-go-s3-jks/aws_client"
	"github.com/fpmoles/kata-go-s3-jks/certs"
	"github.com/fpmoles/kata-go-s3-jks/utils"
	"github.com/pavel-v-chernykh/keystore-go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(utils.GetLogLevel())
}

func main() {
	config := &aws.Config{
		Region: aws.String("us-east-2"),
	}
	dd := aws_client.GetDownloadDetails()
	file, err := aws_client.DownloadFileFromBucket(dd, config)
	if err != nil {
		log.Errorf("Error while downloading file")
		os.Exit(1)
	}
	if file == nil {
		log.Errorf("File is nil")
		os.Exit(11)
	}
	creds, err := readCredsFromZipFile(dd.Item, dd.BucketName)
	if err != nil {
		log.Errorf("Error while parsing creds file")
		os.Exit(2)
	}
	log.Debugf("Public key: %v", hex.EncodeToString(creds.cert))

	identityStore := certs.CreateIdentityStore("client", creds.pvtKey, creds.cert)
	trustStore := certs.CreateTrustStore("remote", creds.remoteCert)
	writeNewZip(dd, identityStore, trustStore)
	aws_client.UploadZipFile(dd, config)
	cleanup(dd)
}

type Creds struct {
	pvtKey     []byte
	cert       []byte
	remoteCert []byte
}

func readCredsFromZipFile(filename string, bucket string) (c Creds, err error) {
	z, err := zip.OpenReader(filename)
	if err != nil {
		log.Errorf("Error opening zipfile: %v", filename)
		return
	}
	defer z.Close()
	var r io.ReadCloser

	clientKey := utils.GetClientKeyName(bucket)
	clientCert := utils.GetClientCertName(bucket)
	remoteCert := utils.GetRemoteCertName(bucket)

	for _, f := range z.File {
		log.Debugf("Filename: %v", f.Name)

		switch f.Name {
		case clientCert:
			r, err = f.Open()
			if err != nil {
				log.Errorf("Error opening client cert with name %v", clientCert)
				return
			}
			byteBuf := new(bytes.Buffer)
			byteBuf.ReadFrom(r)
			c.cert = byteBuf.Bytes()
		case clientKey:
			r, err = f.Open()
			if err != nil {
				log.Errorf("Error opening client key with name %v", clientKey)
				return
			}
			byteBuf := new(bytes.Buffer)
			byteBuf.ReadFrom(r)
			c.pvtKey = byteBuf.Bytes()
		case remoteCert:
			r, err = f.Open()
			if err != nil {
				log.Errorf("Error opening remote cert with name %v", clientKey)
				return
			}
			byteBuf := new(bytes.Buffer)
			byteBuf.ReadFrom(r)
			c.remoteCert = byteBuf.Bytes()
		}
	}
	return
}

func writeNewZip(dd aws_client.DownloadDetails, identityStore keystore.KeyStore, trustStore keystore.KeyStore) {
	z, err := zip.OpenReader(dd.Item)
	if err != nil {
		log.Errorf("Error opening zipfile: %v", dd.Item)
		return
	}
	defer z.Close()

	if err := os.MkdirAll(dd.BucketName, 0755); err != nil {
		log.Errorf("Error creating directory: %v", dd.BucketName)
		return
	}

	for _, file := range z.File {
		path := file.Name
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			log.Errorf("Error opening file: %v", file.Name)
			return
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			log.Errorf("Error opening new file: %v", file.Name)
			return
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			log.Error("Error copying file")
			return
		}
	}
	identityPwd := utils.SecureRandomAlphaString(17)
	trustStorePwd := utils.SecureRandomAlphaString(17)

	certs.WriteKeyStore(identityStore, utils.GetIdentityStoreName(dd.BucketName), []byte(identityPwd))
	certs.WriteKeyStore(trustStore, utils.GetTrustStoreName(dd.BucketName), []byte(trustStorePwd))

	pwdText := "identitystore_pwd: " + identityPwd + "\nidentity_alias: client\ntruststore_pwd: " + trustStorePwd + "\ntruststore_alias: remote"
	err = ioutil.WriteFile(dd.BucketName+"/pwd.txt", []byte(pwdText), 0644)
	if err != nil {
		log.Errorf("Error creating pwd file")
		return
	}

	zipfile, err := os.Create(dd.Item)
	if err != nil {
		log.Errorf("Error creating new zip file")
		return
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(dd.BucketName)
	if err != nil {
		log.Errorf("Issue stating the directory")
		return
	}
	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(dd.BucketName)
	}
	filepath.Walk(dd.BucketName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, dd.BucketName))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return
}

func cleanup(dd aws_client.DownloadDetails) {
	os.Remove(dd.Item)
	os.RemoveAll(dd.BucketName)
}
