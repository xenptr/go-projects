package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type PageData struct {
	Category string
	Value    float64
	From     string
	To       string
	Result   float64
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
			"tmpl/length.html",
			"tmpl/weight.html",
			"tmpl/temperature.html",
		),
)

func renderTemplate(w http.ResponseWriter, tmpl string, data *PageData) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/length", http.StatusFound)
}

func lengthHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Parses only length.html into a new template set on every request.
		// t, err := template.ParseFiles("length.html")

		// Creates a new template set (named ""), registers custom functions,
		// then parses length.html into that template set.
		// t := template.Must(template.New("").Funcs(template.FuncMap{"format": formatFloat}).ParseFiles("length.html"))

		// Executes the already parsed template named "length.html"
		// from the global template set.
		// err := templates.ExecuteTemplate(w, "length.html", nil)

		// Wrapper around ExecuteTemplate that also handles errors.
		renderTemplate(w, "length", nil)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		category := r.FormValue("category")
		valueStr := r.FormValue("value")
		from := r.FormValue("from")
		to := r.FormValue("to")

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, "Invalid number", http.StatusBadRequest)
			return
		}

		result := convert(category, value, from, to)
		data := &PageData{
			Category: category,
			Value:    value,
			From:     from,
			To:       to,
			Result:   result,
		}

		renderTemplate(w, "length", data)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func weightHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		renderTemplate(w, "weight", nil)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		category := r.FormValue("category")
		valueStr := r.FormValue("value")
		from := r.FormValue("from")
		to := r.FormValue("to")

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, "Invalid number", http.StatusBadRequest)
			return
		}

		result := convert(category, value, from, to)
		data := &PageData{
			Category: category,
			Value:    value,
			From:     from,
			To:       to,
			Result:   result,
		}

		renderTemplate(w, "weight", data)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func temperatureHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		renderTemplate(w, "temperature", nil)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		category := r.FormValue("category")
		valueStr := r.FormValue("value")
		from := r.FormValue("from")
		to := r.FormValue("to")

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, "Invalid number", http.StatusBadRequest)
			return
		}

		result := convert(category, value, from, to)
		data := &PageData{
			Category: category,
			Value:    value,
			From:     from,
			To:       to,
			Result:   result,
		}

		renderTemplate(w, "temperature", data)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func convert(cat string, value float64, from, to string) float64 {
	if from == to {
		return value
	}

	switch cat {
	case "length":
		return convertLength(value, from, to)
	case "weight":
		return convertWeight(value, from, to)
	case "temperature":
		return convertTemperature(value, from, to)
	default:
		panic("unknown category")
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
	http.HandleFunc("/length", lengthHandler)
	http.HandleFunc("/weight", weightHandler)
	http.HandleFunc("/temperature", temperatureHandler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
