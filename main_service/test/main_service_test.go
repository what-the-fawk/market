package main_service_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"soa/common"
	service "soa/main_service/include"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {

	testCases := []common.AuthInfo{
		{
			Login:    "test1",
			Password: "pwd1",
		},
		{
			Login:    "test2",
			Password: "pwd2",
		},
		{
			Login:    "test3",
			Password: "pwd3",
		},
	}

	for _, tc := range testCases {

		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		hasher := sha256.New()
		passwordHash := hex.EncodeToString(hasher.Sum([]byte(tc.Password)))

		log.Println("Current user: ", tc.Login, " ", tc.Password)

		mock.ExpectExec("INSERT INTO Users").WithArgs(tc.Login, passwordHash).WillReturnResult(sqlmock.NewResult(1, 1))

		ms := &service.MainServiceHandler{Db: db}
		handler := http.HandlerFunc(ms.Register)

		auth_data, err := json.Marshal(&tc)

		if err != nil {
			log.Fatal(err.Error())
		}

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(auth_data))

		if err != nil {
			log.Fatal(err.Error())
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestAuth(t *testing.T) {
	testCases := []common.AuthInfo{
		{
			Login:    "test1",
			Password: "pwd1",
		},
		{
			Login:    "test2",
			Password: "pwd2",
		},
		{
			Login:    "test3",
			Password: "pwd3",
		},
	}

	for _, tc := range testCases {

		private, err := rsa.GenerateKey(rand.Reader, 128)

		if err != nil {
			log.Fatal(err.Error())
		}

		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		hasher := sha256.New()
		passwordHash := hex.EncodeToString(hasher.Sum([]byte(tc.Password)))

		log.Println("Current user: ", tc.Login, " ", tc.Password)

		mock.ExpectQuery("SELECT").WithArgs(tc.Login).WillReturnRows(sqlmock.NewRows([]string{"login", "password"}).AddRow(tc.Login, passwordHash))

		ms := &service.MainServiceHandler{Db: db, JwtPrivate: private, JwtPublic: &private.PublicKey}
		handler := http.HandlerFunc(ms.Auth)

		auth_data, err := json.Marshal(&tc)

		if err != nil {
			log.Fatal(err.Error())
		}

		req, err := http.NewRequest(http.MethodPost, "/auth", bytes.NewReader(auth_data))

		if err != nil {
			log.Fatal(err.Error())
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}
