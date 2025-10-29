package daytime

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDaytime_Valid(t *testing.T) {
	tests := []struct {
		name string
		d    Daytime
		want bool
	}{
		// --- Valid Cases (d <= EndOfDay) ---

		{
			name: "StartOfDay (00:00:00) is valid",
			d:    StartOfDay,
			want: true,
		},
		{
			name: "Midday (12:00:00) is valid",
			d:    Must(12, 0, 0), // 43200
			want: true,
		},
		{
			name: "Just before EndOfDay (23:59:59) is valid",
			d:    Must(23, 59, 59), // 86399
			want: true,
		},
		{
			name: "EndOfDay (24:00:00) is valid",
			d:    EndOfDay, // 86400
			want: true,
		},

		// --- Invalid Cases (d > EndOfDay) ---

		{
			name: "Just beyond EndOfDay is invalid",
			d:    EndOfDay + 1, // 86401
			want: false,
		},
		{
			name: "Large invalid value is invalid",
			d:    Daytime(100000),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.Valid()
			if got != tt.want {
				t.Errorf("Daytime(%d).Valid() got = %v, want %v", tt.d, got, tt.want)
			}
		})
	}

	// Also explicitly test a zero value Daytime (which is StartOfDay)
	t.Run("Zero value is valid", func(t *testing.T) {
		var d Daytime // d = 0
		if !d.Valid() {
			t.Errorf("Zero-valued Daytime should be valid, got false")
		}
	})
}

