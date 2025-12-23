// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package commithealth defines the data model for commit health reports.
//
// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md
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
