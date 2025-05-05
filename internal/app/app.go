package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"impulse_yadro/internal/storage"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse(storage.TimeFmt, timeStr)
}

func formatTime(t time.Time) string {
	return t.Format(storage.TimeFmt)
}

func ReadConfig(filename string) (storage.Config, error) {
	var config storage.Config
	file, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}

func ReadEvents(filename string) ([]storage.Event, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []storage.Event
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)

		if len(parts) < 2 {
			continue
		}

		eventTime, err := ConvertToTime(parts[0])

		if err != nil {
			return nil, err
		}

		eventID, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		competitorID, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, err
		}
		extraParams := ""
		if len(parts) > 3 {
			for i := 3; i < len(parts); i++ {
				extraParams += parts[i] + " "
			}
		}

		events = append(events, storage.Event{
			Time:         eventTime,
			EventID:      eventID,
			CompetitorID: competitorID,
			ExtraParams:  extraParams,
		})
	}

	return events, scanner.Err()
}

func newCompetitor(id int, config *storage.Config) *storage.Competitor {
	return &storage.Competitor{
		ID:   id,
		Laps: make([]storage.Lap, 0, config.Laps),
	}
}

func ConvertToTime(time string) (time.Time, error) {
	timeStr := strings.TrimPrefix(time, "[")
	timeStr = timeStr[:len(timeStr)-1]
	eventTime, err := parseTime(timeStr)
	if err != nil {
		return eventTime, err
	}
	return eventTime, nil
}

func ParseStartDelta(deltaStr string) (time.Duration, error) {
	parts := strings.Split(deltaStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format")
	}

	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	s, _ := strconv.Atoi(parts[2])

	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(s)*time.Second, nil
}

func ResultingTable(competitors map[int]*storage.Competitor, config *storage.Config) {
	results := make([]*storage.Result, 0)

	// Сначала собираем всех зарегистрированных участников
	for _, comp := range competitors {
		if !comp.Registered {
			continue
		}

		res := &storage.Result{
			ID_competitor: comp.ID,
			Laps:          make([]storage.LapsInfo, 0),
			Penalty:       make([]storage.PenaltyInfo, 0),
		}

		// Определяем статус
		switch {
		case comp.Disqualified:
			res.Finished = "Disqualified"
		case comp.NotFinished:
			res.Finished = "NotFinished"
		default:
			res.Finished = formatDuration(comp.Finished)
		}

		// Заполняем данные по кругам и по пенальти
		// Считаем попадания
		totalShots := config.FiringLines * 5
		hits := 0
		for _, lapTime := range comp.Laps {
			if lapTime.LapDuration.Seconds() > 0 {
				avgSpeed := float64(config.Laplen) / lapTime.LapDuration.Seconds()
				res.Laps = append(res.Laps, storage.LapsInfo{
					TimeLaps:     lapTime.LapDuration,
					AvgSpeedLaps: avgSpeed,
				})
			}
			for _, firing := range lapTime.FiringRanges {
				hits += len(firing.Shots)
				if len(firing.Shots) != 5 {
					avgSpeed := float64(config.PenaltyLen*(5-len(firing.Shots))) / firing.Duration.Seconds()
					res.Penalty = append(res.Penalty, storage.PenaltyInfo{
						TimePenalty:     firing.Duration,
						AvgSpeedPenalty: avgSpeed,
					})
				}

			}
		}

		res.Shots = fmt.Sprintf("%d/%d", hits, totalShots)

		results = append(results, res)
	}

	// Сортируем по времени финиша
	sort.Slice(results, func(i, j int) bool {
		if len(results[i].Laps) == 0 || len(results[j].Laps) == 0 {
			return false
		}
		return results[i].Laps[0].TimeLaps < results[j].Laps[0].TimeLaps
	})

	for _, res := range results {
		fmt.Printf("[%s] %d ", res.Finished, res.ID_competitor)
		// Круги
		var laps []string
		for _, lap := range res.Laps {
			laps = append(laps, fmt.Sprintf("{%s, %.3f}",
				formatDuration(lap.TimeLaps),
				lap.AvgSpeedLaps))
		}
		fmt.Print(laps, " ")

		// Штрафы
		var pens []string
		for _, pen := range res.Penalty {
			pens = append(pens, fmt.Sprintf("{%s, %.3f}",
				formatDuration(pen.TimePenalty),
				pen.AvgSpeedPenalty))
		}
		fmt.Print(pens, " ")

		// Попадания
		fmt.Println(res.Shots)
	}
}

// func avgSpeed(duration time.Duration, laps int) float64 {
// 	if duration.Seconds() != 0 {
// 		return float64(laps) / duration.Seconds()
// 	}
// 	return 0
// }

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	ms := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, ms)
}

