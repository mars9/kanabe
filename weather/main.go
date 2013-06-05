package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"text/template"
)

const ApiUrl = "http://api.wunderground.com/api/"

var (
	forecast   = flag.Bool("f", false, "reports the current (3-day) forecast")
	forecast10 = flag.Bool("f10", false, "reports the current (10-day) forecast")

	apiKey  string
	station string = "Innsbruck, AT"
)

type Response struct {
	Error Error
}

type Error struct {
	Type        string
	Description string
}

type Conditions struct {
	Current_observation Current
	Response            Response
}

type Current struct {
	Observation_time   string
	Display_location   Location
	Station_id         string
	Weather            string
	Temperature_string string
	Wind_string        string
	Feelslike_string   string
	Relative_humidity  string
}

type Location struct {
	Latitude  string
	Longitude string
	Elevation string
}

var ConditionsTmpl string = `{{if .Response.Error.Type}}Error: {{.Response.Error.Description}}{{else}}
Conditions at ` + station + `:
  Latitude:   {{.Current_observation.Display_location.Latitude}}
  Longitude:  {{.Current_observation.Display_location.Longitude}}
  Elevation:  {{.Current_observation.Display_location.Elevation}}m
  Station ID: {{.Current_observation.Station_id}}
  {{.Current_observation.Observation_time}} 

  Conditions:        {{.Current_observation.Weather}}
  Temperatur:        {{.Current_observation.Temperature_string}}
  Feels Like:        {{.Current_observation.Feelslike_string}}
  Wind:              {{.Current_observation.Wind_string}}
  Relative humidity: {{.Current_observation.Relative_humidity}}
{{end}}
`

type ForecastConditions struct {
	Forecast Forecast
	Response Response
}

type Forecast struct {
	Txt_forecast Txt_forecast
}

type Txt_forecast struct {
	Forecastday []Forecastday
}

type Forecastday struct {
	Title          string
	Fcttext_metric string
}

var ForecastTmpl string = `{{if .Response.Error.Type}}Error: {{.Response.Error.Description}}{{else}}
Forecast for ` + station + `:
{{range $i, $e := .Forecast.Txt_forecast.Forecastday}}  {{$e.Title}}:
    {{$e.Fcttext_metric}}

{{end}}{{end}}
`

func fetch(query string) (b []byte, err error) {
	r, err := http.Get(ApiUrl + apiKey + "/" + query + "/q/" + station + ".json")
	if err != nil {
		return
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("fetch: %s: %s", station, r.Status)
	}
	return ioutil.ReadAll(r.Body)
}

func formatter(buf *bytes.Buffer) (err error) {
	cmd := exec.Command("fmt")
	in, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}
	if _, err = in.Write(buf.Bytes()); err != nil {
		return
	}
	in.Close()
	formated, err := ioutil.ReadAll(out)
	if err != nil {
		return
	}
	if err = cmd.Wait(); err != nil {
		return
	}
	fmt.Printf("%s", formated)
	return
}

type cmd struct {
	result   interface{}
	query    string
	template string
}

func (c cmd) weather(format bool) (err error) {
	b, err := fetch(c.query)
	if err != nil {
		return
	}
	if err = json.Unmarshal(b, c.result); err != nil {
		return
	}
	tmpl, err := template.New(c.query).Parse(c.template)
	if err != nil {
		return
	}
	if format {
		buf := bytes.NewBuffer(nil)
		if err = tmpl.Execute(buf, c.result); err != nil {
			return
		}
		if err = formatter(buf); err != nil {
			return
		}
	} else {
		if err = tmpl.Execute(os.Stdout, c.result); err != nil {
			return
		}
	}
	return
}

func init() {
	if apiKey = os.Getenv("WEATHERAPIKEY"); apiKey == "" {
		fmt.Fprintf(os.Stderr, "%s: WEATHERAPIKEY not set\n", os.Args[0])
		os.Exit(2)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] station\n", os.Args[0])
		fmt.Fprint(os.Stderr, usageMsg)
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()

	if ws := os.Getenv("WEATHERSTATION"); ws != "" {
		station = ws
	}
	if flag.NArg() == 1 {
		station = flag.Arg(0)
	}

	if flag.NFlag() == 0 {
		c := cmd{
			result:   new(Conditions),
			query:    "conditions",
			template: ConditionsTmpl,
		}
		if err := c.weather(false); err != nil {
			fmt.Fprintf(os.Stderr, "weather: %v\n", err)
			os.Exit(1)
		}
	} else {
		var c cmd
		if *forecast {
			c = cmd{
				result:   new(ForecastConditions),
				query:    "forecast",
				template: ForecastTmpl,
			}
		} else if *forecast10 {
			c = cmd{
				result:   new(ForecastConditions),
				query:    "forecast10day",
				template: ForecastTmpl,
			}
		}
		if err := c.weather(true); err != nil {
			fmt.Fprintf(os.Stderr, "weather: %v\n", err)
			os.Exit(1)
		}
	}
}

const usageMsg = `
Weather prints the local conditions and forecast (3- and 10-day) most
recently reported at Underground Weather (http://www.wunderground.com)
with location identifier station. Where station is a "city,
state-abbreviation", (US or Canadian) zipcode, 3- or 4-letter airport
code, or "LAT, LONG".

The arguments are mutually exclusive and case-insensitive. If neither
is given, station defaults to location identifier Innsbruck, Austria

Environment:
  WEATHERSTATION - location identifier station
`
