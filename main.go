package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shopspring/decimal"
)

// Define global variables

type config struct {
	address string
	apiKey  string
	coin    string
	chatID  string
	token   string
	name    string
}

func main() {
	// flags
	address := flag.String("address", "", "Address to monitor for changes")
	apiKey := flag.String("apikey", "", "APIKEY from https://chainz.cryptoid.info")
	coin := flag.String("coin", "", "Coin of address. You can lookup all supported coins at: https://chainz.cryptoid.info/ (use abbreviated form as you see in the url, e.g. Litecoin = ltc)")
	chatID := flag.String("chatid", "", "Look up the ChatID of your channel by inviting @RawDataBot to your channel (and removing it afterwards)")
	token := flag.String("token", "", "Telegram token for your bot, you need to create a bot by talking to @botfather first")
	name := flag.String("name", "", "(Nick)name of address, will appear in Telegram message")
	flag.Parse()

	c := config{
		address: *address,
		apiKey:  *apiKey,
		coin:    *coin,
		chatID:  *chatID,
		token:   *token,
		name:    *name,
	}

	switch {
	case c.address == "":
		log.Fatalln("Please specify an address using -address")
	case c.apiKey == "":
		log.Fatalln("Please specify an API key using -apikey")
	case c.chatID == "":
		log.Fatalln("Please specify a chat ID using -chatid")
	case c.coin == "":
		log.Fatalln("Please specify a coin using -coin")
	case c.token == "":
		log.Fatalln("Please specify a token using -token")
	case c.name == "":
		log.Fatalln("Please specify a name using -name")
	}

	// all checks completed starting script
	log.Println("Passed all checks, starting watcher now!")

	// setup scheduler
	tick := time.NewTicker(time.Minute)
	go scheduler(tick, c)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	tick.Stop()
}

func scheduler(tick *time.Ticker, c config) error {
	// variable to hold amound of last call
	var previousAmount float64

	// run once (directly on start)
	previousAmount = task(c, previousAmount)

	// loop every time interval
	for range tick.C {
		previousAmount = task(c, previousAmount)
	}
	return nil
}

func task(c config, previousAmount float64) float64 {
	b, err := checkAPI()
	if err != nil {
		log.Fatalln(err)
	}

	if b == true {
		amount, err := parseAmount(c)
		if err != nil {
			log.Fatalf("can't parse amount, error is: %s\n", err)
		}

		previousAmount, err = compareAmount(previousAmount, amount, c)
		if err != nil {
			log.Fatalf("can't compare amount, error is: %s\n", err)
		}
	} else {
		log.Println("API seems down. Skipping this round")
	}
	return previousAmount
}

func checkAPI() (bool, error) {
	url := "https://chainz.cryptoid.info/explorer/api.dws?q=summary"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("can't request %s, error is: %s", url, err)
	}
	request.Header.Set("Content-Type", "application/json")

	// request response
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return false, fmt.Errorf("HTTPS request to url %s failed with error %s", url, err)
	}
	// read response into data variable
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("can't read body %s, error is %s", response.Body, err)
	}

	if len(data) > 100 {
		// API available
		return true, nil
	}
	// API not available
	return false, nil
}

func parseAmount(c config) (float64, error) {
	// Define Variables
	var amount float64

	// assemble url
	url := "https://chainz.cryptoid.info/" + c.coin + "/api.dws?q=getbalance&a=" + c.address + "&key=" + c.apiKey

	// form url
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("can't request %s, error is: %s", url, err)
	}

	// add headers
	request.Header.Set("Content-Type", "application/json")

	// request response
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("HTTPS request to url %s failed with error %s", url, err)
	}
	// read response into data variable
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("can't read body %s, error is %s", response.Body, err)
	}

	// convert []byte to int
	amount, err = strconv.ParseFloat(string(data), 64)
	if err != nil {
		// can't parse string to number (this happens when website is in maintenance mode)
		return 0, nil
	}

	// output amount to console
	log.Printf("Current amount of %s: %f\n", strings.ToUpper(c.coin), amount)

	return amount, nil
}

// compareAmount compares the current amount a with the previous amount p
func compareAmount(p float64, a float64, c config) (f float64, err error) {
	if a == p {
		// current amount is equal to previous
	} else if a > p {
		// current amount is more than previous
		err = sendToTelegram(decimal.NewFromFloat(a), decimal.NewFromFloat(p), false, c)
	} else {
		// current amount is less than previous
		err = sendToTelegram(decimal.NewFromFloat(a), decimal.NewFromFloat(p), true, c)
	}
	if err != nil {
		return 0, err
	}
	return a, nil
}

// sendToTelegram sends amount a, and previous amount p, and bool less to telegram
func sendToTelegram(a decimal.Decimal, p decimal.Decimal, isLess bool, c config) error {
	// Variables
	url := "https://api.telegram.org/bot" + c.token + "/sendMessage?chat_id=" + c.chatID + "&text="
	var text string

	// make coin uppercase
	c.coin = strings.ToUpper(c.coin)

	// see if less bool is true or false, to send out messages for added or withdrawn coins
	if isLess {
		// the current amount is less than previous
		diff := p.Sub(a)
		text = fmt.Sprintf("%s %s have been withdrawn from %s, totalling now %s %s", diff.String(), c.coin, c.name, a.String(), c.coin)
	} else {
		// amount is more than previous amount
		diff := a.Sub(p)
		text = fmt.Sprintf("%s %s have been deposited to %s, totalling now %s %s", diff.String(), c.coin, c.name, a.String(), c.coin)
	}

	// construct url to be called
	url = url + text

	// call url
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("can't request %s, error is: %s", url, err)
	}

	// add headers
	request.Header.Set("Content-Type", "application/json")

	// request response
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("HTTPS request to url %s failed with error %s", url, err)
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("unexpected status code from telegram API, error is %d", response.StatusCode)
	}
	return nil
}
