# Unit Converter

A web-based unit converter built with Go's standard library. Supports length, weight, and temperature conversions through a simple browser interface.

## Project URL

https://roadmap.sh/projects/unit-converter

## Features

- Convert between units of length, weight, and temperature
- Clean web UI served entirely by the Go standard library (no external dependencies)
- HTML templates with a shared layout

### Supported Units

**Length:** Millimeter, Centimeter, Meter, Kilometer, Inch, Foot, Yard, Mile

**Weight:** Milligram, Gram, Kilogram, Ounce, Pound

**Temperature:** Celsius, Fahrenheit, Kelvin

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd unit-converter
```

Run the server:

```bash
go run .
```

Then open your browser at [http://localhost:8080](http://localhost:8080).

Or build an executable first:

```bash
go build -o unit-converter
./unit-converter
```

## Usage

Navigate to one of the three converter pages:

| URL | Converter |
|-----|-----------|
| `/length` | Length units |
| `/weight` | Weight units |
| `/temperature` | Temperature units |

The root path (`/`) redirects to `/length` by default.

Enter a value, pick the source and target units, and submit the form to see the result.

## Project Structure

```text
.
├── main.go             # HTTP server, route handlers, and conversion logic
├── tmpl/
│   ├── layout.html     # Shared page layout
│   └── converter.html  # Converter form and result template
├── go.mod
└── README.md
```
