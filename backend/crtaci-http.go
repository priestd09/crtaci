// Author: Milan Nikolic <gen2brain@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.google.com/p/google-api-go-client/googleapi/transport"
	youtube "code.google.com/p/google-api-go-client/youtube/v3"
	"github.com/gen2brain/vidextr"
)

var (
	appName    = "crtaci-http"
	appVersion = "1.3"
)

type Cartoon struct {
	Id             string    `json:"id"`
	Character      string    `json:"character"`
	Title          string    `json:"title"`
	FormattedTitle string    `json:"formattedTitle"`
	Episode        int       `json:"episode"`
	Season         int       `json:"season"`
	Service        string    `json:"service"`
	Url            string    `json:"url"`
	Thumbnails     Thumbnail `json:"thumbnails"`
}

type Thumbnail struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

type Character struct {
	Name     string `json:"name"`
	AltName  string `json:"altname"`
	AltName2 string `json:"altname2"`
	Duration string `json:"duration"`
	Query    string `json:"query"`
}

var characters = []Character{
	{"atomski mrav", "", "", "medium", ""},
	{"a je to", "", "", "medium", "a je to crtani"},
	{"bananamen", "", "", "medium", ""},
	{"blinki bil", "", "блинки бил", "long", ""},
	{"blufonci", "", "", "medium", ""},
	{"bombončići", "bomboncici", "", "medium", ""},
	{"braća grim", "braca grim", "najlepse bajke", "long", ""},
	{"brzi gonzales", "", "", "medium", ""},
	{"čarli braun", "carli braun", "", "medium", ""},
	{"čili vili", "cili vili", "", "medium", ""},
	{"cipelići", "cipelici", "", "medium", ""},
	{"denis napast", "", "", "long", ""},
	{"droidi", "", "", "long", ""},
	{"duško dugouško", "dusko dugousko", "dusko 20dugousko", "medium", ""},
	{"džoni test", "dzoni test", "", "long", ""},
	{"elmer", "", "", "medium", "elmer crtani"},
	{"eustahije brzić", "eustahije brzic", "", "medium", ""},
	{"evoksi", "", "", "long", ""},
	{"generalova radnja", "", "", "medium", ""},
	{"gustav", "gustavus", "", "medium", "gustavus crtani"},
	{"helo kiti", "", "", "medium", ""},
	{"hi men i gospodari svemira", "himen i gospodari svemira", "", "long", ""},
	{"inspektor radiša", "inspektor radisa", "", "medium", ""},
	{"iznogud", "", "", "medium", ""},
	{"kalimero", "", "kalimero - ", "medium", "kalimero- crtani"},
	{"kasper", "", "", "medium", "kasper crtani"},
	{"kuče dragoljupče", "kuce dragoljupce", "", "medium", ""},
	{"lale gator", "", "", "medium", ""},
	{"la linea", "", "", "medium", ""},
	{"legenda o tarzanu", "", "", "long", ""},
	{"le piaf", "", "", "short", ""},
	{"mali leteći medvjedići", "mali leteci medvjedici", "", "long", ""},
	{"masa i medved", "masha i medved", "masa i medvjed", "medium", "masa i medved crtani"},
	{"mačor mika", "macor mika", "", "long", ""},
	{"mece dobrići", "mece dobrici", "", "medium", ""},
	{"miki maus", "", "", "medium", ""},
	{"mornar popaj", "", "", "medium", ""},
	{"nindža kornjače", "nindza kornjace", "ninja kornjace", "long", ""},
	{"ogi i žohari", "ogi i zohari", "", "long", ""},
	{"otkrića bez granica", "otkrica bez granica", "", "long", ""},
	{"paja patak", "", "", "medium", ""},
	{"patak dača", "patak daca", "", "medium", ""},
	{"pepa prase", "", "", "medium", ""},
	{"pepe le tvor", "", "", "medium", ""},
	{"pera detlić", "pera detlic", "", "medium", ""},
	{"pera kojot", "", "", "medium", ""},
	{"pink panter", "", "", "medium", ""},
	{"plava princeza", "", "", "long", ""},
	{"porodica kremenko", "", "", "long", ""},
	{"poručnik draguljče", "porucnik draguljce", "", "medium", ""},
	{"profesor baltazar", "", "", "medium", ""},
	{"ptica trkačica", "ptica trkacica", "", "medium", ""},
	{"rakuni", "", "", "long", ""},
	{"ren i stimpi", "", "", "medium", ""},
	{"robotek", "", "robotech", "long", ""},
	{"šalabajzerići", "salabajzerici", "", "medium", ""},
	{"silvester", "", "silvester i tviti", "medium", "silvester crtani"},
	{"šilja", "silja", "", "medium", "silja crtani"},
	{"snorkijevci", "", "", "medium", ""},
	{"sofronije", "", "", "medium", ""},
	{"super miš", "super mis", "", "medium", "super mis crtani"},
	{"supermen", "", "", "medium", "supermen crtani"},
	{"sport bili", "", "", "medium", ""},
	{"srle i pajče", "srle i pajce", "", "medium", ""},
	{"stanlio i olio", "", "", "medium", ""},
	{"stari crtaći", "stari crtaci", "stari sinhronizovani crtaci", "medium", ""},
	{"stripi", "", "", "medium", ""},
	{"štrumfovi", "strumpfovi", "strumfovi", "medium", "strumfovi crtani"},
	{"sundjer bob kockalone", "sundjer bob", "sunđer bob", "medium", ""},
	{"talični tom", "talicni tom", "", "long", ""},
	{"tarzan gospodar džungle", "tarzan gospodar dzungle", "", "long", ""},
	{"tom i džeri", "tom i dzeri", "", "medium", ""},
	{"transformersi", "", "", "long", ""},
	{"vitez koja", "", "", "medium", ""},
	{"voltron force", "", "", "long", "voltron force crtani"},
	{"vuk vučko", "vuk vucko", "", "medium", ""},
	{"wumi", "", "wummi", "short", "wumi crtani"},
	{"zamenik boža", "zamenik boza", "", "medium", ""},
	{"zmajeva kugla", "zmajeva kugla", "zmajeva kugla z", "long", ""},
}

