package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gosimple/slug"
	"github.com/ritchies/ctftool/pkg/ctftime"
	"github.com/spf13/cobra"
)

// !TODO: Change this to use an emoji parser
type Emoji string // Emoji is a custom type for emojis

var EventID int // EventID is the ID of the event to retrieve

var (
	SpeechBalloon  Emoji = "\U0001f4ac" // speech balloon
	TriangularFlag Emoji = "\U0001f6a9" // triangular flag
	// SUBGROUP: keycaps
	KeycapHash     Emoji = "#\ufe0f\u20e3" // keycap: #
	KeycapAsterisk Emoji = "*\ufe0f\u20e3" // keycap: *
	Keycap0        Emoji = "0\ufe0f\u20e3" // keycap: 0
	Keycap1        Emoji = "1\ufe0f\u20e3" // keycap: 1
	Keycap2        Emoji = "2\ufe0f\u20e3" // keycap: 2
	Keycap3        Emoji = "3\ufe0f\u20e3" // keycap: 3
	Keycap4        Emoji = "4\ufe0f\u20e3" // keycap: 4
	Keycap5        Emoji = "5\ufe0f\u20e3" // keycap: 5
	Keycap6        Emoji = "6\ufe0f\u20e3" // keycap: 6
	Keycap7        Emoji = "7\ufe0f\u20e3" // keycap: 7
	Keycap8        Emoji = "8\ufe0f\u20e3" // keycap: 8
	Keycap9        Emoji = "9\ufe0f\u20e3" // keycap: 9
	Keycap10       Emoji = "\U0001f51f"    // keycap: 10
	// SUBGROUP: time
	HourglassDone    Emoji = "\u231b"           // hourglass done
	HourglassNotDone Emoji = "\u23f3"           // hourglass not done
	Watch            Emoji = "\u231a"           // watch
	AlarmClock       Emoji = "\u23f0"           // alarm clock
	Stopwatch        Emoji = "\u23f1\ufe0f"     // stopwatch
	TimerClock       Emoji = "\u23f2\ufe0f"     // timer clock
	MantelpieceClock Emoji = "\U0001f570\ufe0f" // mantelpiece clock
	TwelveOClock     Emoji = "\U0001f55b"       // twelve o’clock
	TwelveThirty     Emoji = "\U0001f567"       // twelve-thirty
	OneOClock        Emoji = "\U0001f550"       // one o’clock
	OneThirty        Emoji = "\U0001f55c"       // one-thirty
	TwoOClock        Emoji = "\U0001f551"       // two o’clock
	TwoThirty        Emoji = "\U0001f55d"       // two-thirty
	ThreeOClock      Emoji = "\U0001f552"       // three o’clock
	ThreeThirty      Emoji = "\U0001f55e"       // three-thirty
	FourOClock       Emoji = "\U0001f553"       // four o’clock
	FourThirty       Emoji = "\U0001f55f"       // four-thirty
	FiveOClock       Emoji = "\U0001f554"       // five o’clock
	FiveThirty       Emoji = "\U0001f560"       // five-thirty
	SixOClock        Emoji = "\U0001f555"       // six o’clock
	SixThirty        Emoji = "\U0001f561"       // six-thirty
	SevenOClock      Emoji = "\U0001f556"       // seven o’clock
	SevenThirty      Emoji = "\U0001f562"       // seven-thirty
	EightOClock      Emoji = "\U0001f557"       // eight o’clock
	EightThirty      Emoji = "\U0001f563"       // eight-thirty
	NineOClock       Emoji = "\U0001f558"       // nine o’clock
	NineThirty       Emoji = "\U0001f564"       // nine-thirty
	TenOClock        Emoji = "\U0001f559"       // ten o’clock
	TenThirty        Emoji = "\U0001f565"       // ten-thirty
	ElevenOClock     Emoji = "\U0001f55a"       // eleven o’clock
	ElevenThirty     Emoji = "\U0001f566"       // eleven-thirty
	HyperLink        Emoji = "\U0001f517"       // link
)

// ctftimeEventCmd represents the event command
var ctftimeEventCmd = &cobra.Command{
	Use:     "event",
	Short:   "Get information about a CTF event by ID",
	Long:    `Display information about a CTF event by ID.`,
	Example: `  ctftool ctftime event --event-id 12345`,
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) { // if args is not an integer, exit
		if len(args) > 0 {
			if _, err := fmt.Sscanf(args[0], "%d", &EventID); err != nil {
				log.Errorf("%v", err)
				return
			}
		}

		if EventID != 0 {
			event, err := ctftime.GetCTFEvent(EventID)
			CheckErr(err)

			json, err := json.MarshalIndent(event, "", "  ")
			CheckErr(err)

			fmt.Println(string(json))
		} else {
			events, err := ctftime.GetCTFEvents()
			CheckErr(err)

			now := time.Now()
			nextWeek := now.AddDate(0, 0, 7-int(now.Weekday()))

			thisWeekEvents := make([]ctftime.Event, 0)

			for _, event := range events {
				if event.Start.After(now) && event.Start.Before(nextWeek) {
					thisWeekEvents = append(thisWeekEvents, event)
				}
			}

			numberEmojis := []string{string(Keycap1), string(Keycap2), string(Keycap3), string(Keycap4), string(Keycap5), string(Keycap6), string(Keycap7), string(Keycap8), string(Keycap9), string(Keycap10)}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 0, '\t', 0)
			fmt.Fprintf(w, "# This week's CTF lineup\n\n")
			for i, event := range thisWeekEvents {
				ctfName := event.Title
				ctftimeURL := event.CTFTimeURL
				ctfDescription := event.Description

				// if ctf weight != 0, add it to the name
				if event.Weight != 0 {
					ctfName = fmt.Sprintf("%s (%.2f)", ctfName, event.Weight)
				}

				discordChannel := slug.Make(event.Title)

				// make sure the channel name is not too long (100 chars max)
				if len(discordChannel) > 99 {
					for words := strings.Split(discordChannel, "-"); len(words) > 2; words = strings.Split(discordChannel, "-") {
						discordChannel = strings.Join(words[:len(words)-1], "-")
					}
				}

				ctfTimes := make(map[string]string)
				ctfTimes["UTC"] = fmt.Sprintf("%s — %s", event.Start.UTC().Format("Mon, 02 Jan 2006 15:04"), event.Finish.UTC().Format("Mon, 02 Jan 2006 15:04"))

				fmt.Fprintf(w, "%s %s\n", numberEmojis[i], ctfName)
				fmt.Fprintf(w, "%s #%s\n", SpeechBalloon, discordChannel)
				fmt.Fprintf(w, "%s %s UTC\n", TwelveOClock, ctfTimes["UTC"])
				fmt.Fprintf(w, "%s %s\n", TriangularFlag, ctftimeURL)
				fmt.Fprintf(w, "%s %s\n\n", HyperLink, event.URL)

				// for every line in CTF description, add >>> to the start of the line
				for _, line := range strings.Split(ctfDescription, "\n") {
					fmt.Fprintf(w, ">>> %s\n", line)
				}

				// separator
				fmt.Fprintf(w, "---\n\n")
			}

			// flush
			w.Flush()
		}
	},
}

func init() {
	ctftimeCmd.AddCommand(ctftimeEventCmd)

	ctftimeEventCmd.Flags().IntVar(&EventID, "event-id", 0, "Unique identifier of the event")
}
