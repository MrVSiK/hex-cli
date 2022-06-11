package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
)

func toBase64(filePath string) string {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(bytes)
}

func readTokenFromConfigFile(filePath string) string {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(jsonFile)

	config := map[string]string{}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}
	err1 := json.Unmarshal(byteValue, &config)
	if err1 != nil {
		log.Fatal(err1)
	}

	return config["token"]
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var filePath string
	var email string

	client := req.C()

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
					Required:    true,
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

					result := map[string]string{}
					body := map[string]string{}
					headers := map[string]string{}

					body["file"] = base64File
					body["name"] = name

					token := readTokenFromConfigFile("../config.json")
					headers["token"] = token

					fmt.Printf("Uploading %s...\n", name)
					response, err := client.R().SetHeaders(headers).SetBodyJsonMarshal(body).SetResult(&result).Post(os.Getenv("URI") + "/image/")

					if err != nil {
						log.Fatal(err)
					}

					if response.IsSuccess() {
						fmt.Printf("%s\n", result["message"])
					}

					if response.IsError() {
						fmt.Println(response.GetStatusCode())
						fmt.Println(response)
						statusCode := response.GetStatusCode()
						responseBody := map[string]string{}

						byteValue, err := response.ToBytes()
						if err != nil {
							log.Fatal(err)
						}
						err1 := json.Unmarshal(byteValue, &responseBody)
						if err1 != nil {
							return err1
						}
						if statusCode < 500 {
							fmt.Println(responseBody["message"])
						} else {
							fmt.Println("Server error")
						}
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
					Usage:       "set email",
					Destination: &email,
					Required:    true,
				}},
				Action: func(c *cli.Context) error {
					fmt.Printf("Password: ")
					bytePassword, err := term.ReadPassword(int(syscall.Stdin))
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("\n")
					password := strings.TrimSpace(string(bytePassword))

					result := map[string]string{}
					body := map[string]string{}

					body["email"] = email
					body["password"] = password

					response, err := client.R().SetBodyJsonMarshal(body).SetResult(&result).Post(os.Getenv("URI") + "/user/")

					if err != nil {
						log.Fatal(err)
					}

					if response.IsSuccess() {
						file, _ := json.MarshalIndent(result, "", "	")

						_ = ioutil.WriteFile("../config.json", file, 0644)
						fmt.Println("Logged in")
					}

					if response.IsError() {
						statusCode := response.GetStatusCode()
						responseBody := map[string]string{}

						byteValue, err := response.ToBytes()
						if err != nil {
							log.Fatal(err)
						}
						err1 := json.Unmarshal(byteValue, &responseBody)
						if err1 != nil {
							return err1
						}
						if statusCode < 500 {
							fmt.Println(responseBody["message"])
						} else {
							fmt.Println("Server error")
						}
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
