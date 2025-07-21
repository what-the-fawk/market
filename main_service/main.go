package main

import (
	"log"
	"net/http"
	service "soa/main_service/include"
	"time"
)

func main() {

	time.Sleep(time.Second * 5) // wait for DB

	serv := service.CreateMainServiceHandler()
	log.Println("created")

	http.HandleFunc("/register", serv.Register)
	log.Println("Register")

	http.HandleFunc("/auth", serv.Auth)
	log.Println("Auth")

	http.HandleFunc("/create", serv.CreatePost)
	log.Println("Create")

	http.HandleFunc("/getPost", serv.GetPost)
	log.Println("Get post")

	http.HandleFunc("/list", serv.GetPostList)
	log.Println("Get post list")

	err := http.ListenAndServe(":3333", nil)

	if err != nil {
		log.Println(err.Error())
	}

	defer serv.Close()
}
