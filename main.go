package main

import (
	"context"
	"flag"
	"fmt"
	"gh-inspect/internal"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	// コマンドライン引数を取得
	var owner, repo, base string
	flag.StringVar(&owner, "o", "", "repository owner")
	flag.StringVar(&repo, "r", "", "repository name")
	flag.StringVar(&base, "b", "main", "base branch name")
	flag.Parse()

	// GITHUB_API_TOKEN 環境変数からトークンを取得
	token := os.Getenv("GITHUB_API_TOKEN")
	if token == "" || owner == "" || repo == "" {
		log.Fatal("token, owner, and repo are required")
	}

	// GitHub API へのクライアントの作成
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oAuthClient := oauth2.NewClient(ctx, tokenSource)
	ghClient := github.NewClient(oAuthClient)

	// PRのメトリクスを取得
	mergedPRCount, timeToMerge, err := internal.GetPRMetrics(ghClient, ctx, owner, repo, base)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Merged PR count within 6 months: %d\n\n", mergedPRCount)
	fmt.Printf("Average time to merge: %v\n\n", timeToMerge/time.Duration(mergedPRCount))

	// リポジトリのコミット履歴を取得して、ファイルの変更数をカウント
	fileList, err := internal.GetSortedFileList(ghClient, ctx, owner, repo)
	if err != nil {
		log.Fatal(err)
	}

	// 変更数が多い順にトップ 10 のファイルを表示
	fmt.Printf("Top 10 changed files within 6 months:\n")
	for i := 0; i < 10 && i < len(fileList); i++ {
		fmt.Printf("%s: %d\n", fileList[i].Filename, fileList[i].Count)
	}
}
