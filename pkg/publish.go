package dox

import (
	"errors"
	"os"

	"github.com/jesselang/go-confluence"
	"github.com/spf13/viper"
	"github.com/jesselang/dox/pkg/source"
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

// Creates pages if they don't exist
func PrePublish(file string, rootID string, dryRun bool) (id string, err error) {
	err = getConfigVars()
	if err != nil {
		return id, err
	}

	wiki, err := confluence.NewWiki(
		uri,
		confluence.BasicAuth(
			username,
			password,
		),
	)
	if err != nil {
		return
	}

	src, err := source.New(file, source.Opts{StripComments: true, TrimSpace: true})
	if err != nil {
		return
	}

	if src.Ignore() {
		return
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

// Updates page content
func Publish(file string, repoRoot string, dryRun bool) (id string, err error) {
	err = getConfigVars()
	if err != nil {
		return id, err
	}

	wiki, err := confluence.NewWiki(
		uri,
		confluence.BasicAuth(
			username,
			password,
		),
	)
	if err != nil {
		return
	}

	src, err := source.New(file, source.Opts{
		StripComments: true,
		TrimSpace: true,
		DoxNoticeFileUrl: fileBrowseUrl(browseUrlBase, repoRoot, file),
	})
	if err != nil {
		return
	}

	if src.Ignore() {
		return
	}

	if dryRun {
		return src.ID(), nil
	}

	sourceOutput := src.Output()

	c, err := wiki.GetContent(src.ID(),
		[]string{"body.storage", "space", "version"})
	if err != nil {
		// TODO: handle 404 where dox id exists in source, but published page does not
		return "", err
	}

	imageSrcFiles, err := getImageSrcFiles(sourceOutput, file)
	if err != nil {
		return "", err
	}

	pageContent, err := replaceImagesWithAttachments(imageSrcFiles, file, sourceOutput, c.ID, wiki, uri)
	if err != nil {
		return "", err
	}

	pageContent, err = replaceRelativeLinks(file, pageContent, uri, browseUrlBase, repoRoot)
	if err != nil {
		return "", err
	}

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
