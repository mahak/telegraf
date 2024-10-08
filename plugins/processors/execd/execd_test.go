package execd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/processors"
	_ "github.com/influxdata/telegraf/plugins/serializers/all"
	serializers_influx "github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestExternalProcessorWorks(t *testing.T) {
	e := New()
	e.Log = testutil.Logger{}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	e.SetParser(parser)

	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	e.SetSerializer(serializer)

	exe, err := os.Executable()
	require.NoError(t, err)
	t.Log(exe)
	e.Command = []string{exe, "-countmultiplier"}
	e.Environment = []string{"PLUGINS_PROCESSORS_EXECD_MODE=application", "FIELD_NAME=count"}
	e.RestartDelay = config.Duration(5 * time.Second)

	acc := &testutil.Accumulator{}

	require.NoError(t, e.Start(acc))

	now := time.Now()
	orig := now
	for i := 0; i < 10; i++ {
		m := metric.New("test",
			map[string]string{
				"city": "Toronto",
			},
			map[string]interface{}{
				"population": 6000000,
				"count":      1,
			},
			now)
		now = now.Add(1)

		require.NoError(t, e.Add(m, acc))
	}

	acc.Wait(1)
	e.Stop()
	acc.Wait(9)

	metrics := acc.GetTelegrafMetrics()
	m := metrics[0]

	expected := testutil.MustMetric("test",
		map[string]string{
			"city": "Toronto",
		},
		map[string]interface{}{
			"population": 6000000,
			"count":      2,
		},
		orig,
	)
	testutil.RequireMetricEqual(t, expected, m)

	metricTime := m.Time().UnixNano()

	// make sure the other 9 are ordered properly
	for i := 0; i < 9; i++ {
		m = metrics[i+1]
		require.EqualValues(t, metricTime+1, m.Time().UnixNano())
		metricTime = m.Time().UnixNano()
	}
}

func TestParseLinesWithNewLines(t *testing.T) {
	e := New()
	e.Log = testutil.Logger{}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	e.SetParser(parser)

	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	e.SetSerializer(serializer)

	exe, err := os.Executable()
	require.NoError(t, err)
	t.Log(exe)
	e.Command = []string{exe, "-countmultiplier"}
	e.Environment = []string{"PLUGINS_PROCESSORS_EXECD_MODE=application", "FIELD_NAME=count"}
	e.RestartDelay = config.Duration(5 * time.Second)

	acc := &testutil.Accumulator{}

	require.NoError(t, e.Start(acc))

	now := time.Now()
	orig := now

	m := metric.New("test",
		map[string]string{
			"author": "Mr. Gopher",
		},
		map[string]interface{}{
			"phrase": "Gophers are amazing creatures.\nAbsolutely amazing.",
			"count":  3,
		},
		now)

	require.NoError(t, e.Add(m, acc))

	acc.Wait(1)
	e.Stop()

	processedMetric := acc.GetTelegrafMetrics()[0]

	expectedMetric := testutil.MustMetric("test",
		map[string]string{
			"author": "Mr. Gopher",
		},
		map[string]interface{}{
			"phrase": "Gophers are amazing creatures.\nAbsolutely amazing.",
			"count":  6,
		},
		orig,
	)

	testutil.RequireMetricEqual(t, expectedMetric, processedMetric)
}

var countmultiplier = flag.Bool("countmultiplier", false,
	"if true, act like line input program instead of test")

func TestMain(m *testing.M) {
	flag.Parse()
	runMode := os.Getenv("PLUGINS_PROCESSORS_EXECD_MODE")
	if *countmultiplier && runMode == "application" {
		runCountMultiplierProgram()
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}

func runCountMultiplierProgram() {
	fieldName := os.Getenv("FIELD_NAME")
	parser := influx.NewStreamParser(os.Stdin)
	serializer := &serializers_influx.Serializer{}
	//nolint:errcheck // this should always succeed
	serializer.Init()

	for {
		m, err := parser.Next()
		if err != nil {
			if errors.Is(err, influx.EOF) {
				return // stream ended
			}
			var parseErr *influx.ParseError
			if errors.As(err, &parseErr) {
				fmt.Fprintf(os.Stderr, "parse ERR %v\n", parseErr)
				//nolint:revive // os.Exit called intentionally
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			//nolint:revive // os.Exit called intentionally
			os.Exit(1)
		}

		c, found := m.GetField(fieldName)
		if !found {
			fmt.Fprintf(os.Stderr, "metric has no %s field\n", fieldName)
			//nolint:revive // os.Exit called intentionally
			os.Exit(1)
		}
		switch t := c.(type) {
		case float64:
			t *= 2
			m.AddField(fieldName, t)
		case int64:
			t *= 2
			m.AddField(fieldName, t)
		default:
			fmt.Fprintf(os.Stderr, "%s is not an unknown type, it's a %T\n", fieldName, c)
			//nolint:revive // os.Exit called intentionally
			os.Exit(1)
		}
		b, err := serializer.Serialize(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			//nolint:revive // os.Exit called intentionally
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, string(b))
	}
}

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	// Set up for file inputs
	processors.AddStreaming("execd", func() telegraf.StreamingProcessor {
		return New()
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		t.Run(fname, func(t *testing.T) {
			testdataPath := filepath.Join("testcases", fname)
			configFilename := filepath.Join(testdataPath, "telegraf.conf")
			inputFilename := filepath.Join(testdataPath, "input.influx")
			expectedFilename := filepath.Join(testdataPath, "expected.out")

			// Get parser to parse input and expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			input, err := testutil.ParseMetricsFromFile(inputFilename, parser)
			require.NoError(t, err)

			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Processors, 1, "wrong number of outputs")
			plugin := cfg.Processors[0].Processor

			// Process the metrics
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			for _, m := range input {
				require.NoError(t, plugin.Add(m, &acc))
			}
			plugin.Stop()

			require.Eventually(t, func() bool {
				acc.Lock()
				defer acc.Unlock()
				return acc.NMetrics() >= uint64(len(expected))
			}, time.Second, 100*time.Millisecond)

			// Check the expectations
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual)
		})
	}
}

func TestTracking(t *testing.T) {
	now := time.Now()

	// Setup the raw  input and expected output data
	inputRaw := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{
				"city": "Toronto",
			},
			map[string]interface{}{
				"population": 6000000,
				"count":      1,
			},
			now,
		),
		metric.New(
			"test",
			map[string]string{
				"city": "Tokio",
			},
			map[string]interface{}{
				"population": 14000000,
				"count":      8,
			},
			now,
		),
	}

	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{
				"city": "Toronto",
			},
			map[string]interface{}{
				"population": 6000000,
				"count":      2,
			},
			now,
		),
		metric.New(
			"test",
			map[string]string{
				"city": "Tokio",
			},
			map[string]interface{}{
				"population": 14000000,
				"count":      16,
			},
			now,
		),
	}

	// Create a testing notifier
	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	// Convert raw input to tracking metrics
	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	// Setup the plugin
	exe, err := os.Executable()
	require.NoError(t, err)

	plugin := &Execd{
		Command:      []string{exe, "-countmultiplier"},
		Environment:  []string{"PLUGINS_PROCESSORS_EXECD_MODE=application", "FIELD_NAME=count"},
		RestartDelay: config.Duration(5 * time.Second),
		Log:          testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Process expected metrics and compare with resulting metrics
	for _, in := range input {
		require.NoError(t, plugin.Add(in, &acc))
	}
	require.Eventually(t, func() bool {
		return int(acc.NMetrics()) >= len(expected)
	}, 3*time.Second, 100*time.Millisecond)

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
