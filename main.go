package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
)

type Config struct {
	token string
}

type FileDetails struct {
	file string
	name string
}

type FileUploadSuccess struct {
	message string
}

type LoginDetails struct {
	email    string
	password string
}

type LoginSuccess struct {
	token string
}

func toBase64(filePath string) string {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(bytes)
}

func main() {
	var filePath string
	var email string

	client := req.C()
	jsonFile, err := os.Open("../config.json")
	if err != nil {
		log.Fatal(err)
	}

	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(jsonFile)

	var config Config

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}
	err1 := json.Unmarshal(byteValue, &config)
	if err1 != nil {
		log.Fatal(err1)
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "upload",
				Usage: "upload a file to your cloud storage",
				Flags: []cli.Flag{&cli.StringFlag{
					Name:        "file",
					Aliases:     []string{"f"},
					Usage:       "Load configuration from `FILEPATH`",
					Destination: &filePath,
				}},
				Action: func(c *cli.Context) error {
					file, err := os.Stat(filePath)

					if err != nil {
						log.Fatal(err)
					}

					size := file.Size()
					name := file.Name()

					fmt.Printf("File size: %d bytes\n", size)

					base64File := toBase64(filePath)

					var result FileUploadSuccess

					fmt.Printf("Uploading %s...\n", name)
					response, err := client.R().SetBody(&FileDetails{file: base64File, name: name}).SetResult(&result).SetHeader("token", config.token).Post("http://localhost:3000/image/")

					if err != nil {
						log.Fatal(err)
					}

					if response.IsSuccess() {
						fmt.Printf("%s\n", result.message)
					}
					return nil
				},
			},
			{
				Name:  "login",
				Usage: "login using email and password",
				Flags: []cli.Flag{&cli.StringFlag{
					Name:        "email",
					Aliases:     []string{"e"},
					Usage:       "email",
					Destination: &email,
				}},
				Action: func(c *cli.Context) error {
					fmt.Printf("Password: ")
					bytePassword, err := term.ReadPassword(int(syscall.Stdin))
					if err != nil {
						log.Fatal(err)
					}
					password := strings.TrimSpace(string(bytePassword))

					var result LoginSuccess

					response, err := client.R().SetBody(&LoginDetails{email: email, password: password}).SetResult(&result).Post("http://localhost:3000/user/")

					if err != nil {
						log.Fatal(err)
					}

					if response.IsSuccess() {
						file, _ := json.MarshalIndent(result, "", "	")

						_ = ioutil.WriteFile("../config.json", file, 0644)
						fmt.Println()
						fmt.Println("Logged in")
					}
					return nil
				},
			},
		},
	}

	err3 := app.Run(os.Args)
	if err3 != nil {
		log.Fatal(err)
	}
}
