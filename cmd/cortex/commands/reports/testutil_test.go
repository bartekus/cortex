// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package reports

import (
	"github.com/bartekus/cortex/internal/reports/commithealth"
)

// fakeHistorySource is a test helper that implements HistorySource without shelling out.
type fakeHistorySource struct {
	commits []commithealth.CommitMetadata
}

func (f fakeHistorySource) Commits() ([]commithealth.CommitMetadata, error) {
	return f.commits, nil
}
