package restoreservice

import (
	"context"
	"os"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	fileName := "tmpFile.json"
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove(fileName)
		if err != nil {
			t.Fatal(err)
		}
	}()

	data := "{\"id\":\"Counter-1\",\"type\":\"counter\",\"delta\":1}\n{\"id\":\"Counter-1\",\"type\":\"counter\",\"delta\":2}\n{\"id\":\"Counter-2\",\"type\":\"counter\",\"delta\":3}\n{\"id\":\"Gauge-1\",\"type\":\"gauge\",\"value\":1.1}\n{\"id\":\"Gauge-1\",\"type\":\"gauge\",\"value\":2.1}"
	/*
		{"id":"Counter-1","type":"counter","delta":1}
		{"id":"Counter-1","type":"counter","delta":2}
		{"id":"Counter-2","type":"counter","delta":3}
		{"id":"Gauge-1","type":"gauge","value":1.1}
		{"id":"Gauge-1","type":"gauge","value":2.1}`
	*/
	_, err = file.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	store := memstorage.New()
	rs := New(store, fileName)
	err = rs.Start()
	if err != nil {
		t.Fatal(err)
	}

	expectArr := []metric.MetricDB{
		metric.NewMetricDB(
			"Counter-1",
			metric.Counter(2),
		),
		metric.NewMetricDB(
			"Counter-2",
			metric.Counter(3),
		),
		metric.NewMetricDB(
			"Gauge-1",
			metric.Gauge(2.1),
		),
	}
	actualArr := store.List(context.Background())

	assert.ElementsMatch(t, expectArr, actualArr)
}
