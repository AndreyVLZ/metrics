package grps

import (
	"strings"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	pb "github.com/AndreyVLZ/metrics/internal/proto"
)

func TestProtoArrray(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		arr := []model.Metric{
			model.NewCounterMetric("counter-1", 10),
			model.NewGaugeMetric("gauge-1", 10.01),
		}

		protoArr := buildProtoArray(arr)
		compare(t, arr, protoArr)
	})
}

func compare(t *testing.T, arr []model.Metric, protoArr []*pb.Metric) {
	if len(arr) != len(protoArr) {
		t.Errorf("compare len: %d - %d", len(arr), len(protoArr))
	}

	for i := range arr {
		if arr[i].MName != protoArr[i].GetInfo().GetName() {
			t.Errorf("compare name: %s - %s", arr[i].MName, protoArr[i].GetInfo().GetName())
		}

		pType := strings.ToLower(protoArr[i].GetInfo().GetType().String())
		if arr[i].MType.String() != pType {
			t.Errorf("compare types: %s - %s", arr[i].MType.String(), pType)
		}
	}
}
