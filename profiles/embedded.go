package profiles

import _ "embed"

// We embed these profiles because we need to install
// them in the DB for new DART installations.

//go:embed aptrust-v2.2.json
var APTrust_V_2_2 string

//go:embed btr-v1.0-1.3.0.json
var BTR_V_1_0 string

//go:embed empty_profile.json
var Empty_V_1_0 string
