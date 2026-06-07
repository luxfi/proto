# Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
# See the file LICENSE for licensing terms.
#
# Wire schemas for the SQLiteMaterializer's internal blobs, replacing
# json/v2 for: collection schemas, dynamic-table user data, and the
# counter operation payload. Hand-rolled codec lives in
# node/vms/chainadapter/appchain_zap.go.

package chainadapter

# Collection schema persisted in _schemas.schema_zap.
struct CollectionSchemaWire {
    Kind        u8    @0       # = 1
    CRDTType    u8    @1
    Encrypted   u8    @2
    Domain      u8    @3
    Name        bytes @4
    PrimaryKey  bytes @12
    Fields      bytes @20       # concatenated FieldSchema records (var-len)
    Indexes     bytes @28       # concatenated IndexSchema records (var-len)
}

# User-data envelope (flat map[string]any with typed leaves).
struct UserDataWire {
    Kind     u8    @0           # = 2
    Entries  bytes @1            # flat (varbytes key, tag, value) records
}

# Counter operation payload (engine-internal Field/Delta).
struct CounterOpWire {
    Kind   u8    @0              # = 3
    Delta  i64   @1
    Field  bytes @9
}
