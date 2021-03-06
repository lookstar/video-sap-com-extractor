package collector

import (
	"fmt"
	"io/ioutil"
	"crypto/tls"
	"net/http"
	"gopkg.in/headzoo/surf.v1"
	"github.com/headzoo/surf/browser"
	"github.com/robertkrimen/otto"
	"os"
	"io"
	"runtime"
	"strings"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"encoding/json"
)

type CollectorProvider struct {
	videoURL string
}

type Credential struct {
	Username string	`json:"username"`
	Password string	`json:"password"`
}

func NewCollectorProvider(url string) *CollectorProvider {
	return &CollectorProvider{
		videoURL: url,
	}
}

func (p *CollectorProvider) processUrl(url string) string {
	// all sap video url looks like this: https://video.sap.com/media/t/1_q7ykdtqu#
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return strings.Trim(parts[len(parts)-1], "#")
	}
	return ""
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
	identification := p.processUrl(p.videoURL)

	redisHost := os.Getenv("REDIS_URL")
	redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	session, err := redis.Dial("tcp", redisHost, redis.DialDatabase(redisPort))
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return err
	}
	defer session.Close()

	isDone, err := session.Do("GET", identification)
	if err != nil {
		return err
	}
	if isDone != nil {
		fmt.Println(identification + " duplicated")
		return nil
	}

	bow := p.constructBrowser()
	err = p.handleMediaInitForm(bow)
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

	err = bow.Open(p.videoURL)
	if err != nil {
		panic(err)
	}

	videoTitle := bow.Title()

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
	ks := pdataMap["flashvars"].(map[string]interface{})["ks"].(string)

	urlPattern := "http://cdnapi.kaltura.com/p/1921661/sp/0/playManifest/entryId/%s/format/url/protocol/http/flavorParamId/301971/ks/%s/video.mp4"
	targeturl := fmt.Sprintf(urlPattern, entryId, ks)

	fmt.Println(targeturl)
	fmt.Println(videoTitle)

	if runtime.GOOS == "windows" {
		DownloadBody(&http.Client{Transport: p.getTr()}, targeturl, "c:\\download", identification + ".mp4")
		ioutil.WriteFile("c:\\download\\" + identification + ".txt", []byte(videoTitle), 0655)
	} else {
		DownloadBody(&http.Client{Transport: p.getTr()}, targeturl, "/hypercd/demo/video", identification + ".mp4")
		ioutil.WriteFile("/hypercd/demo/video/" + identification + ".txt", []byte(videoTitle), 0655)
	}

	_, err = session.Do("SET", identification, "done")

	return nil
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