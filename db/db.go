package db

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/zutrixpog/CMS/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/zutrixpog/CMS/fs"
)

type DB struct {
	DbName     string
	client     *mongo.Client
	cancel     context.CancelFunc
	dateLayout string
	collection *mongo.Collection
}

var catmap = map[string]*model.Categories{
	"TECH":   &model.AllCategories[0],
	"CRYPTO": &model.AllCategories[1],
	"CS":     &model.AllCategories[2],
}

var mapcat = map[*model.Categories]string{
	&model.AllCategories[0]: "TECH",
	&model.AllCategories[1]: "CRYPTO",
	&model.AllCategories[2]: "CS",
}

func (this *DB) Connect(url string) *mongo.Client {
	this.dateLayout = "01-02-2006 15:04:05"
	var ctx context.Context
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	checkError(err)

	ctx, this.cancel = context.WithCancel(context.Background())
	err = client.Connect(ctx)
	this.client = client
	checkError(err)
	return client
}

func (this *DB) FindPosts(page int) ([]*model.Post, error) {
	ctx := context.Background()

	limit := 10
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", -1}})
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))

	this.collection = this.client.Database(this.DbName).Collection("posts")
	cursor, err := this.collection.Find(ctx, bson.M{}, findOptions)
	checkError(err)
	defer cursor.Close(ctx)

	var posts []*model.Post
	for cursor.Next(ctx) {
		var post bson.M
		err = cursor.Decode(&post)
		if err != nil {
			return posts, nil
		}

		posts = append(posts, &model.Post{
			ID:           primitive.ObjectID(post["_id"].(primitive.ObjectID)).String(),
			Title:        post["Title"].(string),
			Author:       this.getAuthor(post["Author"].(primitive.ObjectID)),
			Banner:       post["Banner"].(string),
			Date:         post["Date"].(string),
			MarkdownText: this.getMarkdown(post["MarkdownFile"].(string)),
			Category:     this.getCategories(post["Category"].(primitive.A)),
		})
	}

	if posts == nil {
		return nil, errors.New("Problem fetching data!")
	}

	return posts, nil
}

func (this *DB) InsertPost(post model.NewPost) error {
	this.collection = this.client.Database(this.DbName).Collection("posts")

	var categories bson.A
	for i := range post.Category {
		categories = append(categories, mapcat[post.Category[i]])
	}

	err := fs.WriteMarkdown(post.Title, strings.NewReader(post.MarkdownFile))
	checkError(err)

	authorid, _ := primitive.ObjectIDFromHex(post.Author)
	newPost := bson.D{
		{"Title", post.Title},
		{"Author", authorid},
		{"Banner", post.Banner},
		{"Date", post.Date},
		{"MarkdownFile", "Markdown-" + post.Title},
		{"Category", categories},
	}

	_, err = this.collection.InsertOne(context.Background(), newPost)
	if err != nil {
		return errors.New("Problem inserting the document!")
	}
	return nil
}

func (this *DB) FindPost(id string) (*model.Post, error) {
	objectid, _ := primitive.ObjectIDFromHex(id)
	this.collection = this.client.Database(this.DbName).Collection("posts")

	var post bson.M
	err := this.collection.FindOne(context.Background(), bson.D{{"_id", objectid}}).Decode(&post)
	if err != nil {
		return nil, err
	}

	return &model.Post{
		ID:           primitive.ObjectID(post["_id"].(primitive.ObjectID)).String(),
		Title:        post["Title"].(string),
		Author:       this.getAuthor(post["Author"].(primitive.ObjectID)),
		Banner:       post["Banner"].(string),
		Date:         post["Date"].(string),
		MarkdownText: this.getMarkdown(post["MarkdownFile"].(string)),
		Category:     this.getCategories(post["Category"].(primitive.A)),
	}, nil
}

func (this *DB) getPosts(coll string, findOptions *options.FindOptions) ([]*model.Post, error) {
	this.collection = this.client.Database(this.DbName).Collection(coll)
	ctx := context.Background()

	cursor, err := this.collection.Find(ctx, bson.M{}, findOptions)
	checkError(err)
	defer cursor.Close(ctx)

	var posts []*model.Post
	for cursor.Next(ctx) {
		var post bson.M
		err = cursor.Decode(&post)
		if err != nil {
			return posts, nil
		}
		if t, _ := time.Parse(this.dateLayout, post["Date"].(string)); t.Day() != time.Now().Day() && t.Month() != time.Now().Month() {
			break
		}

		posts = append(posts, &model.Post{
			ID:           primitive.ObjectID(post["_id"].(primitive.ObjectID)).String(),
			Title:        post["Title"].(string),
			Author:       this.getAuthor(post["Author"].(primitive.ObjectID)),
			Banner:       post["Banner"].(string),
			Date:         post["Date"].(string),
			MarkdownText: this.getMarkdown(post["MarkdownFile"].(string)),
			Category:     this.getCategories(post["Category"].(primitive.A)),
		})
	}

	if posts == nil {
		return nil, errors.New("Problem fetching data!")
	}

	return posts, nil
}

