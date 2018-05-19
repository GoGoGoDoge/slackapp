package main

import (
	"fmt"
	"net/http"
	"log"
	"io/ioutil"
	"os"
	"encoding/json"
	"strings"
	"strconv"
	"bytes"
)

const (
	//coinMarketListApi = "https://api.coinmarketcap.com/v2/listings/"
	coinMarketPriceApi = "https://api.coinmarketcap.com/v2/ticker/"
)

var (
	slackWebhood = "https://hooks.slack.com/services/your_slack_authentication_link_here"
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
			BTC unitData `json:"BTC"`
		} `json:"quotes"`
	} `json:"data"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Must provide only one argument [name of crypto asset]")
	}
	asset := os.Args[1]
	if len(os.Args) > 2 {
		slackWebhood=os.Args[2]
	}
	id := QueryId(asset)
	jsonStr := QueryPrice(id)
	req, _ := http.NewRequest("POST", slackWebhood, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
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

func QueryPrice(id int) []byte{
	fmt.Println("Start the exciting apps!")
	requestUri := coinMarketPriceApi + strconv.Itoa(id) + "/?convert=BTC"
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


	var info requestInfo
	info.Text = FormatText(cmResp)
	info.Markdown = true

	rawByte,_ := json.Marshal(info)
	return rawByte
}

func FormatText(resp coinMarketResp) string {
	var res bytes.Buffer
	res.WriteString("*Price Monkey* :monkey_face:\n")
	res.WriteString(resp.Data.Name + " (*_" + resp.Data.Symbol + "_*)\n")
	res.WriteString(fmt.Sprintf("USD: %f\n", resp.Data.Quotes.USD.Price))
	res.WriteString(fmt.Sprintf("\t1 Hour: %f%%\n", resp.Data.Quotes.USD.OneHourChange))
	res.WriteString(fmt.Sprintf("\t1 Day : %f%%\n", resp.Data.Quotes.USD.OneDayChange))
	res.WriteString(fmt.Sprintf("\t1 Week: %f%%\n", resp.Data.Quotes.USD.OneWeekChange))
	res.WriteString(fmt.Sprintf("BTC: %f\n", resp.Data.Quotes.BTC.Price))
	return res.String()
}
