package urlpath

import (
	"testing"

	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestNewGetURLPath(t *testing.T) {
	tc := []struct {
		name string
		arr  []string
		want *GetURLPath
	}{
		{
			name: "positive #1",
			arr:  []string{"gauge", "myGauge"},
			want: &GetURLPath{
				typeStr: metric.GaugeType.String(),
				name:    "myGauge",
			},
		},
		{
			name: "positive #2",
			arr:  []string{"counter", "myCounter"},
			want: &GetURLPath{
				typeStr: metric.CounterType.String(),
				name:    "myCounter",
			},
		},
		{
			name: "positive #3",
			arr:  []string{"counter", "myCounter", "test"},
			want: &GetURLPath{
				typeStr: metric.CounterType.String(),
				name:    "myCounter",
			},
		},
		{
			name: "positive #4",
			arr:  []string{"co", ""},
			want: &GetURLPath{
				typeStr: "co",
				name:    "",
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			urlPath := NewGetURLPath(test.arr...)
			assert.Equal(t, test.want, urlPath)
		})
	}
}

func TestGetURLPathValidate(t *testing.T) {
	tc := []struct {
		name string
		arr  []string
		want *GetURLPath
		err  error
	}{
		{
			name: "positive #1",
			arr:  []string{"gauge", "myGauge"},
			want: &GetURLPath{
				typeStr: metric.GaugeType.String(),
				name:    "myGauge",
			},
			err: nil,
		},
		{
			name: "positive #2",
			arr:  []string{"counter", "myCounter"},
			want: &GetURLPath{
				typeStr: metric.CounterType.String(),
				name:    "myCounter",
			},
			err: nil,
		},
		{
			name: "positive #3",
			arr:  []string{"counter", "myCounter", "test"},
			want: &GetURLPath{
				typeStr: metric.CounterType.String(),
				name:    "myCounter",
			},
			err: nil,
		},
		{
			name: "negative #1",
			arr:  []string{"co", ""},
			want: &GetURLPath{
				typeStr: "co",
				name:    "",
			},
			err: ErrNoCorrectURLPath,
		},
		{
			name: "negative #2",
			arr:  []string{"", "1"},
			want: &GetURLPath{
				typeStr: "",
				name:    "1",
			},
			err: ErrNoCorrectURLPath,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			urlPath := NewGetURLPath(test.arr...)
			assert.Equal(t, test.err, urlPath.Validate())
		})
	}

}
