package ctf

import (
	"testing"
	"time"
)

func TestIsCTFEventActive(t *testing.T) {

	now := time.Now()

	type args struct {
		event Event
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test 1",
			args: args{
				event: Event{
					Start:  now.Add(time.Hour * -1),
					Finish: now.Add(time.Hour * 24 * 14),
				},
			},
			want: true,
		},
		{
			name: "Test 2",
			args: args{
				event: Event{
					Start:  now.Add(time.Hour * 42),
					Finish: now.Add(time.Hour * 24 * 14),
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCTFEventActive(tt.args.event); got != tt.want {
				t.Errorf("IsCTFEventActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Normal description",
			args: args{
				description: "This is a description without a link",
			},
			want: "This is a description without a link",
		},
		{
			name: "Description with a link",
			args: args{
				description: "This is a description with a link: http://google.com",
			},
			want: "This is a description with a link: http://google.com",
		},
		{
			name: "Long Description over 1024 characters",
			args: args{
				description: "The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog.",
			},
			want: "", // "The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanDescription(tt.args.description); got != tt.want {
				t.Errorf("CleanDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

// test CleanCTFEvents(events []Event) ([]Event, error)
func TestCleanCTFEvents(t *testing.T) {
	now := time.Now()

	fakeEvents := []Event{
		{
			Title:        "Working event",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     1,
			Start:        now.Add(time.Hour * -1),
			Finish:       now.Add(time.Hour * 24 * 14),
			Restrictions: "Open",
		},
		{
			Title:        "Onsite event",
			Description:  "This is a description",
			Onsite:       true,
			FormatID:     1,
			Start:        now.Add(time.Hour * -1),
			Finish:       now.Add(time.Hour * 24 * 14),
			Restrictions: "Open",
		},
		{
			Title:        "Invalid formatID",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     2,
			Start:        now.Add(time.Hour * -1),
			Finish:       now.Add(time.Hour * 24 * 14),
			Restrictions: "Open",
		},
		{
			Title:        "Finished event",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     1,
			Start:        now.Add(time.Hour * -200),
			Finish:       now.Add(time.Hour * -42),
			Restrictions: "Open",
		},
		{
			Title:        "Closed event",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     1,
			Start:        now.Add(time.Hour * -1),
			Finish:       now.Add(time.Hour * 24 * 14),
			Restrictions: "Closed",
		},
		{
			Title:        "Finished event 1",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     1,
			Start:        now.Add(time.Hour * 24 * 180),
			Finish:       now.Add(time.Hour * 24 * 365),
			Restrictions: "Open",
		},
		{
			Title:        "Future event",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     1,
			Start:        now.Add(time.Hour * 24 * 14),
			Finish:       now.Add(time.Hour * 24 * 21),
			Restrictions: "Open",
		},
		{
			Title:        "Future event 2",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     1,
			Start:        now.Add(time.Hour * 24 * 15),
			Finish:       now.Add(time.Hour * 24 * 22),
			Restrictions: "Open",
		},
		{
			Title:        "Active event 1",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     1,
			Start:        now.Add(time.Hour * -1),
			Finish:       now.Add(time.Hour * 24 * 14),
			Restrictions: "Open",
		},
		{
			Title:        "Active event 2",
			Description:  "This is a description",
			Onsite:       false,
			FormatID:     1,
			Start:        now.Add(time.Hour * -2),
			Finish:       now.Add(time.Hour * 24 * 14),
			Restrictions: "Open",
		},
	}

	// start testing
	for _, tt := range fakeEvents {
		t.Run(tt.Title, func(t *testing.T) {
			_, err := CleanCTFEvents(fakeEvents)
			if err != nil {
				t.Errorf("CleanCTFEvents(1) error = %v", err)
			}

		})
	}

	// test sorting
	t.Run("Sort", func(t *testing.T) {
		events, err := CleanCTFEvents(fakeEvents)
		if err != nil {
			t.Errorf("CleanCTFEvents(2) error = %v", err)
		}

		if events[0].Title != "Working event" {
			t.Errorf("CleanCTFEvents(2) = %v, want %v", events[0].Title, "Working event")
		}
	})

	// no events
	t.Run("No events", func(t *testing.T) {
		_, err := CleanCTFEvents([]Event{})
		if err == nil {
			t.Errorf("CleanCTFEvents(3) expected error, got nil")
		}
	})
}
