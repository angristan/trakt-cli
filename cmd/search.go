package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/angristan/trakt-cli/api"
	"github.com/briandowns/spinner"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/muesli/termenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for movies and TV shows",
	Long:  `Search for movies and TV shows on Trakt.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewAPIClient()

		query := strings.Join(args, " ")

		searchType, err := cmd.Flags().GetString("type")
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get type flag")
		}

		s := spinner.New(spinner.CharSets[2], 100*time.Millisecond)
		s.Start()
		s.Prefix = fmt.Sprintf("Searching for '%s'... ", query)

		results, err := client.Search(query, searchType)
		if err != nil {
			s.Stop()
			logrus.WithError(err).Fatal("Search failed")
		}

		s.Stop()

		if len(results) == 0 {
			fmt.Println("No results found.")
			return
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{
			termenv.String("Type").Bold(),
			termenv.String("Title").Bold(),
			termenv.String("Year").Bold(),
			termenv.String("IMDB").Bold(),
		})

		for _, r := range results {
			switch r.Type {
			case "movie":
				if r.Movie != nil {
					t.AppendRow([]interface{}{
						"Movie",
						r.Movie.Title,
						r.Movie.Year,
						r.Movie.Ids.Imdb,
					})
				}
			case "show":
				if r.Show != nil {
					t.AppendRow([]interface{}{
						"TV Show",
						r.Show.Title,
						r.Show.Year,
						r.Show.Ids.Imdb,
					})
				}
			}
		}

		t.SetStyle(table.StyleRounded)
		t.Render()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("type", "t", "movie,show", "Type to search for (movie, show, or movie,show)")
}
