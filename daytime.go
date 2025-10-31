package daytime

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Daytime represents a time moment within a day [0, 86400].
// Zero value corresponds to the start of day (00:00:00).
// Value 86400 represents the end of day (24:00:00).
type Daytime uint32

const (
	secondsInDay = 86400
	hoursInDay   = 24

	// StartOfDay represents the start of day (00:00:00)
	StartOfDay = Daytime(0)

	// EndOfDay represents the end of day (24:00:00)
	EndOfDay = Daytime(secondsInDay)
)

// Error represents a daytime package error.
type Error struct {
	op    string
	value any
	err   error
}

func (e *Error) Error() string {
	if e.value != nil {
		return fmt.Sprintf("daytime: %s: %v: %v", e.op, e.value, e.err)
	}
	return fmt.Sprintf("daytime: %s: %v", e.op, e.err)
}

func (e *Error) Unwrap() error {
	return e.err
}

// Operation returns the operation that caused the error.
func (e *Error) Operation() string {
	return e.op
}

// Value returns the problematic value associated with the error.
func (e *Error) Value() any {
	return e.value
}

var (
	// ErrDivisionByZero indicates division by zero.
	ErrDivisionByZero = errors.New("division by zero")

	// ErrInvalidTimeComponent indicates invalid time components (hour, minute, second).
	ErrInvalidTimeComponent = errors.New("invalid time component")

	// ErrInvalidFormat indicates the input string has invalid format.
	ErrInvalidFormat = errors.New("invalid format")

	// ErrValueOutOfRange indicates a value is outside valid range.
	ErrValueOutOfRange = errors.New("value out of range")

	// ErrInvalidModulus indicates an invalid modulus value.
	ErrInvalidModulus = errors.New("modulus must be positive")

	// ErrEndOfDayExceeded indicates that 24:00:00 was specified with non-zero minutes or seconds.
	// This replaces the previous unexported error string for better errors.Is support.
	ErrEndOfDayExceeded = errors.New("daytime 24:00:00 must have zero minutes and seconds")
)

// errorf creates a new wrapped error with operation context.
func errorf(op string, value any, err error) error {
	if err == nil {
		return nil
	}
	var existingErr *Error
	if errors.As(err, &existingErr) {
		// Do not double-wrap an existing daytime.Error
		return err
	}
	return &Error{op: op, value: value, err: err}
}

// Valid checks if the daytime represents a valid time value [StartOfDay, EndOfDay].
func (d Daytime) Valid() bool {
	return d <= EndOfDay
}

// IsEndOfDay checks if the daytime represents the end of day (24:00:00).
func (d Daytime) IsEndOfDay() bool {
	return d == EndOfDay
}

// IsInDay checks if the daytime is strictly within the day [StartOfDay, EndOfDay).
func (d Daytime) IsInDay() bool {
	return d < EndOfDay
}

// --- Creation ---

// New creates a new daytime from hours, minutes, and seconds.
//
// Valid ranges:
//
//   - hour: [0, 24]
//   - minute: [0, 59]
//   - second: [0, 59]
//
// Returns an error if any component is out of range or if 24:00:00
// is specified with non-zero minutes or seconds.
func New(hour, minute, second int) (Daytime, error) {
	if hour < 0 || hour > hoursInDay ||
		minute < 0 || minute > 59 ||
		second < 0 || second > 59 {
		return 0, errorf("New", fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), ErrInvalidTimeComponent)
	}

	if hour == hoursInDay && (minute != 0 || second != 0) {
		return 0, errorf("New", fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), ErrEndOfDayExceeded)
	}

	total := hour*3600 + minute*60 + second
	if total > secondsInDay {
		// This should only happen if hour=24 and minute=0, second=0, which is EndOfDay (86400)
		return EndOfDay, nil
	}
	return Daytime(total), nil
}

// Must creates a new daytime from hours, minutes, and seconds, panicking on error.
//
// This is intended for use with compile-time constants where invalid values
// should cause immediate failure.
func Must(hour, minute, second int) Daytime {
	d, err := New(hour, minute, second)
	if err != nil {
		panic(err)
	}
	return d
}

// FromTime creates a daytime from time.Time, extracting the time-of-day portion.
func FromTime(t time.Time) Daytime {
	return fromTime(t)
}

