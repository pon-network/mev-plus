package data

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type AggregatorData struct {
	mu sync.Mutex

	lastSlot                  uint64
	lastSlotHeaderReleaseTime time.Time // The time at which the slot header is released

	selectedSlotHeaders map[uint64][]SlotHeader
}

func NewAggregatorData() *AggregatorData {
	return &AggregatorData{
		selectedSlotHeaders: make(map[uint64][]SlotHeader),
	}
}

func (d *AggregatorData) SetLastSlot(slot uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.lastSlot = slot
}

func (d *AggregatorData) GetLastSlot() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.lastSlot
}

func (d *AggregatorData) SetLastSlotHeaderReleaseTime(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.lastSlotHeaderReleaseTime = t
}

func (d *AggregatorData) GetLastSlotHeaderReleaseTime() time.Time {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.lastSlotHeaderReleaseTime
}

func (d *AggregatorData) AddSlotHeader(slotHeader SlotHeader) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if slotHeader.Slot < d.lastSlot {
		return fmt.Errorf("addition of old slot header")
	}

	d.selectedSlotHeaders[slotHeader.Slot] = append(d.selectedSlotHeaders[slotHeader.Slot], slotHeader)

	// sort the headers by value and remove the least valuable if more than 3
	// 3 are kept in the case of when trying to retrieve the payload from
	// any module, there is a chance that the payload is not available for the
	// first 2 headers
	sort.Slice(d.selectedSlotHeaders[slotHeader.Slot], func(i, j int) bool {
		return d.selectedSlotHeaders[slotHeader.Slot][i].Value.Cmp(d.selectedSlotHeaders[slotHeader.Slot][j].Value) > 0
	})
	if len(d.selectedSlotHeaders[slotHeader.Slot]) > 3 {
		d.selectedSlotHeaders[slotHeader.Slot] = d.selectedSlotHeaders[slotHeader.Slot][:3]
	}

	d.lastSlot = slotHeader.Slot

	// delete the slot headers for slots older than 2 epochs
	for slot := range d.selectedSlotHeaders {
		if slot < d.lastSlot-64 {
			delete(d.selectedSlotHeaders, slot)
		}
	}

	return nil

}

func (d *AggregatorData) GetSlotHeaderByHash(blockHash string) (SlotHeader, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, slotHeaders := range d.selectedSlotHeaders {
		for _, slotHeader := range slotHeaders {
			if slotHeader.BlockHash == blockHash {
				return slotHeader, nil
			}
		}
	}

	return SlotHeader{}, fmt.Errorf("slot header with BlockHash %s not found", blockHash)
}

func (d *AggregatorData) GetSelectedSlotHeaders(slot uint64) (SlotHeader, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if the slot is in the map
	if _, ok := d.selectedSlotHeaders[slot]; !ok {
		return SlotHeader{}, fmt.Errorf("slot %v not found", slot)
	}

	return d.selectedSlotHeaders[slot][0], nil
}