func TestDaytime_IsEndOfDay(t *testing.T) {
	tests := []struct {
		name string
		d    Daytime
		want bool
	}{
		{
			name: "EndOfDay (24:00:00) returns true",
			d:    EndOfDay, // 86400
			want: true,
		},
		{
			name: "StartOfDay (00:00:00) returns false",
			d:    StartOfDay, // 0
			want: false,
		},
		{
			name: "Just before EndOfDay (23:59:59) returns false",
			d:    Must(23, 59, 59), // 86399
			want: false,
		},
		{
			name: "Midday (12:00:00) returns false",
			d:    Must(12, 0, 0), // 43200
			want: false,
		},
		{
			name: "Invalid daytime (beyond EndOfDay) returns false",
			d:    EndOfDay + 1, // 86401
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.IsEndOfDay()
			if got != tt.want {
				t.Errorf("Daytime(%d).IsEndOfDay() got = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

func TestDaytime_IsInDay(t *testing.T) {
	tests := []struct {
		name string
		d    Daytime
		want bool
	}{
		{
			name: "StartOfDay (00:00:00) is in day",
			d:    StartOfDay, // 0
			want: true,
		},
		{
			name: "Just after StartOfDay is in day",
			d:    StartOfDay + 1, // 1
			want: true,
		},
		{
			name: "Midday (12:00:00) is in day",
			d:    Must(12, 0, 0), // 43200
			want: true,
		},
		{
			name: "Just before EndOfDay (23:59:59) is in day",
			d:    Must(23, 59, 59), // 86399
			want: true,
		},
		{
			name: "EndOfDay (24:00:00) is NOT in day (exclusive upper bound)",
			d:    EndOfDay, // 86400
			want: false,
		},
		{
			name: "Invalid daytime (beyond EndOfDay) is NOT in day",
			d:    EndOfDay + 1, // 86401
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.IsInDay()
			if got != tt.want {
				t.Errorf("Daytime(%d).IsInDay() got = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		h    int
		m    int
		s    int
		want Daytime
		err  error
	}{
		// --- Valid Cases (Success) ---
		{
			name: "StartOfDay 00:00:00",
			h:    0, m: 0, s: 0,
			want: StartOfDay,
			err:  nil,
		},
		{
			name: "EndOfDay 24:00:00",
			h:    24, m: 0, s: 0,
			want: EndOfDay,
			err:  nil,
		},
		{
			name: "Midday 12:00:00",
			h:    12, m: 0, s: 0,
			want: Daytime(12 * 3600), // 43200
			err:  nil,
		},
		{
			name: "Just before midnight 23:59:59",
			h:    23, m: 59, s: 59,
			want: Daytime(86399),
			err:  nil,
		},
		{
			name: "Example time 01:02:03",
			h:    1, m: 2, s: 3,
			want: Daytime(1*3600 + 2*60 + 3), // 3723
			err:  nil,
		},

		// --- Invalid Cases: ErrInvalidTimeComponent (Range) ---
		{
			name: "Negative hour -1:00:00",
			h:    -1, m: 0, s: 0,
			want: 0,
			err:  ErrInvalidTimeComponent,
		},
		{
			name: "Hour too large 25:00:00",
			h:    25, m: 0, s: 0,
			want: 0,
			err:  ErrInvalidTimeComponent,
		},
		{
			name: "Negative minute 00:-1:00",
			h:    0, m: -1, s: 0,
			want: 0,
			err:  ErrInvalidTimeComponent,
		},
		{
			name: "Minute too large 00:60:00",
			h:    0, m: 60, s: 0,
			want: 0,
			err:  ErrInvalidTimeComponent,
		},
		{
			name: "Negative second 00:00:-1",
			h:    0, m: 0, s: -1,
			want: 0,
			err:  ErrInvalidTimeComponent,
		},
		{
			name: "Second too large 00:00:60",
			h:    0, m: 0, s: 60,
			want: 0,
			err:  ErrInvalidTimeComponent,
		},

		// --- Invalid Cases: ErrEndOfDayExceeded (24:xx:xx) ---
		{
			name: "24:01:00 should error",
			h:    24, m: 1, s: 0,
			want: 0,
			err:  ErrEndOfDayExceeded,
		},
		{
			name: "24:00:01 should error",
			h:    24, m: 0, s: 1,
			want: 0,
			err:  ErrEndOfDayExceeded,
		},
		{
			name: "24:01:01 should error",
			h:    24, m: 1, s: 1,
			want: 0,
			err:  ErrEndOfDayExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.h, tt.m, tt.s)

			// Check for error type
			if tt.err != nil {
				if err == nil {
					t.Fatalf("New(%d, %d, %d) expected error %v, got nil", tt.h, tt.m, tt.s, tt.err)
				}
				if !errors.Is(err, tt.err) {
					t.Fatalf("New(%d, %d, %d) got error type %v, want %v. Full error: %v", tt.h, tt.m, tt.s, errors.Unwrap(err), tt.err, err)
				}
				return
			}

			// Check for success cases
			if err != nil {
				t.Fatalf("New(%d, %d, %d) expected nil error, got %v", tt.h, tt.m, tt.s, err)
			}
			if got != tt.want {
				t.Errorf("New(%d, %d, %d) got daytime %d, want %d", tt.h, tt.m, tt.s, got, tt.want)
			}
		})
	}
}

func TestMust_Success(t *testing.T) {
	tests := []struct {
		name string
		h    int
		m    int
		s    int
		want Daytime
	}{
		{
			name: "StartOfDay 00:00:00",
			h:    0, m: 0, s: 0,
			want: StartOfDay,
		},
		{
			name: "EndOfDay 24:00:00",
			h:    24, m: 0, s: 0,
			want: EndOfDay,
		},
		{
			name: "Midday 12:00:00",
			h:    12, m: 0, s: 0,
			want: Daytime(12 * 3600),
		},
		{
			name: "Just before midnight 23:59:59",
			h:    23, m: 59, s: 59,
			want: Daytime(86399),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Must should not panic for valid inputs
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Must(%d, %d, %d) panicked unexpectedly: %v", tt.h, tt.m, tt.s, r)
				}
			}()

			got := Must(tt.h, tt.m, tt.s)
			if got != tt.want {
				t.Errorf("Must(%d, %d, %d) got daytime %d, want %d", tt.h, tt.m, tt.s, got, tt.want)
			}
		})
	}
}

func TestMust_Panic(t *testing.T) {
	tests := []struct {
		name string
		h    int
		m    int
		s    int
		err  error
	}{
		// --- Invalid Cases: ErrInvalidTimeComponent (Range) ---
		{
			name: "Panic on Hour too large 25:00:00",
			h:    25, m: 0, s: 0,
			err: ErrInvalidTimeComponent,
		},
		{
			name: "Panic on Negative minute 00:-1:00",
			h:    0, m: -1, s: 0,
			err: ErrInvalidTimeComponent,
		},
		{
			name: "Panic on Second too large 00:00:60",
			h:    0, m: 0, s: 60,
			err: ErrInvalidTimeComponent,
		},

		// --- Invalid Cases: ErrEndOfDayExceeded (24:xx:xx) ---
		{
			name: "Panic on 24:01:00 (Exceeded EndOfDay)",
			h:    24, m: 1, s: 0,
			err: ErrEndOfDayExceeded,
		},
		{
			name: "Panic on 24:00:01 (Exceeded EndOfDay)",
			h:    24, m: 0, s: 1,
			err: ErrEndOfDayExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use defer/recover to catch the expected panic
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("Must(%d, %d, %d) did not panic, expected panic with error %v", tt.h, tt.m, tt.s, tt.err)
				}

				// The panic value should be a daytime.Error (the one returned by New)
				err, ok := r.(error)
				if !ok {
					t.Fatalf("Must(%d, %d, %d) panicked with non-error value: %v", tt.h, tt.m, tt.s, r)
				}

				// Check if the unwrapped error matches the expected error type
				if !errors.Is(err, tt.err) {
					t.Fatalf("Must(%d, %d, %d) panicked with wrong underlying error. Got: %v, Want: %v. Full panic value: %v", tt.h, tt.m, tt.s, errors.Unwrap(err), tt.err, err)
				}

				// Optionally check the Operation and Value of the wrapped error
				var daytimeErr *Error
				if errors.As(err, &daytimeErr) {
					if daytimeErr.Operation() != "New" {
						t.Errorf("Wrapped error operation got %q, want 'New'", daytimeErr.Operation())
					}
					wantValue := fmt.Sprintf("%02d:%02d:%02d", tt.h, tt.m, tt.s)
					if daytimeErr.Value() != wantValue {
						t.Errorf("Wrapped error value got %q, want %q", daytimeErr.Value(), wantValue)
					}
				}
			}()

			Must(tt.h, tt.m, tt.s)
		})
	}
}

