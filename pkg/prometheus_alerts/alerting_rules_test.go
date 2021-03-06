package prometheus_alerts_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestAlertGeneration(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it makes prometheus alerts from the thresholds", func(t *testing.T) {
		g = NewGomegaWithT(t)
		document := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{
					{
						Thresholds: []v1.Threshold{{}, {}},
					},
					{
						Thresholds: []v1.Threshold{{}},
					},
				},
			},
		}

		alertDoc := prometheus_alerts.AlertDocumentFrom(document)

		g.Expect(alertDoc.Groups[0].Rules).To(HaveLen(3))
	})

	t.Run("it generates a promql statement for less than statements", func(t *testing.T) {
		g = NewGomegaWithT(t)

		exprFor := func(op v1.ThresholdOperator) string {
			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							PromQL: `metric{source_id="fake-source"}`,
							Thresholds: []v1.Threshold{{
								Level:    "warning",
								Operator: op,
								Value:    0.99999999999999, //keep many nines to ensure we don't blow float parsing to 1
							}},
						},
					},
				},
			}

			return getFirstRule(doc).Expr
		}

		g.Expect(exprFor(v1.LessThanOrEqualTo)).To(Equal(`metric{source_id="fake-source"} <= 0.99999999999999`))
		g.Expect(exprFor(v1.LessThan)).To(Equal(`metric{source_id="fake-source"} < 0.99999999999999`))
		g.Expect(exprFor(v1.EqualTo)).To(Equal(`metric{source_id="fake-source"} == 0.99999999999999`))
		g.Expect(exprFor(v1.NotEqualTo)).To(Equal(`metric{source_id="fake-source"} != 0.99999999999999`))
		g.Expect(exprFor(v1.GreaterThan)).To(Equal(`metric{source_id="fake-source"} > 0.99999999999999`))
		g.Expect(exprFor(v1.GreaterThanOrEqualTo)).To(Equal(`metric{source_id="fake-source"} >= 0.99999999999999`))
	})

	t.Run("sets the name to the indicator's name", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{{
					Name:       "indicator_lol",
					Thresholds: []v1.Threshold{{}},
				}},
			},
		}

		g.Expect(getFirstRule(doc).Alert).To(Equal("indicator_lol"))
	})

	t.Run("sets the labels", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"meta-lol": "data-lol"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "product-lol", Version: "beta.9"},
				Indicators: []v1.IndicatorSpec{{
					Name: "indicator_lol",
					Thresholds: []v1.Threshold{{
						Level: "warning",
					}},
				}},
			},
		}

		g.Expect(getFirstRule(doc).Labels).To(Equal(map[string]string{
			"product":  "product-lol",
			"version":  "beta.9",
			"level":    "warning",
			"meta-lol": "data-lol",
		}))
	})

	t.Run("sets the annotations to the documentation block", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{{
					Documentation: map[string]string{"title-lol": "Indicator LOL"},
					Thresholds:    []v1.Threshold{{}},
				}},
			},
		}

		g.Expect(getFirstRule(doc).Annotations).To(Equal(map[string]string{
			"title-lol": "Indicator LOL",
		}))
	})

	t.Run("sets the alert for", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"meta-lol": "data-lol"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "product-lol", Version: "beta.9"},
				Indicators: []v1.IndicatorSpec{{
					Name:   "indicator_lol",
					PromQL: "promql_expression",
					Thresholds: []v1.Threshold{{
						Level:    "boo",
						Operator: v1.NotEqualTo,
						Value:    0,
						Alert: v1.Alert{
							For: "40h",
						},
					}},
				}},
			},
		}

		g.Expect(getFirstRule(doc).For).To(Equal("40h"))
	})

	t.Run("interpolates $step", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"meta-lol": "data-lol"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "product-lol", Version: "beta.9"},
				Indicators: []v1.IndicatorSpec{{
					Name:   "indicator_lol",
					PromQL: "super_query(promql_expression[$step])[$step]",
					Thresholds: []v1.Threshold{{
						Level:    "warning",
						Operator: v1.LessThan,
						Value:    0,
						Alert: v1.Alert{
							Step: "12m",
						},
					}},
				}},
			},
		}

		g.Expect(getFirstRule(doc).Expr).To(Equal("super_query(promql_expression[12m])[12m] < 0"))
	})

	t.Run("creates a filename based on product name and document contents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"deployment": "test_deployment"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "test_product",
					Version: "v1.2.3",
				},
				Indicators: []v1.IndicatorSpec{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Thresholds: []v1.Threshold{{
						Level:    "critical",
						Operator: v1.LessThan,
						Value:    5,
						Alert:    test_fixtures.DefaultAlert(),
					}},
					Presentation:  test_fixtures.DefaultPresentation(),
					Documentation: map[string]string{"title": "Test Indicator Title"},
				}},
				Layout: v1.Layout{
					Title: "Test Dashboard",
					Sections: []v1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		docBytes, err := json.Marshal(document)

		g.Expect(err).ToNot(HaveOccurred())
		filename := prometheus_alerts.AlertDocumentFilename(docBytes, "test_product")
		g.Expect(filename).To(MatchRegexp("test_product_[0-9a-f]{40}.yml"))
	})
}

func getFirstRule(from v1.IndicatorDocument) prometheus_alerts.Rule {
	return prometheus_alerts.AlertDocumentFrom(from).Groups[0].Rules[0]
}
