package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"

	"go.mongodb.org/mongo-driver/bson"
)

func (ar *AuctionRepository) UpdateAuctionStatusById(
	ctx context.Context,
	id string,
	status auction_entity.AuctionStatus) *internal_error.InternalError {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status}}
	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(fmt.Sprintf("Error trying to update auction status by id = %s", id), err)
		return internal_error.NewInternalServerError("Error trying to update auction status by id")
	}

	return nil
}
