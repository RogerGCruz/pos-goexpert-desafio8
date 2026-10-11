package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}

type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

// closeAuctionFunc is a replaceable function used to perform the DB update
// when an auction should be closed. Tests can replace this to observe the
// scheduled close without requiring a real MongoDB instance.
var closeAuctionFunc = func(ar *AuctionRepository, auctionId string) {
	// perform the update in DB to set status to Completed
	if ar == nil || ar.Collection == nil {
		return
	}
	filter := bson.M{"_id": auctionId}
	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}
	ctx := context.Background()
	if _, err := ar.Collection.UpdateOne(ctx, filter, update); err != nil {
		logger.Error("Error trying to update auction status to Completed", err)
	}
}

func getAuctionDuration() time.Duration {
	// Read AUCTION_INTERVAL (e.g. "20s") from env. Default to 5 seconds when missing or invalid.
	duration, err := time.ParseDuration(os.Getenv("AUCTION_INTERVAL"))
	if err != nil || duration <= 0 {
		return 5 * time.Second
	}
	return duration
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}

	// Insert only if collection is available (production path). Tests may set
	// a nil collection and still expect the scheduling to run.
	if ar.Collection != nil {
		_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
		if err != nil {
			logger.Error("Error trying to insert auction", err)
			return internal_error.NewInternalServerError("Error trying to insert auction")
		}
	}

	// Schedule the auction to be closed after configured duration in background.
	duration := getAuctionDuration()
	go func(id string) {
		// non-blocking; sleep then attempt to close auction
		time.Sleep(duration)
		closeAuctionFunc(ar, id)
	}(auctionEntity.Id)

	return nil
}
