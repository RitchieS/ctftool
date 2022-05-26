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

type Emoji string

var EventID int

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

// add "conference" and "ctf" to the IgnoreList words
type IgnoreList []string

var (
	Ignored = IgnoreList{
		"conference",
		"ctf",
		"qualifiers",
		"qualifier",
		"quals",
		"qual",
	}
)

// eventCmd represents the event command
var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Get information about CTF events",
	Long:  `Display the current and upcoming CTF events from CTFTime.`,
	Aliases: []string{
		"events",
	},
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("event called")

		// if args is not an integer, exit
		if len(args) > 0 {
			if _, err := fmt.Sscanf(args[0], "%d", &EventID); err != nil {
				log.Errorf("%v", err)
				return
			}
		}

		if EventID != 0 {

			event, err := ctftime.GetCTFEvent(EventID)
			if err != nil {
				log.Fatalf("Error getting event: %s", err)
			}

			json, err := json.MarshalIndent(event, "", "  ")
			if err != nil {
				log.Fatalf("Error marshalling event: %s", err)
			}
			fmt.Println(string(json))

		} else {

			events, err := ctftime.GetCTFEvents()
			if err != nil {
				log.Fatalf("Error getting events: %s", err)
			}

			// Get the events for the next 7 days
			now := time.Now()

			nextWeek := now.AddDate(0, 0, 7)

			thisWeekEvents := make([]ctftime.Event, 0)
			// nextWeekEvents := make([]ctftime.Event, 0)

			for _, event := range events {
				if event.Start.After(now) && event.Start.Before(nextWeek) {
					thisWeekEvents = append(thisWeekEvents, event)
				} /*else {
					nextWeekEvents = append(nextWeekEvents, event)
				} */
			}

			// print events in markdown
			// # This week's CTF lineup
			// 1️⃣ XXCTF (CTF Name)
			// 💬 #XXCTF (discord channel)
			// 🕛 Fri, 15 April 2022, 16:00 UTC — Sun, 17 April 2022, 16:00 UTC (time in UTC)
			// 🕖 Fri, 15 April 2022, 12:00 EDT — Sun, 17 April 2022, 12:00 EDT (time in EDT)
			// 🚩 <ctftime url>
			// >>> <description>

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

				// discord channel
				var discordChannel string

				discordChannel = event.Title

				// use slug to clean up the discord channel name
				discordChannel = slug.Make(discordChannel)

				// make sure the channel name is not too long (100 chars max), split on dashes
				if len(discordChannel) > 100 {
					for words := strings.Split(discordChannel, "-"); len(words) > 2; words = strings.Split(discordChannel, "-") {
						discordChannel = strings.Join(words[:len(words)-1], "-")
					}
				}

				// get the current time zone of the operating system
				currentTimeZone, err := time.LoadLocation(time.Now().Location().String())
				if err != nil {
					log.Fatalf("Error loading time zone: %s", err)
				}

				localLoc, _ := time.LoadLocation(currentTimeZone.String())

				loc, _ := time.LoadLocation("America/New_York")
				ctfTimes := make(map[string]string)
				ctfTimes["LOCAL"] = fmt.Sprintf("%s — %s", event.Start.In(localLoc).Format("Mon, 02 Jan 2006 15:04"), event.Finish.In(localLoc).Format("Mon, 02 Jan 2006 15:04"))
				ctfTimes["UTC"] = fmt.Sprintf("%s — %s", event.Start.UTC().Format("Mon, 02 Jan 2006 15:04"), event.Finish.UTC().Format("Mon, 02 Jan 2006 15:04"))
				ctfTimes["EDT"] = fmt.Sprintf("%s — %s", event.Start.In(loc).Format("Mon, 02 Jan 2006 15:04"), event.Finish.In(loc).Format("Mon, 02 Jan 2006 15:04"))

				fmt.Fprintf(w, "%s  %s\n", numberEmojis[i], ctfName)
				fmt.Fprintf(w, "%s #%s\n", SpeechBalloon, discordChannel)
				fmt.Fprintf(w, "%s %s\n", TwelveOClock, ctfTimes["LOCAL"])
				fmt.Fprintf(w, "%s %s UTC\n", TwelveOClock, ctfTimes["UTC"])
				fmt.Fprintf(w, "%s %s EDT\n", SevenOClock, ctfTimes["EDT"])
				fmt.Fprintf(w, "%s %s\n", TriangularFlag, ctftimeURL)
				fmt.Fprintf(w, "%s %s\n", HyperLink, event.URL)
				fmt.Fprintf(w, "%s\n\n", ctfDescription)

				// separator
				fmt.Fprintf(w, "---\n\n")
			}

			// flush
			w.Flush()

		}

	},
}

func init() {
	ctftimeCmd.AddCommand(eventCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eventCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// eventCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	eventCmd.Flags().IntVar(&EventID, "id", 0, "Event ID")
}
