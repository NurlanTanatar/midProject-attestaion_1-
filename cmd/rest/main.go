package main

import (
	"context"
	"log"
	"midProject/internal/http"
	"midProject/internal/store/inmemory"
)

func main() {
	store := inmemory.Init()

	srv := http.NewServer(context.Background(), ":8080", store)
	if err := srv.Run(); err != nil {
		log.Println(err)
	}

	srv.WaitForGracefulTermination()

}
