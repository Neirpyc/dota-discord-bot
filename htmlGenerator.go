package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
)

type Replacement map[string]string

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (r Replacement) applyTemplate(template string) (string, error) {
	var out *os.File
	path := "tmp/" + RandStringRunes(64) + ".html"
	for _, err := os.Stat("assets/" + path); os.IsExist(err); _, err = os.Stat("assets/" + path) {
		path = "tmp/" + RandStringRunes(64) + ".html"
	}
	out, err := os.Create("assets/" + path)
	if err != nil {
		return "", err
	}
	input, err := os.Open(template)
	if err != nil {
		return "", err
	}
	contentByte, err := ioutil.ReadAll(input)
	if err != nil {
		return "", err
	}
	content := string(contentByte)
	for k, v := range r {
		content = strings.Replace(content, "$"+k, v, -1)
		content = strings.Replace(content, "$("+k+")", v, -1)
	}
	if _, err = out.Write([]byte(content)); err != nil {
		return "", err
	}
	_ = out.Close()
	return path, nil
}
