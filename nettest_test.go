package main

import (
	"testing"
)

func TestValidConfigParse(t *testing.T) {
	config, err := parseConfig("resources/validTestConfig.yaml")
	if err != nil {
		t.Fatalf("Configuration parsing failed. err: %v.", err)
	}

	//Set Expected and Actual values
	expectedConfigSize := 2
	actualConfigSize := len(config.Config)
	expectedNetworkName := "amazon"
	actualNetworkName := config.Config[1].NetworkName
	expectedPort := 443
	actualPort := config.Config[1].Port

	//Tests
	if actualConfigSize != expectedConfigSize {
		t.Fatalf("Expected %d config tests, found %d.", expectedConfigSize, actualConfigSize)
	}
	if actualNetworkName != expectedNetworkName {
		t.Fatalf("Expected network name %s, found %s.", expectedNetworkName, actualNetworkName)
	}
	if actualPort != expectedPort {
		t.Fatalf("Expected network name %s, found %s.", expectedPort, actualPort)
	}
}

func TestInvalidConfigParse(t *testing.T) {
	_, err := parseConfig("testResources/invalidTestConfig.yaml")
	if err == nil {
		t.Fatal("Config parsing returned a success for an invalid YAML file.")
	}
}
