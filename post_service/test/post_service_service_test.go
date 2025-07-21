package post_service_test

import (
	"context"
	"os/exec"
	"soa/post_service/posts_service/pkg/pb"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPostServiceRPCGet(t *testing.T) {

	cmd1 := exec.Command("docker", "compose", "up")
	cmd2 := exec.Command("docker", "compose", "down")
	defer cmd2.Run()
	err := cmd1.Run()

	if err != nil {
		t.Fatal(err)
	}

	conn, err := grpc.Dial("localhost:6666", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	client := pb.NewPostServiceClient(conn)

	req_create := &pb.PostInfo{
		Author:   "test_pixie",
		Headline: "Headline 1",
		Content:  "test content",
		Location: "somewhere",
		Value:    69,
	}

	post_id, err := client.NewPost(context.Background(), req_create)

	if err != nil {
		t.Fatal(err.Error())
	}

	post, err := client.GetPost(context.Background(), post_id)

	if err != nil {
		t.Fatal(err.Error())
	}

	if post.Author != req_create.Author {
		t.Fatal("Author incorrect")
	}

	if post.Headline != req_create.Headline {
		t.Fatal("Comment section ID incorrect")
	}

	if post.Content != req_create.Content {
		t.Fatal("Content incorrect")
	}

	if post.Value != req_create.Value {
		t.Fatal("Date of creation incorrect")
	}

	if post.Id != post_id.Id {
		t.Fatal("Incorrect PostID generation")
	}

	if post.Location != req_create.Location {
		t.Fatal("Incorrect location")
	}
}
