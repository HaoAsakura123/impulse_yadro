package storage

import "time"

const (
	TimeFmt = "15:04:05.000"
)

type Config struct {
	Laps        int    `json:"laps"`
	Laplen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

type Event struct {
	Time         time.Time
	EventID      int
	CompetitorID int
	ExtraParams  string
	Original     string
}

type FiringRange struct {
	Shots          []bool
	Id_FiringRange int
	PenTimeStart   time.Time
	Duration       time.Duration
}

type Lap struct {
	FiringRanges      []FiringRange
	LapStart          time.Time
	LapDuration       time.Duration
	CurrentFiringLine int // обновлять когда проходит круг
}

// Состояние участника
type Competitor struct {
	ID             int
	Registered     bool
	ScheduledStart time.Time
	WasOnStartLine bool
	OnStartLine    time.Time
	ActualStart    time.Time
	Finished       time.Duration
	Disqualified   bool
	NotFinished    bool
	CurrentLap     int
	OnPenalty      bool

	Laps []Lap // при добавление круга обновлять данные

	LastEventTime time.Time
	Comment       string
}

type Result struct {
	Finished      string        //+
	ID_competitor int           // +
	Laps          []LapsInfo    // +
	Penalty       []PenaltyInfo //+
	Shots         string
	AllDuration   time.Duration
}

type LapsInfo struct {
	TimeLaps     time.Duration
	AvgSpeedLaps float64
}

type PenaltyInfo struct {
	TimePenalty     time.Duration
	AvgSpeedPenalty float64
}
