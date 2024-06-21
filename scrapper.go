package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Scrapper interface {
	ScrapMajors() error
	ScrapMinors() error
}

var tickers = make(map[string]*Ticker)

func StartScrapping() {
	for {
		err := scrapMajors()
		if err != nil {
			log.Println("Error scrapping majors:", err)
		}
		err = scrapMinors()
		if err != nil {
			log.Println("Error Scraping minors:", err)
		}
		time.Sleep(1 * time.Minute) // 1-minute interval
	}
}

func scrapMajors() error {
	url := "https://www.tradingview.com/markets/currencies/rates-major/"
	return scrapUrl(url)
}

func scrapMinors() error {
	url := "https://www.tradingview.com/markets/currencies/rates-minor/"
	return scrapUrl(url)
}

func scrapUrl(url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}
	elements := doc.Find("tbody tr")
	elements.Each(func(index int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() >= 8 {

			symbol := strings.TrimSpace(cells.Eq(0).Text()[0:6])
			livePriceStr := strings.TrimSpace(cells.Eq(1).Text())
			dailyHighStr := strings.TrimSpace(cells.Eq(6).Text())
			dailyLowStr := strings.TrimSpace(cells.Eq(7).Text())
			cleanLivePrice := strings.Replace(livePriceStr, ",", "", -1)
			cleanDailyhigh := strings.Replace(dailyHighStr, ",", "", -1)
			cleanDailyLow := strings.Replace(dailyLowStr, ",", "", -1)

			livePrice, err := strconv.ParseFloat(cleanLivePrice, 64)
			if err != nil {
				log.Printf("Error parsing live price for %s: %v", symbol, err)
				return
			}
			dailyHigh, err := strconv.ParseFloat(cleanDailyhigh, 64)
			if err != nil {
				log.Printf("Error parsing daily high price for %s: %v", symbol, err)
				return
			}
			dailyLow, err := strconv.ParseFloat(cleanDailyLow, 64)
			if err != nil {
				log.Printf("Error parsing daily low price for %s: %v", symbol, err)
				return
			}
			ticker, exists := tickers[symbol]
			if exists {
				ticker.Update(livePrice, dailyHigh, dailyLow)
			} else {
				t := NewTicker(symbol, livePrice, dailyHigh, dailyLow)
				tickers[symbol] = t
			}
		} else {
			log.Println("Scrapper does not have sufficeint table columns.")
		}
	})
	return nil
}
