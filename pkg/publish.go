package dox

import (
	"errors"
	"github.com/jesselang/dox/pkg/source"
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
	if err != nil {
		return
	}

	src, err := source.New(file)
	if err != nil {
		return
	}

	if dryRun {
		id = src.ID()
	} else {
		c := &confluence.Content{
			ID:    src.ID(),
			Type:  "page",
			Title: src.Title(),
		}
		c.Body.Storage.Value = src.Output()
		c.Body.Storage.Representation = "storage"
		c.Space.Key = space // should be taken from repo config
		c.Version.Number = 1

		if c.ID == "" {
			var resp []byte
			c, resp, err = wiki.CreateContent(c)
			if err != nil {
				// confluence does not support duplicate title in a space
				return "", errors.New(string(resp))
			}
			id = c.ID
			// create
			// set content id
			err = src.SetID(id)
			if err != nil {
				return
			}
		} else {
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
