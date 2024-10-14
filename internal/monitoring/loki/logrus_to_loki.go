package loki

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	httpClient = &http.Client{}
)

func InitializeLoki() {
	log.Info("Setting up Loki hook!")
	log.AddHook(
		&LokiHook{
			LokiHost:  os.Getenv("LOKI_HOST"),
			LokiUser:  os.Getenv("LOKI_USER"),
			LokiToken: os.Getenv("LOKI_TOKEN"),
		},
	)
}

type LokiHook struct {
	LokiHost  string
	LokiUser  string
	LokiToken string
}

func (hook *LokiHook) Levels() []log.Level {
	return log.AllLevels
}

type StreamData struct {
	Level log.Level `json:"level"`
	log.Fields
	ServiceName string `json:"service_name"`
}

type Stream struct {
	StreamData `json:"stream"`
	Values     [][]string `json:"values"`
}

type StreamsData struct {
	Streams []Stream `json:"streams"`
}

func (hook *LokiHook) Fire(entry *log.Entry) error {
	logData, err := getLokiIngestCompatibleStream(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", hook.LokiHost, bytes.NewBuffer(logData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(hook.LokiUser, hook.LokiToken)
	httpClient.Do(req)

	return nil
}

func getLokiIngestCompatibleStream(entry *log.Entry) ([]byte, error) {
	var currentTime = strconv.FormatInt(time.Now().UnixNano(), 10)

	streamsData := StreamsData{
		Streams: []Stream{
			{
				StreamData: StreamData{
					entry.Level,
					entry.Data,
					"prog-demos-backend",
				},
				Values: [][]string{
					{currentTime, entry.Message},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(streamsData)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}
