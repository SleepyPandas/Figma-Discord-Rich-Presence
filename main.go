package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hugolgst/rich-go/client"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	
	if err != nil {
		panic("Error loading .env file")
	}

	err = client.Login(os.Getenv("DISCORD_CLIENT_ID"))

	if err != nil {
		panic(err)
	}

	now := time.Now()
	err = client.SetActivity(client.Activity{
		State:      "Ligma",
		Details:    "This GOOOOOOOOO GO LANGGGGGGGGGGGGGGGGGGGG :)",
		LargeImage: "largeimageid",
		LargeText:  "This is the large image :D",
		SmallImage: "smallimageid",
		SmallText:  "And this is the small image",
		Party: &client.Party{
			ID:         "-1",
			Players:    12,
			MaxPlayers: 9000,
		},
		Timestamps: &client.Timestamps{
			Start: &now,
		},
		Buttons: []*client.Button{
			&client.Button{
				Label: "GitHub",
				Url:   "https://github.com/hugolgst/rich-go",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	// Discord will only show the presence if the app is running
	// Sleep for a few seconds to see the update
	fmt.Println("Sleeping...")
	time.Sleep(time.Second * 10000)

	for i := 0; i < 2; i++ {
		main()
	}

}
