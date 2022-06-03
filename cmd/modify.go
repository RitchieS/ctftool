package cmd

import (
	"time"

	"github.com/ritchies/ctftool/internal/storage"
	"github.com/ritchies/ctftool/pkg/ctftime"
	"github.com/spf13/cobra"
)

// ctftimeCustomCmd represents the modify command
var ctftimeCustomCmd = &cobra.Command{
	Use:     "custom",
	Aliases: []string{"modify"},
	Short:   "Add a custom field to the ctf events database",
	Long:    `Modify the ctf events database with a custom field`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("modify called")

		// Get ID
		id, err := cmd.Flags().GetUint64("id")
		if err != nil {
			log.Fatal(err)
		}

		// Check if ID is set
		if id == 0 {
			log.Fatal("ID is not set")
		}

		// Get title, description, url, start, finish or hidden are set
		title, err := cmd.Flags().GetString("title")
		if err != nil {
			log.Fatal(err)
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			log.Fatal(err)
		}

		uri, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal(err)
		}

		start, err := cmd.Flags().GetString("start")
		if err != nil {
			log.Fatal(err)
		}

		finish, err := cmd.Flags().GetString("finish")
		if err != nil {
			log.Fatal(err)
		}

		hidden, err := cmd.Flags().GetBool("hidden")
		if err != nil {
			log.Fatal(err)
		}

		// Check if any of the fields are set
		if title == "" && description == "" && uri == "" && start == "" && finish == "" && hidden {
			log.Fatal("No fields set")
		}

		// Get the database
		db, err := dB.Get()
		if err != nil {
			log.Fatalf("Error getting db: %s", err)
		}

		// Check if the event exists
		var event ctftime.Event
		if err := db.Where("id = ?", id).First(&event).Error; err != nil {
			log.Fatalf("Error getting event: %s", err)
		}

		// Modify the event
		if title != "" {
			if event.Title != title {
				// Create a custom title
				var customTitle storage.EventCustomTitle

				customTitle.ID = event.ID
				customTitle.Title = title

				// Save the custom title
				if err := db.Create(&customTitle).Error; err != nil {
					log.Fatalf("Error creating custom title: %s", err)
				}
			}
		}

		if description != "" {
			if event.Description != description {
				// Create a custom description
				var customDescription storage.EventCustomDescription

				customDescription.ID = event.ID
				customDescription.Description = description

				if err := db.Create(&customDescription).Error; err != nil {
					log.Fatalf("Error creating custom description: %s", err)
				}
			}
		}

		if uri != "" {
			if event.URL != uri {
				// Create a custom url
				var customURL storage.EventCustomURL

				customURL.ID = event.ID
				customURL.URL = uri

				if err := db.Create(&customURL).Error; err != nil {
					log.Fatalf("Error creating custom url: %s", err)
				}
			}
		}

		var customDate storage.EventCustomDate

		if start != "" {
			parsedTime, err := time.Parse(time.RFC3339, start)
			if err != nil {
				// print example time format
				exampleTime := time.Now().Format(time.RFC3339)
				log.Fatalf("Error parsing start time: %s. Example time format: %s", err, exampleTime)
			}

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
			if err != nil {
				// print example time format
				exampleTime := time.Now().Format(time.RFC3339)
				log.Fatalf("Error parsing finish time: %s. Example time format: %s", err, exampleTime)
			}

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
		if err := db.Save(&event).Error; err != nil {
			log.Fatalf("Error saving event: %s", err)
		}

	},
}

func init() {
	ctftimeCmd.AddCommand(ctftimeCustomCmd)

	// Here you will define your flags and configuration settings.
	// id
	ctftimeCustomCmd.Flags().Uint64("id", 0, "The ID of the event to modify")
	cobra.MarkFlagRequired(ctftimeCustomCmd.Flags(), "id")

	// title
	ctftimeCustomCmd.Flags().String("title", "", "The title of the event")

	// description
	ctftimeCustomCmd.Flags().String("description", "", "The description of the event")

	// url
	ctftimeCustomCmd.Flags().String("url", "", "The URL of the event")

	// start
	ctftimeCustomCmd.Flags().String("start", "", "The start date of the event")

	// finish
	ctftimeCustomCmd.Flags().String("finish", "", "The finish date of the event")

	// hidden
	ctftimeCustomCmd.Flags().Bool("hidden", false, "Whether the event is hidden")
}
