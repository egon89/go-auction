package auction_test

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestShouldCompleteAuctionWhenTimeExpires(t *testing.T) {
	beforeTest()
	defer afterTest()

	mongoContainer, databaseConnection, ctx := setUpDatabase(t)
	defer mongoContainer.Terminate(ctx)
	auctionRepository := auction.NewAuctionRepository(ctx, databaseConnection)

	auction := &auction_entity.Auction{
		Id:          uuid.New().String(),
		ProductName: "product",
		Category:    "category",
		Description: "description",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now().Add(-10 * time.Second),
	}

	auctionRepository.CreateAuction(ctx, auction)

	actives, _ := auctionRepository.FindActiveAuctions(ctx)

	time.Sleep(5 * time.Second)

	activesAfter, _ := auctionRepository.FindActiveAuctions(ctx)

	assert.Len(t, actives, 1)
	assert.Equal(t, auction.Id, actives[0].Id)
	assert.Len(t, activesAfter, 0)
}

func TestShouldNotCompleteAuctionWhenTimeNotExpires(t *testing.T) {
	beforeTest()
	defer afterTest()

	mongoContainer, databaseConnection, ctx := setUpDatabase(t)
	defer mongoContainer.Terminate(ctx)
	auctionRepository := auction.NewAuctionRepository(ctx, databaseConnection)

	auction := &auction_entity.Auction{
		Id:          uuid.New().String(),
		ProductName: "product",
		Category:    "category",
		Description: "description",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now().Add(10 * time.Second),
	}

	auctionRepository.CreateAuction(ctx, auction)

	actives, _ := auctionRepository.FindActiveAuctions(ctx)

	time.Sleep(5 * time.Second)

	activesAfter, _ := auctionRepository.FindActiveAuctions(ctx)

	assert.Len(t, actives, 1)
	assert.Equal(t, auction.Id, actives[0].Id)
	assert.Len(t, activesAfter, 1)
	assert.Equal(t, auction.Id, activesAfter[0].Id)
}

func beforeTest() {
	os.Setenv("AUCTION_DURATION", "2s")
}

func afterTest() {
	os.Unsetenv("AUCTION_DURATION")
}

func setUpDatabase(t *testing.T) (*mongodb.MongoDBContainer, *mongo.Database, context.Context) {
	ctx := context.Background()
	mongoContainer, endpoint := startMongoContainer(t, ctx)
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	require.NoError(t, err)

	databaseConnection := mongoClient.Database("test")

	err = databaseConnection.Drop(ctx)
	require.NoError(t, err)

	return mongoContainer, databaseConnection, ctx
}

func startMongoContainer(t *testing.T, ctx context.Context) (*mongodb.MongoDBContainer, string) {
	mongoContainer, err := mongodb.Run(ctx, "mongo:6")
	require.NoError(t, err)

	endpoint, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	return mongoContainer, endpoint
}
