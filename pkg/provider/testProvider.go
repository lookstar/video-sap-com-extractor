package provider

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"crypto/tls"
	"net/http"
	"gopkg.in/headzoo/surf.v1"
	"github.com/headzoo/surf/browser"
	"github.com/robertkrimen/otto"
	"os"
	"io"
	"runtime"
	"github.com/streadway/amqp"
)

type CollectorProvider struct {
}

type Credential struct {
	Username string	`json:"username"`
	Password string	`json:"password"`
}

func NewCollectorProvider() *CollectorProvider {
	return &CollectorProvider{
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
	err := bow.Open("https://accounts.sap.com/saml2/idp/usso/sap?sp=video.sap.com")
	if err != nil {
		panic(err)
	}

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

func (p *CollectorProvider) getTr() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Renegotiation:      tls.RenegotiateFreelyAsClient,
			MaxVersion:         tls.VersionTLS10,
		},
	}
}

func (p *CollectorProvider) constructBrowser() *browser.Browser {
	bow := surf.NewBrowser()
	bow.SetTransport(p.getTr())
	return bow
}

func (p *CollectorProvider) DoWork() error {
	fmt.Println("hello world!\n")

	bow := p.constructBrowser()
	err := p.handleMediaInitForm(bow)
	if err != nil {
		panic(err)
	}

	for bow.Title() == "Error" {
		bow = p.constructBrowser()
		err = p.handleMediaInitForm(bow)
		if err != nil {
			panic(err)
		}
		fmt.Println(bow.Title())
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
	
	//part 2
	//err = bow.Open("https://video.sap.com/media/t/1_4i0i6naj")
	err = bow.Open("https://video.sap.com/media/t/1_1yr8abip")
	if err != nil {
		panic(err)
	}

	selection := bow.Find("#player")
	lastScript := selection.Find("script").Last()
	vm := otto.New()

	_, err = vm.Run("var kms_kWidgetJsLoader = {};")
	_, err = vm.Run("kms_kWidgetJsLoader.embed = function(id, kms) {};")

	_, err = vm.Run(lastScript.Text())
	if err != nil {
		panic(err)
	}

	pdata, err := vm.Get("kMainPlayerEmbedObject")
	if err != nil {
		panic(err)
	}

	pdataInterface, _ := pdata.Export()
	pdataMap := pdataInterface.(map[string]interface{})

	entryId := pdataMap["entry_id"]
	//fmt.Println(pdataMap["wid"])
	ks := pdataMap["flashvars"].(map[string]interface{})["ks"].(string)

	//part 3
	urlPattern := "http://cdnapi.kaltura.com/p/1921661/sp/0/playManifest/entryId/%s/format/url/protocol/http/flavorParamId/301971/ks/%s/video.mp4"
	targeturl := fmt.Sprintf(urlPattern, entryId, ks)

	fmt.Println(targeturl)

	DownloadBody(&http.Client{Transport: p.getTr()}, targeturl, "c:\\download", "b.mp4")

	return nil
}

func connectMQ() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	defer ch.Close()
	

}

func DownloadBody(client *http.Client, url, dir, filename string) error {
	os.MkdirAll(dir, os.ModePerm)

	var path string
	if runtime.GOOS == "windows" {
		path = dir + "\\" + filename
	} else {
		path = dir + "/" + filename
	}

	file, err := os.Open(path)
	if err == nil {
		stat, _ := file.Stat()
		if stat.Size() > 10 {
			defer file.Close()
			return nil
		}
	} else {
		//file not exist
		if os.IsNotExist(err) {
			file.Close()
		} else {
			return err
		}
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}

	defer out.Close()

	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}