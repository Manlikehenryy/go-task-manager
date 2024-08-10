package controllers

import "go.mongodb.org/mongo-driver/mongo"

var usersCollection *mongo.Collection
var tasksCollection *mongo.Collection


func InitDB(DB *mongo.Database) {

	usersCollection = DB.Collection("users")
	tasksCollection = DB.Collection("tasks")
}
