// This program pulls a list of trading information for a given symbol (i.e. PLTR) and displays a Daily Time Series
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

// This is the main container for the returned JSON from Alpha Vantage
type Response struct {
	MetaData MetaData `json:"Meta Data"`
	// The returned JSON uses the date as a key and a nested object with
	// key/value pairs for the actual for that date i.e.
	// "2020-11-11" { "1. open": 200.0, "2. high": 205.0} etc.
	// Because of this we create a hashmap that stores the date as a key
	// and the contained JSON as the value, which converted into a
	// TimeSeries object, mirroring the JSON layout.
	TimeSeries map[string]TimeSeries `json:"Time Series (Daily)"`
}

// This object houses the meta data from the request.
type MetaData struct {
	Info        string `json:"1. Information"`
	Symbol      string `json:"2. Symbol"`
	LastRefresh string `json:"3. Last Refreshed"`
	TimeZone    string `json:"5. Time Zone"`
}

// This object is what holds the daily values from the JSON
type TimeSeries struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"6. volume"`
}

// Quick helper function to convert the Date string to a time.Time
// object. I cannot stress this enough: when using a layout to define
// the formatting to be parsed, you HAVE TO use the reference time
// for the formatting to work as it is in the docs.
// 01/02 03:04:05PM '06 -0700
func makeDate(value string) time.Time {
	layout := "2006-01-02"
	t, err := time.Parse(layout, value)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func main() {
	// The symbol to retrieve data for
	symbol := "PLTR"
	// A HTTP GET request to the API that concatenates in the symbol part of the URL
	response, err := http.Get("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=" + symbol + "&apikey=T0JGQ9XTGLNNY2U8")
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	// ioutil is used to parse the response body as []bytes
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate a Response object in preperation for unmarshalling the JSON
	var responseObject Response
	// Unmarshal/decode the JSON into the structs that we defined at the
	// start of the program
	json.Unmarshal(responseData, &responseObject)

	// Print out the MetaData from the response JSON
	fmt.Println("=============================")
	fmt.Println("Information:\t " + responseObject.MetaData.Info)
	fmt.Println("Symbol:\t\t " + responseObject.MetaData.Symbol)
	fmt.Println("Last Refresh:\t " + responseObject.MetaData.LastRefresh)
	fmt.Println("Timezone:\t " + responseObject.MetaData.TimeZone)

	// Because of the way we recieved the data (with the Date as the key and
	// the data consisting of nest JSON objects instead of lists) we converted
	// it to a hashmap using the date as the key. However, this means that the
	// data is unsorted, so we use this next block of code to convert the strings
	// to time.Time objects, which we are then able to sort.
	// We start by creating an empty array of time.Time objects.
	dateKeys := make([]time.Time, 0)
	// Then, for each entry in our TimeSeries hashmap, we append the keys (which
	// are the dates stored as strings) adter we use the makeDate helper function
	// to convert them to time.Time objects.
	for k, _ := range responseObject.TimeSeries {
		dateKeys = append(dateKeys, makeDate(k))
	}

	// Sorts the array of time.Time objects
	sort.Sort(ByDate(dateKeys))

	// Print out all of our data, iterating through each entry in the TimeSeries map.
	for _, value := range dateKeys {
		// Here we convert out time.Time object back to a string so it can be used
		// as a key for the hashmap. This is convoluted and there's probably a better
		// way of doing it but it works.
		value := value.Format("2006-01-02")
		fmt.Println("=============================")
		fmt.Println("Date:\t ", value)
		fmt.Println("Open:\t ", responseObject.TimeSeries[value].Open)
		fmt.Println("High:\t ", responseObject.TimeSeries[value].High)
		fmt.Println("Low:\t ", responseObject.TimeSeries[value].Low)
		fmt.Println("Close\t ", responseObject.TimeSeries[value].Close)
		fmt.Println("Volume\t ", responseObject.TimeSeries[value].Volume)
	}
	// Print a final line at the end for aesthetics.
	fmt.Println("=============================")

}

// Ngl idk why this works but I will at some point. Thanks to
// https://gist.github.com/santiaago/9027077 for making this.
type ByDate []time.Time

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Before(a[j]) }
