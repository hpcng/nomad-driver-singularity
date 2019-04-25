// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package singularity

import (
	hclog "github.com/hashicorp/go-hclog"

	"testing"
)

func TestNewSingularityDriver(t *testing.T) {
	tests := []struct {
		name   string
		logger hclog.Logger
	}{
		{"DefaultLogger", hclog.Default()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewSingularityDriver(tt.logger)

			if d == nil {
				t.Fatalf("got nil logger")
			}
		})
	}
}
