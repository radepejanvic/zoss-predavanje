package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var mongoCollection *mongo.Collection
var rdb *redis.Client
var ctx = context.Background()

func main() {

	mongoClient, err := mongo.Connect(
		options.Client().ApplyURI(os.Getenv("MONGO_URI")),
	)
	if err != nil {
		panic(err)
	}

	db := mongoClient.Database("routes_db")
	mongoCollection = db.Collection("transport_routes")

	rdb = redis.NewClient(&redis.Options{
		Addr:      os.Getenv("REDIS_HOST"),
		Password:  os.Getenv("REDIS_PASSWORD"),
		DB:        0,
		TLSConfig: nil,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("Redis connection failed: " + err.Error())
	}

	r := gin.Default()
	r.GET("/routes", getRoutes)

	r.Run(":8080")
}

func getRoutes(c *gin.Context) {

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	val, err := rdb.Get(ctxTimeout, "routes").Result()
	if err == nil {
		var cachedRoutes []bson.M
		if err := json.Unmarshal([]byte(val), &cachedRoutes); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"source": "cache",
				"routes": cachedRoutes,
			})
			return
		}
	}

	cursor, err := mongoCollection.Find(ctxTimeout, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	defer cursor.Close(ctxTimeout)

	var routes []bson.M
	if err := cursor.All(ctxTimeout, &routes); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	jsonData, err := json.Marshal(routes)
	if err == nil {
		rdb.Set(ctxTimeout, "routes", jsonData, 60*time.Second)
	}

	c.JSON(http.StatusOK, gin.H{
		"source": "db",
		"routes": routes,
	})
}
