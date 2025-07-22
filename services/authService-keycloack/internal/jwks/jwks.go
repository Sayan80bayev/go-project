package jwks

import (
	"auth_service/internal/config"
	"crypto/tls"
	"crypto/x509"
	"github.com/MicahParks/keyfunc/v2"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

var JWKS *keyfunc.JWKS

func Init(cfg *config.Config, log *logrus.Logger) {
	jwksURL := cfg.KeycloakURL + "/realms/" + cfg.KeycloakRealm + "/protocol/openid-connect/certs"

	certData, err := ioutil.ReadFile(cfg.KeycloakCertPath)
	if err != nil {
		log.Fatalf("Не удалось прочитать сертификат: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certData) {
		log.Fatal("Не удалось добавить сертификат в пул доверенных")
	}

	customClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		},
	}

	JWKS, err = keyfunc.Get(jwksURL, keyfunc.Options{
		Client:          customClient,
		RefreshInterval: time.Hour,
		RefreshErrorHandler: func(err error) {
			log.Errorf("Ошибка обновления JWKS: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("Ошибка получения JWKS: %v", err)
	}
}
