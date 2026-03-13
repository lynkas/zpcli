package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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

func Load() (*StoreData, error) {
	data := &StoreData{
		Series: make([]*Series, 0),
	}

	dataFile, err := getConfigFile()
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(dataFile)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
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

	if data.Series == nil {
		data.Series = make([]*Series, 0)
	}

	if data.Version == 0 {
		data.Version = 1
	}

	return data, nil
}

func (s *StoreData) Save() error {
	dataFile, err := getConfigFile()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, bytes, 0644)
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
