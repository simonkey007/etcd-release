package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Flags struct {
	EtcdAddress    string
	CACertFilePath string
	CertFilePath   string
	KeyFilePath    string
}

func main() {
	flags := Flags{}
	flag.StringVar(&flags.EtcdAddress, "etcdAddress", "", "path to the etcd machine")
	flag.StringVar(&flags.CACertFilePath, "caCert", "", "path to the etcd ca cert")
	flag.StringVar(&flags.CertFilePath, "clientCert", "", "path to the etcd client certificate")
	flag.StringVar(&flags.KeyFilePath, "clientKey", "", "path to the etcd client key")
	flag.Parse()

	proxyUrl, err := url.Parse(flags.EtcdAddress)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyUrl)
	proxy.Transport = &http.Transport{
		TLSClientConfig: buildTLSConfig(flags.CACertFilePath, flags.CertFilePath, flags.KeyFilePath),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%+v", r)
		proxy.ServeHTTP(w, r)
	})

	http.ListenAndServe(":4001", nil)
}

func buildTLSConfig(caCertFilePath, certFilePath, keyFilePath string) *tls.Config {
	tlsCert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		log.Fatal(err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: false,
		ClientAuth:         tls.RequireAndVerifyClientCert,
	}

	certBytes, err := ioutil.ReadFile(caCertFilePath)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
		log.Fatal(err)
	}

	tlsConfig.RootCAs = caCertPool
	tlsConfig.ClientCAs = caCertPool

	return tlsConfig
}
