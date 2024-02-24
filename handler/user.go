package handler

import (
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"
    "time"

    "around/model"
    "around/service"

    jwt "github.com/form3tech-oss/jwt-go"
)

var mySigningKey = []byte("secret")

func signinHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received one signin request")
    w.Header().Set("Content-Type", "text/plain")

    // 1.process request
    //  Get User information from client
    decoder := json.NewDecoder(r.Body)
    var user model.User
    if err := decoder.Decode(&user); err != nil {
        http.Error(w, "Cannot decode user data from client", http.StatusBadRequest)
        fmt.Printf("Cannot decode user data from client %v\n", err)
        return
    }

    //2. handle request
    // check if user existed
    success, err := service.CheckUser(user.Username, user.Password)
    
    //err有错误是service端的查找错误， 返回500
    if err != nil {
        http.Error(w, "Failed to read user from Elasticsearch", http.StatusInternalServerError)
        fmt.Printf("Failed to read user from Elasticsearch %v\n", err)
        return
    }

    //success为false是用户名和密码有问题， 返回400
    if !success {
        http.Error(w, "User doesn't exists or wrong password", http.StatusUnauthorized)
        fmt.Printf("User doesn't exists or wrong password\n")
        return
    }

    //3. generate token and return
    //调用jwt包生成token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": user.Username,
        //exp表示什么时候到期
        "exp":      time.Now().Add(time.Hour * 24).Unix(),
    })

    //在token中加入自定义的sign
    tokenString, err := token.SignedString(mySigningKey)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        fmt.Printf("Failed to generate token %v\n", err)
        return
    }

    //传入前端的时候写入token
    w.Write([]byte(tokenString))
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received one signup request")
    w.Header().Set("Content-Type", "text/plain")

    // 1. process request
    decoder := json.NewDecoder(r.Body)
    var user model.User
    if err := decoder.Decode(&user); err != nil {
        http.Error(w, "Cannot decode user data from client", http.StatusBadRequest)
        fmt.Printf("Cannot decode user data from client %v\n", err)
        return
    }

    // 1.5 正则表达式check用户名密码格式是否正确
    if user.Username == "" || user.Password == "" || regexp.MustCompile(`^[a-z0-9]$`).MatchString(user.Username) {
        http.Error(w, "Invalid username or password", http.StatusBadRequest)
        fmt.Printf("Invalid username or password\n")
        return
    }

    // 2. handle request
    success, err := service.AddUser(&user)

    // 后端失败，返回500
    if err != nil {
        http.Error(w, "Failed to save user to Elasticsearch", http.StatusInternalServerError)
        fmt.Printf("Failed to save user to Elasticsearch %v\n", err)
        return
    }

    //用户名密码错误， 返回400
    if !success {
        http.Error(w, "User already exists", http.StatusBadRequest)
        fmt.Println("User already exists")
        return
    }
    fmt.Printf("User added successfully: %s.\n", user.Username)
}