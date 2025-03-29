package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"log"
	"os"
	"sync"
	"time"

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
	expiredChannel chan string
	Collection     *mongo.Collection
}

func NewAuctionRepository(ctx context.Context, database *mongo.Database) *AuctionRepository {
	log.Println("Creating new auction repository")

	repository := &AuctionRepository{
		Collection:     database.Collection("auctions"),
		expiredChannel: make(chan string, 5),
	}

	go repository.startExpiredAuctionMonitor(ctx)

	go repository.processExpiredAuction()

	return repository
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
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	return nil
}

func (ar *AuctionRepository) startExpiredAuctionMonitor(ctx context.Context) {
	log.Println("Starting expired auction monitor")

	ticker := time.NewTicker(getAuctionDuration())
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Checking for expired auctions")
		ar.checkExpiredAuction(ctx)
	}
}

func (ar *AuctionRepository) checkExpiredAuction(ctx context.Context) {
	log.Println("Checking expired auctions")

	activeAuctions, err := ar.FindAuctions(ctx, auction_entity.Active, "", "")
	if err != nil {
		log.Println("Error trying to find auctions")
		return
	}

	log.Printf("Found %d active auctions\n", len(activeAuctions))

	var wg sync.WaitGroup
	for _, auction := range activeAuctions {
		wg.Add(1)
		go func(auction auction_entity.Auction) {
			defer wg.Done()

			now := time.Now().UTC()
			auctionTimestamp := auction.Timestamp.UTC()
			auctionWithDuration := auctionTimestamp.Add(getAuctionDuration())

			if now.After(auctionWithDuration) {
				log.Printf("Auction %s expired\n", auction.Id)
				ar.expiredChannel <- auction.Id
			}
		}(auction)
	}
	wg.Wait()
}

func (ar *AuctionRepository) processExpiredAuction() {
	log.Println("Process expired auction started")

	for auctionId := range ar.expiredChannel {
		log.Printf("Processing auction id %s\n", auctionId)
		// TODO: update auction status to completed
	}

	log.Println("Process expired auction finished")
}

func getAuctionDuration() time.Duration {
	interval := os.Getenv("AUCTION_DURATION")
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}
