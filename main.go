package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	var (
		token string
		owner string
		repo  string
		base  string
	)

	flag.StringVar(&token, "token", "", "GitHub API token")
	flag.StringVar(&owner, "owner", "", "repository owner")
	flag.StringVar(&repo, "repo", "", "repository name")
	flag.StringVar(&base, "base", "main", "base branch name")
	flag.Parse()

	if token == "" || owner == "" || repo == "" {
		log.Fatal("token, owner, and repo are required")
	}

	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oAuthClient := oauth2.NewClient(ctx, tokenSource)
	ghClient := github.NewClient(oAuthClient)

	options := &github.PullRequestListOptions{
		State:     "closed",
		Base:      base,
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var mergedPRCount int
	var timeToMerge time.Duration

	for {
		prs, res, err := ghClient.PullRequests.List(ctx, owner, repo, options)
		if err != nil {
			log.Fatal(err)
		}

		for _, pr := range prs {
			if pr.CreatedAt.AddDate(0, 0, 14).Before(time.Now()) {
				break
			}

			if pr.MergedAt != nil {
				mergedPRCount++
				timeToMerge += pr.MergedAt.Sub(*pr.CreatedAt)
			}
		}

		if res.NextPage == 0 {
			break
		}
		options.Page = res.NextPage
	}

	fmt.Printf("Time to merge: %v\n", timeToMerge/time.Duration(mergedPRCount))
}
