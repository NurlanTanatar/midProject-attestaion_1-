package inmemory

import (
	"context"
	"errors"
	"fmt"
	"log"
	"midProject/internal/models"
	"midProject/internal/store"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	collection *mongo.Collection

	mu *sync.RWMutex
}

func Init() store.Store {
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	return &DB{
		collection: client.Database("2fa").Collection("users"),
		mu:         new(sync.RWMutex),
	}
}

// func NewDB() store.Store {
// 	return &DB{
// 		collection: make(map[int]*models.User),
// 		mu:         new(sync.RWMutex),
// 	}
// }

func (db *DB) Create(ctx context.Context, user *models.User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.collection.InsertOne(ctx, user)
	return err
}

func (db *DB) All(ctx context.Context) ([]*models.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	filter := bson.D{{}}

	return db.filterTasks(ctx, filter)
}

func (db *DB) filterTasks(ctx context.Context, filter interface{}) ([]*models.User, error) {
	var users []*models.User

	cur, err := db.collection.Find(ctx, filter)
	if err != nil {
		return users, err
	}

	for cur.Next(ctx) {
		var t models.User
		err := cur.Decode(&t)
		if err != nil {
			return users, err
		}

		users = append(users, &t)
	}

	if err := cur.Err(); err != nil {
		return users, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func (db *DB) ByID(ctx context.Context, id string) (*models.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	u := &models.User{}
	ok := db.collection.FindOne(ctx, filter).Decode(u)
	if ok != nil {
		return nil, fmt.Errorf("no user with id %s", id)
	}

	return u, nil
}

func (db *DB) Update(ctx context.Context, user *models.User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	filter := bson.D{primitive.E{Key: "_id", Value: user.ID}}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "completed", Value: true},
	}}}

	u := &models.User{}

	return db.collection.FindOneAndUpdate(ctx, filter, update).Decode(u)
}

func (db *DB) Delete(ctx context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	res, err := db.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("no tasks were deleted")
	}

	return nil
}
