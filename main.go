package main

import (
    "fmt"
    "github.com/gorilla/mux"
    "github.com/gorilla/securecookie"
    "net/http"
    "io/ioutil"
    "strings"
)

var router = mux.NewRouter()

func main() {

    router.HandleFunc("/", onboardingPageHandler)
    router.HandleFunc("/internal", inWorldPageHandler)

    router.HandleFunc("/login", loginHandler).Methods("POST")
    router.HandleFunc("/logout", logoutHandler).Methods("POST")
    router.HandleFunc("/post", postHandler).Methods("POST")

    http.Handle("/", router)
    http.ListenAndServe(":8000", nil)
}

// cookie handling

var cookieHandler = securecookie.New(
    securecookie.GenerateRandomKey(64),
    securecookie.GenerateRandomKey(32))

//gets the username from the cookie data
func getUserName(request *http.Request) (userName string) {
    if cookie, err := request.Cookie("session"); err == nil {
        cookieValue := make(map[string]string)
        if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
            userName = cookieValue["name"]
        }
    }
    return userName
}

//sets the session's username
func setSession(userName string, response http.ResponseWriter) {
    value := map[string]string{
        "name": userName,
    }
    if encoded, err := cookieHandler.Encode("session", value); err == nil {
        cookie := &http.Cookie{
            Name:  "session",
            Value: encoded,
            Path:  "/",
        }
        http.SetCookie(response, cookie)
    }
}

//clear a session
func clearSession(response http.ResponseWriter) {
    cookie := &http.Cookie{
        Name:   "session",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    }
    http.SetCookie(response, cookie)
}

// login handler

//User data object
const newUser = `
<li>%s</li>
`

//handles login and stores the user object in the user list
func loginHandler(response http.ResponseWriter, request *http.Request) {
    name := request.FormValue("name")
    pass := request.FormValue("password")
    redirectTarget := "/"
    if name != "" && pass != "" {
        // .. check credentials ..
        setSession(name, response)

        //add to user list
        body, err := ioutil.ReadFile("users.txt")
        if err != nil {
            http.Error(response, err.Error(), http.StatusInternalServerError)
            return
        }
        output := fmt.Sprintf(newUser,name)

        byteOutput := []byte(output)
        //byteOutput = body + byteOutput
        byteOutput = append(body, byteOutput...)
        ioutil.WriteFile("users.txt", byteOutput, 0666)
        redirectTarget = "/internal"
    }
    http.Redirect(response, request, redirectTarget, 302)
}

// logout handler

//logs out the user and clears the cookie
func logoutHandler(response http.ResponseWriter, request *http.Request) {
    clearSession(response)
    username := getUserName(request)
    replaceString := fmt.Sprintf(newUser,username)
    userBytes, err := ioutil.ReadFile("users.txt")
    if err != nil {
        http.Error(response, err.Error(), http.StatusInternalServerError)
        return
    }

    s := string(userBytes)

    newList := strings.Replace(s,replaceString,"", -1)
    newListBytes := []byte(newList)
    ioutil.WriteFile("users.txt", newListBytes, 0666)
    http.Redirect(response, request, "/", 302)
}

//Post data object
const post = `
<p>%s – %s</p>
`

//Handles the posting of new data to the world data and reloads the page
func postHandler(response http.ResponseWriter, request *http.Request) {
    userName := request.FormValue("username")
    newPost := request.FormValue("newPost")
    redirectTarget := "/internal"
    if userName != "" && newPost != "" {
        // .. check credentials ..
        //add code to write to file
        body, err := ioutil.ReadFile("world.txt")
        if err != nil {
            http.Error(response, err.Error(), http.StatusInternalServerError)
            return
        }

        //appending new Post to world data
        output := fmt.Sprintf(post, newPost,userName)
        byteOutput := []byte(output)
        byteOutput = append(body, byteOutput...)
        ioutil.WriteFile("world.txt", byteOutput, 0666)

        redirectTarget = "/internal"

    }
    http.Redirect(response, request, redirectTarget, 302)
}

// index page

const indexPage = `
<head>
<link href="//netdna.bootstrapcdn.com/twitter-bootstrap/2.3.1/css/bootstrap-combined.min.css" rel="stylesheet">
            <script src="//ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js"></script>
            <script src="//netdna.bootstrapcdn.com/twitter-bootstrap/2.3.1/js/bootstrap.min.js"></script>
</head>
<body>
<div class="container">
<h1>Welcome to this Virtual World</h1>
<form method="post" action="/login">
    <label for="name">User name</label>
    <input type="text" id="name" name="name">
    <label for="password">Password</label>
    <input type="password" id="password" name="password">
    <button type="submit">Login</button>
</form>
</div>
</body>
`

func onboardingPageHandler(response http.ResponseWriter, request *http.Request) {
    fmt.Fprintf(response, indexPage)
}

// internal page

const internalPage = `
<head>
<meta http-equiv="refresh" content="10" />
<link href="//netdna.bootstrapcdn.com/twitter-bootstrap/2.3.1/css/bootstrap-combined.min.css" rel="stylesheet">
            <script src="//ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js"></script>
            <script src="//netdna.bootstrapcdn.com/twitter-bootstrap/2.3.1/js/bootstrap.min.js"></script>
</head>
<body>
<div class="container">
<h1>Virtual World</h1>
<h3>Users</h3>
<div><ul>%s</ul></div>
<hr>
<h3>World</h3>
<div>%s</div>
<form method="post" action="/post">
	<input type="hidden" id="username" name="username" value="%s">
	<input type="text" id="newPost" name="newPost"></input>
	<input type="submit" value="Post to World"></input>
</form>
<hr>
<small>User: %s</small>
<form method="post" action="/logout">
    <button type="submit">Logout</button>
</form>
</div>
</body>
`
//handles loading the world data and displaying it to the user
func inWorldPageHandler(response http.ResponseWriter, request *http.Request) {
    userName := getUserName(request)
    body, err := ioutil.ReadFile("world.txt")
    users,err := ioutil.ReadFile("users.txt");
    if err != nil {
        http.Error(response, err.Error(), http.StatusInternalServerError)
        return
    }
    if userName != "" {
        fmt.Fprintf(response, internalPage, users,body,userName,userName)
    } else {
        http.Redirect(response, request, "/internal", 302)
    }
}
