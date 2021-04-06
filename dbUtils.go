package shortly_url

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseConnection will help you create a mongo connection and will return 1. only one collection in this project need 2. context 3. error
func DatabaseConnection() (*mongo.Collection, context.Context, error) {
	dbClient, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Panicln(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = dbClient.Connect(ctx); err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	collection := dbClient.Database(os.Getenv("MONGO_DATABASE")).Collection("shortlyData")
	return collection, ctx, nil
}

func SetupMongoCollectionIndex() error {
	collection, ctx, err := DatabaseConnection()
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"expireAt": 1,
		},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	return nil
}

func InsertIntoDB(collection *mongo.Collection, shortlyData *ShortlyData) error {
	_, err := collection.InsertOne(context.Background(), shortlyData)
	return err
}

func FindShortlyData(shortUUID string, ctx context.Context, collection *mongo.Collection) (ShortlyData, error) {
	filter := bson.M{
		"shortUUID": shortUUID,
	}
	mongoData := collection.FindOne(ctx, filter)
	var shortlyData ShortlyData
	if err := mongoData.Decode(&shortlyData); err != nil {
		return shortlyData, err
	}
	return shortlyData, nil
}

func DeleteShortlyData(shortUUID string, ctx context.Context, collection *mongo.Collection) error {
	filter := bson.M{
		"shortUUID": shortUUID,
	}
	mongoData := collection.FindOne(ctx, filter)
	var shortlyData ShortlyData
	if err := mongoData.Decode(&shortlyData); err != nil {
		return err
	}
	id, _ := primitive.ObjectIDFromHex(shortlyData.ID)
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})

	return err
}
