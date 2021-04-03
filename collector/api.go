package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type HitronRouter struct {
	URL      string
	Username string
	Password string

	client    *http.Client
	parsedUrl *url.URL
}

type Info struct {
	HwVersion       string `json:"hwVersion"`       // 1A",
	SwVersion       string `json:"swVersion"`       // 4.12.34.567-XX-YYY",
	SerialNumber    string `json:"serialNumber"`    // VCAP12345678",
	RfMac           string `json:"rfMac"`           // 68:8F:12:34:12:34",
	WanIp           string `json:"wanIp"`           // 84.12.34.56/21",
	AftrName        string `json:"aftrName"`        // ",
	AftrAddr        string `json:"aftrAddr"`        // ",
	DelegatedPrefix string `json:"delegatedPrefix"` // ",
	LanIPv6Addr     string `json:"lanIPv6Addr"`     // ",
	SystemUptime    string `json:"systemUptime"`    // 04 Days,22 Hours,23 Minutes,48 Seconds",
	SystemTime      string `json:"systemTime"`      // Sat Apr 03, 2021, 14:16:41",
	Timezone        string `json:"timezone"`        // 1",
	WRecPkt         string `json:"WRecPkt"`         // 815.00M Bytes",
	WSendPkt        string `json:"WSendPkt"`        // 527.44M Bytes",
	LanIp           string `json:"lanIp"`           // 192.168.0.1/24",
	LRecPkt         string `json:"LRecPkt"`         // 779.79M Bytes",
	LSendPkt        string `json:"LSendPkt"`        // 1.15G Bytes"
}

func NewHitronRouter(rawUrl, username, password string) *HitronRouter {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalln("creating cookie jar:", err)
	}
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		log.Fatalln("invalid hostname:", err)
	}
	return &HitronRouter{
		URL:       rawUrl,
		parsedUrl: parsedUrl,
		Username:  username,
		Password:  password,
		client: &http.Client{
			Jar:     cookieJar,
			Timeout: time.Second * 30,
		},
	}
}

func (r *HitronRouter) Login() error {
	// login check to get preSession cookie
	resp, err := r.client.Get(r.URL + "/index.html")
	if err != nil {
		log.Infof("login: %+v err: %+v", resp, err)
		return err
	}
	log.Debug("Login Check:", resp)
	//if resp.StatusCode != 302 {
	// no need to login
	//return nil
	//}

	form := url.Values{
		"usr":         {r.Username},
		"pwd":         {r.Password},
		"forcelogoff": {"1"},
		"preSession":  {r.getCookie("preSession")},
	}
	resp, err = r.client.Post(r.URL+"/goform/login",
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()))

	if err != nil {
		log.Warnf("Login error: %+v / %+v", err, resp)
		return err
	}
	r.getCookie("sessionindex") // for debug
	return nil
}

func (r *HitronRouter) getCookie(name string) string {
	list := r.client.Jar.Cookies(r.parsedUrl)
	log.Debug("cookies:", list)
	for _, c := range list {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

func (r *HitronRouter) Info() (*Info, error) {
	resp, err := r.client.Get(r.URL + "/data/getSysInfo.asp")
	if err != nil {
		return nil, errors.Wrap(err, "getting info")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Info raw: %+v : %v", resp, string(data))
	var info []Info
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, errors.Wrap(err, "parsing info")
	}
	log.Debugf("Info: %+v", info)
	if len(info) != 1 {
		return nil, errors.New(fmt.Sprintf("Info gave wrong length: %d", len(info)))
	}
	return &info[0], nil
}
