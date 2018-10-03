package certs

import (
	"encoding/pem"
	"github.com/pavel-v-chernykh/keystore-go"
	"log"
	"os"
	"time"
)

func CreateIdentityStore(alias string, key []byte, cert []byte) keystore.KeyStore {
	privateKey, _ := pem.Decode(key)
	certificate := keystore.Certificate{
		Type:    "X509",
		Content: cert,
	}
	certificateChain := []keystore.Certificate{certificate}

	identityStore := keystore.KeyStore{
		alias: &keystore.PrivateKeyEntry{
			Entry: keystore.Entry{
				CreationDate: time.Now(),
			},
			PrivKey:   privateKey.Bytes,
			CertChain: certificateChain,
		},
	}
	return identityStore
}

func CreateTrustStore(alias string, remoteCert []byte) keystore.KeyStore {
	serverCertificate := keystore.Certificate{
		Type:    "X509",
		Content: remoteCert,
	}

	trustStore := keystore.KeyStore{
		alias: &keystore.TrustedCertificateEntry{
			Entry: keystore.Entry{
				CreationDate: time.Now(),
			},
			Certificate: serverCertificate,
		},
	}
	return trustStore
}

func WriteKeyStore(keyStore keystore.KeyStore, filename string, password []byte) {
	o, err := os.Create(filename)
	defer o.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = keystore.Encode(o, keyStore, password)
	if err != nil {
		log.Fatal(err)
	}
}
