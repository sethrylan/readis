package main

import (
	"fmt"
	"time"

	"github.com/sethrylan/readis/internal/data"

	"charm.land/lipgloss/v2"
	"github.com/dustin/go-humanize"
)

// keyItem represents a Redis key, and implements [list.Item]
type keyItem struct {
	data.Key
}

func (k keyItem) String() string {
	return fmt.Sprintf("%s (%s)", k.Name, k.Datatype)
}

func (k keyItem) TTLString() string {
	if k.TTL == -1 {
		return "∞"
	}
	return humanize.RelTime(time.Now(), time.Now().Add(k.TTL), "", "")
}

func (k keyItem) SizeString() string {
	return humanize.Bytes(k.Size)
}

func (k keyItem) Title() string {
	typeLabel := lipgloss.NewStyle().Background(colorForKeyType(k.Datatype)).Render(k.Datatype)
	return lipgloss.NewStyle().Width(typeLabelWidth).Render(typeLabel) +
		lipgloss.NewStyle().Width(keyNameWidth).Inline(true).Render(k.Name) +
		lipgloss.NewStyle().Width(ttlWidth).Render(k.TTLString()) +
		lipgloss.NewStyle().Width(sizeWidth).Render(k.SizeString())
}

func (k keyItem) Description() string {
	return ""
}

func (k keyItem) FilterValue() string {
	return k.Name
}
