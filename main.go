package main

import (
	"fmt"
	"os"

	"github.com/borerer/nlib-app-logs/database"
	nlibgo "github.com/borerer/nlib-go"
)

var (
	mongoClient *database.MongoClient
)

func mustString(in map[string]interface{}, key string) (string, error) {
	raw, ok := in[key]
	if !ok {
		return "", fmt.Errorf("missing %s", key)
	}
	str, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("invalid type %s", key)
	}
	return str, nil
}

func log(level string, in map[string]interface{}) interface{} {
	message, err := mustString(in, "message")
	if err != nil {
		return err.Error()
	}
	err = mongoClient.AddLogs(level, message, in["details"])
	if err != nil {
		return err.Error()
	}
	return "ok"
}

func debug(in map[string]interface{}) interface{} {
	return log("debug", in)
}

func info(in map[string]interface{}) interface{} {
	return log("info", in)
}

func warn(in map[string]interface{}) interface{} {
	return log("warn", in)
}

func error_(in map[string]interface{}) interface{} {
	return log("error", in)
}

func wait() {
	ch := make(chan bool)
	<-ch
}

func main() {
	mongoClient = database.NewMongoClient(&database.MongoConfig{
		URI:      os.Getenv("NLIB_MONGO_URI"),
		Database: os.Getenv("NLIB_MONGO_DATABASE"),
	})
	if err := mongoClient.Start(); err != nil {
		println(err.Error())
		return
	}
	nlib := nlibgo.NewClient(os.Getenv("NLIB_SERVER"), "logs")
	nlib.RegisterFunction("debug", debug)
	nlib.RegisterFunction("info", info)
	nlib.RegisterFunction("warn", warn)
	nlib.RegisterFunction("error", error_)
	wait()
}
