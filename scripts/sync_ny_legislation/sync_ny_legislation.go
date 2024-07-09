package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/jehiah/nysenateapi"
	log "github.com/sirupsen/logrus"
)

var localTimezone *time.Location

type SyncApp struct {
	api       nysenateapi.API
	targetDir string

	billLookup map[string]bool

	LastSync
}

type LastSync struct {
	Bills time.Time

	LastRun time.Time
}

func (s *SyncApp) Load() error {
	fn := filepath.Join(s.targetDir, "last_sync.json")
	_, err := os.Stat(fn)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	b, err := os.ReadFile(fn)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &s.LastSync)
	if err != nil {
		return err
	}
	err = s.LoadBills()
	if err != nil {
		return err
	}
	return nil
}

func (s *SyncApp) Run() error {
	os.MkdirAll(s.targetDir, 0777)
	os.MkdirAll(filepath.Join(s.targetDir, "bills"), 0777)
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

	e := json.NewEncoder(f)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	err = e.Encode(o)
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
	targetDir := flag.String("target-dir", "../../", "Target Directory")
	timezone := flag.String("tz", "America/New_York", "timezone")
	updateAll := flag.Bool("update-all", false, "update all")
	skipIndexUpdate := flag.Bool("skip-index-update", false, "skip updating year index files and last_sync.json")

	flag.Parse()
	log.SetLevel(log.DebugLevel)
	if *targetDir == "" {
		log.Fatal("set --target-dir")
	}

	s := &SyncApp{
		api:        nysenateapi.NewAPI(os.Getenv("NY_SENATE_TOKEN")),
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
	case *updateAll:
		err = s.UpdateAllBills(ctx)
	default:
		err = s.Run()
	}
	if err != nil {
		log.Fatal(err)
	}
	if !*skipIndexUpdate {
		if err := s.Save(); err != nil {
			log.Fatal(err)
		}
	}
}
