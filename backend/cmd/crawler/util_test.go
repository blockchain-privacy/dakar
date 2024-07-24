package main

import (
	ory "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheckHTTPModuleConfig(t *testing.T) {
	type moduleTest struct {
		module     APIModule
		shouldFail bool
	}

	var cases = []moduleTest{
		{
			module:     APIModule{},
			shouldFail: true,
		},
		{
			module: APIModule{
				Active:               true,
				Port:                 300,
				KratosPublicEndpoint: "123456",
				KratosAdminEndpoint:  "123456",
			},
			shouldFail: false,
		},
		{
			module: APIModule{
				Active:               false,
				Port:                 0,
				KratosPublicEndpoint: "localhost:4433",
				KratosAdminEndpoint:  "localhost:1234567",
			},
			shouldFail: false,
		},
	}

	for _, c := range cases {
		if c.shouldFail {
			require.Error(t, checkAPIModuleConfig(c.module))
		} else {
			require.NoError(t, checkAPIModuleConfig(c.module))
		}
	}
}

func TestPrintVersion(t *testing.T) {
	require.NotPanics(t, printVersion)
}

func TestNewKratosClient(t *testing.T) {
	type clientTest struct {
		endpoint   string
		shouldFail bool
	}

	var cases = []clientTest{
		{
			endpoint:   "",
			shouldFail: true,
		},
		{
			endpoint:   "123456",
			shouldFail: false,
		},
		{
			endpoint:   "some_random_text",
			shouldFail: false,
		},
	}

	for _, c := range cases {
		client, err := newKratosClient(c.endpoint)

		if c.shouldFail {
			require.Error(t, err)
			require.Nil(t, client)
		} else {
			require.NoError(t, err)
			require.NotNil(t, client)
		}
	}
}

func TestIsKratosAlive(t *testing.T) {
	type kratosTest struct {
		auth       *ory.APIClient
		shouldFail bool
	}

	client, err := newKratosClient("invalid_endpoint")
	if err != nil {
		t.Failed()
	}

	var cases = []kratosTest{
		{
			auth:       nil,
			shouldFail: true,
		},
		{
			auth: &ory.APIClient{
				MetadataAPI: nil,
			},
			shouldFail: true,
		},
		{
			auth:       client,
			shouldFail: true,
		},
	}

	for _, c := range cases {
		require.Equal(t, c.shouldFail, !isKratosAlive(c.auth))
	}
}