// Parse parses a daytime from string.
//
// Supported input formats:
//
//   - <seconds>: integer seconds since midnight (e.g., "3600")
//   - "HH:MM:SS": hours:minutes:seconds (e.g., "01:00:00")
//
// Returns error if the string cannot be parsed or the value exceeds 86400 seconds.
func Parse(s string) (Daytime, error) {
	if s == "" {
		return 0, errorf("Parse", s, ErrInvalidFormat)
	}

	// Try parsing as integer seconds
	if sec, err := parseSeconds(s); err == nil {
		return Daytime(sec), nil
	}

	// Try parsing as HH:MM:SS time string
	if sec, err := parseTimeString(s); err == nil {
		return Daytime(sec), nil
	}

	return 0, errorf("Parse", s, ErrInvalidFormat)
}

// --- Time Components ---

// Clock returns the hour, minute, and second components of the daytime.
//
// For EndOfDay (24:00:00), returns (24, 0, 0).
func (d Daytime) Clock() (hour, minute, second int) {
	if d == EndOfDay {
		return hoursInDay, 0, 0
	}

	sec := int(d)
	hour = sec / 3600
	minute = (sec % 3600) / 60
	second = sec % 60
	return
}

// Hour returns the hour component [0, 24].
func (d Daytime) Hour() int {
	hour, _, _ := d.Clock()
	return hour
}

// Minute returns the minute component [0, 59].
func (d Daytime) Minute() int {
	_, minute, _ := d.Clock()
	return minute
}

// Second returns the second component [0, 59].
func (d Daytime) Second() int {
	_, _, second := d.Clock()
	return second
}

// Duration returns the daytime as time.Duration since midnight.
func (d Daytime) Duration() time.Duration {
	return time.Duration(d) * time.Second
}

// --- Comparison Operations ---

// Compare compares two daytimes.
//
// Returns:
//
//	-1 if d < other
//	 0 if d == other
//	+1 if d > other
func (d Daytime) Compare(other Daytime) int {
	if d == other {
		return 0
	}
	if d.Before(other) {
		return -1
	}
	return 1
}

// Before reports whether the daytime occurs before the other.
//
// EndOfDay is considered after all other daytimes except itself.
func (d Daytime) Before(other Daytime) bool {
	// If d is EndOfDay, it is only before EndOfDay if they are not equal, which is impossible.
	if d == EndOfDay {
		return false
	}
	// Any daytime before EndOfDay is before EndOfDay.
	if other == EndOfDay {
		return true
	}
	return d < other
}

// After reports whether the daytime occurs after the other.
func (d Daytime) After(other Daytime) bool {
	return other.Before(d)
}

// Equal reports whether two daytimes represent the same time.
func (d Daytime) Equal(other Daytime) bool {
	return d == other
}

// Between reports whether the daytime is between start and end (start; end).
//
// Handles midnight wraparound: if start > end, the period spans across midnight.
func (d Daytime) Between(start, end Daytime) bool {
	if !d.Valid() || !start.Valid() || !end.Valid() {
		return false
	}

	// Equal case - quick return
	if start == end {
		return d == start
	}

	if start.Before(end) {
		// Normal case: start <= d <= end
		return !d.Before(start) && !d.After(end)
	}

	// Midnight wraparound case: start > end
	// This means the period goes from start through EndOfDay, then from StartOfDay to end
	// So d is in the interval if: d >= start OR d <= end
	return !d.Before(start) || !d.After(end)
}

// BeforeTime reports whether the daytime occurs before the daytime extracted from the given time.
func (d Daytime) BeforeTime(t time.Time) bool { return d.Before(fromTime(t)) }

// AfterTime reports whether the daytime occurs after the daytime extracted from the given time.
func (d Daytime) AfterTime(t time.Time) bool { return d.After(fromTime(t)) }

// EqualTime reports whether the daytime equals the daytime extracted from the given time.
func (d Daytime) EqualTime(t time.Time) bool { return d.Equal(fromTime(t)) }

// --- Arithmetic Operations ---

