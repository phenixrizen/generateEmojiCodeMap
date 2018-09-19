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
	"strings"
	"text/template"
	"unicode/utf8"
)

// original unicode data file: https://unicode.org/Public/emoji/11.0/emoji-test.txt
// - processed to only include data lines
// - cat emoji-data.txt | grep '^[0-9]'
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

	var es []Emoji
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		e := lineToEmoji(line)
		es = append(es, e)
	}

	matchExp := ""

	emojiCodeMap := make(map[string]string)
	for i, e := range es {
		emojiCodeMap[e.Emoji] = e.Description
		matchExp = matchExp + e.Match
		if i != len(es)-1 {
			matchExp = matchExp + "|"
		}
	}

	// Template GenerateSource

	var buf bytes.Buffer
	t := template.Must(template.New("template").Parse(templateMapCode))
	if err := t.Execute(&buf, TemplateData{pkgName, matchExp, emojiCodeMap}); err != nil {
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

func lineToEmoji(line string) Emoji {
	e := Emoji{}

	line = strings.ToLower(line)

	// get match expr
	matchStr := strings.Split(line, ";")[0]
	matchAr := strings.Fields(matchStr)
	match := ""
	for _, m := range matchAr {
		match = match + fmt.Sprintf("\\\\x{%s}", m)
	}
	e.Match = match

	// get emoji
	emojiStr := line[strings.Index(line, "#")+1:]
	emojiAr := strings.Split(emojiStr, " ")
	e.Emoji = emojiAr[1]

	// get description
	r, _ := utf8.DecodeRuneInString(e.Emoji)
	rLength := len(e.Emoji)
	idx := strings.IndexRune(emojiStr, r)
	e.Description = emojiStr[idx+rLength:]

	return e
}
