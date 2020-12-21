package main

//go:generate cp $GOROOT/misc/wasm/wasm_exec.js .

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/spddl/go-twitch-ws"
)

var window = js.Global().Get("window")

type Store struct {
	Badges map[string]map[string]map[string]string
	mutex  sync.RWMutex
}

func main() {
	js.Global().Set("StartTwitchChat", StartTwitchChat())
	select {}
}

func startBot(Channel []string) {
	store := new(Store)
	store.Badges = make(map[string]map[string]map[string]string)
	store.Badges["global"] = make(map[string]map[string]string)

	go store.getTwitchTVBadges("global", "global")

	bot, err := twitch.NewClient(&twitch.Client{
		Server:  "wss://irc-ws.chat.twitch.tv",
		Channel: Channel,
		Debug:   true,
	})
	if err != nil {
		panic(err)
	}

	bot.OnRoomStateMessage = func(ircMsg twitch.IRCMessage) { // Load new Badges
		go store.getTwitchTVBadges(string(ircMsg.Params[0][1:]), string(ircMsg.Tags["room-id"]))
	}

	bot.OnPongLatency = func(message time.Duration) {
		fmt.Printf("OnPongLatency: %s\n", message)
	}

	bot.OnNoticeMessage = func(ircMsg twitch.IRCMessage) {
		window.Call("OnNoticeMessage", map[string]interface{}{
			"channel": string(ircMsg.Params[0][1:]),
			"msgid":   string(ircMsg.Tags["msg-id"]),
			"message": string(ircMsg.Params[1]),
		})
	}

	bot.OnPrivateMessage = func(ircMsg twitch.IRCMessage) {
		tags := TagsToStrings(ircMsg.Tags)
		channel := string(ircMsg.Params[0][1:])
		msgByte := ircMsg.Params[1]

		// ignore "ACTION" Type
		if bytes.HasPrefix(msgByte, []byte{1, 65, 67, 84, 73, 79, 78, 32}) {
			b := bytes.TrimPrefix(msgByte, []byte{1, 65, 67, 84, 73, 79, 78, 32})
			msgByte = bytes.TrimSuffix(b, []byte{1})
		}

		msg := string(msgByte)

		var htmlDom string
		var styleColor string

		// Channel
		htmlDom += "<span id='" + channel + "' class='line'><span class='channel'>" + channel + "</span>"

		// Badges
		if tags["badges"] != "" {
			badgesArray := strings.Split(tags["badges"].(string), ",")
			for _, value := range badgesArray {
				badge := strings.Split(value, "/")

				data, okChan := store.Badges[channel][badge[0]][badge[1]]
				if okChan {
					htmlDom += data
					continue
				}
				data, okGlob := store.Badges["global"][badge[0]][badge[1]]
				if okGlob {
					htmlDom += data
				}
			}
		}

		// Emotes
		if tags["emotes"] != "" {
			msg = formatEmotes(msg, tags["emotes"].(string))
		}

		// UserName
		var UserNameColor string
		if tags["color"] != "" {
			UserNameColor = tags["color"].(string)
		} else {
			UserNameColor = "#" + RandStringBytesMaskImprSrc(6) // RandomColor
		}
		htmlDom += "<span class='username' style=\"color: " + UserNameColor + "\">" + tags["display-name"].(string) + ":</span>"

		// Message
		htmlDom += "<span class='message'"
		if strings.Contains(strings.ToLower(msg), channel) {
			styleColor = "color: red"
		} else if tags["banDuration"] != nil { // ban-duration (Optional) Duration of the timeout, in seconds. If omitted, the ban is permanent.
			styleColor = "text-decoration-line: 'line-through', text-decoration-style: 'wavy', text-decoration-color: 'rgba(255, 193, 7, 0.8)'"
		} else if tags["banReason"] != nil {
			styleColor = "text-decoration-line: 'line-through', text-decoration-style: 'wavy', text-decoration-color: 'rgba(255, 0, 0, 0.8)'"
		}
		if styleColor != "" {
			htmlDom += " style=\"" + styleColor + "\">"
		} else {
			htmlDom += ">"
		}
		htmlDom += msg + "</span>"

		if tags["banMsg"] != nil {
			htmlDom = "<span style=\"font-weight: 'bold'\" class='banMsg'>" + tags["banMsg"].(string) + "</span>"
		}

		htmlDom += "</span>"
		window.Call("OnPrivateMessage", map[string]interface{}{
			"htmlDom": htmlDom,
			"channel": channel,
			"tags":    tags,
			"message": msg,
		})
	}

	bot.Run()
}

// MyGoFunc returns a JavaScript function
func StartTwitchChat() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		len := args[0].Get("channels").Length()
		Channel := make([]string, 0, len)
		for i := 0; i < len; i++ {
			iString := strconv.Itoa(i)
			Channel = append(Channel, args[0].Get("channels").Get(iString).String())
		}

		go startBot(Channel)

		return map[string]interface{}{
			"success": "true",
		}
	})
}

func TagsToStrings(Tags map[string][]byte) map[string]interface{} {
	result := make(map[string]interface{})
	for i, v := range Tags {
		result[i] = string(v)
	}
	return result
}
