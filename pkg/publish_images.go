package dox

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"net/url"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
	"github.com/jesselang/go-confluence"

)

func getImageSrcsFromHTML(content string) ([]string, error) {
	var imageSrcs []string
	z := html.NewTokenizer(strings.NewReader(content))
	for {
		tokenType := z.Next()
		token := z.Token()

		switch tokenType {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return nil, z.Err()
			} else {
				return imageSrcs, nil
			}
		case html.SelfClosingTagToken:
			if token.Data == "img" {
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						imageSrcs = append(imageSrcs, attr.Val)
					}
				}
			}
		}
	}
	return imageSrcs, nil
}

func getImageSrcFiles(content string, file string) ([]string, error) {
	fileDir := filepath.Dir(file)
	imageSrcs, err := getImageSrcsFromHTML(content)
	if err != nil {
		return nil, err
	}

	var imageSrcFiles []string

	for _, imageSrc := range imageSrcs {
		// skip imageSrcs that are URLs
		if _, err := url.ParseRequestURI(imageSrc); err == nil {
			continue
		}

		imageSrcPath := filepath.Join(fileDir, imageSrc)
		if _, err := os.Stat(imageSrcPath); !os.IsNotExist(err) {
			imageSrcFiles = append(imageSrcFiles, imageSrc)
		} else {
			fmt.Printf("warn: could not find image file %s\n", imageSrcPath)
		}
	}

	return imageSrcFiles, nil
}

func replaceImagesWithAttachments(imageSrcFiles []string, file string, pageContent string, pageID string, wiki *confluence.Wiki, uri string) (string, error) {
	fileDir := filepath.Dir(file)
	for _, imageSrcFile := range imageSrcFiles {
		imageSrcPath := filepath.Join(fileDir, imageSrcFile)
		imageSrcFilename := filepath.Base(imageSrcPath)

		results, err := wiki.GetAttachment(pageID, imageSrcFilename)
		if err != nil {
			return "", err
		}

		if len(results.Results) == 0 {
			// create new attachment
			_, err := wiki.CreateAttachment(pageID, imageSrcPath)
			if err != nil {
				return "", err
			}
		} else {
			// update existing attachment
			imageData, err := wiki.GetAttachmentData(pageID, imageSrcFilename)
			if err != nil {
				return "", err
			}

			attachmentSum := getBytesSha256(imageData)
			fileSum, err := getFileSha256(imageSrcPath)
			if err != nil {
				return "", err
			}

			if fileSum != attachmentSum {
				_, err := wiki.UpdateAttachment(pageID, imageSrcPath, results.Results[0].ID)
				if err != nil {
					return "", err
				}
			}
		}

		pageContent = strings.Replace(pageContent, fmt.Sprintf(`"%s"`, imageSrcFile), fmt.Sprintf(`"%s/download/attachments/%s/%s"`, uri, pageID, imageSrcFilename), -1)
	}

	return pageContent, nil
}

func getFileSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func getBytesSha256(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
