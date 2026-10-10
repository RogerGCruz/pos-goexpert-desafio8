package main

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"tests_bdd/steps"
)

func TestMain(m *testing.M) {
	status := godog.TestSuite{
		Name:                 "bdd",
		ScenarioInitializer:  steps.InitializeScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{"features"},
		},
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