var filters = []string{
	"najbolji crtaci",
	"www.crtani-filmovi.org",
	"by crtani svijet",
	"crtanifilmonline",
	"crtani filmovi",
	"crtani film",
	"stari crtani",
	"cijeli crtani",
	"crtani",
	"crtic",
	"sinhronizovano",
	"sihronizovano",
	"sinhronizovani",
	"sinhronizovan",
	"sinhronizacija",
	"sinkronizacija",
	"titlovano",
	"sa prevodom",
	"nove epizode",
	"na srpskom jeziku",
	"na srpskom",
	"srpska",
	"srpski",
	"srb ",
	" srb",
	" sd",
	" hq",
	" svcd",
	"hrvatska",
	"hrv,srp,bos",
	"zagorci",
	"slovenska verzija",
	"b92 tv",
	"za decu",
	"zadecu",
	"youtube",
	"you",
	"youtube",
	"full movie",
	"mashini skazki",
	"the cartooner 100",
	"iz 60-70-80-tih",
	"mpeg4",
	"144p h 264 aac",
	"sihroni fll 2",
	"zlekedik",
	"gusztav allast keres",
	"guszt v k",
	"rtb",
	"djuza stoiljkovic",
	"okrenite preko smplayer-a",
	"new episodes",
}

var censoredWords = []string{
	"kurac",
	"kurcu",
	"sranje",
	"sranja",
	"govna",
	"govno",
	"picka",
	"picke",
	"peder",
	"uzivo",
	"parodija",
	"tretmen",
	"ispaljotka",
	"kinder jaja",
	"video igrice",
	"atomski mravi",
	"igracke od plastelina",
	"sex",
	"sexy",
	"flesh",
	"ubisoft",
	"wanna",
	"special",
	"trailer",
	"teaser",
	"music",
	"monster",
	"intro",
	"countdown",
	"eternity",
	"summer",
	"galaxy",
	"constitution",
	"hunkyard",
	"riders",
	"flash",
	"wanted",
	"instrumental",
	"gamer",
	"remix",
	"tour",
	"party",
	"bjorke",
	"tweety",
	"revolution",
	"halloween",
	"remastered",
	"celebration",
	"experiments",
	"gameplay",
	"surprise",
	"batters",
	"bottle",
	"erasers",
	"series",
	"comics",
	"village",
	"theatre",
	"dolphin",
	"stallone",
	"koniec",
	"latino",
	"lovers",
	"lubochka",
	"tobbe",
	"sushi",
	"prikljuchenija",
	"slovenska",
	"aakerhus",
	"sylvia",
	"deutsch",
	"remue",
	"kespar",
	"splitter",
	"desierto",
	"pelicula",
	"episodio",
	"rwerk",
	"xhaven",
	"erkste",
	"przytul",
	"potenzf",
	"szalony",
	"schweiz",
	"verkackt",
	"sottile",
	"goldene",
	"elasmosaurio",
	"ombra",
	"ehlers",
	"dejas",
	"capitulo",
	"et ses",
	"tu sais",
	"ma vision",
	"how could",
	"new year",
	" del ",
}

