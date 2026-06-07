# Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
# See the file LICENSE for licensing terms.
#
# Wire schema for the REQL→LUX airdrop claims persistence. The hand-
# rolled codec lives in node/vms/platformvm/airdrop/codec_zap.go; this
# schema records the wire shape for the centralized registry.

package airdrop

# Envelope is a single ZAP object carrying a count + variable-tail
# entries blob. Each entry is a length-prefixed ClaimEntry record;
# the entry layout is documented inline in codec_zap.go.
struct AirdropClaimsWire {
    Kind     u8    @0       # = 1 (claim list)
    Count    u32   @4
    Entries  bytes @8        # concatenated, length-prefixed ClaimEntry records
}
