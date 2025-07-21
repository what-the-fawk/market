package service

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"soa/common"
	"soa/post_service/posts_service/pkg/pb"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MainServiceHandler struct {
	Db         *sql.DB
	JwtPrivate *rsa.PrivateKey
	JwtPublic  *rsa.PublicKey
	client     pb.PostServiceClient
}

const dbname = "postgres"
const connectionStringPattern string = "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"

func CreateMainServiceHandler() *MainServiceHandler {

	host, port, user, password, sslmode, err := common.GetPostgresParams()

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
		"CREATE TABLE IF NOT EXISTS Users " +
		"(" +
		"login VARCHAR (50) UNIQUE NOT NULL, " +
		"password TEXT NOT NULL" +
		")"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err = db.ExecContext(ctx, query)

	if err != nil {
		log.Fatal(err.Error())
	}

	pub, pri, err := common.GetRSAKeys()

	if err != nil {
		log.Println("Rsa keys error")
		log.Fatal(err.Error())
	}

	time.Sleep(20 * time.Second)

	grpcServerAddr, ok := os.LookupEnv("GRPC_SERVER")
	if !ok {
		log.Fatal("GRPC_SERVER not set")
	}
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err.Error())
	}
	grpcClient := pb.NewPostServiceClient(conn)

	if err != nil {
		log.Fatalf("Failed to create topics: %s\n", err.Error())
	}

	return &MainServiceHandler{
		Db:         db,
		JwtPublic:  pub,
		JwtPrivate: pri,
		client:     grpcClient,
	}
}

func (s *MainServiceHandler) Close() {
	s.Db.Close()
}

func (s *MainServiceHandler) Register(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		log.Println("Wrong method in Register")
		http.Error(w, "Registration is allowed only with POST method", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.AuthInfo](req)

	if err != nil {
		log.Println("Json unmarshall error")
		http.Error(w, err.Error(), status)
		return
	}

	if len(info.Login) > 50 || len(info.Password) > 120 {
		log.Println("Either login of password too long")
		http.Error(w, "Shorten your login of password", http.StatusBadRequest)
		return
	}

	const query = "" +
		"INSERT INTO Users " +
		"(login, password) " +
		"VALUES ($1, $2)"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	hasher := sha256.New()
	passwordHash := hex.EncodeToString(hasher.Sum([]byte(info.Password)))

	_, err = s.Db.ExecContext(ctx, query, info.Login, passwordHash)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *MainServiceHandler) Auth(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in Auth")
		http.Error(w, "Authentication is allowed only with GET method", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.AuthInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	const query = "SELECT login, password from Users WHERE login=$1"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	row := s.Db.QueryRowContext(ctx, query, info.Login)

	userQueryInfo := common.AuthInfo{}

	err = row.Scan(&userQueryInfo.Login, &userQueryInfo.Password)

	if err != nil {
		log.Println("Row scan error", err.Error())
		http.Error(w, "Incorrect login", http.StatusNotFound)
		return
	}

	hasher := sha256.New()
	passwordHash := hex.EncodeToString(hasher.Sum([]byte(info.Password)))

	if string(passwordHash[:]) != userQueryInfo.Password {
		log.Println("Incorrect password")
		http.Error(w, "Incorrect password", http.StatusNotFound)
		return
	}

	// token gen
	claims := jwt.MapClaims{
		"iss": info.Login,
		"exp": time.Duration(time.Now().Second()) + time.Duration(1800*time.Second),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(s.JwtPrivate)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: tokenStr,
	})
}

// test function
// check token validity, returns jwt issuer and error
func (s *MainServiceHandler) CheckToken(req *http.Request) (string, error) {

	// nolint:all
	cookie, err := req.Cookie("jwt")

	if err != nil {
		log.Println("No jwt?")
		return "", errors.New("no jwt?")
	}

	tokenStr := cookie.Value

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return s.JwtPublic, nil
	})

	if err != nil {
		log.Println("No token")
		return "", errors.New("no token")
	}

	date, err := token.Claims.GetExpirationTime()

	if err != nil {
		log.Println("No expiration date")
		return "", errors.New("no expiration date")
	}

	if date.Time.Before(time.Now()) {
		log.Println("Expired token")
		return "", errors.New("expired token")
	}

	iss, err := token.Claims.GetIssuer()

	if err != nil {
		return "", errors.New("invalid issuer")
	}

	return iss, nil
}

func (s *MainServiceHandler) CreatePost(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in CreatePost")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	user, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	if len(info.Headline) > 100 || len(info.Content) > 3000 || len(info.Location) > 250 {
		log.Println("Incorrect input size")
		http.Error(w, "Make your info shorter", http.StatusBadRequest)
		return
	}

	extension := strings.ToLower(filepath.Ext(info.Location))[1:]
	supportedFormats := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"webp": true,
		"gif":  true,
		"svg":  true,
	}

	if !supportedFormats[extension] {
		log.Println("Unsupported format")
		http.Error(w, "Unsupported format", http.StatusBadRequest)
		return
	}

	id, err := s.client.NewPost(req.Context(), &pb.PostInfo{
		Author:         user,
		Headline:       info.Headline,
		Content:        info.Content,
		Location:       info.Location,
		Value:          info.Value,
		DateOfCreation: uint64(time.Now().Unix()),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *MainServiceHandler) GetPost(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		log.Println("Wrong method in GetPost")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	_, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostId](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	res, err := s.client.GetPost(req.Context(), &pb.PostID{Id: info.Id})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *MainServiceHandler) GetPostList(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in GetPostList")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	_, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PaginationInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	if info.SortBy != "Date" && info.SortBy != "Value" {
		log.Println("Incorrect sort column:", info.SortBy)
		http.Error(w, "Incorrect sort order", http.StatusBadRequest)
		return
	}

	if info.SortBy == "Date" && info.Order != "ASC" && info.Order != "DESC" {
		log.Println("Incorrect date order")
		http.Error(w, "Incorrect date order", http.StatusBadRequest)
		return
	}

	_, _, err = common.CheckRange(info.Order)

	if info.SortBy == "Value" && info.Order != "ASC" && info.Order != "DESC" && err != nil {
		log.Println("Incorrect value order")
		http.Error(w, "Incorrect value order", http.StatusBadRequest)
		return
	}

	res, err := s.client.GetPosts(req.Context(), &pb.PaginationInfo{
		PageNumber: info.PageNumber,
		BatchSize:  info.BatchSize,
		SortBy:     info.SortBy,
		Order:      info.Order,
	})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