var censoredIds = []string{
	"52vfFeJERfQ",
	"DLk4SLmIDUU",
	"VsNOHQfm02M",
	"yfYfGnCVbHs",
	"8zZ6tg2LXiM",
	"-1CnR5qVh5E",
	"DPxb3-7lakw",
	"zKhPpVTUn_Q",
	"vBggIcqV1rc",
	"YrmKYtDnthk",
	"YzqWmqeR43I",
	"Id3kHQC9vPI",
	"Ngke-HPnHok",
	"7VPtdqnHxHw",
	"Q6hTJ11ZGwU",
	"n28-lRu5cpw",
	"dl_kPk276oo",
	"YsdOt6qc6o4",
	"Tm7mOlgPlxA",
	"X8BwFSHJpg4",
	"QYrsrjgGh5g",
	"_z6pgpPDXBY",
	"7_ys2vKapLg",
	"G7SnbTCsj28",
	"2LzVPEoiacY",
	"8QJozzsvPnU",
	"KpLrIWB78sQ",
	"CuV0mDu4GL4",
	"c1-ywGJfS8U",
	"sotlkpiczWk",
	"wPUhMP7aGnw",
	"AR0Jc1rh2N0",
	"xuGex-B3GbQ",
	"drEJEbHDgIA",
	"JF4qkkgQsO4",
	"Y4r7m-Payv8",
	"0HDbPXN-HaE",
	"yONB3IwxtlQ",
	"BjtFDLOmEu8",
	"cpF73znG7UM",
	"2onSjJVgtpg",
	"gRLppbNNCLI",
	"smKnRR0ouds",
	"xy53o1",
	"xy53q1",
	"x3osiz",
	"x25ja2c",
	"x4aha4",
	"x7k5hx",
	"x60rr7",
	"x7k5ko",
	"x4e7mn",
	"xs4jyr",
	"x7wviw",
	"x5nyjf",
	"4562474",
	"21508130",
	"14072389",
	"145041047",
	"163980203",
	"165572645",
	"168855693",
}

var (
	reTitle = regexp.MustCompile(`[0-9A-Za-zžćčšđ_,]+`)
	reAlpha = regexp.MustCompile(`[A-Za-zžćčšđ]+`)
	reDesc  = regexp.MustCompile(`(?U)(\(|\[).*(\)|\])`)
	reYear  = regexp.MustCompile(`(19\d{2}|20\d{2})`)
	reExt   = regexp.MustCompile(`\.(?i:avi|mp4|flv|wmv|mpg|mpeg|mpeg4)$`)
	reRip   = regexp.MustCompile(`(?i:xvid)?(tv|dvd)?(-|\s)(rip)(bg)?(audio)?`)
	reChars = regexp.MustCompile(`(?i:braca grimm|i snupi [sš]ou|i snupi|charlie brown and snoopy|brzi gonzales i patak da[cč]a|patak da[cč]a i brzi gonzales|patak da[cč]a i elmer|patak da[cč]a i gicko prasi[cć]|i hello kitty|tom and jerry|tom i d[zž]eri [sš]ou|spongebob squarepants|paja patak i [sš]ilja|bini i sesil|masha i medved|elmer fudd|blinkibil|kockalone|najlepse bajke|stari sinhronizovani crtaci|popeye the sailor|kasper i drugari,|leghorn)`)
	reTime  = regexp.MustCompile(`(\d{2})h(\d{2})m(\d{2})s`)
	rePart  = regexp.MustCompile(`\s([\diI]{1,2})\.?\s?(?i:/|deo|od|part)\s?([\diI]{1,2})?\s*(?i:deo)?`)
	rePart2 = regexp.MustCompile(`\s(?i:pt)\s?(\d{1,2})\s*`)

	reTitleR     = regexp.MustCompile(`^(\d{1,2}\.?)\s?(\d{1,})?(.*)$`)
	reTitleNotEp = regexp.MustCompile(`\d{2,}\s(?i:razbojnika|sati|malih|pljeskavica)`)
	reTitle20    = regexp.MustCompile(`(\s20)`)

	reE1 = regexp.MustCompile(`(?i:epizoda|epizida|epzioda|episode|epizodas|episoda|Эпизод)\s?(\d{1,3})`)
	reE2 = regexp.MustCompile(`(\d{1,3})\.?-?\s?(?i:epizoda|epizida|epzioda|episode|epizodas|episoda)`)
	reE3 = regexp.MustCompile(`\s(?i:ep|e)\.?\s*(\d{1,3})`)
	reE4 = regexp.MustCompile(`(?:^|-|\.|\s)\s?(\d{1,3}\b)`)
	reE5 = regexp.MustCompile(`(?i:s)(?:\d{1,2})(?i:e)(\d{2})(?:\d{1})?(?:a|b)`)
	reE6 = regexp.MustCompile(`(?i:s)(?:\d{1,2})(?:e)(\d{1,2})`)

	reS1 = regexp.MustCompile(`(?i:sezona|sezon)\s?(\d{1,2})`)
	reS2 = regexp.MustCompile(`(\d{1,2})\.?\s?(?i:sezona|sezon)`)
	reS3 = regexp.MustCompile(`(?i:s)\s?(\d{1,2})`)
	reS4 = regexp.MustCompile(`(\d{1,2})(?i:x)`)
)

