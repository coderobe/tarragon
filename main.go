package main

import (
	"log"
	"strings"
)

var broker *Broker
var instance *Instance

func main() {
	log.Println("Hello, world!")

	broker = NewBroker("localhost:1234")
	broker.Root().AddUser(NewUser("admin").SetPassword("1234"))
	go broker.ListenAndServe()

	instance = NewInstance("localhost:1234", false)

	Command("instance login --username=foo -password bar")

	log.Fatal(instance.ConnectAndRecv())
}

func Command(command string) {
	cmd := strings.Fields(command)

	data := make(map[string]string)
	var verbs []string
	setting := ""

	for _, field := range cmd {
		if len(setting) > 0 {
			data[setting] = field
			setting = ""
		} else if strings.HasPrefix(field, "-") {
			set := strings.Split(strings.TrimLeft(field, "-"), "=")
			if len(set) > 1 {
				data[set[0]] = set[1]
			} else {
				setting = set[0]
			}
		} else {
			verbs = append(verbs, field)
		}
	}

	log.Printf("Command %v, Options %v\n", verbs, data)

	verb := ""
	if len(verbs) > 0 {
		switch verb, verbs = verbs[0], verbs[1:]; verb {
		case "instance":
			if len(verbs) > 0 {
				switch verb, verbs = verbs[0], verbs[1:]; verb {
				case "login":
					if user, ok := data["username"]; ok {
						if pw, ok := data["password"]; ok {
							// TODO login
							log.Printf("login %s %s\n", user, pw)
						} else {
							log.Printf("Missing --password in %v\n", cmd)
						}
					} else {
						log.Printf("Missing --username in %v\n", verbs)
					}
				case "auth":
					// TODO
				default:
					log.Printf("unhandled verb")
				}
			}
		default:
			log.Printf("unhandled verb '%s'\n", verb)
		}
	}
}
