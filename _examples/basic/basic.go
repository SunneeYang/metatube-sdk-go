package main

import (
	"fmt"
	"github.com/metatube-community/metatube-sdk-go/provider/fc2hub"
	"log"

	"github.com/metatube-community/metatube-sdk-go/engine"
)

func main() {
	app := engine.Default()

	results, err := app.SearchMovie(
		"FC2-PPV-1303384",
		fc2hub.Name,
		true,
	)
	//results, err := app.SearchMovieAll("FC2 PPV 1508978 – ロリ体型の後輩に生挿入してみたらキツキツ未開発まんこだった件.mp4", false)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		fmt.Println(result.Provider, result.ID, result.Number, result.Title)
	}
}
