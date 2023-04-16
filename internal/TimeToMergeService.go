package internal

import (
	"context"
	"github.com/google/go-github/github"
	"time"
)

// getPRMetrics は、指定されたリポジトリの PR のメトリクスを取得する関数です。
// mergedPRCount はマージされた PR の数、timeToMerge はマージにかかった平均時間を返します。
func GetPRMetrics(client *github.Client, ctx context.Context, owner, repo, base string) (int, time.Duration, error) {
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
		prs, res, err := client.PullRequests.List(ctx, owner, repo, options)
		if err != nil {
			return 0, 0, err
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

	return mergedPRCount, timeToMerge, nil
}
