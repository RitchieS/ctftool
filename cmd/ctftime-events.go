package cmd

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/internal/storage"
	"github.com/ritchies/ctftool/pkg/ctf"
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

		db, err := dB.Get()
		CheckErr(err)

		// use gorm to get the first xEvents from the events database
		var events []ctf.Event

		// check when the last update was to the database
		var lastUpdate time.Time
		err = db.Model(&events).Select("updated_at").Row().Scan(&lastUpdate)
		if err == nil {
			log.Debugf("Last update: %s", lib.HumanizeTime(lastUpdate))
		}

		if err == nil && time.Since(lastUpdate) <= time.Minute*10 {
			log.Debug("Not updating database, last update was within the last 10 minutes")
		} else {
			log.Info("Updating database")
			events, err = client.GetCTFEvents()
			CheckErr(err)

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
			CheckErr(err)
		}

		eventStringsArray := make([]string, 0)
		newEvents := make([]ctf.Event, 0)

		createdTimes := make([]time.Time, 0)
		for _, event := range events {
			createdTimes = append(createdTimes, event.CreatedAt)
		}

		createdTimes = lib.Unique(createdTimes)

		// Make sure active events are at the top
		err = db.Order("finish asc, start asc, weight desc").Find(&events).Error
		CheckErr(err)

		for _, event := range events {

			if event.Hidden {
				continue
			}

			eventTitle := event.Title
			eventStart := event.Start
			eventFinish := event.Finish
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
				event.URL = customURL.URL
			}

			// Check if CreatedAt is within the last 24 hours
			if event.CreatedAt.After(time.Now().Add(-24*time.Hour)) &&
				len(createdTimes) > 1 {
				eventTags = append(eventTags, "NEW")
				newEvents = append(newEvents, event)
			}

			switch event.FormatID {
			case 2:
				eventTags = append(eventTags, "AD")
			case 3:
				eventTags = append(eventTags, "HQ")
			}

			if event.Onsite {
				eventTags = append(eventTags, "ONSITE")
			}

			eventTitle = cleanTitle(eventTitle)

			// !TODO: BUG
			if event.URLIsCTFD {
				eventTags = append(eventTags, "CTFD")
			}

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

			if time.Since(lastUpdate) > time.Minute*10 {
				db.Save(&event)
			}

			if ctf.IsCTFEventActive(event) {
				prettyETA = lib.RelativeTime(eventFinish, time.Now(), "ago", "left")

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

		if options.Interactive {
			p := tea.NewProgram(newModel(eventStringsArray))
			err := p.Start()
			CheckErr(err)
		} else {

			w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.StripEscape)

			// ID WEIGHT TITLE ETA
			fmt.Fprintln(w, "ID\tWEIGHT\tTITLE\tETA")
			fmt.Fprintln(w, "----\t-----\t-----\t---")

			for i, eventString := range eventStringsArray {
				if i >= limit {
					break
				}
				fmt.Fprintln(w, eventString)
			}
			w.Flush()

			if len(newEvents) > 0 {
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, "New events:")
				fmt.Fprintln(w, "------------")
				for _, event := range newEvents {
					fmt.Fprintf(w, "%d \t%.2f \t%s \t(%s)\n", event.ID, event.Weight, cleanTitle(event.Title), lib.RelativeTime(event.Finish, time.Now(), "ago", "left"))
				}
			}

			// !TODO: add a legend to the bottom of the table
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "Legend:")
			fmt.Fprintln(w, "NEW = Event added in the last 24 hours")
			fmt.Fprintln(w, "AD = Attack Defend")
			fmt.Fprintln(w, "HQ = Hack Quest")
		}

	},
}
var limit int

func init() {
	ctftimeCmd.AddCommand(ctftimeEventsCmd)

	// limit
	ctftimeEventsCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Limit the number of events to display")
}

func cleanTitle(str string) string {
	// remove all the text between ( and ) and [ and ]
	str = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(str, "")
	str = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(str, "")

	// remove anything between () and [] and then remove the empty () and []
	replacers := []string{
		`\(.*?\)`,
		`\[.*?\]`,
	}
	for _, replacer := range replacers {
		str = regexp.MustCompile(replacer).ReplaceAllString(str, "")
	}

	replaceArray := [][]string{
		{"Capture the Flag", "CTF"},
		{"Attack Defend", "AD"},
		{"Hack Quest", "HQ"},
		{"Qualification Round", "Quals"},
		{"Qualification", "Quals"},
	}
	for _, replace := range replaceArray {
		// case insensitive replace
		r := regexp.MustCompile(fmt.Sprintf(`(?i)%s`, replace[0]))
		str = r.ReplaceAllString(str, replace[1])
	}

	str = strings.Trim(str, "-_ ")

	if len(str) > 20 {
		str = strings.Replace(str, fmt.Sprintf(" %d", time.Now().Year()), "", -1)
	}

	r := regexp.MustCompile(`\s+`)
	str = r.ReplaceAllString(str, " ")

	return str
}
