package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func httpClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	t.TLSHandshakeTimeout = 10 * time.Second
	t.ExpectContinueTimeout = 1 * time.Second

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	}

	return client
}

func main() {
	// Load Env variables from .dot file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	// Allow helm chart tests to pass in a valid way
	testMode := os.Getenv("TEST_MODE")
	if len(testMode) > 0 {
		time.Sleep(1200 * time.Second)
	}

	if len(token) < 1 {
		log.Fatal("Error loading Auth Token.")
	}

	if len(appToken) < 1 {
		log.Fatal("Error loading App Token.")
	}

	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Println("You must specify the client ID and client secret from https://api.slack.com/applications")
		os.Exit(1)
	}

	go func() {
		http.HandleFunc("/add", addToSlack)
		http.HandleFunc("/auth", auth)
		http.HandleFunc("/", home)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	log.Println("Listening on port 8080 for OAuth requests")

	// Load HTTP Client
	httpClient := httpClient()

	// Create a new client to slack by giving token
	// Set debug to true while developing
	// Also add a ApplicationToken option to the client
	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	// go-slack comes with a SocketMode package that we need to use that accepts a Slack client and outputs a Socket mode client instead
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		// Option to set a custom logger
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	// Cron goroutines for handling scheduled announcements in parallel
	mainCron := cron.New(cron.WithLocation(time.UTC))
	mainCron, err = rebuildCron(mainCron, client, httpClient)
	if err != nil {
		log.Fatal(err)
	}
	mainCron.Start()

	// Create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// Make this cancel called properly in a real program , graceful shutdown etc
	defer cancel()

	go func(mainCron *cron.Cron, ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		// Create a for loop that selects either the context cancellation or the events incomming
		for {
			select {
			// inscase context cancel is called exit the goroutine
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:
				// We have a new Events, let's type switch the event
				// Add more use cases here if you want to listen to other events.
				switch event.Type {
				// handle EventAPI events
				case socketmode.EventTypeEventsAPI:
					// The Event sent on the channel is not the same as the EventAPI events so we need to type cast it
					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}
					// We need to send an Acknowledge to the slack server
					socketClient.Ack(*event.Request)
					// Now we have an Events API event, but this event type can in turn be many types, so we actually need another type switch
					err := handleEventMessage(eventsAPIEvent, client, socketClient)
					if err != nil {
						// Replace with actual err handeling
						log.Fatal(err)
					}
				// Handle Slash Commands
				case socketmode.EventTypeSlashCommand:
					// Just like before, type cast to the correct event type, this time a SlashEvent
					command, ok := event.Data.(slack.SlashCommand)
					if !ok {
						log.Printf("Could not type cast the message to a SlashCommand: %v\n", command)
						continue
					}
					// handleSlashCommand will take care of the command
					payload, err := handleSlashCommand(command, client, httpClient)
					if err != nil {
						log.Fatal(err)
					}

					// Dont forget to acknowledge the request
					socketClient.Ack(*event.Request, payload)

				case socketmode.EventTypeInteractive:
					interaction, ok := event.Data.(slack.InteractionCallback)
					if !ok {
						log.Printf("Could not type cast the message to a Interaction callback: %v\n", interaction)
						continue
					}

					err := handleInteractionEvent(mainCron, interaction, client, httpClient)
					if err != nil {
						log.Fatal(err)
					}
					socketClient.Ack(*event.Request)

					//end of switch
				}
			}

		}
	}(mainCron, ctx, client, socketClient)

	err = socketClient.Run()
	if err != nil {
		panic(err)
	}
}
