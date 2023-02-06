package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

func main() {
	var (
		config struct {
			BuildpackPath string
			BufferDays    int
			CurrentTime   time.Time
		}
		referenceDate string
		err           error
	)

	flag.StringVar(&config.BuildpackPath, "buildpack", "", "Path to the buildpack")
	flag.IntVar(&config.BufferDays, "buffer-days", 0, "Instructs the program to list all deps that will be deprecated as on this many days in the future")
	flag.StringVar(&referenceDate, "reference-date", "", "(optional) Use a date other than system date as the current date e.g. --reference-date=2006-01-02")
	flag.Parse()

	if config.BuildpackPath == "" {
		log.Fatal(errors.New("missing required input \"buildpack\""))
	}

	config.CurrentTime = time.Now()
	if referenceDate != "" {
		config.CurrentTime, err = time.Parse(dateLayout, referenceDate)
		if err != nil {
			log.Fatal(err)
		}
	}

	logger := libbuildpack.NewLogger(os.Stdout)
	manifest, err := libbuildpack.NewManifest(config.BuildpackPath, logger, config.CurrentTime)
	if err != nil {
		log.Fatal(err)
	}

	deprecationMetadata := manifest.Deprecations
	if len(deprecationMetadata) < 1 {
		fmt.Println("Exiting. Buildpack does not list deprecation dates")
		return
	}

	deprecated, err := getDeprecatedEntries(deprecationMetadata, config.CurrentTime, config.BufferDays)
	if err != nil {
		log.Fatal(err)
	}
	if len(deprecated) < 1 {
		fmt.Println("No deprecated dependencies found in the buildpack")
		return
	}

	var output string
	for _, d := range deprecated {
		output += fmt.Sprintf("- Name: %s\n", d.Name)
		output += fmt.Sprintf("Version Line: %s\n", d.VersionLine)
		output += fmt.Sprintf("Date: %s\n", d.Date)
		output += fmt.Sprintf("Link: %s\n", d.Link)
		output += "\n"
	}
	setOutput("list", string(output))
}

func getDeprecatedEntries(deprecationMetadata []libbuildpack.DeprecationDate,
	currTime time.Time,
	bufferDays int) ([]libbuildpack.DeprecationDate, error) {

	ret := []libbuildpack.DeprecationDate{}
	bufferTime := time.Duration(bufferDays) * 24 * time.Hour

	for _, d := range deprecationMetadata {
		eolTime, err := time.Parse(dateLayout, d.Date)
		if err != nil {
			return ret, err
		}

		if eolTime.Before(currTime) {
			ret = append(ret, d)
			fmt.Printf("%s Version Line %s is already past deprecation date of %s\n", d.Name, d.VersionLine, d.Date)
			continue
		}

		if eolTime.Sub(currTime) < bufferTime {
			ret = append(ret, d)
			fmt.Printf("%s Version Line %s is within %d days of deprecation (%s)\n", d.Name, d.VersionLine, bufferDays, d.Date)
			continue
		}
	}
	return ret, nil
}

func setOutput(key, val string) {
	outputFileName, ok := os.LookupEnv("GITHUB_OUTPUT")
	if !ok {
		log.Fatal("GITHUB_OUTPUT is not set, see https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter")
	}
	file, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#multiline-strings
	delimiter := uuid.New().String()
	fmt.Fprintf(file, "%s<<%s\n", key, delimiter)
	fmt.Fprintf(file, "%s\n", val)
	fmt.Fprintf(file, "%s\n", delimiter)
}