// Add adds seconds to the daytime.
//
// Returns the resulting daytime (normalized to [0, 86400]) and the number of day boundaries crossed.
// Positive values move forward in time, negative values move backward.
func (d Daytime) Add(seconds int) (Daytime, int) {
	if !d.Valid() {
		return d, 0
	}

	// Use the underlying value of d, which is in [0, 86400].
	baseValue := int(d)
	total := baseValue + seconds

	days := total / secondsInDay
	remainder := total % secondsInDay

	// Adjust negative remainder
	if remainder < 0 {
		remainder += secondsInDay
		days--
	}

	// If total is exactly secondsInDay (e.g. 0 + 86400), it should be EndOfDay (24:00:00), 0 days.
	if total == secondsInDay {
		return EndOfDay, 0
	}

	// Normalize special value 0 to StartOfDay
	if remainder == 0 {
		return StartOfDay, days
	}

	return Daytime(remainder), days
}

// Sub subtracts seconds from the daytime.
//
// Returns the resulting daytime and the number of day boundaries crossed.
// This is equivalent to Add with negative seconds.
func (d Daytime) Sub(seconds int) (Daytime, int) {
	return d.Add(-seconds)
}

// AddDuration adds a time duration to the daytime.
//
// Returns the resulting daytime and the number of day boundaries crossed.
func (d Daytime) AddDuration(dur time.Duration) (Daytime, int) {
	return d.Add(int(dur.Seconds()))
}

// SubDuration subtracts a time duration from the daytime.
//
// Returns the resulting daytime and the number of day boundaries crossed.
func (d Daytime) SubDuration(dur time.Duration) (Daytime, int) {
	return d.Add(-int(dur.Seconds()))
}

// Diff calculates the difference between two daytimes (d - other).
//
// Returns the difference in seconds and the number of day boundaries crossed.
// The result is normalized so that seconds is always in [0, 86399].
func (d Daytime) Diff(other Daytime) (seconds int, days int) {
	if !d.Valid() || !other.Valid() {
		return 0, 0
	}

	baseSeconds := int(d)
	otherSeconds := int(other)

	// Treat EndOfDay (86400) as the full 24 hours for differential calculation
	if d == EndOfDay && d != other {
		baseSeconds = secondsInDay
	}
	if other == EndOfDay && d != other {
		otherSeconds = secondsInDay
	}

	diff := baseSeconds - otherSeconds
	days = diff / secondsInDay
	seconds = diff % secondsInDay

	// Adjust for negative difference
	if seconds < 0 {
		seconds += secondsInDay
		days--
	}

	return seconds, days
}

// Mul multiplies the daytime by a factor.
//
// Returns the resulting daytime and the number of day boundaries crossed.
// Useful for scaling time intervals.
func (d Daytime) Mul(factor int) (Daytime, int) {
	if !d.Valid() {
		return d, 0
	}

	// Use the underlying value of d, which is in [0, 86400].
	baseValue := int(d)
	total := baseValue * factor

	days := total / secondsInDay
	remainder := total % secondsInDay

	if remainder < 0 {
		remainder += secondsInDay
		days--
	}

	// If total is exactly secondsInDay, it should be EndOfDay (24:00:00), 0 days.
	if total == secondsInDay {
		return EndOfDay, 0
	}

	// Normalize special value 0 to StartOfDay
	if remainder == 0 {
		return StartOfDay, days
	}

	return Daytime(remainder), days
}

// Div divides the daytime by a divisor.
//
// Returns the quotient daytime, remainder seconds, and any error.
// Returns ErrDivisionByZero if divisor is zero.
func (d Daytime) Div(divisor int) (Daytime, int, error) {
	if divisor == 0 {
		return 0, 0, errorf("Div", divisor, ErrDivisionByZero)
	}
	if !d.Valid() {
		return d, 0, nil
	}

	var value int
	if d == EndOfDay {
		value = secondsInDay
	} else {
		value = int(d)
	}

	quotient := value / divisor
	remainder := value % divisor

	// The quotient must be a valid Daytime. If it exceeds 86400, it's an error.
	if quotient < 0 || quotient > secondsInDay {
		return 0, 0, errorf("Div", quotient, ErrValueOutOfRange) // Возвращаем ошибку диапазона
	}
	return Daytime(quotient), remainder, nil
}

