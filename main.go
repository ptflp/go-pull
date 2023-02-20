package main

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func detectSSHURL(url string) bool {

	if strings.HasPrefix(url, "git@") {
		return true
	}
	return strings.HasPrefix(url, "ssh://")
}

func detectHTTPSURL(url string) bool {
	return strings.HasPrefix(url, "https://")
}

func parsePATH(url string) (string, error) {
	if detectSSHURL(url) {
		url = strings.ReplaceAll(url, "ssh://", "")
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

func detectKeys() (string, error) {
	var keyPath string
	home := os.Getenv("HOME")
	keyPath = filepath.Join(home, ".ssh/id_rsa")
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		return keyPath, nil
	}
	keyPath = filepath.Join(home, ".ssh/id_ed25519")
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		return keyPath, nil
	}

	return "", errors.New("no ssh keys found")
}

func gitClone(url string, path string, useSSH bool) error {
	var err error
	var auth transport.AuthMethod
	if useSSH {
		keyPath, err := detectKeys()
		if err != nil {
			return err
		}
		auth, err = ssh.NewPublicKeysFromFile("git", keyPath, "")
		if err != nil {
			return err
		}
	}
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		Auth: auth,
		URL:  url,
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
	url := args[1]
	var useSSH bool
	if len(args) == 3 {
		if args[2] == "ssh" {
			useSSH = true
			log.Println("using ssh")
			url = replaceHTTPS(args[1])
			url = "ssh://" + url + ".git"
		}
	}
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
	err = gitClone(url, path, useSSH)
	if err != nil {
		log.Println(url, err)
		return
	}
}
