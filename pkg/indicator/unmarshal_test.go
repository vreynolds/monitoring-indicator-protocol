package indicator_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestDocumentFromYAML(t *testing.T) {
	t.Run("returns error if YAML is bad", func(t *testing.T) {
		g := NewGomegaWithT(t)
		t.Run("bad document", func(t *testing.T) {
			reader := ioutil.NopCloser(strings.NewReader(`--`))
			_, errs := indicator.DocumentFromYAML(reader)
			g.Expect(errs).ToNot(BeEmpty())
		})
	})

	t.Run("apiVersion v1", func(t *testing.T) {
		t.Run("parses all document fields", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name
  labels:
    deployment: well-performing-deployment

spec:
  product: 
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test_performance_indicator
    documentation:
      title: Test Performance Indicator
      description: This is a valid markdown description.
      recommendedResponse: Panic!
      thresholdNote: Threshold Note Text
    thresholds:
    - level: warning
      operator: lte
      value: 500
      alert:
        for: 1m
        step: 1m
    promql: prom{deployment="$deployment"}
    presentation:
      currentValue: false
      chartType: step
      frequency: 5
      labels:
      - job
      - ip
      units: nanoseconds

  layout:
    title: Monitoring Test Product
    description: Test description
    sections:
    - title: Test Section
      description: This section includes indicators and metrics
      indicators:
      - test_performance_indicator
`))
			doc, errs := indicator.DocumentFromYAML(reader, indicator.SkipMetadataInterpolation)
			g.Expect(errs).To(BeEmpty())

			indie := v1.IndicatorSpec{
				Name:   "test_performance_indicator",
				PromQL: `prom{deployment="$deployment"}`,
				Thresholds: []v1.Threshold{{
					Level:    "warning",
					Operator: v1.LessThanOrEqualTo,
					Value:    500,
					Alert:    test_fixtures.DefaultAlert(),
				}},
				Presentation: v1.Presentation{
					CurrentValue: false,
					ChartType:    v1.StepChart,
					Frequency:    5,
					Labels:       []string{"job", "ip"},
					Units:        "nanoseconds",
				},
				Documentation: map[string]string{
					"title":               "Test Performance Indicator",
					"description":         "This is a valid markdown description.",
					"recommendedResponse": "Panic!",
					"thresholdNote":       "Threshold Note Text",
				},
			}
			g.Expect(doc).To(BeEquivalentTo(v1.IndicatorDocument{
				TypeMeta: metav1.TypeMeta{
					APIVersion: api_versions.V1,
					Kind:       "IndicatorDocument",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "document name",
					Labels: map[string]string{"deployment": "well-performing-deployment"},
				},
				Spec: v1.IndicatorDocumentSpec{
					Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
					Indicators: []v1.IndicatorSpec{
						indie,
					},
					Layout: v1.Layout{
						Title:       "Monitoring Test Product",
						Description: "Test description",
						Sections: []v1.Section{{
							Title:       "Test Section",
							Description: "This section includes indicators and metrics",
							Indicators:  []string{indie.Name},
						}},
					},
				},
			}))
		})

		t.Run("populates all defaults when not provided", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name
  labels:
    deployment: valid-deployment

spec:
  product:
    name: well-performing-component
    version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: promql_query
    thresholds:
    - operator: lt
      value: 0
      level: warning
`))
			d, errs := indicator.DocumentFromYAML(reader)
			g.Expect(errs).To(BeEmpty())

			g.Expect(d.Spec.Layout).To(Equal(v1.Layout{
				Title: "well-performing-component - 0.0.1",
				Sections: []v1.Section{{
					Title: "Metrics",
					Indicators: []string{
						"test_performance_indicator",
					},
				}},
			}))

			g.Expect(d.Spec.Indicators[0].Thresholds[0].Alert).To(Equal(test_fixtures.DefaultAlert()))

			g.Expect(d.Spec.Indicators[0].Presentation).To(Equal(test_fixtures.DefaultPresentation()))
		})

		t.Run("handles thresholds", func(t *testing.T) {
			t.Run("it handles all the operators", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name

spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test_kpi
    promql: prom
    thresholds:
    - operator: lt
      value: 0
      level: warning
    - operator: lte
      value: 1.2
      level: warning
    - operator: eq
      value: 0.2
      level: warning
    - operator: neq
      value: 123
      level: warning
    - operator: gte
      value: 642
      level: warning
    - operator: gt
      value: 1.222225
      level: warning`))

				d, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).To(BeEmpty())

				g.Expect(d.Spec.Indicators[0].Thresholds).To(Equal([]v1.Threshold{
					{
						Level:    "warning",
						Operator: v1.LessThan,
						Value:    0,
						Alert:    test_fixtures.DefaultAlert(),
					},
					{
						Level:    "warning",
						Operator: v1.LessThanOrEqualTo,
						Value:    1.2,
						Alert:    test_fixtures.DefaultAlert(),
					},
					{
						Level:    "warning",
						Operator: v1.EqualTo,
						Value:    0.2,
						Alert:    test_fixtures.DefaultAlert(),
					},
					{
						Level:    "warning",
						Operator: v1.NotEqualTo,
						Value:    123,
						Alert:    test_fixtures.DefaultAlert(),
					},
					{
						Level:    "warning",
						Operator: v1.GreaterThanOrEqualTo,
						Value:    642,
						Alert:    test_fixtures.DefaultAlert(),
					},
					{
						Level:    "warning",
						Operator: v1.GreaterThan,
						Value:    1.222225,
						Alert:    test_fixtures.DefaultAlert(),
					},
				}))
			})

			t.Run("it handles unknown operator", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test_kpi
    description: desc
    promql: prom
    thresholds:
    - level: warning
      value: 500
      operator: foo
  `))

				_, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).ToNot(BeEmpty())
			})

			t.Run("it handles missing operator", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test_kpi
    description: desc
    promql: prom
    thresholds:
    - level: warning
  `))

				_, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).ToNot(BeEmpty())
			})

			t.Run("it returns an error if value is not a number", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test_kpi
    description: desc
    promql: prom
    thresholds:
    - value: abs
      operator: gt
      level: warning
  `))

				_, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).ToNot(BeEmpty())
			})
		})

		t.Run("handles presentation chart types", func(t *testing.T) {
			t.Run("can set a step chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name

spec:
  product:
   name: test_product
   version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: step
`))
				d, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).To(BeEmpty())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.StepChart))
			})

			t.Run("can set a bar chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name

spec:
  product:
   name: test_product
   version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: bar
`))
				d, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).To(BeEmpty())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.BarChart))
			})

			t.Run("can set a status chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name

spec:
  product:
   name: test_product
   version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: status
`))
				d, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).To(BeEmpty())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.StatusChart))
			})

			t.Run("can set a quota chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name

spec:
  product:
    name: test_product
    version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: quota
`))
				d, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).To(BeEmpty())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.QuotaChart))
			})
		})

		t.Run("handles indicator types", func(t *testing.T) {
			getDoc := func(indiType string) string {
				return fmt.Sprintf(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name

spec:
  product:
   name: test_product
   version: 0.0.1

  indicators:
  - name: test_performance_indicator
    type: %s
    promql: prom{deployment="test"}
    presentation:
      chartType: step
`, indiType)
			}
			g := NewGomegaWithT(t)

			testCases := []struct {
				indiTypeString string
				indiType       v1.IndicatorType
			}{
				{"sli", v1.ServiceLevelIndicator},
				{"kpi", v1.KeyPerformanceIndicator},
				{"other", v1.DefaultIndicator},
			}

			for _, testCase := range testCases {
				yamlString := getDoc(testCase.indiTypeString)
				reader := ioutil.NopCloser(strings.NewReader(yamlString))
				d, errs := indicator.DocumentFromYAML(reader)
				g.Expect(errs).To(BeEmpty())
				g.Expect(d.Spec.Indicators[0].Type).To(Equal(testCase.indiType),
					fmt.Sprintf("Failed indiTypeString: `%s`", testCase.indiTypeString))
			}

			yamlString := getDoc("")
			reader := ioutil.NopCloser(strings.NewReader(yamlString))
			_, errs := indicator.DocumentFromYAML(reader)
			g.Expect(errs).ToNot(BeEmpty())
		})

		t.Run("handles defaulting indicator types", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name

spec:
  product:
    name: test_product
    version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: step
`))
			d, errs := indicator.DocumentFromYAML(reader)
			g.Expect(errs).To(BeEmpty())
			g.Expect(d.Spec.Indicators[0].Type).To(Equal(v1.DefaultIndicator))
		})
	})
}

