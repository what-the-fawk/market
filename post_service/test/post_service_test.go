package post_service_test

import (
	"context"
	"log"
	postservice "soa/post_service/include"
	"soa/post_service/posts_service/pkg/pb"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {

	testCases := []pb.PostInfo{
		{
			Author:         "test1",
			Headline:       "Headline 1",
			Content:        "Imagine",
			Location:       "Nowhere",
			Value:          100,
			DateOfCreation: 120,
		},
		{
			Author:         "test1",
			Headline:       "Headline 2",
			Content:        "Imagine 2",
			Location:       "Nowhere 2",
			Value:          102,
			DateOfCreation: 122,
		},
	}

	for _, tc := range testCases {
		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		serv := &postservice.PostService{
			Db:      db,
			Counter: 0,
		}

		mock.ExpectExec("INSERT INTO").WithArgs(1, tc.Author, tc.Headline, tc.Content, tc.Location, tc.Value, tc.DateOfCreation).WillReturnResult(sqlmock.NewResult(1, 1))

		res, err := serv.NewPost(context.Background(), &tc)

		if err != nil {
			log.Fatal(err.Error())
		}

		assert.Equal(t, res.Id, uint64(1))

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}
