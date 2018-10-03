package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func GetLogLevel() log.Level {
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

func GetClientKeyName(bucket string) string {
	value := os.Getenv("CLIENT_KEY_NAME")
	var clientKey string
	if len(value) == 0 {
		clientKey = bucket + "/key"
	} else {
		clientKey = bucket + "/" + value
	}
	log.Debugf("Client key name is: %v", clientKey)
	return clientKey
}

func GetClientCertName(bucket string) string {
	value := os.Getenv("CLIENT_CERT_NAME")
	var clientCert string
	if len(value) == 0 {
		clientCert = bucket + "/cert"
	} else {
		clientCert = bucket + "/" + value
	}
	log.Debugf("Client cert name is: %v", clientCert)
	return clientCert
}

func GetRemoteCertName(bucket string) string {
	value := os.Getenv("REMOTE_CERT_NAME")
	var remoteCert string
	if len(value) == 0 {
		remoteCert = bucket + "/ca.crt"
	} else {
		remoteCert = bucket + "/" + value
	}
	log.Debugf("Remote cert name is: %v", remoteCert)
	return remoteCert
}

func GetTrustStoreName(bucket string) string {
	value := os.Getenv("TRUST_STORE_NAME")
	var trustStoreName string
	if len(value) == 0 {
		trustStoreName = bucket + "/truststore.jks"
	} else {
		trustStoreName = bucket + "/" + value
	}
	log.Debugf("TrustStore name is: %v", trustStoreName)
	return trustStoreName
}

func GetIdentityStoreName(bucket string) string {
	value := os.Getenv("IDENTITY_STORE_NAME")
	var trustStoreName string
	if len(value) == 0 {
		trustStoreName = bucket + "/identity.jks"
	} else {
		trustStoreName = bucket + "/" + value
	}
	log.Debugf("Keystore name is: %v", trustStoreName)
	return trustStoreName
}