func TestFromTime(t *testing.T) {
	const (
		refYear  = 2024
		refMonth = time.February
		refDay   = 15
	)
	local := time.FixedZone("TestLocal", 3*3600) // UTC+3
	utc := time.UTC

	tests := []struct {
		name string
		h    int
		m    int
		s    int
		loc  *time.Location
		want Daytime
	}{
		{
			name: "00:00:00 at UTC (StartOfDay)",
			h:    0, m: 0, s: 0,
			loc:  utc,
			want: StartOfDay, // 0
		},
		{
			name: "23:59:59 at UTC (Just before EndOfDay)",
			h:    23, m: 59, s: 59,
			loc:  utc,
			want: Daytime(86399), // 86399
		},
		{
			name: "12:30:15 at UTC",
			h:    12, m: 30, s: 15,
			loc:  utc,
			want: Daytime(12*3600 + 30*60 + 15), // 45015
		},
		{
			name: "08:00:00 in Local Zone (UTC+3)",
			h:    8, m: 0, s: 0,
			loc:  local,
			want: Daytime(8 * 3600), // 28800
		},
		{
			name: "17:45:00 in Local Zone (UTC+3)",
			h:    17, m: 45, s: 0,
			loc:  local,
			want: Daytime(17*3600 + 45*60), // 63900
		},
		{
			name: "Time 00:00:01",
			h:    0, m: 0, s: 1,
			loc:  utc,
			want: Daytime(1), // 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the time.Time object
			// Note: We intentionally use hour/minute/second directly because t.Time.Clock()
			// extracts the time in that specific time.Location, which is what FromTime relies on.
			tm := time.Date(refYear, refMonth, refDay, tt.h, tt.m, tt.s, 0, tt.loc)

			got := FromTime(tm)

			if got != tt.want {
				t.Errorf("FromTime(%s) got daytime %d, want %d", tm.Format("15:04:05.999"), got, tt.want)
			}

			// Ensure that time.Time corresponding to 24:00:00 (which is 00:00:00 next day)
			// converts to StartOfDay (0).
			// Example: time.Date(2024, 2, 16, 0, 0, 0, 0, utc) should be 0.
			t.Run("NextDayMidnight", func(t *testing.T) {
				nextDayTm := time.Date(refYear, refMonth, refDay+1, 0, 0, 0, 0, tt.loc)
				gotNextDay := FromTime(nextDayTm)
				if gotNextDay != StartOfDay {
					t.Errorf("FromTime(00:00:00 next day) got %d, want %d", gotNextDay, StartOfDay)
				}
			})
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Daytime
		wantErr bool
	}{
		{"Seconds format success (0)", "0", StartOfDay, false},
		{"Time format success (00:00:00)", "00:00:00", StartOfDay, false},
		{"Error: Empty string fails immediately", "", 0, true},
		{"Error: Invalid input fails both parsers", "invalid-input", 0, true},
		{"Check precedence: 1234 (seconds)", "1234", Daytime(1234), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error status mismatch: got error %v, want error? %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse(%q) got daytime %d, want %d", tt.input, got, tt.want)
			}

			if tt.wantErr && got != 0 {
				t.Errorf("Parse(%q) on error got daytime %d, want 0", tt.input, got)
			}
		})
	}
}

