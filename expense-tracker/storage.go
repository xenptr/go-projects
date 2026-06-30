package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"
)

const dataFile = "expenses.json"
const exportFile = "expenses.csv"

var store *Store

type Store struct {
	db Database
}

func (s *Store) Load() error {
	data, err := os.ReadFile(dataFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &s.db)
}

func (s *Store) Save() error {
	file, err := os.OpenFile(dataFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(s.db, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func (s *Store) AddExpense(exp Expense) error {
	s.db.Expenses = append(s.db.Expenses, exp)
	return s.Save()
}

func (s *Store) UpdateExpense(index int, description, category string, amount float64) error {
	if description != "" {
		s.db.Expenses[index].Description = description
	}
	if amount > 0 {
		s.db.Expenses[index].Amount = amount
	}
	if category != "" {
		s.db.Expenses[index].Category = category
	}
	s.db.Expenses[index].Date = time.Now()
	return s.Save()
}

func (s *Store) DeleteExpense(index int) error {
	s.db.Expenses = append(s.db.Expenses[:index], s.db.Expenses[index+1:]...)
	return s.Save()
}

func (s *Store) SetBudget(m time.Month, amount float64) error {
	year := time.Now().Year()
	for i := range s.db.Budgets {
		if s.db.Budgets[i].Year == year && s.db.Budgets[i].Month == m {
			s.db.Budgets[i].Amount = amount
			return s.Save()
		}
	}
	s.db.Budgets = append(s.db.Budgets, Budget{
		Year:   year,
		Month:  m,
		Amount: amount,
	})
	return s.Save()
}

func (s *Store) ExportCSV(expenses []Expense) error {
	file, err := os.Create(exportFile)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)

	if err := w.Write([]string{"ID", "Date", "Description", "Amount", "Category"}); err != nil {
		return err
	}

	for _, exp := range expenses {
		record := []string{
			strconv.Itoa(exp.ID),
			exp.Date.Format(time.DateOnly),
			exp.Description,
			strconv.FormatFloat(exp.Amount, 'f', 2, 64),
			exp.Category,
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
