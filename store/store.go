package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"zpcli/internal/logx"
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
	logger := logx.Logger("store")
	if envPath := os.Getenv("ZPCLI_CONFIG"); envPath != "" {
		logger.Info("resolve config path from env", "env", "ZPCLI_CONFIG", "path", envPath)
		dir := filepath.Dir(envPath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Error("create config dir from env failed", "dir", dir, "error", err)
			return "", err
		}
		return envPath, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		logger.Error("resolve user config dir failed", "error", err)
		return "", err
	}
	appDir := filepath.Join(configDir, "zpcli")
	if err := os.MkdirAll(appDir, os.ModePerm); err != nil {
		logger.Error("create app config dir failed", "dir", appDir, "error", err)
		return "", err
	}
	logger.Info("resolve config path", "path", filepath.Join(appDir, "sites.json"))
	return filepath.Join(appDir, "sites.json"), nil
}

func ConfigFilePath() (string, error) {
	return getConfigFile()
}

func Load() (*StoreData, error) {
	logger := logx.Logger("store")
	logger.Info("load store start")
	data := &StoreData{
		Version: CurrentVersion,
		Series:  make([]*Series, 0),
	}

	dataFile, err := getConfigFile()
	if err != nil {
		logger.Error("load store resolve config failed", "error", err)
		return nil, err
	}
	logger.Info("load store file", "path", dataFile)

	fileInfo, err := os.Stat(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("store file not found, creating default", "path", dataFile)
			// Ensure file exists even if empty
			if err := data.Save(); err != nil {
				logger.Error("save default store failed", "path", dataFile, "error", err)
				return nil, err
			}
			logger.Info("load store complete", "path", dataFile, "version", data.Version, "series_count", len(data.Series))
			return data, nil
		}
		logger.Error("stat store file failed", "path", dataFile, "error", err)
		return nil, err
	}

	if fileInfo.IsDir() {
		logger.Error("config path is directory", "path", dataFile)
		return nil, fmt.Errorf("config path %s is a directory, not a file. Please check your volume mounts", dataFile)
	}

	if fileInfo.Size() == 0 {
		logger.Info("store file empty", "path", dataFile)
		return data, nil
	}

	bytes, err := os.ReadFile(dataFile)
	if err != nil {
		logger.Error("read store file failed", "path", dataFile, "error", err)
		return nil, err
	}
	logger.Debug("store raw bytes loaded", "path", dataFile, "size", len(bytes), "content", string(bytes))

	err = json.Unmarshal(bytes, data)
	if err != nil {
		logger.Error("decode store file failed", "path", dataFile, "error", err)
		return nil, err
	}
	if err := normalizeStoreData(data); err != nil {
		logger.Error("normalize store failed", "path", dataFile, "error", err)
		return nil, err
	}
	logger.Info("load store complete", "path", dataFile, "version", data.Version, "series_count", len(data.Series))
	logger.Debug("load store result", "data", data)
	return data, nil
}

func (s *StoreData) Save() error {
	logger := logx.Logger("store")
	logger.Info("save store start", "data", s)
	if err := normalizeStoreData(s); err != nil {
		logger.Error("normalize before save failed", "error", err)
		return err
	}
	if err := validateStoreData(s); err != nil {
		logger.Error("validate before save failed", "error", err)
		return err
	}

	dataFile, err := getConfigFile()
	if err != nil {
		logger.Error("resolve config path for save failed", "error", err)
		return err
	}

	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		logger.Error("marshal store failed", "path", dataFile, "error", err)
		return err
	}
	logger.Debug("save store payload", "path", dataFile, "size", len(bytes), "content", string(bytes))

	dir := filepath.Dir(dataFile)
	tempFile, err := os.CreateTemp(dir, "sites-*.json.tmp")
	if err != nil {
		logger.Error("create temp store file failed", "dir", dir, "error", err)
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
		logger.Error("write temp store file failed", "path", tempPath, "error", err)
		return err
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		logger.Error("sync temp store file failed", "path", tempPath, "error", err)
		return err
	}
	if err := tempFile.Chmod(0644); err != nil {
		tempFile.Close()
		logger.Error("chmod temp store file failed", "path", tempPath, "error", err)
		return err
	}
	if err := tempFile.Close(); err != nil {
		logger.Error("close temp store file failed", "path", tempPath, "error", err)
		return err
	}
	if err := os.Rename(tempPath, dataFile); err != nil {
		logger.Error("rename temp store file failed", "from", tempPath, "to", dataFile, "error", err)
		return err
	}

	cleanupTemp = false
	logger.Info("save store complete", "path", dataFile, "size", len(bytes), "series_count", len(s.Series))
	return nil
}

