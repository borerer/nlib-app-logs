package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/borerer/nlib-app-logs/database"
	nlib "github.com/borerer/nlib-go"
	"github.com/borerer/nlib-go/har"
	nlibshared "github.com/borerer/nlib-shared/go"
)

var (
	mongoClient *database.MongoClient
)

func getQuery(req *nlib.FunctionIn, key string) string {
	for _, query := range req.QueryString {
		if query.Name == key {
			return query.Value
		}
	}
	return ""
}

func getQueryAsInt(req *nlib.FunctionIn, key string) int {
	val := getQuery(req, key)
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return i
}

func logGET(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	level := getQuery(req, "level")
	if len(level) == 0 {
		level = "info"
	}
	message := getQuery(req, "message")
	err := mongoClient.AddLogs(level, message, nil)
	if err != nil {
		return nil, err
	}
	return har.Text("ok"), nil
}

func logPOST(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	parseLog := func(req *nlib.FunctionIn) (string, string, interface{}) {
		if req.PostData != nil && req.PostData.Text != nil {
			var j map[string]interface{}
			err := json.Unmarshal([]byte(*req.PostData.Text), &j)
			if err == nil {
				key := j["level"].(string)
				message := j["message"].(string)
				details := j["details"]
				return key, message, details
			}
		}
		return "", "", nil
	}
	level, message, details := parseLog(req)
	err := mongoClient.AddLogs(level, message, details)
	if err != nil {
		return nil, err
	}
	return har.Text("ok"), nil
}

func log(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	if req.Method == "GET" {
		return logGET(req)
	} else if req.Method == "POST" {
		return logPOST(req)
	}
	return har.Err405, nil
}

func debug(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	req.QueryString = append(req.QueryString, nlibshared.Query{Name: "level", Value: "debug"})
	return log(req)
}

func info(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	req.QueryString = append(req.QueryString, nlibshared.Query{Name: "level", Value: "info"})
	return log(req)
}

func warn(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	req.QueryString = append(req.QueryString, nlibshared.Query{Name: "level", Value: "warn"})
	return log(req)
}

func error_(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	req.QueryString = append(req.QueryString, nlibshared.Query{Name: "level", Value: "error"})
	return log(req)
}

func get(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	n := getQueryAsInt(req, "n")
	if n == 0 {
		n = 100
	}
	skip := getQueryAsInt(req, "skip")
	logs, err := mongoClient.GetLogs(n, skip)
	if err != nil {
		return nil, err
	}
	return har.JSON(logs), nil
}

func main() {
	mongoClient = database.NewMongoClient(&database.MongoConfig{
		URI:      os.Getenv("NLIB_MONGO_URI"),
		Database: os.Getenv("NLIB_MONGO_DATABASE"),
	})
	nlib.Must(mongoClient.Start())

	nlib.SetEndpoint(os.Getenv("NLIB_SERVER"))
	nlib.SetAppID("logs")
	nlib.Must(nlib.Connect())

	nlib.RegisterFunction("log", log)
	nlib.RegisterFunction("debug", debug)
	nlib.RegisterFunction("info", info)
	nlib.RegisterFunction("warn", warn)
	nlib.RegisterFunction("error", error_)
	nlib.RegisterFunction("get", get)
	nlib.Wait()
}
