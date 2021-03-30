package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	prompt "github.com/c-bata/go-prompt"
	. "github.com/logrusorgru/aurora/v3"
	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

var namespace string
var broker *Broker
var instance *Instance
var brokers map[string]*Broker
var instances map[string]*Instance
var brokername string
var instancename string

func nameBroker(name string) error {
	if _, exists := brokers[name]; exists {
		return errors.New("Broker name in use")
	}
	brokers[name] = broker
	brokername = name
	return nil
}

func selectBroker(name string) error {
	if b, ok := brokers[name]; ok {
		broker = b
		brokername = name
		return nil
	}
	return errors.New("Broker not found")
}

func nameInstance(name string) error {
	if _, exists := instances[name]; exists {
		return errors.New("Instance name in use")
	}
	instances[name] = instance
	instancename = name
	return nil
}

func selectInstance(name string) error {
	if i, ok := instances[name]; ok {
		instance = i
		instancename = name
		return nil
	}
	return errors.New("Instance not found")
}

func Name() string {
	if namespace == "broker" {
		return "estragon"
	}
	return "tarragon"
}

type CTree struct {
	Help     string
	Branches map[string]CTree
	Leaves   map[string]CLeaf
}

type CLeaf struct {
	Help    string
	Options COpthelp
	Trigger func(COption)
}

type COpthelp map[string]string
type COption func(string) (string, bool)

var Commands CTree
var data map[string]string

