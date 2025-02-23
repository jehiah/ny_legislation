package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jehiah/nysenateapi"
	"github.com/jehiah/nysenateapi/verboseapi"
	log "github.com/sirupsen/logrus"
	"github.com/swaggest/assertjson"
	"golang.org/x/time/rate"
)

var localTimezone *time.Location

type SyncApp struct {
	api       *nysenateapi.API
	targetDir string

	billLookup map[string]bool
	billIndex  map[string]SameAs

	LastSync
}

type LastSync struct {
	Bills time.Time

	LastRun time.Time
}

func (s *SyncApp) Load() error {
	err := s.readFile("last_sync.json", &s.LastSync)
	if err != nil {
		return err
	}
	err = s.LoadBills()
	if err != nil {
		return err
	}
	err = s.LoadBillIndex()
	if err != nil {
		return err
	}
	return nil
}

func (s *SyncApp) CustomAction(ctx context.Context) error {
	// read & re-write each file
	for f := range s.billLookup {
		var bill nysenateapi.Bill
		err := s.readFile(f, &bill)
		if err != nil {
			return err
		}
		s.AddSameAs(bill)
	}

	return nil
}

func (s *SyncApp) Run() error {
	os.MkdirAll(s.targetDir, 0777)
	os.MkdirAll(filepath.Join(s.targetDir, "bills"), 0777)
	os.MkdirAll(filepath.Join(s.targetDir, "resolutions"), 0777)
	s.LastRun = time.Now().UTC().Truncate(time.Second)
	err := s.SyncBills()
	if err != nil {
		return err
	}
	return nil
}

func (s SyncApp) openWriteFile(fn string) (*os.File, error) {
	fn = filepath.Join(s.targetDir, fn)
	err := os.MkdirAll(filepath.Dir(fn), 0777)
	if err != nil {
		return nil, err
	}
	log.Printf("creating %s", fn)
	return os.Create(fn)
}

func (s SyncApp) writeFile(fn string, o interface{}) error {
	f, err := s.openWriteFile(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	var b []byte
	if x, ok := o.([]byte); ok {
		b = x
	} else {
		b, err = assertjson.MarshalIndentCompact(o, "", " ", 120)
		if err != nil {
			return err
		}
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return f.Close()
}

func (s SyncApp) removeFile(fn string) error {
	fn = filepath.Join(s.targetDir, fn)
	log.Printf("removing %s", fn)
	err := os.Remove(fn)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s SyncApp) readFile(fn string, o interface{}) error {
	fn = filepath.Join(s.targetDir, fn)
	body, err := os.ReadFile(fn)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(body, &o)
}

func (s SyncApp) Save() error {
	return s.writeFile("last_sync.json", s.LastSync)
}

func main() {
	targetDir := flag.String("target-dir", "../..", "Target Directory")
	timezone := flag.String("tz", "America/New_York", "timezone")
	updateAll := flag.Bool("update-all", false, "update all")
	updateOne := flag.String("update-one", "", "update one S1234-2021")
	updateMultiple := flag.Bool("update-multiple", false, "update multiple (read on stdin)")
	skipIndexUpdate := flag.Bool("skip-index-update", false, "skip updating year index files and last_sync.json")
	customAction := flag.Bool("custom-action", false, "run custom action")

	flag.Parse()
	log.SetLevel(log.DebugLevel)
	if *targetDir == "" {
		log.Fatal("set --target-dir")
	}
	if *targetDir == "." {
		*targetDir = ""
	}

	// nyassembly.gov SSL has invalid chain
	// https://www.ssllabs.com/ssltest/analyze.html?d=nyassembly.gov
	t := http.DefaultTransport.(*http.Transport)
	t.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	vAPI := verboseapi.NewAPI(os.Getenv("NY_SENATE_TOKEN"))
	vAPI.Limiter = rate.NewLimiter(rate.Every(3*time.Millisecond), 25)

	s := &SyncApp{
		api:        nysenateapi.NewWithVerboseAPI(vAPI),
		billLookup: make(map[string]bool),
		targetDir:  *targetDir,
	}
	ctx := context.Background()

	var err error
	localTimezone, err = time.LoadLocation(*timezone)
	if err != nil {
		log.Fatal(err)
	}

	if err = s.Load(); err != nil {
		log.Fatal(err)
	}
	switch {
	case *customAction:
		err = s.CustomAction(ctx)
	case *updateAll:
		err = s.UpdateAllBills(ctx)
	case *updateOne != "":
		printNo, session, _ := strings.Cut(*updateOne, "-")
		err = s.UpdateOne(ctx, session, printNo)
	case *updateMultiple:
		// read lines on stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			printNo, session, _ := strings.Cut(line, "-")
			err = s.UpdateOne(ctx, session, printNo)
			if err != nil {
				log.Errorf("%s", err)
			}
		}
		err = scanner.Err()
	default:
		err = s.Run()
	}

	if err != nil {
		log.Fatal(err)
	}
	err = s.SaveBillIndex()
	if err != nil {
		log.Fatal(err)
	}
	if !*skipIndexUpdate {
		if err := s.Save(); err != nil {
			log.Fatal(err)
		}
	}
}
