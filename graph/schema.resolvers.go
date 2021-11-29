package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/zutrixpog/CMS/graph/generated"
	"github.com/zutrixpog/CMS/graph/model"
)

func (r *mutationResolver) CreatePost(ctx context.Context, input model.NewPost) (*model.Post, error) {
	err := r.DB.InsertPost(input)
	if err != nil {
		return nil, err
	}
	return &model.Post{Title: "Success"}, nil
}

func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	post, err := r.DB.FindPost(id);
	if post == nil{
		fmt.Errorf("Not found", err)
	}

	userId := ctx.Value("userId")
	err = r.DB.UpdateSeen(userId.(string), id)
	if err != nil {
		fmt.Println(err)
	}

	return post, nil
}

func (r *queryResolver) RecentPosts(ctx context.Context, page int) ([]*model.Post, error) {
	posts, err := r.DB.FindPosts(page);
	if posts == nil{
		fmt.Errorf("not found", err)
	}
	return posts, nil
}

func (r *queryResolver) PickedPosts(ctx context.Context) ([]*model.Post, error) {
	posts, err := r.DB.GetPickedPosts(5);
	if posts == nil{
		fmt.Errorf("not found", err)
	}
	return posts, nil
}

func (r *queryResolver) LovedPosts(ctx context.Context) ([]*model.Post, error) {
	posts, err := r.DB.GetLovedPosts(5);
	if posts == nil{
		fmt.Errorf("not found", err)
	}
	return posts, nil
}

func (r *queryResolver) RecommendPosts(ctx context.Context) ([]*model.Post, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) EditPost(ctx context.Context, id string, input model.NewPost) (string, error) {
	err := r.DB.UpdatePost(id, input)
	if err != nil {
		return "Could not update", err
	}
	return "Updated", nil
}

func (r *mutationResolver) DeletePost(ctx context.Context, id string) (string, error) {
	err := r.DB.DeletePost(id)
	if err != nil {
		return "Could not Delete", err
	}
	return "Deleted", nil
}

func (r *mutationResolver) PickPost(ctx context.Context, id string) (string, error){
	err := r.DB.PickPost(id)
	if err != nil {
		return "Could not Pick", err
	}
	return "Picked", nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
