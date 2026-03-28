// Package execution wires Deribit private order RPCs to local order state, fills, and exposure (WP10).
//
// Post-only is the default for option entries; exits should set reduce_only on the persisted OrderRecord.
// Use [Placer.DryRun] in tests so JSON-RPC is never invoked.
package execution
