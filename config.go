package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const httpConfReqTimeout int = 30

// TestConfig is the highest level struct in the test conig data structure.
// It contains basic information about the overall test case and a list of
// test configurations.
type TestConfig struct {
	TestName string
	Email    string
	Config   []Configuration
}

// Configuration is the unit defining a specific test. Instances of it are
// referenced from a parent TestConfig.
type Configuration struct {
	NetworkName string
	Host        string
	Path        string
	Port        int
	Proto       string
	Timeout     int
	CaptureBody bool
}

// ResponseDetails is the primary structure used to report on results.
// Specifically, its instances are used to generate an output report.
type ResponseDetails struct {
	Request          Configuration
	Success          bool
	Status           int
	FailureMessage   string
	Body             string
	Time             string
	IPResolvedStatus string
}

// Parse nettest config (yaml) and return a structured representation,
// TestConfig.
func parseConfig(fileLocation string) (TestConfig, error) {
	config := TestConfig{}
	var err error = nil

	if strings.HasPrefix(strings.ToLower(fileLocation), "http") {
		configData, err := downloadConfig(fileLocation)
		if err == nil {
			_ = yaml.Unmarshal(configData, &config)
		}
	} else {
		configData, err := ioutil.ReadFile(fileLocation)
		if err != nil {
			log.Printf("Failed to read configuration file. %s\n", err.Error())
		} else {
			_ = yaml.Unmarshal(configData, &config)
		}
	}

	if len(config.Config) < 1 {
		err = errors.New("Nettest config was invalid or didn't contain any test cases")
	}

	return config, err
}

func downloadConfig(fileLocation string) ([]byte, error) {
	req, err := http.NewRequest("GET", fileLocation, nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{Timeout: time.Duration(httpConfReqTimeout) * time.Second}
	resp, errClientReq := client.Do(req)
	if errClientReq != nil {
		log.Printf("Request to retrieve nettest config file failed. Error: %s", errClientReq.Error())
		return nil, errClientReq
	}
	defer resp.Body.Close()
	configResult, _ := ioutil.ReadAll(resp.Body)
	return configResult, nil
}

func (rd ResponseDetails) String() string {
	return fmt.Sprintf(
		"RequestURL: %s://%s:%d\nSuccess: %t\nStatusCode: %d\nTime: %s\nIPResolvedStatus: %s\nBody: %s\n",
		rd.Request.Proto, rd.Request.Host, rd.Request.Port,
		rd.Success, rd.Status, rd.Time, rd.IPResolvedStatus,
		rd.Body)
}
