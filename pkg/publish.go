package dox

import (
	"errors"
	"fmt"
	"github.com/jesselang/go-confluence"
	"os"
)

func Publish(file string, dryRun bool) (id string, err error) {
	uri := os.Getenv("DOX_URI")
	if len(uri) == 0 {
		return id, errors.New("DOX_URI must be set")
	}

	username := os.Getenv("DOX_USERNAME")
	if len(username) == 0 {
		return id, errors.New("DOX_USERNAME must be set")
	}

	password := os.Getenv("DOX_PASSWORD")
	if len(password) == 0 {
		return id, errors.New("DOX_PASSWORD must be set")
	}

	space := os.Getenv("DOX_SPACE")
	if len(space) == 0 {
		return id, errors.New("DOX_PASSWORD must be set")
	}

	wiki, err := confluence.NewWiki(
		uri,
		confluence.BasicAuth(
			username,
			password,
		),
	)

	if dryRun {
		id = "dry-run"
		fmt.Println("dry-run")
	} else {
		fmt.Println("not dry-run")
		id, err = GetContentId(file)
		if err != nil {
			return
		}
	}

	c := &confluence.Content{
		ID:   id,
		Type: "page",
	}

	title, body, err := Convert(file)
	if err != nil {
		return
	}

	c.Title = title
	c.Body.Storage.Value = body
	c.Body.Storage.Representation = "storage"
	c.Space.Key = space // should be taken from repo config
	c.Version.Number = 1

	if c.ID == "" {
		c, _, err = wiki.CreateContent(c)
		if err != nil {
			return
		}
		id = c.ID
		// create
		// set content id
		err = UpdateContentId(file, id)
		if err != nil {
			return
		}
	} else {
		if !dryRun {
			cur, err := wiki.GetContent(c.ID, []string{"version"})
			if err != nil {
				return "", err
			}

			c.Version.Number = cur.Version.Number + 1
			c, _, err = wiki.UpdateContent(c)
			if err != nil {
				return "", err
			}

			id = c.ID
		}

	}

	return
}
