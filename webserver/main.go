package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
)

func main() {
	stripe.Key = os.Getenv("STRIPE_SK")

	r := gin.Default()
	r.POST("/app/upgrade", upgrade)
	r.Run()
}

type Upgrade struct {
	UserId          string `form:"user_id"`
	UserEmail       string `form:"user_email"`
	SuccessUrl      string `form:"success_url"`
	StripeToken     string `form:"stripeToken"`
	StripeTokenType string `form:"stripeTokenType"`
	StripeEmail     string `form:"stripeEmail"`
}

func upgrade(c *gin.Context) {
	var form Upgrade
	if err := c.Bind(&form); err == nil {
		fmt.Println("Received form", spew.Sdump(form))

		customerParams := &stripe.CustomerParams{
			Desc: "Customer for Google email: " + form.UserEmail,
		}
		customerParams.SetSource(form.StripeToken)

		customer, err := customer.New(customerParams)

		if err != nil {
			upgradeError(c, err)
		}

		fmt.Println("Created customer", spew.Sdump(customer))
		subParams := &stripe.SubParams{
			Customer: customer.ID,
			Items: []*stripe.SubItemsParams{
				{
					Plan: os.Getenv("STRIPE_PP_PLAN"),
				},
			},
		}
		subParams.AddMeta("user_id", form.UserId)
		subscription, err := sub.New(subParams)

		if err != nil {
			upgradeError(c, err)
		}

		fmt.Println("Created subscription", spew.Sdump(subscription))

		c.Redirect(http.StatusFound, "/site/payment-success.html")
	} else {
		upgradeError(c, err)
	}

}

func upgradeError(c *gin.Context, err error) {
	// If we're here then we failed, so its not good....
	fmt.Println("Failed to process Stripe subscription")
	spew.Dump(err)
	c.Redirect(http.StatusFound, "/site/payment-failure.html")
}
