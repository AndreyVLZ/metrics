package memstorage

/*
func TestSetMemStore(t *testing.T) {
	type data struct {
		typeStr string
		name    string
		valStr  string
		err     error
	}

	tc := []struct {
		nameTest       string
		data           []data
		wantGaugeList  map[string]string
		wantConterList map[string]string
	}{
		{
			nameTest: "positive #1",
			data: []data{
				{
					typeStr: "gauge",
					name:    "myGauge",
					valStr:  "12.3",
					err:     nil,
				},
				{
					typeStr: "counter",
					name:    "myCounter",
					valStr:  "12",
					err:     nil,
				},
			},
			wantGaugeList: map[string]string{
				"myGauge": "12.3",
			},
			wantConterList: map[string]string{
				"myCounter": "12",
			},
		},
		{
			nameTest: "positive #2",
			data: []data{
				{
					typeStr: "gauge",
					name:    "myGauge",
					valStr:  "12.3",
					err:     nil,
				},
				{
					typeStr: "gauge",
					name:    "myGauge",
					valStr:  "1234",
					err:     nil,
				},
				{
					typeStr: "counter",
					name:    "myCounter",
					valStr:  "12",
					err:     nil,
				},
			},
			wantGaugeList: map[string]string{
				"myGauge": "1234",
			},
			wantConterList: map[string]string{
				"myCounter": "12",
			},
		},
		{
			nameTest: "positive #3",
			data: []data{
				{
					typeStr: "gauge",
					name:    "myGauge",
					valStr:  "12.3",
					err:     nil,
				},
				{
					typeStr: "counter",
					name:    "myCounter",
					valStr:  "12",
					err:     nil,
				},
				{
					typeStr: "counter",
					name:    "myCounter",
					valStr:  "21",
					err:     nil,
				},
			},
			wantGaugeList: map[string]string{
				"myGauge": "12.3",
			},
			wantConterList: map[string]string{
				"myCounter": "33",
			},
		},
		{
			nameTest: "negative #1",
			data: []data{
				{
					typeStr: "gauge",
					name:    "myGauge",
					valStr:  "12.3a",
					err:     metric.ErrStringIsNotValid,
				},
				{
					typeStr: "counter",
					name:    "myCounter",
					valStr:  "12",
					err:     nil,
				},
			},
			wantGaugeList: map[string]string{},
			wantConterList: map[string]string{
				"myCounter": "12",
			},
		},
		{
			nameTest: "negative #2",
			data: []data{
				{
					typeStr: "gauge",
					name:    "myGauge",
					valStr:  "12.3",
					err:     nil,
				},
				{
					typeStr: "counter",
					name:    "myCounter",
					valStr:  "12a",
					err:     metric.ErrStringIsNotValid,
				},
			},
			wantGaugeList: map[string]string{
				"myGauge": "12.3",
			},
			wantConterList: map[string]string{},
		},
		{
			nameTest: "negative #3",
			data: []data{
				{
					typeStr: "gaug",
					name:    "myGauge",
					valStr:  "12.3",
					err:     ErrNotSupportedType,
				},
				{
					typeStr: "counter",
					name:    "myCounter",
					valStr:  "12",
					err:     nil,
				},
			},
			wantGaugeList: map[string]string{},
			wantConterList: map[string]string{
				"myCounter": "12",
			},
		},
		{
			nameTest: "negative #4",
			data: []data{
				{
					typeStr: "gauge",
					name:    "myGauge",
					valStr:  "12.3",
					err:     nil,
				},
				{
					typeStr: "counter1",
					name:    "myCounter",
					valStr:  "12",
					err:     ErrNotSupportedType,
				},
			},
			wantGaugeList: map[string]string{
				"myGauge": "12.3",
			},
			wantConterList: map[string]string{},
		},
	}

	for _, test := range tc {
		t.Run(test.nameTest, func(t *testing.T) {
			store := New()
			for _, d := range test.data {
				err := store.Set(metric.MetricDB{})
				assert.Equal(t, err, d.err)
			}
			assert.Equal(t, test.wantGaugeList, store.GaugeRepo().List())
			assert.Equal(t, test.wantConterList, store.CounterRepo().List())
		})
	}
}
*/
