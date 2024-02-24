package handler

import (
	"around/model"
	"around/service"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

var (
    mediaTypes = map[string]string{
        ".jpeg": "image",
        ".jpg":  "image",
        ".gif":  "image",
        ".png":  "image",
        ".mov":  "video",
        ".mp4":  "video",
        ".avi":  "video",
        ".flv":  "video",
        ".wmv":  "video",
    }
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // // Parse from body of request to get a json object.
    // fmt.Println("Received one upload request")
    // // 1. process the request
	// // json formatted string -> go struct
    // decoder := json.NewDecoder(r.Body)
    // var p model.Post
    // if err := decoder.Decode(&p); err != nil {
    //     panic(err)
    // }

    // fmt.Fprintf(w, "Post received: %s\n", p.Message)

    fmt.Println("Received one upload request")

    

    // 1. process the request, get all the information I need

    //拿到token
    token := r.Context().Value("user")
    claims := token.(*jwt.Token).Claims
    username := claims.(jwt.MapClaims)["username"]
    
    //uuid生成新的unique id， user和message分别从token和http.Request中拿
    p := model.Post{
        Id:      uuid.New(),
        User:    username.(string),
        Message: r.FormValue("message"),
    }
    //获得media_file的file， header（metadata），以及是否有err
    file, header, err := r.FormFile("media_file")
    if err != nil {
        http.Error(w, "Media file is not available", http.StatusBadRequest)
        fmt.Printf("Media file is not available %v\n", err)
        return
    }

    //suffix是为了拿到media的type，例如.jpg, .mp4
    suffix := filepath.Ext(header.Filename)
    //然后把suffix转换成img会这video， 例如jpg -> img, mp4 -> videp
    //通过上面定义的map来实现
    if t, ok := mediaTypes[suffix]; ok {
        p.Type = t
    } else {
        p.Type = "unknown"
    }

    // 2. call service API to save
    err = service.SavePost(&p, file)
    if err != nil {
        http.Error(w, "Failed to save post to backend", http.StatusInternalServerError)
        fmt.Printf("Failed to save post to backend %v\n", err)
        return
    }

    //3. construct response 
    fmt.Println("Post is saved successfully.")

}


func searchHandler(w http.ResponseWriter, r *http.Request) {
    // Parse from body of request to get a json object.

	// 1. process the request
    fmt.Println("Received one request for search")
    w.Header().Set("Content-Type", "application/json")

    //r为传过来的信息
    //直接调用r的api获取user和keywords
    user := r.URL.Query().Get("user")
    keywords := r.URL.Query().Get("keywords")

    
	// 2. call service to handle request
    //posts来存储搜索到的所有posts
    var posts []model.Post
    var err error
    //如果user不为空， 调用backend的searchUser来搜索
    if user != "" {
        posts, err = service.SearchPostsByUser(user)
    } else {
        //如果user为空， 调用backend的searchKeyword来搜索
        //如果keyword也为空， backend已经处理了，直接返回all（所有posts）
        posts, err = service.SearchPostsByKeywords(keywords)
    }

    if err != nil {
        http.Error(w, "Failed to read post from backend", http.StatusInternalServerError)
        fmt.Printf("Failed to read post from backend %v.\n", err)
        return
    }

	// 3.construct response
    //调用json的api，把posts里面的每个post都转换为json字符串
    js, err := json.Marshal(posts)
    if err != nil {
        http.Error(w, "Failed to parse posts into JSON format", http.StatusInternalServerError)
        fmt.Printf("Failed to parse posts into JSON format %v.\n", err)
        return
    }
    //把获得的json写到w里， 并返回前端
    w.Write(js)
    
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received one request for delete")

    user := r.Context().Value("user")
    claims := user.(*jwt.Token).Claims
    username := claims.(jwt.MapClaims)["username"].(string)
    id := mux.Vars(r)["id"]

    if err := service.DeletePost(id, username); err != nil {
        http.Error(w, "Failed to delete post from backend", http.StatusInternalServerError)
        fmt.Printf("Failed to delete post from backend %v\n", err)
        return
    }
    fmt.Println("Post is deleted successfully")
}

