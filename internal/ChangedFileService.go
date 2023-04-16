package internal

import (
	"context"
	"github.com/google/go-github/github"
	"sort"
	"time"
)

// ChangedFile ファイル名とそのファイルの変更数を表す型
type ChangedFile struct {
	Filename string
	Count    int
}

func GetSortedFileList(client *github.Client, ctx context.Context, owner, repo string) ([]ChangedFile, error) {
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

	var fileList []ChangedFile
	for filename, count := range files {
		fileList = append(fileList, ChangedFile{Filename: filename, Count: count})
	}
	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].Count > fileList[j].Count
	})

	return fileList, nil
}
