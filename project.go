package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("mysession"))

var db *sql.DB

var err error

func CreateAconut(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "static/templates/signUP.html")
		return

	}

	username := req.FormValue("Name")
	email := req.FormValue("email")
	password := req.FormValue("Password")

	var user string
	err := db.QueryRow("SELECT Name FROM productdb.Products WHERE Name=?", username).Scan(&user)
	switch {
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		_, err = db.Exec("insert into productdb.Products (Name, Email, Password) values(?,?,?)", username, email, hashedPassword)

		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}
		res.Write([]byte("User created!"))

		return

	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 301)
	}
}

func loginPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "static/templates/signIn.html")
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	var databaseUsername string
	var databasePassword string

	err := db.QueryRow("SELECT Name, Password FROM  productdb.Products  WHERE Name=?", username).Scan(&databaseUsername, &databasePassword)

	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	res.Write([]byte("Hello" + databaseUsername))

}

func Logout(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")
	session.Options.MaxAge = -1
	session.Save(request, response)
	http.Redirect(response, request, "/loginIndex", http.StatusSeeOther)
}

type Article struct {
	id                     uint16
	Title, Anons, FullText string
}

var posts = []Article{}

func index(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("static/templates/index.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	res, err := db.Query("Select * from `articles`")

	if err != nil {
		panic(err)
	}
	posts = []Article{}
	for res.Next() {
		var post Article
		err = res.Scan(&post.id, &post.Title, &post.Anons, &post.FullText)
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Sprintf("Post: %s with id : %d", post.Title, post.id))

		posts = append(posts, post)
	}

	t.ExecuteTemplate(w, "index", posts)
}
func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/templates/create.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	t.Execute(w, nil)
}
func save_article(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	anons := r.FormValue("anons")
	full_text := r.FormValue("full_text")
	if title == "" || anons == "" || full_text == "" {
		fmt.Fprintf(w, "write full data")
	} else {

		insert, err := db.Query(fmt.Sprintf("INSERT INTO `articles`(`title`, `anons`, `full_text`) VALUES('%s', '%s', '%s')", title, anons, full_text))

		if err != nil {
			panic(err)
		}
		defer insert.Close()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func handleRequest() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/signup", CreateAconut)
	http.HandleFunc("/logout", Logout)
	http.HandleFunc("/", index)
	http.HandleFunc("/create", create)
	http.HandleFunc("/save_article", save_article)

	http.ListenAndServe(":8080", nil)
}

func main() {
	db, err = sql.Open("mysql", "root:root@/productdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("database is conecting port localhost:8080")

	handleRequest()

}
