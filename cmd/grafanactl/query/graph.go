package query

import (
	"errors"
	"io"

	"github.com/grafana/grafanactl/internal/format"
	"github.com/grafana/grafanactl/internal/graph"
	"github.com/grafana/grafanactl/internal/query/loki"
	"github.com/grafana/grafanactl/internal/query/prometheus"
)

type queryGraphCodec struct{}

func (c *queryGraphCodec) Format() format.Format {
	return "graph"
}

func (c *queryGraphCodec) Encode(w io.Writer, data any) error {
	var chartData *graph.ChartData
	var err error

	switch resp := data.(type) {
	case *prometheus.QueryResponse:
		chartData, err = graph.FromPrometheusResponse(resp)
		if err != nil {
			return err
		}
	case *loki.QueryResponse:
		chartData, err = graph.FromLokiResponse(resp)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid data type for graph codec (expected *prometheus.QueryResponse or *loki.QueryResponse)")
	}

	opts := graph.DefaultChartOptions()
	return graph.RenderChart(w, chartData, opts)
}

func (c *queryGraphCodec) Decode(io.Reader, any) error {
	return errors.New("graph codec does not support decoding")
}
