package postservice

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"soa/common"
	pb "soa/post_service/posts_service/pkg/pb"
	"strconv"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
)

type PostService struct {
	pb.UnimplementedPostServiceServer
	Db      *sql.DB
	Counter uint64 // atomic
}

const dbname = "postgres"
const connectionStringPattern string = "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"

func NewPostService() *PostService {

	host, port, user, password, sslmode, err := common.GetPostgresParams()

	log.Println(host, port, user, password, sslmode)

	if err != nil {
		log.Fatal(err.Error())
	}

	connectStr := fmt.Sprintf(connectionStringPattern,
		host, port, user, password, dbname, sslmode)

	log.Println("Connecting...")

	db, err := sql.Open(dbname, connectStr)

	for i := 0; i < 10; i++ {

		if err == nil {
			break
		}

		log.Println("Connecting...")

		time.Sleep(time.Second * 2)

		db, err = sql.Open(dbname, connectStr)
	}

	if err != nil {
		log.Fatal(err.Error())
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err.Error())
	}

	const query = "" +
		"CREATE TABLE IF NOT EXISTS Posts " +
		"(" +
		"post_id NUMERIC UNIQUE NOT NULL , " +
		"author TEXT NOT NULL, " +
		"headline TEXT NOT NULL, " +
		"content TEXT NOT NULL, " +
		"location TEXT NOT NULL, " +
		"value NUMERIC NOT NULL, " +
		"date_of_creation NUMERIC" +
		")"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err = db.ExecContext(ctx, query)

	if err != nil {
		log.Fatal(err.Error())
	}

	return &PostService{
		Db:      db,
		Counter: 0,
	}
}

func (s *PostService) NewPost(ctx context.Context, post *pb.PostInfo) (*pb.PostID, error) {

	const query = "INSERT INTO Posts (post_id, author, headline, content, location, value, date_of_creation)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7) "

	newId := atomic.AddUint64(&s.Counter, 1)

	_, err := s.Db.Exec(query, newId, post.Author, post.Headline, post.Content, post.Location, post.Value, post.DateOfCreation)

	if err != nil {
		return nil, err
	}

	retId := &pb.PostID{Id: newId}
	return retId, nil
}

func (s *PostService) GetPost(ctx context.Context, id *pb.PostID) (*pb.Post, error) {

	const query = "SELECT post_id, author, headline, content, location, value from Posts WHERE post_id=$1"

	row := s.Db.QueryRow(query, id.Id)

	post := &pb.Post{}

	err := row.Scan(&post.Id, &post.Author, &post.Headline, &post.Content, &post.Location, &post.Value)

	if err != nil {
		log.Println("Row scan error", err.Error())
	}

	return post, err
}

func (s *PostService) GetPosts(ctx context.Context, info *pb.PaginationInfo) (*pb.PostList, error) {

	query := "SELECT post_id, author, headline, content, location, value FROM Posts LIMIT $1 OFFSET $2 "

	// "Date", "Value"
	if info.SortBy == "Date" {
		query += "ORDER BY date_of_creation "
		query += info.Order
	} else if info.SortBy == "Value" { // SortBy == "Value"
		num1, num2, err := common.CheckRange(info.Order)

		if err != nil {
			query += "WHERE value BETWEEN " + strconv.Itoa(num1) + " AND " + strconv.Itoa(num2)
		} else {
			query += "ORDER BY value "
			query += info.Order
		}
	} // else do nothing

	rows, err := s.Db.Query(query, info.BatchSize, info.PageNumber*uint64(info.BatchSize))

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []*pb.Post
	for rows.Next() {
		var id uint64
		var auth_id string
		var content string
		err := rows.Scan(&id, &content, &auth_id)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &pb.Post{Id: id, Content: content, Author: auth_id}) // TODO: more fields
	}

	return &pb.PostList{Posts: posts}, err
}

func (s *PostService) mustEmbedUnimplementedPostServiceServer() {
	//TODO implement me
}
