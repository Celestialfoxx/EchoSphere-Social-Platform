package service

import (
    "fmt"
    "reflect"

    "around/backend"
    "around/constants"
    "around/model"

    "github.com/olivere/elastic/v7"
)

func CheckUser(username, password string) (bool, error) {
	//boolean query可以设置必须满足
    query := elastic.NewBoolQuery()
	//Must表示username和password必须是指定的内容
    query.Must(elastic.NewTermQuery("username", username))
    query.Must(elastic.NewTermQuery("password", password))
    searchResult, err := backend.ESBackend.ReadFromES(query, constants.USER_INDEX)
    if err != nil {
        return false, err
    }

    var utype model.User
    for _, item := range searchResult.Each(reflect.TypeOf(utype)) {
        u := item.(model.User)
        if u.Password == password {
            fmt.Printf("Login as %s\n", username)
            return true, nil
        }
    }
    return false, nil
}

func AddUser(user *model.User) (bool, error) {
	//1. check is username existed
    query := elastic.NewTermQuery("username", user.Username)
    searchResult, err := backend.ESBackend.ReadFromES(query, constants.USER_INDEX)
    if err != nil {
        return false, err
    }

	//if existed, return false
    if searchResult.TotalHits() > 0 {
        return false, nil
    }

	//if not, save to ES
    err = backend.ESBackend.SaveToES(user, constants.USER_INDEX, user.Username)
    if err != nil {
        return false, err
    }
    fmt.Printf("User is added: %s\n", user.Username)
    return true, nil
}