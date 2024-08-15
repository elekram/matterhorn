package database

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collections struct {
	Sessions string
}

type AppDb struct {
	DBCon       *mongo.Client
	DBName      string
	Collections *Collections
}

func NewConnection(db, dbUser, dbPassWd string) *AppDb {
	DBName := db
	uri := "mongodb://" + dbUser + ":" + dbPassWd + "@mongo/" + db

	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable.")
	}

	DBCon, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Println("Error: Mongo DB Connection Failure")
		panic(err)
	}

	err = DBCon.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connected to MongoDB! 📚")
	}

	Collections := Collections{
		Sessions: "sessions",
	}

	appDatabase := AppDb{
		DBCon,
		DBName,
		&Collections,
	}

	return &appDatabase
}
