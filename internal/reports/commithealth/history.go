// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package commithealth defines the data model for commit health reports.
//
// Feature: GOV_CORE
// Docs: docs/design/commit-reports-go-types.md
package commithealth

// CommitMetadata represents a single commit's metadata.
type CommitMetadata struct {
	SHA         string
	Message     string
	AuthorName  string
	AuthorEmail string
}

// HistorySource provides commit history for analysis.
type HistorySource interface {
	Commits() ([]CommitMetadata, error)
}
