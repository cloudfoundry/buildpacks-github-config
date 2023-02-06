package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	entrypoint string
)

func TestEntrypoint(t *testing.T) {
	var Expect = NewWithT(t).Expect
	var err error
	SetDefaultEventuallyTimeout(5 * time.Second)

	entrypoint, err = gexec.Build("github.com/cloudfoundry/buildpacks-github-config/actions/dependency/deprecation-list/entrypoint")
	Expect(err).NotTo(HaveOccurred())

	spec.Run(t, "deprecation-list", func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually
		)

		context("success cases", func() {
			context("at least one dependency is in the deprecation window", func() {
				var buildpack string
				var err error
				var tempDir string

				it.Before(func() {
					tempDir = t.TempDir()
					buildpack, err = os.MkdirTemp(tempDir, "buildpack")
					Expect(err).NotTo(HaveOccurred())
					err = os.WriteFile(filepath.Join(buildpack, "manifest.yml"),
						[]byte(`
dependency_deprecation_dates:
- version_line: 18.x
  name: first-dep
  date: 2022-12-12
  link: https://first-link
- version_line: 19.x
  name: second-dep
  date: 2023-01-10
  link: https://second-link
- version_line: 20.x
  name: third-dep
  date: 2024-01-01
  link: https://third-link
`), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("outputs deprecated deps", func() {
					command := exec.Command(
						entrypoint,
						"--buildpack", buildpack,
						"--buffer-days", "10",
						"--reference-date", "2023-01-01",
					)
					command.Env = []string{
						fmt.Sprintf("GITHUB_OUTPUT=%s", filepath.Join(tempDir, "github-output")),
					}

					buffer := gbytes.NewBuffer()
					session, err := gexec.Start(command, buffer, buffer)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0),
						func() string { return fmt.Sprintf("output:\n%s\n", buffer.Contents()) },
					)
					Expect(string(buffer.Contents())).
						To(ContainSubstring(`first-dep Version Line 18.x is already past deprecation date`))
					Expect(string(buffer.Contents())).
						To(ContainSubstring(`second-dep Version Line 19.x is within 10 days of deprecation`))

					data, err := os.ReadFile(filepath.Join(tempDir, "github-output"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(data)).To(ContainSubstring(`- Name: first-dep
Version Line: 18.x
Date: 2022-12-12
Link: https://first-link`))
					Expect(string(data)).To(ContainSubstring(`- Name: second-dep
Version Line: 19.x
Date: 2023-01-10
Link: https://second-link`))
					Expect(string(data)).NotTo(ContainSubstring(`20.x`))
				})
			})

			context("no dependency is in the deprecation window", func() {
				var buildpack string
				var err error
				var tempDir string

				it.Before(func() {
					tempDir = t.TempDir()
					buildpack, err = os.MkdirTemp(tempDir, "buildpack")
					Expect(err).NotTo(HaveOccurred())
					err = os.WriteFile(filepath.Join(buildpack, "manifest.yml"),
						[]byte(`
dependency_deprecation_dates:
- version_line: 18.x
  name: first-dep
  date: 2032-12-12
  link: https://first-link
- version_line: 19.x
  name: second-dep
  date: 2033-01-10
  link: https://second-link
- version_line: 20.x
  name: third-dep
  date: 2034-01-01
  link: https://third-link
`), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("outputs no deprecated deps", func() {
					command := exec.Command(
						entrypoint,
						"--buildpack", buildpack,
						"--buffer-days", "10",
						"--reference-date", "2023-01-01",
					)
					command.Env = []string{
						fmt.Sprintf("GITHUB_OUTPUT=%s", filepath.Join(tempDir, "github-output")),
					}

					buffer := gbytes.NewBuffer()
					session, err := gexec.Start(command, buffer, buffer)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0),
						func() string { return fmt.Sprintf("output:\n%s\n", buffer.Contents()) },
					)
					Expect(string(buffer.Contents())).
						To(ContainSubstring(`No deprecated dependencies found in the buildpack`))

					_, err = os.ReadFile(filepath.Join(tempDir, "github-output"))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			context("buildpack does not list deprecation dates", func() {
				var buildpack string
				var err error
				var tempDir string

				it.Before(func() {
					tempDir = t.TempDir()
					buildpack, err = os.MkdirTemp(tempDir, "buildpack")
					Expect(err).NotTo(HaveOccurred())
					err = os.WriteFile(filepath.Join(buildpack, "manifest.yml"),
						[]byte(``), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("exits early with code 0", func() {
					command := exec.Command(
						entrypoint,
						"--buildpack", buildpack,
						"--buffer-days", "10",
					)
					command.Env = []string{
						fmt.Sprintf("GITHUB_OUTPUT=%s", filepath.Join(tempDir, "github-output")),
					}

					buffer := gbytes.NewBuffer()
					session, err := gexec.Start(command, buffer, buffer)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0),
						func() string { return fmt.Sprintf("output:\n%s\n", buffer.Contents()) },
					)
					Expect(string(buffer.Contents())).
						To(ContainSubstring(`Exiting. Buildpack does not list deprecation dates`))

					_, err = os.ReadFile(filepath.Join(tempDir, "github-output"))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})

		context("failure cases", func() {
			context("missing required input buildpack", func() {
				it("exits with error msg", func() {
					command := exec.Command(
						entrypoint,
					)

					buffer := gbytes.NewBuffer()
					session, err := gexec.Start(command, buffer, buffer)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(1),
						func() string { return fmt.Sprintf("output:\n%s\n", buffer.Contents()) },
					)
					Expect(string(buffer.Contents())).
						To(ContainSubstring(`missing required input "buildpack`))
				})
			})

			context("input reference-date is in the wrong format", func() {
				it("exits with error msg", func() {
					command := exec.Command(
						entrypoint,
						"--buildpack", "some-buildpack",
						"--reference-date", "this-is-not-a-date",
					)

					buffer := gbytes.NewBuffer()
					session, err := gexec.Start(command, buffer, buffer)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(1),
						func() string { return fmt.Sprintf("output:\n%s\n", buffer.Contents()) },
					)
					Expect(string(buffer.Contents())).
						To(ContainSubstring(`parsing time "this-is-not-a-date" as "2006-01-02": cannot parse "this-is-not-a-date"`))
				})
			})

			context("buildpack has no top-level manifest.yml", func() {
				var buildpack string
				var err error
				var tempDir string

				it.Before(func() {
					tempDir = t.TempDir()
					buildpack, err = os.MkdirTemp(tempDir, "buildpack")
					Expect(err).NotTo(HaveOccurred())
					_, err = os.Stat(filepath.Join(buildpack, "manifest.yml"))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})

				it("exits with error msg", func() {
					command := exec.Command(
						entrypoint,
						"--buildpack", buildpack,
						"--buffer-days", "10",
					)
					command.Env = []string{
						fmt.Sprintf("GITHUB_OUTPUT=%s", filepath.Join(tempDir, "github-output")),
					}

					buffer := gbytes.NewBuffer()
					session, err := gexec.Start(command, buffer, buffer)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(1),
						func() string { return fmt.Sprintf("output:\n%s\n", buffer.Contents()) },
					)
					Expect(string(buffer.Contents())).
						To(ContainSubstring(`manifest.yml: no such file or directory`))

					_, err = os.ReadFile(filepath.Join(tempDir, "github-output"))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	}, spec.Report(report.Terminal{}), spec.Parallel())
}
