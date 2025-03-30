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

	go repository.startExpiredAuctionMonitor(ctx, getAuctionDuration())

	go repository.processExpiredAuction(ctx)

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

func (ar *AuctionRepository) startExpiredAuctionMonitor(ctx context.Context, auctionDuration time.Duration) {
	log.Println("[expired-monitor] Starting expired auction monitor")

	ticker := time.NewTicker(auctionDuration)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("[expired-monitor] Checking for expired auctions")
		ar.checkExpiredAuction(ctx, auctionDuration)
	}
}

func (ar *AuctionRepository) checkExpiredAuction(ctx context.Context, auctionDuration time.Duration) {
	log.Println("[expired-auction-check] Checking for expired auctions")

	activeAuctions, err := ar.FindActiveAuctions(ctx)
	if err != nil {
		log.Println("[expired-auction-check] Error trying to find auctions")
		return
	}

	log.Printf("[expired-auction-check] Found %d active auctions\n", len(activeAuctions))

	var wg sync.WaitGroup
	for _, auction := range activeAuctions {
		wg.Add(1)
		go func(auction auction_entity.Auction) {
			defer wg.Done()

			now := time.Now().UTC()
			auctionTimestamp := auction.Timestamp.UTC()
			auctionWithDuration := auctionTimestamp.Add(auctionDuration)

			if now.After(auctionWithDuration) {
				log.Printf("[expired-auction-check] Sending auction %s to expired channel. Auction duration deadline: %s. Now: %s\n",
					auction.Id, auctionWithDuration, now)
				ar.expiredChannel <- auction.Id
			}
		}(auction)
	}
	wg.Wait()
}

func (ar *AuctionRepository) processExpiredAuction(ctx context.Context) {
	log.Println("[expired-auction-process] Starting process expired auction")

	for auctionId := range ar.expiredChannel {
		log.Printf("[expired-auction-process]  Updating auction %s status to completed\n", auctionId)
		ar.UpdateAuctionStatusById(ctx, auctionId, auction_entity.Completed)
	}

	log.Println("[expired-auction-process] Process expired auction finished")
}

func getAuctionDuration() time.Duration {
	interval := os.Getenv("AUCTION_DURATION")
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}
