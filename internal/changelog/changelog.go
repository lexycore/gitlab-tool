package changelog

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/xanzy/go-gitlab"

	"github.com/lexycore/gitlab-tools/internal/config"
)

// Item contains changelog item
type Item struct {
	Package    string
	Version    string
	Release    string
	Urgency    string
	Changes    string
	Maintainer string
	Date       string
}

var (
	rePackage        = regexp.MustCompile(`(Package)(\:)( *)(.*)`)
	reItem           = regexp.MustCompile(`(?P<package>[\S]*)( *)\((?P<version>.*)\)( *)(?P<release>.*)\;( *)urgency=(?P<urgency>.*)([\n\r]*)(?P<changes>[\n\r\s\S]*)([\n\r]*)(\s+--\s+)(?P<maintainer>[\s\S]+\<[\S\@]+\>)(\s*)(?P<date>\w+\,\s\d+\s\w+\s\d+\s[\d\:]+\s[\+\d+]*)`)
	reItemGroupNames = reItem.SubexpNames()
)

func Add(git *gitlab.Client, cfg *config.Config, path *config.GitLabPath, newItem *Item) error {
	changelog, err := os.Open("debian/changelog")
	if err != nil {
		return err
	}

	var prevRecord *Item
	prevRecord, err = readChangelogItem(changelog, 0)
	if err != nil {
		fmt.Println("Warning:", err.Error())
	}

	if newItem.Package == "" {
		newItem.Package, err = determinePackage(prevRecord)
		if err != nil {
			return err
		}
	}

	if newItem.Version == "" {
		newItem.Version, err = determineVersion(prevRecord)
		if err != nil {
			return err
		}
	}

	if newItem.Release == "" {
		newItem.Release, err = determineRelease(prevRecord)
		if err != nil {
			return err
		}
	}

	if newItem.Urgency == "" {
		newItem.Urgency, err = determineUrgency(prevRecord)
		if err != nil {
			return err
		}
	}

	if newItem.Changes == "" {
		newItem.Changes, err = determineChanges(prevRecord, git, cfg, path)
		if err != nil {
			return err
		}
	}

	if newItem.Maintainer == "" {
		newItem.Maintainer, err = determineMaintainer(prevRecord)
		if err != nil {
			return err
		}
	}

	if newItem.Date == "" {
		newItem.Date, err = determineDate(prevRecord)
		if err != nil {
			return err
		}
	}

	return appendChangelogItem(changelog, newItem)
}

func determinePackage(prevItem *Item) (string, error) {
	if prevItem == nil || prevItem.Package == "" {
		return getPackageName()
	}
	return prevItem.Package, nil
}

func determineVersion(prevItem *Item) (string, error) {
	if prevItem == nil || prevItem.Version == "" {
		return "", errors.New("error: empty prevItem.Version")
	}
	// TODO: parse version and increment last part of it
	return prevItem.Version, nil
}

func determineRelease(prevItem *Item) (string, error) {
	if prevItem == nil || prevItem.Release == "" {
		return "UNRELEASED", nil
	}
	return prevItem.Release, nil
}

func determineUrgency(prevItem *Item) (string, error) {
	if prevItem == nil || prevItem.Urgency == "" {
		return "medium", nil
	}
	return prevItem.Urgency, nil
}

func determineMaintainer(prevItem *Item) (string, error) {
	if prevItem == nil {
		return "", errors.New("error: empty prevItem")
	}
	return prevItem.Maintainer, nil
}

func determineDate(prevItem *Item) (string, error) {
	// if prevItem == nil {
	// 	return "", errors.New("error: empty prevItem")
	// }
	// return prevItem.Date, nil
	return time.Now().Format(time.RFC1123Z), nil
}

func getPackageName() (string, error) {
	control, err := os.Open("debian/control")
	if err != nil {
		return "", err
	}

	controlText := make([]byte, 0, 1024)
	_, err = control.Read(controlText)
	if err != nil {
		return "", err
	}

	out := rePackage.FindAllStringSubmatch(string(controlText), -1)
	for i, v := range out {
		value := strings.TrimSpace(v[len(v)-1])
		fmt.Printf("%d: %s\n", i, value)
	}

	return "", errors.New("package name not found")
}

func readChangelogItem(changelog *os.File, idx int) (*Item, error) {
	if changelog == nil {
		return nil, errors.New("error: nil changelog file pointer")
	}

	count := 0
	offset := int64(0)
	var buf string
	newItem := Item{}

	_, err := changelog.Seek(offset, 0)
	scanner := bufio.NewScanner(changelog)

	for scanner.Scan() {
		changelogText := scanner.Text()
		if err == nil {
			if len(buf) > 0 {
				buf += "\n"
			}
			buf += changelogText
		}

		for _, match := range reItem.FindAllStringSubmatch(buf, -1) {
			if count < idx {
				buf = buf[len(match[0]):len(buf)]
				offset += int64(len(match[0]))
				count++
				continue
			}
			for groupIdx, value := range match {
				switch reItemGroupNames[groupIdx] {
				case "package":
					newItem.Package = value
				case "version":
					newItem.Version = value
				case "release":
					newItem.Release = value
				case "urgency":
					newItem.Urgency = value
				case "changes":
					newItem.Changes = strings.Trim(strings.Trim(value, "\n"), "\r")
				case "maintainer":
					newItem.Maintainer = value
				case "date":
					newItem.Date = value
				}
			}
			return &newItem, nil
		}
	}
	return nil, errors.New("error: could not find changelog item")
}

func appendChangelogItem(changelog *os.File, newItem *Item) error {
	tpl, err := template.New("changelog").Parse(`{{.Package}} ({{.Version}}) {{.Release}}; urgency={{.Urgency}}

{{.Changes}}

 -- {{.Maintainer}}  {{.Date}}

`)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	err = tpl.Execute(buf, newItem)
	if err != nil {
		return err
	}

	_, err = changelog.Seek(0, 0)
	if err != nil {
		return err
	}

	newChangelog, err := os.Create("debian/changelog.new")
	if err != nil {
		return err
	}
	defer func() {
		_ = newChangelog.Close()
	}()

	// append at the start
	_, err = newChangelog.Write(buf.Bytes())
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(changelog)

	// read the file to be appended to and output all of it
	for scanner.Scan() {

		_, err = newChangelog.WriteString(scanner.Text())
		_, err = newChangelog.WriteString("\n")
	}

	if err = scanner.Err(); err != nil {
		panic(err)
	}
	// ensure all lines are written
	_ = newChangelog.Sync()

	// overwrite the old file with the new one
	err = os.Rename("debian/changelog.new", "debian/changelog")
	if err != nil {
		panic(err)
	}
	return nil
}