func Test_parseSeconds(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
		err  error
	}{
		{"StartOfDay (0)", "0", 0, nil},
		{"EndOfDay (86400)", "86400", 86400, nil},
		{"Mid-day value (43200)", "43200", 43200, nil},
		{"Value too large (86401)", "86401", 0, ErrValueOutOfRange},
		{"Negative value (-1)", "-1", 0, ErrValueOutOfRange},
		{"Non-integer string", "abc", 0, ErrInvalidFormat},
		{"Float string", "1.5", 0, ErrInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSeconds(tt.s)

			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("parseSeconds(%q) got error %v, want error %v", tt.s, err, tt.err)
				}
				if got != tt.want {
					t.Errorf("parseSeconds(%q) got seconds %d on error, want %d", tt.s, got, tt.want)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseSeconds(%q) got unexpected error: %v", tt.s, err)
			}
			if got != tt.want {
				t.Errorf("parseSeconds(%q) got seconds %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func Test_parseTimeString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
		err  error
	}{
		{"StartOfDay (00:00:00)", "00:00:00", 0, nil},
		{"EndOfDay (24:00:00)", "24:00:00", 86400, nil},
		{"Mid-day value (12:30:15)", "12:30:15", 45015, nil},
		{"Single digit components (01:02:03)", "1:2:3", 0, ErrInvalidFormat},
		{"Missing seconds part", "12:30", 0, ErrInvalidFormat},
		{"Hour too large (25)", "25:00:00", 0, ErrInvalidTimeComponent},
		{"Minute too large (60)", "00:60:00", 0, ErrInvalidTimeComponent},
		{"Negative component (Hour)", "-1:00:00", 0, ErrInvalidTimeComponent},
		{"24:01:00 (EndOfDayExceeded)", "24:01:00", 0, ErrEndOfDayExceeded},
		{"24:00:01 (EndOfDayExceeded)", "24:00:01", 0, ErrEndOfDayExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeString(tt.s)

			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("parseTimeString(%q) got error %v, want error %v", tt.s, err, tt.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseTimeString(%q) got unexpected error: %v", tt.s, err)
			}
			if got != tt.want {
				t.Errorf("parseTimeString(%q) got total seconds %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestDaytime_Clock(t *testing.T) {
	tests := []struct {
		name  string
		d     Daytime
		wantH int
		wantM int
		wantS int
	}{
		{name: "Start of day 00:00:00", d: StartOfDay, wantH: 0, wantM: 0, wantS: 0},
		{name: "End of day 24:00:00", d: EndOfDay, wantH: 24, wantM: 0, wantS: 0},
		{name: "Just before end of day 23:59:59", d: Daytime(86399), wantH: 23, wantM: 59, wantS: 59},
		{name: "Midday 12:00:00", d: Daytime(12 * 3600), wantH: 12, wantM: 0, wantS: 0},
		{name: "Complex time 17:30:45", d: Daytime(17*3600 + 30*60 + 45), wantH: 17, wantM: 30, wantS: 45},
		{name: "Just after midnight 00:00:01", d: Daytime(1), wantH: 0, wantM: 0, wantS: 1},
		{name: "Test with 1 hour 01:00:00", d: Daytime(3600), wantH: 1, wantM: 0, wantS: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, m, s := tt.d.Clock()

			if h != tt.wantH || m != tt.wantM || s != tt.wantS {
				t.Errorf("Clock() for daytime %d got %02d:%02d:%02d, want %02d:%02d:%02d",
					tt.d, h, m, s, tt.wantH, tt.wantM, tt.wantS)
			}
		})
	}
}

// TestDaytime_ComponentAccessors tests Hour, Minute, and Second methods.
func TestDaytime_ComponentAccessors(t *testing.T) {
	// Test Daytime for 17:30:45 (63045 seconds)
	d := Daytime(17*3600 + 30*60 + 45) // 63045 seconds

	t.Run("Hour 17:30:45", func(t *testing.T) {
		want := 17
		if got := d.Hour(); got != want {
			t.Errorf("Hour() got %d, want %d", got, want)
		}
	})

	t.Run("Minute 17:30:45", func(t *testing.T) {
		want := 30
		if got := d.Minute(); got != want {
			t.Errorf("Minute() got %d, want %d", got, want)
		}
	})

	t.Run("Second 17:30:45", func(t *testing.T) {
		want := 45
		if got := d.Second(); got != want {
			t.Errorf("Second() got %d, want %d", got, want)
		}
	})

	t.Run("Hour for EndOfDay 24:00:00", func(t *testing.T) {
		want := 24
		if got := EndOfDay.Hour(); got != want {
			t.Errorf("EndOfDay.Hour() got %d, want %d", got, want)
		}
	})

	t.Run("Minute for EndOfDay 24:00:00", func(t *testing.T) {
		want := 0
		if got := EndOfDay.Minute(); got != want {
			t.Errorf("EndOfDay.Minute() got %d, want %d", got, want)
		}
	})
}

func TestDaytime_Duration(t *testing.T) {
	tests := []struct {
		name string
		d    Daytime
		want time.Duration
	}{
		{name: "StartOfDay 00:00:00", d: StartOfDay, want: 0},
		{name: "EndOfDay 24:00:00 (24 hours)", d: EndOfDay, want: 24 * time.Hour},
		{name: "One hour 01:00:00", d: Daytime(3600), want: time.Hour},
		{name: "12 hours and 30 minutes 12:30:00", d: Daytime(12*3600 + 30*60), want: 12*time.Hour + 30*time.Minute},
		{name: "One second 00:00:01", d: Daytime(1), want: time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Duration(); got != tt.want {
				t.Errorf("Duration() for daytime %d got %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

var (
	D000000  = StartOfDay
	D010000  = Must(1, 0, 0)
	D060000  = Must(6, 0, 0)
	D120000  = Must(12, 0, 0)
	D123045  = Must(12, 30, 45)
	D180000  = Must(18, 0, 0)
	D230000  = Must(23, 0, 0)
	D235959  = Must(23, 59, 59)
	D240000  = EndOfDay
	DInvalid = Daytime(secondsInDay + 1)
)

func TestDaytime_CompareAndEqual(t *testing.T) {
	// Test Daytime.Equal
	t.Run("Equal", func(t *testing.T) {
		if !D120000.Equal(D120000) {
			t.Errorf("D120000 should equal D120000")
		}
		if D120000.Equal(D010000) {
			t.Errorf("D120000 should not equal D010000")
		}
		if !D240000.Equal(EndOfDay) {
			t.Errorf("D240000 should equal EndOfDay")
		}
	})

	// Test Daytime.Compare
	t.Run("Compare", func(t *testing.T) {
		tests := []struct {
			name  string
			d1    Daytime
			d2    Daytime
			wantC int // -1 (d1 < d2), 0 (d1 == d2), 1 (d1 > d2)
		}{
			{"d1 < d2 (Normal)", D010000, D120000, -1},
			{"d1 > d2 (Normal)", D120000, D010000, 1},
			{"d1 == d2 (Normal)", D120000, D120000, 0},
			{"d1 < d2 (Boundary: D235959 vs D240000)", D235959, D240000, -1},
			{"d1 > d2 (Boundary: D240000 vs D235959)", D240000, D235959, 1},
			{"d1 == d2 (Boundary: D240000 vs D240000)", D240000, D240000, 0},
			{"d1 < d2 (Start vs Mid)", D000000, D120000, -1},
			{"d1 > d2 (Mid vs Start)", D120000, D000000, 1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.d1.Compare(tt.d2); got != tt.wantC {
					t.Errorf("Compare(%s, %s) got %d, want %d", tt.d1.String(), tt.d2.String(), got, tt.wantC)
				}
			})
		}
	})
}

func TestDaytime_BeforeAndAfter(t *testing.T) {
	tests := []struct {
		name  string
		d1    Daytime
		d2    Daytime
		wantB bool // d1.Before(d2)
		wantA bool // d1.After(d2)
	}{
		{"D010000 vs D120000", D010000, D120000, true, false},
		{"D120000 vs D010000", D120000, D010000, false, true},
		{"D120000 vs D120000", D120000, D120000, false, false},
		// Boundary cases with EndOfDay
		{"D235959 vs D240000", D235959, D240000, true, false},
		{"D240000 vs D235959", D240000, D235959, false, true},
		{"D240000 vs D240000", D240000, D240000, false, false},
		{"D000000 vs D240000", D000000, D240000, true, false},
		{"D240000 vs D000000", D240000, D000000, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d1.Before(tt.d2); got != tt.wantB {
				t.Errorf("Before(%s, %s) got %t, want %t", tt.d1.String(), tt.d2.String(), got, tt.wantB)
			}
			if got := tt.d1.After(tt.d2); got != tt.wantA {
				t.Errorf("After(%s, %s) got %t, want %t", tt.d1.String(), tt.d2.String(), got, tt.wantA)
			}
		})
	}
}

func TestDaytime_Between(t *testing.T) {
	tests := []struct {
		name  string
		d     Daytime
		start Daytime
		end   Daytime
		want  bool
	}{
		// Normal intervals (start < end)
		{"Normal: Inside", D120000, D010000, D230000, true},
		{"Normal: At start", D010000, D010000, D230000, true},
		{"Normal: At end", D230000, D010000, D230000, true},
		{"Normal: Outside before", D000000, D010000, D230000, false},
		{"Normal: Outside after", D235959, D010000, D230000, false},
		{"Normal: Start=End (Inside)", D120000, D120000, D120000, true},
		{"Normal: Start=End (Outside)", D010000, D120000, D120000, false},
		{"Normal: Full day [00:00:00, 24:00:00] (Inside)", D120000, D000000, D240000, true},
		{"Normal: Full day [00:00:00, 24:00:00] (At End)", D240000, D000000, D240000, true},
		{"Normal: Full day [00:00:00, 24:00:00] (At Start)", D000000, D000000, D240000, true},

		// Midnight wraparound intervals (start > end)
		{"Wraparound: Inside near start (23:00-01:00)", D235959, D230000, D010000, true},
		{"Wraparound: Inside near end (23:00-01:00)", D000000, D230000, D010000, true},
		{"Wraparound: At start (23:00-01:00)", D230000, D230000, D010000, true},
		{"Wraparound: At end (23:00-01:00)", D010000, D230000, D010000, true},
		{"Wraparound: Outside (23:00-01:00)", D120000, D230000, D010000, false},
		{"Wraparound: Full day [24:00:00, 00:00:00] (At Start)", D240000, D240000, D000000, true},
		{"Wraparound: Full day [24:00:00, 00:00:00] (At End)", D000000, D240000, D000000, true},
		{"Wraparound: Full day [24:00:00, 00:00:00] (Inside)", D120000, D240000, D000000, false}, // Should be false as D120000 is neither 24 nor 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Between(tt.start, tt.end); got != tt.want {
				t.Errorf("Between(%s, %s, %s) got %t, want %t", tt.d.String(), tt.start.String(), tt.end.String(), got, tt.want)
			}
		})
	}
}

func TestDaytime_TimeComparisonWrappers(t *testing.T) {
	baseTime := time.Date(2024, time.March, 10, 15, 30, 0, 0, time.UTC) // 15:30:00
	dBase := Must(15, 30, 0)

	t.Run("EqualTime", func(t *testing.T) {
		if !dBase.EqualTime(baseTime) {
			t.Errorf("EqualTime should be true for 15:30:00")
		}
		if D120000.EqualTime(baseTime) {
			t.Errorf("EqualTime should be false for 12:00:00")
		}
	})

	t.Run("BeforeTime", func(t *testing.T) {
		// D120000 (12:00:00) is before baseTime (15:30:00)
		if !D120000.BeforeTime(baseTime) {
			t.Errorf("BeforeTime should be true for 12:00:00")
		}
		// dBase (15:30:00) is not before baseTime (15:30:00)
		if dBase.BeforeTime(baseTime) {
			t.Errorf("BeforeTime should be false for equal time")
		}
	})

	t.Run("AfterTime", func(t *testing.T) {
		// D230000 (23:00:00) is after baseTime (15:30:00)
		if !D230000.AfterTime(baseTime) {
			t.Errorf("AfterTime should be true for 23:00:00")
		}
		// dBase (15:30:00) is not after baseTime (15:30:00)
		if dBase.AfterTime(baseTime) {
			t.Errorf("AfterTime should be false for equal time")
		}
	})
}

func TestDaytime_AddAndSub(t *testing.T) {
	tests := []struct {
		name          string
		d             Daytime
		seconds       int
		wantDaytime   Daytime
		wantDaysCross int
	}{
		// --- Add (Forward) ---
		{"Add: Simple forward", D120000, 3600, Must(13, 0, 0), 0},
		{"Add: Exactly one day forward", D120000, secondsInDay, D120000, 1},
		{"Add: Two days forward", D010000, 2 * secondsInDay, D010000, 2},
		{"Add: Cross midnight forward (23:00 + 2h)", D230000, 7200, D010000, 1},
		{"Add: EndOfDay + 1s", D240000, 1, Must(0, 0, 1), 1},
		{"Add: StartOfDay + EndOfDay", D000000, secondsInDay, D240000, 0},

		// --- Sub (Backward / Negative Add) ---
		{"Sub: Simple backward", D120000, -3600, Must(11, 0, 0), 0},
		{"Sub: Exactly one day backward", D120000, -secondsInDay, D120000, -1},
		{"Sub: Two days backward", D010000, -2 * secondsInDay, D010000, -2},
		{"Sub: Cross midnight backward (01:00 - 2h)", D010000, -7200, D230000, -1},
		{"Sub: StartOfDay - 1s", D000000, -1, D235959, -1},
		{"Sub: EndOfDay - 1s", D240000, -1, D235959, 0},

		// --- Edge Cases ---
		{"Edge: Add 0 seconds", D120000, 0, D120000, 0},
		{"Edge: Result exactly EndOfDay (86399 + 1)", D235959, 1, D240000, 0},
	}

	for _, tt := range tests {
		t.Run("Add_"+tt.name, func(t *testing.T) {
			got, days := tt.d.Add(tt.seconds)
			if got != tt.wantDaytime {
				t.Errorf("Add: daytime %s + %d seconds got daytime %s, want %s", tt.d.String(), tt.seconds, got.String(), tt.wantDaytime.String())
			}
			if days != tt.wantDaysCross {
				t.Errorf("Add: daytime %s + %d seconds got %d days, want %d", tt.d.String(), tt.seconds, days, tt.wantDaysCross)
			}
		})

		// Test Sub (using negative seconds in the test case)
		if tt.seconds != 0 {
			t.Run("Sub_"+tt.name, func(t *testing.T) {
				// Reverse the logic for Sub
				got, days := tt.d.Sub(-tt.seconds)
				if got != tt.wantDaytime {
					t.Errorf("Sub: daytime %s - %d seconds got daytime %s, want %s", tt.d.String(), -tt.seconds, got.String(), tt.wantDaytime.String())
				}
				if days != tt.wantDaysCross {
					t.Errorf("Sub: daytime %s - %d seconds got %d days, want %d", tt.d.String(), -tt.seconds, days, tt.wantDaysCross)
				}
			})
		}
	}
}

func TestDaytime_AddSubDurationWrappers(t *testing.T) {
	d := D120000 // 12:00:00
	duration := 1 * time.Hour

	// AddDuration: Check forward movement and day count
	t.Run("AddDuration: 12:00 + 1h", func(t *testing.T) {
		got, days := d.AddDuration(duration)
		want := Must(13, 0, 0)
		if got != want || days != 0 {
			t.Errorf("AddDuration got (%s, %d), want (%s, 0)", got.String(), days, want.String())
		}
	})

	// SubDuration: Check backward movement and day count
	t.Run("SubDuration: 12:00 - 1h", func(t *testing.T) {
		got, days := d.SubDuration(duration)
		want := Must(11, 0, 0)
		if got != want || days != 0 {
			t.Errorf("SubDuration got (%s, %d), want (%s, 0)", got.String(), days, want.String())
		}
	})
}

func TestDaytime_Diff(t *testing.T) {
	tests := []struct {
		name     string
		d        Daytime
		other    Daytime
		wantSec  int
		wantDays int
	}{
		// Normal differences (d > other)
		{"Normal: 12:00 - 01:00", D120000, D010000, 11 * 3600, 0},
		{"Normal: Equal daytimes", D120000, D120000, 0, 0},

		// Negative differences (d < other)
		// 01:00 - 12:00 should be 1h before midnight (23 hours difference), -1 day
		{"Negative: 01:00 - 12:00", D010000, D120000, 13 * 3600, -1},

		// EndOfDay handling
		// 24:00 - 00:00 -> 86400 - 0 = 86400. Normalized to (0, 1 day)
		{"EndOfDay - StartOfDay", D240000, D000000, 0, 1},
		// 00:00 - 24:00 -> 0 - 86400 = -86400. Normalized to (0, -1 day)
		{"StartOfDay - EndOfDay", D000000, D240000, 0, -1},
		// 24:00 - 23:00 -> 3600 seconds, 0 days
		{"EndOfDay - 23:00", D240000, D230000, 3600, 0},
		// 23:00 - 24:00 -> -3600 seconds. Normalized to (86400-3600, -1 day)
		{"23:00 - EndOfDay", D230000, D240000, 82800, -1},

		// Boundary differences
		{"Just before day start: 00:00:01 - 00:00:00", Daytime(1), D000000, 1, 0},
		{"Just before day end: 24:00:00 - 23:59:59", D240000, D235959, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSec, gotDays := tt.d.Diff(tt.other)
			if gotSec != tt.wantSec || gotDays != tt.wantDays {
				t.Errorf("Diff(%s - %s) got (%d sec, %d days), want (%d sec, %d days)",
					tt.d.String(), tt.other.String(), gotSec, gotDays, tt.wantSec, tt.wantDays)
			}
		})
	}
}

func TestDaytime_Mul(t *testing.T) {
	d := D060000 // 21600 seconds

	tests := []struct {
		name          string
		factor        int
		wantDaytime   Daytime
		wantDaysCross int
	}{
		{"Multiply by 1 (identity)", 1, d, 0},
		{"Multiply by 0", 0, D000000, 0},
		{"Multiply by 2 (12:00)", 2, D120000, 0},
		{"Multiply by 4 (24:00 / EndOfDay)", 4, D240000, 0},   // 4 * 21600 = 86400. Consistent with StartOfDay + EndOfDay = 0 days crossed.
		{"Multiply by 5 (Wraps once)", 5, D060000, 1},         // 5 * 21600 = 108000 -> 1 day crossed
		{"Multiply by 10 (Wraps many times)", 10, D120000, 2}, // 10 * 21600 = 216000 -> 2 days crossed

		// Negative factors (should result in backward movement)
		{"Multiply by -1 (18:00)", -1, D180000, -1}, // -21600 -> 64800 (18:00:00), -1 day
		// FIX: Days crossed changed from -3 to -2
		{"Multiply by -5", -5, D180000, -2}, // -108000 -> 64800 (18:00:00), -2 days
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, days := d.Mul(tt.factor)
			if got != tt.wantDaytime || days != tt.wantDaysCross {
				t.Errorf("Mul(%s * %d) got (%s, %d days), want (%s, %d days)",
					d.String(), tt.factor, got.String(), days, tt.wantDaytime.String(), tt.wantDaysCross)
			}
		})
	}
}

func TestDaytime_Div(t *testing.T) {
	// Daytime being divided is now part of the test struct
	tests := []struct {
		name          string
		d             Daytime // Daytime to be divided
		divisor       int
		wantQuotient  Daytime
		wantRemainder int
		wantErr       error
	}{
		// Success cases run on D120000 (43200 seconds)
		{"D120000 / 1 (Identity)", D120000, 1, D120000, 0, nil},
		{"D120000 / 2 (Result 06:00:00)", D120000, 2, D060000, 0, nil},
		{"D120000 / 3 (Result 04:00:00)", D120000, 3, Must(4, 0, 0), 0, nil},
		{"D120000 / 100 (Result 00:07:12)", D120000, 100, Daytime(432), 0, nil}, // 43200 / 100 = 432 seconds (00:07:12)

		// Success case with non-zero remainder (run on D120000)
		{"D120000 / 7 (Remainder 3)", D120000, 7, Must(1, 42, 51), 3, nil}, // 43200 / 7 = 6171 R 3 (6171s = 01:42:51)

		// EndOfDay cases run on D240000 (86400 seconds)
		// 86400 / 2 = 43200 (12:00:00)
		{"EndOfDay / 2", D240000, 2, D120000, 0, nil},
		// 86400 / 3 = 28800 (08:00:00)
		{"EndOfDay / 3", D240000, 3, Daytime(28800), 0, nil},

		// Error cases (run on D120000)
		{"Error: Division by zero", D120000, 0, 0, 0, ErrDivisionByZero},
		{"Error: Quotient out of range (< 0)", D120000, -1, 0, 0, ErrValueOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, gotR, err := tt.d.Div(tt.divisor)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Div: got error %v, want error %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}

			if gotD != tt.wantQuotient {
				t.Errorf("Div: got daytime %s, want %s", gotD.String(), tt.wantQuotient.String())
			}
			if gotR != tt.wantRemainder {
				t.Errorf("Div: got remainder %d, want %d", gotR, tt.wantRemainder)
			}
		})
	}
}

