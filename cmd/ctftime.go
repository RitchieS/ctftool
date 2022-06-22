package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

const padding = 3                         // padding is the amount of padding to add to the left and right of bubbletea's list
const ctftimeURL = "https://ctftime.org/" // ctftimeURL is the URL of the CTFTime website

// ctftimeCmd represents the ctftime command
var ctftimeCmd = &cobra.Command{
	Use:     "ctftime",
	Aliases: []string{"time"},
	Short:   "Query CTFTime",
	Long:    `Retrieve information about upcoming CTF events and teams from CTFTime.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctftimeEventsCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(ctftimeCmd)

	ctftimeCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Limit the number of events to display")
}

func newModel(items []string) paginatorModel {
	for i, item := range items {
		items[i] = strings.TrimSpace(item)
	}

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 10
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	p.SetTotalPages(len(items))

	return paginatorModel{
		paginator: p,
		items:     items,
	}
}

type paginatorModel struct {
	items     []string
	paginator paginator.Model
}

func (m paginatorModel) Init() tea.Cmd {
	return nil
}

func (m paginatorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m paginatorModel) View() string {
	var b strings.Builder

	b.WriteString("\n  Upcoming CTF Events\n\n")

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
