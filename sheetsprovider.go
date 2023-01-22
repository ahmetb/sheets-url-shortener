package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/api/sheets/v4"
)

type sheetsProvider struct {
	sync.RWMutex
	client         *sheets.Service
	googleSheetsID string
	sheetName      string
}

func (s *sheetsProvider) Init() error {
	if s.googleSheetsID == "" {
		return fmt.Errorf("GOOGLE_SHEET_ID not set")
	}

	srv, err := sheets.NewService(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}
	s.client = srv
	return nil
}

func (s *sheetsProvider) Query() ([][]interface{}, error) {
	log.Println("querying sheet")
	readRange := "A:D"
	if s.sheetName != "" {
		readRange = s.sheetName + "!" + readRange
	}
	resp, err := s.client.Spreadsheets.Values.Get(s.googleSheetsID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	log.Printf("queried %d rows", len(resp.Values))
	return resp.Values, nil
}

// Write will write the values rowwise, starting at the given column and row index.
func (s *sheetsProvider) Write(column string, rowIndex int, values []interface{}) error {
	s.Lock()
	defer s.Unlock()
	log.Printf("writing %s to row %v", values, rowIndex)
	writeRange := fmt.Sprintf("%s%d", column, rowIndex)
	if s.sheetName != "" {
		writeRange = s.sheetName + "!" + writeRange
	}
	_, err := s.client.Spreadsheets.Values.Update(s.googleSheetsID, writeRange, &sheets.ValueRange{
		Values: [][]interface{}{values},
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("unable to write data to sheet: %v", err)
	}
	return nil
}

func New() *sheetsProvider {
	return &sheetsProvider{}
}