func ProcessingEvents(competitors map[int]*storage.Competitor, events []storage.Event, config *storage.Config) {

	for _, elem := range events {
		if _, ok := competitors[elem.CompetitorID]; !ok {
			competitors[elem.CompetitorID] = newCompetitor(elem.CompetitorID, config)
		}
		comp := competitors[elem.CompetitorID]
		switch elem.EventID {
		case 1:
			comp.Registered = true
			fmt.Printf("[%s] The competitor(%d) registered\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID)
		case 2:
			eventTime, err := ConvertToTime("[" + strings.TrimSpace(elem.ExtraParams) + "]")
			if err != nil {
				log.Printf("uncorrected time format %d ", comp.ID)
				continue
			}
			comp.ScheduledStart = eventTime
			fmtTime := formatTime(eventTime)
			fmt.Printf("[%s] The start time for the competitor(%d) was set by a draw to %s\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID, fmtTime)
		case 3:
			comp.WasOnStartLine = true
			comp.OnStartLine = elem.Time
			fmt.Printf("[%s] The competitor(%d) is on the start line\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID)
		case 4:
			comp.ActualStart = elem.Time
			comp.Laps = append(comp.Laps, storage.Lap{FiringRanges: make([]storage.FiringRange, 0), LapStart: elem.Time})
			comp.Laps[comp.CurrentLap].LapStart = elem.Time
			startDelay := comp.ActualStart.Sub(comp.ScheduledStart)
			deltaTime, err := ParseStartDelta(config.StartDelta)
			if err != nil {
				log.Printf("uncorrected time format DeltaTime in config")
			}
			if startDelay > deltaTime {
				comp.Disqualified = true

				fmt.Printf("[%s] 32 %d\n",
					comp.ActualStart.Format(storage.TimeFmt),
					comp.ID)
			}
			fmt.Printf("[%s] The competitor(%d) has started\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID)
		case 5:
			comp.Laps[comp.CurrentLap].FiringRanges = append(comp.Laps[comp.CurrentLap].FiringRanges, storage.FiringRange{Shots: make([]bool, 0)})
			// comp.Laps[comp.CurrentLap].CurrentFiringLine++
			fmt.Printf("[%s] The competitor(%d) is on the firing range(%d)\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID, comp.Laps[comp.CurrentLap].CurrentFiringLine+1)

		case 6:
			comp.Laps[comp.CurrentLap].FiringRanges[comp.Laps[comp.CurrentLap].CurrentFiringLine].Shots = append(comp.Laps[comp.CurrentLap].FiringRanges[comp.Laps[comp.CurrentLap].CurrentFiringLine].Shots, true)
			fmt.Printf("[%s] The target(%s) has been hit by competitor(%d)\n",
				elem.Time.Format(storage.TimeFmt),
				strings.TrimSpace(elem.ExtraParams),
				comp.ID)
		case 7:
			fmt.Printf("[%s] The competitor(%d) left the firing range\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID)
		case 8:
			comp.Laps[comp.CurrentLap].FiringRanges[comp.Laps[comp.CurrentLap].CurrentFiringLine].PenTimeStart = elem.Time
			fmt.Printf("[%s] The competitor(%d) entered the penalty laps\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID)
		case 9:
			comp.Laps[comp.CurrentLap].FiringRanges[comp.Laps[comp.CurrentLap].CurrentFiringLine].Duration = elem.Time.Sub(comp.Laps[comp.CurrentLap].FiringRanges[comp.Laps[comp.CurrentLap].CurrentFiringLine].PenTimeStart)
			fmt.Printf("[%s] The competitor(%d) left the penalty laps\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID)
		case 10:
			comp.Laps[comp.CurrentLap].LapDuration = elem.Time.Sub(comp.Laps[comp.CurrentLap].LapStart)
			comp.CurrentLap++
			comp.Laps = append(comp.Laps, storage.Lap{LapStart: elem.Time})
			fmt.Printf("[%s] The competitor(%d) ended the main lap\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID)
			if len(comp.Laps) > config.Laps {
				comp.Finished = elem.Time.Sub(comp.ActualStart)
				fmt.Printf("[%s] 33 The competitor(%d) has finished\n",
					elem.Time.Format(storage.TimeFmt),
					comp.ID)
			}
		case 11:

			comp.Comment = elem.ExtraParams
			comp.NotFinished = true
			fmt.Printf("[%s] The competitor(%d) can`t continue: %s\n",
				elem.Time.Format(storage.TimeFmt),
				comp.ID, elem.ExtraParams)

		default:
			log.Println("unexpected events")
		}

		competitors[elem.CompetitorID] = comp

	}
}
