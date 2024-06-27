package main

import (
	"fmt"
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
		// scrapForex()
		scrapMetals()
		// scrapCryptos()
		time.Sleep(1 * time.Minute) // 1-minute interval
	}
}
func scrapForex() {
	go scrapMinors()
	go scrapMajors()
}

func scrapMajors() {
	url := "https://www.tradingview.com/markets/currencies/rates-major/"
	if err := scrapUrl(url); err != nil {
		log.Println("Error scrapping majors:", err)
	}
}

func scrapMinors() {
	url := "https://www.tradingview.com/markets/currencies/rates-minor/"
	if err := scrapUrl(url); err != nil {
		log.Println("Error scrapping minors:", err)
	}
}

func scrapMetals() {
	url := "https://www.tradingview.com/markets/futures/quotes-metals/"
	if err := scrapUrlMetals(url); err != nil {
		log.Println("Error scrapping metals:", err)
	}
}

func scrapCryptos() {
	url := "https://www.tradingview.com/markets/cryptocurrencies/prices-all/"
	if err := scrapUrlCrypto(url); err != nil {
		log.Println("Error scrapping metals:", err)
	}
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

	symbol := strings.TrimSpace(strings.Split(cells.Eq(0).Text(), " ")[0])
	livePriceStr := strings.TrimSpace(cells.Eq(1).Text())
	dailyHighStr := strings.TrimSpace(cells.Eq(6).Text())
	dailyLowStr := strings.TrimSpace(cells.Eq(7).Text())
	processPrices(symbol, "forex", livePriceStr, dailyHighStr, dailyLowStr)
}

func processFeatures(row *goquery.Selection) {
	cells := row.Find("td")
	if cells.Length() < 3 {
		log.Println("Scrapper does not have sufficient table columns.")
		return
	}

	symbol := strings.TrimSpace(cells.Eq(0).Find("a").Eq(0).Text())
	name := strings.TrimSpace(cells.Eq(0).Find("sup").Eq(0).Text())
	livePriceStr := strings.TrimSpace(cells.Eq(1).Text())
	processPrices(symbol, name, livePriceStr, "0", "0")
}

func processCryptos(row *goquery.Selection) {
	cells := row.Find("td")
	if cells.Length() < 8 {
		log.Println("Scrapper does not have sufficient table columns.")
		return
	}

	symbol := strings.TrimSpace(cells.Eq(0).Find("a").Eq(0).Text())
	name := strings.TrimSpace(cells.Eq(0).Find("sup").Eq(0).Text())
	livePriceStr := strings.TrimSpace(strings.Replace(cells.Eq(2).Text(), "USD", "", -1))
	processPrices(symbol, name, livePriceStr, "0", "0")
}

func processPrices(symbol, name, livePriceStr, dailyHighStr, dailyLowStr string) {
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
		t := NewTicker(symbol, name, "forex", livePrice, dailyHigh, dailyLow)
		tickers[symbol] = t
	}

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

			symbol := strings.TrimSpace(strings.Split(cells.Eq(0).Text(), " ")[0])
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
				t := NewTicker(symbol, "anything", "forex", livePrice, dailyHigh, dailyLow)
				tickers[symbol] = t
			}
		} else {
			log.Println("Scrapper does not have sufficeint table columns.")
		}
	})
	return nil
}

func scrapUrlCrypto(url string) error {
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

			symbol := strings.TrimSpace(cells.Eq(0).Find("a").Eq(0).Text())
			name := strings.TrimSpace(cells.Eq(0).Find("sup").Eq(0).Text())
			livePriceStr := strings.TrimSpace(strings.Replace(cells.Eq(2).Text(), "USD", "", -1))

			fmt.Println(symbol, name, livePriceStr)
			// dailyHighStr := strings.TrimSpace(cells.Eq(6).Text())
			// dailyLowStr := strings.TrimSpace(cells.Eq(7).Text())
			cleanLivePrice := strings.Replace(livePriceStr, ",", "", -1)
			// cleanDailyhigh := strings.Replace(dailyHighStr, ",", "", -1)
			// cleanDailyLow := strings.Replace(dailyLowStr, ",", "", -1)

			livePrice, err := strconv.ParseFloat(cleanLivePrice, 64)
			if err != nil {
				log.Printf("Error parsing live price for %s: %v", symbol, err)
				return
			}
			// dailyHigh, err := strconv.ParseFloat(cleanDailyhigh, 64)
			// if err != nil {
			// 	log.Printf("Error parsing daily high price for %s: %v", symbol, err)
			// 	return
			// }
			// dailyLow, err := strconv.ParseFloat(cleanDailyLow, 64)
			// if err != nil {
			// 	log.Printf("Error parsing daily low price for %s: %v", symbol, err)
			// 	return
			// }
			ticker, exists := tickers[symbol]
			if exists {
				ticker.Update(livePrice, 0, 0)
			} else {
				t := NewTicker(symbol, name, "crypto", livePrice, 0, 0)
				tickers[symbol] = t
			}
		} else {
			log.Println("Scrapper does not have sufficeint table columns.")
		}
	})
	return nil
}

func scrapUrlMetals(url string) error {
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
		if cells.Length() >= 3 {

			symbol := strings.TrimSpace(cells.Eq(0).Find("a").Eq(0).Text())
			name := strings.TrimSpace(cells.Eq(0).Find("sup").Eq(0).Text())
			livePriceStr := strings.TrimSpace(cells.Eq(1).Text())

			dailyHighStr := strings.TrimSpace(cells.Eq(4).Text())
			dailyLowStr := strings.TrimSpace(cells.Eq(5).Text())
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
			fmt.Println(symbol, name, livePrice, dailyHigh, dailyLow)
			ticker, exists := tickers[symbol]
			if exists {
				ticker.Update(livePrice, 0, 0)
			} else {
				t := NewTicker(symbol, name, "crypto", livePrice, 0, 0)
				tickers[symbol] = t
			}
		} else {
			log.Println("Scrapper does not have sufficeint table columns.")
		}
	})
	return nil
}
