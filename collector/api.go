package collector

import (
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/prometheus/common/log"
)

type HitronRouter struct {
	URL      string
	Username string
	Password string
	client   *http.Client
}

func NewHitronRouter(url, username, password string) *HitronRouter {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalln("creating cookie jar:", err)
	}
	return &HitronRouter{
		URL:      url,
		Username: username,
		Password: password,
		client: &http.Client{
			Jar:     cookieJar,
			Timeout: time.Second * 30,
		},
	}
}