func TestPatchFromYAML(t *testing.T) {
	t.Run("apiVersion indicatorprotocol.io/v1", func(t *testing.T) {
		t.Run("parses all the fields", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocumentPatch

match:
  product:
    name: my-other-component
    version: 1.2.3

operations:
- type: replace
  path: /spec/indicators/0/thresholds?/-
  value:
    level: warning
    operator: gt
    value: 100
`))
			p, err := indicator.PatchFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			var patchedThreshold interface{}
			patchedThreshold = map[string]interface{}{
				"level":    "warning",
				"operator": "gt",
				"value":    float64(100),
			}
			expectedPatch := indicator.Patch{
				APIVersion: api_versions.V1,
				Match: indicator.Match{
					Name:    test_fixtures.StrPtr("my-other-component"),
					Version: test_fixtures.StrPtr("1.2.3"),
				},
				Operations: []patch.OpDefinition{{
					Type:  "replace",
					Path:  test_fixtures.StrPtr("/spec/indicators/0/thresholds?/-"),
					Value: &patchedThreshold,
				}},
			}

			g.Expect(p).To(BeEquivalentTo(expectedPatch))
		})

		t.Run("parses empty product name and version", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocumentPatch

match:
  metadata:
    deployment: test-deployment

operations:
- type: replace
  path: /spec/indicators/name=success_percentage
  value:
    promql: success_percentage_promql{source_id="origin"}
    documentation:
      title: Success Percentage

`))
			p, err := indicator.PatchFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(p.Match.Name).To(BeNil())
			g.Expect(p.Match.Version).To(BeNil())
		})
	})
}

