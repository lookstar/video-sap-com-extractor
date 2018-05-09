package provider

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"crypto/tls"
	"net/http"
	"gopkg.in/headzoo/surf.v1"
	"github.com/headzoo/surf/browser"
	//"github.com/PuerkitoBio/goquery"
)

type CollectorProvider struct {
	Url        string
	Output     string
}

type Credential struct {
	Username string
	Password string
}

func NewCollectorProvider(url, output string) *CollectorProvider {
	return &CollectorProvider{
		Url:        url,
		Output:     output,
	}
}

func (p *CollectorProvider) ReadCredential() *Credential {
	content, err := ioutil.ReadFile("./data/credential.json")
	if err != nil {
		fmt.Println("ReadCredential " + err.Error())
		panic(err)
	}
	ret := &Credential{}
	json.Unmarshal(content, ret)
	return ret
}

func (p *CollectorProvider) handleMediaInitForm(bow *browser.Browser) error {
	form, err := bow.Form("#logOnForm")
	if err != nil {
		return err
	}

	cred := p.ReadCredential()
	form.Input("j_username", cred.Username)
	form.Input("j_password", cred.Password)
	err = form.Submit()
	return err
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

	err = p.handleMediaInitForm(bow)
	if err != nil {
		panic(err)
	}

	if bow.Title() == "Error" {
		bow.Back()
		err = p.handleMediaInitForm(bow)
		if err != nil {
			panic(err)
		}
	}

	form, err := bow.Form("#samlRedirect")
	if err != nil {
		panic(err)
	}

	err = form.Submit()
	if err != nil {
		panic(err)
	}

	fmt.Println(bow.Title())
	
	err = bow.Open("https://video.sap.com/media/t/1_q7ykdtqu#")
	if err != nil {
		panic(err)
	}

	fmt.Println(bow.Title())

	return nil
}

