package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const CurrentVersion = 1

type Domain struct {
	URL          string `json:"url"`
	FailureScore int    `json:"failure_score"`
}

type Series struct {
	Domains []*Domain `json:"domains"`
}

type StoreData struct {
	Version int       `json:"version"`
	Series  []*Series `json:"series"`
}

func getConfigFile() (string, error) {
	if envPath := os.Getenv("ZPCLI_CONFIG"); envPath != "" {
		dir := filepath.Dir(envPath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "", err
		}
		return envPath, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "zpcli")
	if err := os.MkdirAll(appDir, os.ModePerm); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "sites.json"), nil
}

func ConfigFilePath() (string, error) {
	return getConfigFile()
}

func Load() (*StoreData, error) {
	data := &StoreData{
		Version: CurrentVersion,
		Series: make([]*Series, 0),
	}

	dataFile, err := getConfigFile()
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Ensure file exists even if empty
			if err := data.Save(); err != nil {
				return nil, err
			}
			return data, nil
		}
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("config path %s is a directory, not a file. Please check your volume mounts", dataFile)
	}

	if fileInfo.Size() == 0 {
		return data, nil
	}

	bytes, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}
	if err := normalizeStoreData(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *StoreData) Save() error {
	if err := normalizeStoreData(s); err != nil {
		return err
	}
	if err := validateStoreData(s); err != nil {
		return err
	}

	dataFile, err := getConfigFile()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(dataFile)
	tempFile, err := os.CreateTemp(dir, "sites-*.json.tmp")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(bytes); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Chmod(0644); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, dataFile); err != nil {
		return err
	}

	cleanupTemp = false
	return nil
}

func (s *StoreData) CreateSeries(url string) error {
	for _, series := range s.Series {
		for _, dom := range series.Domains {
			if dom.URL == url {
				return fmt.Errorf("domain %s already exists", url)
			}
		}
	}
	s.Series = append(s.Series, &Series{
		Domains: []*Domain{{URL: url, FailureScore: 0}},
	})
	return s.Save()
}

func (s *StoreData) AddDomainToSeries(seriesIndex int, url string) error {
	if seriesIndex < 0 || seriesIndex >= len(s.Series) {
		return fmt.Errorf("series %d does not exist", seriesIndex+1)
	}

	for _, s2 := range s.Series {
		for _, dom := range s2.Domains {
			if dom.URL == url {
				return fmt.Errorf("domain %s already exists", url)
			}
		}
	}

	s.Series[seriesIndex].Domains = append(s.Series[seriesIndex].Domains, &Domain{
		URL:          url,
		FailureScore: 0,
	})

	return s.Save()
}

func (s *StoreData) RemoveDomain(domainStr string) error {
	parts := strings.Split(domainStr, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid formatting, expected format: seriesId.domainId (e.g. 1.1)")
	}
	seriesId, _ := strconv.Atoi(parts[0])
	domainId, _ := strconv.Atoi(parts[1])

	seriesIndex := seriesId - 1
	domainIndex := domainId - 1

	if seriesIndex < 0 || seriesIndex >= len(s.Series) {
		return fmt.Errorf("series %d does not exist", seriesId)
	}

	series := s.Series[seriesIndex]
	if domainIndex < 0 || domainIndex >= len(series.Domains) {
		return fmt.Errorf("domain %s not found", domainStr)
	}

	series.Domains = append(series.Domains[:domainIndex], series.Domains[domainIndex+1:]...)
	if len(series.Domains) == 0 {
		s.Series = append(s.Series[:seriesIndex], s.Series[seriesIndex+1:]...)
	}

	return s.Save()
}

func (s *StoreData) RemoveSeries(seriesIndex int) error {
	if seriesIndex < 0 || seriesIndex >= len(s.Series) {
		return fmt.Errorf("series %d does not exist", seriesIndex+1)
	}
	s.Series = append(s.Series[:seriesIndex], s.Series[seriesIndex+1:]...)
	return s.Save()
}

func normalizeStoreData(data *StoreData) error {
	if data == nil {
		return fmt.Errorf("store data is nil")
	}
	if data.Version == 0 {
		data.Version = CurrentVersion
	}
	if data.Version > CurrentVersion {
		return fmt.Errorf("unsupported config version %d", data.Version)
	}
	if data.Series == nil {
		data.Series = make([]*Series, 0)
	}
	for i, series := range data.Series {
		if series == nil {
			data.Series[i] = &Series{Domains: make([]*Domain, 0)}
			continue
		}
		if series.Domains == nil {
			series.Domains = make([]*Domain, 0)
		}
	}
	return nil
}

func validateStoreData(data *StoreData) error {
	if data == nil {
		return fmt.Errorf("store data is nil")
	}

	seen := make(map[string]string)
	for i, series := range data.Series {
		if series == nil {
			return fmt.Errorf("series %d is nil", i+1)
		}
		for j, dom := range series.Domains {
			if dom == nil {
				return fmt.Errorf("domain %d.%d is nil", i+1, j+1)
			}
			url := strings.TrimSpace(dom.URL)
			if url == "" {
				return fmt.Errorf("domain %d.%d has an empty URL", i+1, j+1)
			}
			if previous, exists := seen[url]; exists {
				return fmt.Errorf("domain %d.%d duplicates %s", i+1, j+1, previous)
			}
			seen[url] = fmt.Sprintf("%d.%d", i+1, j+1)
		}
	}
	return nil
}
