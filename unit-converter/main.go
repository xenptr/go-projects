package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	UnitTypeLength      = "length"
	UnitTypeWeight      = "weight"
	UnitTypeTemperature = "temperature"
)

type PageData struct {
	Value  float64
	From   string
	To     string
	Result float64
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

var templates = template.Must(
	template.New("").
		Funcs(template.FuncMap{
			"format": formatFloat,
		}).
		ParseFiles(
			"tmpl/"+UnitTypeLength+".html",
			"tmpl/"+UnitTypeWeight+".html",
			"tmpl/"+UnitTypeTemperature+".html",
		),
)

func renderTemplate(w http.ResponseWriter, tmpl string, data *PageData) {
	if err := templates.ExecuteTemplate(w, tmpl+".html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/"+UnitTypeLength, http.StatusFound)
}

func unitHandler(w http.ResponseWriter, r *http.Request) {
	unitType := strings.TrimPrefix(r.URL.Path, "/")
	switch unitType {
	case UnitTypeLength, UnitTypeWeight, UnitTypeTemperature:
		// valid
	default:
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		renderTemplate(w, unitType, nil)
	case http.MethodPost:
		processConversion(w, r, unitType)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func processConversion(w http.ResponseWriter, r *http.Request, unitType string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	valueStr := r.FormValue("value")
	from := r.FormValue("from")
	to := r.FormValue("to")

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	result, err := convert(unitType, value, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data := &PageData{
		Value:  value,
		From:   from,
		To:     to,
		Result: result,
	}

	renderTemplate(w, unitType, data)
}

func convert(unitType string, value float64, from, to string) (float64, error) {
	if from == to {
		return value, nil
	}
	switch unitType {
	case UnitTypeLength:
		return convertLength(value, from, to), nil
	case UnitTypeWeight:
		return convertWeight(value, from, to), nil
	case UnitTypeTemperature:
		return convertTemperature(value, from, to), nil
	default:
		return 0, fmt.Errorf("unknown unit type: %q", unitType)
	}
}

func convertLength(value float64, from, to string) float64 {
	meters := toMeters(value, from)
	return fromMeters(meters, to)
}

func toMeters(value float64, unit string) float64 {
	switch unit {
	case "mm":
		return value / 1000
	case "cm":
		return value / 100
	case "m":
		return value
	case "km":
		return value * 1000
	case "in":
		return value * 0.0254
	case "ft":
		return value * 0.3048
	case "yd":
		return value * 0.9144
	case "mi":
		return value * 1609.344
	}
	return value
}

func fromMeters(value float64, unit string) float64 {
	switch unit {
	case "mm":
		return value * 1000
	case "cm":
		return value * 100
	case "m":
		return value
	case "km":
		return value / 1000
	case "in":
		return value / 0.0254
	case "ft":
		return value / 0.3048
	case "yd":
		return value / 0.9144
	case "mi":
		return value / 1609.344
	}
	return value
}

func convertWeight(value float64, from, to string) float64 {
	grams := toGrams(value, from)
	return fromGrams(grams, to)
}

func toGrams(value float64, unit string) float64 {
	switch unit {
	case "mg":
		return value / 1000
	case "g":
		return value
	case "kg":
		return value * 1000
	case "oz":
		return value * 28.3495
	case "lb":
		return value * 453.592
	}
	return value
}

func fromGrams(value float64, unit string) float64 {
	switch unit {
	case "mg":
		return value * 1000
	case "g":
		return value
	case "kg":
		return value / 1000
	case "oz":
		return value / 28.3495
	case "lb":
		return value / 453.592
	}
	return value
}

func convertTemperature(value float64, from, to string) float64 {
	celsius := toCelsius(value, from)
	return fromCelsius(celsius, to)
}

func toCelsius(value float64, unit string) float64 {
	switch unit {
	case "c":
		return value
	case "f":
		return (value - 32) * 5 / 9
	case "k":
		return value - 273.15
	}
	return value
}

func fromCelsius(value float64, unit string) float64 {
	switch unit {
	case "c":
		return value
	case "f":
		return (value * 1.8) + 32
	case "k":
		return value + 273.15
	}
	return value
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/"+UnitTypeLength, unitHandler)
	http.HandleFunc("/"+UnitTypeWeight, unitHandler)
	http.HandleFunc("/"+UnitTypeTemperature, unitHandler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