var (
	wg       sync.WaitGroup
	cartoons []Cartoon
)

type lessFunc func(p1, p2 *Cartoon) bool

type multiSorter struct {
	cartoons []Cartoon
	less     []lessFunc
}

func (ms *multiSorter) Sort(cartoons []Cartoon) {
	ms.cartoons = cartoons
	sort.Sort(ms)
}

func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

func (ms *multiSorter) Len() int {
	return len(ms.cartoons)
}

func (ms *multiSorter) Swap(i, j int) {
	ms.cartoons[i], ms.cartoons[j] = ms.cartoons[j], ms.cartoons[i]
}

func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.cartoons[i], &ms.cartoons[j]
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return ms.less[k](p, q)
}

func YouTube(character Character) {

	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered in YouTube: ", r)
		}
	}()

	const apiKey = "YOUR_API_KEY"

	httpClient := &http.Client{
		Transport: &transport.APIKey{Key: apiKey},
	}

	yt, err := youtube.New(httpClient)
	if err != nil {
		log.Print("Error creating YouTube client: %v", err)
		return
	}

	name := strings.ToLower(character.Name)
	altname := strings.ToLower(character.AltName)
	altname2 := strings.ToLower(character.AltName2)

	getResponse := func(token string) *youtube.SearchListResponse {
		apiCall := yt.Search.List("id,snippet").
			Q(getQuery(character, false)).
			MaxResults(50).
			VideoDuration(character.Duration).
			Type("video").
			PageToken(token)

		response, err := apiCall.Do()
		if err != nil {
			log.Print("Error making YouTube API call: %v", err.Error())
			return nil
		}
		return response
	}

	parseResponse := func(response *youtube.SearchListResponse) {
		for _, video := range response.Items {
			videoId := video.Id.VideoId
			videoTitle := strings.ToLower(video.Snippet.Title)
			videoThumbSmall := video.Snippet.Thumbnails.Default.Url
			videoThumbMedium := video.Snippet.Thumbnails.Medium.Url
			videoThumbLarge := video.Snippet.Thumbnails.High.Url

			if isValidTitle(videoTitle, name, altname, altname2, videoId) {
				formattedTitle := getFormattedTitle(videoTitle, name, altname, altname2)

				cartoon := Cartoon{
					videoId,
					name,
					videoTitle,
					formattedTitle,
					getEpisode(videoTitle),
					getSeason(videoTitle),
					"youtube",
					"https://www.youtube.com/watch?v=" + videoId,
					Thumbnail{videoThumbSmall, videoThumbMedium, videoThumbLarge},
				}

				cartoons = append(cartoons, cartoon)
			}
		}
	}

	response := getResponse("")
	parseResponse(response)

	if response.NextPageToken != "" {
		response = getResponse(response.NextPageToken)
		parseResponse(response)
	}

}

