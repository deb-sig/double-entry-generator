package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func SplitFindContains(str, target, sep string, match bool) bool {
	ss := strings.Split(str, sep)
	isContain := false
	for _, s := range ss {
		if strings.Contains(target, s) {
			isContain = true
			break
		}
	}

	return isContain && match
}

func SplitFindEquals(str, target, sep string, match bool) bool {
	ss := strings.Split(str, sep)
	isEqual := false
	for _, s := range ss {
		if target == s {
			isEqual = true
			break
		}
	}

	return isEqual && match
}

func SplitFindTimeInterval(timeStr string, targetTime time.Time, match bool) (bool, error) {
	isContain := false

	timeStrs := strings.Split(timeStr, "-")
	if len(timeStrs) != 2 {
		return match, fmt.Errorf("fail to parse the time condition `%s`", timeStr)
	}

	startTimeStrs := strings.Split(timeStrs[0], ":")
	// time can be "00:00" or "00:00:00"
	if len(startTimeStrs) < 2 {
		return match, fmt.Errorf("fail to parse the start time `%s`", timeStrs[0])
	}

	endTimeStrs := strings.Split(timeStrs[1], ":")
	if len(endTimeStrs) < 2 {
		return match, fmt.Errorf("fail to parse the start time `%s`", timeStrs[1])
	}

	startTimeHour, err := strconv.Atoi(startTimeStrs[0])
	if err != nil {
		return match, fmt.Errorf("fail to parse the start time hour `%s`", startTimeStrs[0])
	}
	startTimeMinute, err := strconv.Atoi(startTimeStrs[1])
	if err != nil {
		return match, fmt.Errorf("fail to parse the start time minute `%s`", startTimeStrs[1])
	}
	startTimeSecond := 0
	if len(startTimeStrs) > 2 {
		startTimeSecond, err = strconv.Atoi(startTimeStrs[2])
		if err != nil {
			return match, fmt.Errorf("fail to parse the start time second `%s`", startTimeStrs[2])
		}
	}
	if startTimeHour < 0 || startTimeHour >= 24 ||
		startTimeMinute < 0 || startTimeMinute >= 60 ||
		startTimeSecond < 0 || startTimeSecond >= 60 {
		return match, fmt.Errorf("fail to parse the start time, hour:[0,23], minute:[0,59], second:[0,59]")
	}

	endTimeHour, err := strconv.Atoi(endTimeStrs[0])
	if err != nil {
		return match, fmt.Errorf("fail to parse the end time hour `%s`", endTimeStrs[0])
	}
	endTimeMinute, err := strconv.Atoi(endTimeStrs[1])
	if err != nil {
		return match, fmt.Errorf("fail to parse the end time minute `%s`", endTimeStrs[1])
	}
	endTimeSecond := 0
	if len(endTimeStrs) > 2 {
		endTimeSecond, err = strconv.Atoi(endTimeStrs[2])
		if err != nil {
			return match, fmt.Errorf("fail to parse the end time second `%s`", endTimeStrs[2])
		}
	}
	if endTimeHour < 0 || endTimeHour >= 24 ||
		endTimeMinute < 0 || endTimeMinute >= 60 ||
		endTimeSecond < 0 || endTimeSecond >= 60 {
		return match, fmt.Errorf("fail to parse the end time, hour:[0,23], minute:[0,59], second:[0,59]")
	}

	var startTime, endTime time.Time
	if targetTime.Hour() >= startTimeHour {
		startTime = time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), startTimeHour, startTimeMinute, startTimeSecond, 0, targetTime.Location())
		endTime = time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), endTimeHour, endTimeMinute, endTimeSecond, 0, targetTime.Location())
		if startTimeHour > endTimeHour {
			// 23:00T to 01:00T+1
			endTime = endTime.AddDate(0, 0, 1)
		}
	} else { // target time hour < start time hour, maybe in T-1 cycle
		startTime = time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), startTimeHour, startTimeMinute, startTimeSecond, 0, targetTime.Location())
		endTime = time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), endTimeHour, endTimeMinute, endTimeSecond, 0, targetTime.Location())
		if startTimeHour > endTimeHour {
			// 23:00T-1 to 01:00T
			startTime = startTime.AddDate(0, 0, -1)
		}
	}

	if targetTime.After(startTime) && targetTime.Before(endTime) {
		isContain = true
	}

	return isContain && match, nil
}

func EscapeString(str string) string {
	str1 := strings.ReplaceAll(str, "\\", "\\'")
	return strings.ReplaceAll(str1, `"`, `\"`)
}

func SplitFindTimeStampInterval(timeRangeStr string, targetTime time.Time, match bool) (bool, error) {
	isContain := false

	timeRangeStrs := strings.Split(timeRangeStr, "-")
	if len(timeRangeStrs) != 2 {
		return match, fmt.Errorf("fail to parse the time condition `%s`", timeRangeStr)
	}

	startTimeInt, err := strconv.ParseInt(timeRangeStrs[0], 10, 64)
	if err != nil {
		return match, fmt.Errorf("fail to parse the start timestamp `%s`", timeRangeStrs[0])
	}
	startTime := time.Unix(startTimeInt, 0)

	endTimeInt, err := strconv.ParseInt(timeRangeStrs[1], 10, 64)
	if err != nil {
		return match, fmt.Errorf("fail to parse the end timestamp `%s`", timeRangeStrs[1])
	}
	endTime := time.Unix(endTimeInt, 0)

	if targetTime.After(startTime) && targetTime.Before(endTime) {
		isContain = true
	}
	if match {
		println("start time", startTime.String(), endTime.String(), targetTime.String())
	}

	return isContain && match, nil
}
