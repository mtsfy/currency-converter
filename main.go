package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	base       string
	baseValue  float64
	quote      string
	quoteValue float64
	inputValue string
)

type Rate map[string]float64

type Data struct {
	Rates map[string]float64 `json:"rates"`
}

func print() {
	p := message.NewPrinter(language.English)

	currencies := make(map[string]string)
	currencies["USD"] = "$"
	currencies["GBP"] = "£"
	currencies["EUR"] = "€"
	currencies["JPY"] = "¥"
	currencies["ETB"] = "Br"

	if base == quote {
		fmt.Println("Whoa! You picked the same currency for conversion. No conversion needed.")
		p.Printf("%s%.2f (%s) converts to %s%.2f (%s)\n", currencies[base], baseValue, base, currencies[quote], baseValue, quote)
		return
	}

	p.Printf("%s%.2f (%s) converts to %s%.2f (%s)\n", currencies[base], baseValue, base, currencies[quote], quoteValue, quote)
}

func convert() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Warning: Could not load .env file, using system environment")
	}

	appID := os.Getenv("APP_ID")
	if appID == "" {
		fmt.Println("Error: APP_ID not found in environment variables")
		fmt.Println("Please create a .env file with APP_ID=your_api_key")
		os.Exit(1)
	}

	res, err := http.Get(fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s", appID))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	var d Data
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &d)
	if err != nil {
		panic(err)
	}

	baseRate := 1.0
	if base != "USD" {
		baseRate = d.Rates[base]
	}

	quoteRate := 1.0
	if quote != "USD" {
		quoteRate = d.Rates[quote]
	}

	baseValue, err = strconv.ParseFloat(inputValue, 64)
	if err != nil {
		panic(err)
	}

	quoteValue = (quoteRate / baseRate) * baseValue
}

func main() {

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("What is your base currency?").Options(
				huh.NewOption("$ USD", "USD"),
				huh.NewOption("£ GBP", "GBP"),
				huh.NewOption("€ EUR", "EUR"),
				huh.NewOption("¥ JPY", "JPY"),
				huh.NewOption("Br BIRR", "ETB"),
			).Value(&base),
			huh.NewSelect[string]().Title("What do you want to convert to?").Options(
				huh.NewOption("$ USD", "USD"),
				huh.NewOption("£ GBP", "GBP"),
				huh.NewOption("€ EUR", "EUR"),
				huh.NewOption("¥ JPY", "JPY"),
				huh.NewOption("Br BIRR", "ETB"),
			).Value(&quote),
			huh.NewInput().Title("How much do you want to convert").Value(&inputValue),
		),
	)

	err := form.Run()
	if err != nil {
		panic(err)
	}

	convert()
	print()
}
