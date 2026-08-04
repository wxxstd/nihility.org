package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongo_client *mongo.Client
var db *mongo.Database

var dbSTAFF         *mongo.Collection
var dbREGISTER_SHA  *mongo.Collection
var dbTRANSLATIONS  *mongo.Collection
var dbROLES         *mongo.Collection

func dbConnect() {
    var err error

    // Use the SetServerAPIOptions() method to set the Stable API version to 1
    serverAPI := options.ServerAPI(options.ServerAPIVersion1)
    opts := options.Client().ApplyURI(Config.Dbase.Url).SetServerAPIOptions(serverAPI)

    // Create a new client and connect to the server
    mongo_client, err = mongo.Connect(context.Background(), opts)
    if err != nil {
        panic(err)
    }
    db = mongo_client.Database(Config.Dbase.Name)

    // Send a ping to confirm a successful connection
    var result bson.M
    if err := db.RunCommand(context.Background(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
        panic(err)
    }

    dbSTAFF =           db.Collection("staff")
	dbREGISTER_SHA =    db.Collection("register_sha")
    dbTRANSLATIONS =    db.Collection("translations")
    dbROLES =           db.Collection("roles")
}

// =====================================================================================================================
// Internal Staff Listing CRUD

func (staff *Staff) List() ([]Staff, error) {
    var staffList []Staff

    cursor, err := dbSTAFF.Find(context.Background(), bson.D{{}})
    if nil != err {
        return staffList, err
    }
    defer cursor.Close(context.Background())

    err = cursor.All(context.Background(), &staffList)

    return staffList, err
}

func (staff *Staff) Select(id primitive.ObjectID) error {
    return dbSTAFF.FindOne(context.Background(), bson.D{{"_id", id}}).Decode(staff)
}

func (staff *Staff) FindByName(name string) error {
    return dbSTAFF.FindOne(context.Background(), bson.D{{"nick", name}}).Decode(staff)
}

func (staff *Staff) Add() error {
    _, err := dbSTAFF.InsertOne(context.Background(), staff)
    return err
}

func (staff *Staff) Update() error {
    _, err := dbSTAFF.ReplaceOne(context.Background(), bson.D{{"_id", staff.Id}}, staff)
    return err
}

func (staff *Staff) Delete() error {
    filter := bson.D{{"_id", staff.Id}}
    _, err := dbSTAFF.DeleteOne(context.Background(), filter)
    return err
}

// =====================================================================================================================
// Internal Roles Listing CRUD

func (role *Role) List() ([]Role, error) {
    var rolesList []Role

    cursor, err := dbROLES.Find(context.Background(), bson.D{{}})
    if nil != err {
        return rolesList, err
    }
    defer cursor.Close(context.Background())

    err = cursor.All(context.Background(), &rolesList)

    return rolesList, err
}

func (role *Role) Select(id primitive.ObjectID) error {
    return dbROLES.FindOne(context.Background(), bson.D{{"_id", id}}).Decode(role)
}

func (role *Role) FindByName(name string) error {
    return dbROLES.FindOne(context.Background(), bson.D{{"nick", name}}).Decode(role)
}

func (role *Role) Add() error {
    _, err := dbROLES.InsertOne(context.Background(), role)
    return err
}

func (role *Role) Update() error {
    _, err := dbROLES.ReplaceOne(context.Background(), bson.D{{"_id", role.Id}}, role)
    return err
}

func (role *Role) Delete() error {
    filter := bson.D{{"_id", role.Id}}
    _, err := dbROLES.DeleteOne(context.Background(), filter)
    return err
}

// =====================================================================================================================
// Internal Translations Listing CRUD

func (tr *Translation) List() ([]Translation, error) {
    var trList []Translation

    cursor, err := dbTRANSLATIONS.Find(context.Background(), bson.D{{}})
    if nil != err {
        return trList, err
    }
    defer cursor.Close(context.Background())

    err = cursor.All(context.Background(), &trList)

    return trList, err
}

func (tr *Translation) Select(id primitive.ObjectID) error {
    return dbTRANSLATIONS.FindOne(context.Background(), bson.D{{"_id", id}}).Decode(tr)
}

func (tr *Translation) FindByName(name string) error {
    return dbTRANSLATIONS.FindOne(context.Background(), bson.D{{"nick", name}}).Decode(tr)
}

func (tr *Translation) Add() error {
    _, err := dbTRANSLATIONS.InsertOne(context.Background(), tr)
    return err
}

func (tr *Translation) Update() error {
    _, err := dbTRANSLATIONS.ReplaceOne(context.Background(), bson.D{{"_id", tr.Id}}, tr)
    return err
}

func (tr *Translation) Delete() error {
    filter := bson.D{{"_id", tr.Id}}
    _, err := dbTRANSLATIONS.DeleteOne(context.Background(), filter)
    return err
}

// =====================================================================================================================
// Register SHA CRUD

func (sha *RegisterSha) Find(sha_str string) error {
	return dbREGISTER_SHA.FindOne(context.TODO(), bson.D{{"sha", sha_str}}).Decode(sha)
}

func (sha *RegisterSha) Add() error {
	_, err := dbREGISTER_SHA.InsertOne(context.TODO(), sha)
	return err
}

func (sha *RegisterSha) Delete() error {
	filter := bson.D{{"_id", sha.Id}}
	_, err := dbREGISTER_SHA.DeleteOne(context.TODO(), filter)
	return err
}
