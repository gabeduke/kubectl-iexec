package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v57/github"
)

type TagVersion struct {
	Tag    *github.RepositoryTag
	SemVer *semver.Version
}

func getGitHubToken() (string, error) {
	token, found := os.LookupEnv("GITHUB_TOKEN")
	if found {
		return token, nil
	}

	cmd := exec.Command("gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return token, fmt.Errorf("failed to get GITHUB_TOKEN: %w", err)
	}

	return string(out), nil
}
func getGitHubClient() (*github.Client, error) {
	token, err := getGitHubToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub token: %w", err)
	}

	token = strings.TrimSpace(token)

	client := github.NewClient(nil).WithAuthToken(token)

	return client, nil
}

func getLatestTag(ctx context.Context, client *github.Client, owner string, repo string) (*TagVersion, error) {
	opts := &github.ListOptions{PerPage: 100}
	var allTags []*TagVersion

	for {
		tags, resp, err := client.Repositories.ListTags(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			v, err := semver.NewVersion(*tag.Name)
			if err != nil {
				return nil, fmt.Errorf("invalid semver tag: %s", *tag.Name)
			}
			allTags = append(allTags, &TagVersion{Tag: tag, SemVer: v})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	sort.Slice(allTags, func(i, j int) bool {
		return allTags[i].SemVer.GreaterThan(allTags[j].SemVer)
	})

	return allTags[0], nil
}

func incrementVersion(tag *semver.Version, bumpType string) (string, error) {
	var newTag semver.Version

	switch bumpType {
	case "major":
		newTag = tag.IncMajor()
	case "minor":
		newTag = tag.IncMinor()
	case "patch":
		newTag = tag.IncPatch()
	default:
		return "", fmt.Errorf("invalid bump type: %s", bumpType)
	}

	return newTag.String(), nil
}

func createNewTag(ctx context.Context, client *github.Client, owner string, repo string, newTag string, dryRun bool) {
	sha := os.Getenv("GITHUB_SHA")
	ref := &github.Reference{
		Ref: github.String("refs/tags/" + newTag),
		Object: &github.GitObject{
			SHA: &sha,
		},
	}

	if dryRun {
		fmt.Printf("Dry run: would have created ref with owner: %s, repo: %s, ref: %v\n", owner, repo, ref)
	} else {
		_, _, err := client.Git.CreateRef(ctx, owner, repo, ref)
		if err != nil {
			panic(err)
		}
	}
}

func getBumpTypeFromCommitMessage(ctx context.Context, client *github.Client, owner string, repo string) string {
	commits, _, err := client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		panic(err)
	}

	commitMessage := *commits[0].Commit.Message
	commitMessage = strings.ToLower(commitMessage)

	if strings.Contains(commitMessage, "#major") {
		return "major"
	} else if strings.Contains(commitMessage, "#minor") {
		return "minor"
	} else {
		return "patch"
	}
}

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", true, "If true, print API calls but do not make them.")
	flag.Parse()

	ctx := context.Background()

	client, err := getGitHubClient()
	if err != nil {
		fmt.Printf("Error getting GitHub client: %v\n", err)
		os.Exit(1)
	}

	repoFullName := os.Getenv("GITHUB_REPOSITORY")
	split := strings.Split(repoFullName, "/")
	owner := split[0]
	repo := split[1]

	latestTagVersion, err := getLatestTag(ctx, client, owner, repo)
	if err != nil {
		fmt.Printf("Error getting latest tag: %v\n", err)
		os.Exit(1)
	}

	bumpType := getBumpTypeFromCommitMessage(ctx, client, owner, repo)
	newTag, err := incrementVersion(latestTagVersion.SemVer, bumpType)
	if err != nil {
		fmt.Printf("Error incrementing version: %v\n", err)
		os.Exit(1)
	}

	createNewTag(ctx, client, owner, repo, newTag, dryRun)
}
