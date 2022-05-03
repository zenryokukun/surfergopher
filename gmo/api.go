package gmo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	Public  = "https://api.coin.z.com/public"
	Private = "https://api.coin.z.com/private"
)

func pubURI(dir string, param Bmap) string {
	query := Querystr(param)
	return Public + dir + query
}

func priURI(dir string, param Bmap) string {
	query := Querystr(param)
	return Private + dir + query
}

type Bmap map[string]string
type Imap map[string]interface{}

type ReqHandler struct {
	Client *http.Client
	Auth   *authKeys
}

func (request *ReqHandler) Get(dir string, param Bmap, i GMOAPI) error {
	uri := pubURI(dir, param)

	res, err := request.Client.Get(uri)

	if err != nil {
		fmt.Println(err)
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = json.Unmarshal(body, i)
	if err != nil {
		fmt.Println(err)
		return err
	}

	i.ErrorLog()
	return nil
}

func (request *ReqHandler) GetAuth(dir string, param Bmap, i GMOAPI) error {
	api, secret := request.Auth.Keys()
	uri := priURI(dir, param)
	sign, nonce := getSign("GET", dir, "", secret)
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		fmt.Println(err)
		return err
	}

	setHeaderAuth(nonce, sign, api, req)

	res, err := request.Client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = json.Unmarshal(body, i)
	if err != nil {
		fmt.Println(err)
		return err
	}

	i.ErrorLog()
	return nil
}

func (request *ReqHandler) Post(dir string, param Imap, i GMOAPI) error {
	api, secret := request.Auth.Keys()
	uri := priURI(dir, nil)

	reqbody, err := json.Marshal(param)
	if err != nil {
		fmt.Println(err)
		return err
	}

	sign, nonce := getSign("POST", dir, string(reqbody), secret)

	req, err := http.NewRequest("POST", uri, strings.NewReader(string(reqbody)))
	if err != nil {
		fmt.Println(err)
		return err
	}
	setHeaderAuth(nonce, sign, api, req)
	res, err := request.Client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = json.Unmarshal(body, i)

	if err != nil {
		fmt.Println(err)
		return err
	}

	i.ErrorLog()
	return nil
}

func getNonce() string {
	var nowNanoSec int64 = time.Now().UnixNano()
	var milliSec int64 = int64(time.Millisecond)
	var _nonce int64 = nowNanoSec / milliSec
	return fmt.Sprint(_nonce)
}

func getSign(method string, dir string, body string, secret string) (string, string) {
	var nonce string = getNonce()
	var message string = nonce + method + dir + body
	hc := hmac.New(sha256.New, []byte(secret))
	hc.Write([]byte(message))
	var sign string = hex.EncodeToString(hc.Sum(nil))
	return sign, nonce
}

func setHeaderAuth(nonce string, sign string, apikey string, req *http.Request) {
	req.Header.Set("API-KEY", apikey)
	req.Header.Set("API-TIMESTAMP", nonce)
	req.Header.Set("API-SIGN", sign)
}

func InitGMO(path string) *ReqHandler {
	return &ReqHandler{
		Client: &http.Client{},
		Auth:   NewAuthKeys(path),
	}
}
