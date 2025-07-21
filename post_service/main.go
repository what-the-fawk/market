package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	post_service "soa/post_service/include"
	"soa/post_service/posts_service/pkg/pb"
	"time"
)

func main() {

	time.Sleep(time.Second * 5) // wait for DB

	lis, err := net.Listen("tcp", ":6666")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServ := post_service.NewPostService()
	log.Println("created")

	newServ := grpc.NewServer()
	pb.RegisterPostServiceServer(newServ, grpcServ)

	err = newServ.Serve(lis)

	if err != nil {
		log.Println(err.Error())
	}

	//defer serv.Close()
}
