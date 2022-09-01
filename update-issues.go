package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"html/template"
	"math"
	"os"
	"strings"
	"time"
)

type Ticket struct {
	Title            string
	State            string
	URL              string
	NumComments      int
	TimeToFirstTouch string
	TimeToClose      string
	CreatedAt        string
}

type PageData struct {
	Tickets []Ticket
}

var (
	Namespace string
	Repo      string
)

func initGlobalSettings() {
	flag.StringVar(&Namespace, "n", "", "namespace of repository")
	flag.StringVar(&Repo, "r", "", "repository name")
	flag.Parse()

	if Namespace == "" {
		fmt.Fprintf(os.Stderr, "error: namespace of repository must be set\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if Repo == "" {
		fmt.Fprintf(os.Stderr, "error: repository name must be set\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {

	initGlobalSettings()

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN"),
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Only started issue tracking in April 2022
	start := time.Date(2022, time.Month(4), 8, 0, 0, 0, 0, time.UTC)

	opts := &github.IssueListByRepoOptions{
		State:       "all",
		Since:       start,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	issues, _, err := client.Issues.ListByRepo(context.Background(), Namespace, Repo, opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var tickets []Ticket

	// fmt.Println()
	for _, i := range issues {
		if i.PullRequestLinks != nil {
			continue
		}

		if i.CreatedAt.Sub(start) < 0 {
			continue
		}

		if *i.Number == 62 {
			continue
		}

		createdAt := fmt.Sprintf("%s", (*i.CreatedAt).Format("2006-01-02 15:04"))
		// fmt.Printf("%-7s   %-70s  %-30s  %s\n", *i.State, *i.Title, *i.HTMLURL, createdAt)
		comments, _, err := client.Issues.ListComments(context.Background(), "E4S-Project", "e4s", *i.Number, nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		newTicket := Ticket{
			Title:       *i.Title,
			State:       strings.ToUpper(*i.State),
			URL:         *i.HTMLURL,
			NumComments: *i.Comments,
			CreatedAt:   createdAt,
		}

		if i.ClosedAt != nil {
			timeToClose := i.ClosedAt.Sub(*i.CreatedAt)
			nHours := timeToClose.Round(time.Hour).Hours()
			nDays := nHours / 24
			nWeeks := nDays / 7

			if nHours < 24 {
				newTicket.TimeToClose = "< 1 day"
			} else if nWeeks < 1 {
				newTicket.TimeToClose = fmt.Sprintf("%d days", int(math.Round(nDays)))
			} else if nWeeks >= 1 {
				nWeeksRnd := int(math.Floor(nWeeks))
				nRemDays := float64((int(math.Round(nHours)) % 168)) / 24
				if nRemDays < 1 {
					if nWeeksRnd == 1 {
						newTicket.TimeToClose = fmt.Sprintf("1 week")
					} else {
						newTicket.TimeToClose = fmt.Sprintf("%d weeks", nWeeksRnd)
					}
				} else if nRemDays == 1 {
					if nWeeksRnd == 1 {
						newTicket.TimeToClose = fmt.Sprintf("1 week 1 day")
					} else {
						newTicket.TimeToClose = fmt.Sprintf("%d weeks 1 day", nWeeksRnd)
					}
				} else {
					if nWeeksRnd == 1 {
						newTicket.TimeToClose = fmt.Sprintf("1 week %d days", int(math.Floor(nRemDays)))
					} else {
						newTicket.TimeToClose = fmt.Sprintf("%d weeks %d days", nWeeksRnd, int(math.Floor(nRemDays)))
					}
				}
			}
		} else {
			newTicket.TimeToClose = "--"
		}

		for _, c := range comments {
			timeToFirstTouch := c.CreatedAt.Sub(*i.CreatedAt)
			newTicket.TimeToFirstTouch = fmt.Sprintf("%s", timeToFirstTouch)
			// fmt.Printf("- %-20s %s (%s)\n", *c.User.Login, *c.HTMLURL, dt)
			break
		}

		tickets = append(tickets, newTicket)
		// fmt.Println()
	}
	// fmt.Println()

	// fmt.Println()
	// for _, t := range tickets {
	// 	fmt.Printf("%10s  %-70s %-20s   %s\n", t.State, t.Title, t.URL, t.TimeToFirstTouch)
	// }
	// fmt.Println()

	data := PageData{
		Tickets: tickets,
	}

	f, e := os.Create("index.html")
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}

	tpl, e := template.ParseFiles("index.tpl.html")
	if e != nil {
		fmt.Println("parse:", e)
		os.Exit(1)
	}
	tpl.Execute(f, data)
}
