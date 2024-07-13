package agent_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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

	ctxStart, cancelStart := context.WithTimeout(ctx, 2*time.Second)
	defer cancelStart()

	ctxTimeoutStop, cancelStop := context.WithTimeout(ctx, time.Second)
	defer cancelStop()

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

	cfg := config.Default()
	agent := agent.New(cfg, slog.Default())

	if err := agent.Start(ctxStart); err != nil {
		t.Errorf("start agent err: %v\n", err)
	}

	select {
	case <-ctxStart.Done():
	case err := <-agent.Err():
		t.Errorf("run agent err: %v\n", err)
	case <-exit:
	}

	if err := agent.Stop(ctxTimeoutStop); err != nil {
		t.Logf("agent stop err: %v\n", err)
	}
}
