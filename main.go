//Filename: Main.go
//Author: Nyah Check
//Purpose: GitHub Bot to increase following and print following list.
//Licence: GNU PL 2017


package main

import (

    "bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/oauth2"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
)

const (
	// LOGGER is what is printed for help/info output
	LOGGER = "GHBot - %s\n"
	// VERSION is the binary version.
	VERSION = "v0.1.0"
)

var (
	token    string
	interval string

	lastChecked time.Time

	debug   bool
	version bool
)

type UserDetails struct {
    Login string `json:"login"`
    Id int64 `json:"id"`
    Url int64 `json:"url"`
    HtmlUrl string `json:"html_url"`
}

func init() {
	// parse flags
	flag.StringVar(&token, "token", "", "GitHub API token")
	flag.StringVar(&interval, "interval", "30s", "check interval (ex. 5ms, 10s, 1m, 3h)")

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(LOGGER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if token == "" {
		usageAndExit("GitHub token cannot be empty.", 1)
	}
}

//fchecks panic in file.
func fcheck(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
	var ticker *time.Ticker
	
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			ticker.Stop()
			logrus.Infof("Received %s, exiting.", sig.String())
			os.Exit(0)
		}
	}()

	// Create the http client.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	// Create the github client.
	client := github.NewClient(tc)
	gt := github.NewClient(nil)//client for GitHub user whose following list will be used.
	

	// Get the authenticated user, the empty string being passed let's the GitHub
	// API know we want ourself.
	user, _, err := client.Users.Get("")
	if err != nil {
		logrus.Fatal(err)
	}
	username := *user.Login

	// parse the duration
	dur, err := time.ParseDuration(interval)
	if err != nil {
		logrus.Fatalf("parsing %s as duration failed: %v", interval, err)
	}
	ticker = time.NewTicker(dur)

	logrus.Infof("Bot started for user %s.", username)

	for range ticker.C {
		page := 1
		perPage := 30
		if err := getFollowing(client, username, page, perPage); err != nil {
			logrus.Fatal(err)
		}
		
		if err = getFollowers(client, username, page, perPage); err != nil {
		    logrus.Fatal(err)
		}
		
		if err := followUsers(client, username, page, perPage); err != nil {
		    logrus.Fatal(err)
		}
	}
    
    /**
     * This part will be tested and worked on later.
     */
    for range ticker.C {
        page := 1
        perPage := 30
        if err := Unfollow(client, username, page, perPage); err := nil {
            logrus.Fatal(err)
        }
    }
}

// getFollowers iterates over all followers received for user.
func getFollowers(client *github.Client, username string, page, perPage int) error {
    ctx := context.Background()
	opt := &github.ListOptions{
		All:   true,
		Since: lastChecked,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}
	if lastChecked.IsZero() {
		lastChecked = time.Now()
	}
	

    followers, resp, err := github.UsersService.ListFollowers(ctx, username, opt)
	if err != nil {
		return err
	}

	for _, flwr := range followers {
		// handle follower details.
		user, e := handleUser(flwr.Body)
        if err != nil {
            panic(err.Error())
        }
        
        //writes user details to file.
        followersJson, _ := json.Marshal(rankings)
        err = ioutil.WriteFile("results/followers.json", followersJson, 0644)
        fmt.Printf("%+v", rankings)
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return getfollowers(client, username, page, perPage)
}

func handleUser(body []byte) (*UserDetails, error) {
    var s = new(UserDetails)
    err := json.Unmarshal(body, &s)
    if(err != nil){
        fmt.Println("whoops:", err)
    }
    return s, err
}

// getFollowing iterates over the list of following and writes to file.
func getFollowing(client *github.Client, username string, page, perPage int) error {
    ctx := context.Background()
	opt := &github.ListOptions{
		All:   true,
		Since: lastChecked,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}
	if lastChecked.IsZero() {
		lastChecked = time.Now()
	}
	

    following, resp, err := github.UsersService.ListFollowing(ctx, username, opt)//to test properly whether to parse resp instead inloop
	if err != nil {
		return err
	}

	for _, flwg := range following {
		// handle following details.
		user, e := handleUser(flwr.Body)
        if err != nil {
            panic(err.Error())
        }
        
        //writes user details to file.
        followingJson, _ := json.Marshal(rankings)
        err = ioutil.WriteFile("results/following.json", followingJson, 0644)
        fmt.Printf("%+v", rankings)
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return getfollowing(client, username, page, perPage)


	return nil
}

// followUsers, gets the list of followers for a particular user and followers them on GitHub.
// This requires authentication with the API.
func followUsers(client *github.Client, username string, page, perPage int) {
    ctx := context.Background()
	opt := &github.ListOptions{
		All:   true,
		Since: lastChecked,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}
	if lastChecked.IsZero() {
		lastChecked = time.Now()
	}
	

    usrs, resp, err := github.UsersService.ListFollowing(ctx, username, opt)//to test properly whether to parse resp instead inloop
	if err != nil {
		return err
	}

	for _, usr := range usrs {
		//Follow user
		res, e := github.UsersService.Follow(ctx, usr)
        if err != nil {
            panic(e.Error())
        }
        
        fmt.Printf("%+v", res)
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return followUsers(client, username, page, perPage)
}


// Unfollow all GitHub users on one's follower list.
func Unfollow(client *github.Client, username string, page, perPage int) {
    ctx := context.Background()
	opt := &github.ListOptions{
		All:   true,
		Since: lastChecked,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}
	if lastChecked.IsZero() {
		lastChecked = time.Now()
	}
	

    usrs, resp, err := github.UsersService.ListFollowing(ctx, username, opt)//to test properly whether to parse resp instead inloop
	if err != nil {
		return err
	}

	for _, usr := range usrs {
		//Follow user
		res, e := github.UsersService.Unfollow(ctx, usr)
        if err != nil {
            panic(e.Error())
        }
        
        fmt.Printf("%+v", res)
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return Unfollow(client, username, page, perPage)
}


func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
