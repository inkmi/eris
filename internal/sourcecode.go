package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type TocEntry struct {
	Level int
	Entry string
	Id    string
}
type RenderContext struct {
	LeanPub           bool
	IncludeLinkToFile bool
	WithToc           bool
	Out               *bufio.Writer
	Toc               []TocEntry
}

func InsertSourceCode(rtx *RenderContext, markdown string) string {
	re := regexp.MustCompile(`!\[\]\(([^,]+),\s*(\d+)\)`)

	// Replacement function
	replacer := func(s string) string {
		matches := re.FindStringSubmatch(s)
		if len(matches) < 3 {
			// Not enough matches, return the original string
			return s
		}
		filename := matches[1]
		markerNum, err := strconv.Atoi(matches[2])
		if err != nil {
			// If conversion fails, return the original string
			fmt.Println("Error converting marker number:", err)
			return s
		}

		replacement, err := cutPartsFromFile(filename, markerNum)
		if err != nil {
			// Handle error, for now, just print it
			fmt.Println("Error calling cutPartsFromFile:", err)
			return s
		}

		lexer := lexers.Get("go")
		if lexer == nil {
			lexer = lexers.Fallback
		}
		lexer = chroma.Coalesce(lexer)
		style := styles.Get("swapoff")
		if style == nil {
			style = styles.Fallback
		}
		formatter := html.New(html.Standalone(false))

		iterator, err := lexer.Tokenise(nil, string(replacement))
		var buf bytes.Buffer // Create a new bytes.Buffer
		if err := formatter.Format(&buf, style, iterator); err != nil {
			panic(err)
		}

		result := "```go\n" + normalizeIndentation(replacement) + "```\n"
		if rtx.IncludeLinkToFile {
			result = result + fmt.Sprintf("From [%s](%s)", filename, filename)
		}
		return result
	}

	// Replace all markers with their corresponding file content
	return re.ReplaceAllStringFunc(markdown, replacer)
}

func normalizeIndentation(code string) string {
	// Split the code into lines
	lines := strings.Split(code, "\n")

	// Initialize the variables
	var result []string
	indentLevel := 0
	indentSize := 2 // number of spaces per indentation level

	for _, line := range lines {
		// Trim leading and trailing spaces to check for braces
		trimmedLine := strings.TrimSpace(line)

		// Decrease the indentation level if the line starts with a closing brace
		if strings.HasPrefix(trimmedLine, "}") || strings.HasPrefix(trimmedLine, ")") {
			indentLevel--
		}

		// Apply indentation
		indentation := strings.Repeat(" ", max(indentLevel*indentSize, 0))
		result = append(result, fmt.Sprintf("%s%s", indentation, trimmedLine))

		// Increase the indentation level for the next line if the current line ends with an opening brace
		if strings.HasSuffix(trimmedLine, "{") || strings.HasSuffix(trimmedLine, "(") {
			indentLevel++
		}
	}

	// Join the lines back together
	return strings.Join(result, "\n")
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// cutPartsFromFile opens a file, searches for parts between markers // S:<number> and // E:<number>, and returns them as a string.
func cutPartsFromFile(filename string, markerNum int) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	startMarker := fmt.Sprintf("// S:%d", markerNum)
	endMarker := fmt.Sprintf("// E:%d", markerNum)
	isBetweenMarkers := false

	for scanner.Scan() {
		line := scanner.Text()
		trimLine := strings.TrimSpace(line)
		// Check if the line matches the marker format, allowing for spaces
		if matchMarker(trimLine, startMarker) {
			isBetweenMarkers = true
			continue // Skip the start marker itself
		} else if matchMarker(trimLine, endMarker) {
			isBetweenMarkers = false
			continue // Skip the end marker itself
		}

		if isBetweenMarkers {
			result.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}

// matchMarker checks if a given line matches a marker, ignoring additional spaces.
func matchMarker(line, marker string) bool {
	// Remove all spaces to allow for flexible matching
	cleanLine := strings.ReplaceAll(line, " ", "")
	cleanMarker := strings.ReplaceAll(marker, " ", "")
	return cleanLine == cleanMarker
}
