package main

import (
	"net/http"
	"io/ioutil"
	"log"
	"fmt"
	"strconv"
	"encoding/json"
	"bytes"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slack/checkasset", HandleSlackCheckAsset)
	err := http.ListenAndServe(":9090", mux)
	log.Fatalln(err)
}

func HandleSlackCheckAsset(w http.ResponseWriter, req *http.Request) {
	input := []string{
		`base:85000.0`,
		`api:https://hooks.slack.com/services/your_slack_link_here`,
		`btc:0.01`,
		`eth:0.01`,
		`ht:0.01`,
		`bch:0.01`,
		`eos:0.01`,
		`ont:0.01`,
		`gxs:0.01`}
	var res bytes.Buffer
	res.WriteString("*Asset Monkey* :monkey:\n")

	var sum float64
	var base float64

	for _, val := range input {
		pairs := strings.Split(val, ":")
		if pairs[0] == "api" {
			slackWebhood = pairs[1] + ":" + pairs[2]
			continue
		}
		if pairs[0] == "base" {
			base,_ = strconv.ParseFloat(pairs[1], 64)
			continue
		}
		asset := pairs[0]
		amount, err := strconv.ParseFloat(pairs[1], 64)
		if err != nil {
			log.Fatalln("Parsing float error: ", err)
		}
		total := TotalAsset(asset, amount)
		res.WriteString(FormatText(asset, total))
		sum += total
	}

	res.WriteString(FormatText("Total", sum))
	if base > float64(0) {
		res.WriteString(fmt.Sprintf("*Net Value*: %.2f\n", sum / base))
	}


	fmt.Fprint(w, res.String())
}

const (
	coinMarketPriceApi = "https://api.coinmarketcap.com/v2/ticker/"
)

var (
	slackWebhood = "https://hooks.slack.com/services/your_slack_link_here"
)

type coin struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Symbol string `json:"symbol"`
}
type listing struct {
	Data  []coin `json:"data"`
}
type requestInfo struct {
	Text string `json:"text"`
	Markdown bool `json:"mrkdwn"`
}
type unitData struct {
	Price float64 `json:"price"`
	OneHourChange float64 `json:"percent_change_1h"`
	OneDayChange float64  `json:"percent_change_24h"`
	OneWeekChange float64 `json:"percent_change_7d"`
}
type coinMarketResp struct {
	Data struct {
		Name string `json:"name"`
		Symbol string `json:"symbol"`
		Quotes struct {
			USD unitData `json:"USD"`
			CNY unitData `json:"CNY"`
		} `json:"quotes"`
	} `json:"data"`
}

func SendToSlack(jsonStr []byte) {
	req, _ := http.NewRequest("POST", slackWebhood, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func TotalAsset(asset string, amount float64) float64 {
	return amount * QueryPrice(QueryId(asset))
}

func QueryId(asset string) int {
	raw, err := ioutil.ReadFile("./coinmarket_id.json")
	if err != nil {
		log.Fatalln("File error: ", err)
	}
	var assetList listing
	json.Unmarshal(raw, &assetList)

	for _, val := range assetList.Data {
		if strings.ToLower(asset) == strings.ToLower(val.Symbol) {
			return val.Id
		}
	}

	return 0
}

func QueryPrice(id int) float64{
	fmt.Println("Start the exciting apps!")
	requestUri := coinMarketPriceApi + strconv.Itoa(id) + "/?convert=CNY"
	resp, err := http.Get(requestUri)
	if err != nil {
		log.Fatalln("Http get error: ", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Read resp body error: ", err)
	}

	var cmResp coinMarketResp
	err = json.Unmarshal(body, &cmResp)
	if err != nil {
		log.Fatalln("Unmarshal body error: ", err)
	}

	return cmResp.Data.Quotes.CNY.Price
}

func FormatText(asset string, total float64) string {
	return fmt.Sprintf("%s: %f CNY\n", asset, total)
}