func init() {
	brokers = make(map[string]*Broker)
	instances = make(map[string]*Instance)
	brokername = ""
	instancename = ""

	Commands = CTree{
		Branches: map[string]CTree{
			"instance": CTree{
				Help: "Subcommands for instance configuration ('client' mode)",
				Leaves: map[string]CLeaf{
					"alias": CLeaf{
						Help:    "Assign an alias to the currently selected instance",
						Options: COpthelp{"name": "name to use for this instance"},
						Trigger: func(option COption) {
							if name, ok := option("name"); ok {
								if err := nameInstance(name); err != nil {
									log.Println(err)
								}
							}
						},
					},
					"select": CLeaf{
						Help:    "Select an instance by alias",
						Options: COpthelp{"name": "named alias to select"},
						Trigger: func(option COption) {
							if name, ok := option("name"); ok {
								if err := selectInstance(name); err != nil {
									log.Println(err)
								}
							}
						},
					},
					"state": CLeaf{
						Help: "Display current instance state",
						Trigger: func(option COption) {
							instance.State().PrettyPrint()
						},
					},
					"connect": CLeaf{
						Help: "Connect to a broker",
						Options: COpthelp{
							"broker":   "Broker address, i.e. '127.0.0.1:42069'",
							"insecure": "Set to any value to force insecure connection (do not use in prod)",
						},
						Trigger: func(option COption) {
							if brokerAddr, ok := option("broker"); ok {
								if _, ok := data["insecure"]; ok {
									instance = NewInstance(brokerAddr, false)
								} else {
									instance = NewInstance(brokerAddr, true)
								}
								go instance.ConnectAndRecv()
							}
						},
					},
					"login": CLeaf{
						Help: "Log in to broker account (not required for basic endpoint functionality)",
						Options: COpthelp{
							"username": "Account username",
							"password": "Account password",
						},
						Trigger: func(option COption) {
							if user, ok := option("username"); ok {
								if pw, ok := option("password"); ok {
									if err := instance.Login(user, pw); err != nil {
										log.Fatal(err)
									}
								}
							}
						},
					},
					"logoff": CLeaf{
						Help: "Log out of broker account (does not deauthenticate)",
						Trigger: func(option COption) {
							if err := instance.Logoff(); err != nil {
								log.Fatal(err)
							}
						},
					},
					"deauth": CLeaf{
						Help: "Deauthenticate the current session, this disconnects your endpoint",
						Trigger: func(option COption) {
							if err := instance.Deauth(); err != nil {
								log.Fatal(err)
							}
						},
					},
					"auth": CLeaf{
						Help:    "Authenticate the current session, required to connect as endpoint",
						Options: COpthelp{"token": "Authentication token, see 'instance token'"},
						Trigger: func(option COption) {
							if token, ok := option("token"); ok {
								if err := instance.Auth(token); err != nil {
									log.Fatal(err)
								}
							}
						},
					},
					"identify": CLeaf{
						Help:    "Identify to the network, configures the current session as endpoint",
						Options: COpthelp{"name": "Endpoint hostname to use, has to be unique"},
						Trigger: func(option COption) {
							if name, ok := option("name"); ok {
								if err := instance.Identify(name); err != nil {
									log.Fatal(err)
								}
							}
						},
					},
				},
				Branches: map[string]CTree{
					"token": CTree{
						Help: "Authentication token management",
						Leaves: map[string]CLeaf{
							"new": CLeaf{
								Help: "Generate a new authentication token for your current user",
								Trigger: func(option COption) {
									if token, err := instance.NewAuthToken(); err == nil {
										log.Printf("New auth token: %s\n", token)
										instance.Auth(token) // TODO remove
									} else {
										log.Fatal(err)
									}
								},
							},
							"delete": CLeaf{
								Help:    "Revoke a previously generated authentication token",
								Options: COpthelp{"token": "Authentication token to revoke"},
								Trigger: func(option COption) {
									if token, ok := option("token"); ok {
										if err := instance.DeleteAuthToken(token); err == nil {
											log.Printf("Token %s deleted\n", token)
										} else {
											log.Fatal(err)
										}
									}
								},
							},
						},
					},
				},
			},
			"broker": CTree{
				Help: "Subcommands for broker configuration ('server' mode)",
				Leaves: map[string]CLeaf{
					"alias": CLeaf{
						Help:    "Assign an alias to the currently selected broker",
						Options: COpthelp{"name": "name to use for this broker"},
						Trigger: func(option COption) {
							if name, ok := option("name"); ok {
								if err := nameBroker(name); err != nil {
									log.Println(err)
								}
							}
						},
					},
					"select": CLeaf{
						Help:    "Select a broker by alias",
						Options: COpthelp{"name": "named broker to select"},
						Trigger: func(option COption) {
							if name, ok := option("name"); ok {
								if err := selectBroker(name); err != nil {
									log.Println(err)
								}
							}
						},
					},
					"state": CLeaf{
						Help: "Display current broker state",
						Trigger: func(option COption) {
							broker.State().PrettyPrint()
						},
					},
					"statuspage": CLeaf{
						Help: "Enable optional http status page on listen address",
						Trigger: func(option COption) {
							broker.HandleStatus()
						},
					},
					"listen": CLeaf{
						Help:    "Start broker",
						Options: COpthelp{"address": "listen address, i.e. '127.0.0.1:42069'"},
						Trigger: func(option COption) {
							if listenAddr, ok := option("address"); ok {
								broker = NewBroker(listenAddr)
								go broker.ListenAndServe()
								time.Sleep(100 * time.Millisecond) // without this, when scripting broker & instance in same process, instance may be faster than broker and fail to connect
							}
						},
					},
				},
				Branches: map[string]CTree{
					"group": CTree{
						Help: "Administrative group management",
						Leaves: map[string]CLeaf{
							"add": CLeaf{
								Help: "Add new endpoint group",
								Options: COpthelp{
									"name":  "Group name, has to be unique",
									"owner": "Group owner",
								},
								Trigger: func(option COption) {
									if name, ok := option("name"); ok {
										if owner, ok := option("owner"); ok {
											if ownerUser, err := broker.State().GetUser(owner); err == nil {
												if _, err := broker.State().NewGroup(name, ownerUser); err != nil {
													log.Println(err)
												}
											} else {
												log.Printf("Owner user %v not found\n", owner)
											}
										}
									}
								},
							},
							"remove": CLeaf{
								Help:    "Remove endpoint group",
								Options: COpthelp{"name": "Group name"},
								Trigger: func(option COption) {
									if name, ok := option("name"); ok {
										if group, err := broker.State().GetGroup(name); err == nil {
											broker.State().RemoveGroup(group)
										} else {
											log.Printf("Group %s does not exist\n", name)
										}
									}
								},
							},
						},
					},
					"user": CTree{
						Help: "Administrative user management",
						Leaves: map[string]CLeaf{
							"add": CLeaf{
								Help: "Add new user",
								Options: COpthelp{
									"username": "User name, has to be unique",
									"password": "User password",
								},
								Trigger: func(option COption) {
									if username, ok := option("username"); ok {
										if password, ok := option("password"); ok {
											if user, err := broker.State().NewUser(username); err == nil {
												user.SetPassword(password)
											} else {
												log.Println(err)
											}
										}
									}
								},
							},
							"remove": CLeaf{
								Help:    "Remove user",
								Options: COpthelp{"username": "User name of target user"},
								Trigger: func(option COption) {
									if username, ok := option("username"); ok {
										if user, err := broker.State().GetUser(username); err == nil {
											broker.State().RemoveUser(user)
											log.Printf("User %s deleted\n", username)
										} else {
											log.Printf("User %s does not exist\n", username)
										}
									}
								},
							},
							"chpw": CLeaf{
								Help: "Change password",
								Options: COpthelp{
									"username": "User name of target user",
									"password": "New password",
								},
								Trigger: func(option COption) {
									if username, ok := option("username"); ok {
										if password, ok := option("password"); ok {
											if user, err := broker.State().GetUser(username); err == nil {
												user.SetPassword(password)
												log.Printf("Password for user %s changed\n", username)
											} else {
												log.Printf("User %s does not exist\n", username)
											}
										}
									}
								},
							},
							"list": CLeaf{
								Help: "List users",
								Trigger: func(option COption) {
									log.Printf("Users:")
									for _, user := range broker.State().Users() {
										var machines []string
										for _, box := range user.Endpoints() {
											machines = append(machines, box.Name())
										}
										var groups []string
										for _, group := range broker.State().GetUserGroups(user) {
											groups = append(groups, group.Name())
										}
										log.Printf("\t%v\tendpoints: %v\tgroups: %v\n", user.Name(), machines, groups)
									}
								},
							},
						},
						Branches: map[string]CTree{
							"group": CTree{
								Help: "Manage group memberships",
								Leaves: map[string]CLeaf{
									"add": CLeaf{
										Help: "Add user to group",
										Options: COpthelp{
											"username": "User to add",
											"group":    "Group to add user to",
										},
										Trigger: func(option COption) {
											if username, ok := option("username"); ok {
												if groupname, ok := option("group"); ok {
													if group, err := broker.State().GetGroup(groupname); err == nil {
														if user, err := broker.State().GetUser(username); err == nil {
															group.AddGroup(user.Group())
															log.Printf("User %s added to group %s\n", username, groupname)
														} else {
															log.Printf("User %s does not exist\n", username)
														}
													} else {
														log.Printf("Group %s does not exist\n", groupname)
													}
												}
											}
										},
									},
									"remove": CLeaf{
										Help: "Remove user from group",
										Options: COpthelp{
											"username": "User to remove",
											"group":    "Group to remove user from",
										},
										Trigger: func(option COption) {
											if username, ok := option("username"); ok {
												if groupname, ok := option("group"); ok {
													if group, err := broker.State().GetGroup(groupname); err == nil {
														if user, err := broker.State().GetUser(username); err == nil {
															group.RemoveGroup(user.Group())
															log.Printf("User %s removed from group %s\n", username, groupname)
														} else {
															log.Printf("User %s does not exist\n", username)
														}
													} else {
														log.Printf("Group %s does not exist\n", groupname)
													}
												}
											}
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func tabComplete(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}

	cmdset := Commands
	cmd := strings.Fields(strings.Split(d.TextBeforeCursor(), " -")[0])
	for _, subcommand := range cmd {
		if def, ok := cmdset.Branches[subcommand]; ok {
			cmdset = def
		}
	}
	for bname, b := range cmdset.Branches {
		s = append(s, prompt.Suggest{Text: bname, Description: b.Help})
	}
	if len(cmd) > 0 {
		if l, ok := cmdset.Leaves[cmd[len(cmd)-1]]; ok {
			for optname, opthelp := range l.Options {
				s = append(s, prompt.Suggest{Text: Sprintf("--%s", optname), Description: opthelp})
			}
		} else {
			for lname, l := range cmdset.Leaves {
				s = append(s, prompt.Suggest{Text: lname, Description: l.Help})
			}
		}
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

type RawWriter struct {
	prompt.PosixWriter
}

func (w *RawWriter) Write(data []byte) {
	w.WriteRaw(data)
}
func (w *RawWriter) WriteStr(data string) {
	w.WriteRawStr(data)
}

func exit(_ *prompt.Buffer) {
	// go-prompt eats some terminal attrs without this :(
	in, _ := unix.Open("/dev/tty", unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|unix.O_RDWR, 0666)
	tio, _ := termios.Tcgetattr(uintptr(in))
	tio.Lflag |= unix.ECHO
	termios.Tcsetattr(uintptr(in), termios.TCSANOW, tio)
	log.SetFlags(0)
	log.SetPrefix("\t |")
	log.Println(Red("Goodbye.\n"))
	os.Exit(0)
}

func main() {
	log.SetFlags(log.Ltime | log.Lmsgprefix)
	log.SetPrefix("|")
	log.Println(Bold("Hello, world!").Underline())

	if len(os.Args) == 2 {
		configfile, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(Sprintf("%s %s", Red("Reading config file failed:"), err))
		} else {
			log.Printf("Reading %s", os.Args[1])
		}
		defer configfile.Close()
		config := bufio.NewScanner(configfile)
		for config.Scan() {
			line := config.Text()
			if len(line) > 0 {
				Command(line)
			}
		}
		if err := config.Err(); err != nil {
			log.Fatal(err)
		}
	}

	for {
		block := make([]byte, 1)
		os.Stdin.Read(block)
		if block[0] < 5 {
			exit(nil)
		}
		log.SetFlags(0)
		log.SetPrefix("\t |")
		log.Println(Bold("Accepting commands."))
		log.SetFlags(log.Ltime | log.Lmsgprefix)
		log.SetPrefix("|")
		log.Println(White("Tab completion available. Use '? <command>' for help."))

		prompt.New(
			Command,
			tabComplete,
			prompt.OptionTitle("tarragon"),
			prompt.OptionPrefix(Sprintf("%s |>", Faint(Name()))),
			prompt.OptionLivePrefix(func() (string, bool) {
				return Sprintf("%s |%s> ", Faint(Name()), BrightWhite(namespace)), true
			}),
			prompt.OptionShowCompletionAtStart(),
			prompt.OptionCompletionOnDown(),
			prompt.OptionWriter(&RawWriter{}), // go-prompt eats escape sequences without this :(
			prompt.OptionAddKeyBind(prompt.KeyBind{
				Key: prompt.ControlC,
				Fn:  exit,
			}),
		).Run()
	}
}

func unquote(s string) string {
	ret := strings.Trim(s, "\"")
	if (len(s)-len(ret))%2 == 0 {
		s = ret
	}
	ret = strings.Trim(s, "'")
	if (len(s)-len(ret))%2 == 0 {
		s = ret
	}
	return s
}

func Command(command string) {
	command = strings.TrimSpace(command)
	if len(command) == 0 || strings.HasPrefix(command, "#") {
		// command is a comment or empty, ignore
		return
	}

	cmd := strings.Fields(command)
	data = make(map[string]string)
	var verbs []string

	setting := ""
	for _, field := range cmd {
		isOpt := strings.HasPrefix(field, "-")
		if len(setting) > 0 && !isOpt {
			data[setting] = unquote(field)
			setting = ""
			continue
		} else if len(setting) > 0 {
			data[setting] = ""
			setting = ""
		}
		if isOpt {
			set := strings.Split(strings.TrimLeft(field, "-"), "=")
			if len(set) > 1 {
				data[set[0]] = unquote(set[1])
			} else {
				setting = strings.ToLower(set[0])
			}
		} else {
			verbs = append(verbs, strings.ToLower(field))
		}
	}
	if len(setting) != 0 {
		data[setting] = ""
		setting = ""
	}

	if len(verbs) > 0 && verbs[0] == "ns" {
		if len(verbs) > 1 {
			namespace = verbs[1]
			log.Println(Sprintf(BrightBlack("Command %v: %s %s"), Green(verbs), BrightGreen("Namespace set to"), Green(namespace)))
		} else {
			namespace = ""
			log.Println(Sprintf(BrightBlack("Command %v: %s"), Green(verbs), BrightGreen("Namespace cleared")))
		}
		return
	}

	if namespace != "" {
		verbs = append([]string{namespace}, verbs...)
	}

	log.Println(Sprintf(BrightBlack("Command %v %v"), Green(verbs), BrightGreen(data)))

	branch := Commands
	showHelp := true
	onlyHelp := false
	foundLeaf := ""
	if len(verbs) > 0 && verbs[0] == "?" {
		verbs = verbs[1:]
		onlyHelp = true
	}

	for _, verb := range verbs {
		if next, ok := branch.Branches[verb]; ok {
			branch = next
		} else if call, ok := branch.Leaves[verb]; ok {
			foundLeaf = verb
			if !onlyHelp {
				showHelp = false
				call.Trigger(func(key string) (string, bool) {
					if value, ok := data[key]; ok {
						return value, true
					}
					showHelp = true
					log.Println(Sprintf(Red("Missing %s"), Sprintf(Cyan("--%s").Bold(), key)))
					return "", false
				})
			}
		} else {
			log.Println(Sprintf(Red("Unknown command %v"), Cyan(cmd)))
			break
		}
	}
	if showHelp {
		if len(branch.Help) > 0 {
			log.Println(Sprintf(Bold("Help: %s"), BrightMagenta(branch.Help)))
		}
		if foundLeaf == "" && len(branch.Branches) > 0 {
			log.Println("Available submenus:")
			for mname, mbranch := range branch.Branches {
				log.Println(Sprintf(" %s: %s", Bold(mname), BrightMagenta(mbranch.Help)))
			}
		}
		if len(branch.Leaves) > 0 {
			if foundLeaf == "" {
				log.Println("Available commands:")
			}
			for cname, cleaf := range branch.Leaves {
				if foundLeaf != "" && foundLeaf != cname {
					continue
				}
				log.Println(Sprintf(" %s: %s", Bold(cname), BrightMagenta(cleaf.Help)))
				if len(cleaf.Options) > 0 {
					for optname, opthelp := range cleaf.Options {
						log.Println(Sprintf("   %s: %s", Sprintf(Magenta("--%s"), optname), BrightMagenta(opthelp)))
					}
				}
			}
		}
	}
}
