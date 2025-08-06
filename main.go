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
	currencies CurrenciesData
)

type CurrenciesData map[string]string

type RatesData struct {
	Rates map[string]float64 `json:"rates"`
}

func print() {
	p := message.NewPrinter(language.English)

	if base == quote {
		fmt.Println("Whoa! You picked the same currency for conversion. No conversion needed.")
		p.Printf("%.2f %s equals %.2f %s\n", baseValue, currencies[base], baseValue, currencies[base])
		return
	}

	p.Printf("%.2f %s equals %.2f %s\n", baseValue, currencies[base], quoteValue, currencies[quote])
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

	var rd RatesData
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &rd)
	if err != nil {
		panic(err)
	}

	baseRate := 1.0
	if base != "USD" {
		baseRate = rd.Rates[base]
	}

	quoteRate := 1.0
	if quote != "USD" {
		quoteRate = rd.Rates[quote]
	}

	baseValue, err = strconv.ParseFloat(inputValue, 64)
	if err != nil {
		panic(err)
	}

	quoteValue = (quoteRate / baseRate) * baseValue
}

func getCurrencies() {
	f, err := os.ReadFile("currencies.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(f, &currencies)
	if err != nil {
		panic(err)
	}
}

func main() {
	getCurrencies()

	var options []huh.Option[string]
	for code, name := range currencies {
		display := fmt.Sprintf("%s (%s)", code, name)
		options = append(options, huh.NewOption(display, code))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Base Currency").
				Description("💡 Press / to search (e.g. 'dollar', 'eur', 'yen')").
				Options(options...).
				Height(12).
				Value(&base),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Target Currency").
				Description("💡 Press / to search").
				Options(options...).
				Height(12).
				Value(&quote),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("How much do you want to convert?").
				Placeholder("Enter amount (e.g. 100)").
				Value(&inputValue),
		))

	err := form.Run()
	if err != nil {
		panic(err)
	}

	convert()
	print()
}
