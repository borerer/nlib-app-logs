package database

import (
	"time"
)

const (
	CollectionLogs = "logs"
)

func (mc *MongoClient) AddLogs(level string, message string, details interface{}) error {
	doc := DBLogs{
		Level:     level,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().UnixMilli(),
	}
	if err := mc.InsertDocument(CollectionLogs, doc); err != nil {
		return err
	}
	return nil
}
