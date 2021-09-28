// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package quota

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetUsedProjectIDs(t *testing.T) {
	type Expectation struct {
		ProjectIDs []int
		Error      string
	}
	tests := []struct {
		Name        string
		Input       string
		InputErr    error
		Expectation Expectation
	}{
		{
			Name:  "no projects",
			Input: "",
		},
		{
			Name:  "single project",
			Input: "#0              0      0      0  00 [------]",
			Expectation: Expectation{
				ProjectIDs: []int{0},
			},
		},
		{
			Name:  "multiple projects",
			Input: "#0              0      0      0  00 [------]\n#100            0     5M     5M  00 [------]\n#200            0    10M    10M  00 [------]",
			Expectation: Expectation{
				ProjectIDs: []int{0, 100, 200},
			},
		},
		{
			Name:     "exec failure",
			InputErr: fmt.Errorf("exec failed"),
			Expectation: Expectation{
				Error: "exec failed",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			xfs := &XFS{
				exec: func(dir, command string) (output string, err error) {
					return test.Input, test.InputErr
				},
			}

			var (
				act Expectation
				err error
			)
			act.ProjectIDs, err = xfs.getUsedProjectIDs()
			if err != nil {
				act.Error = err.Error()
			}

			if diff := cmp.Diff(test.Expectation, act); diff != "" {
				t.Errorf("unexpected getUsedProjectIDs (-want +got):\n%s", diff)
			}
		})
	}
}