func (s *StoreData) CreateSeries(url string) error {
	logger := logx.Logger("store")
	logger.Info("create series start", "url", url)
	for _, series := range s.Series {
		for _, dom := range series.Domains {
			if dom.URL == url {
				logger.Warn("create series duplicate domain", "url", url)
				return fmt.Errorf("domain %s already exists", url)
			}
		}
	}
	s.Series = append(s.Series, &Series{
		Domains: []*Domain{{URL: url, FailureScore: 0}},
	})
	logger.Debug("create series updated data", "data", s)
	return s.Save()
}

func (s *StoreData) AddDomainToSeries(seriesIndex int, url string) error {
	logger := logx.Logger("store")
	logger.Info("add domain to series start", "series_index", seriesIndex, "url", url)
	if seriesIndex < 0 || seriesIndex >= len(s.Series) {
		logger.Warn("add domain invalid series", "series_index", seriesIndex, "series_count", len(s.Series))
		return fmt.Errorf("series %d does not exist", seriesIndex+1)
	}

	for _, s2 := range s.Series {
		for _, dom := range s2.Domains {
			if dom.URL == url {
				logger.Warn("add domain duplicate url", "url", url)
				return fmt.Errorf("domain %s already exists", url)
			}
		}
	}

	s.Series[seriesIndex].Domains = append(s.Series[seriesIndex].Domains, &Domain{
		URL:          url,
		FailureScore: 0,
	})

	logger.Debug("add domain updated data", "data", s)
	return s.Save()
}

func (s *StoreData) RemoveDomain(domainStr string) error {
	logger := logx.Logger("store")
	logger.Info("remove domain start", "domain_id", domainStr)
	parts := strings.Split(domainStr, ".")
	if len(parts) != 2 {
		logger.Warn("remove domain invalid format", "domain_id", domainStr)
		return fmt.Errorf("invalid formatting, expected format: seriesId.domainId (e.g. 1.1)")
	}
	seriesId, _ := strconv.Atoi(parts[0])
	domainId, _ := strconv.Atoi(parts[1])

	seriesIndex := seriesId - 1
	domainIndex := domainId - 1

	if seriesIndex < 0 || seriesIndex >= len(s.Series) {
		logger.Warn("remove domain invalid series", "domain_id", domainStr, "series_id", seriesId, "series_count", len(s.Series))
		return fmt.Errorf("series %d does not exist", seriesId)
	}

	series := s.Series[seriesIndex]
	if domainIndex < 0 || domainIndex >= len(series.Domains) {
		logger.Warn("remove domain not found", "domain_id", domainStr)
		return fmt.Errorf("domain %s not found", domainStr)
	}

	series.Domains = append(series.Domains[:domainIndex], series.Domains[domainIndex+1:]...)
	if len(series.Domains) == 0 {
		s.Series = append(s.Series[:seriesIndex], s.Series[seriesIndex+1:]...)
	}

	logger.Debug("remove domain updated data", "data", s)
	return s.Save()
}

func (s *StoreData) RemoveSeries(seriesIndex int) error {
	logger := logx.Logger("store")
	logger.Info("remove series start", "series_index", seriesIndex)
	if seriesIndex < 0 || seriesIndex >= len(s.Series) {
		logger.Warn("remove series invalid series", "series_index", seriesIndex, "series_count", len(s.Series))
		return fmt.Errorf("series %d does not exist", seriesIndex+1)
	}
	s.Series = append(s.Series[:seriesIndex], s.Series[seriesIndex+1:]...)
	logger.Debug("remove series updated data", "data", s)
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
