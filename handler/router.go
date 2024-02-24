package handler

import (
	"net/http"
    //middleware用来接收并检验token
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/form3tech-oss/jwt-go"
    //handler用来处理不同环境发来的http请求
	"github.com/gorilla/handlers"
    //mux是相当于spring用annotation处理rest请求的library
	"github.com/gorilla/mux"
)

func InitRouter() http.Handler {
    //middleware用来接收并检验token
    jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
        ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
            return []byte(mySigningKey), nil
        },

        SigningMethod: jwt.SigningMethodHS256,
    })


    //mux是相当于spring用annotation处理rest请求的library
    router := mux.NewRouter()
    //当url的后缀是/upload，并且操作类型是POST时，先通过jwt检验token， 通过后执行uploadHandler方法
    router.Handle("/upload", jwtMiddleware.Handler(http.HandlerFunc(uploadHandler))).Methods("POST")
    //当url的后缀是/search并且操作类型是GET时，先通过jwt检验token， 通过后执行searchHandler方法
    router.Handle("/search", jwtMiddleware.Handler(http.HandlerFunc(searchHandler))).Methods("GET")
    router.Handle("/post/{id}", jwtMiddleware.Handler(http.HandlerFunc(deleteHandler))).Methods("DELETE")

    router.Handle("/signup", http.HandlerFunc(signupHandler)).Methods("POST")
    router.Handle("/signin", http.HandlerFunc(signinHandler)).Methods("POST")

    //规定接受的http请求的要求
    //1.表示接受任何其他环境的请求， 例如后端在GCP上， 从AWS发来的前端请求
    originsOk := handlers.AllowedOrigins([]string{"*"})
    //2.表示只接受header上只包含Authorization和Content-Type的请求
    headersOk := handlers.AllowedHeaders([]string{"Authorization", "Content-Type"})
    //3.表示只接受GET， POST， DELETE的http请求
    methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "DELETE"})

    return handlers.CORS(originsOk, headersOk, methodsOk)(router)
}