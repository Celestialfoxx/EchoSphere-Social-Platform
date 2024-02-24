package model

//用于接收json文件然后转变为object
type Post struct {
    Id      string `json:"id"`
    User    string `json:"user"`
    Message string `json:"message"`
    Url     string `json:"url"`
    Type    string `json:"type"`
}

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Age      int64  `json:"age"`
    Gender   string `json:"gender"`
}