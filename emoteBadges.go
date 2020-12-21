package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

const ClientID = "joj3f4vzfu20ush2zsackrnbm8n9nd"

// type BTTVStruct []struct {
// 	Emote struct {
// 		ID   string `json:"id"`
// 		Code string `json:"code"`
// 	} `json:"emote"`
// }

// func (s Store) GetBTTVEmote(c *cache.Cache) *cache.Cache {
// 	body := GetRequest("https://api.betterttv.net/3/emotes/shared/top?offset=0&limit=50", map[string]string{}) // BTTV

// 	var data BTTVStruct
// 	json.Unmarshal(body, &data)
// 	for _, e := range data {
// 		c.Set(e.Emote.Code, "<img className='emoticon' title='BTTV: "+e.Emote.Code+"' alt='"+e.Emote.Code+"' src='//cdn.betterttv.net/emote/"+e.Emote.ID+"/1x' srcSet='//cdn.betterttv.net/emote/"+e.Emote.ID+"/2x 2x, //cdn.betterttv.net/emote/"+e.Emote.ID+"/3x 3x'/>", cache.NoExpiration)
// 	}
// 	return c
// }

type TwitchBages struct {
	BadgeSets struct {
		One979Revolution1 struct {
			Versions struct {
				Num1 struct {
					ImageURL1X  string      `json:"image_url_1x"`
					ImageURL2X  string      `json:"image_url_2x"`
					ImageURL4X  string      `json:"image_url_4x"`
					Description string      `json:"description"`
					Title       string      `json:"title"`
					ClickAction string      `json:"click_action"`
					ClickURL    string      `json:"click_url"`
					LastUpdated interface{} `json:"last_updated"`
				} `json:"1"`
			} `json:"versions"`
		} `json:"1979-revolution_1"`
	} `json:"badge_sets"`
}

type Version struct {
	ImageURL1X  string      `json:"image_url_1x"`
	ImageURL2X  string      `json:"image_url_2x"`
	ImageURL4X  string      `json:"image_url_4x"`
	Description string      `json:"description"`
	Title       string      `json:"title"`
	ClickAction string      `json:"click_action"`
	ClickURL    string      `json:"click_url"`
	LastUpdated interface{} `json:"last_updated"`
}

func (s *Store) getTwitchTVBadges(ChannelName, RoomID string) { // func getTwitchTVBadges(c *cache.Cache) *cache.Cache
	headers := map[string]string{
		"Accept":    "application/vnd.twitchtv.v5+json",
		"Client-ID": ClientID,
	}

	var body []byte
	if RoomID == "global" {
		body = GetRequest("https://badges.twitch.tv/v1/badges/global/display", headers) // twitch.tv Global bages
	} else {
		body = GetRequest("https://badges.twitch.tv/v1/badges/channels/"+RoomID+"/display", headers) // twitch.tv Channel bages
	}

	var jsonData map[string]interface{}
	json.Unmarshal(body, &jsonData)
	badgeSets := jsonData["badge_sets"].(map[string]interface{})

	s.mutex.Lock()
	s.Badges[ChannelName] = make(map[string]map[string]string)
	s.mutex.Unlock()

	for k, v := range badgeSets {
		mv, ok := v.(map[string]interface{})
		if ok {
			s.Badges[ChannelName][k] = make(map[string]string)
			for _, v2 := range mv {
				mv2, ok2 := v2.(map[string]interface{})
				if ok2 {
					for version, v3 := range mv2 {
						versionData, ok3 := v3.(map[string]interface{})
						if ok3 {
							description, _ := versionData["description"]
							title, _ := versionData["title"]
							image_url_1x, _ := versionData["image_url_1x"]
							image_url_2x, _ := versionData["image_url_2x"]
							image_url_4x, _ := versionData["image_url_4x"]
							s.mutex.Lock()
							s.Badges[ChannelName][k][version] = "<img title='" + title.(string) + "' alt='" + description.(string) + "' aria-label='" + title.(string) + "' class='line badge' src='" + image_url_1x.(string) + "' srcset='" + image_url_1x.(string) + " 1x, " + image_url_2x.(string) + " 2x, " + image_url_4x.(string) + " 4x'>"
							s.mutex.Unlock()
						} else {
							log.Printf("%v : %+v\n", version, v3)
						}
					}
				} else {
					log.Printf("%v : %v\n", k, v)
				}
			}
		} else {
			log.Printf("%v: %v\n", k, v)
		}
	}
}

func GetRequest(url string, headers map[string]string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	for key, prop := range headers {
		req.Header.Set(key, prop)
	}

	client := &http.Client{Timeout: time.Second * 15}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response. ", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body. ", err)
	}
	resp.Body.Close()
	return body
}
