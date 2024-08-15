package main

import (
	"context"
	"fmt"

	database "github.com/elekram/matterhorn/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongostore struct {
	dbCon          *mongo.Client
	dbName         string
	collectionName string
}

type Sess struct {
	Session_id     string
	Session_expiry primitive.DateTime `bson:"session_expiry"`
	Email          string             `bson:"email"`
	Name           string             `bson:"name"`
	Family_name    string             `bson:"family_name"`
	Given_name     string             `bson:"given_name"`
	Account_id     string             `bson:"account_id"`
	Hd             string             `bson:"hd"`
	Picture        string             `bson:"picture"`
}

func newMongoStore(db *database.AppDb) *mongostore {

	ms := mongostore{
		db.DBCon,
		db.DBName,
		db.Collections.Sessions,
	}

	return &ms
}

func (ms *mongostore) addSession(id string, s session) {
	expiry := primitive.NewDateTimeFromTime(s.expiry)

	newSession := Sess{
		Session_id:     id,
		Session_expiry: expiry,
		Email:          s.profile.email,
		Name:           s.profile.name,
		Family_name:    s.profile.family_name,
		Given_name:     s.profile.given_name,
		Hd:             s.profile.hd,
		Account_id:     s.profile.id,
		Picture:        s.profile.picture,
	}

	coll := ms.dbCon.Database(ms.dbName).Collection(ms.collectionName)

	result, err := coll.InsertOne(context.TODO(), newSession)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	fmt.Println(result)
}

func (ms *mongostore) getSession(id string) (session, bool) {
	s := session{}

	coll := ms.dbCon.Database(ms.dbName).Collection(ms.collectionName)

	filter := bson.D{{"session_id", id}}

	var result Sess
	err := coll.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println(err)
			return s, false
		}
		fmt.Println(err)
		return s, false
	}

	s.expiry = result.Session_expiry.Time()
	s.profile.email = result.Email
	s.profile.name = result.Name
	s.profile.family_name = result.Email
	s.profile.given_name = result.Given_name
	s.profile.id = result.Account_id
	s.profile.picture = result.Picture

	return s, true
}
