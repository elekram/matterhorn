package database

import (
	"context"
	"fmt"
	"log"

	appcfg "github.com/elekram/matterhorn/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DBCon *mongo.Client
)

func NewConnection() *mongo.Client {
	uri := "mongodb://" + appcfg.Props.MongoUsername + ":" + appcfg.Props.MongoPassword + "@mongo/dev_db"

	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable.")
	}

	DBCon, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Println("Error: Mongo DB Connection")
		panic(err)
	}

	err = DBCon.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connected to MongoDB! 📚")
	}

	return DBCon
}
