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

type Session struct {
	*HitronRouter
}

type SysInfo struct {
	HwVersion       string `json:"hwVersion"`       // 1A
	SwVersion       string `json:"swVersion"`       // 4.12.34.567-XX-YYY
	SerialNumber    string `json:"serialNumber"`    // VCAP12345678
	RfMac           string `json:"rfMac"`           // 68:8F:12:34:12:34
	WanIp           string `json:"wanIp"`           // 84.12.34.56/21
	AftrName        string `json:"aftrName"`        //
	AftrAddr        string `json:"aftrAddr"`        //
	DelegatedPrefix string `json:"delegatedPrefix"` //
	LanIPv6Addr     string `json:"lanIPv6Addr"`     //
	SystemUptime    string `json:"systemUptime"`    // 04 Days,22 Hours,23 Minutes,48 Seconds
	SystemTime      string `json:"systemTime"`      // Sat Apr 03, 2021, 14:16:41
	Timezone        string `json:"timezone"`        // 1
	WRecPkt         string `json:"WRecPkt"`         // 815.00M Bytes
	WSendPkt        string `json:"WSendPkt"`        // 527.44M Bytes
	LanIp           string `json:"lanIp"`           // 192.168.0.1/24
	LRecPkt         string `json:"LRecPkt"`         // 779.79M Bytes
	LSendPkt        string `json:"LSendPkt"`        // 1.15G Bytes
}

type CMInit struct {
	HwInit         string `json:"hwInit"`         // Success
	FindDownstream string `json:"findDownstream"` // Success
	Ranging        string `json:"ranging"`        // Success
	Dhcp           string `json:"dhcp"`           // Success
	TimeOfday      string `json:"timeOfday"`      // Secret
	DownloadCfg    string `json:"downloadCfg"`    // Success
	Registration   string `json:"registration"`   // Success
	EaeStatus      string `json:"eaeStatus"`      // Secret
	BpiStatus      string `json:"bpiStatus"`      // AUTH:authorized, TEK:operational
	NetworkAccess  string `json:"networkAccess"`  // Permitted
	TrafficStatus  string `json:"trafficStatus"`  // Enabl
}

var (
	StatusSuccess          = "Success"
	NetworkAccessPermitted = "Permitted"

	contentType = "application/x-www-form-urlencoded"

	ErrorBackingOff  = errors.New("Backing off, because the router told us to do so")
	ErrorLoginAnswer = errors.New("Login response unknown")

	// Time to wait for a previous session to end before failing a scrape.
	WaitTimeout = time.Second * 3
	// Timeout for individual requests.
	RequestTimeout = time.Second * 30

	accessToken chan bool
)

func init() {
	accessToken = make(chan bool, 1)
	accessToken <- true
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
			Timeout: RequestTimeout,
		},
	}
}

func (r *HitronRouter) Login() (*Session, error) {
	// get a backoff token
	session := &Session{r}
	if t := session.getToken(); !t {
		return nil, ErrorBackingOff
	}

	// login check to get preSession cookie
	resp, err := r.client.Get(r.URL + "/index.html")
	if err != nil {
		log.Infof("login: %+v err: %+v", resp, err)
		session.abort()
		return nil, err
	}
	defer resp.Body.Close()
	log.Debug("Login Check:", resp)
	//if resp.StatusCode != 302 {
	// no need to login
	//return nil
	//}

	form := url.Values{
		"usr": {r.Username},
		"pwd": {r.Password},
		//"forcelogoff": {"1"},
		"preSession": {r.getCookie("preSession")},
	}
	resp, err = r.client.Post(r.URL+"/goform/login",
		contentType,
		strings.NewReader(form.Encode()))
	if err != nil {
		log.Warnf("Login error: %+v / %+v", err, resp)
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	response := string(data)
	log.Debugf("Login response: %+v %+v Body: %s", err, resp, response)
	log.Debug("cookies: ", r.client.Jar.Cookies(r.parsedUrl))
	if response != "success" {
		return nil, r.handleLoginError(session, response)
	}

	return session, nil
}

func (r *HitronRouter) handleLoginError(session *Session, response string) error {
	if strings.Contains(response, "LoginProtect=") {
		//TODO backoff login!!
		// LoginProtect=9|58|21
		// --> 9 failed attempts, wait 58min21s
		var failedAttempts, minutes, seconds time.Duration
		if _, err := fmt.Sscanf(response, "LoginProtect=%d|%d|%d", &failedAttempts, &minutes, &seconds); err != nil {
			session.abort()
			return errors.New("parsing backoff '" + response + "': " + err.Error())
		}
		go func() {
			wait := time.After(time.Minute*minutes + time.Second*seconds)
			<-wait
			session.abort()
		}()
		return ErrorBackingOff
	}

	return errors.Wrap(ErrorLoginAnswer, response)
}

func (r *Session) getToken() bool {
	timeout := time.After(WaitTimeout)
	select {
	case <-accessToken:
		return true
	case <-timeout:
		return false
	}
}

func (r *Session) abort() {
	accessToken <- true
}

func (r *Session) Logout() {
	defer r.abort()
	form := url.Values{
		"data": {"byebye"},
	}
	resp, err := r.client.Post(r.URL+"/goform/logout", contentType, strings.NewReader(form.Encode()))
	if err != nil {
		log.Warnf("Logout error: %+v: %+v", err, resp)
		return
	}
	// body should be empty
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

func (r *HitronRouter) fetch(name string, output interface{}) error {
	resp, err := r.client.Get(r.URL + "/data/get" + name + ".asp")
	if err != nil {
		return errors.Wrap(err, "getting "+name)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debugf("%s raw: %+v : %v", name, resp, string(data))
	if strings.Contains(string(data), "Unknown error.") {
		//TODO
	}
	err = json.Unmarshal(data, output)
	if err != nil {
		return errors.Wrap(err, "parsing "+name)
	}
	log.Debugf("%s: %+v", name, output)
	return nil
}

func (r *HitronRouter) Info() (*SysInfo, error) {
	var data []SysInfo
	err := r.fetch("SysInfo", &data)
	if err != nil {
		return nil, err
	}
	if len(data) != 1 {
		return nil, errors.New(fmt.Sprintf("SysInfo gave wrong length: %d", len(data)))
	}
	return &data[0], err
}

func (r *HitronRouter) CMInit() (*CMInit, error) {
	var data []CMInit
	err := r.fetch("CMInit", &data)
	if err != nil {
		return nil, err
	}
	if len(data) != 1 {
		return nil, errors.New(fmt.Sprintf("CMInit gave wrong length: %d", len(data)))
	}
	return &data[0], err
}
