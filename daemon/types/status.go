package types

import "time"

type (
	// HostAlertID the unique id of an alert
	HostAlertID string

	// HostAlert an alert
	HostAlert struct {
		Type     string `json:"type"`
		Text     string `json:"text"`
		Severity string `json:"severity"`
	}

	// HostStatus status information about the host
	HostStatus struct {
		HostMeta
		Online             bool      `json:"online"`
		AcceptingContracts bool      `json:"accepting_contracts"`
		WalletUnlocked     bool      `json:"wallet_unlocked"`
		Version            string    `json:"version"`
		StartTime          time.Time `json:"start_time"`
	}
)
