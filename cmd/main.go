package main

import (
	"fmt"
	"impulse_yadro/internal/app"
	"impulse_yadro/internal/storage"
	"log"
	"os"
)

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
// Конечно лучше использовать Postgres, т к он хранит просто огромные данные, что подразумевается программой, но тогда увеличиться время выполнения
// программы и существенно увеличиться код, в виду всех "миграций" и операций с базой данных
// Лично я бы тут использовал MongoDB для более удобного поиска по коллекциям
//

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: cmd/main.go <config_file> <events_file>")
		os.Exit(1)
	}

	config, err := app.ReadConfig(os.Args[1])
	if err != nil {
		log.Printf("error: error reading config file \n")
		return
	}

	events, err := app.ReadEvents(os.Args[2])
	if err != nil {
		fmt.Printf("Error reading events: %v\n", err)
		os.Exit(1)
	}

	competitors := make(map[int]*storage.Competitor)

	app.ProcessingEvents(competitors, events, &config)

	//на основе competitors составлять результаты

	app.ResultingTable(competitors, &config)

}
