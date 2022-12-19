package main

import (
	"fmt"
	"os"
	"strconv"

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

func mustInt(in map[string]interface{}, key string) (int, error) {
	raw, ok := in[key]
	if !ok {
		return 0, fmt.Errorf("missing %s", key)
	}
	var ret int
	switch v := raw.(type) {
	case int:
		ret = v
	case float64:
		ret = int(v)
	case float32:
		ret = int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		ret = i
	default:
		return 0, fmt.Errorf("invalid type %s", key)
	}
	return ret, nil
}

func log(in map[string]interface{}) interface{} {
	level, err := mustString(in, "level")
	if err != nil {
		level = "info"
	}
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
	in["level"] = "debug"
	return log(in)
}

func info(in map[string]interface{}) interface{} {
	in["level"] = "info"
	return log(in)
}

func warn(in map[string]interface{}) interface{} {
	in["level"] = "warn"
	return log(in)
}

func error_(in map[string]interface{}) interface{} {
	in["level"] = "error"
	return log(in)
}

func get(in map[string]interface{}) interface{} {
	n, err := mustInt(in, "n")
	if err != nil {
		n = 100
	}
	skip, err := mustInt(in, "skip")
	if err != nil {
		skip = 0
	}
	logs, err := mongoClient.GetLogs(n, skip)
	if err != nil {
		return err.Error()
	}
	return logs
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
	nlib.RegisterFunction("log", log)
	nlib.RegisterFunction("debug", debug)
	nlib.RegisterFunction("info", info)
	nlib.RegisterFunction("warn", warn)
	nlib.RegisterFunction("error", error_)
	nlib.RegisterFunction("get", get)
	wait()
}
