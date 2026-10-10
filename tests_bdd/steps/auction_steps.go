package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"bytes"
	"os"
	"strconv"
	"time"
	"github.com/cucumber/godog"
)

type auctionOutput struct {
	Id          string `json:"id"`
	ProductName string `json:"product_name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Condition   int    `json:"condition"`
	Status      int    `json:"status"`
}

var (
	baseURL   = "http://localhost:8080"
	lastFound []auctionOutput
	currentId string
)

func theAppIsRunning() error {
	// simple check
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 20; i++ {
		resp, err := client.Get(baseURL + "/auction?status=0&productName=healthcheck")
		if err == nil && resp.StatusCode < 500 {
			if resp.Body != nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("app not reachable at %s", baseURL)
}

func iCreateAnAuction(productName, category, description string, condition int) error {
	body := map[string]interface{}{
		"product_name": productName,
		"category":     category,
		"description":  description,
		"condition":    condition,
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/auction", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected 201 Created, got %d: %s", resp.StatusCode, string(data))
	}
	return nil
}

func iSearchForActiveAuctionsWithProductName(productName string) error {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/auction?status=0&productName=" + productName)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
	var auctions []auctionOutput
	if err := json.NewDecoder(resp.Body).Decode(&auctions); err != nil {
		return err
	}
	lastFound = auctions
	return nil
}

func iShouldFindAtLeastOneAuction() error {
	if len(lastFound) == 0 {
		return fmt.Errorf("no auctions found in last search")
	}
	currentId = lastFound[0].Id
	return nil
}

func iWaitForTheAuctionDuration() error {
	// read env var
	secsStr := os.Getenv("AUCTION_DURATION")
	secs := 5
	if secsStr != "" {
		if v, err := strconv.Atoi(secsStr); err == nil && v > 0 {
			secs = v
		}
	}
	// wait a bit more than duration to allow background job to finish
	time.Sleep(time.Duration(secs+1) * time.Second)
	return nil
}

func iFetchTheAuctionById() error {
	if currentId == "" {
		return fmt.Errorf("no current auction id")
	}
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/auction/" + currentId)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK when fetching auction, got %d", resp.StatusCode)
	}
	var au auctionOutput
	if err := json.NewDecoder(resp.Body).Decode(&au); err != nil {
		return err
	}
	// store single-element list
	lastFound = []auctionOutput{au}
	return nil
}

func theAuctionStatusShouldBe(expected int) error {
	if len(lastFound) == 0 {
		return fmt.Errorf("no auction fetched")
	}
	if lastFound[0].Status != expected {
		return fmt.Errorf("expected status %d but got %d", expected, lastFound[0].Status)
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the app is running$`, theAppIsRunning)
	ctx.Step(`^I create an auction with product name "([^"]*)" and category "([^"]*)" and description "([^"]*)" and condition (\d+)$`, iCreateAnAuction)
	ctx.Step(`^I search for active auctions with product name "([^"]*)"$`, iSearchForActiveAuctionsWithProductName)
	ctx.Step(`^I should find at least one auction$`, iShouldFindAtLeastOneAuction)
	ctx.Step(`^I wait for the auction duration$`, iWaitForTheAuctionDuration)
	ctx.Step(`^I fetch the auction by id$`, iFetchTheAuctionById)
	ctx.Step(`^the auction status should be (\d+)$`, theAuctionStatusShouldBe)
}
