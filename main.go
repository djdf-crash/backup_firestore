package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"errors"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/urfave/cli/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func getClient(ctx context.Context, serviceAccountPath string) (*firestore.Client, error) {
	sa := option.WithCredentialsFile(serviceAccountPath)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func saveDocuments(ctx context.Context, client *firestore.Client, collections []string) error {
	for _, collection := range collections {
		var allDocs []map[string]map[string]interface{}
		iter := client.Collection(collection).Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			nameDoc := strings.Split(doc.Ref.Path, "/")
			item := map[string]map[string]interface{}{
				nameDoc[len(nameDoc)-1]: doc.Data(),
			}
			allDocs = append(allDocs, item)
		}
		if len(allDocs) == 0 {
			return errors.New("No found documents in collection " + collection)
		}
		data, err := json.MarshalIndent(allDocs, "", " ")
		if err != nil {
			return err
		}
		nameFile := collection + ".json"
		err = ioutil.WriteFile(nameFile, data, 0644)
		if err != nil {
			return err
		}
		currentDir, _ := os.Getwd()
		fullPath := currentDir + string(os.PathSeparator) + nameFile
		fmt.Println("Saved collection " + collection + " to " + fullPath)
	}
	return nil
}

func main() {

	app := &cli.App{
		Name:  "Backup firestore",
		Usage: "Backup firestore database collections in firebase!",
		Commands: []*cli.Command{
			{
				Name:  "backup",
				Usage: "Backup firestore",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "collections",
						Aliases:  []string{"c"},
						Usage:    "Choice collections for backup",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Usage:    "Path to file *-firebase-adminsdk-*.json",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					col := c.StringSlice("collections")
					if len(col) == 0 {
						return errors.New("Set collections!")
					}
					serviceAccountPath := c.String("path")
					ctx := context.Background()
					client, err := getClient(ctx, serviceAccountPath)
					if err != nil {
						return err
					}
					defer client.Close()
					err = saveDocuments(ctx, client, col)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

	// start our application
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
