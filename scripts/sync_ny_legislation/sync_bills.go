package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jehiah/nysenateapi"
	log "github.com/sirupsen/logrus"
)

func fileName(bill nysenateapi.Bill) string {
	dirName := "bills"
	if bill.Resolution {
		dirName = "resolutions"
	}
	return filepath.Join(dirName, fmt.Sprintf("%d", bill.Session), bill.PrintNo+".json")
}

func (s *SyncApp) SyncBills() error {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	updates, err := s.api.GetBillUpdates(ctx, s.LastSync.Bills, now)
	if err != nil {
		return err
	}
	for _, billID := range updates.Bills {
		// get bill
		bill, err := s.api.GetBill(ctx, fmt.Sprintf("%d", billID.Session), billID.PrintNo)
		if err != nil {
			return err
		}
		err = s.writeFile(fileName(*bill), bill)
		if err != nil {
			return err
		}
	}
	s.LastSync.Bills = now
	return nil
}

func (s *SyncApp) LoadBills() error {
	files, err := filepath.Glob(filepath.Join(s.targetDir, "bills", "*", "*.json"))
	if err != nil {
		return err
	}
	for _, fn := range files {
		if filepath.Base(fn) == "index.json" {
			continue
		}
		if strings.Contains(fn, "_raw") {
			continue
		}
		fn = strings.TrimPrefix(fn, s.targetDir+"/")
		s.billLookup[fn] = true
	}

	log.Printf("loaded %d bills", len(s.billLookup))
	return nil
}

// UpdateAllBills
func (s *SyncApp) UpdateAllBills(ctx context.Context) error {
	year := "2017"
	// get all bills for the year in batches of 1000

	offset := 0
	offset = 16000
	for {
		res, err := s.api.Bills(ctx, year, offset)
		if err != nil {
			return err
		}
		log.Infof("got %d bills", len(res.Bills))
		offset += len(res.Bills)
		if len(res.Bills) == 0 {
			break
		}
		for _, billID := range res.Bills {
			// get bill
			bill, err := s.api.GetBill(ctx, fmt.Sprintf("%d", billID.Session), billID.PrintNo)
			if err != nil {
				return err
			}
			err = s.writeFile(fileName(*bill), bill)
			if err != nil {
				return err
			}

		}
	}
	return nil
}
