package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var collection *mongo.Collection

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	client, err := mongo.Connect(
		options.Client().ApplyURI(os.Getenv("MONGO_URI")),
	)

	if err != nil {
		panic(err)
	}

	db := client.Database("subscriptions_db")
	collection = db.Collection("transport_subscriptions")

	r := gin.Default()

	r.GET("/my-subscriptions", getFilteredSubscriptions)

	r.Run(":8080")
}

func getFilteredSubscriptions(c *gin.Context) {

	filterString := c.Query("filter")

	if filterString == "" {
		c.JSON(400, "missing filter")
		return
	}

	var filter bson.M

	err := json.Unmarshal([]byte(filterString), &filter)
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	var results []bson.M
	cursor.All(ctx, &results)

	c.JSON(http.StatusOK, results)
}
