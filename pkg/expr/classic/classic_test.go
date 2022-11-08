package classic

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/require"
	ptr "github.com/xorcare/pointer"

	"github.com/grafana/grafana/pkg/expr/mathexp"
)

func TestConditionsCmd(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *ConditionsCmd
		vars     mathexp.Vars
		expected func() mathexp.Results
	}{
		{
			name: "single query and single condition",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedSeries(ptr.Float64(30), ptr.Float64(40)),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: 34},
					},
				}},
			expected: func() mathexp.Results {
				v := valBasedNumber(ptr.Float64(1))
				v.SetMeta([]EvalMatch{{Value: ptr.Float64(35)}})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "single query and single condition - empty series",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedSeries(),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: 34},
					},
				}},
			expected: func() mathexp.Results {
				v := valBasedNumber(nil)
				v.SetMeta([]EvalMatch{{Metric: "NoData"}})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "single query and single condition - empty series and not empty series",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedSeries(),
						valBasedSeries(ptr.Float64(3)),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: .5},
					},
				}},
			expected: func() mathexp.Results {
				v := valBasedNumber(ptr.Float64(1))
				v.SetMeta([]EvalMatch{{Value: ptr.Float64(3)}})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "single query and two conditions",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedSeries(ptr.Float64(30), ptr.Float64(40)),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("max"),
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: 34},
					},
					{
						InputRefID: "A",
						Reducer:    reducer("min"),
						Operator:   "or",
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: 12},
					},
				}},
			expected: func() mathexp.Results {
				v := valBasedNumber(ptr.Float64(1))
				v.SetMeta([]EvalMatch{{Value: ptr.Float64(40)}, {Value: ptr.Float64(30)}})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "single query and single condition - multiple series (one true, one not == true)",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedSeriesWithLabels(data.Labels{"h": "1"}, ptr.Float64(30), ptr.Float64(40)),
						valBasedSeries(ptr.Float64(0), ptr.Float64(10)),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: 34},
					},
				}},
			expected: func() mathexp.Results {
				v := valBasedNumber(ptr.Float64(1))
				v.SetMeta([]EvalMatch{{Value: ptr.Float64(35), Labels: data.Labels{"h": "1"}}})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "single query and single condition - multiple series (one not true, one true == true)",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedSeries(ptr.Float64(0), ptr.Float64(10)),
						valBasedSeries(ptr.Float64(30), ptr.Float64(40)),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: 34},
					},
				}},
			expected: func() mathexp.Results {
				v := valBasedNumber(ptr.Float64(1))
				v.SetMeta([]EvalMatch{{Value: ptr.Float64(35)}})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "single query and single condition - multiple series (2 not true == false)",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedSeries(ptr.Float64(0), ptr.Float64(10)),
						valBasedSeries(ptr.Float64(20), ptr.Float64(30)),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: 34},
					},
				}},
			expected: func() mathexp.Results {
				v := valBasedNumber(ptr.Float64(0))
				v.SetMeta([]EvalMatch{})
				return mathexp.Results{Values: mathexp.Values{v}}
			},
		},
		{
			name: "single query and single ranged condition",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedSeries(ptr.Float64(30), ptr.Float64(40)),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("diff"),
						Operator:   "and",
						Evaluator:  &rangedEvaluator{Type: "within_range", Lower: 2, Upper: 3},
					},
				},
			},
			expected: func() mathexp.Results {
				v := valBasedNumber(ptr.Float64(0))
				v.SetMeta([]EvalMatch{})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "single query with no data",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{mathexp.NoData{}.New()},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{"gt", 1},
					},
				},
			},
			expected: func() mathexp.Results {
				v := valBasedNumber(nil)
				v.SetMeta([]EvalMatch{{Metric: "NoData"}})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "single query with no values",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{"gt", 1},
					},
				},
			},
			expected: func() mathexp.Results {
				v := valBasedNumber(nil)
				v.SetMeta([]EvalMatch{{Metric: "NoData"}})
				return mathexp.NewResults(v)
			},
		},
		{
			name: "should accept numbers",
			vars: mathexp.Vars{
				"A": mathexp.Results{
					Values: []mathexp.Value{
						valBasedNumber(ptr.Float64(5)),
						valBasedNumber(ptr.Float64(10)),
						valBasedNumber(ptr.Float64(15)),
					},
				},
			},
			cmd: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{"gt", 1},
					},
				},
			},
			expected: func() mathexp.Results {
				v := valBasedNumber(ptr.Float64(1))
				v.SetMeta([]EvalMatch{
					{Value: ptr.Float64(5)},
					{Value: ptr.Float64(10)},
					{Value: ptr.Float64(15)},
				})
				return mathexp.NewResults(v)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cmd.Execute(context.Background(), time.Now(), tt.vars)
			require.NoError(t, err)

			require.Equal(t, 1, len(res.Values))

			require.Equal(t, tt.expected(), res)
		})
	}
}

func TestUnmarshalConditionsCmd(t *testing.T) {
	var tests = []struct {
		name            string
		rawJSON         string
		expectedCommand *ConditionsCmd
		needsVars       []string
	}{
		{
			name: "basic threshold condition",
			rawJSON: `{
				"conditions": [
				  {
					"evaluator": {
					  "params": [
						2
					  ],
					  "type": "gt"
					},
					"operator": {
					  "type": "and"
					},
					"query": {
					  "params": [
						"A"
					  ]
					},
					"reducer": {
					  "params": [],
					  "type": "avg"
					},
					"type": "query"
				  }
				]
			}`,
			expectedCommand: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("avg"),
						Operator:   "and",
						Evaluator:  &thresholdEvaluator{Type: "gt", Threshold: 2},
					},
				},
			},
			needsVars: []string{"A"},
		},
		{
			name: "ranged condition",
			rawJSON: `{
				"conditions": [
				  {
					"evaluator": {
					  "params": [
						2,
						3
					  ],
					  "type": "within_range"
					},
					"operator": {
					  "type": "or"
					},
					"query": {
					  "params": [
						"A"
					  ]
					},
					"reducer": {
					  "params": [],
					  "type": "diff"
					},
					"type": "query"
				  }
				]
			}`,
			expectedCommand: &ConditionsCmd{
				Conditions: []condition{
					{
						InputRefID: "A",
						Reducer:    reducer("diff"),
						Operator:   "or",
						Evaluator:  &rangedEvaluator{Type: "within_range", Lower: 2, Upper: 3},
					},
				},
			},
			needsVars: []string{"A"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rq map[string]interface{}

			err := json.Unmarshal([]byte(tt.rawJSON), &rq)
			require.NoError(t, err)

			cmd, err := UnmarshalConditionsCmd(rq, "")
			require.NoError(t, err)
			require.Equal(t, tt.expectedCommand, cmd)

			require.Equal(t, tt.needsVars, cmd.NeedsVars())
		})
	}
}
