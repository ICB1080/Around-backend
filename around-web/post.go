package main

import (
    "reflect"
    "mime/multipart"
    "github.com/olivere/elastic/v7"
)

const (
    POST_INDEX  = "post"
)

type Post struct {
	// 首字母大写为public 首字母小写是private
    Id      string `json:"id"` // 反引号 raw string 中间有引号当初普通字符；
    User    string `json:"user"`
    Message string `json:"message"`
    Url     string `json:"url"`
    Type    string `json:"type"`
}

func searchPostsByUser(user string) ([]Post, error) {   // []post：这个user发的内容组成的数组
    query := elastic.NewTermQuery("user", user)     // 搜索某user的query 
    searchResult, err := readFromES(query, POST_INDEX)
    if err != nil {
        return nil, err
    }
    return getPostFromSearchResult(searchResult), nil
}

// keywords not array: 搜索时如果有多个关键词，每个关键词之间会用加号连接
// elastic search可以识别加号
func searchPostsByKeywords(keywords string) ([]Post, error) {
    query := elastic.NewMatchQuery("message", keywords)
    // 搜索多个关键词时：match到一个就返回：OR（默认，可以不写）， match全部才返回AND。
    query.Operator("AND")
    // 未提供关键词时返回所有内容
    if keywords == "" {
        query.ZeroTermsQuery("all")
    }
    searchResult, err := readFromES(query, POST_INDEX)
    if err != nil {
        return nil, err
    }
    return getPostFromSearchResult(searchResult), nil
}

func getPostFromSearchResult(searchResult *elastic.SearchResult) []Post {
    var ptype Post
    var posts []Post

    // TypeOf: in order to check type
    for _, item := range searchResult.Each(reflect.TypeOf(ptype)) {
        p := item.(Post)
        posts = append(posts, p)
    }
    return posts
}

func savePost(post *Post, file multipart.File) error {
    medialink, err := saveToGCS(file, post.Id)
    if err != nil {
        return err
    }
    post.Url = medialink

    return saveToES(post, POST_INDEX, post.Id)
}

func deletePost(id string, user string) error {
    query := elastic.NewBoolQuery()
    query.Must(elastic.NewTermQuery("id", id))
    query.Must(elastic.NewTermQuery("user", user))

    return deleteFromES(query, POST_INDEX)
}

