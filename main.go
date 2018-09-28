package main

import (
	"net/http"
	"html/template"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var templates *template.Template
var client *redis.Client
var store = sessions.NewCookieStore([]byte("top-secret"))

func main(){
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	templates = template.Must(template.ParseGlob("templates/*.html"))
	fs  := http.FileServer(http.Dir("./static"))
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	r.HandleFunc("/", AuthRequired(indexHandler)).Methods("GET")
	r.HandleFunc("/", indexPostHandler).Methods("POST")

	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/login", loginPostHandler).Methods("POST")


	r.HandleFunc("/register", registerHandler).Methods("GET")
	r.HandleFunc("/register", registerPostHandler).Methods("POST")

	http.ListenAndServe(":8080", r)
}

func AuthRequired(handler http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){

		session, _ := store.Get(r,"session")
		_, ok := session.Values["username"]
		if !ok {
			http.Redirect(w, r, "/login", 302)
			return
		}

		handler.ServeHTTP(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request){
	comments, err := client.LRange("comments", 0, 10).Result()
	if err != nil{
		return
	}
	templates.ExecuteTemplate(w, "index.html", comments)
}

func indexPostHandler(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	comment := r.PostForm.Get("comment")
	client.LPush("comments", comment)
	http.Redirect(w, r,"/", 302)

}

func loginHandler(w http.ResponseWriter, r *http.Request){
	templates.ExecuteTemplate(w, "login.html", nil)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	hash, err := client.Get("user:"+username).Bytes()

	if err == redis.Nil{
		templates.ExecuteTemplate(w, "login.html", "Unknown User")
		return
	}else if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil{
		templates.ExecuteTemplate(w, "login.html", "Username is wrong!")
		return
	}
	session, _ := store.Get(r,"session")
	session.Values["username"] = username
	session.Save(r, w)
	http.Redirect(w, r,"/", 302)
}

func registerHandler(w http.ResponseWriter, r *http.Request){
	templates.ExecuteTemplate(w, "register.html", nil)
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	cost := bcrypt.DefaultCost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	err = client.Set("user:"+username, hash, 0).Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	http.Redirect(w, r,"/login", 302)
}