func DailyMotion(character Character) {

	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered in DailyMotion: ", r)
		}
	}()

	uri := "https://api.dailymotion.com/videos?search=%s&fields=id,title,url,duration,thumbnail_120_url,thumbnail_360_url,thumbnail_480_url&limit=50&page=%s&sort=relevance"

	name := strings.ToLower(character.Name)
	altname := strings.ToLower(character.AltName)
	altname2 := strings.ToLower(character.AltName2)

	timeout := time.Duration(6 * time.Second)

	dialTimeout := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}

	transport := http.Transport{
		Dial: dialTimeout,
	}

	httpClient := http.Client{
		Transport: &transport,
	}

	getResponse := func(page string) ([]interface{}, bool) {
		res, err := httpClient.Get(fmt.Sprintf(uri, getQuery(character, true), page))
		if err != nil {
			log.Print("Error making DailyMotion API call: %v", err.Error())
			return nil, false
		}
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Print("Error unmarshaling json: %v", err.Error())
			return nil, false
		}

		hasMore := data["has_more"].(bool)
		response, ok := data["list"].([]interface{})
		if !ok {
			return nil, false
		}

		if len(response) == 0 {
			return nil, false
		}

		return response, hasMore
	}

	parseResponse := func(response []interface{}) {
		for _, obj := range response {
			video, ok := obj.(map[string]interface{})
			if !ok {
				continue
			}

			videoId := video["id"].(string)
			videoTitle := strings.ToLower(video["title"].(string))
			videoUrl := video["url"].(string)
			videoThumbSmall := video["thumbnail_120_url"].(string)
			videoThumbMedium := video["thumbnail_360_url"].(string)
			videoThumbLarge := video["thumbnail_480_url"].(string)

			videoDuration := getDuration(video["duration"].(float64))

			if isValidTitle(videoTitle, name, altname, altname2, videoId) && character.Duration == videoDuration {
				formattedTitle := getFormattedTitle(videoTitle, name, altname, altname2)

				cartoon := Cartoon{
					videoId,
					name,
					videoTitle,
					formattedTitle,
					getEpisode(videoTitle),
					getSeason(videoTitle),
					"dailymotion",
					videoUrl,
					Thumbnail{videoThumbSmall, videoThumbMedium, videoThumbLarge},
				}

				cartoons = append(cartoons, cartoon)
			}
		}
	}

	response, hasMore := getResponse("1")
	if response != nil {
		parseResponse(response)
	}

	if hasMore {
		response, _ := getResponse("2")
		if response != nil {
			parseResponse(response)
		}
	}

}

func Vimeo(character Character) {

	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered in Vimeo: ", r)
		}
	}()

	const apiKey = "YOUR_API_KEY"
	uri := "https://api.vimeo.com/videos?query=%s&page=%s&per_page=100&sort=relevant"

	name := strings.ToLower(character.Name)
	altname := strings.ToLower(character.AltName)
	altname2 := strings.ToLower(character.AltName2)

	timeout := time.Duration(6 * time.Second)

	dialTimeout := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}

	transport := http.Transport{
		Dial: dialTimeout,
	}

	httpClient := http.Client{
		Transport: &transport,
	}

	getResponse := func(page string) []interface{} {
		req, err := http.NewRequest("GET", fmt.Sprintf(uri, getQuery(character, true), page), nil)
		if err != nil {
			log.Print("Error making Vimeo API call: %v", err.Error())
			return nil
		}

		req.Header.Set("Authorization", "bearer "+apiKey)
		req.Header.Set("Accept", "application/vnd.vimeo.video+json;version=3.2")
		res, err := httpClient.Do(req)
		if err != nil {
			log.Print("Error making Vimeo API call: %v", err.Error())
			return nil
		}
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Print("Error unmarshaling json: %v", err.Error())
			return nil
		}

		response, ok := data["data"].([]interface{})
		if !ok {
			return nil
		}

		if len(response) == 0 {
			return nil
		}

		return response
	}

	parseResponse := func(response []interface{}) {
		for _, obj := range response {
			video, ok := obj.(map[string]interface{})
			if !ok {
				continue
			}

			videoId := strings.Replace(video["link"].(string), "https://vimeo.com/", "", -1)
			videoTitle := strings.ToLower(video["name"].(string))
			videoUrl := video["link"].(string)

			pictures, ok := video["pictures"].(map[string]interface{})
			if !ok {
				continue
			}

			sizes := pictures["sizes"].([]interface{})

			if len(sizes) < 4 {
				continue
			}

			videoThumbSmall := sizes[3].(map[string]interface{})["link"].(string)
			videoThumbMedium := sizes[2].(map[string]interface{})["link"].(string)
			videoThumbLarge := sizes[1].(map[string]interface{})["link"].(string)

			videoDuration := getDuration(video["duration"].(float64))

			if isValidTitle(videoTitle, name, altname, altname2, videoId) && character.Duration == videoDuration {
				formattedTitle := getFormattedTitle(videoTitle, name, altname, altname2)

				cartoon := Cartoon{
					videoId,
					name,
					videoTitle,
					formattedTitle,
					getEpisode(videoTitle),
					getSeason(videoTitle),
					"vimeo",
					videoUrl,
					Thumbnail{videoThumbSmall, videoThumbMedium, videoThumbLarge},
				}

				cartoons = append(cartoons, cartoon)
			}
		}
	}

	response := getResponse("1")
	if response != nil {
		parseResponse(response)
	}

}

