package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/lipgloss"
	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/pkg/ctftime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

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
		events, err := ctftime.GetCTFEvents()
		if err != nil {
			log.Fatalf("Error getting events: %s", err)
		}

		eventStringsArray := make([]string, len(events))

		for i, event := range events {
			prettyETA := lib.HumanizeTime(event.Start)
			prettyWeight := lib.FtoaWithDigits(event.Weight, 2)

			if event.Weight == 0 {
				prettyWeight = "TBD"
				prettyWeight = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "222", Dark: "222"}).Render(prettyWeight)
			} else {
				prettyWeight = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render(prettyWeight)
			}

			if ctftime.IsCTFEventActive(event) {
				prettyETA = lib.RelativeTime(event.Finish, time.Now(), "ago", "left")
				eventURL := event.URL

				log.WithFields(logrus.Fields{
					"id":     event.ID,
					"weight": event.Weight,
					"eta":    fmt.Sprintf("active (%s)", prettyETA),
					"url":    eventURL,
				}).Debug(event.Title)

				if event.Finish.Sub(event.Start).Hours() > 1 && event.Finish.Sub(event.Start).Hours() < 120 {
					prettyETA = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#00ff00",
						Dark:  "#00ff00",
					}).Render(fmt.Sprintf("%s - active", prettyETA))
				} else if event.Finish.Sub(event.Start).Hours() > 120 {
					prettyETA = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#ff0000",
						Dark:  "#ff0000",
					}).Render(fmt.Sprintf("%s - active", prettyETA))

					prettyWeight = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#ff0000",
						Dark:  "#ff0000",
					}).Render("UNR")
				}

				eventStringsArray[i] = fmt.Sprintf("%d \t%s \t%s (%s)", event.ID, prettyWeight, event.Title, prettyETA)
			} else {
				log.WithFields(logrus.Fields{
					"id":     event.ID,
					"weight": event.Weight,
					"eta":    prettyETA,
				}).Debug(event.Title)

				prettyETA = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
					Light: "#888888",
					Dark:  "#888888",
				}).Render(prettyETA)

				prettyEND := lib.FtoaWithDigits(event.Finish.Sub(event.Start).Hours(), 2)
				if event.Finish.Sub(event.Start).Hours() > 120 {
					prettyEND = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#ff0000",
						Dark:  "#ff0000",
					}).Render(prettyEND)

					prettyWeight = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#ff0000",
						Dark:  "#ff0000",
					}).Render("UNR")
				} else if event.Finish.Sub(event.Start).Hours() < 120 && event.Finish.Sub(event.Start).Hours() > 1 {
					prettyEND = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
						Light: "#00bb00",
						Dark:  "#00bb00",
					}).Render(prettyEND)
				}

				eventStringsArray[i] = fmt.Sprintf("%d \t%s \t%s (%s for %s hours)", event.ID, prettyWeight, event.Title, prettyETA, prettyEND)
			}
		}

		p := tea.NewProgram(newModel(eventStringsArray))
		if err := p.Start(); err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(ctftimeCmd)
}
