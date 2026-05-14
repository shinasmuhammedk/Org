package billing

import (
	"context"
	"log"

	pb "org/api-core/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Client pb.BillingServiceClient

func Connect() {
	conn, err := grpc.Dial(
		"localhost:50052",
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	Client = pb.NewBillingServiceClient(conn)

	log.Println("Connected to billing service")
}


func GetUserSubscription(
	userID string,
) (string, string, error) {

	res, err := Client.GetUserSubscription(
		context.Background(),
		&pb.GetUserSubscriptionRequest{
			UserId: userID,
		},
	)

	if err != nil {
		return "", "", err
	}

	return res.Plan, res.Status, nil
}