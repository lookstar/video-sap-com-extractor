package provider

import (
	"fmt"
	"crypto/tls"
	"net/http"
	"gopkg.in/headzoo/surf.v1"
	"github.com/PuerkitoBio/goquery"
)

type CollectorProvider struct {
	Url        string
	Output     string
}

func NewCollectorProvider(url, output string) *CollectorProvider {
	return &CollectorProvider{
		Url:        url,
		Output:     output,
	}
}

func (p *CollectorProvider) DoWork() error {
	fmt.Println("hello world!\n")

	bow := surf.NewBrowser()

	clientCrt, _ := tls.LoadX509KeyPair("./data/test.crt", "./data/test.key")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{clientCrt},
			Renegotiation:      tls.RenegotiateFreelyAsClient,
			MaxVersion:         tls.VersionTLS10,
		},
	}

	bow.SetTransport(tr)

	err := bow.Open("https://accounts.sap.com/saml2/idp/usso/sap?sp=video.sap.com")
	if err != nil {
		panic(err)
	}

	bow.Find("form").Each(func(_ int, s *goquery.Selection){
		fmt.Println(s.Attr("id"))
	})

	return nil
}