func TestProductFromYAML(t *testing.T) {
	t.Run(api_versions.V1, func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
spec:
  product:
    name: indi-pro
    version: 1.2.3
`))
		p, err := indicator.ProductFromYAML(reader)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(p).To(BeEquivalentTo(v1.Product{
			Name:    "indi-pro",
			Version: "1.2.3",
		}))
	})
}

func TestMetadataFromYAML(t *testing.T) {
	t.Run("parses all the fields in v1 documents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1

spec:
  product:
    name: indi-pro
    version: 1.2.3

metadata:
  name: document name
  labels:
    sound: meow
    size: small
    color: tabby
`))
		p, err := indicator.MetadataFromYAML(reader)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(p).To(BeEquivalentTo(map[string]string{
			"sound": "meow",
			"size":  "small",
			"color": "tabby",
		}))
	})
}

func TestProcessesDocument(t *testing.T) {
	t.Run("does not mess up thresholds in apiVersion v1", func(t *testing.T) {
		g := NewGomegaWithT(t)
		doc := []byte(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  name: document name

spec:
  product:
    name: testing
    version: v123
  indicators:
  - name: test_indicator
    promql: test_expr
    thresholds:
    - level: critical
      operator: neq
      value: 100
`)
		resultDoc, err := indicator.ProcessDocument([]indicator.Patch{}, doc)
		g.Expect(err).To(HaveLen(0))
		g.Expect(resultDoc.Spec.Indicators[0].Thresholds[0]).To(BeEquivalentTo(v1.Threshold{
			Level:    "critical",
			Operator: v1.NotEqualTo,
			Value:    100,
			Alert:    test_fixtures.DefaultAlert(),
		}))
	})

}

func TestApiVersionFromYAML(t *testing.T) {
	t.Run("returns a useful error when document is not valid YAML", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := []byte(`---
apiVersion: v1

product:
  name: indi-pro
  - lol
  version: 1.2.3
`)

		_, err := indicator.ApiVersionFromYAML(doc)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("could not unmarshal apiVersion, check that document contains valid YAML"))
	})
}
