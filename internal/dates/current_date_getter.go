package dates

import "time"

type CurrentDateGetter struct {
}

func NewCurrentDateGetter() *CurrentDateGetter {
	return &CurrentDateGetter{}
}

func (getter *CurrentDateGetter) GetCurrentDate() time.Time {
	return time.Now().UTC()
}
