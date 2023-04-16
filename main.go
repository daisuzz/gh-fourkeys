package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// 変更頻度が多いファイルを10個出力
type fileCount struct {
	filename string
	count    int
}

func main() {
	var (
		owner string
		repo  string
		base  string
	)
	flag.StringVar(&owner, "o", "", "repository owner")
	flag.StringVar(&repo, "r", "", "repository name")
	flag.StringVar(&base, "b", "main", "base branch name")
	flag.Parse()
	token := os.Getenv("GITHUB_API_TOKEN")

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
			if pr.CreatedAt.AddDate(0, 6, 0).Before(time.Now()) {
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

	fmt.Printf("Merged PR count within 6 months: %d\n\n", mergedPRCount)
	fmt.Printf("Average time to merge: %v\n\n", timeToMerge/time.Duration(mergedPRCount))

	// リポジトリのコミット履歴を取得
	fileList, err := getSortedFileList(ghClient, ctx, owner, repo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Top 10 changed files within 6 months:\n")
	for i := 0; i < 10 && i < len(fileList); i++ {
		fmt.Printf("%s: %d\n", fileList[i].filename, fileList[i].count)
	}
}

func getSortedFileList(client *github.Client, ctx context.Context, owner, repo string) ([]fileCount, error) {
	files := make(map[string]int)

	// 直近半年分のコミット履歴を取得するための日付を計算
	now := time.Now()
	halfYearAgo := now.AddDate(0, -6, 0)

	// リポジトリのコミット履歴を取得(直近半年分)
	commits, _, err := client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
		Since: halfYearAgo,
	})
	if err != nil {
		return nil, err
	}

	// 変更されたファイルを数えるためのマップを作成
	for _, commit := range commits {
		// コミットに含まれるファイルの変更差分を取得
		commitFiles, _, err := client.Repositories.CompareCommits(ctx, owner, repo, commit.Parents[0].GetSHA(), commit.GetSHA())
		if err != nil {
			return nil, err
		}
		// 変更されたファイルをマップに追加
		for _, file := range commitFiles.Files {
			files[file.GetFilename()]++
		}
	}

	var fileList []fileCount
	for filename, count := range files {
		fileList = append(fileList, fileCount{filename, count})
	}
	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].count > fileList[j].count
	})

	return fileList, nil
}
