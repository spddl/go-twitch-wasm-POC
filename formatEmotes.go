package main

import (
	"log"
	"sort"
	"strconv"
	"strings"
)

type emoticons struct {
	von      int
	bis      int
	template []string
}

func formatEmotes(text, emotes string) string {
	splitText := strings.Split(text, "")

	sortEmotes := parseEmoticons(emotes)
	sort.Slice(sortEmotes, func(i, j int) bool { return sortEmotes[i].von > sortEmotes[j].von })

	for _, v := range sortEmotes {
		splitText = RemoveIndex(splitText, v.von, v.bis, v.template)
	}
	return strings.Join(splitText, "")
}

// emoticonID:von-bis,von-bis/emoticonID:von-bis,von-bis
func parseEmoticons(emotes string) []emoticons {
	result := []emoticons{}
	emotesArray := strings.Split(emotes, "/")
	for _, e := range emotesArray {
		EmoteSep := strings.Index(e, ":")
		EID := e[:EmoteSep]
		EposArray := strings.Split(e[EmoteSep+1:], ",")

		for i := len(EposArray) - 1; i >= 0; i-- {
			VonBis := EposArray[i]
			VonBisArray := strings.Split(VonBis, "-")
			von, err := strconv.Atoi(VonBisArray[0])
			if err != nil {
				log.Println("err", err)
			}
			bis, err := strconv.Atoi(VonBisArray[1])
			if err != nil {
				log.Println("err", err)
			}
			result = append(result, emoticons{
				von:      von,
				bis:      bis + 1,
				template: strings.Split("<img class='emoticon' src='//static-cdn.jtvnw.net/emoticons/v1/"+EID+"/1.0' srcset='//static-cdn.jtvnw.net/emoticons/v1/"+EID+"/1.0 1x,//static-cdn.jtvnw.net/emoticons/v1/"+EID+"/2.0 2x,//static-cdn.jtvnw.net/emoticons/v1/"+EID+"/3.0 4x'>", ""),
			})
		}
	}
	return result
}

func RemoveIndex(s []string, index, endIndex int, template []string) []string {
	return append(s[:index], append(template, s[endIndex:]...)...)
}
