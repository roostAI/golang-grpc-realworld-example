package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"testing"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"github.com/raahii/golang-grpc-realworld-example/proto"
)

type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext
}
/*
ROOST_METHOD_HASH=run_9594c70ad3
ROOST_METHOD_SIG_HASH=run_9bb183262c


 */
func Testrun(t *testing.T) {
	type testScenario struct {
		description                string
		setupMocks                 func() (endFunc func(), regUserErr, regArticleErr error)
		expectedError              string
		assertListenAndServeCalled bool
	}

	tests := []testScenario{
		{
			description: "Successful Server Start",
			setupMocks: func() (func(), error, error) {
				return func() {}, nil, nil
			},
			expectedError:              "",
			assertListenAndServeCalled: true,
		},
		{
			description: "Failure on User Handler Registration",
			setupMocks: func() (func(), error, error) {
				return func() {}, errors.New("user handler error"), nil
			},
			expectedError:              "user handler error",
			assertListenAndServeCalled: false,
		},
		{
			description: "Failure on Article Handler Registration",
			setupMocks: func() (func(), error, error) {
				return func() {}, nil, errors.New("article handler error")
			},
			expectedError:              "article handler error",
			assertListenAndServeCalled: false,
		},
	}

	var buf bytes.Buffer
	flag.CommandLine.SetOutput(&buf)
	flag.CommandLine.Parse([]string{"-endpoint=localhost:50051"})

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			endFunc, regUserErr, regArticleErr := test.setupMocks()
			defer endFunc()

			origRegisterUsersHandlerFromEndpoint := gw.RegisterUsersHandlerFromEndpoint
			origRegisterArticlesHandlerFromEndpoint := gw.RegisterArticlesHandlerFromEndpoint
			gw.RegisterUsersHandlerFromEndpoint = func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
				return regUserErr
			}
			gw.RegisterArticlesHandlerFromEndpoint = func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
				return regArticleErr
			}
			defer func() {
				gw.RegisterUsersHandlerFromEndpoint = origRegisterUsersHandlerFromEndpoint
				gw.RegisterArticlesHandlerFromEndpoint = origRegisterArticlesHandlerFromEndpoint
			}()

			var outputBuffer bytes.Buffer
			stdout := log.Writer()
			defer func() { log.SetOutput(stdout) }()
			log.SetOutput(&outputBuffer)

			err := run()

			if test.expectedError == "" {
				assert.NoError(t, err, "expected no error but got an error")
				assert.Contains(t, outputBuffer.String(), "starting gateway server on port 3000", "should log start of server")
			} else {
				assert.Error(t, err, "expected an error but got no error")
				assert.EqualError(t, err, test.expectedError, fmt.Sprintf("expected error message to be '%s'", test.expectedError))
			}

			if test.assertListenAndServeCalled {
				assert.Contains(t, outputBuffer.String(), "starting gateway server on port 3000", "server should start successfully")
			} else {
				assert.NotContains(t, outputBuffer.String(), "starting gateway server on port 3000", "server should not start on failure")
			}

			t.Logf("Test Scenario: %s It was Successful", test.description)
		})
	}
}

