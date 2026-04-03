package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func GetHttpResponse(url string) (body []byte, err error) {
	var (
		req *http.Request
		res *http.Response
	)
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	res, err = httpClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	return
}

func BuildTLSConfig(cert, clientKey, clientCert string) (tlsConfig *tls.Config, err error) {
	if strings.HasPrefix(cert, "LS0t") {
		deBytes, _ := base64.StdEncoding.DecodeString(cert)
		cert = string(deBytes)
	}
	if strings.HasPrefix(clientKey, "LS0t") {
		deBytes, _ := base64.StdEncoding.DecodeString(clientKey)
		clientKey = string(deBytes)
	}
	if strings.HasPrefix(clientCert, "LS0t") {
		deBytes, _ := base64.StdEncoding.DecodeString(clientCert)
		clientCert = string(deBytes)
	}
	var client tls.Certificate
	client, err = tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
	if err != nil {
		return
	}
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{client},
	}
	if len(cert) > 0 {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(cert))
		tlsConfig.RootCAs = caCertPool
	}
	return
}
func GetRemoteAddress(address string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.Split(address, ":")[0], "]"), "[")
}

func GetBrowserDetail(useAgent string) (browser, platformDetail string) {
	re := regexp.MustCompile(`\((.*?)\)`)
	match := re.FindStringSubmatch(useAgent)
	if len(match) > 1 {
		platformDetail = match[1]
	}
	if (strings.Index(useAgent, "Firefox")) > -1 {
		browser = "Firefox"
	} else if (strings.Index(useAgent, "QQBrowser")) > -1 {
		browser = "QQBrowser"
	} else if (strings.Index(useAgent, "QQ")) > -1 {
		browser = "QQ"
	} else if (strings.Index(useAgent, "UCBrowser")) > -1 {
		browser = "UCBrowser"
	} else if (strings.Index(useAgent, "Opera")) > -1 || (strings.Index(useAgent, "OPR")) > -1 {
		browser = "Opera"
	} else if (strings.Index(useAgent, "Wechat")) > -1 {
		browser = "Wechat"
	} else if (strings.Index(useAgent, "Trident")) > -1 {
		browser = "InternetExplorer"
	} else if (strings.Index(useAgent, "Edge")) > -1 {
		browser = "Edge"
	} else if (strings.Index(useAgent, "Chrome")) > -1 {
		browser = "Chrome"
	} else if (strings.Index(useAgent, "Safari")) > -1 {
		browser = "Safari"
	} else {
		browser = "Unknown"
	}
	return

}
