// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/client/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type AppsV1Interface interface {
	RESTClient() rest.Interface
	IndicatorsGetter
	IndicatorDocumentsGetter
}

// AppsV1Client is used to interact with features provided by the indicatorprotocol.io group.
type AppsV1Client struct {
	restClient rest.Interface
}

func (c *AppsV1Client) Indicators(namespace string) IndicatorInterface {
	return newIndicators(c, namespace)
}

func (c *AppsV1Client) IndicatorDocuments(namespace string) IndicatorDocumentInterface {
	return newIndicatorDocuments(c, namespace)
}

// NewForConfig creates a new AppsV1Client for the given config.
func NewForConfig(c *rest.Config) (*AppsV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &AppsV1Client{client}, nil
}

// NewForConfigOrDie creates a new AppsV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *AppsV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new AppsV1Client for the given RESTClient.
func New(c rest.Interface) *AppsV1Client {
	return &AppsV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *AppsV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
