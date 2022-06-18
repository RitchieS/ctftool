package cmd

import (
	"time"

	"github.com/ritchies/ctftool/internal/storage"
	"github.com/ritchies/ctftool/pkg/ctf"
	"github.com/spf13/cobra"
)

// ctftimeModifyCmd represents the modify command
var ctftimeModifyCmd = &cobra.Command{
	Use:     "modify",
	Aliases: []string{"custom"},
	Short:   "Add a custom field to the ctf events database",
	Long:    `Modify the ctf events database with a custom field`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("modify called")

		// Get ID
		id, err := cmd.Flags().GetUint64("id")
		CheckErr(err)

		// Check if ID is set
		if id == 0 {
			log.Fatal("ID is not set")
		}

		// Get title, description, url, start, finish or hidden are set
		title, err := cmd.Flags().GetString("title")
		CheckErr(err)

		description, err := cmd.Flags().GetString("description")
		CheckErr(err)

		uri, err := cmd.Flags().GetString("url")
		CheckErr(err)

		start, err := cmd.Flags().GetString("start")
		CheckErr(err)

		finish, err := cmd.Flags().GetString("finish")
		CheckErr(err)

		hidden, err := cmd.Flags().GetBool("hidden")
		CheckErr(err)

		// Check if any of the fields are set
		if title == "" && description == "" && uri == "" && start == "" && finish == "" && hidden {
			log.Fatal("No fields set")
		}

		// Get the database
		db, err := dB.Get()
		CheckErr(err)

		// Check if the event exists
		var event ctf.Event
		err = db.Where("id = ?", id).First(&event).Error
		CheckErr(err)

		// Modify the event
		if title != "" {
			if event.Title != title {
				// Create a custom title
				var customTitle storage.EventCustomTitle

				customTitle.ID = event.ID
				customTitle.Title = title

				// Save the custom title
				err := db.Create(&customTitle).Error
				CheckErr(err)
			}
		}

		if description != "" {
			if event.Description != description {
				// Create a custom description
				var customDescription storage.EventCustomDescription

				customDescription.ID = event.ID
				customDescription.Description = description

				err := db.Create(&customDescription).Error
				CheckErr(err)
			}
		}

		if uri != "" {
			if event.URL != uri {
				// Create a custom url
				var customURL storage.EventCustomURL

				customURL.ID = event.ID
				customURL.URL = uri

				err := db.Create(&customURL).Error
				CheckErr(err)
			}
		}

		var customDate storage.EventCustomDate

		if start != "" {
			parsedTime, err := time.Parse(time.RFC3339, start)
			CheckErr(err)

			inUTC := parsedTime.UTC()

			if event.Start != inUTC {
				db.Where("id = ?", event.ID).Find(&customDate)

				customDate.ID = event.ID
				customDate.Start = inUTC

				db.Save(&customDate)
			}
		}

		if finish != "" {
			parsedTime, err := time.Parse(time.RFC3339, finish)
			CheckErr(err)

			inUTC := parsedTime.UTC()

			if event.Finish != inUTC {
				db.Where("id = ?", event.ID).Find(&customDate)

				customDate.ID = event.ID
				customDate.Finish = inUTC

				db.Save(&customDate)
			}
		}

		if hidden {
			event.Hidden = hidden
		}

		// Save the event
		err = db.Save(&event).Error
		CheckErr(err)
	},
}

func init() {
	ctftimeCmd.AddCommand(ctftimeModifyCmd)

	ctftimeModifyCmd.Flags().Uint64("id", 0, "The ID of the event to modify")
	cobra.MarkFlagRequired(ctftimeModifyCmd.Flags(), "id")

	ctftimeModifyCmd.Flags().String("title", "", "The title of the event")
	ctftimeModifyCmd.Flags().String("description", "", "The description of the event")
	ctftimeModifyCmd.Flags().String("url", "", "The URL of the event")
	ctftimeModifyCmd.Flags().String("start", "", "The start date of the event")
	ctftimeModifyCmd.Flags().String("finish", "", "The finish date of the event")
	ctftimeModifyCmd.Flags().Bool("hidden", false, "Whether the event is hidden")
}
