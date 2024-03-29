package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/pkg/ctftime"
	"github.com/spf13/cobra"
)

// ctftimeEventsCmd represents the events command
var ctftimeEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Get information about CTF events",
	Long: `Display the current and upcoming CTF events from CTFTime.
	
Legend:
 ONSITE = CTF requires team to be in person
 AD = Attack Defend
 HQ = Hack Quest`,
	Example: `  ctftool ctftime events --limit 10`,
	Run: func(cmd *cobra.Command, args []string) {
		var events []ctftime.Event

		events, err := ctftime.GetCTFEvents()
		CheckErr(err)

		eventStringsArray := make([]string, 0)

		for _, event := range events {
			eventTitle := event.Title
			eventStart := event.Start
			eventFinish := event.Finish
			eventTags := []string{}

			prettyETA := lib.HumanizeTime(eventStart)
			prettyWeight := lib.FtoaWithDigits(event.Weight, 2)

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

			if len(eventTags) > 0 {
				eventTitle = fmt.Sprintf("%s (%s)", eventTitle, strings.Join(eventTags, ", "))
			}

			switch {
			case event.Weight == 0 && eventFinish.Sub(eventStart).Hours() < 120:
				prettyWeight = "TBD"
				prettyWeight = colorize(prettyWeight, "222", "222")
			case event.Weight == 0:
				prettyWeight = "N/A"
				prettyWeight = colorize(prettyWeight, "223", "223")
			default:
				prettyWeight = colorize(prettyWeight, "235", "252")
			}

			if ctftime.IsActive(event) {
				prettyETA = lib.RelativeTime(eventFinish, time.Now(), "ago", "left")

				if eventFinish.Sub(eventStart).Hours() > 1 && eventFinish.Sub(eventStart).Hours() < 120 {
					prettyETA = colorize(fmt.Sprintf("%s - active", prettyETA), "#00ff00", "#00ff00")
				} else if eventFinish.Sub(eventStart).Hours() >= 120 {
					prettyETA = colorize(fmt.Sprintf("%s - active", prettyETA), "#ffa500", "#ffa500")
				}
			} else {
				if event.Finish.Before(time.Now()) {
					continue
				}

				prettyEND := lib.FtoaWithDigits(eventFinish.Sub(eventStart).Hours(), 1)
				prettyETA = fmt.Sprintf("%s for %s hours", prettyETA, prettyEND)
			}

			eventStringsArray = append(eventStringsArray, fmt.Sprintf("%d \t%s \t%s \t(%s)", event.ID, prettyWeight, eventTitle, prettyETA))
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', tabwriter.StripEscape)

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

	},
}

var limit int

func init() {
	ctftimeCmd.AddCommand(ctftimeEventsCmd)

	// limit
	ctftimeEventsCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Maximum number of events to display")
}

func cleanTitle(str string) string {
	// remove all the text between ( and ) and [ and ]
	replacers := [][]string{
		{`\(.*?\)`, ""},
		{`\[.*?\]`, ""},
		{"Capture the Flag", "CTF"},
		{"Attack Defend", "AD"},
		{"Hack Quest", "HQ"},
		{"Qualification Round", "Quals"},
		{"Qualification", "Quals"},
	}

	for _, replace := range replacers {
		str = regexp.MustCompile(fmt.Sprintf(`(?i)%s`, replace[0])).ReplaceAllString(str, replace[1])
	}

	str = strings.Trim(str, "-_ ")

	if len(str) > 20 {
		str = strings.Replace(str, fmt.Sprintf(" %d", time.Now().Year()), "", -1)
	}

	str = regexp.MustCompile(`\s+`).ReplaceAllString(str, " ")

	return str
}

func colorize(text string, light string, dark string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: light,
		Dark:  dark,
	}).Render(text)
}
