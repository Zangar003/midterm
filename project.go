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

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/templates/index.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	t.Execute(w, nil)
}
func handleRequest() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/signup", CreateAconut)
	http.HandleFunc("/logout", Logout)
	http.HandleFunc("/", index)

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
