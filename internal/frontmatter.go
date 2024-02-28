package internal

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
)

type Content struct {
	Title     string
	Body      string
	PageBreak bool
	Sample    bool
	// BACKMATTER
	Backmatter  bool
	Frontmatter bool
	Rule        int
	Index       bool
	Id          string
}

func ParseFrontMatter(c string) Content {
	delimiterRegex := regexp.MustCompile(`^\s*-{3,}\s*$`)
	scanner := bufio.NewScanner(strings.NewReader(c))
	inFrontMatter := false
	var bodyLines []string
	var frontMatterLines []string

	var content Content

	for scanner.Scan() {
		line := scanner.Text()
		if delimiterRegex.MatchString(line) && !inFrontMatter {
			inFrontMatter = true
			continue
		} else if delimiterRegex.MatchString(line) {
			inFrontMatter = false
			continue
		}

		if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line)
		} else {
			bodyLines = append(bodyLines, line)
		}
	}
	content = parseContent(bufio.NewScanner(strings.NewReader(strings.Join(frontMatterLines, "\n"))))
	content.Body = strings.Join(bodyLines, "\n")

	return content
}

func parseContent(scanner *bufio.Scanner) Content {
	var c Content
	for scanner.Scan() {
		line := scanner.Text()
		key, value := parseLine(line)

		switch key {
		case "title":
			c.Title = value
		case "id":
			c.Id = value
		case "page-break":
			c.PageBreak = value == "yes"
		case "sample":
			c.Sample = value == "yes"
		case "backmatter":
			c.Backmatter = value == "yes"
		case "frontmatter":
			c.Frontmatter = value == "yes"
		case "rule":
			v, err := strconv.Atoi(value)
			if err != nil {
				panic(err)
			}
			c.Rule = v
		}
	}
	return c
}

func parseLine(line string) (key, value string) {
	if idx := strings.Index(line, ":"); idx != -1 {
		key = strings.ToLower(line[:idx]) // Ensure key is lowercase for consistent comparison
		value = strings.TrimSpace(line[idx+1:])
	}
	return
}