func TestDaytime_Mod(t *testing.T) {
	d := Must(1, 0, 0) // 3600 seconds
	dEOD := D240000    // 86400 seconds

	tests := []struct {
		name        string
		d           Daytime
		modulus     int
		wantDaytime Daytime
		wantErr     error
	}{
		// Success cases
		{"Success: Modulus 3600 (0)", d, 3600, D000000, nil},
		{"Success: Modulus 3601 (Remainder 3600)", d, 3601, Daytime(3600), nil},
		{"Success: Modulus 60 (Remainder 0)", d, 60, D000000, nil},

		// EndOfDay case (86400)
		{"EndOfDay: Modulus 86400 (0)", dEOD, 86400, D000000, nil},
		{"EndOfDay: Modulus 1 (0)", dEOD, 1, D000000, nil},
		{"EndOfDay: Modulus 100", dEOD, 100, D000000, nil}, // 86400 is divisible by 100

		// Error cases
		{"Error: Modulus 0", d, 0, 0, ErrInvalidModulus},
		{"Error: Modulus -1", d, -1, 0, ErrInvalidModulus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, err := tt.d.Mod(tt.modulus)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Mod: got error %v, want error %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}

			if gotD != tt.wantDaytime {
				t.Errorf("Mod: got daytime %s, want %s", gotD.String(), tt.wantDaytime.String())
			}
		})
	}
}

