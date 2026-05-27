// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"gitlab.com/blockchain-privacy/dakar/analytics"
	"gitlab.com/blockchain-privacy/dakar/cmd/cliutil"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// Tag represents a tag with an address and a flag indicating if it's a cluster definer
type Tag struct {
	Address     string `yaml:"address"`
	Label       string `yaml:"label"`
	Source      string `yaml:"source"`
	Category    string `yaml:"category"`
	Currency    string `yaml:"currency"`
	Description string `yaml:"description"`
}

type TagPack struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
	Confidence  string `yaml:"confidence"`
	Currency    string `yaml:"currency"`
	Label       string `yaml:"label"`
	Source      string `yaml:"source"`
	Tags        []Tag  `yaml:"tags"`
}

var excludedFiles = map[string]bool{"samourai.yaml": true}

func doInsertTagPacks(ctx context.Context, c external.Database, directory string) {
	tags := getBitcoinTags(directory)
	if len(tags) == 0 {
		info("No tags found")
		return
	}

	const steps = 1000
	stop := false
	var insertCountSum int
	for i := 0; !stop; i += steps {
		to := i + steps
		if to >= len(tags) {
			to = len(tags)
			stop = true
		}
		insertCount, err := analytics.PublicAttributionImport(ctx, c, tags[i:to])
		if err != nil {
			warn(err)
			return
		}
		insertCountSum += insertCount
	}

	info("finished inserting attributions", "attribution count", len(tags), "insert count", insertCountSum)
}

// getBitcoinTags reads all YAML files in the given directory and returns found attributions.
// The format is defined in https://github.com/graphsense/graphsense-tagpacks
func getBitcoinTags(directory string) []analytics.Attribution {
	const currency = "BTC"
	bitcoinTags := map[string]analytics.Attribution{}
	items, _ := os.ReadDir(directory)
	for _, item := range items {
		itemName := item.Name()
		if !item.IsDir() && (strings.HasSuffix(itemName, ".yml") || strings.HasSuffix(itemName, ".yaml")) && !excludedFiles[itemName] {
			file, err := os.ReadFile(directory + "/" + itemName)
			if err != nil {
				return nil
			}

			if file == nil {
				continue
			}

			pack := TagPack{}
			if err = yaml.Unmarshal(file, &pack); err != nil {
				warn(serror.New(err))
				continue
			}

			if pack.Currency != "" && pack.Currency != currency {
				continue
			}

			for _, tag := range pack.Tags {
				if tag.Currency != "" && tag.Currency != currency {
					continue
				}
				setValues(pack, &tag)
				bitcoinTags[tag.Address] = analytics.Attribution{
					AddressHash: tag.Address,
					Tag:         tag.Label,
					Description: tag.Description,
					Source:      tag.Source,
					Category:    tag.Category,
				}
			}
		}
	}
	return cliutil.GetMapValues(bitcoinTags)
}

func setValues(pack TagPack, tag *Tag) {
	if tag.Currency == "" {
		tag.Currency = pack.Currency
	}

	if tag.Label == "" {
		tag.Label = pack.Label
	}
	if tag.Source == "" {
		tag.Source = pack.Source
	}

	if tag.Category == "" {
		tag.Category = pack.Category
	}

	if tag.Description == "" {
		tag.Description = pack.Description
	}
}
