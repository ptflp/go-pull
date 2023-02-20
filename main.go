package main

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func detectSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@")
}

func detectHTTPSURL(url string) bool {
	return strings.HasPrefix(url, "https://")
}

func parsePATH(url string) (string, error) {
	if detectSSHURL(url) {
		url = strings.ReplaceAll(url, "git@", "")
		url = strings.ReplaceAll(url, ":", "/")
		url = strings.ReplaceAll(url, ".git", "")
		return url, nil
	}

	if detectHTTPSURL(url) {
		url = strings.ReplaceAll(url, "https://", "")
		url = strings.ReplaceAll(url, ".git", "")
		return url, nil
	}

	return "", errors.New("unknown url type")
}

func gitClone(url string, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
	})

	return err
}

func preparePath(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	return os.MkdirAll(path, os.ModePerm)
}

func absPATH(goPATH, path string) string {
	return filepath.Join(goPATH, "src", path)
}

func replaceHTTPS(url string) string {
	return strings.ReplaceAll(url, "https://", "git@")
}

func main() {
	// get url from command line arguments
	args := os.Args
	if len(args) < 2 {
		log.Println("please provide a url")
		return
	}
	if len(args) == 3 {
		if args[2] == "ssh" {
			log.Println("using ssh")
			args[1] = replaceHTTPS(args[1])
		}
	}
	url := args[1]
	path, err := parsePATH(url)
	if err != nil {
		log.Println(err)
		return
	}
	goPATH := os.Getenv("GOPATH")
	if goPATH == "" {
		goPATH = build.Default.GOPATH
	}
	path = absPATH(goPATH, path)
	log.Println("path:", path)
	err = preparePath(path)
	if err != nil {
		log.Println(err)
		return
	}
	if args[2] == "ssh" {
		url = url + ".git"
	}
	err = gitClone(url, path)
	if err != nil {
		log.Println(url, err)
		return
	}
}
