package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/lipgloss"
	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/internal/storage"
	"github.com/ritchies/ctftool/pkg/ctftime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm/clause"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	PrintPretty bool
)

const padding = 3

func newModel(items []string) model {
	for i, item := range items {
		items[i] = strings.TrimSpace(item)
	}

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 10
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	p.SetTotalPages(len(items))

	return model{
		paginator: p,
		items:     items,
	}
}

type model struct {
	items     []string
	paginator paginator.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.paginator, cmd = m.paginator.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	// set msg to full white
	b.WriteString("\n  Upcoming CTF Events\n\n")

	// b.WriteString("\n    N \tWEIGHT \tID \tNAME\n")
	prettyHeader := fmt.Sprintf("    %s \t%s \t%s\n", "CTFID", "WEIGHT", "NAME")
	prettyHeader = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render(prettyHeader)
	b.WriteString("\n" + prettyHeader + "\n")

	start, end := m.paginator.GetSliceBounds(len(m.items))
	for _, item := range m.items[start:end] {
		b.WriteString("  • " + item + "\n")
	}

	if m.paginator.TotalPages > 1 {
		b.WriteString("  " + m.paginator.View())
	}

	b.WriteString("\n\n  h/l ←/→ page • q: quit\n")
	return b.String()
}

// ctftimeCmd represents the ctftime command
var ctftimeCmd = &cobra.Command{
	Use:     "ctftime",
	Aliases: []string{"time"},
	Short:   "Query CTFTime",
	Long:    `Retrieve information about upcoming CTF events and teams from CTFTime.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := ctftime.NewClient(nil)
		client.BaseURL, _ = url.Parse("https://ctftime.org/")

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

		eventStringsArray := make([]string, len(events))

		result := db.Order("start asc, finish asc, weight desc").Find(&events)
		if result.Error != nil {
			log.Fatal(result.Error)
		}

		for i, event := range events {

			if event.Hidden {
				continue
			}

			eventTitle := event.Title
			eventStart := event.Start
			eventFinish := event.Finish
			eventURL := event.URL

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

			if event.CreatedAt.Second() == event.UpdatedAt.Second() {
				// add (NEW)
				eventTitle = fmt.Sprintf("%s (NEW)", eventTitle)
			}

			if event.Weight == 0 && eventFinish.Sub(eventStart).Hours() < 120 {
				prettyWeight = "TBD"
				prettyWeight = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "222", Dark: "222"}).Render(prettyWeight)
			} else {
				prettyWeight = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render(prettyWeight)
			}

			db.Save(&event)

			if ctftime.IsCTFEventActive(event) {
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
				} else if eventFinish.Sub(eventStart).Hours() > 120 {
					prettyETA = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#ffa500",
						Dark:  "#ffa500",
					}).Render(fmt.Sprintf("%s - active", prettyETA))
				}

				eventStringsArray[i] = fmt.Sprintf("%d \t%s \t%s \t(%s)", event.ID, prettyWeight, eventTitle, prettyETA)
			} else {
				log.WithFields(logrus.Fields{
					"id":     event.ID,
					"weight": event.Weight,
					"eta":    prettyETA,
				}).Debug(event.Title)

				prettyEND := lib.FtoaWithDigits(eventFinish.Sub(eventStart).Hours(), 2)
				if eventFinish.Sub(eventStart).Hours() > 120 {
					prettyEND = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#ffa500",
						Dark:  "#ffa500",
					}).Render(prettyEND)
				}

				prettyETA = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
					Light: "#888888",
					Dark:  "#888888",
				}).Render(fmt.Sprintf("%s for %s hours", prettyETA, prettyEND))

				eventStringsArray[i] = fmt.Sprintf("%d \t%s \t%s \t(%s)", event.ID, prettyWeight, eventTitle, prettyETA)
			}
		}

		// remomve empty strings
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
				// fmt.Println(eventString)
				fmt.Fprintln(w, eventString)
			}
			w.Flush()
		}

	},
}

func init() {
	rootCmd.AddCommand(ctftimeCmd)

	ctftimeCmd.Flags().BoolVar(&PrintPretty, "interactive", false, "Interactive mode")
}
