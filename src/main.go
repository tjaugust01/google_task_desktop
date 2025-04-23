package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

func main() {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Konfiguration des OAuth2-Clients
	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Erstellen eines OAuth2-Clients mit dem Config und einem http.Client
	client := getClient(config)

	// Erstellen eines Google Tasks API Service
	srv, err := tasks.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Tasks client: %v", err)
	}

	fmt.Println("Google Tasks API Client erfolgreich erstellt!")

	fmt.Println("\n--- Deine Aufgabenlisten ---")
	tasklists, err := srv.Tasklists.List().MaxResults(10).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve tasklists: %v", err)
	}

	if len(tasklists.Items) == 0 {
		fmt.Println("Keine Aufgabenlisten gefunden.")
	} else {
		for _, tl := range tasklists.Items {
			fmt.Printf("- %s (%s)\n", tl.Title, tl.Id)

			fmt.Printf("  --- Aufgaben in '%s' ---\n", tl.Title)
			tasksInList, err := srv.Tasks.List(tl.Id).MaxResults(20).Do()
			if err != nil {
				log.Printf("Fehler beim Abrufen der Aufgaben für Liste '%s': %v", tl.Title, err)
				continue
			}

			if len(tasksInList.Items) == 0 {
				fmt.Println("    Keine Aufgaben in dieser Liste.")
			} else {
				for _, task := range tasksInList.Items {
					fmt.Printf("    - [%s] %s\n", statusSymbol(task.Status), task.Title)
				}
			}
			fmt.Println()
		}
	}
}

// getClient ruft ein Token vom Web ab, wenn es nicht zwischengespeichert ist,
// und gibt andernfalls den zwischengespeicherten Token zurück.
func getClient(config *oauth2.Config) *http.Client {
	// Das Token wird normalerweise in einer Datei zwischengespeichert.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(oauth2.NoContext, tok)
}

// getTokenFromWeb ruft ein Token vom Web ab.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Öffne den folgenden Link in deinem Browser und gib den Autorisierungscode ein: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// tokenFromFile liest ein Token aus der Token-Datei.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken speichert das Token in einer Datei.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
func statusSymbol(status string) string {
	if status == "completed" {
		return "X"
	}
	return " "
}
