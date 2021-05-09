package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/urfave/cli/v2"
)

func main() {
	tagger, err := newGitTagger()
	if err != nil {
		exit(err)
	}
	if err := run(tagger, os.Args); err != nil {
		exit(err)
	}
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func run(t tagger, args []string) error {
	return initCLIApp(t).Run(args)
}

func initCLIApp(tagger tagger) *cli.App {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name: "major",
			Action: func(*cli.Context) error {
				return major(tagger)
			},
		},
		{
			Name: "minor",
			Action: func(*cli.Context) error {
				return minor(tagger)
			},
		},
		{
			Name: "patch",
			Action: func(*cli.Context) error {
				return patch(tagger)
			},
		},
	}

	return app
}

func major(tagger tagger) error {
	return increment(tagger, (vtag).incrementMajor)
}

func minor(tagger tagger) error {
	return increment(tagger, (vtag).incrementMinor)
}

func patch(tagger tagger) error {
	return increment(tagger, (vtag).incrementPatch)
}

func increment(tagger tagger, increment func(vtag) vtag) error {
	latest, err := tagger.latest()
	if err != nil {
		return fmt.Errorf("failed to get latest tag: %w", err)
	}

	return tagger.create(
		increment(latest),
	)
}

type tagger interface {
	tags() ([]vtag, error)
	latest() (vtag, error)
	create(t vtag) error
}

func newGitTagger() (*gitTagger, error) {
	repo, err := git.PlainOpen(".git")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &gitTagger{
		repo: repo,
	}, nil
}

type gitTagger struct {
	repo *git.Repository
}

func (t gitTagger) tags() ([]vtag, error) {
	tags, err := t.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var vtags []vtag

	if err := tags.ForEach(func(tag *plumbing.Reference) error {
		var vtag vtag
		if _, err := fmt.Fscan(strings.NewReader(tag.Name().Short()), &vtag); err != nil {
			return err
		}

		vtags = append(vtags, vtag)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	sortVtags(vtags)

	return vtags, nil
}

func (t gitTagger) latest() (vtag, error) {
	tags, err := t.tags()
	if err != nil {
		return vtag{}, err
	}

	return latestVtag(tags), nil
}

func (t gitTagger) create(tag vtag) error {
	head, err := t.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get head: %w", err)
	}

	if _, err := t.repo.CreateTag(tag.String(), head.Hash(), nil); err != nil {
		return err
	}

	return nil
}

func latestVtag(tags []vtag) vtag {
	sortVtags(tags)
	return tags[len(tags)-1]
}

func sortVtags(tags []vtag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].String() < tags[j].String()
	})
}

type vtag struct {
	major, minor, patch int
}

func (t *vtag) Scan(state fmt.ScanState, _ rune) error {
	if _, err := fmt.Fscanf(state, "v%d.%d.%d", &t.major, &t.minor, &t.patch); err != nil {
		return err
	}

	return nil
}

func (t vtag) String() string {
	return fmt.Sprintf("v%d.%d.%d", t.major, t.minor, t.patch)
}

func (t vtag) incrementMajor() vtag {
	return vtag{
		major: t.major + 1,
	}
}

func (t vtag) incrementMinor() vtag {
	return vtag{
		major: t.major,
		minor: t.minor + 1,
	}
}

func (t vtag) incrementPatch() vtag {
	return vtag{
		major: t.major,
		minor: t.minor,
		patch: t.patch + 1,
	}
}
