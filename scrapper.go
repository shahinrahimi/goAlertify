package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var tickers = make(map[string]*Ticker)

func StartScrapping() {
	for {
		scrapForex()
		scrapFeatures()
		scrapCryptos()
		time.Sleep(5 * time.Minute) // 1-minute interval
	}
}

func scrapForex() {
	go scrap("https://www.tradingview.com/markets/currencies/rates-major/", processForex)
	go scrap("https://www.tradingview.com/markets/currencies/rates-minor/", processForex)
}

func scrapFeatures() {
	go scrap("https://www.tradingview.com/markets/futures/quotes-metals/", processFeatures)
	go scrap("https://www.tradingview.com/markets/futures/quotes-energy/", processFeatures)
}

func scrapCryptos() {
	go scrap("https://www.tradingview.com/markets/cryptocurrencies/prices-all/", processCryptos)
}

func scrap(url string, processFunc func(*goquery.Selection)) {
	res, err := http.Get(url)
	if err != nil {
		log.Println("Error gettomg URL:", err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("Error reading docuemnt:", err)
	}
	doc.Find("tbody tr").Each(func(index int, row *goquery.Selection) {
		processFunc(row)
	})
}

func processForex(row *goquery.Selection) {
	cells := row.Find("td")
	if cells.Length() < 8 {
		log.Println("Scrapper does not have sufficient table columns.")
		return
	}

	symbol := strings.TrimSpace(cells.Eq(0).Find("a").Text())
	name := strings.TrimSpace(cells.Eq(0).Find("sup").Eq(0).Text())
	livePriceStr := strings.TrimSpace(cells.Eq(1).Text())
	dailyHighStr := strings.TrimSpace(cells.Eq(6).Text())
	dailyLowStr := strings.TrimSpace(cells.Eq(7).Text())
	processPrices(symbol, name, livePriceStr, dailyHighStr, dailyLowStr, "forex")
}

func processFeatures(row *goquery.Selection) {
	cells := row.Find("td")
	if cells.Length() < 6 {
		log.Println("Scrapper does not have sufficient table columns.")
		return
	}

	symbol := strings.TrimSpace(strings.Replace(cells.Eq(0).Find("a").Text(), "!", "", -1))
	name := strings.TrimSpace(cells.Eq(0).Find("sup").Eq(0).Text())
	livePriceStr := strings.TrimSpace(cells.Eq(1).Text())
	dailyHighStr := strings.TrimSpace(cells.Eq(4).Text())
	dailyLowStr := strings.TrimSpace(cells.Eq(5).Text())
	processPrices(symbol, name, livePriceStr, dailyHighStr, dailyLowStr, "feature")
}

func processCryptos(row *goquery.Selection) {
	cells := row.Find("td")
	if cells.Length() < 3 {
		log.Println("Scrapper does not have sufficient table columns.")
		return
	}

	symbol := strings.TrimSpace(cells.Eq(0).Find("a").Eq(0).Text())
	name := strings.TrimSpace(cells.Eq(0).Find("sup").Eq(0).Text())
	livePriceStr := strings.TrimSpace(strings.Replace(cells.Eq(2).Text(), "USD", "", -1))
	processPrices(symbol, name, livePriceStr, "0", "0", "crypto")
}

func processPrices(symbol, name, livePriceStr, dailyHighStr, dailyLowStr, category string) {
	cleanLivePrice := strings.Replace(livePriceStr, ",", "", -1)
	livePrice, err := strconv.ParseFloat(cleanLivePrice, 64)
	if err != nil {
		log.Printf("Error parsing live price for %s: %v", symbol, err)
		return
	}

	var dailyHigh, dailyLow float64
	if dailyHighStr != "" {
		cleanDailyHigh := strings.Replace(dailyHighStr, ",", "", -1)
		dailyHigh, err = strconv.ParseFloat(cleanDailyHigh, 64)
		if err != nil {
			log.Printf("Error parsing daily high price for %s: %v", symbol, err)
			return
		}
	}
	if dailyLowStr != "" {
		cleanDailyLow := strings.Replace(dailyLowStr, ",", "", -1)
		dailyLow, err = strconv.ParseFloat(cleanDailyLow, 64)
		if err != nil {
			log.Printf("Error parsing daily low price for %s: %v", symbol, err)
			return
		}
	}

	ticker, exists := tickers[symbol]
	if exists {
		ticker.Update(livePrice, dailyHigh, dailyLow)

	} else {
		t := NewTicker(strings.ToLower(symbol), strings.ToLower(name), category, livePrice, dailyHigh, dailyLow)
		tickers[symbol] = t
		// fmt.Println(t.Category, t.Symbol, t.Name, t.LivePrice, t.DailyHigh, t.DailyLow)
	}
}
