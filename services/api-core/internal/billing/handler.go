package billing

import (
	"context"
	"net/http"

	pb "org/api-core/proto"

	"github.com/gin-gonic/gin"
)

func CreateCheckoutSession(c *gin.Context) {

	var body struct {
		Plan string `json:"plan"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid body",
		})
		return
	}

	res, err := Client.CreateCheckoutSession(
		context.Background(),
		&pb.CreateCheckoutSessionRequest{
			Plan: body.Plan,
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