func VK(character Character) {

	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered in VK: ", r)
		}
	}()

	const apiKey = "YOUR_API_KEY"
	uri := "https://api.vk.com/method/video.search?q=%s&count=100&sort=2&access_token=%s&v=5.26"

	name := strings.ToLower(character.Name)
	altname := strings.ToLower(character.AltName)
	altname2 := strings.ToLower(character.AltName2)

	timeout := time.Duration(6 * time.Second)

	dialTimeout := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}

	transport := http.Transport{
		Dial: dialTimeout,
	}

	httpClient := http.Client{
		Transport: &transport,
	}

	getResponse := func() []interface{} {
		res, err := httpClient.Get(fmt.Sprintf(uri, getQuery(character, true), apiKey))
		if err != nil {
			log.Print("Error making VK API call: %v", err.Error())
			return nil
		}
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Print("Error unmarshaling json: %v", err.Error())
			return nil
		}

		response, ok := data["response"].(map[string]interface{})
		if !ok {
			return nil
		}

		items, ok := response["items"].([]interface{})
		if !ok {
			return nil
		}

		if len(items) == 0 {
			return nil
		}

		return items
	}

	parseResponse := func(response []interface{}) {
		for _, obj := range response {
			video, ok := obj.(map[string]interface{})
			if !ok {
				continue
			}

			videoId := fmt.Sprintf("%.0f", video["id"].(float64))
			videoTitle := strings.ToLower(video["title"].(string))
			videoUrl := video["player"].(string)
			videoThumbSmall := video["photo_130"].(string)
			videoThumbMedium := video["photo_320"].(string)
			videoThumbLarge := video["photo_320"].(string)

			videoDuration := getDuration(video["duration"].(float64))

			if strings.Contains(videoUrl, "youtube") || strings.Contains(videoUrl, "vimeo") {
				continue
			}

			if isValidTitle(videoTitle, name, altname, altname2, videoId) && character.Duration == videoDuration {
				formattedTitle := getFormattedTitle(videoTitle, name, altname, altname2)

				cartoon := Cartoon{
					videoId,
					name,
					videoTitle,
					formattedTitle,
					getEpisode(videoTitle),
					getSeason(videoTitle),
					"vk",
					videoUrl,
					Thumbnail{videoThumbSmall, videoThumbMedium, videoThumbLarge},
				}

				cartoons = append(cartoons, cartoon)
			}
		}
	}

	response := getResponse()
	if response != nil {
		parseResponse(response)
	}

}

func getDuration(videoDuration float64) string {
	minutes := videoDuration / 60
	switch {
	case minutes < 4 && minutes > 0:
		return "short"
	case minutes >= 4 && minutes <= 20:
		return "medium"
	case minutes > 20:
		return "long"
	default:
		return "any"
	}
}