func (this *DB) GetPickedPosts(limit int64) ([]*model.Post, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", -1}})
	findOptions.SetLimit(limit)

	posts, err := this.getPosts("picks", findOptions)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (this *DB) GetLovedPosts(limit int64) ([]*model.Post, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", -1}})
	findOptions.SetSort(bson.D{{"views", -1}})
	findOptions.SetLimit(limit)

	posts, err := this.getPosts("posts", findOptions)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (this *DB) UpdatePost(id string, newPost model.NewPost) error {
	this.collection = this.client.Database(this.DbName).Collection("posts")
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", objectId}}
	update := bson.D{{"$set", bson.A{bson.D{{"Title", newPost.Title}}, bson.D{{"Author", newPost.Author}}, bson.D{{"Banner", newPost.Banner}}, bson.D{{"Category", newPost.Category}}}}}

	fs.WriteMarkdown(this.getFilename(objectId), strings.NewReader(newPost.MarkdownFile))

	_, err := this.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (this *DB) DeletePost(id string) error {
	this.collection = this.client.Database(this.DbName).Collection("posts")
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", objectId}}

	err := fs.RemoveMarkdown(this.getFilename(objectId))

	_, err = this.collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (this *DB) PickPost(id string) error {
	this.collection = this.client.Database(this.DbName).Collection("picks")
	objectId, _ := primitive.ObjectIDFromHex(id)

	newPick := bson.D{
		{"post", objectId},
		{"date", time.Now().Format(this.dateLayout)},
	}
	_, err := this.collection.InsertOne(context.Background(), newPick)
	if err != nil {
		return err
	}
	return nil
}

func (this *DB) CreateUser(ip string) string {
	this.collection = this.client.Database(this.DbName).Collection("users")

	if exists, userId := this.userExists(ip); exists {
		return userId
	}

	userId, _ := uuid.NewV4()

	newUser := bson.D{
		{"userId", userId.String()},
		{"ip", ip},
	}
	_, err := this.collection.InsertOne(context.Background(), newUser)
	if err != nil {
		log.Fatal(errors.New("Problem inserting the document!"))
	}

	return userId.String()
}

func (this *DB) UpdateSeen(userId string, postId string) error {
	ctx := context.Background()
	this.collection = this.client.Database(this.DbName).Collection("users")
	postColl := this.client.Database(this.DbName).Collection("posts")

	var seen bson.D
	err := this.collection.FindOne(ctx, bson.D{{"userId", userId}}).Decode(&seen)
	if err != nil {
		return err
	}

	seen = append(seen, bson.E{"postId", postId})
	_, err = this.collection.UpdateOne(ctx, bson.D{{"userId", userId}}, seen)
	if err != nil {
		return err
	}

	post, _ := this.FindPost(postId)
	postColl.UpdateOne(ctx, bson.D{{"_id", postId}}, bson.D{{"$set", bson.D{{"views", post.Views + 1}}}})

	return nil
}

func (this *DB) userExists(ip string) (bool, string) {
	this.collection = this.client.Database(this.DbName).Collection("users")

	var user bson.M
	err := this.collection.FindOne(context.Background(), bson.D{{"ip", ip}}).Decode(&user)
	if err != nil {
		return false, ""
	}
	return true, user["userId"].(string)
}

func (this *DB) getMarkdown(filename string) string {
	markdown, err := fs.ReadMarkdown(filename)
	if err != nil {
		markdown = "there was a problem"
	}
	return markdown
}

func (this *DB) getFilename(id primitive.ObjectID) string {
	this.collection = this.client.Database(this.DbName).Collection("posts")
	var post bson.M
	this.collection.FindOne(context.Background(), bson.D{{"_id", id}}).Decode(&post)
	return post["MarkdownFile"].(string)
}

func (this *DB) getCategories(categories primitive.A) []*model.Categories {
	var result []*model.Categories
	for i := range categories {
		result = append(result, catmap[categories[i].(string)])
	}
	return result
}

func (this *DB) getAuthor(id primitive.ObjectID) *model.Author {
	this.collection = this.client.Database(this.DbName).Collection("authors")

	var Author bson.M
	err := this.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&Author)
	checkError(err)
	return &model.Author{
		ID:          Author["_id"].(primitive.ObjectID).String(),
		Name:        Author["Name"].(string),
		Description: Author["Description"].(string),
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
	return
}
