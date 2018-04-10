package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jaytaylor/html2text"
	"github.com/spf13/cobra"
)

var (
	Quiet        bool
	Verbose      bool
	OmitLinks    bool
	PrettyTables bool
	Markdown     bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "Activate quiet log output")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Activate verbose log output")
	rootCmd.PersistentFlags().BoolVarP(&OmitLinks, "omit-links", "o", false, "Omit special links output")
	rootCmd.PersistentFlags().BoolVarP(&PrettyTables, "pretty-tables", "p", false, "Activate pretty tables output")
	rootCmd.PersistentFlags().BoolVarP(&Markdown, "markdown", "m", false, "Activate markdown mode")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		errorExit(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "html2text",
	Short: "Converts HTML to plaintext or markdown",
	Long:  "Converts HTML to plaintext or markdown",
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(_ *cobra.Command, _ []string) {
		initLogging()
	},
	Run: func(cmd *cobra.Command, args []string) {
		var (
			options = html2text.Options{
				OmitLinks:    OmitLinks,
				PrettyTables: PrettyTables,
				Markdown:     Markdown,
			}
			f   *os.File
			s   string
			err error
		)
		if args[0] == "-" {
			f = os.Stdin
		} else {
			if f, err = os.Open(args[0]); err != nil {
				errorExit(fmt.Errorf("opening %q: %s", args[0], err))
			}
		}
		if s, err = html2text.FromReader(f, options); err != nil {
			errorExit(err)
		}
		fmt.Fprint(os.Stdout, s)
	},
}

func errorExit(err interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	os.Exit(1)
}

func initLogging() {
	level := log.InfoLevel
	if Verbose {
		level = log.DebugLevel
	}
	if Quiet {
		level = log.ErrorLevel
	}
	log.SetLevel(level)
}
