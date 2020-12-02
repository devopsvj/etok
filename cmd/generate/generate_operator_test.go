package generate

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	cmdutil "github.com/leg100/etok/cmd/util"
	"github.com/leg100/etok/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOperator(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		err        bool
		setup      func(*testutil.T, *GenerateOperatorOptions)
		assertions func(*testutil.T, *bytes.Buffer)
	}{
		{
			name: "local",
			args: []string{"generate", "operator", "--local"},
			setup: func(t *testutil.T, o *GenerateOperatorOptions) {
				// Default local path to CRDs is relative to repo root
				t.Chdir("../../")
			},
			assertions: func(t *testutil.T, out *bytes.Buffer) {
				clusterRole, _ := ioutil.ReadFile(clusterRolePath)
				assert.Regexp(t, string(clusterRole), out.String())
			},
		},
		{
			name: "remote",
			args: []string{"generate", "operator"},
			setup: func(t *testutil.T, o *GenerateOperatorOptions) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, "---\ntest: yaml")
				}))
				o.RemoteClusterRoleURL = ts.URL
				t.Cleanup(ts.Close)
			},
			assertions: func(t *testutil.T, out *bytes.Buffer) {
				assert.Regexp(t, "^---\ntest: yaml\n", out.String())
			},
		},
		{
			name: "remote failure",
			args: []string{"generate", "operator"},
			setup: func(t *testutil.T, o *GenerateOperatorOptions) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
				o.RemoteClusterRoleURL = ts.URL
				t.Cleanup(ts.Close)
			},
			err: true,
		},
	}
	for _, tt := range tests {
		testutil.Run(t, tt.name, func(t *testutil.T) {
			out := new(bytes.Buffer)
			opts, err := cmdutil.NewFakeOpts(out)
			require.NoError(t, err)

			cmd, cmdOpts := GenerateOperatorCmd(opts)
			cmd.SetOut(out)
			cmd.SetArgs(tt.args)

			if tt.setup != nil {
				tt.setup(t, cmdOpts)
			}

			t.CheckError(tt.err, cmd.ExecuteContext(context.Background()))

			if tt.assertions != nil {
				tt.assertions(t, out)
			}
		})
	}
}
