package db

import (
	"context"
	"fmt"
)

// SetupTiKVLists initializes blacklist/whitelist with sample data
func (c *Client) SetupTiKVLists() error {
	ctx := context.Background()
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tikv txn: %w", err)
	}
	defer txn.Rollback()

	// Sample blacklist entry
	err = txn.Set([]byte("blacklist:citizen003"), []byte("Blocked for fraud"))
	if err != nil {
		return fmt.Errorf("failed to set blacklist: %w", err)
	}

	// Sample whitelist entry
	err = txn.Set([]byte("whitelist:citizen001"), []byte("Trusted user"))
	if err != nil {
		return fmt.Errorf("failed to set whitelist: %w", err)
	}

	return txn.Commit(ctx)
}

// AddToBlacklist adds a citizen to the blacklist
func (c *Client) AddToBlacklist(citizenID, reason string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte("blacklist:" + citizenID)
	err = txn.Set(key, []byte(reason))
	if err != nil {
		return fmt.Errorf("failed to set blacklist entry: %w", err)
	}

	return txn.Commit(context.Background())
}

// IsBlacklisted checks if a citizen is on the blacklist
func (c *Client) IsBlacklisted(citizenID string) (bool, string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return false, "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte("blacklist:" + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return false, "", nil // Not found = not blacklisted
	}
	return true, string(val), nil
}

// AddToWhitelist adds a citizen to the whitelist
func (c *Client) AddToWhitelist(citizenID, reason string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte("whitelist:" + citizenID)
	err = txn.Set(key, []byte(reason))
	if err != nil {
		return fmt.Errorf("failed to set whitelist entry: %w", err)
	}

	return txn.Commit(context.Background())
}
