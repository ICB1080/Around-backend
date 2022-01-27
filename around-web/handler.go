package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "path/filepath"

    "github.com/pborman/uuid"
    "regexp"
    "time"
    // jwt alias
    jwt "github.com/form3tech-oss/jwt-go"
    "github.com/gorilla/mux"  
)

var (
    // 根据后缀判断image & video
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

// 加安全密钥
var mySigningKey = []byte("secret")

// r 指向request变量的指针 如果不是指针，那么相当于copy了一个新的request对象
// 改变这个新对象对原对象没有影响，加了指针才会产生影响。（借鉴了c，在java里不存在这种，java其实pass的是c语言里的reference
// 但java 本身也被称为pass by value）
// ResponseWriter是个interface类型，没有指针和非指针的区别。
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // Parse from body of request to get a json object.
    fmt.Println("Received one post request")

    // 支持所有地址的访问
    w.Header().Set("Access-Control-Allow-Origin", "*")
    // content type:request发送当中request的格式 如json；
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

    // options 只要告诉浏览器支不支持跨域，上面header已经达到目的了，现在可以直接返回。
    if r.Method == "OPTIONS" {
        return
    }

    user := r.Context().Value("user")
    claims := user.(*jwt.Token).Claims
    username := claims.(jwt.MapClaims)["username"]

    p := Post{
        Id: uuid.New(),
        User: username.(string),
        Message: r.FormValue("message"),
    }

    file, header, err := r.FormFile("media_file")
    if err != nil {
        http.Error(w, "Media file is not available", http.StatusBadRequest)
        fmt.Printf("Media file is not available %v\n", err)
        return
    }
    // ext读取文件后缀
    suffix := filepath.Ext(header.Filename)
    // mediaTypes key-value pair 
    if t, ok := mediaTypes[suffix]; ok {
        p.Type = t
    } else {
        p.Type = "unknown"
    }

    err = savePost(&p, file)
    if err != nil {
        http.Error(w, "Failed to save post to GCS or Elasticsearch", http.StatusInternalServerError)
        fmt.Printf("Failed to save post to GCS or Elasticsearch %v\n", err)
        return
    }

    fmt.Println("Post is saved successfully.")
}


func searchHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received one request for search")
    // Header options 支持所有人跨域访问
    w.Header().Set("Access-Control-Allow-Origin", "*")
    // request的的content-type
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
    // 明确告诉前段，我返回数据的类型是json, response的header
    w.Header().Set("Content-Type", "application/json")

    if r.Method == "OPTIONS" {
        return
    }

    // 参数写在URL里
    user := r.URL.Query().Get("user")
    keywords := r.URL.Query().Get("keywords")

    var posts []Post
    var err error
    if user != "" {
        posts, err = searchPostsByUser(user)
    } else {
        posts, err = searchPostsByKeywords(keywords)
    }

    if err != nil {
        // factor: response, error message, error type
        // http.StatusInternalServerError 是常量 = 500
        http.Error(w, "Failed to read post from Elasticsearch", http.StatusInternalServerError)
        fmt.Printf("Failed to read post from Elasticsearch %v.\n", err)
        return
    }
    // Marshall 与decode相反
    js, err := json.Marshal(posts)
    if err != nil {
        http.Error(w, "Failed to parse posts into JSON format", http.StatusInternalServerError)
        fmt.Printf("Failed to parse posts into JSON format %v.\n", err)
        return
    }
    w.Write(js)
}


func signinHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received one signin request")
    // sign in 会返回token， token的格式不是json，text就行
    w.Header().Set("Content-Type", "text/plain")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == "OPTIONS" {
        return
    }

    //  Get User information from client
    // read request body
    decoder := json.NewDecoder(r.Body)
    var user User
    if err := decoder.Decode(&user); err != nil {
        http.Error(w, "Cannot decode user data from client", http.StatusBadRequest)
        fmt.Printf("Cannot decode user data from client %v\n", err)
        return
    }

    exists, err := checkUser(user.Username, user.Password)
    if err != nil {
        http.Error(w, "Failed to read user from Elasticsearch", http.StatusInternalServerError)
        fmt.Printf("Failed to read user from Elasticsearch %v\n", err)
        return
    }

    if !exists {
        http.Error(w, "User doesn't exists or wrong password", http.StatusUnauthorized)
        fmt.Printf("User doesn't exists or wrong password\n")
        return
    }
    //  create token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": user.Username,
        "exp":      time.Now().Add(time.Hour * 24).Unix(),
    })

    tokenString, err := token.SignedString(mySigningKey)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        fmt.Printf("Failed to generate token %v\n", err)
        return
    }

    w.Write([]byte(tokenString))
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received one signup request")
    w.Header().Set("Content-Type", "text/plain")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == "OPTIONS" {
        return
    }

    decoder := json.NewDecoder(r.Body)
    var user User
    if err := decoder.Decode(&user); err != nil {
        http.Error(w, "Cannot decode user data from client", http.StatusBadRequest)
        fmt.Printf("Cannot decode user data from client %v\n", err)
        return
    }

    if user.Username == "" || user.Password == "" || regexp.MustCompile(`^[a-z0-9]$`).MatchString(user.Username) {
        http.Error(w, "Invalid username or password", http.StatusBadRequest)
        fmt.Printf("Invalid username or password\n")
        return
    }

    success, err := addUser(&user)
    if err != nil {
        http.Error(w, "Failed to save user to Elasticsearch", http.StatusInternalServerError)
        fmt.Printf("Failed to save user to Elasticsearch %v\n", err)
        return
    }
    // bool = false, err = nil 唯一的情况是已有此用户。
    if !success {
        http.Error(w, "User already exists", http.StatusBadRequest)
        fmt.Println("User already exists")
        return
    }
    fmt.Printf("User added successfully: %s.\n", user.Username)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received one delete for search")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

    if r.Method == "OPTIONS" {
        return
    }

    user := r.Context().Value("user")
    claims := user.(*jwt.Token).Claims
    username := claims.(jwt.MapClaims)["username"].(string)
    id := mux.Vars(r)["id"]

    if err := deletePost(id, username); err != nil {
        http.Error(w, "Failed to delete post from Elasticsearch", http.StatusInternalServerError)
        fmt.Printf("Failed to delete post from Elasticsearch %v\n", err)
        return
    }
    fmt.Println("Post is deleted successfully")
}


