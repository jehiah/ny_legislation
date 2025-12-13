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
	offset := 0
	for {
		updates, err := s.api.GetBillUpdates(ctx, s.LastSync.Bills, now, offset)
		if err != nil {
			return err
		}
		offset = updates.OffsetEnd + 1

		if len(updates.Bills) == 0 {
			break
		}

		for _, billID := range updates.Bills {
			// get bill
			bill, err := s.api.GetBill(ctx, fmt.Sprintf("%d", billID.Session), billID.PrintNo)
			if err != nil {
				return err
			}
			if !bill.Resolution {
				s.AddSameAs(*bill)
			}
			err = s.writeFile(fileName(*bill), bill)
			if err != nil {
				return err
			}
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
	year := "2025"
	// get all bills for the year in batches of 1000

	offset := 0
	for {
		res, err := s.api.Bills(ctx, year, offset)
		if err != nil {
			return err
		}
		log.Infof("year=%s got %d bills (query offset=%d)", year, len(res.Bills), offset)
		if len(res.Bills) == 0 {
			break
		}
		for _, billID := range res.Bills {
			offset++
			// every 100 bills - output offset
			if offset%100 == 0 {
				log.Infof("offset=%d", offset)
			}

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

// UpdateAllBills
func (s *SyncApp) UpdateOne(ctx context.Context, session, printNo string) error {
	// get bill
	bill, err := s.api.GetBill(ctx, session, printNo)
	if err != nil {
		return err
	}
	if bill == nil {
		return fmt.Errorf("bill not found %s-%s", printNo, session)
	}
	err = s.writeFile(fileName(*bill), bill)
	if err != nil {
		return err
	}

	return nil
}