func TestDaytime_Time(t *testing.T) {
	var fixedBaseTime = time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)

	berlinLoc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("Failed to load location Europe/Berlin: %v", err)
	}
	fixedBaseTimeBerlin := time.Date(2025, time.January, 10, 0, 0, 0, 0, berlinLoc)

	tests := []struct {
		name string
		d    Daytime
		base time.Time
		want time.Time
	}{
		{"UTC: StartOfDay", D000000, fixedBaseTime, fixedBaseTime},
		{"UTC: MidDay", D120000, fixedBaseTime, fixedBaseTime.Add(12 * time.Hour)},
		{"UTC: EndOfDayMinusOne", D235959, fixedBaseTime, fixedBaseTime.Add(secondsInDay*time.Second - time.Second)},
		{"UTC: EndOfDay (Next Day)", D240000, fixedBaseTime, fixedBaseTime.Add(24 * time.Hour)},
		{"Berlin: MidDay", D120000, fixedBaseTimeBerlin, fixedBaseTimeBerlin.Add(12 * time.Hour)},
		{"Berlin: EndOfDay (Next Day)", D240000, fixedBaseTimeBerlin, fixedBaseTimeBerlin.Add(24 * time.Hour)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.Time(tt.base)
			if !got.Equal(tt.want) {
				t.Errorf("Daytime(%s).Time(base=%s) got %s, want %s",
					tt.d.String(), tt.base.Format(time.TimeOnly), got.Format(time.DateTime), tt.want.Format(time.DateTime))
			}
			if got.Location() != tt.base.Location() {
				t.Errorf("Daytime(%s).Time() got location %v, want %v", tt.d.String(), got.Location(), tt.base.Location())
			}
		})
	}
}