func getFormattedTitle(videoTitle string, name string, altname string, altname2 string) string {

	title := videoTitle

	part := ""
	p := rePart.FindAllStringSubmatch(title, -1)
	if len(p) > 0 {
		part = p[0][1]
	}

	p2 := rePart2.FindAllStringSubmatch(title, -1)
	if len(p2) > 0 {
		part = p2[0][1]
	}

	title = reYear.ReplaceAllString(title, "")

	re20 := reTitle20.FindAllStringSubmatch(title, -1)
	if len(re20) > 1 {
		title = reTitle20.ReplaceAllString(title, " ")
	}

	for _, filter := range filters {
		title = strings.Replace(title, filter, " ", -1)
	}

	for _, re := range []*regexp.Regexp{
		reDesc, reExt, reRip, reChars, reTime, rePart, rePart2,
		reE1, reS1, reS4, reE2, reE5, reE6, reE3, reS2, reS3} {
		title = re.ReplaceAllString(title, "")
	}

	matches := reTitle.FindAllString(title, -1)
	title = strings.Join(matches, " ")
	title = strings.Replace(title, "_", " ", -1)

	name = strings.Replace(name, "-", "", -1)
	name = strings.TrimRight(name, " ")

	if altname2 != "" {
		title = strings.Replace(title, altname2, "", 1)
	}
	if altname != "" {
		title = strings.Replace(title, altname+" ", "", 1)
	}
	title = strings.Replace(title, name, "", 1)

	title = strings.TrimLeft(title, " ")
	title = strings.TrimRight(title, " ")

	title = reTitleR.ReplaceAllString(title, "$3")

	if strings.HasPrefix(title, "i ") || strings.HasPrefix(title, "and ") || strings.HasPrefix(title, " i ") {
		title = fmt.Sprintf("%s %s", name, title)
	}

	if !reAlpha.MatchString(title) {
		title = name
	}

	if part != "" {
		title = fmt.Sprintf("%s - %s deo", title, part)
	}

	return title
}

func getEpisode(videoTitle string) int {
	title := videoTitle

	title = reYear.ReplaceAllString(title, "")

	re20 := reTitle20.FindAllStringSubmatch(title, -1)
	if len(re20) > 1 {
		title = reTitle20.ReplaceAllString(title, " ")
	}

	for _, filter := range filters {
		title = strings.Replace(title, filter, " ", -1)
	}
	for _, re := range []*regexp.Regexp{reDesc, reYear, reTime, rePart, rePart2, reS1, reS4} {
		title = re.ReplaceAllString(title, "")
	}

	ep := -1
	e1 := reE1.FindAllStringSubmatch(title, -1)
	if len(e1) > 0 {
		ep, _ = strconv.Atoi(e1[0][1])
		return ep
	}

	e2 := reE2.FindAllStringSubmatch(title, -1)
	if len(e2) > 0 {
		ep, _ = strconv.Atoi(e2[0][1])
		return ep
	}

	e5 := reE5.FindAllStringSubmatch(title, -1)
	if len(e5) > 0 {
		ep, _ = strconv.Atoi(e5[0][1])
		return ep
	}

	e6 := reE6.FindAllStringSubmatch(title, -1)
	if len(e6) > 0 {
		ep, _ = strconv.Atoi(e6[0][1])
		return ep
	}

	e3 := reE3.FindAllStringSubmatch(title, -1)
	if len(e3) > 0 {
		ep, _ = strconv.Atoi(e3[0][1])
		return ep
	}

	e4 := reE4.FindAllStringSubmatch(title, -1)
	notEp := reTitleNotEp.MatchString(title)
	if len(e4) > 0 && !notEp {
		ep, _ = strconv.Atoi(e4[0][1])
		if ep > 100 || ep == 0 {
			return -1
		}
		return ep
	}

	return ep
}

func getSeason(videoTitle string) int {
	title := videoTitle

	title = reYear.ReplaceAllString(title, "")

	re20 := reTitle20.FindAllStringSubmatch(title, -1)
	if len(re20) > 1 {
		title = reTitle20.ReplaceAllString(title, " ")
	}

	for _, re := range []*regexp.Regexp{reDesc, reYear, reTime, rePart, rePart2, reE1} {
		title = re.ReplaceAllString(title, "")
	}

	s := -1
	s1 := reS1.FindAllStringSubmatch(title, -1)
	if len(s1) > 0 {
		s, _ = strconv.Atoi(s1[0][1])
		return s
	}

	s2 := reS2.FindAllStringSubmatch(title, -1)
	if len(s2) > 0 {
		s, _ = strconv.Atoi(s2[0][1])
		return s
	}

	s3 := reS3.FindAllStringSubmatch(title, -1)
	if len(s3) > 0 {
		s, _ = strconv.Atoi(s3[0][1])
		if s >= 20 || s == 0 {
			return -1
		}
		return s
	}

	s4 := reS4.FindAllStringSubmatch(title, -1)
	if len(s4) > 0 {
		s, _ = strconv.Atoi(s4[0][1])
		if s >= 20 || s == 0 {
			return -1
		}
		return s
	}

	return s
}

