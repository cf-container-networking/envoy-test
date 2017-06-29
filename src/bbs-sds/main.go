package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type Host struct {
	IPAddress string `json:"ip_address"`
	Port      int    `json:"port"`
}

type BBSClient struct {
	Logger    lager.Logger
	BBSClient bbs.ExternalActualLRPClient
}

func (c *BBSClient) getHostsFromBBS(targetAppGUID string) []Host {
	logger := c.Logger.Session("getHosts")
	logger.Info("start", lager.Data{"appGUID": targetAppGUID})
	defer logger.Info("done")

	hosts := []Host{}
	lrpGroups, err := c.BBSClient.ActualLRPGroups(logger, models.ActualLRPFilter{})
	if err != nil {
		logger.Error("actualLRPGroups", err)
		return nil
	}
	for _, lrpGroup := range lrpGroups {
		if lrpGroup.Instance == nil {
			logger.Info("nil lrp instance")
			continue
		}
		processGUID := lrpGroup.Instance.ActualLRPKey.ProcessGuid
		lrpAppGUID := processGUID[:36]

		if lrpAppGUID != targetAppGUID {
			logger.Info("non-match", lager.Data{"actualLRPProcessGUID": processGUID})
			continue
		}
		host := Host{
			IPAddress: lrpGroup.Instance.ActualLRPNetInfo.InstanceAddress,
			Port:      8443,
		}
		logger.Info("match", lager.Data{"host": host})
		hosts = append(hosts, host)
	}

	return hosts
}

func failOnError(err error) {
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

const (
	clientSessionCacheSize int = 0
	maxIdleConnsPerHost    int = 0
)

func main() {
	bbsCertFile := "tmp/bbs_client.crt"
	bbsKeyFile := "tmp/bbs_client.key"

	os.MkdirAll("tmp", os.ModePerm)

	failOnError(ioutil.WriteFile(bbsCertFile, []byte(bbsClientCert), 0600))
	failOnError(ioutil.WriteFile(bbsKeyFile, []byte(bbsClientKey), 0600))

	bbsHostname := os.Getenv("BBS_HOSTNAME")
	bbsURL := fmt.Sprintf("https://%s:8889", bbsHostname)
	bbsClient, err := bbs.NewSecureSkipVerifyClient(
		bbsURL,
		bbsCertFile,
		bbsKeyFile,
		clientSessionCacheSize,
		maxIdleConnsPerHost,
	)
	if err != nil {
		log.Fatalf("new-secure-client: %s", err)
	}
	logger := lager.NewLogger("sds")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	client := &BBSClient{
		Logger:    logger,
		BBSClient: bbsClient,
	}

	http.HandleFunc("/v1/registration/", func(responseWriter http.ResponseWriter, request *http.Request) {
		parts := strings.Split(request.URL.Path, "/")
		appName := parts[len(parts)-1]

		var response struct {
			Hosts []Host `json:"hosts"`
		}
		response.Hosts = client.getHostsFromBBS(appName)
		err := json.NewEncoder(responseWriter).Encode(response)
		if err != nil {
			log.Printf("writing response: %s\n", err)
		}
	})
	http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		err := json.NewEncoder(responseWriter).Encode(map[string]string{"hello": "there"})
		if err != nil {
			log.Printf("writing response: %s\n", err)
		}
	})

	port := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

const bbsClientCert = `-----BEGIN CERTIFICATE-----
MIIDNjCCAh6gAwIBAgIRALr4Tu4PzPuXYJZSarz25kAwDQYJKoZIhvcNAQELBQAw
OzEMMAoGA1UEBhMDVVNBMRYwFAYDVQQKEw1DbG91ZCBGb3VuZHJ5MRMwEQYDVQQD
EwppbnRlcm5hbENBMB4XDTE3MDYyNjIyNTYwM1oXDTE4MDYyNjIyNTYwM1owOzEM
MAoGA1UEBhMDVVNBMRYwFAYDVQQKEw1DbG91ZCBGb3VuZHJ5MRMwEQYDVQQDEwpi
YnMgY2xpZW50MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9OxALoFZ
utIQqox2Kkt5GOAovxGYFlAX7mpnz0TPw8q14v+02CS0JPuK6tSKDNsYlwC0gn4c
hdXwCN+ijHpYLvw0FA+ffkcClbMh3+3xnzr5dnNkgXhbNRQpeIcYD/wtJIVDAIG2
5+WqqVKhKKQ4Aioyag5YdFB+8L0Z2wCG7lIKyxCpVjoITWv1a8B/KFhBL87IV313
sxMtU/SO7oo5qtr8R63tEvK0fw7tl2gxPwuOHseHow8YrX9DfvfDNRlB5PxAHrIs
aF8lgdQCbM6e4blEYzNqQoEoW9SJ8d5d/40gTwOWi/WW+KNvZ6VyBzykD84jrLhc
v0SLQEqph7rSgQIDAQABozUwMzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYI
KwYBBQUHAwIwDAYDVR0TAQH/BAIwADANBgkqhkiG9w0BAQsFAAOCAQEANQQhYwq/
6o9TqHNNIIX6xkfW8f3KxRtZ0waQfRb7DcFSQlStvw6hVLQO2TTVTy216AjCcfXC
cAd6ZqWaoI1TElhA/81pbjeOtUvm5jnr4Gl7iHSm+0ZkDSnSGsu+e1wwKmRK5BHM
20A55yAjjC+4do/4S8fIOH+czYVvlnogdtYv5rFXYEo6EYDGph52NfygGGJC+pw9
evZcQdaNRFzIGBosY6CyG0zOZiNOUbVK9hXDeh2JU2ULCXu1fTOdayGtHVTEzMAa
c384mJ5wKPTb6uKUYPm6U1bX7JqIBzVyDPj4Xj3SE+TcUrmcxyHXaQncQtyhSU3E
8aubV31f3sJbSA==
-----END CERTIFICATE-----`

const bbsClientKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA9OxALoFZutIQqox2Kkt5GOAovxGYFlAX7mpnz0TPw8q14v+0
2CS0JPuK6tSKDNsYlwC0gn4chdXwCN+ijHpYLvw0FA+ffkcClbMh3+3xnzr5dnNk
gXhbNRQpeIcYD/wtJIVDAIG25+WqqVKhKKQ4Aioyag5YdFB+8L0Z2wCG7lIKyxCp
VjoITWv1a8B/KFhBL87IV313sxMtU/SO7oo5qtr8R63tEvK0fw7tl2gxPwuOHseH
ow8YrX9DfvfDNRlB5PxAHrIsaF8lgdQCbM6e4blEYzNqQoEoW9SJ8d5d/40gTwOW
i/WW+KNvZ6VyBzykD84jrLhcv0SLQEqph7rSgQIDAQABAoIBAQCxLOps+fOgOvAF
gCDHDdvnS9kOBzs/AOee9+hqvvuRRmX3dUUsiriqfDENGX1YOXJ7Ye4y6+nUQ2Ql
9ylOd/6s1pMR5A9buSC8jF438Jg2uOHXdzhAlIFeT1yErS2R+rnpTmGezzcyYCjp
3jVpAgrmPgJESGZilgyOOC2pCKOTZ+m/55q5KvRVz4e2FPt5eaY+dVsKsNKyXHzQ
gFcvNDsY+6OzuJLDl5IeDRSICYRT9a7vtkYYanOzYkoxbrgusUSsuAMkFQkpeu3A
RQraTLdzryhhmu57oFO8CwAZwJQMXBc1vNzXS6eDWhimI0p//O6L5kM6vLRXoMot
LprekEUxAoGBAP9AOJ9BB/ozv4ukGFmuPa2UBrblE1W1uvbHsc8AdM3Vf9Ex7icz
tgQMxfouLQz3aWoQZDFU8oDjuZY4bk0zyxMX3TJPvt6wqByDFdSENBAL37dFhFqp
/S+g8iYEJfG3ws1VnQmsqFLvC4XkX4Ys04uKR7MX1S48Rg1Y5o/jzV/lAoGBAPWk
RQ2M+z/mEBBktlLEpy24AM0k9dcWdqrHb7+ERaERry+PONFr3Nw4KmgDnJAysVus
FUzMN1JG9izHDZi6ptiKCoIloJlAvuVh2My+D5btcZf0ZSczZOYrOT2BPABdikCP
z+iOEHQ8ePzFzfuEO4S3+ghyMwileKMC9rMWCSZtAoGAXNnfvw7I+Bsa8pEeyoC3
rwzJ5H4wKl2RRXQfGk3wL3Aart6a42fMLmz3F6r0eGMH1a1gxRFBpeExAZRFi4/r
r2Ze8I5RwHBCtxx4NHZi+fNXzjNbkh+EGm9RpsKbivJtyoP6PCqykHikmHAaz5Q+
3+PNcTiaM9d5JCHSvUUA0IECgYEAqFM50nBSV1Yimek5mvwRB144hlsWb56AEMT7
iYRtZlNE9dUx/Sfpv6ppPL+E0Lc8G/KO4gJqwmHIHaUFZyw4Wtg1HTwFkh7w8SSc
uKhg7G6nUZZynms0cBkcb04YvLNcoqMpuFVpZw1tZuFxJjJVyrt7hcAjwoAJa8MD
JHMsL/0CgYARtd62CBfxtdTNclfVzWd9eUNJsA9EFvXTIjawWuctSKZcjSyERJLJ
IATXGosVk3ZD+b2lHiwl4+aKyYfKrmgwsHnCJRvTfdpoz32EAuSzvHEsCediYTdm
ExcJG7eBVx+CGgjkIYp5gZ4sFUNpo5xym4DgOrjyJkKEWfeVhZ7UyA==
-----END RSA PRIVATE KEY-----`
