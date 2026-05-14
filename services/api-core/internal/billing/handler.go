package billing

import (
	"context"
	"log"
	"net/http"

	pb "org/api-core/proto"

	"github.com/gin-gonic/gin"
)

func CreateCheckoutSession(c *gin.Context) {
	userID := c.GetString("user_id")
	log.Println("USER ID FROM MIDDLEWARE:", userID)
	var body struct {
		PriceID string `json:"price_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid body",
		})
		return
	}

	if body.PriceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "price_id is required",
		})
		return
	}

	res, err := Client.CreateCheckoutSession(
		context.Background(),
		&pb.CreateCheckoutSessionRequest{
			UserId:  userID,
			PriceId: body.PriceID,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": res.CheckoutUrl,
	})
}
