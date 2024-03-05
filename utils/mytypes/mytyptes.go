package mytypes

import "time"

type PgTime struct {
	time.Time
}

func (pgt *PgTime) UnMarshalJSON(data []byte) error {
	parsedTime, err := time.Parse(time.DateTime, string(data))

	if err != nil {
		return err
	}

	pgt.Time = parsedTime
	return nil
}