// Mod computes the daytime modulo a modulus.
//
// Returns the remainder after division by modulus.
// The result is always in [0, modulus-1].
// Returns error if modulus is not positive.
func (d Daytime) Mod(modulus int) (Daytime, error) {
	if modulus <= 0 {
		return 0, errorf("Mod", modulus, ErrInvalidModulus)
	}
	if !d.Valid() {
		return d, nil
	}

	var value int
	if d == EndOfDay {
		value = secondsInDay
	} else {
		value = int(d)
	}

	result := value % modulus
	if result < 0 {
		result += modulus
	}

	// Modulo result is always less than modulus, and modulus is positive, so no range check needed.
	return Daytime(result), nil
}

// --- Conversions ---

// Time creates a time.Time by combining the daytime with a base date.
//
// For EndOfDay (24:00:00), returns the start of the next day.
// The time zone is taken from the base time.
func (d Daytime) Time(base time.Time) time.Time {
	year, month, day := base.Date()
	location := base.Location()

	if d == EndOfDay {
		// End of day (24:00:00) is equivalent to 00:00:00 of the next day.
		return time.Date(year, month, day, 0, 0, 0, 0, location).
			Add(24 * time.Hour)
	}

	hour, minute, second := d.Clock()
	return time.Date(year, month, day, hour, minute, second, 0, location)
}

// Since computes the duration from the given time to this daytime on the base date.
//
// A positive result means the daytime occurs after the given time.
func (d Daytime) Since(t time.Time, base time.Time) time.Duration {
	daytimeTime := d.Time(base)
	return daytimeTime.Sub(t)
}

// Until computes the duration from this daytime on the base date to the given time.
//
// A positive result means the daytime occurs before the given time.
func (d Daytime) Until(t time.Time, base time.Time) time.Duration {
	daytimeTime := d.Time(base)
	return t.Sub(daytimeTime)
}

// --- String Representations ---

// String returns the string representation in HH:MM:SS format.
//
// Returns "invalid" for invalid daytimes.
// EndOfDay returns "24:00:00".
func (d Daytime) String() string {
	if !d.Valid() {
		return "invalid"
	}

	if d == EndOfDay {
		return "24:00:00"
	}

	hour, minute, second := d.Clock()
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

// Format formats the daytime according to the layout string applied to the base date.
//
// The layout follows time.Time.Format conventions.
func (d Daytime) Format(layout string, base time.Time) string {
	return d.Time(base).Format(layout)
}

// --- Helper functions ---

// parseSeconds parses a string as integer seconds since midnight.
func parseSeconds(s string) (int, error) {
	var sec int
	sec, err := strconv.Atoi(s)
	if err != nil {
		return 0, errorf("parseSeconds", s, ErrInvalidFormat)
	}
	if sec < 0 || sec > secondsInDay {
		return 0, errorf("parseSeconds", sec, ErrValueOutOfRange)
	}
	return sec, nil
}

// parseTimeString parses a string in HH:MM:SS format.
func parseTimeString(s string) (int, error) {
	if len(s) != 8 {
		return 0, errorf("parseTimeString", s, ErrInvalidFormat)
	}
	var hour, minute, second int
	n, err := fmt.Sscanf(s, "%d:%d:%d", &hour, &minute, &second)
	if err != nil || n != 3 {
		return 0, errorf("parseTimeString", s, ErrInvalidFormat)
	}

	if hour < 0 || hour > hoursInDay ||
		minute < 0 || minute > 59 ||
		second < 0 || second > 59 {
		return 0, errorf("parseTimeString", fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), ErrInvalidTimeComponent)
	}

	if hour == hoursInDay && (minute != 0 || second != 0) {
		return 0, errorf("parseTimeString", fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), ErrEndOfDayExceeded)
	}

	total := hour*3600 + minute*60 + second
	if total > secondsInDay {
		// Should be unreachable if the above checks are correct, but safe to guard
		return 0, errorf("parseTimeString", fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), ErrValueOutOfRange)
	}

	return total, nil
}

// fromTime extracts the daytime from a time.Time.
func fromTime(t time.Time) Daytime {
	// time.Time.Clock() components are always valid and normalized to [0, 23]:[0, 59]:[0, 59]
	hour, min, sec := t.Clock()
	total := uint32(hour)*3600 + uint32(min)*60 + uint32(sec)
	return Daytime(total)
}
