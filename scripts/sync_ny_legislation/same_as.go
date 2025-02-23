package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jehiah/nysenateapi"
	log "github.com/sirupsen/logrus"
)

type SameAs struct {
	SameAsPrintNo    string   `json:",omitempty"`
	PreviousVersions []string `json:",omitempty"`
}

func (s *SyncApp) AddSameAs(bill nysenateapi.Bill) {
	printNo := fmt.Sprintf("%s-%d", bill.PrintNo, bill.Session)
	s.billIndex[printNo] = SameAs{
		SameAsPrintNo:    bill.SameAsPrintNo,
		PreviousVersions: bill.PreviousVersions,
	}
}

func (s *SyncApp) LoadBillIndex() error {
	s.billIndex = make(map[string]SameAs, 10000)
	files, err := filepath.Glob(filepath.Join(s.targetDir, "bills", "*-index.json"))
	if err != nil {
		return err
	}
	for _, fn := range files {
		fn = strings.TrimPrefix(fn, s.targetDir+"/")
		var index map[string]SameAs
		if err = s.readFile(fn, &index); err != nil {
			return err
		}
		for k, v := range index {
			s.billIndex[k] = v
		}
	}

	log.Printf("loaded %d index", len(s.billIndex))
	return nil
}

func comparePrintNo(a, b string) bool {
	// split session
	printA, sessionA, _ := strings.Cut(a, "-")
	printB, sessionB, _ := strings.Cut(b, "-")
	if sessionA != sessionB {
		return sessionA < sessionB
	}
	// compare first digit as a string
	if printA[0] != printB[0] {
		return printA[0] < printB[0]
	}
	// compare the rest as integers
	n := func(s string) int {
		i, _ := strconv.Atoi(s)
		return i
	}
	return n(printA[1:]) < n(printB[1:])
}

func (s *SyncApp) SaveBillIndex() error {
	// group by year
	sessions := make(map[string][]string)
	for k := range s.billIndex {
		_, session, _ := strings.Cut(k, "-")
		sessions[session] = append(sessions[session], k)
	}

	for session, bills := range sessions {
		sort.Slice(bills, func(i, j int) bool {
			return comparePrintNo(bills[i], bills[j])
		})
		var o bytes.Buffer
		o.Write([]byte("{\n"))
		for i, printNo := range bills {
			suffix := ",\n"
			if i == len(bills)-1 {
				suffix = "\n"
			}
			row, _ := json.Marshal(s.billIndex[printNo])
			o.WriteString(fmt.Sprintf("%q: %s%s", printNo, row, suffix))
		}
		o.Write([]byte("}\n"))
		err := s.writeFile(filepath.Join("bills", session+"-index.json"), o.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}
