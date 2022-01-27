package main

import (
    "context"

    "github.com/olivere/elastic/v7"
)

const (
        ES_URL = "http://10.128.0.2:9200"
)


func readFromES(query elastic.Query, index string) (*elastic.SearchResult, error) {
    // create a client to work with elasticsearch
    client, err := elastic.NewClient(
        elastic.SetURL(ES_URL),
        elastic.SetBasicAuth("Qin", "123456"))
    if err != nil {
        return nil, err
    }

    searchResult, err := client.Search().
        Index(index).	// search in index ""index"
        Query(query).	// specify the query
        Pretty(true).	// pretty print request and response JSON
        Do(context.Background())	// execute
    if err != nil {
        return nil, err
    }

    return searchResult, nil
}

// interface? saveToES为了支持各种不同类型的存储:Post,User(struct)
// id string as primary key
func saveToES(i interface{}, index string, id string) error{

    client, err := elastic.NewClient(
        elastic.SetURL(ES_URL),
        elastic.SetBasicAuth("Qin", "123456"))
    if err != nil {
        return err
    }

    _, err = client.Index().
        Index(index).
        Id(id).
        BodyJson(i).
        // context: not associated with function
        Do(context.Background())
    return err
}

func deleteFromES(query elastic.Query, index string) error {
    client, err := elastic.NewClient(
        elastic.SetURL(ES_URL),
        elastic.SetBasicAuth("Qin", "123456"))
    if err != nil {
        return err
    }

    _, err = client.DeleteByQuery().
        Index(index).
        Query(query).
        Pretty(true).
        Do(context.Background())

    return err
}