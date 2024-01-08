package main

import (
	"context"
	"fmt"
	"gifmanager-backend/categories"
	"gifmanager-backend/dal"
	"gifmanager-backend/gifs"
	"gifmanager-backend/groups"
	"gifmanager-backend/httputil"
	"gifmanager-backend/server"
)

func main() {
	ctx := context.Background()
	mongoDal, err := dal.NewMongoDal(ctx, "mongodb://localhost:27017", dal.DbName)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := mongoDal.Disconnect(ctx); err != nil {
			fmt.Println(err)
		}
	}()

	parser := httputil.NewGifsApiQueryParamParser()
	apiGif := gifs.NewGifApi(mongoDal, parser)

	apiGroup := groups.NewGroupApi(mongoDal)
	apiCategory := categories.NewApi(mongoDal, parser)
	s := server.NewServer(mongoDal, apiGif, apiGroup, apiCategory)

	fmt.Println("HTTP SERVER SUCCESSFULLY RUNNING ON PORT 8888")
	if err := s.Run("localhost:8888"); err != nil {
		panic(err)
	}
}
