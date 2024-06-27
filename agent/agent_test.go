package agent_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/agent/config"
	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	exit := make(chan struct{})
	wantEndpoint := "/updates/"

	ctxStart, cancelStart := context.WithCancel(ctx)
	defer cancelStart()

	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 2*time.Second)
	defer cancelTimeout()

	tsrv := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			exit <- struct{}{}
		}()
		assert.Equal(t, wantEndpoint, req.URL.String())

		body := req.Body
		defer body.Close()

		data, err := io.ReadAll(body)
		if err != nil {
			t.Errorf("read body err: %v\n", err)
		}

		if len(data) == 0 {
			t.Error("len data ==0")
		}
	}))
	defer tsrv.Close()

	cfg, err := config.New(
		config.SetAddr(strings.TrimPrefix(tsrv.URL, "http://")),
		config.SetPollInterval(1*time.Second),
		config.SetReportInterval(1*time.Second),
		config.SetRateLimit(4),
	)
	if err != nil {
		t.Errorf("new config: %v\n", err)
	}

	agent := agent.New(cfg, slog.Default())

	if err := agent.Start(ctxStart); err != nil {
		t.Errorf("start agent err: %v\n", err)
	}

	select {
	case <-ctxTimeout.Done():
		t.Errorf("ctx is done")
	case err := <-agent.Err():
		t.Errorf("run agent err: %v\n", err)
	case <-exit:
	}

	cancelStart()

	ctxTimeoutStop, cancelStop := context.WithTimeout(ctx, time.Second)
	defer cancelStop()

	if err := agent.Stop(ctxTimeoutStop); err != nil {
		t.Logf("agent stop err: %v\n", err)
	}
}
