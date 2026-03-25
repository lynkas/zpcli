package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveRejectsDuplicateDomains(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sites.json")
	t.Setenv("ZPCLI_CONFIG", configPath)

	data := &StoreData{
		Version: CurrentVersion,
		Series: []*Series{
			{
				Domains: []*Domain{
					{URL: "example.com", FailureScore: 0},
				},
			},
			{
				Domains: []*Domain{
					{URL: "example.com", FailureScore: 1},
				},
			},
		},
	}

	err := data.Save()
	if err == nil {
		t.Fatal("expected duplicate-domain save to fail")
	}
	if !strings.Contains(err.Error(), "duplicates") {
		t.Fatalf("expected duplicate-domain error, got %v", err)
	}
}

func TestSaveWritesConfigAtomically(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sites.json")
	t.Setenv("ZPCLI_CONFIG", configPath)

	data := &StoreData{
		Version: CurrentVersion,
		Series: []*Series{
			{
				Domains: []*Domain{
					{URL: "example.com", FailureScore: 0},
				},
			},
		},
	}

	if err := data.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(bytes), "\"version\": 1") {
		t.Fatalf("expected saved config to include version, got %s", string(bytes))
	}

	matches, err := filepath.Glob(filepath.Join(filepath.Dir(configPath), "sites-*.json.tmp"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no temp files after save, found %v", matches)
	}
}

func TestLoadNormalizesLegacyData(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sites.json")
	t.Setenv("ZPCLI_CONFIG", configPath)

	content := `{"series":[{"domains":[{"url":"example.com","failure_score":0}]}]}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if data.Version != CurrentVersion {
		t.Fatalf("expected version %d, got %d", CurrentVersion, data.Version)
	}
	if len(data.Series) != 1 || len(data.Series[0].Domains) != 1 {
		t.Fatalf("expected normalized series/domain data, got %+v", data)
	}
}
