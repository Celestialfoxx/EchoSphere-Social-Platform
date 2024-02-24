// 相当于java的repository
package backend

import (
	"context"
	"fmt"

	"around/constants"

	"github.com/olivere/elastic/v7"
)

//创建一个instance，所有的handler都会调用这一个instance
//所以该instance是一个singleton，每次连接数据库都会由他执行，不需要重复创建，节省了再次创建的资源
var (
    ESBackend *ElasticsearchBackend
)

//对于第三方的包， 为了方便处理error， 可以自己wrap
type ElasticsearchBackend struct {
    client *elastic.Client
}

func InitElasticsearchBackend() {
	//从constants中获取参数， 创建一个client
	//client可以理解为是用来handle NoSQL的管理员
    client, err := elastic.NewClient(
        elastic.SetURL(constants.ES_URL),
        elastic.SetBasicAuth(constants.ES_USERNAME, constants.ES_PASSWORD))
    if err != nil {
        panic(err)
    }

	//Do(context.Background())表示在获取数据库的信息的时候backend程序该干什么
	//这里采用了默认设置， 即是一直等着， 也就是页面上一直转圈
    exists, err := client.IndexExists(constants.POST_INDEX).Do(context.Background())
    if err != nil {
        panic(err)
    }

	//如果不存在post的index（可以理解为名字为post的table），创建一个新的
	//keyword代表必须要全部输入正确才能搜索，例如course number
	//text这种可以输入一部分来搜索， 例如输入chicken可以显示chicken soup
	//这里的index默认为true，是Elasticsearch会自动给数据库的条目进行catogorize，来方便检索；如果为false代表不需要分类，节省资源
    if !exists {
        mapping := `{
            "mappings": {
                "properties": {
                    "id":       { "type": "keyword" },
                    "user":     { "type": "keyword" },
                    "message":  { "type": "text" },
                    "url":      { "type": "keyword", "index": false },
                    "type":     { "type": "keyword", "index": false }
                }
            }
        }`

		//创建新的post index（table）
        _, err := client.CreateIndex(constants.POST_INDEX).Body(mapping).Do(context.Background())
        if err != nil {
            panic(err)
        }
    }

	//user index同理
    exists, err = client.IndexExists(constants.USER_INDEX).Do(context.Background())
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
        _, err = client.CreateIndex(constants.USER_INDEX).Body(mapping).Do(context.Background())
        if err != nil {
            panic(err)
        }
    }
    fmt.Println("Indexes are created.")

    ESBackend = &ElasticsearchBackend{client: client}
}

func (backend *ElasticsearchBackend) ReadFromES(query elastic.Query, index string) (*elastic.SearchResult, error) {
    searchResult, err := backend.client.Search().
        Index(index).
        Query(query).
        Pretty(true).
        Do(context.Background())
    if err != nil {
        return nil, err
    }

    return searchResult, nil
}

//i interface{}是一个空的接口， GO中认为实现了接口的方法即为继承，因此什么都没实现的接口是所有类的父类
//因为save的type可能是Post，也可能是User， 因此这里定义一个公共父类，即是empty interface
func (backend *ElasticsearchBackend) SaveToES(i interface{}, index string, id string) error {
    //第一个index是store(存储)的意思
    //第二个index是找到要存的table
    _, err := backend.client.Index().
        Index(index).
        Id(id).
        BodyJson(i).
        Do(context.Background())
    return err
}

func (backend *ElasticsearchBackend) DeleteFromES(query elastic.Query, index string) error {
    _, err := backend.client.DeleteByQuery().
        Index(index).
        Query(query).
        Pretty(true).
        Do(context.Background())

    return err
}