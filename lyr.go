package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const Ua = "Mozilla/5.0"

var Artist, Title string

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s \"artist\" \"title\"\n", os.Args[0])
		os.Exit(1)
	}

	a, _ := url.QueryUnescape(os.Args[1])
	Artist = strings.ToLower(a)
	t, _ := url.QueryUnescape(os.Args[2])
	Title = strings.ToLower(t)

	res := GetSL() || GetWikia() || GetGenius() || GetMA() || GetMega() || GetMania() || GetLyr()
	if !res {
		fmt.Println("Nothing Found")
	}
}

func GetMania() bool {
	rep := strings.NewReplacer(" ", "_")
	Url := "https://www.lyricsmania.com/" + rep.Replace(Title+"_lyrics_"+Artist) + ".html"

	str := Get(Url)
	if i := strings.Index(str, `<div class="lyrics-body">`); i != -1 {
		s1 := str[i:]
		if j := strings.Index(s1, `</div>`); j != -1 {
			s2 := s1[j+6:]
			if end := strings.Index(s2, `</div>	<script>try`); end != -1 {
				r := strings.NewReplacer("<br/>", "", "<br>", "\n",
					`<div class="p402_premium">`, "")

				fmt.Println(r.Replace(s2[:end]))
				return true
			}
		}
	}
	return false
}

func GetLyr() bool {
	rep := strings.NewReplacer("_", "+", " ", "+")
	Url := "https://www.lyrics.com/serp.php?st=" + rep.Replace(Title) + "&stype=1"

	rep2 := strings.NewReplacer("_", " ")
	q := rep.Replace(strings.Title(rep2.Replace(Artist + "/" + Title)))

	var lUrl string
	str := Get(Url)
	if i := strings.Index(str, q+"\">"); i != -1 {
		s1 := str[i-38 : i]
		if j := strings.Index(s1, `/lyric/`); j != -1 {
			s2 := s1[j:]
			lUrl = "https://www.lyrics.com" + s2 + q
		}
	}

	if lUrl != "" {
		str := Get(lUrl)
		if i := strings.Index(str, `<pre id="lyric-body-text`); i != -1 {
			s1 := str[i:]
			if j := strings.Index(s1, `</pre>`); j != -1 {
				s2 := s1[:j]
				reg := regexp.MustCompile("<[^>]*>")
				res := reg.ReplaceAllString(s2, "")
				fmt.Println(res)
				return true
			}
		}
	}
	return false
}

func GetMega() bool {
	rep := strings.NewReplacer("_", "-", " ", "-")
	Url := "http://megalyrics.ru/lyric/" + rep.Replace(Artist+"/"+Title) + ".htm"

	str := Get(Url)
	if i := strings.Index(str, `text_inner">`); i != -1 {
		s1 := str[i+12:]
		if end := strings.Index(s1, `</div></div>`); end != -1 {
			r := strings.NewReplacer("&quot;", "\"", "<br />", "\n",
				`<div id="native_roll"></div>`, "")

			fmt.Println(r.Replace(s1[:end]))
			return true
		}
	}
	return false
}

func GetSL() bool {
	rep := strings.NewReplacer("_", "-", " ", "-")
	Url := "http://www.songlyrics.com/" + rep.Replace(Artist+"/"+Title) + "-lyrics/"

	str := Get(Url)
	if i := strings.Index(str, `songLyricsV14 iComment-text">`); i != -1 {
		s1 := str[i+29:]
		if end := strings.Index(s1, `</p>`); end != -1 {
			if !strings.Contains(s1[:end], "Sorry, we have no") &&
				!strings.Contains(s1[:end], "do not have the lyrics") {
				r := strings.NewReplacer("&quot;", "\"", "<br />", "",
					`&#039;`, "'", "&amp;amp;amp;amp;amp;#039;", "'")

				fmt.Println(r.Replace(s1[:end]))
				return true
			}
		}
	}
	return false
}

func GetWikia() bool {
	rep := strings.NewReplacer("_", " ")
	q := strings.Title(rep.Replace(Artist + ":" + Title))
	Url := "https://lyrics.fandom.com/api.php?action=query&prop=revisions&rvprop=content&format=xml&titles=" + url.PathEscape(q)

	str := Get(Url)
	if i := strings.Index(str, `lyrics&gt;`); i != -1 {
		s1 := str[i+10:]
		if end := strings.Index(s1, `&lt;`); end != -1 {
			r := strings.NewReplacer("&quot;", "\"", `&#039;`, "'")
			fmt.Println(r.Replace(s1[:end]))
			return true
		}
	}
	return false
}

func GetMA() bool {
	rep := strings.NewReplacer("_", "+", " ", "+")

	Url := "https://www.metal-archives.com/search/ajax-advanced/searching/songs?songTitle=" + rep.Replace(Title+"&bandName="+Artist) + "&releaseTitle=&lyrics=&genre=#songs"

	str := Get(Url)
	if i := strings.Index(str, `lyricsLink_`); i != -1 {
		s1 := str[i+11:]
		if end := strings.Index(s1, `\`); end != -1 {
			url := "https://www.metal-archives.com/release/ajax-view-lyrics/id/" + s1[:end]
			s2 := strings.ReplaceAll(Get(url), `<br />`, "")
			if !strings.Contains(s2, "lyrics not available") {
				fmt.Println(s2)
				return true
			}
		}
	}
	return false
}

func GetGenius() bool {
	rep := strings.NewReplacer("_", "-", " ", "-")
	Url := "https://genius.com/" + rep.Replace(Artist+"-"+Title) + "-lyrics"

	str := Get(Url)
	if i := strings.Index(str, `<div class="lyrics">`); i != -1 {
		s1 := str[i:]
		if start := strings.Index(s1, `<p>`); start != -1 {
			s2 := s1[start+3:]
			if end := strings.Index(s2, `</p>`); end != -1 {
				fmt.Println(strings.ReplaceAll(s2[:end], "<br>", ""))
				return true
			}
		}
	}
	return false
}

func Get(Url string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", Ua)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}
