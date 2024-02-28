package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/inkmi/eris/internal"
)

type sortedEntries struct {
	path  string
	order int
	depth int
	isDir bool
}

// sortByOrder implements sort.Interface based on the `order` field
//type sortByOrder []sortedEntries
//
//func (s sortByOrder) Len() int {
//	return len(s)
//}
//func (s sortByOrder) Swap(i, j int) {
//	s[i], s[j] = s[j], s[i]
//}
//func (s sortByOrder) Less(i, j int) bool {
//	return s[i].order < s[j].order
//}

func main() {
	var dirPath, outputPath string
	var leanPub bool
	var withToc bool
	flag.StringVar(&dirPath, "dir", "", "Directory to search for markdown files.")
	flag.StringVar(&outputPath, "out", "", "Output markdown file name.")
	flag.BoolVar(&leanPub, "leanpub", false, "Format as Leanpub")
	flag.BoolVar(&withToc, "withtoc", true, "Add a table of contents")
	flag.Parse()

	rtx := internal.RenderContext{
		LeanPub:           leanPub,
		IncludeLinkToFile: true,
		WithToc:           withToc,
		Toc:               make([]internal.TocEntry, 0),
	}

	if dirPath == "" || outputPath == "" {
		fmt.Println("Usage: --dir <directory path> --out <output file> --leanpub=false --withtoc=false")
		return
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	var builder strings.Builder
	rtx.Out = bufio.NewWriter(&builder)
	defer rtx.Out.Flush()

	if rtx.LeanPub {
		outputFile.WriteString("{:: encoding=\"utf-8\" /}\n\n")
	}
	processDirectory(&rtx, dirPath, 1)

	if rtx.WithToc {
		outputFile.WriteString("<p><b>Table Of Contents</b></p>")
		outputFile.WriteString(RenderTOC(rtx.Toc))
		outputFile.WriteString("\n\n")
	}

	outputFile.WriteString(builder.String())
}

func RenderTOC(toc []internal.TocEntry) string {
	var sb strings.Builder
	currentLevel := 0

	for _, entry := range toc {
		for entry.Level > currentLevel {
			sb.WriteString("<ul>")
			currentLevel++
		}
		for entry.Level < currentLevel {
			sb.WriteString("</ul>")
			currentLevel--
		}
		if len(entry.Id) > 0 {
			sb.WriteString(fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>", entry.Id, entry.Entry))
		} else {
			sb.WriteString(fmt.Sprintf("<li>%s</li>", entry.Entry))
		}
	}

	for currentLevel > 0 {
		sb.WriteString("</ul>")
		currentLevel--
	}

	return sb.String()
}

func stringToAnchor(input string) string {
	// Replace multiple spaces with a single dash
	spaceRegex := regexp.MustCompile(`\s+`)
	result := spaceRegex.ReplaceAllString(input, "-")

	// Remove non-alphanumeric characters
	alphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	result = alphanumericRegex.ReplaceAllString(result, "")

	return strings.ToLower(result)
}

func processDirectory(rtx *internal.RenderContext, dirPath string, depth int) {
	entries, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", dirPath, err)
		return
	}

	// Prepare entries for sorting
	var sorted []sortedEntries
	for _, entry := range entries {
		name := entry.Name()
		order := getOrder(name)
		sorted = append(sorted, sortedEntries{
			path:  filepath.Join(dirPath, name),
			order: order,
			depth: depth,
			isDir: entry.IsDir(),
		})
	}

	// Sort entries
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].order == sorted[j].order {
			return sorted[i].path < sorted[j].path
		}
		return sorted[i].order < sorted[j].order
	})

	// Process entries
	for _, entry := range sorted {
		if entry.isDir {
			processDirectory(rtx, entry.path, depth+1)
		} else {
			if strings.HasSuffix(entry.path, ".md") {
				content, err := ioutil.ReadFile(entry.path)
				if err != nil {
					fmt.Printf("Error reading file %s: %v\n", entry.path, err)
					continue
				}
				c := internal.ParseFrontMatter(string(content))
				if strings.HasSuffix(entry.path, "index.md") {
					c.Index = true
				}
				if rtx.LeanPub {
					if c.PageBreak {
						rtx.Out.WriteString("{pagebreak}\n")
					}
					if c.Sample {
						rtx.Out.WriteString("{sample: true}\n")
					} else {
						rtx.Out.WriteString("{sample: false}\n")
					}
					if c.Frontmatter {
						rtx.Out.WriteString("{frontmatter}\n")
					} else if c.Backmatter {
						rtx.Out.WriteString("{backmatter}\n")
					}
					if c.Rule > 0 {
						rtx.Out.WriteString(fmt.Sprintf("C> Rule \\#%d\n\n", c.Rule))
					}
				}
				if len(c.Id) == 0 {
					c.Id = stringToAnchor(c.Title)
				}
				if c.Title != "" {
					depth := entry.depth
					if c.Index {
						depth = depth - 1
					}
					if rtx.WithToc {
						rtx.Toc = append(rtx.Toc, internal.TocEntry{
							Level: depth,
							Entry: c.Title,
							Id:    c.Id,
						})
					}
					if len(c.Id) > 0 && rtx.LeanPub {
						rtx.Out.WriteString(fmt.Sprintf("%s %s {#%s}\n\n", strings.Repeat("#", depth), c.Title, c.Id))
					} else {
						rtx.Out.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", depth), c.Title))

					}
				}

				rtx.Out.WriteString(internal.InsertSourceCode(rtx, c.Body) + "\n\n")
			}
		}
	}
}

func getOrder(name string) int {
	if name == "index.md" {
		return -1 // Ensures index.md is processed first
	}
	num, err := strconv.Atoi(strings.Split(name, " ")[0])
	if err == nil {
		return num // Use numeric prefix as order if present
	}
	return 1000000
}
