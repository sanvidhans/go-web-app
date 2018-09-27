package main

import (
	"net/http"
	"fmt"
)

func handlerFunc(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "<h1>Welcome to my Site</h1>")
}

func main(){
	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":3000", nil)
}