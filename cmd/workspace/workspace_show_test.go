package workspace

import (
	"bytes"
	"context"
	"testing"

	"github.com/leg100/stok/pkg/app"
	"github.com/leg100/stok/pkg/env"
	"github.com/leg100/stok/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceShow(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  env.StokEnv
		out  string
		err  bool
	}{
		{
			name: "WithEnvironmentFile",
			args: []string{"show"},
			env:  env.StokEnv("default/workspace-1"),
			out:  "default/workspace-1\n",
		},
		{
			name: "WithoutEnvironmentFile",
			args: []string{"show"},
			out:  "default/default\n",
		},
	}

	for _, tt := range tests {
		testutil.Run(t, tt.name, func(t *testutil.T) {
			path := t.NewTempDir().Chdir().Root()

			// Write .terraform/environment
			if tt.env != "" {
				require.NoError(t, tt.env.Write(path))
			}

			out := new(bytes.Buffer)

			opts, err := app.NewFakeOpts(out)
			require.NoError(t, err)

			cmd := ShowCmd(opts)
			cmd.SetOut(opts.Out)
			cmd.SetArgs(tt.args)

			t.CheckError(tt.err, cmd.ExecuteContext(context.Background()))

			assert.Equal(t, tt.out, out.String())
		})
	}
}