func TestDaytime_Since(t *testing.T) {
	var fixedBaseTime = time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)

	d := D120000
	dTime := d.Time(fixedBaseTime) // 2025-01-10 12:00:00 UTC

	prevDayBase := fixedBaseTime.AddDate(0, 0, -1) // 2025-01-09 00:00:00 UTC

	tests := []struct {
		name      string
		otherTime time.Time
		baseTime  time.Time
		wantDur   time.Duration
	}{
		{"Same Day: 2 hours ago", fixedBaseTime.Add(10 * time.Hour), fixedBaseTime, 2 * time.Hour},
		{"Same Day: 2 hours ahead (Negative duration)", fixedBaseTime.Add(14 * time.Hour), fixedBaseTime, -2 * time.Hour},
		{"Same Day: Equal time", dTime, fixedBaseTime, 0},

		{"Cross Day: Exactly 24h", prevDayBase.Add(12 * time.Hour), fixedBaseTime, 24 * time.Hour},
		{"Cross Day: 22h ago", prevDayBase.Add(14 * time.Hour), fixedBaseTime, 22 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.Since(tt.otherTime, tt.baseTime)
			if got != tt.wantDur {
				t.Errorf("Daytime(%s).Since(%s) with base %s got %v, want %v",
					d.String(), tt.otherTime.Format(time.DateTime), tt.baseTime.Format(time.DateOnly), got, tt.wantDur)
			}
		})
	}
}

func TestDaytime_Until(t *testing.T) {
	var fixedBaseTime = time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)

	d := D120000
	dTime := d.Time(fixedBaseTime) // 2025-01-10 12:00:00 UTC

	nextDayBase := fixedBaseTime.AddDate(0, 0, 1) // 2025-01-11 00:00:00 UTC

	tests := []struct {
		name      string
		otherTime time.Time
		baseTime  time.Time
		wantDur   time.Duration
	}{
		{"Same Day: 2 hours ahead", fixedBaseTime.Add(14 * time.Hour), fixedBaseTime, 2 * time.Hour},
		{"Same Day: 2 hours ago (Negative duration)", fixedBaseTime.Add(10 * time.Hour), fixedBaseTime, -2 * time.Hour},
		{"Same Day: Equal time", dTime, fixedBaseTime, 0},

		{"Cross Day: Exactly 24h", nextDayBase.Add(12 * time.Hour), fixedBaseTime, 24 * time.Hour},
		{"Cross Day: 22h ahead", nextDayBase.Add(10 * time.Hour), fixedBaseTime, 22 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.Until(tt.otherTime, tt.baseTime)
			if got != tt.wantDur {
				t.Errorf("Daytime(%s).Until(%s) with base %s got %v, want %v",
					d.String(), tt.otherTime.Format(time.DateTime), tt.baseTime.Format(time.DateOnly), got, tt.wantDur)
			}
		})
	}
}

