package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	_ "strconv"
	"strings"
	"time"
)





type Config struct{
	Laps int `json:"laps"`
	Laplen int `json:"lapLen"`
	PenaltyLen int `json:"penaltyLen"`
	FiringLines int `json:"firingLines"`
	Start string `json:"start"`
	StartDelta string `json:"startDelta"`
}

type Event struct {
	Time        time.Time
	EventID     int
	CompetitorID int
	ExtraParams string
	Original    string
}

type FiringRange struct{
	Shots []bool
	Id_FiringRange int
}

// Состояние участника
type Competitor struct {
	ID             int
	Registered     bool
	ScheduledStart time.Time
	WasOnStartLine bool
	OnStartLine time.Time
	ActualStart    time.Time
	Finished       bool
	Disqualified   bool
	NotFinished    bool
	CurrentLap     int
	OnFiringRange  int
	OnPenalty      bool
	FiringRanges []FiringRange
	// Shots          []bool // true - попадание, false - промах
	LapTimes       []time.Duration
	PenaltyTime    time.Duration
	LastEventTime  time.Time
	Comment        string
}

// Результаты участника
type Result struct {
	ID            int
	TotalTime     string
	LapTimes      []string
	LapSpeeds     []float64
	PenaltyTime   string
	PenaltySpeed  float64
	HitRatio      string
	Status        string
}

func ParseTime(timeStr string) (time.Time, error) {
    return time.Parse("05:04:05.000", timeStr)
}

func FormatTime(t time.Time) string {
	return t.Format("09:05:59.867")
}

func ReadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}


func ReadEvents(filename string) ([]Event, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []Event
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
		if err != nil{
			return nil, err
		}
		
		competitorID, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, err
		}
		extraParams := ""
		if len(parts)>3{
			extraParams+=parts[3]
		}

		events = append(events, Event{
			Time:        eventTime,
			EventID:     eventID,
			CompetitorID: competitorID,
			ExtraParams: extraParams,
		})
	}

	return events, scanner.Err()
}
// <log_time> <ID_EVENTS> <ID_COMPETITOR> <COMMENT>
//events 1 : участник зарегестрирован
//events 2 : назначает время старта для участника
//events 3 : участник на стартовой позиции
//events 4 : участник стартовал
//events 5 : участник на стрельбище
//events 6 : участник стреляет // если нет записи о попадание - значит промазал
//events 7 : участник покидает стрельбище
//events 8 : участник бежит на штрафные круги :)
//events 9 : участник покидает штрафные круги :)))
//events 10 : закончил основной круг


// основное решение - при добавление участника буду хранить его "позицию","ситуцию" и "результат" в map[id_comp]
// Конечно лучше использовать Postgres, т к он хранит просто огромные данные, что подразумевается программой, но тогда увеличиться время выполнения программы


