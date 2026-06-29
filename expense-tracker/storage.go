package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"
)

func load() error {
	file, err := os.ReadFile("expenses.json")
	if errors.Is(err, os.ErrNotExist) {
		expenses = []Expense{}
		return nil
	}
	if err != nil {
		return err
	}

	if len(file) == 0 {
		expenses = []Expense{}
		return nil
	}

	return json.Unmarshal(file, &expenses)
}

func save() error {
	file, err := os.OpenFile("expenses.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(expenses, "", " ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func export() error {
	file, err := os.Create("expenses.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)

	if err := w.Write([]string{"Id", "Date", "Description", "Amount", "Category"}); err != nil {
		return err
	}

	for _, exp := range expenses {
		record := []string{
			strconv.Itoa(exp.ID),
			exp.Date.Format(time.DateOnly),
			exp.Description,
			strconv.FormatFloat(exp.Amount, 'f', 2, 64),
		}

		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()

	return w.Error()
}
