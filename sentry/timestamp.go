package sentry

import "time"

const timestampFormat = `"2006-01-02T15:04:05.00"`

// Timestamp is JSON wrapper for time.Time
type Timestamp time.Time

// MarshalJSON converts timestamp into JSON representation
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(t).UTC().Format(timestampFormat)), nil
}

// UnmarshalJSON converts JSON to Timestamp
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	ts, err := time.Parse(timestampFormat, string(data))
	if err != nil {
		return err
	}

	*t = Timestamp(ts)
	return nil
}
