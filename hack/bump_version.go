package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v57/github"
)

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

func getLatestTag(ctx context.Context, client *github.Client, owner string, repo string) (*github.RepositoryTag, error) {
	tags, _, err := client.Repositories.ListTags(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}

	sort.Slice(tags, func(i, j int) bool {
		return *tags[i].Name > *tags[j].Name
	})

	return tags[0], nil
}

func incrementVersion(tag string, bumpType string) string {
	parts := strings.Split(tag, ".")
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func createNewTag(ctx context.Context, client *github.Client, owner string, repo string, newTag string, sha string, dryRun bool) {
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

	latestTag, err := getLatestTag(ctx, client, owner, repo)
	if err != nil {
		fmt.Printf("Error getting latest tag: %v\n", err)
		os.Exit(1)
	}

	bumpType := getBumpTypeFromCommitMessage(ctx, client, owner, repo)
	newTag := incrementVersion(*latestTag.Name, bumpType)

	createNewTag(ctx, client, owner, repo, newTag, *latestTag.Commit.SHA, dryRun)
}
