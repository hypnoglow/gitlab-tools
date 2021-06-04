package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xanzy/go-gitlab"
)

const (
	appName = "gitlab-runner-janitor"

	envGitlabToken = "GITLAB_TOKEN"

	defaultMaxDurationSinceLastContact = 72 * time.Hour
)

func main() {
	flagset := flag.NewFlagSet(appName, flag.ExitOnError)
	var (
		groupID                     string
		maxDurationSinceLastContact time.Duration
		dryRun                      bool
	)
	flagset.StringVar(&groupID, "group-id", "", "GitLab group ID (required)")
	flagset.DurationVar(&maxDurationSinceLastContact, "max-duration-since-last-contact", defaultMaxDurationSinceLastContact, "Remove runners with last contact duration bigger than this value")
	flagset.BoolVar(&dryRun, "dry-run", false, "Preview what will be done but do nothing")
	_ = flagset.Parse(os.Args[1:])

	if groupID == "" {
		flagset.Usage()
		os.Exit(1)
	}

	gitlabToken := os.Getenv(envGitlabToken)
	if gitlabToken == "" {
		fmt.Fprintf(os.Stderr, "Environment variable %v must not be empty.\n", envGitlabToken)
		os.Exit(2)
	}

	client, err := gitlab.NewClient(gitlabToken)
	if err != nil {
		log.Fatalf("Create GitLab client: %v", err)
	}

	runnerService := client.Runners

	page := 1
	perPage := 20
	removed := 0

listing:
	for {
		runners, _, err := runnerService.ListGroupsRunners(groupID, &gitlab.ListGroupsRunnersOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			log.Fatalf("List group runners: %v", err)
		}

		log.Printf("Retrieved %v runners on page %v", len(runners), page)

		for _, runner := range runners {
			log.Printf("Runner: %#v", runner)
			if runner.Online {
				continue
			}

			details, _, err := runnerService.GetRunnerDetails(runner.ID)
			if err != nil {
				log.Fatalf("Get runner details: %v", err)
			}

			durationSinceLastContact := time.Now().Sub(*details.ContactedAt)

			log.Printf("Last contact: %v (%v)", details.ContactedAt, durationSinceLastContact)

			if durationSinceLastContact > maxDurationSinceLastContact {
				log.Printf("Removing runner %v", runner.ID)

				if !dryRun {
					_, err := runnerService.RemoveRunner(runner.ID)
					if err != nil {
						log.Fatalf("Remove runner: %v", err)
					}
					removed++
				}
			}
		}

		if len(runners) < perPage {
			break listing
		}

		page++
	}

	log.Printf("Removed total %d runners.", removed)
}
