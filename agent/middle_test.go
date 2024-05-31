package agent

import (
	"net/http"
)

type fakeRound struct {
	err   error
	count int
}

func (fr *fakeRound) RoundTrip(req *http.Request) (*http.Response, error) {
	fr.count++
	return nil, fr.err
}

/*
func TestRetry(t *testing.T) {
	werr := errors.New("send err")
	maxRetries := 2
	fr := fakeRound{count: 0, err: werr}
	rrt := retryRoundTripper{
		next:           &fr,
		maxRetries:     maxRetries,
		delayIncrement: 10 * time.Millisecond,
		log:            slog.Default(),
	}

	req, err := http.NewRequest(http.MethodGet, "test", nil)
	if err != nil {
		t.Errorf("build req: %v\n", err)
	}

	resp, err2 := rrt.RoundTrip(req)

	assert.Equal(t, werr, err2)

	if err := resp.Body.Close(); err != nil {
		t.Error(err)
	}

	assert.Equal(t, maxRetries, fr.count)
}
*/