func getQuery(char Character, escape bool) string {
	query := ""
	if char.Query != "" {
		query = char.Query
	} else if char.AltName != "" {
		query = char.AltName
	} else {
		query = char.Name
	}
	if escape {
		query = url.QueryEscape(query)
	}
	return query
}

func isCensored(videoTitle string, videoId string) bool {
	for _, word := range censoredWords {
		if strings.Contains(videoTitle, word) {
			return true
		}
	}
	for _, id := range censoredIds {
		if id == videoId {
			return true
		}
	}
	return false
}

func isValidTitle(videoTitle string, name string, altname string, altname2 string, videoId string) bool {
	videoTitle = reTitleR.ReplaceAllString(videoTitle, "$3")
	videoTitle = strings.TrimLeft(videoTitle, " ")

	if strings.HasPrefix(videoTitle, name) {
		if !isCensored(videoTitle, videoId) {
			return true
		}
	}
	if altname != "" {
		if strings.HasPrefix(videoTitle, altname) {
			if !isCensored(videoTitle, videoId) {
				return true
			}
		}
	}
	if altname2 != "" {
		if strings.HasPrefix(videoTitle, altname2) {
			if !isCensored(videoTitle, videoId) {
				return true
			}
		}
	}
	return false
}

func sortCartoons(cartoons []Cartoon) {
	episode := func(c1, c2 *Cartoon) bool {
		if c1.Episode == -1 {
			return false
		} else if c2.Episode == -1 {
			return true
		}
		return c1.Episode < c2.Episode
	}
	season := func(c1, c2 *Cartoon) bool {
		if c1.Season == -1 {
			return false
		} else if c2.Season == -1 {
			return true
		}
		return c1.Season < c2.Season
	}
	OrderedBy(season, episode).Sort(cartoons)
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Server", fmt.Sprintf("%s/%s", appName, appVersion))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func handleList(w http.ResponseWriter, r *http.Request) {
	setHeader(w)

	js, err := json.MarshalIndent(characters, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	setHeader(w)

	//path := html.EscapeString(r.URL.Path[1:])
	path := r.URL.Path[1:]
	path = strings.TrimRight(path, "/")
	paths := strings.Split(path, "/")

	if len(paths) > 1 {
		query := paths[1]

		char := new(Character)
		for _, character := range characters {
			if query == character.Name || query == character.AltName {
				char = &character
				break
			}
		}

		if char.Name != "" {
			wg.Add(3)
			cartoons = make([]Cartoon, 0)
			go YouTube(*char)
			go DailyMotion(*char)
			go Vimeo(*char)
			//go VK(*char)
			wg.Wait()

			sortCartoons(cartoons)

			js, err := json.MarshalIndent(cartoons, "", "    ")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Write(js)
		} else {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
}

func handleExtract(w http.ResponseWriter, r *http.Request) {
	setHeader(w)

	//path := html.EscapeString(r.URL.Path[1:])
	path := r.URL.Path[1:]
	path = strings.TrimRight(path, "/")
	paths := strings.Split(path, "/")

	if len(paths) == 3 {
		var url string
		var err error
		service := paths[1]
		videoId := paths[2]

		switch {
		case service == "youtube":
			url, err = vidextr.YouTube(videoId)
		case service == "dailymotion":
			url, err = vidextr.DailyMotion(videoId)
		case service == "vimeo":
			url, err = vidextr.Vimeo(videoId)
		case service == "vk":
			url, err = vidextr.VK(videoId)
		}

		if err != nil {
			log.Print(err.Error())
			return
		}

		js, err := json.Marshal(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(js)
	} else {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
}

func main() {
	bind := flag.String("bind", ":7313", "Bind address")
	flag.Parse()

	http.HandleFunc("/list", handleList)
	http.HandleFunc("/search/", handleSearch)
	http.HandleFunc("/extract/", handleExtract)

	l, err := net.Listen("tcp4", *bind)
	if err != nil {
		log.Fatal(err)
	}
	http.Serve(l, nil)
}
