package main

import (
        "context"
        "fmt"

        "github.com/olivere/elastic/v7"
)

const (
        POST_INDEX = "post"
        USER_INDEX = "user"
		// external IP 会变, internal IP 不会变, 用internal更合适。
        ES_URL = "http://10.128.0.2:9200"
)
func main() {
    client, err := elastic.NewClient(

        elastic.SetURL(ES_URL),
		// 验证用户名和密码
        elastic.SetBasicAuth("Qin", "123456"))
    if err != nil {
        panic(err)
    }

    exists, err := client.IndexExists(POST_INDEX).Do(context.Background())
    if err != nil {
        panic(err)
    }

    if !exists {
        mapping := `{
            "mappings": {
                "properties": {
					// 没写就是index:true，有快速匹配功能
                    "id":       { "type": "keyword" },
                    "user":     { "type": "keyword" },
					// “keyword” 必须完全相同才能搜到。“text"支持部分match就能搜到（关键词就能搜到）
					// text需要额外的维护hashmap，效率降低。
                    "message":  { "type": "text" },
					// false 不能快速匹配，但是keyword本身还是可以全文匹配的
                    "url":      { "type": "keyword", "index": false },
                    "type":     { "type": "keyword", "index": false }
                }
            }
        }`
        _, err := client.CreateIndex(POST_INDEX).Body(mapping).Do(context.Background())
        if err != nil {
            panic(err)
        }
    }

    exists, err = client.IndexExists(USER_INDEX).Do(context.Background())
    if err != nil {
        panic(err)
    }

    if !exists {
        mapping := `{
                        "mappings": {
                                "properties": {
                                        "username": {"type": "keyword"},
                                        "password": {"type": "keyword"},
                                        "age":      {"type": "long", "index": false},
                                        "gender":   {"type": "keyword", "index": false}
                                }
                        }
                }`
		// _返回的值对我没用 
        _, err = client.CreateIndex(USER_INDEX).Body(mapping).Do(context.Background())
        if err != nil {
            panic(err)
        }
    }
    fmt.Println("Indexes are created.")
}
