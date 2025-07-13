package data

import "embed"

// Embedded data for daily foraging hours.
//
// # Available data
//
//   - foraging-period/berlin2000.txt
//   - foraging-period/berlin2001.txt
//   - foraging-period/berlin2002.txt
//   - foraging-period/berlin2003.txt
//   - foraging-period/berlin2004.txt
//   - foraging-period/berlin2005.txt
//   - foraging-period/berlin2006.txt
//   - foraging-period/foragingHoursListExample.txt
//   - foraging-period/Sweden2010.txt
//   - foraging-period/Valencia2010.txt
//	 - foraging-period/tunnel.txt
//   - foraging-period/rothamsted2009.txt
//
//go:embed foraging-period
var ForagingPeriod embed.FS

// Embedded data for daily water needs.
//
// # Available data
//
//   - ETOX_waterforcooling_daily/waterlistExample.txt
//   - ETOX_waterforcooling_daily/waterlistempty.txt
//   - ETOX_waterforcooling_daily/waterlistValencia.txt
//
//go:embed ETOX_waterforcooling_daily
var WaterNeedsDaily embed.FS
