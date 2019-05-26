package main

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb1-client/v2"
	"log"
	"math/rand"
	"time"
)

const (
	database = "nodes"
	username = "monitor"
	password = "secret"
)

var clusters = []string{"public", "private"}

func main() {
	c := influxDBClient()
	createMetrics(c)
	for _, cluster := range clusters {
		log.Printf("Mean values: cluster='%s', cpu_usage='%f'", cluster, meanCpuUsage(c, cluster))
	}
}

func influxDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	return c
}

func createMetrics(c client.Client) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	if err != nil {
		log.Fatalln("Error: ", err)
	}

	eventTime := time.Now().Add(time.Second * -20)

	var t time.Duration
	for t = 0; t < 20; t++ {
		for i := 0; i < 100; i++ {
			clusterIndex := rand.Intn(len(clusters))
			tags := map[string]string{
				"cluster": clusters[clusterIndex],
				"host":    fmt.Sprintf("192.168.%d.%d", clusterIndex, rand.Intn(100)),
			}

			fields := map[string]interface{}{
				"cpu_usage":  rand.Float64() * 100.0,
				"disk_usage": rand.Float64() * 100.0,
			}

			point, err := client.NewPoint(
				"node_status",
				tags,
				fields,
				eventTime.Add(time.Second*10),
			)
			if err != nil {
				log.Fatalln("Error: ", err)
			}

			bp.AddPoint(point)
		}
	}

	err = c.Write(bp)
	if err != nil {
		log.Fatal(err)
	}
}

func meanCpuUsage(c client.Client, cluster string) float64 {
	q := client.Query{
		Command:  fmt.Sprintf("select mean(cpu_usage) from node_status where cluster = '%s'", cluster),
		Database: database,
	}

	resp, err := c.Query(q)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	if resp.Error() != nil {
		log.Fatalln("Error: ", resp.Error())
	}

	res, err := resp.Results[0].Series[0].Values[0][1].(json.Number).Float64()
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	return res
}
