//go:generate ../../../tools/readme_config_includer/generator
package minmax

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

//go:embed sample.conf
var sampleConfig string

type MinMax struct {
	cache map[uint64]aggregate
}

type aggregate struct {
	fields map[string]minmax
	name   string
	tags   map[string]string
}

type minmax struct {
	min float64
	max float64
}

func (*MinMax) SampleConfig() string {
	return sampleConfig
}

func (m *MinMax) Add(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := m.cache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		a := aggregate{
			name:   in.Name(),
			tags:   in.Tags(),
			fields: make(map[string]minmax),
		}
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				a.fields[k] = minmax{
					min: fv,
					max: fv,
				}
			}
		}
		m.cache[id] = a
	} else {
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				if _, ok := m.cache[id].fields[k]; !ok {
					// hit an uncached field of a cached metric
					m.cache[id].fields[k] = minmax{
						min: fv,
						max: fv,
					}
					continue
				}
				if fv < m.cache[id].fields[k].min {
					tmp := m.cache[id].fields[k]
					tmp.min = fv
					m.cache[id].fields[k] = tmp
				} else if fv > m.cache[id].fields[k].max {
					tmp := m.cache[id].fields[k]
					tmp.max = fv
					m.cache[id].fields[k] = tmp
				}
			}
		}
	}
}

func (m *MinMax) Push(acc telegraf.Accumulator) {
	for _, aggregate := range m.cache {
		fields := make(map[string]interface{}, len(aggregate.fields))
		for k, v := range aggregate.fields {
			fields[k+"_min"] = v.min
			fields[k+"_max"] = v.max
		}
		acc.AddFields(aggregate.name, fields, aggregate.tags)
	}
}

func (m *MinMax) Reset() {
	m.cache = make(map[uint64]aggregate)
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func newMinMax() telegraf.Aggregator {
	mm := &MinMax{}
	mm.Reset()
	return mm
}

func init() {
	aggregators.Add("minmax", func() telegraf.Aggregator {
		return newMinMax()
	})
}
