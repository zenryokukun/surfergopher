package main

/*
NewTwitter -- constructor
(*twitter)tweet -- tweets text
(*twitter)tweetImage --tweets text with Image
*/

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
)

type (
	keys struct {
		API_KEY       string `json:"API_KEY"`
		API_SECRET    string `json:"API_SECRET"`
		ACCESS_TOKEN  string `json:"ACCESS_TOKEN"`
		ACCESS_SECRET string `json:"ACCESS_SECRET"`
	}

	twitter struct {
		api *anaconda.TwitterApi
	}
)

func NewTwitter() *twitter {
	t := &twitter{}
	t.setAPIkeys()
	return t
}

func (t *twitter) setAPIkeys() {
	b, err := os.ReadFile("./twitter_conf.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	_keys := &keys{}
	json.Unmarshal(b, _keys)
	anaconda.SetConsumerKey(_keys.API_KEY)
	anaconda.SetConsumerSecret(_keys.API_SECRET)
	t.api = anaconda.NewTwitterApi(_keys.ACCESS_TOKEN, _keys.ACCESS_SECRET)
}

func (t *twitter) tweet(text string, v url.Values) {
	_, err := t.api.PostTweet(text, v)
	if err != nil {
		fmt.Println(err)
	}
}

func (t *twitter) tweetImage(text, imgPath string) {
	bs := imgBase64(imgPath)
	media, err := t.api.UploadMedia(bs)
	if err != nil {
		fmt.Println(err)
		return
	}
	v := url.Values{}
	v.Add("media_ids", media.MediaIDString)
	t.tweet(text, v)
}

func imgBase64(imgPath string) string {
	b, err := os.ReadFile(imgPath)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	bstr := base64.StdEncoding.EncodeToString(b)
	return bstr
}
