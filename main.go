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
		fmt.Println("Warning: Error loading .env file")
	}
	// Not a Secret But you can Either set your own VAR "Discord_Client_ID here or in an .env file"
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	if clientID == "" {
		panic("DISCORD_CLIENT_ID is not set")
	}

	err = client.Login(clientID)
	if err != nil {
		panic(err)
	}

	fmt.Println("Figma Discord Rich Presence is running...")

	// Start time for the session
	now := time.Now()

	for {
		// Get the current file from Figma
		filename, err := GetFigmaTitle()
		if err != nil {
			fmt.Println("Error reading Figma title:", err)
		}

		if filename == "" {
			// Figma not found or no file open

			fmt.Println("Figma not detected or no file open.")

			// We can choose to not update or set a default "Idling" state
			// For now, let's just wait and retry.

		} else {
			fmt.Println("Detected:", filename)

			details := "Editing File"
			state := filename

			if filename == "Browsing Files" {
				details = "In Home"
				state = "Browsing Files"
			}

			err = client.SetActivity(client.Activity{
				State:      state,
				Details:    details,
				LargeImage: "largeimageid",
				LargeText:  "Figma",
				Timestamps: &client.Timestamps{
					Start: &now,
				},
			})

			if err != nil {
				fmt.Println("Failed to set activity:", err)
			}
		}

		// Update every 15 seconds (Discord's rate limit is 15s for visual updates usually, but 5s is fine for logic)
		time.Sleep(15 * time.Second)
	}
}
