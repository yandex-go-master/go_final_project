package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("empty repeat rule")
	}

	startDate, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}

	repeatRule := strings.Split(repeat, " ")

	switch repeatRule[0] {
	case "d":
		if len(repeatRule) != 2 {
			return "", fmt.Errorf("invalid day repeat rule format")
		}
		days, err := strconv.Atoi(repeatRule[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("invalid day interval")
		}
		nextDate := startDate
		for {
			nextDate = nextDate.AddDate(0, 0, days)
			if nextDate.After(now) {
				return nextDate.Format(DateFormat), nil
			}
		}
	case "y":
		nextDate := startDate.AddDate(1, 0, 0)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format(DateFormat), nil
	case "w":
		if len(repeatRule) != 2 {
			return "", fmt.Errorf("invalid week repeat rule format")
		}
		days := strings.Split(repeatRule[1], ",")
		var repeatDays []int
		for _, day := range days {
			dayNumber, err := strconv.Atoi(day)
			if err != nil || dayNumber < 1 || dayNumber > 7 {
				return "", fmt.Errorf("invalid week interval")
			}
			repeatDays = append(repeatDays, dayNumber)
		}
		nextDate := startDate
		for {
			nextDate = nextDate.AddDate(0, 0, 1)
			if nextDate.After(now) {
				for _, day := range repeatDays {
					if int(nextDate.Weekday()) == (day % 7) {
						return nextDate.Format(DateFormat), nil
					}
				}
			}
		}
	case "m":
		if len(repeatRule) < 2 || len(repeatRule) > 3 {
			return "", fmt.Errorf("invalid month repeat rule format")
		}

		var repeatDays []int
		days := strings.Split(repeatRule[1], ",")
		for _, day := range days {
			dayNumber, err := strconv.Atoi(day)
			if err != nil || dayNumber < -2 || dayNumber > 31 || dayNumber == 0 {
				return "", fmt.Errorf("invalid month interval: %w", err)
			}
			repeatDays = append(repeatDays, dayNumber)
		}

		var repeatMonths []int
		if len(repeatRule) == 3 {
			months := strings.Split(repeatRule[2], ",")
			for _, month := range months {
				monthNumber, err := strconv.Atoi(month)
				if err != nil || monthNumber < 1 || monthNumber > 12 {
					return "", fmt.Errorf("invalid month interval: %w", err)
				}
				repeatMonths = append(repeatMonths, monthNumber)
			}
		}

		nextDate := startDate
		for {
			nextDate = nextDate.AddDate(0, 0, 1)
			if nextDate.After(now) {
				for _, day := range repeatDays {
					if day == -1 && nextDate.Day() == lastDayOfMonth(nextDate) {
						if len(repeatMonths) == 0 || contains(repeatMonths, int(nextDate.Month())) {
							return nextDate.Format(DateFormat), nil
						}
					} else if day == -2 && nextDate.Day() == lastDayOfMonth(nextDate)-1 {
						if len(repeatMonths) == 0 || contains(repeatMonths, int(nextDate.Month())) {
							return nextDate.Format(DateFormat), nil
						}
					} else if day > 0 && nextDate.Day() == day {
						if len(repeatMonths) == 0 || contains(repeatMonths, int(nextDate.Month())) {
							return nextDate.Format(DateFormat), nil
						}
					}
				}
			}
		}
	default:
		return "", fmt.Errorf("invalid repeat rule format")
	}
}

func lastDayOfMonth(t time.Time) int {
	nextMonth := t.Month() + 1
	if nextMonth > 12 {
		nextMonth = 1
	}
	nextYear := t.Year()
	if nextMonth == 1 {
		nextYear++
	}
	firstOfNextMonth := time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, t.Location())
	return firstOfNextMonth.Add(-24 * time.Hour).Day()
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
