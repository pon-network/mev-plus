package data

import (
	"math/big"
	"testing"
	"time"
)

func TestAggregatorData(t *testing.T) {
	t.Run("SetLastSlot", func(t *testing.T) {
		aggregator := NewAggregatorData()
		aggregator.SetLastSlot(42)
		lastSlot := aggregator.GetLastSlot()
		if lastSlot != 42 {
			t.Errorf("Expected last slot to be 42, got %v", lastSlot)
		}
	})

	t.Run("SetLastSlotHeaderReleaseTime", func(t *testing.T) {
		aggregator := NewAggregatorData()
		releaseTime := time.Now()
		aggregator.SetLastSlotHeaderReleaseTime(releaseTime)
		lastReleaseTime := aggregator.GetLastSlotHeaderReleaseTime()
		if lastReleaseTime != releaseTime {
			t.Errorf("Expected last release time to be %v, got %v", releaseTime, lastReleaseTime)
		}
	})

	t.Run("AddSlotHeader", func(t *testing.T) {
		aggregator := NewAggregatorData()
		slotHeader := SlotHeader{
			Slot:      142,
			BlockHash: "hash123",
			Value:     big.NewInt(123),
		}

		// Add a new slot header
		err := aggregator.AddSlotHeader(slotHeader)
		if err != nil {
			t.Errorf("Error adding slot header: %v", err)
		}

		// Retrieve the added slot header
		retrievedSlotHeader, err := aggregator.GetSlotHeaderByHash("hash123")
		if err != nil {
			t.Errorf("Error retrieving slot header: %v", err)
		}

		if retrievedSlotHeader != slotHeader {
			t.Errorf("Retrieved slot header does not match the added one.")
		}
	})

	t.Run("GetSelectedSlotHeaders", func(t *testing.T) {
		aggregator := NewAggregatorData()
		slotHeader := SlotHeader{
			Slot:      142,
			BlockHash: "hash123",
			Value:     big.NewInt(123),
		}

		// Add a new slot header
		err := aggregator.AddSlotHeader(slotHeader)
		if err != nil {
			t.Errorf("Error adding slot header: %v", err)
		}

		// Retrieve the selected slot header for slot 142
		selectedSlotHeader, err := aggregator.GetSelectedSlotHeaders(142)
		if err != nil {
			t.Errorf("Error retrieving selected slot header: %v", err)
		}

		if selectedSlotHeader != slotHeader {
			t.Errorf("Retrieved selected slot header does not match the added one.")
		}
	})
}

func TestAggregatorDataNegative(t *testing.T) {
	t.Run("AddOldSlotHeader", func(t *testing.T) {
		aggregator := NewAggregatorData()
		// Set the last slot to 100
		aggregator.SetLastSlot(100)

		// Try to add a slot header for an older slot (99), which should result in an error
		err := aggregator.AddSlotHeader(SlotHeader{
			Slot:      99,
			BlockHash: "old_hash",
			Value:     big.NewInt(1000),
		})

		if err == nil {
			t.Error("Expected error for adding an old slot header, got nil")
		}
	})

	t.Run("GetMissingSlotHeaders", func(t *testing.T) {
		aggregator := NewAggregatorData()

		// Try to retrieve slot headers for a slot that has not been added, should result in an error
		_, err := aggregator.GetSelectedSlotHeaders(42)

		if err == nil {
			t.Error("Expected error for retrieving missing slot headers, got nil")
		}
	})

	t.Run("GetMissingSlotHeaderByHash", func(t *testing.T) {
		aggregator := NewAggregatorData()

		// Try to retrieve a slot header by hash that has not been added, should result in an error
		_, err := aggregator.GetSlotHeaderByHash("missing_hash")

		if err == nil {
			t.Error("Expected error for retrieving missing slot header by hash, got nil")
		}
	})
}
