package service

import (
	"mime/multipart"
	"reflect"

	"around/backend"
	"around/constants"
	"around/model"

	"github.com/olivere/elastic/v7"
)

func SearchPostsByUser(user string) ([]model.Post, error) {
	// 1. create a query
    query := elastic.NewTermQuery("user", user)

	//2. call backend to search
    searchResult, err := backend.ESBackend.ReadFromES(query, constants.POST_INDEX)
	//if error happened, return error
    if err != nil {
        return nil, err
    }

	// 3. process and return results
    return getPostFromSearchResult(searchResult), nil
}

func SearchPostsByKeywords(keywords string) ([]model.Post, error) {
	//1. create a query， strings connects with '+'， like ann+john
    query := elastic.NewMatchQuery("message", keywords)
	// 'AND' means 只要match一个keyword就算hit
    query.Operator("AND")
	// if keyword is "", return all documents(返回所有条目)
    if keywords == "" {
        query.ZeroTermsQuery("all")
    }

	//2. call backend to search
    searchResult, err := backend.ESBackend.ReadFromES(query, constants.POST_INDEX)
    if err != nil {
        return nil, err
    }

	//3. process and return results
    return getPostFromSearchResult(searchResult), nil
}

func getPostFromSearchResult(searchResult *elastic.SearchResult) []model.Post {
    var ptype model.Post
    var posts []model.Post

    for _, item := range searchResult.Each(reflect.TypeOf(ptype)) {
        p := item.(model.Post)
        posts = append(posts, p)
    }
    return posts
}

func SavePost(post *model.Post, file multipart.File) error {
    // 1. save file to GCS by calling backend API
    medialink, err := backend.GCSBackend.SaveToGCS(file, post.Id)
    if err != nil {
        return err
    }

    // 2. set post.url
    post.Url = medialink

    // 3. save post to ES
    // 4. return or response(opitonal)
    return backend.ESBackend.SaveToES(post, constants.POST_INDEX, post.Id)
}

func DeletePost(id string, user string) error {
    query := elastic.NewBoolQuery()
    query.Must(elastic.NewTermQuery("id", id))
    query.Must(elastic.NewTermQuery("user", user))

    return backend.ESBackend.DeleteFromES(query, constants.POST_INDEX)
}