package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"
)

func TestCreateAuction_SchedulesClose(t *testing.T) {
	// set short duration for test
	os.Setenv("AUCTION_INTERVAL", "1s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// replace closeAuctionFunc to observe invocation
	orig := closeAuctionFunc
	ch := make(chan string, 1)
	closeAuctionFunc = func(ar *AuctionRepository, auctionId string) {
		ch <- auctionId
	}
	defer func() { closeAuctionFunc = orig }()

	repo := &AuctionRepository{Collection: nil}
	auction := &auction_entity.Auction{
		Id:          "test-auction-id",
		ProductName: "Product",
		Category:    "Cat",
		Description: "A very nice product",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now(),
	}

	if err := repo.CreateAuction(context.Background(), auction); err != nil {
		t.Fatalf("CreateAuction returned error: %v", err)
	}

	select {
	case id := <-ch:
		if id != auction.Id {
			t.Fatalf("expected id %s but got %s", auction.Id, id)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for scheduled close to run")
	}
}
