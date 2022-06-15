package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/internal/storage"
	"github.com/ritchies/ctftool/pkg/ctf"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm/clause"
)

// ctftimeEventsCmd represents the events command
var ctftimeEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Get information about CTF events",
	Long:  `Display the current and upcoming CTF events from CTFTime.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := ctf.NewClient(nil)
		client.BaseURL, _ = url.Parse(ctftimeURL)

		events, err := client.GetCTFEvents()
		if err != nil {
			log.Fatalf("Error getting events: %s", err)
		}

		db, err := dB.Get()
		if err != nil {
			log.Fatalf("Error getting db: %s", err)
		}

		err = db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"title",
				"description",
				"url",
				"url_is_ctfd",
				"logo",
				"weight",
				"onsite",
				"location",
				"restrictions",
				"format",
				"format_id",
				"participants",
				"start",
				"finish",
			}),
		}).Create(&events).Error
		if err != nil {
			log.Fatalf("Error creating events in DB: %s", err)
		}

		eventStringsArray := make([]string, 0)

		// Make sure active events are at the top
		result := db.Order("finish asc, start asc, weight desc").Find(&events)
		if result.Error != nil {
			log.Fatal(result.Error)
		}

		for _, event := range events {

			if event.Hidden {
				continue
			}

			eventTitle := event.Title
			eventStart := event.Start
			eventFinish := event.Finish
			eventURL := event.URL
			eventTags := []string{}

			prettyETA := lib.HumanizeTime(eventStart)
			prettyWeight := lib.FtoaWithDigits(event.Weight, 2)

			var customTitle storage.EventCustomTitle
			err := db.Where("id = ?", event.ID).Find(&customTitle).Error
			if err == nil && customTitle.Title != "" {
				eventTitle = customTitle.Title
				event.Title = customTitle.Title
			}

			var customDate storage.EventCustomDate
			err = db.Where("id = ?", event.ID).Find(&customDate).Error
			if err == nil && customDate != (storage.EventCustomDate{}) {
				eventStart = customDate.Start
				eventFinish = customDate.Finish

				event.Start = customDate.Start
				event.Finish = customDate.Finish
			}

			var customURL storage.EventCustomURL
			err = db.Where("id = ?", event.ID).Find(&customURL).Error
			if err == nil && customURL.URL != "" {
				eventURL = customURL.URL
				event.URL = customURL.URL
			}

			// Check if CreatedAt is within the last 24 hours
			if event.CreatedAt.After(time.Now().Add(-24 * time.Hour)) {
				eventTags = append(eventTags, "NEW")
			}

			if event.FormatID != 1 {
				format := event.Format
				eventTags = append(eventTags, strings.Replace(format, "Attack-Defense", "AD", -1))
			}

			// !TODO: BUG
			/* if event.URLIsCTFD {
				eventTags = append(eventTags, "CTFD")
			} */

			if len(eventTags) > 0 {
				eventTitle = fmt.Sprintf("%s (%s)", eventTitle, strings.Join(eventTags, ", "))
				eventTitle = strings.TrimSuffix(eventTitle, ", ")
			}

			if event.Weight == 0 && eventFinish.Sub(eventStart).Hours() < 120 {
				prettyWeight = "TBD"
				prettyWeight = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "222", Dark: "222"}).Render(prettyWeight)
			} else {
				prettyWeight = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render(prettyWeight)
			}

			db.Save(&event)

			if ctf.IsCTFEventActive(event) {
				prettyETA = lib.RelativeTime(eventFinish, time.Now(), "ago", "left")

				log.WithFields(logrus.Fields{
					"id":     event.ID,
					"weight": event.Weight,
					"eta":    fmt.Sprintf("active (%s)", prettyETA),
					"url":    eventURL,
				}).Debug(event.Title)

				if eventFinish.Sub(eventStart).Hours() > 1 && eventFinish.Sub(eventStart).Hours() < 120 {
					prettyETA = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#00ff00",
						Dark:  "#00ff00",
					}).Render(fmt.Sprintf("%s - active", prettyETA))
				} else if eventFinish.Sub(eventStart).Hours() >= 120 {
					prettyETA = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#ffa500",
						Dark:  "#ffa500",
					}).Render(fmt.Sprintf("%s - active", prettyETA))
				}

				eventStringsArray = append(eventStringsArray, fmt.Sprintf("%d \t%s \t%s \t(%s)", event.ID, prettyWeight, eventTitle, prettyETA))
			} else {

				if event.Finish.Before(time.Now()) {
					continue
				}

				log.WithFields(logrus.Fields{
					"id":     event.ID,
					"weight": event.Weight,
					"eta":    prettyETA,
				}).Debug(event.Title)

				prettyEND := lib.FtoaWithDigits(eventFinish.Sub(eventStart).Hours(), 2)
				if eventFinish.Sub(eventStart).Hours() >= 120 {
					prettyEND = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#ffa500",
						Dark:  "#ffa500",
					}).Render(prettyEND)
				}

				prettyETA = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
					Light: "#888888",
					Dark:  "#888888",
				}).Render(fmt.Sprintf("%s for %s hours", prettyETA, prettyEND))

				eventStringsArray = append(eventStringsArray, fmt.Sprintf("%d \t%s \t%s \t(%s)", event.ID, prettyWeight, eventTitle, prettyETA))
			}
		}

		for i := len(eventStringsArray) - 1; i >= 0; i-- {
			if eventStringsArray[i] == "" {
				eventStringsArray = append(eventStringsArray[:i], eventStringsArray[i+1:]...)
			}
		}

		if PrintPretty {
			p := tea.NewProgram(newModel(eventStringsArray))
			if err := p.Start(); err != nil {
				log.Fatalf("Error creating tea program: %s", err)
			}
		} else {

			w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

			for _, eventString := range eventStringsArray {
				fmt.Fprintln(w, eventString)
			}
			w.Flush()
		}

	},
}

func init() {
	ctftimeCmd.AddCommand(ctftimeEventsCmd)

}
