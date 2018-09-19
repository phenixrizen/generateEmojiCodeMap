package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"net/http"
	"os"
	"text/template"
)

const emojiDataURL = "https://raw.githubusercontent.com/phenixrizen/generateEmojiCodeMap/master/emoji-data.txt"

type TemplateData struct {
	PkgName string
	Pattern string
	CodeMap map[string]string
}

type Emoji struct {
	Description string `json:"description"`
	Emoji       string `json:"emoji"`
	Match       string `json:"match"`
}

const templateMapCode = `
package {{.PkgName}}

// NOTE: THIS FILE WAS PRODUCED BY THE
// EMOJICODEMAP CODE GENERATION TOOL (github.com/phenixrizen/generateEmojiCodeMap)
// DO NOT EDIT

const emojiPattern = "{{.Pattern}}"

// Mapping from character to concrete escape code.
var emojiCodeMap = map[string]string{
	{{range $key, $val := .CodeMap}}"{{$key}}": "{{$val}}",
{{end}}
}
`

var pkgName string
var fileName string

func init() {
	log.SetFlags(log.Llongfile)

	flag.StringVar(&pkgName, "pkg", "main", "output package")
	flag.StringVar(&fileName, "o", "emoji_codemap.go", "output file")
	flag.Parse()
}

func main() {

	codeMap, err := generateFromData(pkgName)
	if err != nil {
		log.Fatalln(err)
	}

	os.Remove(fileName)

	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	if _, err := file.Write(codeMap); err != nil {
		log.Fatalln(err)
	}
}

func generateFromData(pkgName string) ([]byte, error) {

	// Read Emoji file

	res, err := http.Get(emojiDataURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	/*emojiFile, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}*/

	var es []Emoji
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}

	emojiCodeMap := make(map[string]string)
	for _, e := range es {
		emojiCodeMap[e.Emoji] = e.Description
	}

	// Template GenerateSource

	var buf bytes.Buffer
	t := template.Must(template.New("template").Parse(templateMapCode))
	if err := t.Execute(&buf, TemplateData{pkgName, "", emojiCodeMap}); err != nil {
		return nil, err
	}

	// gofmt

	fmtBytes, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(string(buf.Bytes()))
		return nil, fmt.Errorf("gofmt: %s", err)
	}

	return fmtBytes, nil
}

func lineToEmoji(line string) *Emoji {
	return nil
}