func main(){

	if len(os.Args) < 3 {
		fmt.Println("Usage: cmd/main.go <config_file> <events_file>")
		os.Exit(1)
	}
	
	config, err := ReadConfig(os.Args[1])
	if err != nil {
		log.Printf("error: error reading config file \n")
		return
	}

	events, err := ReadEvents(os.Args[2])
	if err != nil {
		fmt.Printf("Error reading events: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(config)
	for _, elem := range events{
		fmt.Println(elem)
	}
	competitors := make(map[int]*Competitor)

	ProcessingEvents(competitors, events, &config)

	//на основе competitors составлять результаты

	// resultingTable(competitors)




}

func ProcessingEvents(competitors map[int]*Competitor, events []Event, config *Config){
	// type Competitor struct {
	// 	ID             int
	// 	Registered     bool
	// 	ScheduledStart time.Time
	// 	ActualStart    time.Time
	// 	Finished       bool
	// 	Disqualified   bool
	// 	NotFinished    bool
	// 	CurrentLap     int
	// 	OnFiringRange  bool
	// 	OnPenalty      bool
	// 	Shots          []bool // true - попадание, false - промах
	// 	LapTimes       []time.Duration
	// 	PenaltyTime    time.Duration
	// 	LastEventTime  time.Time
	// 	Comment        string
	// }

	for _, elem := range events{
		if _, ok := competitors[elem.CompetitorID]; !ok{
			competitors[elem.CompetitorID] = NewCompetitor(elem.CompetitorID, config)
		}
		comp := competitors[elem.CompetitorID]
		switch elem.EventID{
		case 1:
			comp.Registered = true
			fmt.Printf("[%s] The competitor(%d) registered\n", 
			elem.Time.Format("05:04:05.000"),
			comp.ID)
		case 2:
			eventTime, err := ConvertToTime(elem.ExtraParams)
			if err != nil{
				log.Printf("uncorrected time format")
				continue
			}
			comp.ScheduledStart = eventTime
			fmt.Printf("[%s] The start time for the competitor(%d) was set by a draw to [%s]\n", 
			elem.Time.Format("05:04:05.000"),
			comp.ID, eventTime)
		case 3:
			comp.WasOnStartLine = true
			comp.OnStartLine = elem.Time
			fmt.Printf("[%s] The competitor(%d) is on the start line\n", 
			elem.Time.Format("05:04:05.000"),
			comp.ID)
		case 4:
			comp.ActualStart = elem.Time
			startDelay := comp.ActualStart.Sub(comp.ScheduledStart)
			deltaTime, err := ParseStartDelta(config.StartDelta)
			if err != nil{
				log.Printf("uncorrected time format DeltaTime in config")
			}
			if startDelay > deltaTime {
				comp.Disqualified = true

				fmt.Printf("[%s] 32 %d\n", 
					comp.ActualStart.Format("05:04:05.000"),
					comp.ID)
			} else{
				comp.CurrentLap++
			}
			fmt.Printf("[%s] The competitor(%d) has started\n", 
			elem.Time.Format("05:04:05.000"),
			comp.ID)
		case 5:
			comp.OnFiringRange++
			fmt.Printf("[%s] The competitor(%d) is on the firing range(%d)\n", 
			elem.Time.Format("05:04:05.000"),
			comp.ID, comp.OnFiringRange)

		case 6:
			if len(comp.FiringRanges)<comp.CurrentLap{
				comp.FiringRanges = append(comp.FiringRanges, FiringRange{Shots: make([]bool, 0), Id_FiringRange: comp.OnFiringRange})				
			}
			comp.FiringRanges[comp.OnFiringRange-1].Shots = append(comp.FiringRanges[comp.OnFiringRange-1].Shots, true)
			fmt.Printf("[%s] The target(%s) has been hit by competitor(%d)\n", 
			elem.Time.Format("05:04:05.000"),
			elem.ExtraParams,
			comp.ID)
		case 7:
			fmt.Printf("[%s] The competitor(%d) left the firing range\n", 
			elem.Time.Format("05:04:05.000"),
			comp.ID)
		case 8:
			fmt.Printf("[%s] The competitor(%d) entered the penalty laps\n", 
			elem.Time.Format("05:04:05.000"),
			comp.ID)
		case 9:
			fmt.Printf("[%s] The competitor(%d) left the penalty laps\n", 
			elem.Time.Format("05:04:05.000"),
			comp.ID)
		case 10:
			comp.CurrentLap++
			if comp.CurrentLap > config.Laps{
				comp.Finished = true
				fmt.Printf("[%s] 33 The competitor(%d) has finished\n", 
				elem.Time.Format("05:04:05.000"),
				comp.ID)
			}
		case 11:

		default:
			log.Println("unexpected events")
		}

		competitors[elem.CompetitorID] = comp
		
	}
}

func NewCompetitor(id int, config *Config) *Competitor {
    return &Competitor{
        ID:       id,           
        LapTimes: make([]time.Duration, 0, config.Laps),
		FiringRanges: make([]FiringRange, 0, config.Laps * config.FiringLines),
    }
}


func ConvertToTime(time string)(time.Time, error){
	timeStr := strings.TrimPrefix(time, "[")
	timeStr = timeStr[:len(timeStr)-1]
	eventTime, err := ParseTime(timeStr)
	if err != nil{
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
