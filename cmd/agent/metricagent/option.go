package metricagent

type FuncOpt func(c *MetricClient)

func SetAddr(addr string) FuncOpt {
	return func(c *MetricClient) {
		c.addr = addr
	}
}

func SetPollInterval(pollInterval int) FuncOpt {
	return func(c *MetricClient) {
		c.pollInterval = pollInterval
	}
}

func SetReportInterval(reportInterval int) FuncOpt {
	return func(c *MetricClient) {
		c.reportInterval = reportInterval
	}
}

func SetKey(key string) FuncOpt {
	return func(c *MetricClient) {
		c.key = key
	}
}
func SetRateLimit(rateLimit int) FuncOpt {
	return func(c *MetricClient) {
		c.rateLimit = rateLimit
	}
}