func TestDaytime_String(t *testing.T) {
	tests := []struct {
		name string
		d    Daytime
		want string
	}{
		{"Start of day (00:00:00)", D000000, "00:00:00"},
		{"One hour after midnight (01:00:00)", D010000, "01:00:00"},
		{"Mid-day daytime (12:30:45)", D123045, "12:30:45"},
		{"Just before end of day (23:59:59)", D235959, "23:59:59"},
		{"End of day (24:00:00)", D240000, "24:00:00"},
		{"Invalid daytime (> 86400)", DInvalid, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("Daytime.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDaytime_Format(t *testing.T) {
	var baseTime = time.Date(2025, time.October, 26, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		d      Daytime
		layout string
		want   string
	}{
		{"Hour:Minute:Second (12:30:45)", D123045, "15:04:05", "12:30:45"},
		{"Hour and AM/PM (12:30:45 PM)", D123045, "3:04 PM", "12:30 PM"},
		{"Hour and AM/PM (01:00:00 AM)", D010000, "3:04 PM", "1:00 AM"},
		{"Full date/time (StartOfDay)", D000000, time.RFC3339, "2025-10-26T00:00:00Z"},
		{"Full date/time (EndOfDay)", D240000, time.RFC3339, "2025-10-27T00:00:00Z"}, // Should roll over to the next day (Oct 27)
		{"Day of week (EndOfDay)", D240000, "Monday", "Monday"},                      // Oct 27, 2025 is a Monday
		// Note: Invalid daytimes are not tested here as they are handled inside d.Time() which returns a valid time.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Format(tt.layout, baseTime); got != tt.want {
				t.Errorf("Daytime.Format(%s) = %v, want %v", tt.layout, got, tt.want)
			}
		})
	}
}

func TestDaytime_MarshalText(t *testing.T) {
	tests := []struct {
		name string
		d    Daytime
		want string
		err  error
	}{
		{"Start of day", D000000, "00:00:00", nil},
		{"Mid-day daytime", D123045, "12:30:45", nil},
		{"End of day", D240000, "24:00:00", nil},
		{"Invalid daytime", DInvalid, "invalid", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, err := tt.d.MarshalText()

			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("MarshalText() got error %v, want error %v", err, tt.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("MarshalText() got unexpected error: %v", err)
			}

			got := string(gotBytes)
			if got != tt.want {
				t.Errorf("MarshalText() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("JSON Marshal", func(t *testing.T) {
		type data struct {
			D Daytime `json:"d"`
		}
		d := data{D: D123045}
		want := `{"d":"12:30:45"}`
		gotBytes, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("JSON Marshal failed: %v", err)
		}
		if got := string(gotBytes); got != want {
			t.Errorf("JSON Marshal got %q, want %q", got, want)
		}
	})
}

func TestDaytime_UnmarshalText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Daytime
		err   error
	}{
		{"Valid HH:MM:SS (10:30:15)", "10:30:15", Must(10, 30, 15), nil},
		{"Valid Seconds (3600)", "3600", D010000, nil},
		{"Start of day (0)", "0", D000000, nil},
		{"End of day (24:00:00)", "24:00:00", D240000, nil},
		{"End of day (86400 seconds)", "86400", D240000, nil},

		{"Invalid format (too short)", "1:2", D000000, ErrInvalidFormat},
		{"Invalid format (non-numeric)", "1a:2b:3c", D000000, ErrInvalidFormat},
		{"Invalid format (empty string)", "", D000000, ErrInvalidFormat},

		{"Seconds out of range (>86400)", "86401", D000000, ErrInvalidFormat},
		{"Seconds out of range (negative)", "-1", D000000, ErrInvalidFormat},

		{"End of day exceeded (24:00:01)", "24:00:01", D000000, ErrInvalidFormat},
		{"End of day exceeded (24:01:00)", "24:01:00", D000000, ErrInvalidFormat},
		{"Time component out of range (hour)", "25:00:00", D000000, ErrInvalidFormat},
		{"Time component out of range (minute)", "10:60:00", D000000, ErrInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Daytime
			err := d.UnmarshalText([]byte(tt.input))

			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("UnmarshalText(%q) got error %v, want error %v", tt.input, err, tt.err)
				}
				// We expect 'd' to be unchanged on error (zero value) or the error-handler's desired default
				if d != tt.want {
					t.Errorf("UnmarshalText(%q) got daytime %v on error, want %v", tt.input, d, tt.want)
				}
				return
			}

			if err != nil {
				t.Fatalf("UnmarshalText(%q) got unexpected error: %v", tt.input, err)
			}
			if d != tt.want {
				t.Errorf("UnmarshalText(%q) got daytime %v, want %v", tt.input, d, tt.want)
			}
		})
	}

	// Test with JSON to confirm standard library integration
	t.Run("JSON Unmarshal", func(t *testing.T) {
		type data struct {
			D Daytime `json:"d"`
		}
		inputJSON := `{"d":"12:30:45"}`
		var d data
		want := D123045
		err := json.Unmarshal([]byte(inputJSON), &d)
		if err != nil {
			t.Fatalf("JSON Unmarshal failed: %v", err)
		}
		if d.D != want {
			t.Errorf("JSON Unmarshal got daytime %v, want %v", d.D, want)
		}
	})
}
