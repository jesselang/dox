package dox

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jesselang/dox/pkg/source"
	"github.com/jesselang/go-confluence"
	"github.com/spf13/viper"
)

var uri string
var space string
var browseUrlBase string
var username string
var password string

func getConfigVars() error {
	uri = viper.GetString("uri")
	if len(uri) == 0 {
		return errors.New("uri must be set in config")
	}

	space = viper.GetString("space")
	if len(space) == 0 {
		return errors.New("space must be set in config")
	}

	browseUrlBase = viper.GetString("browse_url_base")
	if len(browseUrlBase) == 0 {
		return errors.New("browse_url_base must be set in config")
	}

	username = os.Getenv("DOX_USERNAME")
	if len(username) == 0 {
		return errors.New("DOX_USERNAME must be set")
	}

	password = os.Getenv("DOX_PASSWORD")
	if len(password) == 0 {
		return errors.New("DOX_PASSWORD must be set")
	}

	return nil
}

func Publish(files []string, repoRoot string, verbose bool, dryRun bool) error {
	err := getConfigVars()
	if err != nil {
		return err
	}

	wiki, err := confluence.NewWiki(
		uri,
		confluence.BasicAuth(
			username,
			password,
		),
	)
	if err != nil {
		return err
	}

	// make sources out of each file
	var sources []source.Source
	for _, file := range files {
		src, err := source.New(file, source.Opts{
			StripComments: true,
			TrimSpace: true,
			DoxNoticeFileUrl: fileBrowseUrl(browseUrlBase, repoRoot, file),
		})
		if err != nil {
			return err
		}
		if src.Ignore() {
			continue
		}
		sources = append(sources, src)
	}

	// try to find root page
	rootPageSrc, err := getRootPageSrc(sources)
	if err != nil {
		return err
	}

	if rootPageSrc == nil {
		// create dox default root page
		rootPageSrc, err = source.New("", source.Opts{})
		if err != nil {
			return err
		}
		sources = append(sources, rootPageSrc)
	}

	// createStub only, we require root page's ID
	rootID, err := createStub(wiki, rootPageSrc, "", dryRun)
	if err != nil {
		return err
	}

	// TODO: this prints even if we did not stub the page
	if verbose {
		fmt.Printf("root page stubbed to %s\n", rootID)
	}

	for _, src := range sources {
		_, err := createStub(wiki, src, rootID, dryRun)
		if err != nil {
			return err
		}
	}

	for _, src := range sources {
		id, err := updateContent(wiki, src, repoRoot, dryRun)
		if err != nil {
			return err
		}
		if verbose {
			srcFile := src.File()
			if srcFile == "" {
				srcFile = "root"
			}
			fmt.Printf("%s published to %s\n", srcFile, id)
		}
	}
	return nil
}

func createStub(wiki *confluence.Wiki, src source.Source, rootID string, dryRun bool) (id string, err error) {
	if src.Ignore() {
		return "", fmt.Errorf("should not publish an ignored page")
	}

	if dryRun || src.ID() != "" {
		return src.ID(), nil
	}

	// NEW
	c := &confluence.Content{
		ID:    src.ID(),
		Type:  "page",
		Title: src.Title(),
	}

	if rootID != "" {
		c.Ancestors = []confluence.ContentAncestor{{ID: rootID}}
	}
	c.Body.Storage.Value = "This is a page stub created by dox."
	c.Body.Storage.Representation = "storage"
	c.Space.Key = space
	c.Version.Number = 1

	c, err = wiki.CreateContent(c)
	if err != nil {
		// TODO: confluence does not support duplicate title in a space
		return "", err
	}

	err = src.SetID(c.ID)
	if err != nil {
		return "", err
	}

	return src.ID(), nil
}

func updateContent(wiki *confluence.Wiki, src source.Source, repoRoot string, dryRun bool) (id string, err error) {
	if src.Ignore() {
		return "", fmt.Errorf("should not publish an ignored page")
	}

	if dryRun {
		return src.ID(), nil
	}

	c, err := wiki.GetContent(src.ID(),
		[]string{"body.storage", "space", "version"})
	if err != nil {
		// TODO: handle 404 where dox id exists in source, but published page does not
		return "", err
	}

	sourceOutput := src.Output()

	imageSrcFiles, err := getImageSrcFiles(sourceOutput, src.File())
	if err != nil {
		return "", err
	}

	pageContent, err := replaceImagesWithAttachments(imageSrcFiles, src.File(), sourceOutput, c.ID, wiki, uri)
	if err != nil {
		return "", err
	}

	pageContent, err = replaceRelativeLinks(src.File(), pageContent, uri, browseUrlBase, repoRoot)
	if err != nil {
		return "", err
	}

	// TODO: Confluence assigns a unique macro id to each macro in page. If a page contains macros,
	//       this condition will always be true since dox source does not contain macro ids.
	if c.Body.Storage.Value != pageContent {
		c.Body.Storage.Value = pageContent
		c.Version.Number += 1

		c, err = wiki.UpdateContent(c)
		if err != nil {
			return "", err
		}
	}

	return c.ID, nil
}

func getRootPageSrc(sources []source.Source) (source.Source, error) {
	var rootPages []source.Source
	for _, src := range sources {
		if src.IsRootPage() {
			rootPages = append(rootPages, src)
		}
	}

	if len(rootPages) == 1 {
		return rootPages[0], nil
	} else if len(rootPages) > 1 {
		var rootPagePaths []string
		for _, rp := range rootPages {
			rootPagePaths = append(rootPagePaths, rp.File())
		}
		return nil, fmt.Errorf("multiple root pages found: %s\n", strings.Join(rootPagePaths, ", "))
	} else {
		return nil, nil
	}
}
