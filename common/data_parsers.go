package common

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func GetPostgresParams() (string, int, string, string, string, error) {

	host := os.Getenv("POSTGR_HOST")
	port, err := strconv.Atoi(os.Getenv("POSTGR_PORT"))
	user := os.Getenv("POSTGR_USR")
	password := os.Getenv("POSTGR_PASSWORD")
	security := os.Getenv("POSTGR_SSL")

	return host, port, user, password, security, err
}

func GetRSAKeys() (*rsa.PublicKey, *rsa.PrivateKey, error) {

	pubFile, priFile := os.Getenv("RSA_PUB"), os.Getenv("RSA_PRI")

	if len(pubFile) == 0 {
		log.Fatal("Empty pub file")
	}

	pub, err := os.ReadFile(pubFile)

	if err != nil {
		log.Println("Public key file failed")
		return nil, nil, err
	}

	pri, err := os.ReadFile(priFile)

	if err != nil {
		log.Println("Private key file failed")
		return nil, nil, err
	}

	pub_key, err := jwt.ParseRSAPublicKeyFromPEM(pub)

	if err != nil {
		log.Println("Error in pub parsing")
		return nil, nil, err
	}

	pri_key, err := jwt.ParseRSAPrivateKeyFromPEM(pri)

	if err != nil {
		return nil, nil, err
	}

	return pub_key, pri_key, nil

}

func GetJsonStruct[T UserInfo | NewUserInfo | AuthInfo | PostInfo | PostId | PaginationInfo | ReactionInfo | PostIsLike](req *http.Request) (*T, int, error) {

	body := make([]byte, req.ContentLength)
	read, err := req.Body.Read(body)
	defer req.Body.Close()

	if read != int(req.ContentLength) {
		//status = http.StatusBadRequest
		return nil, http.StatusBadRequest, err
	}

	var info T

	err = json.Unmarshal(body, &info)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &info, http.StatusOK, nil
}

func CheckRange(input string) (int, int, error) {
	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid range format")
	}

	part1 := strings.TrimSpace(parts[0])
	part2 := strings.TrimSpace(parts[1])

	num1, err := strconv.Atoi(part1)
	if err != nil {
		return 0, 0, err
	}

	num2, err := strconv.Atoi(part2)
	if err != nil {
		return 0, 0, err
	}

	if num1 > num2 {
		return 0, 0, errors.New("invalid range")
	}

	return num1, num2, nil
}
